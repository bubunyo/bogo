package bogo

import (
	"bytes"
	"errors"
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

func encodeString(v string) ([]byte, error) {
	strLen, err := encodeInt(int64(len(v)))
	buf := bytes.Buffer{}
	buf.Grow(3 + len(strLen))
	if err = buf.WriteByte(byte(TypeString)); err != nil {
		return []byte{}, err
	}
	if err = buf.WriteByte(byte(len(strLen))); err != nil {
		return []byte{}, err
	}
	if _, err = buf.Write(strLen); err != nil {
		return []byte{}, err
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
