package bogo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArrayEncoding(t *testing.T) {
	expected := []any{1, true, nil, "yes", 0.5, uint16(120), []any{7, 8, 9}, -98.0005}
	res, err := encodeArray(expected)
	require.NoError(t, err)

	assert.Equal(t, TypeArray, int(res[0]))
	arrLen := uint64(res[1])
	size, err := decodeUint(res[2 : 2+arrLen])
	require.NoError(t, err)
	assert.Equal(t, size, uint64(len(res[2+arrLen:])))

	var actual []any
	err = decodeArray(res[3:], &actual)
	require.NoError(t, err)

	assert.Equal(t, int64(expected[0].(int)), actual[0])
	assert.Equal(t, expected[1], actual[1])
	assert.Equal(t, expected[2], actual[2])
	assert.Equal(t, expected[3], actual[3])
	assert.Equal(t, expected[4], actual[4])
	assert.Equal(t, uint64(expected[5].(uint16)), actual[5])

	expInner := expected[6].([]interface{})
	actInner := actual[6].([]interface{})
	for i := range expInner {
		assert.Equal(t, int64(expInner[i].(int)), actInner[i])
	}

	assert.Equal(t, expected[7], actual[7])
}
