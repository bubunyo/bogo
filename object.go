package bogo

import (
	"bytes"
	"errors"
	"fmt"
)

var objEncErr = errors.New("object encoder error")

func encodeObject(v any) ([]byte, error) {
	switch obj := v.(type) {
	case map[string]any:
		return encodeMap(obj)
	default:
		return nil, wrapError(objEncErr, "object type not supported")
	}
}

func encodeMap(m map[string]any) ([]byte, error) {
	fieldsBuf := bytes.Buffer{}
	
	// Encode each key-value pair as field entries
	for key, value := range m {
		fieldEntry, err := encodeFieldEntry(key, value)
		if err != nil {
			return nil, wrapError(objEncErr, "failed to encode field", key, err.Error())
		}
		fieldsBuf.Write(fieldEntry)
	}
	
	fieldsData := fieldsBuf.Bytes()
	fieldsSize := len(fieldsData)
	
	// Encode the total size of all fields
	encodedSizeData, err := encodeUint(uint64(fieldsSize))
	if err != nil {
		return nil, wrapError(objEncErr, "failed to encode fields size", err.Error())
	}
	
	// Build final object: TypeObject + LenSize + DataSize + FieldData
	result := bytes.Buffer{}
	result.WriteByte(TypeObject)
	result.Write(encodedSizeData[1:]) // remove type byte from size encoding
	result.Write(fieldsData)
	
	return result.Bytes(), nil
}

func encodeFieldEntry(key string, value any) ([]byte, error) {
	// Encode the value first to know its size
	encodedValue, err := encode(value)
	if err != nil {
		return nil, err
	}
	
	keyBytes := []byte(key)
	keyLen := len(keyBytes)
	
	if keyLen > 255 {
		return nil, errors.New("key too long, maximum 255 bytes")
	}
	
	// Calculate entry size: keyLen(1) + key + value
	entrySize := 1 + keyLen + len(encodedValue)
	
	// Encode entry size
	encodedEntrySize, err := encodeUint(uint64(entrySize))
	if err != nil {
		return nil, err
	}
	
	// Build field entry: LenSize + EntrySize + KeyLength + Key + Value
	entry := bytes.Buffer{}
	entry.Write(encodedEntrySize[1:]) // remove type byte
	entry.WriteByte(byte(keyLen))
	entry.Write(keyBytes)
	entry.Write(encodedValue)
	
	return entry.Bytes(), nil
}

var objDecErr = errors.New("object decoder error")

func decodeObject(data []byte) (map[string]any, error) {
	if len(data) == 0 {
		return map[string]any{}, nil
	}
	
	// Read the size of all field data
	sizeLen := int(data[0])
	if len(data) < sizeLen+1 {
		return nil, wrapError(objDecErr, "insufficient data for field size")
	}
	
	fieldsSize, err := decodeUint(data[1 : 1+sizeLen])
	if err != nil {
		return nil, wrapError(objDecErr, "failed to decode fields size", err.Error())
	}
	
	fieldsStart := 1 + sizeLen
	fieldsEnd := fieldsStart + int(fieldsSize)
	if len(data) < fieldsEnd {
		return nil, wrapError(objDecErr, "insufficient data for fields")
	}
	
	fieldsData := data[fieldsStart:fieldsEnd]
	
	// Parse all field entries
	result := make(map[string]any)
	pos := 0
	
	for pos < len(fieldsData) {
		key, value, bytesRead, err := decodeFieldEntry(fieldsData[pos:])
		if err != nil {
			return nil, wrapError(objDecErr, "failed to decode field entry", err.Error())
		}
		
		result[key] = value
		pos += bytesRead
	}
	
	return result, nil
}

func decodeFieldEntry(data []byte) (key string, value any, bytesRead int, err error) {
	if len(data) == 0 {
		return "", nil, 0, errors.New("empty field entry data")
	}
	
	// Read entry size
	entrySizeLen := int(data[0])
	if len(data) < entrySizeLen+1 {
		return "", nil, 0, errors.New("insufficient data for entry size")
	}
	
	entrySize, err := decodeUint(data[1 : 1+entrySizeLen])
	if err != nil {
		return "", nil, 0, err
	}
	
	entryStart := 1 + entrySizeLen
	entryEnd := entryStart + int(entrySize)
	if len(data) < entryEnd {
		return "", nil, 0, errors.New("insufficient data for entry content")
	}
	
	entryData := data[entryStart:entryEnd]
	
	// Read key length and key
	if len(entryData) < 1 {
		return "", nil, 0, errors.New("insufficient data for key length")
	}
	
	keyLen := int(entryData[0])
	if len(entryData) < 1+keyLen {
		return "", nil, 0, errors.New("insufficient data for key")
	}
	
	key = string(entryData[1 : 1+keyLen])
	
	// Decode value
	valueData := entryData[1+keyLen:]
	value, err = decodeValue(valueData)
	if err != nil {
		return "", nil, 0, err
	}
	
	return key, value, entryEnd, nil
}

func decodeValue(data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	
	// Use the existing decode function but without version check
	switch Type(data[0]) {
	case TypeNull:
		return nil, nil
	case TypeBoolTrue:
		return true, nil
	case TypeBoolFalse:
		return false, nil
	case TypeString:
		sizeLen := int(data[1])
		return decodeString(data[2:], sizeLen)
	case TypeByte:
		return decodeByte(data[1:])
	case TypeInt:
		sizeLen := int(data[1])
		return decodeInt(data[2 : 2+sizeLen])
	case TypeUint:
		sizeLen := int(data[1])
		return decodeUint(data[2 : 2+sizeLen])
	case TypeFloat:
		sizeLen := int(data[1])
		return decodeFloat(data[2 : 2+sizeLen])
	case TypeBlob:
		blob, err := decodeBlob(data[1:])
		if err != nil {
			return nil, err
		}
		return blob, nil
	case TypeTimestamp:
		timestamp, err := decodeTimestamp(data[1:])
		if err != nil {
			return nil, err
		}
		return timestamp, nil
	case TypeArray:
		// Decode array within object
		array, err := decodeArrayValue(data[1:])
		if err != nil {
			return nil, err
		}
		return array, nil
	case TypeTypedArray:
		// Decode typed array within object
		typedArray, err := decodeTypedArray(data[1:])
		if err != nil {
			return nil, err
		}
		return typedArray, nil
	case TypeObject:
		// Recursive object decoding
		obj, err := decodeObject(data[1:])
		if err != nil {
			return nil, err
		}
		return obj, nil
	default:
		return nil, fmt.Errorf("unsupported value type: %d", data[0])
	}
}

// decodeArrayValue decodes an array and returns the result as any
func decodeArrayValue(data []byte) (any, error) {
	if len(data) == 0 {
		return []any{}, nil
	}
	
	// Get array size
	sizeLen := int(data[0])
	if len(data) < sizeLen+1 {
		return nil, errors.New("insufficient data for array size")
	}
	
	arraySize, err := decodeUint(data[1 : 1+sizeLen])
	if err != nil {
		return nil, err
	}
	
	arrayStart := 1 + sizeLen
	arrayEnd := arrayStart + int(arraySize)
	if len(data) < arrayEnd {
		return nil, errors.New("insufficient data for array content")
	}
	
	arrayData := data[arrayStart:arrayEnd]
	
	// Parse all array elements
	result := []any{}
	pos := 0
	
	for pos < len(arrayData) {
		// Find the end of the current element by decoding its header
		if pos >= len(arrayData) {
			break
		}
		
		// Decode the element
		element, err := decodeValue(arrayData[pos:])
		if err != nil {
			return nil, err
		}
		
		result = append(result, element)
		
		// Find the size of this encoded element to advance position
		elementSize, err := getElementSize(arrayData[pos:])
		if err != nil {
			return nil, err
		}
		
		pos += elementSize
	}
	
	return result, nil
}

// getElementSize calculates the size of an encoded element
func getElementSize(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, errors.New("empty data")
	}
	
	switch Type(data[0]) {
	case TypeNull:
		return 1, nil
	case TypeBoolTrue, TypeBoolFalse:
		return 1, nil
	case TypeByte:
		return 2, nil
	case TypeString:
		if len(data) < 2 {
			return 0, errors.New("insufficient data for string size")
		}
		sizeLen := int(data[1])
		if len(data) < 2+sizeLen {
			return 0, errors.New("insufficient data for string size value")
		}
		stringSize, err := decodeUint(data[2 : 2+sizeLen])
		if err != nil {
			return 0, err
		}
		return 2 + sizeLen + int(stringSize), nil
	case TypeInt, TypeUint, TypeFloat:
		if len(data) < 2 {
			return 0, errors.New("insufficient data for numeric size")
		}
		sizeLen := int(data[1])
		return 2 + sizeLen, nil
	case TypeBlob:
		if len(data) < 2 {
			return 0, errors.New("insufficient data for blob size")
		}
		sizeLen := int(data[1])
		if len(data) < 2+sizeLen {
			return 0, errors.New("insufficient data for blob size value")
		}
		blobSize, err := decodeUint(data[2 : 2+sizeLen])
		if err != nil {
			return 0, err
		}
		return 2 + sizeLen + int(blobSize), nil
	case TypeTimestamp:
		return 9, nil // 1 + 8 bytes for timestamp
	case TypeArray:
		if len(data) < 2 {
			return 0, errors.New("insufficient data for array size")
		}
		sizeLen := int(data[1])
		if len(data) < 2+sizeLen {
			return 0, errors.New("insufficient data for array size value")
		}
		arraySize, err := decodeUint(data[2 : 2+sizeLen])
		if err != nil {
			return 0, err
		}
		return 2 + sizeLen + int(arraySize), nil
	case TypeTypedArray:
		if len(data) < 2 {
			return 0, errors.New("insufficient data for typed array size")
		}
		sizeLen := int(data[1])
		if len(data) < 2+sizeLen {
			return 0, errors.New("insufficient data for typed array size value")
		}
		arraySize, err := decodeUint(data[2 : 2+sizeLen])
		if err != nil {
			return 0, err
		}
		return 2 + sizeLen + int(arraySize), nil
	case TypeObject:
		if len(data) < 2 {
			return 0, errors.New("insufficient data for object size")
		}
		sizeLen := int(data[1])
		if len(data) < 2+sizeLen {
			return 0, errors.New("insufficient data for object size value")
		}
		objSize, err := decodeUint(data[2 : 2+sizeLen])
		if err != nil {
			return 0, err
		}
		return 2 + sizeLen + int(objSize), nil
	default:
		return 0, fmt.Errorf("unsupported type: %d", data[0])
	}
}
