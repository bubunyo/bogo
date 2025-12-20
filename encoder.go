package bogo

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"time"
)

func Encode(v any) ([]byte, error) {
	res, err := encode(v)
	if err != nil {
		return nil, err
	}
	return append([]byte{Version}, res...), nil
}

func encode(v any) ([]byte, error) {
	if v == nil {
		return []byte{TypeNull}, nil
	}
	
	data := reflect.ValueOf(v)

	if !data.IsValid() {
		return []byte{TypeNull}, nil
	}
	
	// Handle nil pointers, slices, maps, etc.
	if data.CanInterface() && data.Kind() != reflect.Invalid {
		switch data.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
			if data.IsNil() {
				return []byte{TypeNull}, nil
			}
		}
	}

	// Special case for time.Time
	if t, ok := v.(time.Time); ok {
		return encodeTimestamp(t.UnixMilli())
	}

	switch data.Kind() {
	case reflect.Ptr:
		// Dereference pointer
		return encode(data.Elem().Interface())
	case reflect.Bool:
		return encodeBool(data.Interface().(bool)), nil

	case reflect.String:
		buf, err := encodeString(data.Interface().(string))
		if err != nil {
			return nil, err
		}
		return buf, nil

	case reflect.Uint8:
		// Special case for byte (uint8)
		return encodeByte(data.Interface().(byte))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		res, err := encodeNum(v)
		if err != nil {
			return nil, err
		}
		return res, nil

	case reflect.Slice, reflect.Array:
		// Special case for []byte - encode as blob
		if data.Type().Elem().Kind() == reflect.Uint8 {
			byteSlice := data.Interface().([]byte)
			return encodeBlob(byteSlice)
		}
		return encodeArray(data.Interface())
	case reflect.Map:
		// object
	}

	return nil, fmt.Errorf("bogo error: unsupported type. type=%T", v)
}

var stringEncodeError = errors.New("string encoding error")

func wrapError(err1 error, err2 ...string) error {
	return fmt.Errorf("%w, %s", err1, err2)
}

func encodeString(v string) ([]byte, error) {
	if len(v) == 0 {
		return []byte{byte(TypeString), 0}, nil
	}
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

func encodeBool(b bool) []byte {
	var by = TypeBoolFalse
	if b {
		by = TypeBoolTrue
	}
	return []byte{byte(by)}
}

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

func encodeTimestamp(timestamp int64) ([]byte, error) {
	buf := make([]byte, 9) // 1 byte type + 8 bytes int64
	buf[0] = byte(TypeTimestamp)
	binary.LittleEndian.PutUint64(buf[1:], uint64(timestamp))
	return buf, nil
}

func encodeByte(b byte) ([]byte, error) {
	return []byte{byte(TypeByte), b}, nil
}
