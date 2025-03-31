package bogo

import (
	"fmt"
	"reflect"
)

const TagName = "bogo"

// Parser extracts fields with a specific tag
type Parser struct{}

// FieldInfo holds the field name, type, and value
// type FieldInfo struct {
// 	FieldName string
// 	FieldType string
// 	Value     any
// }

// ParseFields returns a map of tag values to their corresponding field info
func (p *Parser) ParseFields(in interface{}) map[string]FieldInfo {
	result := make(map[string]FieldInfo)
	p.extractFields(reflect.ValueOf(in), reflect.TypeOf(in), &result)
	return result
}

// extractFields is a helper function to process struct fields, including nested structs and slices
func (p *Parser) extractFields(v reflect.Value, t reflect.Type, result *map[string]FieldInfo) {
	// Ensure input is a struct
	if t.Kind() != reflect.Struct {
		return
	}

	// Iterate through the struct fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Handle embedded structs (anonymous fields)
		if field.Anonymous && fieldValue.Kind() == reflect.Struct {
			p.extractFields(fieldValue, fieldValue.Type(), result)
			continue
		}

		// Get tag value
		tagValue := field.Tag.Get(TagName)
		key := []byte(tagValue)
		if len(key) == 0 {
			key = []byte(field.Name)
		}

		// Handle struct fields (nested structs)
		if fieldValue.Kind() == reflect.Struct {
			p.extractFields(fieldValue, fieldValue.Type(), result)
			continue
		}

		// Handle slices of structs
		if fieldValue.Kind() == reflect.Slice && fieldValue.Type().Elem().Kind() == reflect.Struct {
			for j := 0; j < fieldValue.Len(); j++ {
				element := fieldValue.Index(j)
				p.extractFields(element, element.Type(), result)
			}
			continue
		}

		var fieldType Type

		switch field.Type.Kind() {
		case reflect.String:
			fieldType = TypeString
		case reflect.Int:
			fieldType = TypeString
		default:
			fmt.Println(">>>>>>>>>>>> unknown type", field.Type.String())
		}

		// Add field to result
		// (*result)[key] = FieldInfo{
		(*result)[""] = FieldInfo{
			Key:   key,
			Type:  fieldType,
			Value: fieldValue.Interface(),
		}
	}
}
