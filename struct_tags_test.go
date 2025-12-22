package bogo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test struct with various tag scenarios
type User struct {
	ID       int64     `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Active   bool      `json:"active"`
	Balance  float64   `json:"balance"`
	Created  time.Time `json:"created"`
	Tags     []string  `json:"tags"`
	Settings map[string]any `json:"settings"`
	
	// Test omitempty
	OptionalField string `json:"optional,omitempty"`
	
	// Test field without tag (should use field name)
	NoTag string
	
	// Test ignored field
	IgnoredField string `json:"-"`
	
	// Private field (should be ignored)
	privateField string `json:"private"`
}

type NestedStruct struct {
	User   User   `json:"user"`
	Status string `json:"status"`
	Count  int    `json:"count"`
}

func TestStructTagsBasicEncoding(t *testing.T) {
	user := User{
		ID:       12345,
		Username: "testuser",
		Email:    "test@example.com", 
		Active:   true,
		Balance:  99.99,
		Created:  time.Unix(1609459200, 0), // 2021-01-01 00:00:00
		Tags:     []string{"admin", "user"},
		Settings: map[string]any{"theme": "dark", "notifications": true},
		NoTag:    "no tag value",
		OptionalField: "", // empty, should be omitted if omitempty works
		IgnoredField:  "should be ignored",
		privateField:  "private value",
	}

	// Test encoding
	data, err := Marshal(user)
	if err != nil {
		t.Logf("Encoding error: %v", err)
	}
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Test decoding
	var decoded User
	err = Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify all tagged fields are preserved
	assert.Equal(t, user.ID, decoded.ID)
	assert.Equal(t, user.Username, decoded.Username)
	assert.Equal(t, user.Email, decoded.Email)
	assert.Equal(t, user.Active, decoded.Active)
	assert.Equal(t, user.Balance, decoded.Balance)
	
	
	assert.Equal(t, user.Created.Unix(), decoded.Created.Unix()) // Compare Unix timestamp
	assert.Equal(t, user.Tags, decoded.Tags)
	assert.Equal(t, user.Settings, decoded.Settings)
	
	// Field without tag should use field name
	assert.Equal(t, user.NoTag, decoded.NoTag)
	
	// Ignored field should not be preserved
	assert.Empty(t, decoded.IgnoredField)
	
	// Private field should not be preserved  
	assert.Empty(t, decoded.privateField)
}

func TestStructTagsNestedStructs(t *testing.T) {
	nested := NestedStruct{
		User: User{
			ID:       999,
			Username: "nested",
			Email:    "nested@test.com",
			Active:   false,
			Balance:  123.45,
			Created:  time.Unix(1640995200, 0), // 2022-01-01 00:00:00
			Tags:     []string{"nested", "test"},
			Settings: map[string]any{"lang": "en"},
			NoTag:    "nested no tag",
		},
		Status: "active",
		Count:  42,
	}

	// Test encoding
	data, err := Marshal(nested)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Test decoding
	var decoded NestedStruct
	err = Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify nested struct fields
	assert.Equal(t, nested.Status, decoded.Status)
	assert.Equal(t, nested.Count, decoded.Count)
	assert.Equal(t, nested.User.ID, decoded.User.ID)
	assert.Equal(t, nested.User.Username, decoded.User.Username)
	assert.Equal(t, nested.User.Email, decoded.User.Email)
	assert.Equal(t, nested.User.Active, decoded.User.Active)
	assert.Equal(t, nested.User.Balance, decoded.User.Balance)
	assert.Equal(t, nested.User.Tags, decoded.User.Tags)
	assert.Equal(t, nested.User.Settings, decoded.User.Settings)
	assert.Equal(t, nested.User.NoTag, decoded.User.NoTag)
}

func TestStructTagsCustomTag(t *testing.T) {
	// Define struct with custom tags
	type CustomTagStruct struct {
		Name  string `bogo:"name"`
		Age   int    `bogo:"age"`
		Email string `bogo:"email"`
	}

	original := CustomTagStruct{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	// Create encoder with custom tag
	encoder := NewConfigurableEncoder(WithStructTag("bogo"))
	decoder := NewConfigurableDecoder(WithDecoderStructTag("bogo"))

	// Test encoding
	data, err := encoder.Encode(original)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Test decoding
	result, err := decoder.Decode(data)
	require.NoError(t, err)

	// Convert result to struct using custom decoder
	var decoded CustomTagStruct
	SetDefaultDecoder(decoder)
	defer SetDefaultDecoder(NewConfigurableDecoder()) // Reset to default
	
	err = assignResult(result, &decoded)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, original.Name, decoded.Name)
	assert.Equal(t, int64(original.Age), int64(decoded.Age)) // Handle int/int64 conversion
	assert.Equal(t, original.Email, decoded.Email)
}

func TestStructTagsOmitEmpty(t *testing.T) {
	type OmitEmptyStruct struct {
		Required string `json:"required"`
		Optional string `json:"optional,omitempty"`
		AlwaysOmit string `json:"-"`
	}

	// Test with empty optional field
	original := OmitEmptyStruct{
		Required:   "required value",
		Optional:   "", // Empty, should be omitted
		AlwaysOmit: "never included",
	}

	data, err := Marshal(original)
	require.NoError(t, err)

	var decoded OmitEmptyStruct
	err = Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Required, decoded.Required)
	assert.Empty(t, decoded.Optional) // Should remain empty
	assert.Empty(t, decoded.AlwaysOmit) // Should be empty (omitted)

	// Test with non-empty optional field
	original.Optional = "optional value"
	
	data, err = Marshal(original)
	require.NoError(t, err)

	err = Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Required, decoded.Required)
	assert.Equal(t, original.Optional, decoded.Optional) // Should be preserved
	assert.Empty(t, decoded.AlwaysOmit) // Still omitted
}

func TestStructTagsPointers(t *testing.T) {
	type PointerStruct struct {
		Name  *string `json:"name"`
		Age   *int    `json:"age"`
		Email *string `json:"email"`
	}

	name := "John Doe"
	age := 30
	email := "john@example.com"

	original := PointerStruct{
		Name:  &name,
		Age:   &age,
		Email: &email,
	}

	data, err := Marshal(original)
	require.NoError(t, err)

	var decoded PointerStruct
	err = Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Check pointer values
	require.NotNil(t, decoded.Name)
	assert.Equal(t, *original.Name, *decoded.Name)
	require.NotNil(t, decoded.Age)
	assert.Equal(t, int64(*original.Age), int64(*decoded.Age))
	require.NotNil(t, decoded.Email)
	assert.Equal(t, *original.Email, *decoded.Email)

	// Test with nil pointers
	nilPointers := PointerStruct{
		Name:  nil,
		Age:   nil, 
		Email: nil,
	}

	data, err = Marshal(nilPointers)
	require.NoError(t, err)

	var decodedNil PointerStruct
	err = Unmarshal(data, &decodedNil)
	require.NoError(t, err)

	assert.Nil(t, decodedNil.Name)
	assert.Nil(t, decodedNil.Age)
	assert.Nil(t, decodedNil.Email)
}

func TestStructTagsSlicesAndMaps(t *testing.T) {
	type ComplexStruct struct {
		StringSlice []string           `json:"string_slice"`
		IntSlice    []int             `json:"int_slice"`
		FloatSlice  []float64         `json:"float_slice"`
		StringMap   map[string]string `json:"string_map"`
		AnyMap      map[string]any    `json:"any_map"`
	}

	original := ComplexStruct{
		StringSlice: []string{"a", "b", "c"},
		IntSlice:    []int{1, 2, 3},
		FloatSlice:  []float64{1.1, 2.2, 3.3},
		StringMap:   map[string]string{"key1": "value1", "key2": "value2"},
		AnyMap:      map[string]any{"number": int64(42), "text": "hello", "flag": true},
	}

	data, err := Marshal(original)
	require.NoError(t, err)

	var decoded ComplexStruct
	err = Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.StringSlice, decoded.StringSlice)
	// Note: IntSlice will be decoded as []int64, so we need to check differently
	require.Len(t, decoded.IntSlice, len(original.IntSlice))
	for i, v := range original.IntSlice {
		assert.Equal(t, int64(v), int64(decoded.IntSlice[i]))
	}
	assert.Equal(t, original.FloatSlice, decoded.FloatSlice)
	assert.Equal(t, original.StringMap, decoded.StringMap)
	assert.Equal(t, original.AnyMap, decoded.AnyMap)
}

func TestStructTagsJSONCompatibility(t *testing.T) {
	// Test that bogo can decode structs the same way as JSON package
	type JSONCompatStruct struct {
		ID       int     `json:"id"`
		Name     string  `json:"name"`
		Active   bool    `json:"active"`
		Balance  float64 `json:"balance"`
		Tags     []string `json:"tags"`
		Settings map[string]any `json:"settings"`
	}

	original := JSONCompatStruct{
		ID:      12345,
		Name:    "test user",
		Active:  true,
		Balance: 99.99,
		Tags:    []string{"admin", "user"},
		Settings: map[string]any{
			"theme": "dark",
			"notifications": true,
			"max_items": int64(100),
		},
	}

	// Encode and decode with bogo
	data, err := Marshal(original)
	require.NoError(t, err)

	var decoded JSONCompatStruct
	err = Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Name, decoded.Name)
	assert.Equal(t, original.Active, decoded.Active)
	assert.Equal(t, original.Balance, decoded.Balance)
	assert.Equal(t, original.Tags, decoded.Tags)
	assert.Equal(t, original.Settings, decoded.Settings)
}