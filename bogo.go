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
	TypeArray
	TypeObject
	TypeByte
	TypeInt
	TypeUint
	TypeFloat
)

type TypeNumber interface {
	int64 | float64
}

type FieldInfo struct {
	Type  Type
	Key   []byte
	Value any
}
