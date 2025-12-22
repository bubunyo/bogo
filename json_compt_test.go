package bogo

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that demonstrates API compatibility with encoding/json package
func TestJSONAPICompatibility(t *testing.T) {
	t.Run("Marshal/Unmarshal API compatibility", func(t *testing.T) {
		testData := map[string]any{
			"name":    "Alice",
			"age":     int64(30), // Use int64 directly for bogo
			"active":  true,
			"scores":  []int64{95, 87, 92, 88, 91},
			"details": map[string]any{"city": "New York", "country": "USA"},
		}

		// JSON way
		jsonData, err := json.Marshal(testData)
		require.NoError(t, err)

		var jsonResult map[string]any
		err = json.Unmarshal(jsonData, &jsonResult)
		require.NoError(t, err)

		// Bogo way (same API signatures)
		bogoData, err := Marshal(testData)
		require.NoError(t, err)

		var bogoResult map[string]any
		err = Unmarshal(bogoData, &bogoResult)
		require.NoError(t, err)

		// Results should have the same structure
		assert.Equal(t, jsonResult["name"], bogoResult["name"])
		assert.Equal(t, jsonResult["active"], bogoResult["active"])

		// Note: JSON numbers become float64, bogo preserves int64
		// This is expected difference in type handling
		assert.Equal(t, float64(30), jsonResult["age"])
		assert.Equal(t, int64(30), bogoResult["age"])

		// Note: Size comparison varies by data type and complexity
		// Binary formats aren't always smaller for small, complex nested data
		t.Logf("JSON size: %d bytes, Bogo size: %d bytes", len(jsonData), len(bogoData))
	})

	t.Run("Streaming API compatibility", func(t *testing.T) {
		testData := map[string]any{
			"message": "Hello, World!",
			"count":   42,
			"enabled": true,
		}

		// JSON streaming
		var jsonBuf bytes.Buffer
		jsonEncoder := json.NewEncoder(&jsonBuf)
		err := jsonEncoder.Encode(testData)
		require.NoError(t, err)

		jsonDecoder := json.NewDecoder(&jsonBuf)
		var jsonResult map[string]any
		err = jsonDecoder.Decode(&jsonResult)
		require.NoError(t, err)

		// Bogo streaming (identical API)
		var bogoBuf bytes.Buffer
		bogoEncoder := NewEncoder(&bogoBuf)
		err = bogoEncoder.Encode(testData)
		require.NoError(t, err)

		bogoDecoder := NewDecoder(&bogoBuf)
		var bogoResult map[string]any
		err = bogoDecoder.Decode(&bogoResult)
		require.NoError(t, err)

		// Both should decode successfully
		assert.Equal(t, jsonResult["message"], bogoResult["message"])
		assert.Equal(t, jsonResult["enabled"], bogoResult["enabled"])

		// Again, number type difference is expected
		assert.Equal(t, float64(42), jsonResult["count"])
		assert.Equal(t, int64(42), bogoResult["count"])
	})

	t.Run("Type-specific unmarshaling", func(t *testing.T) {
		// Test unmarshaling into specific types
		testString := "Hello, World!"
		testInt := int64(42)
		testBool := true

		// String test
		bogoStringData, err := Marshal(testString)
		require.NoError(t, err)

		var resultString string
		err = Unmarshal(bogoStringData, &resultString)
		require.NoError(t, err)
		assert.Equal(t, testString, resultString)

		// Int test
		bogoIntData, err := Marshal(testInt)
		require.NoError(t, err)

		var resultInt int64
		err = Unmarshal(bogoIntData, &resultInt)
		require.NoError(t, err)
		assert.Equal(t, testInt, resultInt)

		// Bool test
		bogoBoolData, err := Marshal(testBool)
		require.NoError(t, err)

		var resultBool bool
		err = Unmarshal(bogoBoolData, &resultBool)
		require.NoError(t, err)
		assert.Equal(t, testBool, resultBool)
	})

	t.Run("Error handling compatibility", func(t *testing.T) {
		// Test invalid input handling
		invalidData := []byte("not valid data")

		// Both should return errors
		var jsonResult any
		jsonErr := json.Unmarshal(invalidData, &jsonResult)
		assert.Error(t, jsonErr)

		var bogoResult any
		bogoErr := Unmarshal(invalidData, &bogoResult)
		assert.Error(t, bogoErr)
	})

	t.Run("Nil pointer error", func(t *testing.T) {
		// Test nil pointer handling
		testData := "test"
		bogoData, err := Marshal(testData)
		require.NoError(t, err)

		// Both should error on nil pointer
		var jsonResult *string
		jsonErr := json.Unmarshal([]byte(`"test"`), jsonResult)
		assert.Error(t, jsonErr)

		var bogoResult *string
		bogoErr := Unmarshal(bogoData, bogoResult)
		assert.Error(t, bogoErr)
		assert.Contains(t, bogoErr.Error(), "non-nil pointer")
	})
}

// Test demonstrating the API usage similarity
func TestAPIUsageSimilarity(t *testing.T) {
	t.Run("Example usage comparison", func(t *testing.T) {
		type Person struct {
			Name   string `json:"name"`
			Age    int    `json:"age"`
			Active bool   `json:"active"`
		}

		person := Person{
			Name:   "John Doe",
			Age:    30,
			Active: true,
		}

		// Convert struct to map for bogo (since bogo doesn't use struct tags)
		personMap := map[string]any{
			"name":   person.Name,
			"age":    int64(person.Age), // Explicit int64 for bogo
			"active": person.Active,
		}

		// JSON usage pattern
		jsonData, err := json.Marshal(person)
		require.NoError(t, err)

		var jsonPerson Person
		err = json.Unmarshal(jsonData, &jsonPerson)
		require.NoError(t, err)

		// Bogo usage pattern (nearly identical)
		bogoData, err := Marshal(personMap)
		require.NoError(t, err)

		var bogoPersonMap map[string]any
		err = Unmarshal(bogoData, &bogoPersonMap)
		require.NoError(t, err)

		// Verify data integrity
		assert.Equal(t, person.Name, bogoPersonMap["name"])
		assert.Equal(t, int64(person.Age), bogoPersonMap["age"])
		assert.Equal(t, person.Active, bogoPersonMap["active"])
	})

	t.Run("Streaming usage comparison", func(t *testing.T) {
		data := []map[string]any{
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"},
			{"id": 3, "name": "Charlie"},
		}

		// JSON streaming pattern
		var jsonBuf bytes.Buffer
		jsonEnc := json.NewEncoder(&jsonBuf)
		for _, item := range data {
			err := jsonEnc.Encode(item)
			require.NoError(t, err)
		}

		jsonDec := json.NewDecoder(strings.NewReader(jsonBuf.String()))
		var jsonItems []map[string]any
		for i := 0; i < len(data); i++ {
			var item map[string]any
			err := jsonDec.Decode(&item)
			require.NoError(t, err)
			jsonItems = append(jsonItems, item)
		}

		// Bogo streaming pattern (identical structure)
		var bogoItems []map[string]any
		for _, item := range data {
			var bogoBuf bytes.Buffer
			bogoEnc := NewEncoder(&bogoBuf)
			err := bogoEnc.Encode(item)
			require.NoError(t, err)

			bogoDec := NewDecoder(&bogoBuf)
			var bogoItem map[string]any
			err = bogoDec.Decode(&bogoItem)
			require.NoError(t, err)
			bogoItems = append(bogoItems, bogoItem)
		}

		// Results should have same structure
		require.Equal(t, len(jsonItems), len(bogoItems))
		for i := range jsonItems {
			assert.Equal(t, jsonItems[i]["name"], bogoItems[i]["name"])
			// JSON converts numbers to float64, bogo preserves int64
			assert.Equal(t, float64(i+1), jsonItems[i]["id"])
			assert.Equal(t, int64(i+1), bogoItems[i]["id"])
		}
	})
}

// Benchmark JSON vs Bogo performance
func BenchmarkJSONvsBogo(b *testing.B) {
	testData := map[string]any{
		"name":    "John Doe",
		"age":     int64(30),
		"active":  true,
		"scores":  []int64{95, 87, 92, 88, 91},
		"details": map[string]any{"city": "New York", "country": "USA"},
	}

	b.Run("JSON Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(testData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Bogo Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Marshal(testData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Prepare data for unmarshal benchmarks
	jsonData, _ := json.Marshal(testData)
	bogoData, _ := Marshal(testData)

	b.Run("JSON Unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result map[string]any
			err := json.Unmarshal(jsonData, &result)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Bogo Unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result map[string]any
			err := Unmarshal(bogoData, &result)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
