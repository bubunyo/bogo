package bogo_test

import (
	"fmt"
	"log"

	"github.com/bubunyo/bogo"
)

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func Example() {

	// Create a user
	user := User{
		ID:    12345,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	// Marshal to Bogo binary format
	data, err := bogo.Marshal(user)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Encoded %d bytes\n", len(data))

	// Unmarshal back to struct
	var decoded User
	err = bogo.Unmarshal(data, &decoded)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %s, Age: %d\n", decoded.Name, decoded.Age)
	// Output: Encoded 68 bytes
	// Name: John Doe, Age: 30
}

func ExampleEncoder() {
	// Advanced configuration for high-performance scenarios
	decoder := bogo.NewConfigurableDecoder(
		bogo.WithSelectiveFields([]string{"id", "name"}), // Only decode these fields
		bogo.WithDecoderMaxDepth(10),                     // Limit nesting
	)

	// Create some test data
	testData := map[string]any{
		"id":    int64(123),
		"name":  "Alice",
		"email": "alice@example.com", // This will be skipped
		"profile": map[string]any{ // This will be skipped
			"bio":      "Long biography...",
			"settings": map[string]any{"theme": "dark"},
		},
	}

	// Encode normally
	encoded, err := bogo.Marshal(testData)
	if err != nil {
		log.Fatal(err)
	}

	// Decode selectively (much faster for large objects)
	result, err := decoder.Decode(encoded)
	if err != nil {
		log.Fatal(err)
	}

	resultMap := result.(map[string]any)
	fmt.Printf("Selective decode - ID: %v, Name: %v\n", resultMap["id"], resultMap["name"])
	// Output: Selective decode - ID: 123, Name: Alice
}

