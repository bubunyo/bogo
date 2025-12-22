package bogo

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test streaming API compatibility with json package patterns
func TestStreamingAPICompatibility(t *testing.T) {
	t.Run("NewEncoder and Encoder.Encode like json", func(t *testing.T) {
		var buf bytes.Buffer
		encoder := NewEncoder(&buf)
		
		// Test encoding like json.NewEncoder(w).Encode(v)
		err := encoder.Encode("hello world")
		require.NoError(t, err)
		
		// Verify the data was written to the buffer
		encoded := buf.Bytes()
		assert.NotEmpty(t, encoded)
		
		// Should match direct encoding
		expected, err := Encode("hello world")
		require.NoError(t, err)
		assert.Equal(t, expected, encoded)
	})
	
	t.Run("NewDecoder and Decoder.Decode like json", func(t *testing.T) {
		// Prepare test data
		testData := map[string]any{
			"name": "Alice",
			"age":  int64(30),
			"active": true,
		}
		encoded, err := Encode(testData)
		require.NoError(t, err)
		
		// Test decoding like json.NewDecoder(r).Decode(&v)
		reader := bytes.NewReader(encoded)
		decoder := NewDecoder(reader)
		
		var result map[string]any
		err = decoder.Decode(&result)
		require.NoError(t, err)
		
		// Verify the decoded result
		assert.Equal(t, "Alice", result["name"])
		assert.Equal(t, int64(30), result["age"])
		assert.Equal(t, true, result["active"])
	})
	
	t.Run("Round trip streaming", func(t *testing.T) {
		testCases := []any{
			"test string",
			int64(42),
			true,
			[]byte{1, 2, 3, 4, 5},
			[]string{"a", "b", "c"},
			map[string]any{"key": "value", "number": int64(123)},
		}
		
		for _, testCase := range testCases {
			// Encode using streaming encoder
			var encodeBuf bytes.Buffer
			encoder := NewEncoder(&encodeBuf)
			err := encoder.Encode(testCase)
			require.NoError(t, err, "Failed to encode %T: %v", testCase, testCase)
			
			// Decode using streaming decoder  
			decoder := NewDecoder(&encodeBuf)
			var result any
			err = decoder.Decode(&result)
			require.NoError(t, err, "Failed to decode %T: %v", testCase, testCase)
			
			// Verify round trip
			assert.Equal(t, testCase, result, "Round trip failed for %T", testCase)
		}
	})
	
	t.Run("Multiple values in stream", func(t *testing.T) {
		// Note: The current streaming implementation reads all data at once
		// This test demonstrates individual encode/decode cycles
		values := []any{"first", int64(2), "third"}
		
		for i, expected := range values {
			// Encode each value individually
			var buf bytes.Buffer
			encoder := NewEncoder(&buf)
			err := encoder.Encode(expected)
			require.NoError(t, err)
			
			// Decode each value individually
			decoder := NewDecoder(&buf)
			var result any
			err = decoder.Decode(&result)
			require.NoError(t, err, "Failed to decode value %d", i)
			assert.Equal(t, expected, result, "Value %d mismatch", i)
		}
	})
	
	t.Run("Encoder with custom options", func(t *testing.T) {
		var buf bytes.Buffer
		encoder := NewEncoderWithOptions(&buf, WithStrictMode(true), WithCompactArrays(false))
		
		testData := []int{1, 2, 3, 4, 5}
		err := encoder.Encode(testData)
		require.NoError(t, err)
		
		// Verify data was written
		assert.NotEmpty(t, buf.Bytes())
	})
	
	t.Run("Decoder with custom options", func(t *testing.T) {
		// Encode test data
		testData := "hello, 世界"
		encoded, err := Encode(testData)
		require.NoError(t, err)
		
		// Decode with custom options
		reader := bytes.NewReader(encoded)
		decoder := NewDecoderWithOptions(reader, WithUTF8Validation(true), WithDecoderStrictMode(true))
		
		var result string
		err = decoder.Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, testData, result)
	})
	
	t.Run("Error handling like json", func(t *testing.T) {
		// Test with invalid data
		reader := strings.NewReader("invalid data")
		decoder := NewDecoder(reader)
		
		var result any
		err := decoder.Decode(&result)
		assert.Error(t, err)
		// The error could be about unsupported type or insufficient data
		assert.True(t, strings.Contains(err.Error(), "insufficient data") || 
			strings.Contains(err.Error(), "unsupported type"), 
			"Expected error about insufficient data or unsupported type, got: %v", err)
	})
	
	t.Run("Type-specific decoding", func(t *testing.T) {
		testCases := []struct {
			input    any
			target   any
		}{
			{"hello", new(string)},
			{int64(42), new(int64)},
			{true, new(bool)},
			{[]byte{1, 2, 3}, new([]byte)},
		}
		
		for _, tc := range testCases {
			// Encode
			var buf bytes.Buffer
			encoder := NewEncoder(&buf)
			err := encoder.Encode(tc.input)
			require.NoError(t, err)
			
			// Decode to specific type
			decoder := NewDecoder(&buf)
			err = decoder.Decode(tc.target)
			require.NoError(t, err)
			
			// Extract the decoded value and compare
			switch target := tc.target.(type) {
			case *string:
				assert.Equal(t, tc.input, *target)
			case *int64:
				assert.Equal(t, tc.input, *target)
			case *bool:
				assert.Equal(t, tc.input, *target)
			case *[]byte:
				assert.Equal(t, tc.input, *target)
			}
		}
	})
}

// Test that streaming API follows json package naming and behavior
func TestStreamingAPINaming(t *testing.T) {
	t.Run("Constructor names match json", func(t *testing.T) {
		var buf bytes.Buffer
		
		// Should work like json.NewEncoder(w)
		encoder := NewEncoder(&buf)
		assert.NotNil(t, encoder)
		
		reader := bytes.NewReader([]byte{})
		
		// Should work like json.NewDecoder(r)  
		decoder := NewDecoder(reader)
		assert.NotNil(t, decoder)
	})
	
	t.Run("Method names match json", func(t *testing.T) {
		var buf bytes.Buffer
		encoder := NewEncoder(&buf)
		
		// Should work like encoder.Encode(v)
		err := encoder.Encode("test")
		assert.NoError(t, err)
		
		decoder := NewDecoder(&buf)
		var result string
		
		// Should work like decoder.Decode(&v) 
		err = decoder.Decode(&result)
		assert.NoError(t, err)
		assert.Equal(t, "test", result)
	})
}

// Benchmark streaming vs direct encoding
func BenchmarkStreamingEncoding(b *testing.B) {
	testData := map[string]any{
		"name":    "John Doe",
		"age":     int64(30),
		"active":  true,
		"scores":  []int{95, 87, 92, 88, 91},
		"details": map[string]any{"city": "New York", "country": "USA"},
	}
	
	b.Run("Direct encoding", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Encode(testData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("Streaming encoding", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			encoder := NewEncoder(&buf)
			err := encoder.Encode(testData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkStreamingDecoding(b *testing.B) {
	// Prepare test data
	testData := map[string]any{
		"name":   "John Doe",
		"age":    int64(30),
		"active": true,
	}
	encoded, err := Encode(testData)
	if err != nil {
		b.Fatal(err)
	}
	
	b.Run("Direct decoding", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Decode(encoded)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("Streaming decoding", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reader := bytes.NewReader(encoded)
			decoder := NewDecoder(reader)
			var result any
			err := decoder.Decode(&result)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}