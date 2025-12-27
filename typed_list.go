package bogo

import (
	"bytes"
	"fmt"
	"reflect"
)

func encodeTypedList(arr any) ([]byte, error) {
	v := reflect.ValueOf(arr)

	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, fmt.Errorf("typed list encoder: type is not list/slice")
	}

	if v.Len() == 0 {
		// Empty list - encode as regular list
		return encodeList(arr)
	}

	// Get the element type
	elemType := v.Type().Elem()
	firstElem := v.Index(0).Interface()

	// Check if all elements are the same type
	for i := 1; i < v.Len(); i++ {
		elem := v.Index(i).Interface()
		if reflect.TypeOf(elem) != reflect.TypeOf(firstElem) {
			// Mixed types - fall back to regular list
			return encodeList(arr)
		}
	}

	buf := bytes.Buffer{}

	// Determine the element type for optimization
	var elementTypeCode byte
	switch elemType.Kind() {
	case reflect.String:
		elementTypeCode = TypeString
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		elementTypeCode = TypeInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if elemType.Kind() == reflect.Uint8 {
			elementTypeCode = TypeByte
		} else {
			elementTypeCode = TypeUint
		}
	case reflect.Float32, reflect.Float64:
		elementTypeCode = TypeFloat
	case reflect.Bool:
		elementTypeCode = TypeBoolTrue // We'll handle true/false during encoding
	default:
		// Unsupported type for optimization - fall back to regular list
		return encodeList(arr)
	}

	// Encode each element without type headers
	elementsBuf := bytes.Buffer{}
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i).Interface()

		switch elementTypeCode {
		case TypeString:
			str := elem.(string)
			strBytes := []byte(str)
			lenData, err := encodeUint(uint64(len(strBytes)))
			if err != nil {
				return nil, err
			}
			elementsBuf.Write(lenData[1:]) // Remove type byte
			elementsBuf.Write(strBytes)
		case TypeInt:
			val := reflect.ValueOf(elem).Int()
			intBytes, err := encodeIntValue(val)
			if err != nil {
				return nil, err
			}
			elementsBuf.Write(intBytes[1:]) // Remove type byte
		case TypeUint:
			val := reflect.ValueOf(elem).Uint()
			uintBytes, err := encodeUintValue(val)
			if err != nil {
				return nil, err
			}
			elementsBuf.Write(uintBytes[1:]) // Remove type byte
		case TypeByte:
			b := elem.(byte)
			elementsBuf.WriteByte(b)
		case TypeFloat:
			val := reflect.ValueOf(elem).Float()
			floatBytes, err := encodeFloatValue(val)
			if err != nil {
				return nil, err
			}
			elementsBuf.Write(floatBytes[1:]) // Remove type byte
		case TypeBoolTrue:
			b := elem.(bool)
			if b {
				elementsBuf.WriteByte(1)
			} else {
				elementsBuf.WriteByte(0)
			}
		}
	}

	elementsData := elementsBuf.Bytes()

	// Encode count
	countData, err := encodeUint(uint64(v.Len()))
	if err != nil {
		return nil, err
	}

	// Encode total data length
	totalLength := 1 + len(countData) - 1 + len(elementsData) // 1 for element type + count + elements
	lengthData, err := encodeUint(uint64(totalLength))
	if err != nil {
		return nil, err
	}

	// Build final typed list: TypeTypedList + LenSize + DataSize + ElementType + Count + Elements
	buf.WriteByte(TypeTypedList)
	buf.Write(lengthData[1:]) // Remove type byte from length encoding
	buf.WriteByte(elementTypeCode)
	buf.Write(countData[1:]) // Remove type byte from count encoding
	buf.Write(elementsData)

	return buf.Bytes(), nil
}

// Helper functions for encoding numeric values without type headers
func encodeIntValue(val int64) ([]byte, error) {
	return encodeNum(val)
}

func encodeUintValue(val uint64) ([]byte, error) {
	return encodeNum(val)
}

func encodeFloatValue(val float64) ([]byte, error) {
	return encodeNum(val)
}

func decodeTypedList(data []byte) (any, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("typed list decode error: insufficient data for size")
	}

	// Read the total data size
	sizeLen := int(data[0])
	if len(data) < sizeLen+1 {
		return nil, fmt.Errorf("typed list decode error: insufficient data for size info")
	}

	totalSize, err := decodeUint(data[1 : 1+sizeLen])
	if err != nil {
		return nil, fmt.Errorf("typed list decode error: %w", err)
	}

	dataStart := 1 + sizeLen
	dataEnd := dataStart + int(totalSize)
	if len(data) < dataEnd {
		return nil, fmt.Errorf("typed list decode error: insufficient data for content")
	}

	listData := data[dataStart:dataEnd]

	if len(listData) < 1 {
		return nil, fmt.Errorf("typed list decode error: no element type")
	}

	// Read element type
	elementType := Type(listData[0])
	listData = listData[1:]

	// Read count
	if len(listData) < 1 {
		return nil, fmt.Errorf("typed list decode error: no count info")
	}

	countLen := int(listData[0])
	if len(listData) < countLen+1 {
		return nil, fmt.Errorf("typed list decode error: insufficient count data")
	}

	count, err := decodeUint(listData[1 : 1+countLen])
	if err != nil {
		return nil, fmt.Errorf("typed list decode error: failed to decode count: %w", err)
	}

	elementsData := listData[1+countLen:]

	// Decode elements based on type
	switch elementType {
	case TypeString:
		result := make([]string, count)
		pos := 0
		for i := uint64(0); i < count; i++ {
			if pos >= len(elementsData) {
				return nil, fmt.Errorf("typed list decode error: insufficient string data at index %d", i)
			}

			strLenSize := int(elementsData[pos])
			pos++
			if pos+strLenSize > len(elementsData) {
				return nil, fmt.Errorf("typed list decode error: insufficient string length data")
			}

			strLen, err := decodeUint(elementsData[pos : pos+strLenSize])
			if err != nil {
				return nil, fmt.Errorf("typed list decode error: failed to decode string length: %w", err)
			}
			pos += strLenSize

			if pos+int(strLen) > len(elementsData) {
				return nil, fmt.Errorf("typed list decode error: insufficient string content data")
			}

			result[i] = string(elementsData[pos : pos+int(strLen)])
			pos += int(strLen)
		}
		return result, nil

	case TypeByte:
		if len(elementsData) != int(count) {
			return nil, fmt.Errorf("typed list decode error: byte list size mismatch")
		}
		return elementsData, nil

	case TypeInt:
		result := make([]int64, count)
		pos := 0
		for i := uint64(0); i < count; i++ {
			if pos >= len(elementsData) {
				return nil, fmt.Errorf("typed list decode error: insufficient int data")
			}

			intLenSize := int(elementsData[pos])
			pos++
			if pos+intLenSize > len(elementsData) {
				return nil, fmt.Errorf("typed list decode error: insufficient int length data")
			}

			val, err := decodeInt(elementsData[pos : pos+intLenSize])
			if err != nil {
				return nil, fmt.Errorf("typed list decode error: failed to decode int: %w", err)
			}
			result[i] = val
			pos += intLenSize
		}
		return result, nil

	case TypeUint:
		result := make([]uint64, count)
		pos := 0
		for i := uint64(0); i < count; i++ {
			if pos >= len(elementsData) {
				return nil, fmt.Errorf("typed list decode error: insufficient uint data")
			}

			uintLenSize := int(elementsData[pos])
			pos++
			if pos+uintLenSize > len(elementsData) {
				return nil, fmt.Errorf("typed list decode error: insufficient uint length data")
			}

			val, err := decodeUint(elementsData[pos : pos+uintLenSize])
			if err != nil {
				return nil, fmt.Errorf("typed list decode error: failed to decode uint: %w", err)
			}
			result[i] = val
			pos += uintLenSize
		}
		return result, nil

	case TypeFloat:
		result := make([]float64, count)
		pos := 0
		for i := uint64(0); i < count; i++ {
			if pos >= len(elementsData) {
				return nil, fmt.Errorf("typed list decode error: insufficient float data")
			}

			floatLenSize := int(elementsData[pos])
			pos++
			if pos+floatLenSize > len(elementsData) {
				return nil, fmt.Errorf("typed list decode error: insufficient float length data")
			}

			val, err := decodeFloat(elementsData[pos : pos+floatLenSize])
			if err != nil {
				return nil, fmt.Errorf("typed list decode error: failed to decode float: %w", err)
			}
			result[i] = val
			pos += floatLenSize
		}
		return result, nil

	case TypeBoolTrue:
		result := make([]bool, count)
		if len(elementsData) != int(count) {
			return nil, fmt.Errorf("typed list decode error: bool list size mismatch")
		}

		for i := uint64(0); i < count; i++ {
			result[i] = elementsData[i] == 1
		}
		return result, nil

	default:
		return nil, fmt.Errorf("typed list decode error: unsupported element type: %d", elementType)
	}
}
