package bogo

func UnMarshall() {}
func Marshall()   {}

type Type byte

// Type constants
const (
	maxStorageByteLength = 5
	// Version helps determine which encoders/decoders to use
	Version byte = 0x00 // version 0
)

const (
	TypeNull = iota
	TypeBoolTrue
	TypeBoolFalse
	TypeString
	TypeByte
	TypeInt
	TypeUint
	TypeFloat
	TypeAny

	TypeArray
	TypeObject
)

func (t Type) String() string {
	switch t {
	case TypeNull:
		return "<null>"
	case TypeBoolTrue:
		return "<bool:true>"
	case TypeBoolFalse:
		return "<bool:false>"
	case TypeString:
		return "<string>"
	case TypeArray:
		return "<array>"
	case TypeObject:
		return "<object>"
	case TypeByte:
		return "<byte>"
	case TypeInt:
		return "<int>"
	case TypeUint:
		return "<uint>"
	case TypeFloat:
		return "<float>"
	}
	return "<unknown>"
}

type TypeNumber interface {
	int64 | float64
}

type FieldInfo struct {
	Type  Type
	Key   []byte
	Value any
}
