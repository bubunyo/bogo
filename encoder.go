package bogo

import (
	"fmt"
	"io"
)

// Encoder provides structured encoding with configurable options
type Encoder struct {
	// Configuration options
	MaxDepth        int  // Maximum nesting depth for objects/arrays (0 = unlimited)
	StrictMode      bool // Strict type checking and validation
	CompactArrays   bool // Use typed arrays when beneficial
	ValidateStrings bool // Validate UTF-8 encoding in strings
	
	// Internal state
	depth int
}

// EncoderOption is a function type for configuring an Encoder
type EncoderOption func(*Encoder)

// NewConfigurableEncoder creates a new Encoder with optional configuration
func NewConfigurableEncoder(options ...EncoderOption) *Encoder {
	e := &Encoder{
		MaxDepth:        100, // Default max depth
		StrictMode:      false,
		CompactArrays:   true,
		ValidateStrings: true,
	}
	
	for _, option := range options {
		option(e)
	}
	
	return e
}

// Encoder option functions
func WithMaxDepth(depth int) EncoderOption {
	return func(e *Encoder) {
		e.MaxDepth = depth
	}
}

func WithStrictMode(strict bool) EncoderOption {
	return func(e *Encoder) {
		e.StrictMode = strict
	}
}

func WithCompactArrays(compact bool) EncoderOption {
	return func(e *Encoder) {
		e.CompactArrays = compact
	}
}

func WithStringValidation(validate bool) EncoderOption {
	return func(e *Encoder) {
		e.ValidateStrings = validate
	}
}

// Encode encodes a value using the configured encoder
func (e *Encoder) Encode(v any) ([]byte, error) {
	e.depth = 0 // Reset depth counter
	
	res, err := e.encode(v)
	if err != nil {
		return nil, err
	}
	
	return append([]byte{Version}, res...), nil
}

// EncodeTo encodes a value directly to an io.Writer
func (e *Encoder) EncodeTo(w io.Writer, v any) error {
	data, err := e.Encode(v)
	if err != nil {
		return err
	}
	
	_, err = w.Write(data)
	return err
}

// encode is the internal encoding function with depth tracking
func (e *Encoder) encode(v any) ([]byte, error) {
	// Check max depth
	if e.MaxDepth > 0 && e.depth > e.MaxDepth {
		return nil, fmt.Errorf("bogo encode error: maximum nesting depth exceeded (%d)", e.MaxDepth)
	}
	
	// Handle null values
	if isNullValue(v) {
		return encodeNull(), nil
	}
	
	// Delegate to type-specific encoding with validation
	switch val := v.(type) {
	case string:
		if e.ValidateStrings && !isValidUTF8(val) {
			return nil, fmt.Errorf("bogo encode error: invalid UTF-8 string")
		}
		return encodeString(val)
		
	case bool:
		return encodeBool(val), nil
		
	case byte:
		return encodeByte(val)
		
	case []byte:
		return encodeBlob(val)
		
	case []string:
		if e.CompactArrays {
			return e.encodeTypedArrayWithDepth(val)
		}
		return e.encodeArrayWithDepth(val)
		
	case []int:
		if e.CompactArrays {
			return e.encodeTypedArrayWithDepth(val)
		}
		return e.encodeArrayWithDepth(val)
		
	case []int64:
		if e.CompactArrays {
			return e.encodeTypedArrayWithDepth(val)
		}
		return e.encodeArrayWithDepth(val)
		
	case []float64:
		if e.CompactArrays {
			return e.encodeTypedArrayWithDepth(val)
		}
		return e.encodeArrayWithDepth(val)
		
	case []bool:
		if e.CompactArrays {
			return e.encodeTypedArrayWithDepth(val)
		}
		return e.encodeArrayWithDepth(val)
		
	case map[string]any:
		return e.encodeObjectWithDepth(val)
		
	default:
		// Use reflection for complex types
		return e.encodeGeneric(v)
	}
}

// encodeArrayWithDepth encodes arrays with depth tracking
func (e *Encoder) encodeArrayWithDepth(v any) ([]byte, error) {
	e.depth++
	defer func() { e.depth-- }()
	
	return encodeArray(v)
}

// encodeTypedArrayWithDepth encodes typed arrays with depth tracking
func (e *Encoder) encodeTypedArrayWithDepth(v any) ([]byte, error) {
	e.depth++
	defer func() { e.depth-- }()
	
	return encodeTypedArray(v)
}

// encodeObjectWithDepth encodes objects with depth tracking
func (e *Encoder) encodeObjectWithDepth(v map[string]any) ([]byte, error) {
	// Check depth BEFORE incrementing
	if e.MaxDepth > 0 && e.depth >= e.MaxDepth {
		return nil, fmt.Errorf("bogo encode error: maximum nesting depth exceeded (%d)", e.MaxDepth)
	}
	
	e.depth++
	defer func() { e.depth-- }()
	
	// Validate object keys if in strict mode
	if e.StrictMode {
		for key := range v {
			if len(key) > 255 {
				return nil, fmt.Errorf("bogo encode error: object key too long (%d bytes, max 255)", len(key))
			}
			if e.ValidateStrings && !isValidUTF8(key) {
				return nil, fmt.Errorf("bogo encode error: invalid UTF-8 in object key")
			}
		}
	}
	
	// For now, we need to handle nested encoding manually for proper depth tracking
	// This is a simplified version - a full implementation would recursively encode each value
	return encodeObject(v)
}

// encodeGeneric handles generic types using reflection
func (e *Encoder) encodeGeneric(v any) ([]byte, error) {
	// For generic types, we need to track depth manually
	// This is a simplified implementation - in practice you'd want to implement
	// depth tracking for all reflection-based encoding
	return encode(v)
}

// isValidUTF8 checks if a string is valid UTF-8
func isValidUTF8(s string) bool {
	for _, r := range s {
		if r == 0xFFFD { // Unicode replacement character indicates invalid UTF-8
			return false
		}
	}
	return true
}

// EncodingStats provides statistics about encoding operations
type EncodingStats struct {
	BytesEncoded  int64
	MaxDepthUsed  int
	TypesEncoded  map[Type]int
	ErrorsCount   int64
}

// StatsCollector is an encoder that collects statistics
type StatsCollector struct {
	*Encoder
	Stats EncodingStats
}

// NewStatsCollector creates an encoder that collects encoding statistics
func NewStatsCollector(options ...EncoderOption) *StatsCollector {
	return &StatsCollector{
		Encoder: NewConfigurableEncoder(options...),
		Stats: EncodingStats{
			TypesEncoded: make(map[Type]int),
		},
	}
}

// Encode wraps the parent Encode with statistics collection
func (sc *StatsCollector) Encode(v any) ([]byte, error) {
	data, err := sc.Encoder.Encode(v)
	if err != nil {
		sc.Stats.ErrorsCount++
		return nil, err
	}
	
	sc.Stats.BytesEncoded += int64(len(data))
	if sc.Encoder.depth > sc.Stats.MaxDepthUsed {
		sc.Stats.MaxDepthUsed = sc.Encoder.depth
	}
	
	// Count type usage (simplified - just count the main type)
	if len(data) >= 2 {
		typeVal := Type(data[1])
		sc.Stats.TypesEncoded[typeVal]++
	}
	
	return data, nil
}

// GetStats returns a copy of the current statistics
func (sc *StatsCollector) GetStats() EncodingStats {
	stats := sc.Stats
	// Deep copy the map
	stats.TypesEncoded = make(map[Type]int)
	for k, v := range sc.Stats.TypesEncoded {
		stats.TypesEncoded[k] = v
	}
	return stats
}

// ResetStats resets all statistics
func (sc *StatsCollector) ResetStats() {
	sc.Stats = EncodingStats{
		TypesEncoded: make(map[Type]int),
	}
}