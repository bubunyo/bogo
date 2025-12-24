package bogo

import (
	"fmt"
	"reflect"
)

// Basic type encoding functions

func encodeNull() []byte {
	return []byte{TypeNull}
}

func encodeBool(b bool) []byte {
	var by = TypeBoolFalse
	if b {
		by = TypeBoolTrue
	}
	return []byte{byte(by)}
}

func encodeByte(b byte) ([]byte, error) {
	return []byte{byte(TypeByte), b}, nil
}

// Basic type decoding functions

func decodeNull() (any, error) {
	return nil, nil
}

func decodeBoolTrue() (bool, error) {
	return true, nil
}

func decodeBoolFalse() (bool, error) {
	return false, nil
}

func decodeByte(data []byte) (byte, error) {
	if len(data) < 1 {
		return 0, fmt.Errorf("byte decode error: insufficient data, need 1 byte, got %d", len(data))
	}

	return data[0], nil
}

// Null value detection
func isNullValue(v any) bool {
	if v == nil {
		return true
	}

	data := reflect.ValueOf(v)
	if !data.IsValid() {
		return true
	}

	// Handle nil pointers, slices, maps, etc.
	if data.CanInterface() && data.Kind() != reflect.Invalid {
		switch data.Kind() {
		case reflect.Ptr, reflect.Interface,
			reflect.Slice, reflect.Map,

			// we should never get here add for completeness sakes
			reflect.Chan, reflect.Func:
			if data.IsNil() {
				return true
			}
		}
	}

	return false
}

// isZeroValue checks if a value is the zero value for its type
func isZeroValue(v any) bool {
	if v == nil {
		return true
	}

	switch val := v.(type) {
	case string:
		return val == ""
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(val).Int() == 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(val).Uint() == 0
	case float32, float64:
		return reflect.ValueOf(val).Float() == 0.0
	case bool:
		return !val
	default:
		return false
	}
}

func wrapError(err1 error, err2 ...string) error {
	if len(err2) == 0 {
		return err1
	}
	return fmt.Errorf("%w: %s", err1, err2[0])
}
