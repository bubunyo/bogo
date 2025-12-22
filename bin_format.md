# Bogo Binary Format Specification

This document describes the binary layout for each data type in the Bogo serialization format.

## Overall Format

```
┌─────────────┬─────────────────────────────────────┐
│   Version   │          Encoded Data               │
│   (1 byte)  │         (variable)                  │
└─────────────┴─────────────────────────────────────┘
```

- **Version**: Always `0x00` (version 0) - appears only once at the start
- **Encoded Data**: The actual encoded value using type-specific format below

## Type-Specific Layouts

### 1. Null (Type = 0)
```
┌─────────────┐
│     0x00    │
│   (null)    │
└─────────────┘
```
Total: 1 byte

### 2. Boolean True (Type = 1)
```
┌─────────────┐
│     0x01    │
│   (true)    │
└─────────────┘
```
Total: 1 byte

### 3. Boolean False (Type = 2)
```
┌─────────────┐
│     0x02    │
│   (false)   │
└─────────────┘
```
Total: 1 byte

### 4. String (Type = 3)
```
┌─────────────┬─────────────┬─────────────┬─────────────────┐
│     0x03    │   Len Size  │   Length    │   String Data   │
│  (string)   │  (1 byte)   │  (varint)   │   (len bytes)   │
└─────────────┴─────────────┴─────────────┴─────────────────┘
```
- **Len Size (X)**: Number of bytes used for the varint length encoding
- **Length (Y)**: String length as varint, occupying exactly X bytes
- **String Data**: UTF-8 string bytes, exactly Y bytes

**Parsing Steps:**
1. Read 1 byte → X (number of bytes for length)
2. Read X bytes → decode as varint to get Y (actual string length)
3. Read Y bytes → decode as UTF-8 string

### 5. Signed Integer (Type = 5)
```
┌─────────────┬─────────────┬─────────────────┐
│     0x05    │   Len Size  │   Varint Data   │
│    (int)    │  (1 byte)   │  (len bytes)    │
└─────────────┴─────────────┴─────────────────┘
```
- **Len Size (X)**: Number of bytes used for the varint encoding
- **Varint Data**: Signed varint (zigzag encoded), occupying exactly X bytes

**Parsing Steps:**
1. Read 1 byte → X (number of bytes for varint)
2. Read X bytes → decode as signed varint

### 6. Unsigned Integer (Type = 6)
```
┌─────────────┬─────────────┬─────────────────┐
│     0x06    │   Len Size  │   Uvarint Data  │
│   (uint)    │  (1 byte)   │  (len bytes)    │
└─────────────┴─────────────┴─────────────────┘
```
- **Len Size (X)**: Number of bytes used for the uvarint encoding
- **Uvarint Data**: Unsigned varint, occupying exactly X bytes

**Parsing Steps:**
1. Read 1 byte → X (number of bytes for uvarint)
2. Read X bytes → decode as unsigned varint

### 7. Float (Type = 7)
```
┌─────────────┬─────────────┬─────────────────┬─────────────────┐
│     0x07    │   Len Size  │  Sign+Exponent  │    Mantissa     │
│   (float)   │  (1 byte)   │    (2 bytes)    │   (varint)      │
└─────────────┴─────────────┴─────────────────┴─────────────────┘
```
- **Len Size (X)**: Total size of sign+exponent+mantissa data
- **Sign+Exponent**: 2 bytes little-endian (sign bit + 11-bit exponent)
- **Mantissa**: 52-bit mantissa as uvarint (only if non-zero)

**Parsing Steps:**
1. Read 1 byte → X (total size of float data)
2. Read X bytes → first 2 bytes are sign+exponent, remaining bytes are mantissa varint

### 8. Binary Blob (Type = 8)
```
┌─────────────┬─────────────┬─────────────┬─────────────────┐
│     0x08    │   Len Size  │  Data Size  │   Binary Data   │
│   (blob)    │  (1 byte)   │  (varint)   │   (size bytes)  │
└─────────────┴─────────────┴─────────────┴─────────────────┘
```
- **Len Size (X)**: Number of bytes used for the data size varint
- **Data Size (Y)**: Size of binary data as varint, occupying exactly X bytes
- **Binary Data**: Raw bytes, exactly Y bytes

**Parsing Steps:**
1. Read 1 byte → X (number of bytes for data size)
2. Read X bytes → decode as varint to get Y (binary data size)
3. Read Y bytes → raw binary data

**Use Cases:** Images, files, UUIDs, hashes, encrypted data, certificates

### Binary Blob Example (UUID)
```
┌─────────────┬─────────────┬─────────────┬─────────────┬───────────────────────────────────────┐
│     0x00    │     0x08    │     0x01    │     0x10    │            UUID Bytes                 │
│  (version)  │   (blob)    │   (X=1)     │   (Y=16)    │            (16 bytes)                 │
└─────────────┴─────────────┴─────────────┴─────────────┴───────────────────────────────────────┘
```
16 bytes of raw UUID data (e.g., `550e8400-e29b-41d4-a716-446655440000`)

### 9. Timestamp (Type = 9)
```
┌─────────────┬─────────────────────────────────────┐
│     0x09    │           Timestamp                 │
│(timestamp)  │          (8 bytes)                  │
└─────────────┴─────────────────────────────────────┘
```
- **Timestamp**: Unix timestamp in milliseconds as int64, little-endian, exactly 8 bytes

**Parsing Steps:**
1. Read 8 bytes → decode as int64 little-endian to get timestamp (milliseconds since Unix epoch)

**Notes:**
- Fixed 8-byte encoding (no length header needed)
- Millisecond precision
- Always UTC timezone
- Little-endian byte order
- Supports dates from 1970 to ~292 million years in the future
- Negative values for dates before 1970

### Timestamp Example (2024-01-15T10:30:45.123Z)
```
┌─────────────┬─────────────┬─────────────────────────────────────────┐
│     0x00    │     0x09    │              Timestamp                  │
│  (version)  │(timestamp)  │         (1705317045123)                 │
└─────────────┴─────────────┴─────────────────────────────────────────┘
```
Timestamp 1705317045123 as 8-byte int64 little-endian = Jan 15, 2024 10:30:45.123 UTC

### 10. Untyped Array (Type = 10)
```
┌─────────────┬─────────────┬─────────────┬─────────────────┐
│     0x0A    │   Len Size  │  Data Size  │   Element Data  │
│ (untyped)   │  (1 byte)   │  (varint)   │   (size bytes)  │
└─────────────┴─────────────┴─────────────┴─────────────────┘
```
- **Len Size (X)**: Number of bytes used for the data size varint
- **Data Size (Y)**: Total size of all encoded elements as varint, occupying exactly X bytes
- **Element Data**: Concatenated encoded elements, exactly Y bytes total

**Parsing Steps:**
1. Read 1 byte → X (number of bytes for data size)
2. Read X bytes → decode as varint to get Y (total size of all elements)
3. Read Y bytes → parse as concatenated encoded elements (each with full type headers)

**Note**: Each element includes its own type byte and length information for boundary detection.

### 11. Typed Array (Type = 11)

#### For Fixed-Size Element Types (bool, byte)
```
┌─────────────┬─────────────┬─────────────┬─────────────┬─────────────────┐
│     0x0B    │ Element Type│   Len Size  │ Array Count │   Element Data  │
│(typed array)│  (1 byte)   │  (1 byte)   │  (varint)   │   (size bytes)  │
└─────────────┴─────────────┴─────────────┴─────────────┴─────────────────┘
```

#### For Variable-Size Element Types (string, int, uint, float, blob, timestamp)
```
┌─────────────┬─────────────┬─────────────┬─────────────┬─────────────────┐
│     0x0B    │ Element Type│   Len Size  │  Data Size  │   Element Data  │
│(typed array)│  (1 byte)   │  (1 byte)   │  (varint)   │   (size bytes)  │
└─────────────┴─────────────┴─────────────┴─────────────┴─────────────────┘
```

- **Element Type**: The type of all elements
- **Len Size (X)**: Number of bytes used for the count/size varint
- **Array Count (N)**: Number of elements (for fixed-size types only)
- **Data Size (Y)**: Total size of all element data (for variable-size types)
- **Element Data**: Concatenated element data

**Parsing Steps:**

**For Fixed-Size Types (TypeBoolTrue, TypeBoolFalse, TypeByte, TypeTimestamp):**
1. Read 1 byte → element type
2. Read 1 byte → X (number of bytes for count)
3. Read X bytes → decode as varint to get N (element count)
4. For bools/bytes: Read N bytes → parse as N elements (1 byte each)
5. For timestamps: Read N × 8 bytes → parse as N timestamps (8 bytes each)

**For Variable-Size Types (TypeString, TypeInt, TypeUint, TypeFloat, TypeBlob):**
1. Read 1 byte → element type
2. Read 1 byte → X (number of bytes for data size)
3. Read X bytes → decode as varint to get Y (total data size)
4. Read Y bytes → parse elements sequentially (each with type-specific length headers)

**Predetermined Sizes:**
- TypeBoolTrue/TypeBoolFalse: 1 byte each (no length header needed)
- TypeByte: 1 byte each (no length header needed)  
- TypeTimestamp: 8 bytes each (no length header needed)
- TypeString/TypeInt/TypeUint/TypeFloat/TypeBlob: Variable (need length headers per element)

### 12. Object (Type = 12) - Not Yet Implemented
```
┌─────────────┬─────────────┬─────────────┬─────────────────┐
│     0x0C    │   Len Size  │  Data Size  │   Field Data    │
│  (object)   │  (1 byte)   │  (varint)   │   (size bytes)  │
└─────────────┴─────────────┴─────────────┴─────────────────┘
```
- **Len Size (X)**: Number of bytes used for the data size varint
- **Data Size (Y)**: Total size of all key-value pairs as varint, occupying exactly X bytes
- **Field Data**: Concatenated key-value pairs, exactly Y bytes total

#### Field Entry Format
Each field entry in the Field Data follows this format:
```
┌─────────────┬─────────────┬─────────────┬─────────────┬─────────────────┐
│   Len Size  │ Entry Size  │  Key Length │    Key      │      Value      │
│   (1 byte)  │  (varint)   │   (1 byte)  │ (key bytes) │   (encoded)     │
└─────────────┴─────────────┴─────────────┴─────────────┴─────────────────┘
```

- **Len Size (X)**: Number of bytes for the entry size varint
- **Entry Size (Z)**: Total size of this key-value entry (key length + key + value), occupying X bytes
- **Key Length (K)**: Length of key string (0-255 bytes)
- **Key**: UTF-8 key string, exactly K bytes
- **Value**: Encoded value using standard type encoding

**Parsing Steps:**
1. Read 1 byte → X (number of bytes for data size)
2. Read X bytes → decode as varint to get Y (total size of all fields)
3. For each field entry within Y bytes:
   a. Read 1 byte → X₂ (bytes for entry size)
   b. Read X₂ bytes → decode as varint to get Z (this entry size)
   c. Read 1 byte → K (key length)
   d. Read K bytes → key string
   e. Read remaining bytes of entry → decode value using standard encoding

**Key Constraints:**
- Keys are limited to 255 bytes (single-byte length encoding)
- Keys must be valid UTF-8 strings
- Duplicate keys are allowed (implementation-defined behavior)

## Complete Examples

### String "hello"
```
┌─────────────┬─────────────┬─────────────┬─────────────┬─────────────────────────┐
│     0x00    │     0x03    │     0x01    │     0x05    │   'h' 'e' 'l' 'l' 'o'   │
│  (version)  │  (string)   │   (X=1)     │   (Y=5)     │     (5 bytes)           │
└─────────────┴─────────────┴─────────────┴─────────────┴─────────────────────────┘
```
X=1 means "read next 1 byte for length", Y=5 means "string is 5 bytes long"

### Untyped Array [1, "hi"]
```
┌─────────────┬─────────────┬─────────────┬─────────────┬─────────────────────────────────────────────┐
│     0x00    │     0x0A    │     0x01    │     0x07    │              Element Data                   │
│  (version)  │ (untyped)   │   (X=1)     │   (Y=7)     │              (7 bytes)                      │
└─────────────┴─────────────┴─────────────┴─────────────┴─────────────────────────────────────────────┘
                                                        │
                                                        ├─ 0x05 0x01 0x01 (int: 1)
                                                        └─ 0x03 0x01 0x02 'h' 'i' (string: "hi")
```
X=1 means "read next 1 byte for data size", Y=7 means "element data is 7 bytes total"

### Typed Array [true, false, true] (bool array, fixed-size)
```
┌─────────────┬─────────────┬─────────────┬─────────────┬─────────────┬─────────────────┐
│     0x00    │     0x0B    │     0x01    │     0x01    │     0x03    │  Element Data   │
│  (version)  │  (typed)    │   (bool)    │   (X=1)     │   (N=3)     │   (3 bytes)     │
└─────────────┴─────────────┴─────────────┴─────────────┴─────────────┴─────────────────┘
                                                                      │
                                                                      ├─ 0x01 (true)
                                                                      ├─ 0x00 (false)
                                                                      └─ 0x01 (true)
```
- Element Type=0x01 (TypeBoolTrue, but represents bool type)
- X=1, N=3 (3 elements)
- Each bool is 1 byte: 0x01=true, 0x00=false
- Total: 3 × 1 = 3 bytes (no length headers needed)

### Typed Array [100, 200, 300] (int array, variable-size)
```
┌─────────────┬─────────────┬─────────────┬─────────────┬─────────────┬─────────────────┐
│     0x00    │     0x0B    │     0x05    │     0x01    │     0x09    │  Element Data   │
│  (version)  │  (typed)    │   (int)     │   (X=1)     │   (Y=9)     │   (9 bytes)     │
└─────────────┴─────────────┴─────────────┴─────────────┴─────────────┴─────────────────┘
                                                                      │
                                                                      ├─ 0x01 0x64 (100)
                                                                      ├─ 0x01 0xC8 (200)
                                                                      └─ 0x02 0x2C 0x01 (300)
```
- Element Type=0x05 (TypeInt)
- X=1, Y=9 (total data size is 9 bytes)
- Each int has its own length header (varint encoding)
- Parser reads Y bytes total, parsing elements sequentially

### Typed Array ["a", "bb", "ccc"] (string array, variable-size)
```
┌─────────────┬─────────────┬─────────────┬─────────────┬─────────────┬─────────────────┐
│     0x00    │     0x0B    │     0x03    │     0x01    │     0x0C    │  Element Data   │
│  (version)  │  (typed)    │  (string)   │   (X=1)     │   (Y=12)    │   (12 bytes)    │
└─────────────┴─────────────┴─────────────┴─────────────┴─────────────┴─────────────────┘
                                                                      │
                                                                      ├─ 0x01 0x01 'a' (len=1)
                                                                      ├─ 0x01 0x02 'bb' (len=2)
                                                                      └─ 0x01 0x03 'ccc' (len=3)
```
- Element Type=0x03 (TypeString)
- X=1, Y=12 (total data size is 12 bytes)
- Each string has its own length header since strings are variable-length

### Object {"name": "John", "age": 25}
```
┌─────────────┬─────────────┬─────────────┬─────────────┬───────────────────────────────────────────────┐
│     0x00    │     0x0C    │     0x01    │     0x17    │                Field Data                     │
│  (version)  │  (object)   │   (X=1)     │   (Y=23)    │                (23 bytes)                     │
└─────────────┴─────────────┴─────────────┴─────────────┴───────────────────────────────────────────────┘
                                                        │
                                                        ├─ Entry 1: "name":"John"
                                                        │  0x01 0x0C 0x04 'name' 0x03 0x01 0x04 'John'
                                                        │  (1)(12)(4)(name)(string)(1)(4)(John)
                                                        │
                                                        └─ Entry 2: "age":25  
                                                           0x01 0x07 0x03 'age' 0x05 0x01 0x19
                                                           (1)(7)(3)(age)(int)(1)(25)
```

**Breakdown:**
- Object has Y=23 bytes of field data total
- Entry 1: len_size=1, entry_size=12, key="name"(1+4=5 bytes), value="John"(string: 1+1+1+4=7 bytes)  
  - Total entry: 1 + 1 + (1+4) + (1+1+1+4) = 14 bytes
- Entry 2: len_size=1, entry_size=7, key="age"(1+3=4 bytes), value=25(int: 1+1+1=3 bytes)
  - Total entry: 1 + 1 + (1+3) + (1+1+1) = 9 bytes  
- Field data total: 14 + 9 = 23 bytes

## Notes

- All multi-byte integers use little-endian encoding for fixed-width fields
- Varints follow Go's binary.Varint/Uvarint encoding
- Strings are UTF-8 encoded
- Arrays and objects store total payload size for efficient parsing
- Float encoding separates IEEE 754 components for space efficiency
