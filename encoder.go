package bogo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
)

func UnMarshall() {}
func Marshall()   {}

type Type byte

// Type constants
const (
	maxStorageByteLength = 5
	// Version helps determine which encoders/decoders to use
	Version byte = 0x00 // version 0
)

const (
	TypeNull = iota
	TypeBoolTrue
	TypeBoolFalse
	TypeString
	TypeArray
	TypeObject
	TypeByte
	TypeInt
	TypeUInt
	TypeFloat
)

type FieldInfo struct {
	Type  Type
	Key   []byte
	Value any
}

type number interface {
	int | int32 | int64
}

func Encode(v any) ([]byte, error) {
	prefix := []byte{Version}
	switch v.(type) {
	case nil:
		return append(prefix, byte(TypeNull)), nil
	case string:
		return append(prefix, encodeString(v.(string))...), nil
	case bool:
		return append(prefix, encodeBool(v.(bool))...), nil
	case int, int8, int16, int32, int64, float32, float64:
		res, err := encodeNum(v)
		if err != nil {
			return nil, err
		}
		return append(prefix, res...), nil
	case []any:
	default:
		// result := make(map[string]FieldInfo)
		// p.extractFields(reflect.ValueOf(in), reflect.TypeOf(in), &result)
		// return result
	}
	return nil, errors.New("bogo error: unsupported type")
}

// Encode an integer using varint encoding
func encodeVarint(value int64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, value)
	if n > maxStorageByteLength {
		fmt.Fprintln(os.Stderr, "storage length byte exceeded")
		os.Exit(1)
	}
	binary.PutVarint(buf, value)
	return buf[:maxStorageByteLength]
}

func encodeString(v string) []byte {
	l := encodeVarint(int64(len(v)))
	return append([]byte{byte(TypeString)}, append(l, v...)...)
}

func encodeBool(b bool) []byte {
	var by = TypeBoolFalse
	if b {
		by = TypeBoolTrue
	}
	return []byte{byte(by)}
}

func encodeNum(v any) ([]byte, error) {
	switch data := v.(type) {
	case int:
		result := make([]byte, binary.MaxVarintLen64+2)
		result[0] = byte(TypeInt)
		n := binary.PutVarint(result[2:], int64(data))
		result[1] = byte(n)
		return result[:n+2], nil
	case uint, uint8, uint16, uint32, uint64:
		result := make([]byte, binary.MaxVarintLen64+2)
		result[0] = byte(TypeUInt)
		var val uint64
		switch d := data.(type) {
		case uint:
			val = uint64(d)
		case uint8:
			val = uint64(d)
		case uint16:
			val = uint64(d)
		case uint32:
			val = uint64(d)
		case uint64:
			val = d
		}
		n := binary.PutUvarint(result[2:], val)
		result[1] = byte(n)
		return result[:n+2], nil
	case float32:
		result := make([]byte, 6) // 2 bytes prefix + 4 bytes data
		result[0] = byte(TypeFloat)
		result[1] = 4 // size of float32 in bytes
		bits := math.Float32bits(data)
		binary.LittleEndian.PutUint32(result[2:], bits)
		return result, nil
	case float64:
		result := make([]byte, 10) // 2 bytes prefix + 8 bytes data
		result[0] = byte(TypeFloat)
		result[1] = 8 // size of float64 in bytes
		bits := math.Float64bits(data)
		binary.LittleEndian.PutUint64(result[2:], bits)
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported numeric type: %T", v)
	}
}
