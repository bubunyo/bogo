package bogo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

func Decode(data []byte) (any, error) {
	version := data[0]
	if version != Version {
		fmt.Fprintln(os.Stderr, "unsupported bogo version")
		os.Exit(1)
	}

	bType := Type(data[1])
	switch bType {
	case TypeNull:
		return nil, nil
	case TypeString:
		storageLen := decodeVarint(data[2 : maxStorageByteLength+2])
		return decodeString(data, 7, int(storageLen+7))
	case TypeBoolTrue:
		return true, nil
	case TypeBoolFalse:
		return false, nil
	case TypeInt:
		return decodeNum[int64](data[2:10]), nil
	case TypeFloat:
		return decodeNum[float64](data[2:10]), nil
	}

	return nil, nil
}

func decodeString(data []byte, n, m int) (any, error) {
	s := string(data[n:m])
	return s, nil
}

type TypeNumber interface {
	int64 | float64
}

func decodeNum[T TypeNumber](data []byte) T {
	var decodedNum T
	err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &decodedNum)
	if err != nil {
		log.Fatal(err)
	}
	return decodedNum
}

// Decode an integer from varint encoding
func decodeVarint(data []byte) int64 {
	val, _ := binary.Varint(data)
	return val
}
