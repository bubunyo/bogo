# Bogo Binary Serialization Format Specification

**Version:** 0.0  
**Date:** December 2025
**Authors:** Bogo Development Team  

## Table of Contents

1. [Overview](#overview)
2. [Design Principles](#design-principles)
3. [Format Structure](#format-structure)
4. [Type System](#type-system)
5. [Encoding Specifications](#encoding-specifications)
6. [Examples](#examples)
7. [Implementation Notes](#implementation-notes)
8. [Version History](#version-history)

## Overview

Bogo is a compact, efficient binary serialization format designed for high-performance data interchange. It provides a type-safe, schema-less encoding system optimized for both space efficiency and parsing speed.

### Key Features

- **Compact Binary Format**: Minimal overhead with variable-length encoding
- **Type Safety**: Strong typing system with explicit type information
- **Cross-Platform**: Little-endian encoding with well-defined byte layouts
- **Extensible**: Versioned format supporting future enhancements
- **High Performance**: Optimized for fast encoding/decoding operations

## Design Principles

1. **Efficiency**: Minimize serialized data size through variable-length encoding
2. **Simplicity**: Clear, unambiguous format specification
3. **Performance**: Fast serialization/deserialization with minimal memory allocation
4. **Compatibility**: Consistent cross-platform behavior
5. **Extensibility**: Support for future type additions and format evolution

## Format Structure

### Basic Layout

```
┌─────────────┬─────────────┬─────────────────┐
│   Version   │    Type     │      Data       │
│   (1 byte)  │  (1 byte)   │  (variable)     │
└─────────────┴─────────────┴─────────────────┘
```

### Version Header

- **Size**: 1 byte
- **Current Version**: `0x00`
- **Purpose**: Format version identification and future compatibility

### Type Identifier

- **Size**: 1 byte  
- **Purpose**: Identifies the data type being encoded
- **Values**: See [Type System](#type-system)

## Type System

### Type Constants

| Type ID | Name | Description | Format |
|---------|------|-------------|--------|
| `0x00` | `TypeNull` | Null/nil value | No additional data |
| `0x01` | `TypeBoolTrue` | Boolean true | No additional data |
| `0x02` | `TypeBoolFalse` | Boolean false | No additional data |
| `0x03` | `TypeString` | UTF-8 string | `[SizeLen:1][Size:VarInt][Data:Bytes]` |
| `0x04` | `TypeByte` | Single byte (uint8) | `[Value:1]` |
| `0x05` | `TypeInt` | Signed integer | `[SizeLen:1][Value:VarInt]` |
| `0x06` | `TypeUint` | Unsigned integer | `[SizeLen:1][Value:VarInt]` |
| `0x07` | `TypeFloat` | IEEE 754 floating point | `[SizeLen:1][Value:Bytes]` |
| `0x08` | `TypeBlob` | Binary data ([]byte) | `[SizeLen:1][Size:VarInt][Data:Bytes]` |
| `0x09` | `TypeTimestamp` | 64-bit timestamp (ms) | `[Timestamp:8]` (little-endian) |
| `0x0A` | `TypeUntypedList` | Heterogeneous list | `[SizeLen:1][TotalSize:VarInt][Elements:Variable]` |
| `0x0B` | `TypeTypedList` | Homogeneous list | `[ElementType:1][Count:VarInt][Elements:Variable]` |
| `0x0C` | `TypeObject` | Key-value map/object | `[SizeLen:1][TotalSize:VarInt][FieldEntries:Variable]` |

## Encoding Specifications

### Variable-Length Integer Encoding (VarInt)

Bogo uses Go's standard variable-length integer encoding (`binary.PutUvarint`/`binary.Uvarint`) for efficient size representation:

**Encoding Rules:**
- Values 0-127: 1 byte
- Values 128-16383: 2 bytes  
- Values 16384-2097151: 3 bytes
- Up to 8 bytes for large values

**Structure:**
```
┌─────────────┬─────────────────────────────────┐
│  Size Len   │         Size Value              │
│  (1 byte)   │      (Size Len bytes)           │
└─────────────┴─────────────────────────────────┘
```

- **Size Len**: Number of bytes needed to store the size (1-8)
- **Size Value**: Little-endian encoded size using Go's VarInt encoding
- **Range**: 0 to 2^64-1

### Object Field Entry Format

Objects use a sophisticated field entry format for efficient encoding and field skipping:

```
TypeObject + [SizeLen:1][TotalSize:VarInt] + FieldEntries
```

Each field entry:
```
[EntrySizeLen:1][EntrySize:VarInt][KeyLen:1][Key:Bytes][Value:EncodedValue]
```

**Benefits:**
- Fast field skipping during parsing
- Efficient object traversal
- Self-describing nested structures
- Enables selective field decoding optimization

### Type-Specific Encodings

#### 1. Null (`TypeNull`)
**Purpose**: Represents null/nil values

**Structure:**
```
┌─────────────┬─────────────┐
│   Version   │ TypeNull    │
│    0x00     │    0x00     │
└─────────────┴─────────────┘
```

**Total Size**: 2 bytes

#### 2. Boolean (`TypeBoolTrue`, `TypeBoolFalse`)
**Purpose**: Represents boolean values

**Structure:**
```
┌─────────────┬─────────────────┐
│   Version   │   Bool Type     │
│    0x00     │ 0x01 or 0x02    │
└─────────────┴─────────────────┘
```

**Total Size**: 2 bytes

#### 3. Byte (`TypeByte`)
**Purpose**: Single byte values (uint8)

**Structure:**
```
┌─────────────┬─────────────┬─────────────┐
│   Version   │  TypeByte   │    Value    │
│    0x00     │    0x04     │  (1 byte)   │
└─────────────┴─────────────┴─────────────┘
```

**Total Size**: 3 bytes

#### 4. String (`TypeString`)
**Purpose**: UTF-8 encoded strings

**Structure:**
```
┌─────────────┬─────────────┬─────────────────┬─────────────────┐
│   Version   │ TypeString  │   Length Info   │   String Data   │
│    0x00     │    0x03     │   (VarInt)      │  (Length bytes) │
└─────────────┴─────────────┴─────────────────┴─────────────────┘
```

**Example**: String "hello"
```
00 03 01 05 68 65 6C 6C 6F
│  │  │  │  │
│  │  │  │  └── String content "hello"
│  │  │  └──── Length: 5
│  │  └────── Size length: 1 byte needed
│  └────── Type: String
└────── Version
```

#### 5. Integer (`TypeInt`, `TypeUint`)
**Purpose**: Variable-length integer encoding

**Structure:**
```
┌─────────────┬─────────────┬─────────────────┬─────────────────┐
│   Version   │  Int Type   │   Length Info   │   Integer Data  │
│    0x00     │ 0x05/0x06   │   (VarInt)      │  (Length bytes) │
└─────────────┴─────────────┴─────────────────┴─────────────────┘
```

**Encoding Rules:**
- Little-endian byte order
- Minimal byte representation
- Signed integers use two's complement

#### 6. Float (`TypeFloat`)
**Purpose**: IEEE 754 floating-point numbers

**Structure:**
```
┌─────────────┬─────────────┬─────────────────┬─────────────────┐
│   Version   │  TypeFloat  │   Length Info   │   Float Data    │
│    0x00     │    0x07     │   (VarInt)      │  (Length bytes) │
└─────────────┴─────────────┴─────────────────┴─────────────────┘
```

**Encoding**: 
- Decomposed IEEE 754 representation
- Variable-length based on precision requirements

#### 7. Blob (`TypeBlob`)
**Purpose**: Binary data (byte lists)

**Structure:**
```
┌─────────────┬─────────────┬─────────────────┬─────────────────┐
│   Version   │  TypeBlob   │   Length Info   │   Binary Data   │
│    0x00     │    0x08     │   (VarInt)      │  (Length bytes) │
└─────────────┴─────────────┴─────────────────┴─────────────────┘
```

#### 8. Timestamp (`TypeTimestamp`)
**Purpose**: 64-bit millisecond timestamps

**Structure:**
```
┌─────────────┬─────────────┬─────────────────────────────┐
│   Version   │TypeTimestamp│       Timestamp Value       │
│    0x00     │    0x09     │        (8 bytes LE)         │
└─────────────┴─────────────┴─────────────────────────────┘
```

**Total Size**: 10 bytes  
**Encoding**: Little-endian 64-bit signed integer (milliseconds since Unix epoch)

#### 9. List (`TypeUntypedList`)
**Purpose**: Heterogeneous lists with mixed types

**Structure:**
```
┌─────────────┬─────────────┬─────────────────┬─────────────────┐
│   Version   │TypeUntypedList│   Length Info   │   List Data     │
│    0x00     │    0x0A     │   (VarInt)      │   (Elements)    │
└─────────────┴─────────────┴─────────────────┴─────────────────┘
```

**List Data**: Sequence of encoded elements (each with full type headers)

#### 10. Typed List (`TypeTypedList`)
**Purpose**: Homogeneous lists optimized for same-type elements

**Structure:**
```
┌─────────────┬─────────────┬─────────────────┬─────────────┬─────────────────┬─────────────────┐
│   Version   │TypeTypedArr │   Length Info   │ Element Type│   Count Info    │   Elements      │
│    0x00     │    0x0B     │   (VarInt)      │  (1 byte)   │   (VarInt)      │   (Optimized)   │
└─────────────┴─────────────┴─────────────────┴─────────────┴─────────────────┴─────────────────┘
```

**Optimization**: Elements encoded without individual type headers

#### 11. Object (`TypeObject`)
**Purpose**: Key-value maps and structured objects

**Structure:**
```
┌─────────────┬─────────────┬─────────────────┬─────────────────┐
│   Version   │ TypeObject  │   Length Info   │   Field Data    │
│    0x00     │    0x0C     │   (VarInt)      │   (Fields)      │
└─────────────┴─────────────┴─────────────────┴─────────────────┘
```

**Field Entry Format:**
```
┌─────────────────┬─────────────┬─────────────┬─────────────────┐
│   Entry Size    │ Key Length  │     Key     │     Value       │
│    (VarInt)     │  (1 byte)   │ (Key bytes) │   (Encoded)     │
└─────────────────┴─────────────┴─────────────┴─────────────────┘
```

**Constraints:**
- Key length limited to 255 bytes
- Keys are UTF-8 strings
- Values can be any supported type

## Examples

### Example 1: Simple Object
**Go Code:**
```go
data := map[string]any{
    "name": "Alice",
    "age": 25,
    "active": true
}
```

**Binary Encoding (hexadecimal):**
```
00 0C 01 1F 01 0C 04 6E 61 6D 65 03 01 05 41 6C 69 63 65
01 07 03 61 67 65 05 01 19 01 09 06 61 63 74 69 76 65 01
```

### Example 2: Typed List
**Go Code:**
```go
numbers := []int{1, 2, 3, 4, 5}
```

**Binary Encoding:**
```
00 0B 01 0C 05 01 05 01 01 01 02 01 03 01 04 01 05
```

## Implementation Notes

### Encoder and Decoder Patterns

Bogo provides both simple functions and configurable encoder/decoder patterns:

**Simple API:**
```go
// Basic encoding/decoding
data, err := bogo.Encode(value)
value, err := bogo.Decode(data)

// Marshal/Unmarshal aliases
data, err := bogo.Marshal(value)
value, err := bogo.Unmarshal(data)
```

**Pattern-Based API:**
```go
// Configurable encoder
encoder := bogo.NewEncoder(
    bogo.WithMaxDepth(100),
    bogo.WithStrictMode(true),
    bogo.WithCompactLists(true),
)
data, err := encoder.Encode(value)

// Configurable decoder
decoder := bogo.NewDecoder(
    bogo.WithDecoderStrictMode(true),
    bogo.WithUnknownTypes(false),
    bogo.WithMaxObjectSize(1024*1024),
)
value, err := decoder.Decode(data)

// Statistics collection
statsEncoder := bogo.NewStatsCollector()
data, err := statsEncoder.Encode(value)
stats := statsEncoder.GetStats()
```

### Performance Considerations

1. **Streaming**: Format supports streaming parsers via `EncodeTo`/`DecodeFrom`
2. **Memory**: Minimal allocation during encoding/decoding
3. **CPU**: Simple byte operations, no complex transformations
4. **Statistics**: Optional performance monitoring with stats collectors

### Error Handling

1. **Version Mismatch**: Parsers must validate version compatibility
2. **Corruption Detection**: Length fields provide basic integrity checking
3. **Type Safety**: Strict type validation required

### Extensions

1. **New Types**: Can be added with new type IDs (0x0D+)
2. **Version Evolution**: Major format changes require version increment
3. **Backward Compatibility**: Older versions should remain parseable

## Zero Values vs Null Values

Bogo distinguishes between zero values and null values for all data types:

### Zero Values
Zero values are the default values for each data type when no explicit value is provided:
- **String**: `""` (empty string, zero length)
- **Integer**: `0` (numeric zero)
- **Float**: `0.0` (floating-point zero)
- **Boolean**: `false` (boolean false)
- **Byte**: `0x00` (zero byte)
- **Binary**: Empty binary data (zero length)
- **List**: Empty list (zero elements)
- **Object**: Empty object (zero fields)
- **Timestamp**: Epoch zero (1970-01-01 00:00:00 UTC)

**Encoding**: Zero values are encoded using their respective type identifiers with appropriate data representation.
**Decoding**: Zero values decode back to their language-specific zero/default values, preserving type information.

### Null Values
Null values represent the explicit absence of a value:
- **Explicit null**: Language-specific null/nil/None representation
- **Null pointers**: Uninitialized or explicitly null references
- **Optional types**: Unset optional values

**Encoding**: All null values are encoded as `TypeNull` (0x00).
**Decoding**: All null values decode back to the language's null representation.

### Encoding Examples

| Value Type | Example Value | Encoded As | Decodes To |
|------------|---------------|------------|------------|
| Zero string | `""` | TypeString + 0 length | Empty string |
| Zero integer | `0` | TypeInt + zero | Numeric zero |
| Zero boolean | `false` | TypeBoolFalse | Boolean false |
| Empty list | `[]` | TypeUntypedList + 0 elements | Empty list |
| Null value | `null` | TypeNull | Language null |

### Object Field Handling

Objects support both zero values and null values as field values:

```
{
    "zero_string": "",           // TypeString with zero length
    "zero_number": 0,            // TypeInt with value zero
    "null_value": null           // TypeNull
}
```

### Tri-State Logic

This distinction enables tri-state logic for data types:
- **True/Present**: Actual value (e.g., `true`, `"hello"`, `42`)
- **False/Empty**: Zero value (e.g., `false`, `""`, `0`)  
- **Unknown/Unset**: Null value (`null`)

This is particularly useful for:
- Optional fields that can be unset (null), empty (zero), or have a value
- Boolean fields that can be true, false, or unknown
- Numeric fields that can have a value, be zero, or be unspecified

## Security Considerations

1. **Buffer Bounds**: All length fields must be validated
2. **Resource Limits**: Implement reasonable size limits
3. **Recursion**: Limit nesting depth for objects/lists
4. **Memory**: Prevent excessive memory allocation

## Version History

### Version 0.0 (December 2025)
- Initial specification
- Core type system implementation
- Variable-length encoding
- Object and list support

---

**Note**: This specification is subject to change. Implementations should verify version compatibility before processing bogo-encoded data.
