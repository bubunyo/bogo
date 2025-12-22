package bogo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleStruct(t *testing.T) {
	type SimpleStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	original := SimpleStruct{
		Name: "John",
		Age:  25,
	}

	// Test encoding using new encoder
	data, err := Marshal(original)
	require.NoError(t, err)

	// Test decoding
	var decoded SimpleStruct
	err = Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Name, decoded.Name)
	assert.Equal(t, original.Age, decoded.Age)
}
