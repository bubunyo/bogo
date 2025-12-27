package bogo

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"time"
)

// Encoder provides structured encoding with configurable options
type Encoder struct {
	// Configuration options
	MaxDepth        int    // Maximum nesting depth for objects/lists (0 = unlimited)
	StrictMode      bool   // Strict type checking and validation
	CompactLists   bool   // Use typed lists when beneficial
	ValidateStrings bool   // Validate UTF-8 encoding in strings
	TagName         string // Struct tag name to use (default: "json" for compatibility)

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
		CompactLists:   true,
		ValidateStrings: true,
		TagName:         "json", // Default to json tag for compatibility
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

func WithCompactLists(compact bool) EncoderOption {
	return func(e *Encoder) {
		e.CompactLists = compact
	}
}

func WithStringValidation(validate bool) EncoderOption {
	return func(e *Encoder) {
		e.ValidateStrings = validate
	}
}

func WithStructTag(tagName string) EncoderOption {
	return func(e *Encoder) {
		e.TagName = tagName
	}
}

// Encode encodes a value using the configured encoder
func (e *Encoder) Encode(v any) ([]byte, error) {
	e.depth = 0 // Reset depth counter

	res, err := e.encode(v)
	if err != nil {
		return nil, err
	}

	// todo: can i optimize by definiting the length of the slice before i copy into it?
	// can i estimate the space occupied by a decoded slice by looking at the current
	// memory occupied by the current data? If i can i can allocate a list with a max capacity
	// ensuring the the decoded data is always going to be smaller and based on the used length, we report that only
	// this might not work due to the fact that inner types are decoded first and we might know
	//	their position is the backing list. but still work a short. the risk is that, we might need to
	// padd, tradding size for speed
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

	case time.Time:
		return encodeTimestamp(val.UnixMilli())

	case []string:
		if e.CompactLists {
			return e.encodeTypedListWithDepth(val)
		}
		return e.encodeListWithDepth(val)

	case []int:
		if e.CompactLists {
			return e.encodeTypedListWithDepth(val)
		}
		return e.encodeListWithDepth(val)

	case []int64:
		if e.CompactLists {
			return e.encodeTypedListWithDepth(val)
		}
		return e.encodeListWithDepth(val)

	case []float64:
		if e.CompactLists {
			return e.encodeTypedListWithDepth(val)
		}
		return e.encodeListWithDepth(val)

	case []bool:
		if e.CompactLists {
			return e.encodeTypedListWithDepth(val)
		}
		return e.encodeListWithDepth(val)

	case map[string]any:
		return e.encodeObjectWithDepth(val)

	default:
		// Use reflection for complex types (including structs)
		return e.encodeReflected(v)
	}
}

// encodeListWithDepth encodes lists with depth tracking
func (e *Encoder) encodeListWithDepth(v any) ([]byte, error) {
	e.depth++
	defer func() { e.depth-- }()

	return encodeList(v)
}

// encodeTypedListWithDepth encodes typed lists with depth tracking
func (e *Encoder) encodeTypedListWithDepth(v any) ([]byte, error) {
	e.depth++
	defer func() { e.depth-- }()

	return encodeTypedList(v)
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

	// Encode the object with proper depth tracking
	return e.encodeMapWithDepth(v)
}

// encodeMapWithDepth encodes a map with proper depth tracking and using the encoder
func (e *Encoder) encodeMapWithDepth(obj map[string]any) ([]byte, error) {
	fieldsBuf := &bytes.Buffer{}

	// Encode each key-value pair as field entries
	for key, value := range obj {
		fieldEntry, err := e.encodeFieldEntryWithDepth(key, value)
		if err != nil {
			return nil, fmt.Errorf("bogo encode error: failed to encode field %s: %w", key, err)
		}
		fieldsBuf.Write(fieldEntry)
	}

	fieldsData := fieldsBuf.Bytes()
	fieldsSize := len(fieldsData)

	// Encode the total size of all fields
	encodedSizeData, err := encodeUint(uint64(fieldsSize))
	if err != nil {
		return nil, fmt.Errorf("bogo encode error: failed to encode fields size: %w", err)
	}

	// Build final object: TypeObject + LenSize + DataSize + FieldData
	result := &bytes.Buffer{}
	result.WriteByte(TypeObject)
	result.Write(encodedSizeData[1:]) // remove type byte from size encoding
	result.Write(fieldsData)

	return result.Bytes(), nil
}

// encodeFieldEntryWithDepth encodes a field entry using the encoder for depth tracking
func (e *Encoder) encodeFieldEntryWithDepth(key string, value any) ([]byte, error) {
	// Encode the value first to know its size using the encoder
	encodedValue, err := e.encode(value)
	if err != nil {
		return nil, err
	}

	keyBytes := []byte(key)
	keyLen := len(keyBytes)

	if keyLen > 255 {
		return nil, fmt.Errorf("key too long, maximum 255 bytes")
	}

	// Calculate entry size: keyLen(1) + key + value
	entrySize := 1 + keyLen + len(encodedValue)

	// Encode entry size
	encodedEntrySize, err := encodeUint(uint64(entrySize))
	if err != nil {
		return nil, err
	}

	// Build field entry: LenSize + EntrySize + KeyLength + Key + Value
	entry := &bytes.Buffer{}
	entry.Write(encodedEntrySize[1:]) // remove type byte
	entry.WriteByte(byte(keyLen))
	entry.Write(keyBytes)
	entry.Write(encodedValue)

	return entry.Bytes(), nil
}

// encodeReflected handles reflection-based encoding for structs and other complex types
func (e *Encoder) encodeReflected(v any) ([]byte, error) {
	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)

	// Handle pointers
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return encodeNull(), nil
		}
		rv = rv.Elem()
		rt = rt.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return e.encodeStruct(rv, rt)
	case reflect.Slice, reflect.Array:
		return e.encodeReflectedList(rv)
	case reflect.Map:
		return e.encodeReflectedMap(rv)
	case reflect.Interface:
		// Handle interface{} by encoding the underlying value
		if !rv.IsNil() {
			return e.encode(rv.Interface())
		}
		return encodeNull(), nil
	default:
		// Fall back to basic type encoding for other types
		return encode(v)
	}
}

// encodeStruct converts a struct to a map[string]any and encodes it
func (e *Encoder) encodeStruct(rv reflect.Value, rt reflect.Type) ([]byte, error) {
	obj := make(map[string]any)

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name from tag or use field name
		fieldName := e.getFieldName(field)

		// Skip if tag indicates to omit the field
		if fieldName == "-" {
			continue
		}

		// Skip zero values if omitempty is specified
		if e.shouldOmitEmpty(field) && e.isZeroValue(fieldValue) {
			continue
		}

		// Recursively encode the field value
		obj[fieldName] = fieldValue.Interface()
	}

	return e.encodeObjectWithDepth(obj)
}

// getFieldName returns the field name to use based on struct tags
func (e *Encoder) getFieldName(field reflect.StructField) string {
	tag := field.Tag.Get(e.TagName)
	if tag == "" {
		return field.Name
	}

	// Handle "fieldname" and "fieldname,omitempty" formats
	if commaIdx := len(tag); commaIdx > 0 {
		for i, c := range tag {
			if c == ',' {
				commaIdx = i
				break
			}
		}
		return tag[:commaIdx]
	}

	return tag
}

// shouldOmitEmpty checks if the field has omitempty tag
func (e *Encoder) shouldOmitEmpty(field reflect.StructField) bool {
	tag := field.Tag.Get(e.TagName)
	return len(tag) > 0 && (tag == "omitempty" || len(tag) > 10 && tag[len(tag)-10:] == ",omitempty")
}

// isZeroValue reports whether v is the zero value for its type
func (e *Encoder) isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// encodeReflectedList handles slice/list encoding via reflection
func (e *Encoder) encodeReflectedList(rv reflect.Value) ([]byte, error) {
	length := rv.Len()
	arr := make([]any, length)

	for i := range length {
		arr[i] = rv.Index(i).Interface()
	}

	return e.encodeListWithDepth(arr)
}

// encodeReflectedMap handles map encoding via reflection
func (e *Encoder) encodeReflectedMap(rv reflect.Value) ([]byte, error) {
	obj := make(map[string]any)

	for _, key := range rv.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		obj[keyStr] = rv.MapIndex(key).Interface()
	}

	return e.encodeObjectWithDepth(obj)
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
	BytesEncoded int64
	MaxDepthUsed int
	TypesEncoded map[Type]int
	ErrorsCount  int64
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
