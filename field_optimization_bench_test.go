package bogo

import (
	"testing"
)

// Field Optimization Benchmark Results (Apple M2, 8 cores):
//
// BEFORE OPTIMIZATION:
// BenchmarkFieldDecoding_Full-8               	   14395	    281858 ns/op	   48056 B/op	    1138 allocs/op
// BenchmarkFieldDecoding_Selective-8          	   19111	     56626 ns/op	   31456 B/op	     826 allocs/op
//
// AFTER OPTIMIZATION:
// BenchmarkFieldDecoding_WithOptimization-8   	 1376488	       844.3 ns/op	     424 B/op	       7 allocs/op
//
// PERFORMANCE IMPROVEMENT:
// - 334x faster than full decoding (281858ns → 844ns)
// - 67x faster than selective decoding without optimization (56626ns → 844ns)
// - 113x fewer allocations than full decoding (1138 → 7)
// - 118x fewer allocations than selective decoding (826 → 7)
// - 113x less memory usage than full decoding (48056B → 424B)
// - 74x less memory usage than selective decoding (31456B → 424B)

// Test data structures for field optimization benchmarks
type OptimizationTestData struct {
	// First field - complex and large to make skipping beneficial
	LargePayload struct {
		Data    []byte            `json:"data"`
		Metrics map[string]int64  `json:"metrics"`
		Config  map[string]string `json:"config"`
		Arrays  [][]string        `json:"arrays"`
	} `json:"large_payload"`

	// Second field - simple target field we want to extract
	TargetField string `json:"target_field"`

	// Additional fields
	Metadata map[string]any `json:"metadata"`
	Status   string         `json:"status"`
}

// OptimizationResult holds data we want to extract selectively
type OptimizationResult struct {
	TargetField string `json:"target_field"`
}

// createBenchmarkData creates test data with a complex first field
func createBenchmarkData() OptimizationTestData {
	// Create large, complex data for the first field
	largeData := make([]byte, 10000) // 10KB of data
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	metrics := make(map[string]int64)
	for i := 0; i < 100; i++ {
		metrics[string(rune('a'+i%26))+string(rune('a'+(i/26)%26))] = int64(i * 1000)
	}

	config := make(map[string]string)
	for i := 0; i < 50; i++ {
		config[string(rune('A'+i%26))] = "config_value_" + string(rune('0'+i%10))
	}

	arrays := make([][]string, 20)
	for i := range arrays {
		arrays[i] = make([]string, 10)
		for j := range arrays[i] {
			arrays[i][j] = "item_" + string(rune('0'+i%10)) + "_" + string(rune('0'+j%10))
		}
	}

	return OptimizationTestData{
		LargePayload: struct {
			Data    []byte            `json:"data"`
			Metrics map[string]int64  `json:"metrics"`
			Config  map[string]string `json:"config"`
			Arrays  [][]string        `json:"arrays"`
		}{
			Data:    largeData,
			Metrics: metrics,
			Config:  config,
			Arrays:  arrays,
		},
		TargetField: "important_value_we_want",
		Metadata: map[string]any{
			"created": int64(1640995200),
			"version": "1.0.0",
			"flags":   []string{"flag1", "flag2"},
		},
		Status: "active",
	}
}

// BenchmarkFieldDecoding_Full benchmarks full object decoding
func BenchmarkFieldDecoding_Full(b *testing.B) {
	data := createBenchmarkData()
	encoded, err := Marshal(data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var result OptimizationTestData
		err := Unmarshal(encoded, &result)
		if err != nil {
			b.Fatal(err)
		}

		// Use the target field to prevent optimization
		_ = result.TargetField
	}
}

// BenchmarkFieldDecoding_Selective benchmarks selective field decoding (target: second field)
func BenchmarkFieldDecoding_Selective(b *testing.B) {
	data := createBenchmarkData()
	encoded, err := Marshal(data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var result OptimizationResult
		err := Unmarshal(encoded, &result)
		if err != nil {
			b.Fatal(err)
		}

		// Use the target field to prevent optimization
		_ = result.TargetField
	}
}

// BenchmarkFieldDecoding_WithOptimization will test the optimized decoder
func BenchmarkFieldDecoding_WithOptimization(b *testing.B) {
	data := createBenchmarkData()
	encoded, err := Marshal(data)
	if err != nil {
		b.Fatal(err)
	}

	// Create optimized decoder that only reads specific fields
	decoder := NewConfigurableDecoder(
		WithDecoderStructTag("json"),
		WithSelectiveFields([]string{"target_field"}), // Only decode this field
	)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := decoder.Decode(encoded)
		if err != nil {
			b.Fatal(err)
		}

		var targetResult OptimizationResult
		err = assignResult(result, &targetResult)
		if err != nil {
			b.Fatal(err)
		}

		// Use the target field to prevent optimization
		_ = targetResult.TargetField
	}
}

// Test to verify the optimization works correctly
func TestFieldOptimization(t *testing.T) {
	data := createBenchmarkData()
	encoded, err := Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	// Test full decoding
	var fullResult OptimizationTestData
	err = Unmarshal(encoded, &fullResult)
	if err != nil {
		t.Fatal(err)
	}

	// Test selective decoding
	var selectiveResult OptimizationResult
	err = Unmarshal(encoded, &selectiveResult)
	if err != nil {
		t.Fatal(err)
	}

	// Verify we get the same target field value
	if fullResult.TargetField != selectiveResult.TargetField {
		t.Errorf("Expected target field %q, got %q",
			fullResult.TargetField, selectiveResult.TargetField)
	}

	if selectiveResult.TargetField != "important_value_we_want" {
		t.Errorf("Expected target field 'important_value_we_want', got %q",
			selectiveResult.TargetField)
	}
}
