package bogo

import (
	"encoding/binary"
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
	case TypeBlob:
		blob, err := decodeBlob(data[2:])
		if err != nil {
			return nil, err
		}
		return blob, nil
	case TypeTimestamp:
		timestamp, err := decodeTimestamp(data[2:])
		if err != nil {
			return nil, err
		}
		return timestamp, nil
	case TypeByte:
		byteVal, err := decodeByte(data[2:])
		if err != nil {
			return nil, err
		}
		return byteVal, nil
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

func decodeBlob(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}
	
	sizeLen := int(data[0])
	if len(data) < sizeLen+1 {
		return nil, fmt.Errorf("blob decode error: insufficient data for size")
	}
	
	size, err := decodeUint(data[1 : 1+sizeLen])
	if err != nil {
		return nil, fmt.Errorf("blob decode error: %w", err)
	}
	
	dataStart := 1 + sizeLen
	dataEnd := dataStart + int(size)
	if len(data) < dataEnd {
		return nil, fmt.Errorf("blob decode error: insufficient data for blob content")
	}
	
	return data[dataStart:dataEnd], nil
}

func decodeTimestamp(data []byte) (int64, error) {
	if len(data) < 8 {
		return 0, fmt.Errorf("timestamp decode error: insufficient data, need 8 bytes, got %d", len(data))
	}
	
	timestamp := int64(binary.LittleEndian.Uint64(data[:8]))
	return timestamp, nil
}

func decodeByte(data []byte) (byte, error) {
	if len(data) < 1 {
		return 0, fmt.Errorf("byte decode error: insufficient data, need 1 byte, got %d", len(data))
	}
	
	return data[0], nil
}
