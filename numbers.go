package bogo

import (
	"encoding/binary"
	"fmt"
	"math"
)

func encodeNum(v any) ([]byte, error) {
	switch data := v.(type) {
	case int:
		return encodeInt(int64(data))
	case int8:
		return encodeInt(int64(data))
	case int16:
		return encodeInt(int64(data))
	case int32:
		return encodeInt(int64(data))
	case int64:
		return encodeInt(data)
	case uint:
		return encodeUint(uint64(data))
	case uint8:
		return encodeUint(uint64(data))
	case uint16:
		return encodeUint(uint64(data))
	case uint32:
		return encodeUint(uint64(data))
	case uint64:
		return encodeUint(uint64(data))
	case float32:
		return encodeFloat(float64(data))
	case float64:
		return encodeFloat(data)
	default:
		return nil, fmt.Errorf("unsupported numeric type: %T", v)
	}
}

func encodeUint(data uint64) ([]byte, error) {
	result := make([]byte, binary.MaxVarintLen64+2)
	result[0] = byte(TypeUint)
	n := binary.PutUvarint(result[2:], data)
	result[1] = byte(n)
	return result[:n+2], nil
}

func encodeInt(data int64) ([]byte, error) {
	result := make([]byte, binary.MaxVarintLen64+2)
	result[0] = byte(TypeInt)
	n := binary.PutVarint(result[2:], data)
	result[1] = byte(n)
	return result[:n+2], nil
}

func decodeInt(data []byte) (int64, error) {
	val, n := binary.Varint(data)
	if n <= 0 {
		return 0, fmt.Errorf("failed to decode int at line %d", 119)
	}
	return val, nil
}

func decodeUint(data []byte) (uint64, error) {
	val, n := binary.Uvarint(data)
	if n <= 0 {
		return 0, fmt.Errorf("failed to decode uint at line %d", 126)
	}
	return val, nil
}

func decomposeFloat64(f float64) (signExp uint16, mantissa uint64) {
	bits := math.Float64bits(f)
	sign := int(bits >> 63)
	exponent := uint16((bits >> 52) & 0x7FF)
	mantissa = bits & 0xFFFFFFFFFFFFF

	// pack sign and exponent into 2 bytes
	signExp = (uint16(sign) << 15) | (exponent & 0x7FFF)
	return
}

func encodeFloat(f float64) ([]byte, error) {
	signExp, mant := decomposeFloat64(f)

	buf := make([]byte, 32) // big enough buffer
	buf[0] = byte(TypeFloat)

	binary.LittleEndian.PutUint16(buf[2:4], signExp) // write 2 bytes

	// encode mantisa
	n := binary.PutUvarint(buf[4:], mant)

	buf[1] = byte(n + 4)

	return buf[:n+4], nil
}

func decodeFloat(data []byte) (float64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("empty input")
	}
	signExpo := binary.LittleEndian.Uint16(data[0:2])
	sign := int(signExpo >> 15)
	exp := signExpo & 0x7FFF

	mantissa, n := binary.Uvarint(data[2:])
	if n <= 0 {
		return 0, fmt.Errorf("failed to decode mantissa")
	}

	// Rebuild float64 bits
	bits := (uint64(sign) << 63) | (uint64(exp) << 52) | mantissa
	return math.Float64frombits(bits), nil
}
