package bogo

import (
	"encoding/binary"
	"fmt"
	"testing"
)

const MaxUint = ^uint32(0)
const MinUint = 0

// Example structs with embedded fields, nested structs, and slices
type Address struct {
	City    string `bogo:"city"`
	Country string `bogo:"country"`
}

type Base struct {
	ID   int    `bogo:"base_id"`
	Note string // No tag, should default to "Note"
}

type Example struct {
	// Base            // Embedded struct
	Name string `bogo:"username"`
	Age  int    `bogo:"age"`
	// Email   string  // No tag, should default to "Email"`
	// Score   float64 `bogo:"score"`
	// Address Address // Nested struct
	// Friends []Base  // Slice of structs
}

func TestParser(t *testing.T) {
	ex := Example{
		// Base:    Base{ID: 101, Note: "Embedded struct"},
		Name: "John",
		Age:  30,
		// Email:   "john@example.com",
		// Score:   95.5,
		// Address: Address{City: "Berlin", Country: "Germany"},
		// Friends: []Base{
		// 	{ID: 201, Note: "Friend 1"},
		// 	{ID: 202, Note: "Friend 2"},
		// },
	}

	parser := Parser{}
	taggedFields := parser.ParseFields(ex)

	_ = taggedFields

	// Print extracted fields
	// for key, info := range taggedFields {
	// fmt.Printf("Key: %s -> Field: %s, Type: %v, Value: %v\n", key, info.FieldName, info.FieldType, info.Value)
	// }
}

func TestRandom(t *testing.T) {
	bs := make([]byte, 4)
	fmt.Println(MaxUint)
	binary.LittleEndian.PutUint32(bs, MaxUint)
	fmt.Println(bs)
}
