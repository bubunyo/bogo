# Bogo - High-Performance Binary Serialization for Go

Bogo is a fast, compact binary serialization format for Go with a JSON-compatible API. It provides efficient encoding and decoding of Go data types with performance that significantly outperforms standard JSON serialization.

## Features

- **JSON-Compatible API**: Drop-in replacement for `encoding/json` with familiar `Marshal`/`Unmarshal` functions
- **High Performance**: Up to 3x faster deserialization compared to JSON
- **Compact Binary Format**: Efficient variable-length encoding reduces payload size
- **Comprehensive Type Support**: Supports all primitive types plus arrays, objects, and binary data
- **Streaming API**: Memory-efficient streaming encoding and decoding
- **Configurable Validation**: Optional UTF-8 validation, depth limits, and size constraints
- **Field-Specific Optimization**: Revolutionary selective field decoding with up to 334x performance improvement and 113x memory reduction

## Supported Types

| Go Type | Bogo Type | Description |
|---------|-----------|-------------|
| `nil` | TypeNull | Null values |
| `bool` | TypeBoolTrue/TypeBoolFalse | Boolean values |
| `string` | TypeString | UTF-8 strings with variable-length encoding |
| `byte` | TypeByte | Single byte values |
| `int`, `int8`, `int16`, `int32`, `int64` | TypeInt | Signed integers with VarInt encoding |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | TypeUint | Unsigned integers with VarInt encoding |
| `float32`, `float64` | TypeFloat | IEEE 754 floating-point numbers |
| `[]byte` | TypeBlob | Binary data with length prefix |
| `time.Time` | TypeTimestamp | Unix timestamps |
| `[]interface{}` | TypeArray | Heterogeneous arrays |
| Typed slices | TypeTypedArray | Homogeneous typed arrays |
| `map[string]interface{}` | TypeObject | Key-value objects |

## Installation

```bash
go get github.com/bubunyo/bogo
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "github.com/bubunyo/bogo"
)

type User struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Active   bool   `json:"active"`
}

func main() {
    user := User{
        ID:     123,
        Name:   "John Doe",
        Email:  "john@example.com",
        Active: true,
    }

    // Marshal to binary format
    data, err := bogo.Marshal(user)
    if err != nil {
        log.Fatal(err)
    }

    // Unmarshal from binary format
    var decoded User
    err = bogo.Unmarshal(data, &decoded)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Original: %+v\n", user)
    fmt.Printf("Decoded:  %+v\n", decoded)
}
```

### Streaming API

```go
package main

import (
    "bytes"
    "log"
    "github.com/bubunyo/bogo"
)

func main() {
    data := map[string]interface{}{
        "message": "Hello, World!",
        "count":   42,
        "active":  true,
    }

    // Streaming encoding
    var buf bytes.Buffer
    encoder := bogo.NewEncoder(&buf)
    if err := encoder.Encode(data); err != nil {
        log.Fatal(err)
    }

    // Streaming decoding
    decoder := bogo.NewDecoder(&buf)
    var result map[string]interface{}
    if err := decoder.Decode(&result); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Streamed data: %+v\n", result)
}
```

## Performance

Bogo delivers significant performance improvements over JSON serialization:

### Benchmark Results

```
Simple Data (small objects):
- JSON Deserialize:     1,934 ns/op
- Bogo Deserialize:       560 ns/op  (3.45x faster)
- MessagePack:            625 ns/op

Complex Data (nested structures):
- JSON Serialize:       8,693 ns/op
- Bogo Serialize:       8,446 ns/op  (1.03x faster)
- MessagePack:          2,856 ns/op

- JSON Deserialize:    12,757 ns/op
- Bogo Deserialize:     4,429 ns/op  (2.88x faster)
- MessagePack:          5,001 ns/op
```

## Advanced Configuration

### Decoder Options

```go
decoder := bogo.NewConfigurableDecoder(
    bogo.WithDecoderMaxDepth(50),        // Limit nesting depth
    bogo.WithDecoderStrictMode(true),    // Enable strict validation
    bogo.WithMaxObjectSize(1024*1024),   // 1MB size limit
    bogo.WithUTF8Validation(true),       // Validate UTF-8 strings
    bogo.WithUnknownTypes(false),        // Reject unknown types
    bogo.WithSelectiveFields([]string{"target_field"}), // Optimize for specific fields
)

result, err := decoder.Decode(data)
```

### Field-Specific Optimization

Bogo includes a powerful field-specific decoding optimization that provides **up to 334x performance improvement** when you only need specific fields from large objects.

#### Performance Results (Apple M2, 8 cores)

```
BEFORE OPTIMIZATION:
BenchmarkFieldDecoding_Full-8               	   14395	    281858 ns/op	   48056 B/op	    1138 allocs/op
BenchmarkFieldDecoding_Selective-8          	   19111	     56626 ns/op	   31456 B/op	     826 allocs/op

AFTER OPTIMIZATION:
BenchmarkFieldDecoding_WithOptimization-8   	 1376488	       844.3 ns/op	     424 B/op	       7 allocs/op
```

#### Performance Improvements
- **334x faster** than full decoding (281,858ns → 844ns)
- **67x faster** than selective decoding without optimization (56,626ns → 844ns)
- **113x fewer allocations** than full decoding (1,138 → 7)
- **118x fewer allocations** than selective decoding (826 → 7)
- **113x less memory usage** than full decoding (48,056B → 424B)
- **74x less memory usage** than selective decoding (31,456B → 424B)

#### Usage Examples

**Method 1: Explicit Field Selection**
```go
// Only decode "id" and "name" fields from a complex object
optimizedDecoder := bogo.NewConfigurableDecoder(
    bogo.WithSelectiveFields([]string{"id", "name"}),
)

result, err := optimizedDecoder.Decode(largeObjectData)
// 334x faster than decoding the entire object!
```

**Method 2: Automatic Optimization with Struct Tags**
```go
// Define a struct with only the fields you need
type UserSummary struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
    // Large fields like "profile_data", "history", etc. are automatically skipped
}

var summary UserSummary
err := bogo.Unmarshal(complexUserData, &summary)
// Automatically optimized - only decodes the fields present in the struct!
```

#### How It Works

The optimization uses **field jumping** to skip over unwanted data:

1. **Field Detection**: The decoder identifies which fields are needed
2. **Smart Skipping**: Large, complex fields are skipped entirely using size information
3. **Direct Access**: Only target fields are parsed and decoded
4. **Memory Efficiency**: Unused data never allocates memory

#### Ideal Use Cases

- **Large API responses** where you only need specific fields
- **Database records** with large blob fields you want to skip  
- **Microservices** extracting metadata from complex payloads
- **Log processing** where you need specific fields from large log entries
- **Real-time systems** requiring ultra-low latency field extraction

### Encoder Options

```go
encoder := bogo.NewConfigurableEncoder(
    bogo.WithValidation(true),           // Enable validation
    bogo.WithCompression(false),         // Disable compression
)

data, err := encoder.Encode(value)
```

## Binary Format

Bogo uses a compact, self-describing binary format optimized for performance and space efficiency. The format is fully documented in the [Binary Format Specification](spec.md).

### Format Overview

The binary format consists of:

```
[Version:1][Type:1][TypeSpecificData:Variable]
```

- **Version**: Format version byte (currently `0x00`)
- **Type**: Data type identifier (see type constants below)
- **TypeSpecificData**: Variable-length data specific to each type

### Type Constants

| Type ID | Name | Description | Format |
|---------|------|-------------|--------|
| `0x00` | TypeNull | Null value | No additional data |
| `0x01` | TypeBoolTrue | Boolean true | No additional data |
| `0x02` | TypeBoolFalse | Boolean false | No additional data |
| `0x03` | TypeString | UTF-8 string | `[SizeLen:1][Size:VarInt][Data:Bytes]` |
| `0x04` | TypeByte | Single byte | `[Value:1]` |
| `0x05` | TypeInt | Signed integer | `[SizeLen:1][Value:VarInt]` |
| `0x06` | TypeUint | Unsigned integer | `[SizeLen:1][Value:VarInt]` |
| `0x07` | TypeFloat | IEEE 754 float | `[SizeLen:1][Value:Bytes]` |
| `0x08` | TypeBlob | Binary data | `[SizeLen:1][Size:VarInt][Data:Bytes]` |
| `0x09` | TypeTimestamp | Unix timestamp (ms) | `[Timestamp:8]` (little-endian) |
| `0x0A` | TypeArray | Heterogeneous array | `[SizeLen:1][TotalSize:VarInt][Elements:Variable]` |
| `0x0B` | TypeTypedArray | Homogeneous array | `[ElementType:1][Count:VarInt][Elements:Variable]` |
| `0x0C` | TypeObject | Key-value object | `[SizeLen:1][TotalSize:VarInt][FieldEntries:Variable]` |

### VarInt Encoding

Bogo uses Go's standard variable-length integer encoding (`binary.PutUvarint`/`binary.Uvarint`):
- Values 0-127: 1 byte
- Values 128-16383: 2 bytes  
- Values 16384-2097151: 3 bytes
- Up to 8 bytes for large values

### Object Format

Objects use a sophisticated field entry format for efficient encoding:

```
TypeObject + [SizeLen:1][TotalSize:VarInt] + FieldEntries
```

Each field entry:
```
[EntrySizeLen:1][EntrySize:VarInt][KeyLen:1][Key:Bytes][Value:EncodedValue]
```

This format allows for:
- Fast field skipping during parsing
- Efficient object traversal
- Self-describing nested structures

### Specification

For complete technical details, including encoding algorithms, edge cases, and implementation notes, see the [Binary Format Specification](spec.md).

## API Reference

### Core Functions

```go
// Marshal encodes a value to bogo binary format
func Marshal(v interface{}) ([]byte, error)

// Unmarshal decodes bogo binary data into a value
func Unmarshal(data []byte, v interface{}) error
```

### Streaming API

```go
// NewEncoder creates a streaming encoder
func NewEncoder(w io.Writer) *StreamEncoder

// NewDecoder creates a streaming decoder  
func NewDecoder(r io.Reader) *StreamDecoder
```

### Configuration

```go
// NewConfigurableEncoder creates an encoder with options
func NewConfigurableEncoder(options ...EncoderOption) *Encoder

// NewConfigurableDecoder creates a decoder with options
func NewConfigurableDecoder(options ...DecoderOption) *Decoder
```

## Error Handling

Bogo provides detailed error messages for debugging:

```go
data, err := bogo.Marshal(value)
if err != nil {
    // Handle encoding errors
    fmt.Printf("Encoding failed: %v\n", err)
}

var result interface{}
err = bogo.Unmarshal(data, &result)
if err != nil {
    // Handle decoding errors
    fmt.Printf("Decoding failed: %v\n", err)
}
```

## Testing

Run the test suite:

```bash
go test ./...
```

Run benchmarks:

```bash
go test -bench=. -benchmem
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -am 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Comparison with Other Formats

| Feature | Bogo | JSON | MessagePack | Protocol Buffers |
|---------|------|------|-------------|------------------|
| Human Readable | L |  | L | L |
| Schema Required | L | L | L |  |
| Go API Compatibility |  |  | L | L |
| Size Efficiency |  | L |  |  |
| Speed |  | L |  |  |
| Cross Language | L |  |  |  |

Bogo is ideal when you need JSON-like simplicity with binary format performance in Go applications.
