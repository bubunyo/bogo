package bogo

import (
	"fmt"
	"testing"
)

func TestDebugStructEncoding(t *testing.T) {
	type Person struct {
		Name string `bogo:"name"`
		Age  int    `bogo:"age"`
		City string
	}

	person := Person{Name: "John", Age: 30, City: "New York"}

	// Debug: Check what ParseFields returns
	parser := NewParser("bogo")
	fields := parser.ParseFields(person)
	fmt.Printf("ParseFields result: %+v\n", fields)

	// Debug: Check encoding
	encoded, err := Encode(person)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	fmt.Printf("Encoded length: %d bytes\n", len(encoded))

	// Debug: Check decoding
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	fmt.Printf("Decoded type: %T\n", decoded)
	fmt.Printf("Decoded value: %+v\n", decoded)
}