package bogo

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

func Encode(v any) ([]byte, error) {
	res, err := encode(v)
	if err != nil {
		return nil, err
	}
	return append([]byte{Version}, res...), nil
}

func encode(v any) ([]byte, error) {
	data := reflect.ValueOf(v)

	if !data.IsValid() {
		return []byte{TypeNull}, nil
	}

	switch data.Kind() {
	case reflect.Bool:
		return encodeBool(data.Interface().(bool)), nil

	case reflect.String:
		buf, err := encodeString(data.Interface().(string))
		if err != nil {
			return nil, err
		}
		return buf, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		res, err := encodeNum(v)
		if err != nil {
			return nil, err
		}
		return res, nil

	case reflect.Slice, reflect.Array:
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
