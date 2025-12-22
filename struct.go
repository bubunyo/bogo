package bogo

// encodeStruct encodes a struct using the legacy parser approach
// This function is kept for backward compatibility with old encoding logic
func encodeStruct(v any) ([]byte, error) {
	parser := NewParser("bogo") // Use "bogo" tags for legacy compatibility
	fields := parser.ParseFields(v)
	
	// Convert FieldInfo map to map[string]any
	objMap := make(map[string]any)
	for key, field := range fields {
		objMap[key] = field.Value
	}
	
	return encodeObject(objMap)
}