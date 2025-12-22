package bogo

import (
	"reflect"
	"time"
)

// Parser extracts struct field information for encoding/decoding
type Parser struct {
	TagName string // Configurable tag name (e.g., "json", "bogo")
}

// NewParser creates a new parser with the specified tag name
func NewParser(tagName string) *Parser {
	if tagName == "" {
		tagName = "json" // Default to json for compatibility
	}
	return &Parser{TagName: tagName}
}

// ParseFields returns a map of field names to their corresponding field info
func (p *Parser) ParseFields(in interface{}) map[string]FieldInfo {
	result := make(map[string]FieldInfo)
	p.extractFields(reflect.ValueOf(in), reflect.TypeOf(in), &result)
	return result
}

// extractFields is a helper function to process struct fields safely
func (p *Parser) extractFields(v reflect.Value, t reflect.Type, result *map[string]FieldInfo) {
	// Handle pointers
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
		t = t.Elem()
	}

	// Ensure input is a struct
	if t.Kind() != reflect.Struct {
		return
	}

	// Iterate through the struct fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldValue := v.Field(i)

		// Handle embedded structs (anonymous fields)
		if field.Anonymous && fieldValue.Kind() == reflect.Struct {
			p.extractFields(fieldValue, fieldValue.Type(), result)
			continue
		}

		// Get field name from tag or use field name
		fieldName := p.getFieldName(field)
		
		// Skip if tag indicates to omit the field
		if fieldName == "-" {
			continue
		}

		// Skip zero values if omitempty is specified
		if p.shouldOmitEmpty(field) && p.isZeroValue(fieldValue) {
			continue
		}

		// Determine the bogo type for this field
		fieldType := p.getBogoType(field.Type)

		// Note: We handle structs as objects, not as special cases
		// time.Time is already handled by getBogoType returning TypeTimestamp

		// Add field to result
		(*result)[fieldName] = FieldInfo{
			Key:   []byte(fieldName),
			Type:  fieldType,
			Value: fieldValue.Interface(),
		}
	}
}

// getFieldName returns the field name to use based on struct tags
func (p *Parser) getFieldName(field reflect.StructField) string {
	tag := field.Tag.Get(p.TagName)
	if tag == "" {
		return field.Name
	}
	
	// Handle "fieldname" and "fieldname,omitempty" formats
	for i, c := range tag {
		if c == ',' {
			return tag[:i]
		}
	}
	
	return tag
}

// shouldOmitEmpty checks if the field has omitempty tag
func (p *Parser) shouldOmitEmpty(field reflect.StructField) bool {
	tag := field.Tag.Get(p.TagName)
	return len(tag) > 10 && tag[len(tag)-10:] == ",omitempty"
}

// isZeroValue reports whether v is the zero value for its type
func (p *Parser) isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		// For structs, check if it's the zero value
		return v.IsZero()
	}
	return false
}

// getBogoType maps Go types to bogo types
func (p *Parser) getBogoType(t reflect.Type) Type {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Special cases first
	if t == reflect.TypeOf(time.Time{}) {
		return TypeTimestamp
	}

	switch t.Kind() {
	case reflect.String:
		return TypeString
	case reflect.Bool:
		return TypeBoolTrue // We'll handle false during encoding
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return TypeInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if t.Kind() == reflect.Uint8 {
			return TypeByte
		}
		return TypeUint
	case reflect.Float32, reflect.Float64:
		return TypeFloat
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return TypeBlob // []byte
		}
		return TypeArray
	case reflect.Array:
		return TypeArray
	case reflect.Map:
		return TypeObject
	case reflect.Struct:
		return TypeObject
	case reflect.Interface:
		return TypeNull // Will be determined at runtime
	default:
		return TypeNull // Unknown type
	}
}
