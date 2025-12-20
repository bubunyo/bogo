package bogo

import (
	"bytes"
	"fmt"
	"reflect"
)

func encodeArray(arr any) ([]byte, error) {
	v := reflect.ValueOf(arr)

	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, wrapError(arrEncErr, "type is not a array type")
	}

	buf := bytes.Buffer{}

	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		data, err := encode(elem.Interface())
		if err != nil {
			return nil, wrapError(arrEncErr, "error encoding element in array", err.Error())
		}
		buf.Write(data)
	}
	data := buf.Bytes()
	dataLen := len(data)
	encodedLengthData, err := encodeUint(uint64(dataLen))
	if err != nil {
		return nil, wrapError(arrEncErr, "error encoding array length", err.Error())
	}
	// remove the type information
	encodedLengthData = encodedLengthData[1:]

	buf2 := bytes.Buffer{}
	buf2.WriteByte(TypeArray)
	buf2.Write(encodedLengthData)
	buf2.Write(data)
	return buf2.Bytes(), nil
}

func decodeArray(data []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return wrapError(arrDecErr, fmt.Sprintf("invalid decoder destination, %q", reflect.TypeOf(v)))
	}
	if rv.Kind() == reflect.Array || rv.Kind() == reflect.Slice {
		return wrapError(arrDecErr, fmt.Sprintf("incompatible types %q", rv.Kind()))
	}
	elem := rv.Elem()

	i := 0
	computeDataSize := func() (uint64, error) {
		i++
		sizeLen := uint64(data[i])
		i++
		sizeLenEndI := i + int(sizeLen)
		size, err := decodeUint(data[i:sizeLenEndI])
		if err != nil {
			return 0, wrapError(arrDecErr, err.Error())
		}
		i = sizeLenEndI
		return size, nil
	}

	for i < len(data) {
		var entryVal reflect.Value
		entryType := Type(data[i])
		switch entryType {
		case TypeArray:
			size, err := computeDataSize()
			if err != nil {
				return err
			}
			var zeroSlice []any
			err = decodeArray(data[i:i+int(size)], &zeroSlice)
			if err != nil {
				return wrapError(arrDecErr, err.Error())
			}
			i += int(size)
			entryVal = reflect.ValueOf(zeroSlice)
		case TypeNull:
			entryVal = reflect.Zero(reflect.TypeOf((*any)(nil)).Elem())
			i++
		case TypeBoolTrue:
			entryVal = reflect.ValueOf(true)
			i++
		case TypeBoolFalse:
			entryVal = reflect.ValueOf(false)
			i++
		case TypeByte:
			entryVal = reflect.ValueOf(data[i])
			i++
		case TypeString:
			size, err := computeDataSize()
			if err != nil {
				return err
			}
			entryVal = reflect.ValueOf(string(data[i : i+int(size)]))
			i += int(size)
		case TypeInt:
			i++
			sizeLen := uint64(data[i])
			i++
			endI := i + int(sizeLen)
			n, err := decodeInt(data[i:endI])
			if err != nil {
				return wrapError(arrDecErr, err.Error())
			}
			entryVal = reflect.ValueOf(n)
			i = endI
		case TypeUint:
			i++
			sizeLen := uint64(data[i])
			i++
			endI := i + int(sizeLen)
			n, err := decodeUint(data[i:endI])
			if err != nil {
				return wrapError(arrDecErr, err.Error())
			}
			entryVal = reflect.ValueOf(n)
			i = endI
		case TypeFloat:
			i++
			sizeLen := uint64(data[i])
			i++
			endI := i + int(sizeLen)
			n, err := decodeFloat(data[i:endI])
			if err != nil {
				return wrapError(arrDecErr, err.Error())
			}
			entryVal = reflect.ValueOf(n)

			i = endI
		case TypeObject:
			panic("setup decoder for object")
		}

		if !entryVal.Type().AssignableTo(elem.Type().Elem()) {
			return wrapError(arrDecErr, "item type does not match slice element type")
		}
		elem.Set(reflect.Append(elem, entryVal))
	}

	return nil
}
