package bogo

import (
	"testing"

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

func TestNilEncodingDecoding(t *testing.T) {
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
		{"uint_1", uint8(1), TypeUint},
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
				assert.Equal(t, uint64(in), i)
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
