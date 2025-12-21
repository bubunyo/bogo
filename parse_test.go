package bogo

import (
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
		Name: "John",
		Age:  30,
	}

	parser := NewParser("bogo")
	taggedFields := parser.ParseFields(ex)

	// Check that we got the expected fields
	if len(taggedFields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(taggedFields))
	}

	// Check username field
	if field, exists := taggedFields["username"]; exists {
		if field.Value != "John" {
			t.Errorf("Expected username value 'John', got %v", field.Value)
		}
		if field.Type != TypeString {
			t.Errorf("Expected username type %v, got %v", TypeString, field.Type)
		}
	} else {
		t.Error("Expected 'username' field not found")
	}

	// Check age field
	if field, exists := taggedFields["age"]; exists {
		if field.Value != 30 {
			t.Errorf("Expected age value 30, got %v", field.Value)
		}
		if field.Type != TypeInt {
			t.Errorf("Expected age type %v, got %v", TypeInt, field.Type)
		}
	} else {
		t.Error("Expected 'age' field not found")
	}
}
