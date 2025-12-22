package bogo

import (
	"fmt"
	"reflect"
	"time"
)

// Main encoding function
func Encode(v any) ([]byte, error) {
	return defaultEncoder.Encode(v)
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
		// Convert timestamp back to time.Time (in UTC to maintain consistency)
		return time.UnixMilli(timestamp).UTC(), nil
	case TypeByte:
		byteVal, err := decodeByte(data[2:])
		if err != nil {
			return nil, err
		}
		return byteVal, nil
	case TypeArray:
		array, err := decodeArrayValue(data[2:])
		if err != nil {
			return nil, err
		}
		return array, nil
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
		
	case reflect.Struct:
		// Handle map[string]any -> struct conversion using tags
		if resultMap, ok := result.(map[string]any); ok {
			return assignMapToStruct(resultMap, elem, defaultDecoder.TagName)
		}
	}
	
	return fmt.Errorf("bogo: cannot unmarshal %T into %T", result, v)
}

// assignMapToStruct assigns values from a map[string]any to a struct using struct tags
func assignMapToStruct(resultMap map[string]any, structValue reflect.Value, tagName string) error {
	structType := structValue.Type()
	
	
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		// Get field name from tag or use field name
		fieldName := getStructFieldName(field, tagName)
		
		// Skip if tag indicates to omit the field
		if fieldName == "-" {
			continue
		}
		
		// Check if the map contains this field
		mapValue, exists := resultMap[fieldName]
		if !exists {
			// Field not present in map, leave as zero value
			continue
		}
		
		
		// Recursively assign the value
		if err := assignValueToField(mapValue, fieldValue, tagName); err != nil {
			return fmt.Errorf("bogo: error assigning field %s: %w", fieldName, err)
		}
	}
	
	return nil
}

// getStructFieldName returns the field name to use based on struct tags
func getStructFieldName(field reflect.StructField, tagName string) string {
	tag := field.Tag.Get(tagName)
	if tag == "" {
		return field.Name
	}
	
	// Handle "fieldname" and "fieldname,omitempty" formats
	if commaIdx := len(tag); commaIdx > 0 {
		for i, c := range tag {
			if c == ',' {
				commaIdx = i
				break
			}
		}
		return tag[:commaIdx]
	}
	
	return tag
}

// assignValueToField assigns a value to a struct field with type conversion
func assignValueToField(value any, fieldValue reflect.Value, tagName string) error {
	if value == nil {
		fieldValue.Set(reflect.Zero(fieldValue.Type()))
		return nil
	}
	
	valueReflect := reflect.ValueOf(value)
	
	// Try direct assignment first
	if valueReflect.Type().AssignableTo(fieldValue.Type()) {
		fieldValue.Set(valueReflect)
		return nil
	}
	
	// Handle type conversions
	switch fieldValue.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			fieldValue.SetString(str)
			return nil
		}
		
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, ok := value.(int64); ok {
			if fieldValue.OverflowInt(val) {
				return fmt.Errorf("value %d overflows %s", val, fieldValue.Type())
			}
			fieldValue.SetInt(val)
			return nil
		}
		
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val, ok := value.(uint64); ok {
			if fieldValue.OverflowUint(val) {
				return fmt.Errorf("value %d overflows %s", val, fieldValue.Type())
			}
			fieldValue.SetUint(val)
			return nil
		}
		if val, ok := value.(byte); ok {
			fieldValue.SetUint(uint64(val))
			return nil
		}
		
	case reflect.Float32, reflect.Float64:
		if val, ok := value.(float64); ok {
			if fieldValue.OverflowFloat(val) {
				return fmt.Errorf("value %f overflows %s", val, fieldValue.Type())
			}
			fieldValue.SetFloat(val)
			return nil
		}
		
	case reflect.Bool:
		if val, ok := value.(bool); ok {
			fieldValue.SetBool(val)
			return nil
		}
		
	case reflect.Slice:
		if fieldValue.Type().Elem().Kind() == reflect.Uint8 {
			// Handle []byte
			if val, ok := value.([]byte); ok {
				fieldValue.SetBytes(val)
				return nil
			}
		}
		// Handle other slice types by creating a new slice and converting elements
		if valueReflect.Kind() == reflect.Slice {
			newSlice := reflect.MakeSlice(fieldValue.Type(), valueReflect.Len(), valueReflect.Len())
			for i := 0; i < valueReflect.Len(); i++ {
				elem := valueReflect.Index(i)
				if err := assignValueToField(elem.Interface(), newSlice.Index(i), tagName); err != nil {
					return err
				}
			}
			fieldValue.Set(newSlice)
			return nil
		}
		
	case reflect.Map:
		if valueReflect.Kind() == reflect.Map {
			if valueReflect.Type().AssignableTo(fieldValue.Type()) {
				fieldValue.Set(valueReflect)
				return nil
			}
			
			// Handle map[string]interface{} to map[string]T conversion
			if valueReflect.Type() == reflect.TypeOf(map[string]any{}) && fieldValue.Type().Key() == reflect.TypeOf("") {
				return convertMap(value.(map[string]any), fieldValue, tagName)
			}
		}
		
	case reflect.Struct:
		if valueMap, ok := value.(map[string]any); ok {
			return assignMapToStruct(valueMap, fieldValue, tagName)
		}
		
	case reflect.Ptr:
		// Handle pointers by creating a new instance and assigning to it
		if fieldValue.IsNil() {
			fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
		}
		return assignValueToField(value, fieldValue.Elem(), tagName)
	}
	
	// Handle special cases for specific types
	if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
		if ts, ok := value.(int64); ok {
			// The timestamp is in milliseconds
			fieldValue.Set(reflect.ValueOf(time.UnixMilli(ts)))
			return nil
		}
	}
	
	return fmt.Errorf("cannot assign %T to %s", value, fieldValue.Type())
}

// convertMap converts a map[string]interface{} to a typed map
func convertMap(sourceMap map[string]any, targetMapValue reflect.Value, tagName string) error {
	targetType := targetMapValue.Type()
	valueType := targetType.Elem()
	
	newMap := reflect.MakeMap(targetType)
	
	for key, value := range sourceMap {
		keyValue := reflect.ValueOf(key)
		
		// Convert the map value to the target type
		convertedValue := reflect.New(valueType).Elem()
		if err := assignValueToField(value, convertedValue, tagName); err != nil {
			return fmt.Errorf("failed to convert map value for key %s: %w", key, err)
		}
		
		newMap.SetMapIndex(keyValue, convertedValue)
	}
	
	targetMapValue.Set(newMap)
	return nil
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