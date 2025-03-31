package bogo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicDataTypesEncodingDecoding(t *testing.T) {
	tests := []struct {
		name  string
		input any
		bt    Type
	}{
		{"integer", 47, TypeInt},
		// {"float", 0.0000000000000000000000000000000000000000000000000000000000000000000005, TypeFloat},
		// {"large integer", 1844674407370955161, TypeInt},
		// {"bool", true, TypeBool},
		// {"null", nil, TypeNull},
		// {"string", "Hello, World", TypeString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Encode(tt.input)
			require.NoError(t, err)
			fmt.Println("data:", data)
			assert.Equal(t, tt.bt, Type(data[1]))
			i, err := Decode(data)
			assert.Equal(t, tt.input, i)
		})
	}

}

func Test_EncodeNum(t *testing.T) {
	tests := []struct {
		name  string
		input any
		bt    Type
	}{
		{"integer", 47, TypeInt},
		{"max_unsigned_id", ^uint64(0), TypeInt},
		{"max_unsigned_id_less_1", ^uint64(0) - 1, TypeInt},
		{"min_signed_int", -(1 << 63), TypeUInt},
		{"min_signed_int", uint8(1), TypeUInt},
		{"zero", 0, TypeInt},
		{"float", 0.0000000000000000000000000000000000000000000000000000000000000000000005, TypeFloat},
		// {"large integer", 1844674407370955161, TypeInt},
		// {"bool", true, TypeBool},
		// {"null", nil, TypeNull},
		// {"string", "Hello, World", TypeString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := encodeNum(tt.input)
			fmt.Println(">>", data)
		})
	}

}
