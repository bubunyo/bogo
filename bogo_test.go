package bogo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
