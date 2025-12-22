package bogo

import (
	"bytes"
	"errors"
)

var stringEncodeError = errors.New("string encoding error")

func encodeString(v string) ([]byte, error) {
	// Always use the standard encoding format, even for empty strings
	lenInfoBytes, err := encodeUint(uint64(len(v)))
	if err != nil {
		return []byte{}, wrapError(stringEncodeError, err.Error())
	}
	buf := bytes.Buffer{}
	if err = buf.WriteByte(byte(TypeString)); err != nil {
		return []byte{}, wrapError(stringEncodeError, err.Error())
	}
	if _, err = buf.Write(lenInfoBytes[1:]); err != nil {
		return []byte{}, wrapError(stringEncodeError, err.Error())
	}
	buf.WriteString(v)

	return buf.Bytes(), nil
}

func decodeString(data []byte, sizeLen int) (any, error) {
	size, err := decodeUint(data[:sizeLen])
	if err != nil {
		return nil, err
	}
	return string(data[sizeLen : sizeLen+int(size)]), nil
}