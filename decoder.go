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

	dataTypeLen := Type(data[2])
	switch Type(data[1]) {
	case TypeNull:
		return nil, nil
	case TypeString:
		storageLen, err := decodeInt(data[2 : maxStorageByteLength+2])
		if err != nil {
			return nil, err
		}
		return decodeString(data, 7, int(storageLen+7))
	case TypeBoolTrue:
		return true, nil
	case TypeBoolFalse:
		return false, nil
	case TypeInt:
		return decodeInt(data[3 : 3+dataTypeLen])
	case TypeUint:
		return decodeUint(data[3 : 3+dataTypeLen])
	case TypeFloat:
		return decodeFloat(data[3 : 3+dataTypeLen])
	}

	return nil, nil
}

func decodeString(data []byte, n, m int) (any, error) {
	s := string(data[n:m])
	return s, nil
}
