# DataPack Specification

[https://go.dev/play/p/zhG7mz9P4Oj](https://go.dev/play/p/zhG7mz9P4Oj)

## Table of contents

* DataPack Specification
    * [Type system](#type-system)
        * [Limitation](#limitation)
        * [Extension types](#extension-types)
    * [Formats](#formats)
        * [Overview](#overview)
        * [Notation in diagrams](#notation-in-diagrams)
        * [nil format](#nil-format)
        * [bool format family](#bool-format-family)
        * [int format family](#int-format-family)
        * [float format family](#float-format-family)
        * [str format family](#str-format-family)
        * [bin format family](#bin-format-family)
        * [array format family](#array-format-family)
        * [map format family](#map-format-family)
        * [ext format family](#ext-format-family)
        * [Timestamp extension type](#timestamp-extension-type)
    * [Serialization: type to format conversion](#serialization-type-to-format-conversion)
    * [Deserialization: format to type conversion](#deserialization-format-to-type-conversion)
    * [Future discussion](#future-discussion)
        * [Profile](#profile)
    * [Implementation guidelines](#implementation-guidelines)
        * [Upgrading MessagePack specification](#upgrading-messagepack-specification)

## Type system

* Types
    * **Integer** represents an integer
    * **Nil** represents nil
    * **Boolean** represents true or false
    * **Float** represents a IEEE 754 double precision floating point number including NaN and Infinity
    * **Raw**
        * **String** extending Raw type represents a UTF-8 string
        * **Binary** extending Raw type represents a byte array
    * **Array** represents a sequence of objects
    * **Map** represents key-value pairs of objects
    * **Timestamp** represents an instantaneous point on the time-line in the world that is independent from time zones
      or calendars. Maximum precision is nanoseconds.
        * **Extension** represents a tuple of type information and a byte array where type information is an integer
          whose meaning is defined by 3rd party users or applications

### Limitation

* A value of an Integer object is limited from `-(2^63)` upto `(2^64)-1`
* Maximum length of a Binary object is `(2^32)-1`
* Maximum byte size of a String object is `(2^32)-1`
* String objects may contain invalid byte sequence and the behavior of a deserializer depends on the actual
  implementation when it received invalid byte sequence
    * Deserializers should provide functionality to get the original byte array so that applications can decide how to
      handle the object
* Maximum number of elements of an Array object is `(2^32)-1`
* Maximum number of key-value associations of a Map object is `(2^32)-1`

### Overview of Encoding Specification

| format name     | first byte (in binary) | first byte (in hex) |
|-----------------|------------------------|---------------------|
| positive fixint | 0xxxxxxx               | 0x00 - 0x7f         |
| fixmap          | 1000xxxx               | 0x80 - 0x8f         |
| fixarray        | 1001xxxx               | 0x90 - 0x9f         |
| fixstr          | 101xxxxx               | 0xa0 - 0xbf         |
| nil             | 11000000               | 0xc0                |
| (never used)    | 11000001               | 0xc1                |
| false           | 11000010               | 0xc2                |
| true            | 11000011               | 0xc3                |
| bin 8           | 11000100               | 0xc4                |
| bin 16          | 11000101               | 0xc5                |
| bin 32          | 11000110               | 0xc6                |
| ext 8           | 11000111               | 0xc7                |
| ext 16          | 11001000               | 0xc8                |
| ext 32          | 11001001               | 0xc9                |
| float 32        | 11001010               | 0xca                |
| float 64        | 11001011               | 0xcb                |
| uint 8          | 11001100               | 0xcc                |
| uint 16         | 11001101               | 0xcd                |
| uint 32         | 11001110               | 0xce                |
| uint 64         | 11001111               | 0xcf                |
| int 8           | 11010000               | 0xd0                |
| int 16          | 11010001               | 0xd1                |
| int 32          | 11010010               | 0xd2                |
| int 64          | 11010011               | 0xd3                |
| fixext 1        | 11010100               | 0xd4                |
| fixext 2        | 11010101               | 0xd5                |
| fixext 4        | 11010110               | 0xd6                |
| fixext 8        | 11010111               | 0xd7                |
| fixext 16       | 11011000               | 0xd8                |
| str 8           | 11011001               | 0xd9                |
| str 16          | 11011010               | 0xda                |
| str 32          | 11011011               | 0xdb                |
| array 16        | 11011100               | 0xdc                |
| array 32        | 11011101               | 0xdd                |
| map 16          | 11011110               | 0xde                |
| map 32          | 11011111               | 0xdf                |
| negative fixint | 111xxxxx               | 0xe0 - 0xff         |

### Notation in diagrams

    one byte:
    +--------+
    |        |
    +--------+

    a variable number of bytes:
    +========+
    |        |
    +========+

    variable number of objects stored in MessagePack format:
    +~~~~~~~~~~~~~~~~~+
    |                 |
    +~~~~~~~~~~~~~~~~~+
