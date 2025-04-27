package bogo

import (
	"fmt"
	"os"
)

func Decode(data []byte) (any, error) {
	version := data[0]
	if version != Version {
		fmt.Fprintln(os.Stderr, "unsupported bogo version")
		os.Exit(1)
	}

	switch Type(data[1]) {
	case TypeNull:
		return nil, nil
	case TypeString:
		sizeLen := int(data[2])
		return decodeString(data[3:], sizeLen)
	case TypeBoolTrue:
		return true, nil
	case TypeBoolFalse:
		return false, nil
	case TypeInt:
		sizeLen := int(data[2])
		return decodeInt(data[3 : 3+sizeLen])
	case TypeUint:
		sizeLen := int(data[2])
		return decodeUint(data[3 : 3+sizeLen])
	case TypeFloat:
		sizeLen := int(data[2])
		return decodeFloat(data[3 : 3+sizeLen])
	default:
		return nil, fmt.Errorf("type coder not supported, type=%d", data[1])
	}

}

func decodeString(data []byte, sizeLen int) (any, error) {
	size, err := decodeUint(data[:sizeLen])
	if err != nil {
		return nil, err
	}
	return string(data[sizeLen : sizeLen+int(size)]), nil
}
