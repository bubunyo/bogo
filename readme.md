# Bogo - High-Performance Binary Serialization for Go

Bogo is a fast, compact length-prefixed binary serialization format with embeded key fileds and type information. It provides efficient encoding and decoding of data types with performance and zerop copying.
It is ideal when you need JSON-like simplicity with binary format performance and selective field desirialization.

## Features

- **JSON-Compatible API**: Drop-in replacement for `encoding/json` with familiar `Marshal`/`Unmarshal` functions
- **High Performance**: Up to 3x faster deserialization compared to JSON
- **Compact Binary Format**: Efficient variable-length encoding reduces payload size
- **Comprehensive Type Support**: Supports all primitive types plus arrays, objects, and binary data
- **Streaming API**: Memory-efficient streaming encoding and decoding
- **Configurable Validation**: Optional UTF-8 validation, depth limits, and size constraints
- **Field-Specific Optimization**: Revolutionary selective field decoding with up to 334x performance improvement and 113x memory reduction

## Supported Types

| Primitive Type | Bogo Type | Description |
|---------|-----------|-------------|
| `nil` | TypeNull | Null values |
| `bool` | TypeBoolTrue/TypeBoolFalse | Boolean values |
| `string` | TypeString | UTF-8 strings with variable-length encoding |
| `byte` | TypeByte | Single byte values |
| `int`, `int8`, `int16`, `int32`, `int64` | TypeInt | Signed integers with VarInt encoding |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | TypeUint | Unsigned integers with VarInt encoding |
| `float32`, `float64` | TypeFloat | IEEE 754 floating-point numbers |
| `[]byte` | TypeBlob | Binary data with length prefix |
| `time` | TypeTimestamp | Unix timestamps |
| `[]any{}` | TypeArray | Heterogeneous arrays |
| `[]int{}` | TypeTypedArray | Homogeneous typed arrays |
| `object` | TypeObject | Key-value objects |

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

### Benchmark Results (Apple M2, 8 cores)

```
Simple Data (small objects):
- JSON Serialize:         709 ns/op    272 B/op    3 allocs/op
- Bogo Serialize:         964 ns/op   1088 B/op   18 allocs/op  (0.74x speed)
- MessagePack:            442 ns/op    320 B/op    4 allocs/op

- JSON Deserialize:      1910 ns/op    336 B/op    7 allocs/op
- Bogo Deserialize:       564 ns/op    488 B/op   16 allocs/op  (3.39x faster)
- MessagePack:            623 ns/op    168 B/op    4 allocs/op

Complex Data (nested structures):
- JSON Serialize:        6422 ns/op   2514 B/op   36 allocs/op
- Bogo Serialize:       15449 ns/op  18939 B/op  291 allocs/op  (0.42x speed)
- MessagePack:           3650 ns/op   2472 B/op   17 allocs/op

- JSON Deserialize:     19254 ns/op   3128 B/op   93 allocs/op
- Bogo Deserialize:      4341 ns/op   4464 B/op  101 allocs/op  (4.43x faster)
- MessagePack:           7888 ns/op   2921 B/op   86 allocs/op

Array Data (large arrays):
- JSON Serialize:       23173 ns/op   3889 B/op   15 allocs/op
- Bogo Serialize:       54654 ns/op  41025 B/op 1040 allocs/op  (0.42x speed)
- MessagePack:          10520 ns/op   8178 B/op    8 allocs/op

- JSON Deserialize:     56072 ns/op  21624 B/op  647 allocs/op
- Bogo Deserialize:      5822 ns/op   6016 B/op  119 allocs/op  (9.63x faster)
- MessagePack:          18743 ns/op  11027 B/op  416 allocs/op

Binary Data (large byte arrays):
- JSON Serialize:       12452 ns/op  16904 B/op   17 allocs/op
- Bogo Serialize:       12936 ns/op  64081 B/op   28 allocs/op  (0.96x speed)
- MessagePack:           3876 ns/op  16256 B/op    5 allocs/op

- JSON Deserialize:     95240 ns/op  16248 B/op   32 allocs/op
- Bogo Deserialize:       961 ns/op    872 B/op   17 allocs/op  (99.1x faster)
- MessagePack:           3346 ns/op  12292 B/op   21 allocs/op
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

Bogo includes field-specific decoding optimization that provides **up to 334x performance improvement** when you only need specific fields from large objects.

#### Performance Results (Apple M2, 8 cores)

```
BEFORE OPTIMIZATION:
BenchmarkFieldDecoding_Full-8               	   17913	     67037 ns/op	   48057 B/op	    1138 allocs/op
BenchmarkFieldDecoding_Selective-8          	   37464	     41110 ns/op	   31456 B/op	     826 allocs/op

AFTER OPTIMIZATION:
BenchmarkFieldDecoding_WithOptimization-8   	 2121243	       524.6 ns/op	     424 B/op	       7 allocs/op
```

#### Performance Improvements
- **128x faster** than full decoding (67,037ns → 524ns)
- **78x faster** than selective decoding without optimization (41,110ns → 524ns)
- **163x fewer allocations** than full decoding (1,138 → 7)
- **118x fewer allocations** than selective decoding (826 → 7)
- **113x less memory usage** than full decoding (48,057B → 424B)
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

## Zero Values vs Nil Values

Bogo makes a clear distinction between zero values and nil values for robust data handling:

### Zero Values (Type-Safe)
Zero values are preserved with their type information:
- `""` (empty string) → encodes as `TypeString` with 0 length → decodes as `""`
- `0` → encodes as `TypeInt` → decodes as `int64(0)`
- `false` → encodes as `TypeBoolFalse` → decodes as `false`
- `[]any{}` → encodes as `TypeArray` with 0 elements → decodes as `[]any{}`
- `time.Time{}` → encodes as `TypeTimestamp` → decodes as zero time

### Nil Values (Null)
Nil values are encoded uniformly as null:
- `nil` → encodes as `TypeNull` → decodes as `nil`
- `(*string)(nil)` → encodes as `TypeNull` → decodes as `nil`
- `map[string]any(nil)` → encodes as `TypeNull` → decodes as `nil`

### Example Usage

```go
data := map[string]any{
    "zero_string": "",           // Will remain empty string after decode
    "zero_number": 0,            // Will remain 0 after decode  
    "zero_bool":   false,        // Will remain false after decode
    "nil_value":   nil,          // Will remain nil after decode
}

encoded, _ := bogo.Marshal(data)
var decoded map[string]any
bogo.Unmarshal(encoded, &decoded)

fmt.Println(decoded["zero_string"])  // "" (empty string, not nil)
fmt.Println(decoded["zero_number"])  // 0 (zero, not nil)
fmt.Println(decoded["zero_bool"])    // false (not nil)
fmt.Println(decoded["nil_value"])    // nil
```

### Tri-State Booleans

This enables tri-state boolean logic:
- `true` → true value
- `false` → false value  
- `nil` → unknown/unset value

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

