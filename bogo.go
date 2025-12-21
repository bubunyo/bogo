package bogo

import (
	"fmt"
	"reflect"
	"time"
)

// Main encoding function
func Encode(v any) ([]byte, error) {
	res, err := encode(v)
	if err != nil {
		return nil, err
	}
	return append([]byte{Version}, res...), nil
}

func encode(v any) ([]byte, error) {
	// Handle null values
	if isNullValue(v) {
		return encodeNull(), nil
	}
	
	data := reflect.ValueOf(v)

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
		// Handle map types
		return encodeObject(data.Interface())
	case reflect.Struct:
		// Handle struct types by converting to map using parse.go
		return encodeStruct(data.Interface())
	}

	return nil, fmt.Errorf("bogo error: unsupported type. type=%T", v)
}

// Main decoding function
func Decode(data []byte) (any, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("bogo decode error: insufficient data, need at least 2 bytes for version and type")
	}
	
	version := data[0]
	if version != Version {
		return nil, fmt.Errorf("bogo decode error: unsupported version %d, expected version %d", version, Version)
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
	case TypeTypedArray:
		typedArray, err := decodeTypedArray(data[2:])
		if err != nil {
			return nil, err
		}
		return typedArray, nil
	case TypeObject:
		obj, err := decodeObject(data[2:])
		if err != nil {
			return nil, err
		}
		return obj, nil
	default:
		return nil, fmt.Errorf("type coder not supported, type=%d", data[1])
	}
}

// Default encoder and decoder instances for backward compatibility
var (
	defaultEncoder = NewConfigurableEncoder()
	defaultDecoder = NewConfigurableDecoder()
)

// Marshal encodes v into bogo binary format, similar to json.Marshal
func Marshal(v any) ([]byte, error) {
	return defaultEncoder.Encode(v)
}

// Unmarshal decodes bogo binary data into the value pointed to by v, similar to json.Unmarshal
func Unmarshal(data []byte, v any) error {
	result, err := defaultDecoder.Decode(data)
	if err != nil {
		return err
	}
	
	return assignResult(result, v)
}

// assignResult assigns the decoded result to the pointer provided by the user
func assignResult(result any, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("bogo: Unmarshal destination must be a non-nil pointer")
	}
	
	elem := rv.Elem()
	resultValue := reflect.ValueOf(result)
	
	// Handle nil result
	if result == nil {
		elem.Set(reflect.Zero(elem.Type()))
		return nil
	}
	
	// Try direct assignment first
	if resultValue.Type().AssignableTo(elem.Type()) {
		elem.Set(resultValue)
		return nil
	}
	
	// Handle type conversions for common cases
	switch elem.Kind() {
	case reflect.Interface:
		// For interface{} destinations, assign directly
		elem.Set(resultValue)
		return nil
		
	case reflect.String:
		if str, ok := result.(string); ok {
			elem.SetString(str)
			return nil
		}
		
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, ok := result.(int64); ok {
			if elem.OverflowInt(val) {
				return fmt.Errorf("bogo: value %d overflows %s", val, elem.Type())
			}
			elem.SetInt(val)
			return nil
		}
		
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val, ok := result.(uint64); ok {
			if elem.OverflowUint(val) {
				return fmt.Errorf("bogo: value %d overflows %s", val, elem.Type())
			}
			elem.SetUint(val)
			return nil
		}
		if val, ok := result.(byte); ok {
			elem.SetUint(uint64(val))
			return nil
		}
		
	case reflect.Float32, reflect.Float64:
		if val, ok := result.(float64); ok {
			if elem.OverflowFloat(val) {
				return fmt.Errorf("bogo: value %f overflows %s", val, elem.Type())
			}
			elem.SetFloat(val)
			return nil
		}
		
	case reflect.Bool:
		if val, ok := result.(bool); ok {
			elem.SetBool(val)
			return nil
		}
		
	case reflect.Slice:
		if elem.Type().Elem().Kind() == reflect.Uint8 {
			// Handle []byte
			if val, ok := result.([]byte); ok {
				elem.SetBytes(val)
				return nil
			}
		}
		// Handle other slice types
		if resultValue.Kind() == reflect.Slice && resultValue.Type().ConvertibleTo(elem.Type()) {
			elem.Set(resultValue.Convert(elem.Type()))
			return nil
		}
		
	case reflect.Map:
		if resultValue.Kind() == reflect.Map {
			if resultValue.Type().AssignableTo(elem.Type()) {
				elem.Set(resultValue)
				return nil
			}
			// Handle map[string]any -> map[string]T conversion
			if elem.Type().Key() == reflect.TypeOf("") && resultValue.Type() == reflect.TypeOf(map[string]any{}) {
				elem.Set(resultValue)
				return nil
			}
		}
	}
	
	return fmt.Errorf("bogo: cannot unmarshal %T into %T", result, v)
}

// SetDefaultEncoder sets the default encoder used by Marshal
func SetDefaultEncoder(encoder *Encoder) {
	if encoder != nil {
		defaultEncoder = encoder
	}
}

// SetDefaultDecoder sets the default decoder used by Unmarshal
func SetDefaultDecoder(decoder *Decoder) {
	if decoder != nil {
		defaultDecoder = decoder
	}
}

// GetDefaultEncoder returns the current default encoder
func GetDefaultEncoder() *Encoder {
	return defaultEncoder
}

// GetDefaultDecoder returns the current default decoder
func GetDefaultDecoder() *Decoder {
	return defaultDecoder
}