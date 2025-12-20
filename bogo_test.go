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
	assert.Equal(t, int(TypeString), int(data[1]))        // data type
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
			assert.Equal(t, Version, data[0]) // version
			assert.Equal(t, int(TypeNull), int(data[1])) // null type
			assert.Equal(t, 2, len(data)) // should be exactly 2 bytes
			
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
			assert.Equal(t, Version, encoded[0]) // version
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
			assert.Equal(t, Version, encoded[0]) // version
			assert.Equal(t, byte(TypeTimestamp), encoded[1]) // timestamp type
			assert.Equal(t, 10, len(encoded)) // 1 version + 1 type + 8 data bytes
			
			// Decode using main Decode function
			decoded, err := Decode(encoded)
			require.NoError(t, err)
			
			// Should get back the timestamp in milliseconds
			expectedMillis := tt.time.UnixMilli()
			assert.Equal(t, expectedMillis, decoded.(int64))
			
			// Convert back to time and compare (with millisecond precision)
			decodedTime := time.UnixMilli(decoded.(int64)).UTC()
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
			assert.Equal(t, Version, encoded[0]) // version
			assert.Equal(t, byte(TypeByte), encoded[1]) // byte type
			assert.Equal(t, 3, len(encoded)) // 1 version + 1 type + 1 data byte
			
			// Decode using main Decode function
			decoded, err := Decode(encoded)
			require.NoError(t, err)
			
			// Should get back the same byte
			assert.Equal(t, tt.data, decoded.(byte))
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
