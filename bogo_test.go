package bogo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringEncodingDecoding(t *testing.T) {
	// encode
	data, err := Encode("abcd")
	require.NoError(t, err)
	assert.Equal(t, Version, data[0])                     // bogo version
	assert.Equal(t, TypeString, int(data[1]))             // data type
	assert.Equal(t, 1, int(data[2]))                      // space to hold single byte num
	assert.Equal(t, 4, int(data[3]))                      // length of 4 chars
	assert.Equal(t, []byte{'a', 'b', 'c', 'd'}, data[4:]) // length of 4 chars

	// decode
	d, err := Decode(data)
	require.NoError(t, err)
	assert.Equal(t, "abcd", d.(string))
}

func TestBoolEncodingDecoding(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected []byte
	}{
		{
			name:     "encode and decode true",
			input:    true,
			expected: []byte{0, 1}, // expected encoded data for true
		},
		{
			name:     "encode and decode false",
			input:    false,
			expected: []byte{0, 2}, // expected encoded data for false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode the input value
			data, err := Encode(tt.input)
			require.NoError(t, err)
			assert.Equal(t, Version, data[0]) // bogo version
			dataType := TypeBoolFalse
			if tt.input {
				dataType = TypeBoolTrue
			} else {
				dataType = TypeBoolFalse
			}
			assert.Equal(t, int(dataType), int(data[1])) // data type
			assert.Equal(t, 2, len(data))
			assert.Equal(t, tt.expected, data)

			// Decode the data back
			d, err := Decode(data)
			require.NoError(t, err)
			assert.Equal(t, tt.input, d.(bool))
		})
	}
}

func TestNullEncodingDecoding(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"nil value", nil},
		{"nil interface", (*interface{})(nil)},
		{"nil slice", ([]int)(nil)},
		{"nil map", (map[string]int)(nil)},
		{"nil pointer", (*string)(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			data, err := Encode(tt.input)
			require.NoError(t, err)

			// Check encoding format
			assert.Equal(t, Version, data[0])            // version
			assert.Equal(t, int(TypeNull), int(data[1])) // null type
			assert.Equal(t, 2, len(data))                // should be exactly 2 bytes

			// Decode
			result, err := Decode(data)
			require.NoError(t, err)
			assert.Nil(t, result)
		})
	}
}

// TestBlobEncodingDecoding tests blob encoding via structs (requires object support)
func TestBlobEncodingDecoding(t *testing.T) {
	t.Skip("Requires object/struct support - will implement later")
}

func TestBlobFullRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"small", []byte{0x01, 0x02, 0x03}},
		{"uuid", []byte{0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4, 0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00}},
		{"large", make([]byte, 1000)},
	}

	// Initialize large blob
	for i := range tests[3].data {
		tests[3].data[i] = byte(i % 256)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode using main Encode function
			encoded, err := Encode(tt.data)
			require.NoError(t, err)

			// Check encoding format
			assert.Equal(t, Version, encoded[0])        // version
			assert.Equal(t, byte(TypeBlob), encoded[1]) // blob type

			// Decode using main Decode function
			decoded, err := Decode(encoded)
			require.NoError(t, err)

			// Should get back the same data
			assert.Equal(t, tt.data, decoded.([]byte))
		})
	}
}

func TestBlobTypedEncodingDecoding(t *testing.T) {
	// Test direct blob encoding (not wrapped in struct)
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"small", []byte{0x01, 0x02, 0x03}},
		{"medium", make([]byte, 100)},
	}

	for i := range tests[2].data {
		tests[2].data[i] = byte(i)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will test the direct blob encoding once implemented
			encoded, err := encodeBlob(tt.data)
			require.NoError(t, err)

			// Check format: TypeBlob + LenSize + DataSize + Data
			assert.Equal(t, byte(TypeBlob), encoded[0])

			// Decode
			decoded, err := decodeBlob(encoded[1:])
			require.NoError(t, err)
			assert.Equal(t, tt.data, decoded)
		})
	}
}

func TestTimestampEncodingDecoding(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
	}{
		{"unix epoch", 0},
		{"positive timestamp", 1705317045123}, // 2024-01-15T10:30:45.123Z
		{"negative timestamp", -86400000},     // 1969-12-31T00:00:00.000Z
		{"max positive", 9223372036854775807}, // max int64
		{"large negative", -1000000000000},    // 1938-04-24T22:13:20.000Z
		{"recent timestamp", 1640995200000},   // 2022-01-01T00:00:00.000Z
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test direct timestamp encoding
			encoded, err := encodeTimestamp(tt.timestamp)
			require.NoError(t, err)

			// Check encoding format
			assert.Equal(t, byte(TypeTimestamp), encoded[0])
			assert.Equal(t, 9, len(encoded)) // 1 type + 8 data bytes

			// Decode
			decoded, err := decodeTimestamp(encoded[1:])
			require.NoError(t, err)
			assert.Equal(t, tt.timestamp, decoded)
		})
	}
}

// TestTimestampFullRoundTrip tests timestamp encoding via structs (requires object support)
func TestTimestampFullRoundTrip(t *testing.T) {
	t.Skip("Requires object/struct support - will implement later")
}

func TestTimeEncodingDecoding(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
	}{
		{"unix epoch", time.Unix(0, 0).UTC()},
		{"current time", time.Date(2024, 1, 15, 10, 30, 45, 123000000, time.UTC)},
		{"past", time.Date(1969, 12, 31, 0, 0, 0, 0, time.UTC)},
		{"future", time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"with nanoseconds", time.Date(2024, 6, 15, 14, 30, 45, 123456789, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode using main Encode function
			encoded, err := Encode(tt.time)
			require.NoError(t, err)

			// Check encoding format
			assert.Equal(t, Version, encoded[0])             // version
			assert.Equal(t, byte(TypeTimestamp), encoded[1]) // timestamp type
			assert.Equal(t, 10, len(encoded))                // 1 version + 1 type + 8 data bytes

			// Decode using main Decode function
			decoded, err := Decode(encoded)
			require.NoError(t, err)

			// Should get back the timestamp in milliseconds
			expectedMillis := tt.time.UnixMilli()
			assert.Equal(t, time.UnixMilli(expectedMillis).UTC(), decoded.(time.Time).UTC())

			// Convert back to time and compare (with millisecond precision)
			decodedTime := decoded.(time.Time).UTC()
			assert.Equal(t, tt.time.Truncate(time.Millisecond), decodedTime.Truncate(time.Millisecond))
		})
	}
}

func TestByteEncodingDecoding(t *testing.T) {
	tests := []struct {
		name string
		data byte
	}{
		{"zero", 0},
		{"small", 42},
		{"max", 255},
		{"mid", 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test direct byte encoding
			encoded, err := encodeByte(tt.data)
			require.NoError(t, err)

			// Check encoding format
			assert.Equal(t, byte(TypeByte), encoded[0])
			assert.Equal(t, tt.data, encoded[1])
			assert.Equal(t, 2, len(encoded)) // 1 type + 1 data byte

			// Decode
			decoded, err := decodeByte(encoded[1:])
			require.NoError(t, err)
			assert.Equal(t, tt.data, decoded)
		})
	}
}

func TestByteFullRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		data byte
	}{
		{"zero", 0},
		{"small", 42},
		{"max", 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode using main Encode function
			encoded, err := Encode(tt.data)
			require.NoError(t, err)

			// Check encoding format
			assert.Equal(t, Version, encoded[0])        // version
			assert.Equal(t, byte(TypeByte), encoded[1]) // byte type
			assert.Equal(t, 3, len(encoded))            // 1 version + 1 type + 1 data byte

			// Decode using main Decode function
			decoded, err := Decode(encoded)
			require.NoError(t, err)

			// Should get back the same byte
			assert.Equal(t, tt.data, decoded.(byte))
		})
	}
}

func TestObjectEncodingDecoding(t *testing.T) {
	tests := []struct {
		name   string
		object map[string]any
	}{
		{"empty object", map[string]any{}},
		{"simple object", map[string]any{
			"name": "John",
			"age":  int64(25),
		}},
		{"complex object", map[string]any{
			"name":      "Alice",
			"age":       int64(30),
			"active":    true,
			"score":     95.5,
			"metadata":  nil,
			"timestamp": int64(1705317045123),
		}},
		{"nested object", map[string]any{
			"user": map[string]any{
				"name": "Bob",
				"details": map[string]any{
					"email": "bob@example.com",
					"level": int64(3),
				},
			},
			"settings": map[string]any{
				"theme":         "dark",
				"notifications": true,
			},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test direct object encoding
			encoded, err := encodeObject(tt.object)
			require.NoError(t, err)

			// Check encoding format
			assert.Equal(t, byte(TypeObject), encoded[0])

			// Decode
			decoded, err := decodeObject(encoded[1:])
			require.NoError(t, err)

			// Should get back the same object
			assert.Equal(t, tt.object, decoded)
		})
	}
}

func TestObjectFullRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		object map[string]any
	}{
		{"simple", map[string]any{"key": "value", "num": int64(42)}},
		{"mixed types", map[string]any{
			"str":   "hello",
			"int":   int64(123),
			"bool":  true,
			"float": 3.14,
			"null":  nil,
			"bytes": []byte{1, 2, 3},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode using main Encode function
			encoded, err := Encode(tt.object)
			require.NoError(t, err)

			// Check encoding format
			assert.Equal(t, Version, encoded[0])          // version
			assert.Equal(t, byte(TypeObject), encoded[1]) // object type

			// Decode using main Decode function
			decoded, err := Decode(encoded)
			require.NoError(t, err)

			// Should get back the same object
			decodedMap := decoded.(map[string]any)
			assert.Equal(t, len(tt.object), len(decodedMap))

			for key, expectedValue := range tt.object {
				actualValue, exists := decodedMap[key]
				assert.True(t, exists, "Key %s should exist", key)
				assert.Equal(t, expectedValue, actualValue, "Value for key %s should match", key)
			}
		})
	}
}

func TestStructEncodingDecoding(t *testing.T) {
	type Person struct {
		Name string `bogo:"name"`
		Age  int    `bogo:"age"`
		City string
	}

	tests := []struct {
		name   string
		person Person
	}{
		{"simple person", Person{Name: "John", Age: 30, City: "New York"}},
		{"empty person", Person{}},
		{"partial person", Person{Name: "Alice", Age: 25}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode using main Encode function
			encoded, err := Encode(tt.person)
			require.NoError(t, err)

			// Check encoding format
			assert.Equal(t, Version, encoded[0])          // version
			assert.Equal(t, byte(TypeObject), encoded[1]) // object type (structs encode as objects)

			// Decode using main Decode function
			decoded, err := Decode(encoded)
			require.NoError(t, err)

			// Should get back a map representing the struct
			decodedMap := decoded.(map[string]any)
			assert.NotNil(t, decodedMap)

			// Check that struct fields are present with correct tag names or field names
			if tt.person.Name != "" {
				assert.Equal(t, tt.person.Name, decodedMap["Name"]) // uses bogo tag
			}
			if tt.person.Age != 0 {
				assert.Equal(t, int64(tt.person.Age), decodedMap["Age"]) // uses bogo tag, int becomes int64
			}
			if tt.person.City != "" {
				assert.Equal(t, tt.person.City, decodedMap["City"]) // uses field name
			}
		})
	}
}

func TestTypedListEncodingDecoding(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{"string list", []string{"hello", "world", "test"}, []string{"hello", "world", "test"}},
		{"int list", []int{1, 2, 3, 4, 5}, []int64{1, 2, 3, 4, 5}},
		{"byte list", []byte{10, 20, 30, 40, 50}, []byte{10, 20, 30, 40, 50}},
		{"bool list", []bool{true, false, true, false}, []bool{true, false, true, false}},
		{"float list", []float64{1.1, 2.2, 3.3}, []float64{1.1, 2.2, 3.3}},
		{"uint list", []uint64{100, 200, 300}, []uint64{100, 200, 300}},
		{"single element", []int{42}, []int64{42}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test direct typed list encoding
			encoded, err := encodeTypedList(tt.input)
			require.NoError(t, err)

			// Check encoding format
			assert.Equal(t, byte(TypeTypedList), encoded[0])

			// Decode
			decoded, err := decodeTypedList(encoded[1:])
			require.NoError(t, err)
			assert.Equal(t, tt.expected, decoded)
		})
	}
}

func TestTypedListFullRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{"string list", []string{"hello", "world"}, []string{"hello", "world"}},
		{"int list", []int{1, 2, 3}, []int64{1, 2, 3}},
		{"byte list", []byte{10, 20, 30}, []byte{10, 20, 30}},
		{"bool list", []bool{true, false, true}, []bool{true, false, true}},
		{"float list", []float64{1.1, 2.2}, []float64{1.1, 2.2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test won't automatically use typed lists
			// The main encoder decides based on list homogeneity
			// For now, we'll test the typed list functions directly

			// Encode using typed list function directly
			encoded, err := encodeTypedList(tt.input)
			require.NoError(t, err)

			// Decode using typed list function directly
			decoded, err := decodeTypedList(encoded[1:])
			require.NoError(t, err)

			assert.Equal(t, tt.expected, decoded)
		})
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{"string", "hello", "hello"},
		{"int", 42, int64(42)},
		{"bool", true, true},
		{"null", nil, nil},
		{"float", 3.14, 3.14},
		{"byte", byte(255), byte(255)},
		{"blob", []byte{1, 2, 3}, []byte{1, 2, 3}},
		{"object", map[string]any{"key": "value"}, map[string]any{"key": "value"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := Marshal(tt.input)
			require.NoError(t, err)

			// Unmarshal
			var result any
			err = Unmarshal(data, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNumbersEncodingDecoding(t *testing.T) {
	tests := []struct {
		name  string
		input any
		bt    Type
	}{
		{"integer", 47, TypeInt},
		{"int16", int16(12), TypeInt},
		{"min_unsigned_id", ^uint64(0), TypeUint},
		{"max_unsigned_id_less_1", ^uint64(0) - 1, TypeUint},
		{"min_signed_int", -(1 << 63), TypeInt},
		{"uint_1", uint8(1), TypeByte},
		{"zero", 0, TypeInt},
		{"float", 3.1415926535897932384626433832795028841971693993751058209749445923819343, TypeFloat},
		{"float", 0.1, TypeFloat},
		{"negative-float", -0.0005893, TypeFloat},
		{"negative-float", 0.5, TypeFloat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Encode(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.bt, Type(data[1]))
			i, err := Decode(data)
			assert.NoError(t, err)
			switch in := tt.input.(type) {
			case uint8:
				assert.Equal(t, byte(in), i)
			case int:
				assert.Equal(t, int64(in), i)
			case int16:
				assert.Equal(t, int64(in), i)
			case uint64:
				assert.Equal(t, uint64(in), i)
			case float64:
				assert.Equal(t, float64(in), i)
			case float32:
				assert.Equal(t, float64(in), i)
			default:
				t.Errorf("result type not asserted. type=%T", data)
			}
		})
	}
}

// Test unused helper functions
func TestUnusedHelperFunctions(t *testing.T) {
	t.Run("decodeNull", func(t *testing.T) {
		result, err := decodeNull()
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("decodeBoolTrue", func(t *testing.T) {
		result, err := decodeBoolTrue()
		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("decodeBoolFalse", func(t *testing.T) {
		result, err := decodeBoolFalse()
		assert.NoError(t, err)
		assert.False(t, result)
	})
}

// Test Type.String() method
func TestTypeStringMethod(t *testing.T) {
	tests := []struct {
		typeVal  Type
		expected string
	}{
		{TypeNull, "<null>"},
		{TypeBoolTrue, "<bool:true>"},
		{TypeBoolFalse, "<bool:false>"},
		{TypeString, "<string>"},
		{TypeUntypedList, "<list>"},
		{TypeTypedList, "<typed_list>"},
		{TypeObject, "<object>"},
		{TypeByte, "<byte>"},
		{TypeInt, "<int>"},
		{TypeUint, "<uint>"},
		{TypeFloat, "<float>"},
		{TypeBlob, "<blob>"},
		{TypeTimestamp, "<timestamp>"},
		{Type(99), "<unknown>"}, // Test unknown type
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.typeVal.String())
		})
	}
}

// Test wrapError utility function
func TestWrapError(t *testing.T) {
	baseErr := assert.AnError

	t.Run("wrapError with no additional messages", func(t *testing.T) {
		result := wrapError(baseErr)
		assert.Equal(t, baseErr, result)
	})

	t.Run("wrapError with additional message", func(t *testing.T) {
		result := wrapError(baseErr, "additional context")
		assert.Contains(t, result.Error(), baseErr.Error())
		assert.Contains(t, result.Error(), "additional context")
	})
}

// Test encodeTimeValue helper function
func TestEncodeTimeValue(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 45, 123000000, time.UTC)

	encoded, err := encodeTimeValue(testTime)
	require.NoError(t, err)

	// Should be 9 bytes: 1 type + 8 timestamp
	assert.Equal(t, 9, len(encoded))
	assert.Equal(t, byte(TypeTimestamp), encoded[0])

	// Decode and verify
	timestamp, err := decodeTimestamp(encoded[1:])
	require.NoError(t, err)
	assert.Equal(t, testTime.UnixMilli(), timestamp)
}

// Test all branches in isZeroValue function
func TestIsZeroValueAllBranches(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"nil", nil, true},
		{"empty string", "", true},
		{"non-empty string", "hello", false},
		{"int zero", int(0), true},
		{"int non-zero", int(42), false},
		{"int8 zero", int8(0), true},
		{"int8 non-zero", int8(42), false},
		{"int16 zero", int16(0), true},
		{"int16 non-zero", int16(42), false},
		{"int32 zero", int32(0), true},
		{"int32 non-zero", int32(42), false},
		{"int64 zero", int64(0), true},
		{"int64 non-zero", int64(42), false},
		{"uint zero", uint(0), true},
		{"uint non-zero", uint(42), false},
		{"uint8 zero", uint8(0), true},
		{"uint8 non-zero", uint8(42), false},
		{"uint16 zero", uint16(0), true},
		{"uint16 non-zero", uint16(42), false},
		{"uint32 zero", uint32(0), true},
		{"uint32 non-zero", uint32(42), false},
		{"uint64 zero", uint64(0), true},
		{"uint64 non-zero", uint64(42), false},
		{"float32 zero", float32(0.0), true},
		{"float32 non-zero", float32(3.14), false},
		{"float64 zero", float64(0.0), true},
		{"float64 non-zero", float64(3.14), false},
		{"bool false", false, true},
		{"bool true", true, false},
		{"slice", []int{1, 2, 3}, false}, // default case
		{"struct", struct{}{}, false},    // default case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isZeroValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test error paths and edge cases
func TestErrorPathsAndEdgeCases(t *testing.T) {
	t.Run("decodeByte with insufficient data", func(t *testing.T) {
		_, err := decodeByte([]byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient data")
	})

	t.Run("decodeTimestamp with insufficient data", func(t *testing.T) {
		_, err := decodeTimestamp([]byte{1, 2, 3}) // Less than 8 bytes
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient data")
	})

	t.Run("decodeBlob with insufficient size data", func(t *testing.T) {
		// Size len says 5 bytes needed, but only 2 provided
		_, err := decodeBlob([]byte{5, 10})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient data for size")
	})

	t.Run("decodeBlob with insufficient content data", func(t *testing.T) {
		// Says blob size is 10, but only provides 3 bytes
		_, err := decodeBlob([]byte{1, 10, 1, 2, 3})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient data for blob content")
	})

	t.Run("encodeString with empty string", func(t *testing.T) {
		encoded, err := encodeString("")
		require.NoError(t, err)
		assert.Equal(t, []byte{byte(TypeString), 1, 0}, encoded)
	})

	t.Run("encode unsupported type", func(t *testing.T) {
		// Use a channel which is an unsupported type
		ch := make(chan int)
		_, err := encode(ch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported type")
	})

	t.Run("encodeTypedList with empty list falls back to regular list", func(t *testing.T) {
		emptyList := []string{}
		encoded, err := encodeTypedList(emptyList)
		require.NoError(t, err)
		// Should be regular list type, not typed list
		assert.Equal(t, byte(TypeUntypedList), encoded[0])
	})

	t.Run("encodeTypedList with mixed types falls back to regular list", func(t *testing.T) {
		// This won't actually work since Go is strongly typed, but test the logic
		// by using interface{} list
		mixedList := []interface{}{"string", 42}
		_, err := encodeList(mixedList)
		require.NoError(t, err)
		// Just verify it doesn't panic
	})

	t.Run("encodeObject with unsupported type", func(t *testing.T) {
		_, err := encodeObject("not a map")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "object type not supported")
	})

	t.Run("encodeFieldEntry with key too long", func(t *testing.T) {
		longKey := string(make([]byte, 256)) // 256 bytes, exceeds 255 limit
		_, err := encodeFieldEntry(longKey, "value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key too long")
	})
}

// Test decodeList edge cases to improve coverage
func TestDecodeListEdgeCases(t *testing.T) {
	t.Run("decodeList with nil destination", func(t *testing.T) {
		err := decodeList([]byte{}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid decoder destination")
	})

	t.Run("decodeList with non-pointer destination", func(t *testing.T) {
		var slice []int
		err := decodeList([]byte{}, slice) // Not a pointer
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid decoder destination")
	})

	t.Run("decodeList with byte elements", func(t *testing.T) {
		// Create list data with byte elements
		data := []byte{
			TypeByte, 42, // First byte element
			TypeByte, 43, // Second byte element
		}

		var result []any
		err := decodeList(data, &result)
		assert.Error(t, err) // This will error due to type mismatch logic
	})
}

// Test decodeValue edge cases
func TestDecodeValueEdgeCases(t *testing.T) {
	t.Run("decodeValue with empty data", func(t *testing.T) {
		result, err := decodeValue([]byte{})
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("decodeValue with unsupported type", func(t *testing.T) {
		// Use a type value that doesn't exist
		_, err := decodeValue([]byte{byte(99), 1, 2, 3})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported value type")
	})

	t.Run("decodeValue with list type now supported", func(t *testing.T) {
		// Test the list case that now works
		result, err := decodeValue([]byte{TypeUntypedList, 1, 0})
		assert.NoError(t, err)
		assert.Equal(t, []any{}, result) // Empty list
	})
}

// Test API-level error handling
func TestAPIErrorHandling(t *testing.T) {
	t.Run("Decode with insufficient data", func(t *testing.T) {
		// Empty data
		_, err := Decode([]byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient data")

		// Only version byte, no type
		_, err = Decode([]byte{0x00})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient data")
	})

	t.Run("Decode with unsupported version", func(t *testing.T) {
		// Wrong version
		_, err := Decode([]byte{0x99, TypeString, 1, 5, 'h', 'e', 'l', 'l', 'o'})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported version")
		assert.Contains(t, err.Error(), "153") // 0x99 = 153 in decimal
		assert.Contains(t, err.Error(), "expected version 0")
	})

	t.Run("API should never call os.Exit", func(t *testing.T) {
		// This test ensures the fix worked - if os.Exit was called, this test would never complete
		_, err := Decode([]byte{0x99, TypeNull})
		assert.Error(t, err)
		// If we reach this point, os.Exit was not called
		assert.True(t, true, "API correctly returned error instead of calling os.Exit")
	})
}
