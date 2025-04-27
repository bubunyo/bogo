package bogo

import (
	"bytes"
	"errors"
	"fmt"
)

func Encode(v any) ([]byte, error) {
	prefix := []byte{Version}
	switch v.(type) {
	case nil:
		return append(prefix, byte(TypeNull)), nil
	case string:
		buf, err := encodeString(v.(string))
		if err != nil {
			return nil, err
		}
		return append(prefix, buf...), nil
	case bool:
		return append(prefix, encodeBool(v.(bool))...), nil
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
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

var stringEncodeError = errors.New("string encoding error")

func wrapError(err1, err2 error) error {
	return fmt.Errorf("%w, %s", err1, err2)
}

func encodeString(v string) ([]byte, error) {
	if len(v) == 0 {
		return []byte{byte(TypeString), 0}, nil
	}
	lenInfoBytes, err := encodeUint(uint64(len(v)))
	if err != nil {
		return []byte{}, wrapError(stringEncodeError, err)
	}
	buf := bytes.Buffer{}
	if err = buf.WriteByte(byte(TypeString)); err != nil {
		return []byte{}, wrapError(stringEncodeError, err)
	}
	if _, err = buf.Write(lenInfoBytes[1:]); err != nil {
		return []byte{}, wrapError(stringEncodeError, err)
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
