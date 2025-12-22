package bogo

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestZeroVsNilValues(t *testing.T) {
	tests := []struct {
		name        string
		value       any
		expectNil   bool // true if we expect nil back, false if we expect zero value
		description string
	}{
		// Nil values - should encode as TypeNull and decode back as nil
		{"explicit nil", nil, true, "explicit nil value"},
		{"nil interface", (*interface{})(nil), true, "nil interface pointer"},
		{"nil map", map[string]any(nil), true, "nil map"},
		{"nil slice", []any(nil), true, "nil slice"},

		// Zero values - should encode with their types and decode back as zero values
		{"zero string", "", false, "empty string should remain empty string, not nil"},
		{"zero int", int(0), false, "zero int should remain 0, not nil"},
		{"zero int64", int64(0), false, "zero int64 should remain 0, not nil"},
		{"zero uint", uint(0), false, "zero uint should remain 0, not nil"},
		{"zero uint64", uint64(0), false, "zero uint64 should remain 0, not nil"},
		{"zero float32", float32(0), false, "zero float32 should remain 0.0, not nil"},
		{"zero float64", float64(0), false, "zero float64 should remain 0.0, not nil"},
		{"zero bool", false, false, "false bool should remain false, not nil"},
		{"zero byte", byte(0), false, "zero byte should remain 0, not nil"},
		{"empty byte slice", []byte{}, false, "empty byte slice should remain empty, not nil"},
		{"empty slice", []any{}, false, "empty slice should remain empty, not nil"},
		{"empty map", map[string]any{}, false, "empty map should remain empty, not nil"},
		{"zero time", time.Time{}, false, "zero time should remain zero time, not nil"},

		// Non-zero values - should encode and decode correctly
		{"non-zero string", "hello", false, "non-zero string"},
		{"non-zero int", int(42), false, "non-zero int"},
		{"non-zero bool true", true, false, "true bool"},
		{"non-zero float", 3.14, false, "non-zero float"},
		{"non-empty slice", []any{1, 2, 3}, false, "non-empty slice"},
		{"non-empty map", map[string]any{"key": "value"}, false, "non-empty map"},
		{"non-empty map", map[string]any{"key": nil}, false, "non-empty map with nil value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode the value
			encoded, err := Encode(tt.value)
			require.NoError(t, err, "Failed to encode %s: %v", tt.description, tt.value)

			// Decode the value
			decoded, err := Decode(encoded)
			require.NoError(t, err, "Failed to decode %s", tt.description)

			if tt.expectNil {
				assert.Nil(t, decoded, "%s should decode to nil but got %v (%T)", tt.description, decoded, decoded)
			} else {
				assert.NotNil(t, decoded, "%s should not decode to nil", tt.description)
				// For zero values, check that they're the correct type and zero value
				switch v := tt.value.(type) {
				case string:
					assert.Equal(t, v, decoded, "%s should decode to same string", tt.description)
				case int:
					// Ints decode as int64 in bogo
					assert.Equal(t, int64(v), decoded, "%s should decode to same int64", tt.description)
				case int64:
					assert.Equal(t, v, decoded, "%s should decode to same int64", tt.description)
				case uint:
					// Uints decode as uint64 in bogo
					assert.Equal(t, uint64(v), decoded, "%s should decode to same uint64", tt.description)
				case uint64:
					assert.Equal(t, v, decoded, "%s should decode to same uint64", tt.description)
				case float32:
					// Float32 decode as float64 in bogo
					assert.Equal(t, float64(v), decoded, "%s should decode to same float64", tt.description)
				case float64:
					assert.Equal(t, v, decoded, "%s should decode to same float64", tt.description)
				case bool:
					assert.Equal(t, v, decoded, "%s should decode to same bool", tt.description)
				case byte:
					assert.Equal(t, v, decoded, "%s should decode to same byte", tt.description)
				case []byte:
					assert.Equal(t, v, decoded, "%s should decode to same byte slice", tt.description)
				case []any:
					// For slices, we need to handle type conversions (int -> int64)
					decodedSlice := decoded.([]any)
					assert.Equal(t, len(v), len(decodedSlice), "%s should have same length", tt.description)
					for i, expectedItem := range v {
						actualItem := decodedSlice[i]
						// Handle int -> int64 conversion
						if expectedInt, ok := expectedItem.(int); ok {
							assert.Equal(t, int64(expectedInt), actualItem, "%s slice element %d should convert int to int64", tt.description, i)
						} else {
							assert.Equal(t, expectedItem, actualItem, "%s slice element %d should match", tt.description, i)
						}
					}
				case map[string]any:
					assert.Equal(t, v, decoded, "%s should decode to same map", tt.description)
				case time.Time:
					assert.Equal(t, v, decoded, "%s should decode to same time", tt.description)
				default:
					// For other types, just check equality
					assert.Equal(t, tt.value, decoded, "%s should decode to same value", tt.description)
				}
			}
		})
	}
}

func TestZeroVsNilInObjects(t *testing.T) {
	tests := []struct {
		name   string
		object map[string]any
	}{
		{
			"mixed zero and nil values",
			map[string]any{
				"nil_value":        nil,
				"zero_string":      "",
				"zero_int":         0,
				"zero_bool":        false,
				"zero_float":       0.0,
				"empty_slice":      []any{},
				"empty_map":        map[string]any{},
				"empty_byte_slice": []byte{},
				"non_zero_string":  "hello",
				"non_zero_int":     42,
				"non_zero_bool":    true,
			},
		},
		{
			"all nil values",
			map[string]any{
				"nil1": nil,
				"nil2": nil,
				"nil3": nil,
			},
		},
		{
			"all zero values",
			map[string]any{
				"zero_string": "",
				"zero_int":    0,
				"zero_bool":   false,
				"zero_float":  0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode the object
			encoded, err := Encode(tt.object)
			require.NoError(t, err, "Failed to encode object")

			// Decode the object
			decoded, err := Decode(encoded)
			require.NoError(t, err, "Failed to decode object")

			decodedMap := decoded.(map[string]any)
			assert.Equal(t, len(tt.object), len(decodedMap), "Decoded object should have same number of fields")

			// Check each field
			for key, expectedValue := range tt.object {
				actualValue, exists := decodedMap[key]
				assert.True(t, exists, "Key %s should exist in decoded object", key)

				if expectedValue == nil {
					assert.Nil(t, actualValue, "Nil value for key %s should remain nil", key)
				} else {
					assert.NotNil(t, actualValue, "Non-nil value for key %s should not become nil", key)

					// Type-specific comparisons (accounting for bogo type conversions)
					switch v := expectedValue.(type) {
					case int:
						assert.Equal(t, int64(v), actualValue, "Int value for key %s should convert to int64", key)
					case float32:
						assert.Equal(t, float64(v), actualValue, "Float32 value for key %s should convert to float64", key)
					case uint:
						assert.Equal(t, uint64(v), actualValue, "Uint value for key %s should convert to uint64", key)
					default:
						assert.Equal(t, expectedValue, actualValue, "Value for key %s should match", key)
					}
				}
			}
		})
	}
}

func TestNilTypedValues(t *testing.T) {
	// Test nil values for different pointer types
	tests := []struct {
		name  string
		value any
	}{
		{"nil string pointer", (*string)(nil)},
		{"nil int pointer", (*int)(nil)},
		{"nil bool pointer", (*bool)(nil)},
		{"nil float64 pointer", (*float64)(nil)},
		{"nil byte slice", (*[]byte)(nil)},
		{"nil slice", (*[]any)(nil)},
		{"nil map", (*map[string]any)(nil)},
		{"nil time pointer", (*time.Time)(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode the nil pointer
			encoded, err := Encode(tt.value)
			require.NoError(t, err, "Failed to encode %s", tt.name)

			// Decode the value
			decoded, err := Decode(encoded)
			require.NoError(t, err, "Failed to decode %s", tt.name)

			// Should decode to nil
			assert.Nil(t, decoded, "%s should decode to nil", tt.name)
		})
	}
}

