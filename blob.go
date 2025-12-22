package bogo

import (
	"bytes"
	"errors"
	"fmt"
)

var blobEncodeError = errors.New("blob encoding error")

func encodeBlob(data []byte) ([]byte, error) {
	dataLen := len(data)
	encodedLengthData, err := encodeUint(uint64(dataLen))
	if err != nil {
		return []byte{}, wrapError(blobEncodeError, err.Error())
	}

	buf := bytes.Buffer{}
	if err = buf.WriteByte(byte(TypeBlob)); err != nil {
		return []byte{}, wrapError(blobEncodeError, err.Error())
	}

	// Write length info (remove the type byte from encodeUint result)
	if _, err = buf.Write(encodedLengthData[1:]); err != nil {
		return []byte{}, wrapError(blobEncodeError, err.Error())
	}

	// Write the blob data
	if _, err = buf.Write(data); err != nil {
		return []byte{}, wrapError(blobEncodeError, err.Error())
	}

	return buf.Bytes(), nil
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
