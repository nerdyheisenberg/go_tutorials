# Go Deep Dive — Part 2: Variables, Types & The Go Type System

---

## Chapter 1: The Type System Philosophy

### Why Static Typing Matters

Go is **statically typed** — every variable has a type known at compile time. This:
- Catches entire classes of bugs before the program runs
- Enables editor tooling (autocomplete, refactoring)
- Makes code self-documenting
- Allows compiler optimizations

Contrast with:
- **C++**: Statically typed, but implicit conversions cause surprises
- **Python**: Dynamically typed, bugs appear at runtime
- **Go**: Statically typed with type inference (like C++ `auto`, but safer)

### The Fundamental Rule: No Implicit Conversion

This is the **single biggest source of C++ bugs that Go eliminates**:

```go
// C++ (compiles, but loses precision silently):
// int x = 3.99;  // x = 3 (truncated, no warning by default!)
// float y = 1/3; // y = 0.0 (integer division, silent!)

// Go: ALWAYS explicit
var x int = 3         // OK
// var x int = 3.99   // ERROR: cannot use 3.99 as int (truncated constant 3 to int)

var y float64 = 1.0/3.0 // OK: 0.333...
var z float64 = 1/3     // ERROR: cannot use 1/3 (untyped int constant 0) as float64

// Correct:
var z float64 = float64(1) / float64(3) // 0.333...
```

---

## Chapter 2: Variable Declaration — All Forms Explained

### Form 1: Full `var` Declaration

```go
var name string = "Rohit"
```

**When to use:** Package-level variables, or when you want to be explicit about the type.

**Why explicit type sometimes matters:**
```go
var timeout int64 = 5000  // Clearly int64, NOT int
// vs
timeout := 5000           // inferred as int — might cause issues with APIs expecting int64
```

### Form 2: Type Inference with `var`

```go
var count = 10      // inferred as int
var name = "Rohit"  // inferred as string
var ratio = 0.5     // inferred as float64
```

**What type does the compiler infer?**

```go
var i = 10          // int (platform-dependent: 32 or 64 bit)
var f = 3.14        // float64 (ALWAYS float64, never float32)
var b = true        // bool
var s = "hello"     // string
var r = 'A'         // int32 (rune) — NOT byte!

// Negative case: don't assume int32
var x = 10
_ = x + int32(5)  // ERROR: mismatched types int and int32
```

### Form 3: Short Declaration `:=`

```go
name := "Rohit"     // ONLY inside functions — most idiomatic
```

**Rules of `:=`:**
1. Only inside functions (not at package level)
2. At least one variable on left must be NEW
3. Creates a new scope variable (can shadow outer variables)

```go
package main

import "fmt"

// var name = "global"  // OK at package level
// name := "global"     // ERROR: undefined: name (cannot use := at package level)

func main() {
    x := 10
    fmt.Println(x)
    
    // At least one new variable:
    x, y := 20, 30  // x is REASSIGNED (not new), y is NEW — this is OK
    fmt.Println(x, y)
    
    // x, x := 10, 20  // ERROR: no new variables on left side of :=
    
    // Shadowing (be careful!):
    {
        x := 100    // NEW x in inner scope, SHADOWS outer x
        fmt.Println(x) // 100
    }
    fmt.Println(x) // 20 — original x unchanged
}
```

### Form 4: Multiple `var` Block

```go
var (
    host     string = "localhost"
    port     int    = 5432
    maxConns int    = 10
    debug    bool   // zero value: false
)
```

**Why use this form?**
- Groups related variables visually
- Common for package-level configuration variables
- The Go style for multiple related vars

### Form 5: Blank Identifier `_`

```go
// _ discards a value — compiler requires ALL declared variables to be used
result, err := doSomething()
_, err = doSomethingElse() // ignore result, check error only
x, _ := divide(10, 3)     // ignore error (ONLY do this if you know it can't fail)

// Common in range loops
for _, value := range slice {
    fmt.Println(value) // don't need the index
}
for index, _ := range slice {
    // same as: for index := range slice
}
```

---

## Chapter 3: Integer Types — Choosing the Right One

### The Integer Zoo

```go
// Platform-dependent (most common for general use)
int    // 32-bit on 32-bit systems, 64-bit on 64-bit systems
uint   // unsigned version

// Fixed-size (use when you need specific width)
int8   // -128 to 127
int16  // -32768 to 32767
int32  // -2147483648 to 2147483647
int64  // -9223372036854775808 to 9223372036854775807

uint8  // 0 to 255
uint16 // 0 to 65535
uint32 // 0 to 4294967295
uint64 // 0 to 18446744073709551615

// Aliases (different names, same type)
byte   // = uint8 — for raw binary data
rune   // = int32 — for Unicode code points (characters)

// Pointer type
uintptr // large enough to hold a pointer value (not for arithmetic!)
```

### When to Use Which?

```go
// Use int for:
// - General-purpose counters, indices, sizes
n := 10
for i := 0; i < n; i++ {}
len(slice)  // returns int

// Use int64 for:
// - Time (nanoseconds, Unix timestamps)
var ts int64 = time.Now().UnixNano()
// - Large IDs, database row IDs
var userID int64 = 1234567890

// Use uint8/byte for:
// - Raw bytes, binary data, pixel values
data := []byte{0xFF, 0x00, 0x1A}
pixel := uint8(255)

// Use int32/rune for:
// - Unicode code points
ch := rune('A')           // 65
emoji := rune('😀')       // 128512

// Use uint32 for:
// - IPv4 addresses, CRC-32 checksums
var crc uint32

// Use float64 (NEVER float32 unless specific reason):
// - All floating point math
pi := 3.14159265358979323846
// float32 only has ~7 significant decimal digits — causes precision errors
```

### Integer Overflow — A Critical Gotcha

```go
var x int8 = 127
x++              // What happens? Overflow! x becomes -128
fmt.Println(x)   // -128 (no panic, no error — silent wraparound!)

// C++ has undefined behavior on signed overflow
// Go has defined behavior: wraps around (like unsigned in C)
// But it's still a bug! Always check for overflow in critical code.

// Safe overflow check
import "math"

func safeAdd(a, b int) (int, error) {
    if b > 0 && a > math.MaxInt-b {
        return 0, fmt.Errorf("integer overflow: %d + %d", a, b)
    }
    if b < 0 && a < math.MinInt-b {
        return 0, fmt.Errorf("integer underflow: %d + %d", a, b)
    }
    return a + b, nil
}
```

---

## Chapter 4: Floating Point — Precision and Pitfalls

```go
package main

import (
    "fmt"
    "math"
)

func main() {
    // NEVER compare floats with ==
    a := 0.1 + 0.2
    b := 0.3
    fmt.Println(a == b)        // false! (floating point precision)
    fmt.Println(a)             // 0.30000000000000004
    
    // Correct comparison with epsilon
    epsilon := 1e-9
    fmt.Println(math.Abs(a-b) < epsilon) // true
    
    // Special float values
    inf := math.Inf(1)        // positive infinity
    negInf := math.Inf(-1)   // negative infinity
    nan := math.NaN()        // not a number
    
    fmt.Println(math.IsInf(inf, 1))   // true
    fmt.Println(math.IsNaN(nan))       // true
    fmt.Println(nan == nan)           // FALSE! NaN is never equal to itself
    
    // float32 vs float64 precision
    var f32 float32 = 1234567.89
    var f64 float64 = 1234567.89
    fmt.Printf("float32: %.10f\n", f32) // 1234567.8750000000 (lost precision!)
    fmt.Printf("float64: %.10f\n", f64) // 1234567.8900000000
}
```

### Testing Floats

```go
func TestFloatComparison(t *testing.T) {
    tests := []struct {
        name     string
        a, b     float64
        want     bool
        epsilon  float64
    }{
        {"exact equal", 1.0, 1.0, true, 1e-9},
        {"very close", 0.1+0.2, 0.3, true, 1e-9},
        {"clearly different", 1.0, 2.0, false, 1e-9},
        {"near epsilon", 1.0, 1.0 + 1e-10, true, 1e-9},
        {"just over epsilon", 1.0, 1.0 + 1e-8, false, 1e-9},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := math.Abs(tt.a-tt.b) < tt.epsilon
            if got != tt.want {
                t.Errorf("floatEqual(%v, %v) = %v, want %v",
                    tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

---

## Chapter 5: Strings — Deep Dive

### What IS a String in Go?

A Go string is an **immutable sequence of bytes** (not characters!). Internally:

```go
// A string is essentially:
type StringHeader struct {
    Data uintptr  // pointer to underlying byte array
    Len  int      // number of bytes
}
// The bytes are NOT necessarily UTF-8, but by convention they should be.
```

```go
s := "Hello, 世界"

// len() returns BYTES, not characters!
fmt.Println(len(s)) // 13 (Hello, = 7 bytes, 世 = 3 bytes, 界 = 3 bytes)

// Iterating by BYTE (gives you byte values, not characters!)
for i := 0; i < len(s); i++ {
    fmt.Printf("%d: %x\n", i, s[i])
}
// 7: e4 (first byte of '世' — NOT a complete character)

// Iterating by RUNE (correct for Unicode)
for i, ch := range s {
    fmt.Printf("byte index %d: %c (U+%04X)\n", i, ch, ch)
}
// byte index 0: H (U+0048)
// byte index 7: 世 (U+4E16)  ← index jumps because '世' is 3 bytes
// byte index 10: 界 (U+754C)
```

### String Immutability

```go
s := "hello"
// s[0] = 'H'    // COMPILE ERROR: cannot assign to s[0] (strings are immutable)

// To "modify" a string, create a new one:
bytes := []byte(s)  // convert to mutable byte slice
bytes[0] = 'H'
s = string(bytes)   // convert back
fmt.Println(s) // "Hello"

// Or use strings.Builder for efficient building:
var sb strings.Builder
sb.WriteString("Hello")
sb.WriteString(", ")
sb.WriteString("World")
result := sb.String() // "Hello, World"
```

### String Conversion — Costs and Gotchas

```go
// string ↔ []byte involves COPYING (O(n) cost)
s := "hello"
b := []byte(s)    // copies 5 bytes
s2 := string(b)   // copies 5 bytes again

// In hot paths, avoid these conversions IMPORTANT
// Use: strings.Builder, bytes.Buffer, io.Writer

// string ↔ []rune involves DECODING+COPYING
r := []rune(s)       // decodes UTF-8 → rune slice
s3 := string(r)      // encodes rune slice → UTF-8

// WRONG: int → string (Go 1.15+ gives warning) IMPORTANT
n := 65
s4 := string(n)      // Was: "A" (char 65). Now: compiler warning.
                     // Use string(rune(n)) explicitly
s5 := string(rune(n)) // "A" — correct, explicit
s6 := fmt.Sprintf("%c", n) // "A" — another option
```

### String Operations — When to Use Which

```go
package main

import (
    "fmt"
    "strings"
    "unicode"
)

func main() {
    s := "Hello, World! Hello, Go!"
    
    // strings.Contains vs strings.Index
    fmt.Println(strings.Contains(s, "World"))  // true (just need yes/no)
    fmt.Println(strings.Index(s, "World"))     // 7 (need position)
    fmt.Println(strings.Index(s, "Python"))    // -1 (not found)
    
    // Count occurrences
    fmt.Println(strings.Count(s, "Hello")) // 2
    
    // Split and Join
    parts := strings.Split("a,b,c,d", ",")    // ["a" "b" "c" "d"]
    fmt.Println(strings.Join(parts, " | "))    // "a | b | c | d"
    
    // SplitN — limit number of parts
    parts2 := strings.SplitN("key=value=more", "=", 2) // ["key" "value=more"]
    
    // Fields — split on any whitespace (collapsing multiple)
    words := strings.Fields("  hello   world  ")  // ["hello" "world"]
    
    // Trim functions
    fmt.Println(strings.TrimSpace("  hello  "))  // "hello"
    fmt.Println(strings.Trim("***hello***", "*")) // "hello"
    fmt.Println(strings.TrimLeft("***hello***", "*")) // "hello***"
    fmt.Println(strings.TrimRight("***hello***", "*")) // "***hello"
    fmt.Println(strings.TrimPrefix("/api/v1/users", "/api")) // "/v1/users"
    fmt.Println(strings.TrimSuffix("hello.go", ".go"))       // "hello"
    
    // Replace
    fmt.Println(strings.Replace(s, "Hello", "Hi", 1))  // replace first
    fmt.Println(strings.Replace(s, "Hello", "Hi", -1)) // replace all
    fmt.Println(strings.ReplaceAll(s, "Hello", "Hi"))  // replace all (cleaner)
    
    // Case
    fmt.Println(strings.ToUpper("hello")) // "HELLO"
    fmt.Println(strings.ToLower("HELLO")) // "hello"
    fmt.Println(strings.Title("hello world")) // "Hello World" (deprecated, use golang.org/x/text)
    
    // Check prefix/suffix
    fmt.Println(strings.HasPrefix(s, "Hello")) // true
    fmt.Println(strings.HasSuffix(s, "Go!"))   // true
    
    // Index operations
    fmt.Println(strings.LastIndex(s, "Hello")) // 14
    
    // IndexFunc — custom predicate
    fmt.Println(strings.IndexFunc("hello 世界", unicode.Is(unicode.Han, rune(0)))) // wrong usage, see below:
    idx := strings.IndexFunc("hello 世界", func(r rune) bool {
        return r > 127 // first non-ASCII character
    })
    fmt.Println(idx) // 6
    
    _ = parts2
    _ = words
}
```

### Efficient String Concatenation — Negative Case

```go
// BAD: O(n²) — each + creates a new string, copying all previous bytes
func badConcat(items []string) string {
    result := ""
    for _, item := range items {
        result += item + ", " // creates a new string every iteration!
    }
    return result
}

// GOOD: O(n) — strings.Builder pre-allocates
func goodConcat(items []string) string {
    var b strings.Builder
    b.Grow(len(items) * 10) // pre-allocate estimate
    for i, item := range items {
        if i > 0 {
            b.WriteString(", ")
        }
        b.WriteString(item)
    }
    return b.String()
}

// ALSO GOOD: strings.Join (when separator is consistent)
func bestConcat(items []string) string {
    return strings.Join(items, ", ")
}
```

### Testing Strings

```go
package main

import (
    "strings"
    "testing"
    "unicode/utf8"
)

func TestStringBytes(t *testing.T) {
    s := "Hello, 世界"
    
    byteLen := len(s)
    runeLen := utf8.RuneCountInString(s)
    
    if byteLen != 13 {
        t.Errorf("byte length: got %d, want 13", byteLen)
    }
    if runeLen != 9 {
        t.Errorf("rune length: got %d, want 9", runeLen)
    }
}

func TestStringOperations(t *testing.T) {
    tests := []struct {
        name   string
        input  string
        op     func(string) string
        want   string
    }{
        {"trim space", "  hello  ", strings.TrimSpace, "hello"},
        {"to upper", "hello", strings.ToUpper, "HELLO"},
        {"to lower", "HELLO", strings.ToLower, "hello"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.op(tt.input)
            if got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}

func BenchmarkStringConcat(b *testing.B) {
    items := make([]string, 100)
    for i := range items {
        items[i] = "item"
    }
    
    b.Run("plus operator", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := ""
            for _, item := range items {
                result += item
            }
            _ = result
        }
    })
    
    b.Run("strings.Builder", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            var sb strings.Builder
            for _, item := range items {
                sb.WriteString(item)
            }
            _ = sb.String()
        }
    })
    
    b.Run("strings.Join", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = strings.Join(items, "")
        }
    })
}
```

---

## Chapter 6: Type System Deep Dive — Named Types

### Defining New Types

```go
// type creates a BRAND NEW type (not just an alias)
type Celsius float64
type Fahrenheit float64
type Meters float64
type Miles float64

// The types are INCOMPATIBLE even though both are float64 underneath!
var temp Celsius = 100
var temp2 Fahrenheit = 212

// temp = temp2  // COMPILE ERROR: cannot use Fahrenheit as Celsius
// Go's type system prevents unit mix-ups that kill Mars probes!

// Explicit conversion (only via the same underlying type)
temp3 := Fahrenheit(temp) // OK — same underlying type
_ = temp3

// Methods on custom types
func (c Celsius) ToFahrenheit() Fahrenheit {
    return Fahrenheit(c*9/5 + 32)
}

func (f Fahrenheit) ToCelsius() Celsius {
    return Celsius((f - 32) * 5 / 9)
}

func main() {
    body := Celsius(37.0)
    fmt.Printf("%.1f°C = %.1f°F\n", body, body.ToFahrenheit()) // 37.0°C = 98.6°F
}
```

### Type Aliases (Different from Type Definitions!)

```go
// Type alias — just another name for the SAME type (Go 1.9+)
type MyInt = int      // alias
type NewInt int       // new type

var a MyInt = 5
var b int = a   // OK — MyInt IS int (alias)

var c NewInt = 5
var d int = c   // ERROR — NewInt is a separate type

// Aliases are mainly for:
// 1. Large-scale refactoring (moving types between packages)
// 2. Compatibility layers
// byte = uint8, rune = int32 are the most famous aliases
```

### Type Conversions — Complete Reference

```go
// Numeric conversions (all explicit)
var i int = 42
var i8 int8 = int8(i)        // possible overflow, no error
var i32 int32 = int32(i)
var i64 int64 = int64(i)
var u uint = uint(i)         // negative int → large positive uint
var f32 float32 = float32(i)
var f64 float64 = float64(i)

// String conversions
var r rune = 'A'
var s string = string(r)     // "A"

// WARNING: int to string gives you the character, not the number!
var n int = 65
var s2 string = string(rune(n)) // "A" (correct — via rune) IMPORTANT
var s3 string = fmt.Sprintf("%d", n) // "65" (what you probably want)

// []byte ↔ string (copies!)
var bytes []byte = []byte("hello") // {'h','e','l','l','o'}
var s4 string = string(bytes)       // "hello"

// []rune ↔ string (copies + UTF-8 decode/encode)
var runes []rune = []rune("Hello, 世界")
var s5 string = string(runes)

// Unsafe conversion (zero-copy, advanced use only)
import "unsafe"
b2 := []byte("hello")
s6 := *(*string)(unsafe.Pointer(&b2)) // zero-copy string view of []byte
// DANGER: s6 is only valid while b2 is alive and unmodified!
```

---

## Chapter 7: Constants — Deeper Than You Think

### Untyped Constants — Go's Secret Weapon

```go
// Untyped constants have a "kind" but no fixed type IMPORTANT
// They take on whatever type is needed in context

const x = 10        // untyped integer constant
const y = 3.14      // untyped floating-point constant
const z = "hello"   // untyped string constant

var a int = x       // x becomes int
var b float64 = x   // x becomes float64 (10.0)
var c int32 = x     // x becomes int32
// var d int = y    // ERROR: y is 3.14, truncated constant 3 doesn't... actually:
var d int = y       // ERROR: y (3.14) truncated to integer

// Untyped constants are MORE flexible than typed ones:
const MaxSize = 1000  // works with any integer type

func process(n int64) {}
func render(n int) {}

process(MaxSize) // OK
render(MaxSize)  // OK
// Both work because MaxSize is untyped — it adapts!

// High precision
const BigPi = 3.14159265358979323846264338327950288
// Go internally represents untyped float constants with at least 256-bit precision!
// They're only rounded when assigned to a variable.
```

### iota — Complete Guide

```go
package main

import "fmt"

// iota starts at 0 in each const block
const (
    A = iota // 0
    B         // 1
    C         // 2
)

// iota with expressions
const (
    _   = iota             // 0 (skip with _)
    KB  = 1 << (10 * iota) // 1 << 10 = 1024
    MB                     // 1 << 20 = 1048576
    GB                     // 1 << 30
    TB                     // 1 << 40
    PB                     // 1 << 50
)

// iota in multiple const blocks — resets each time
const (
    X = iota // 0
)
const (
    Y = iota // 0 (resets!)
)

// Week days
type Weekday int // IMPORTANT
const (
    Sunday Weekday = iota // 0
    Monday                // 1
    Tuesday               // 2
    Wednesday             // 3
    Thursday              // 4
    Friday                // 5
    Saturday              // 6
)

// Bitmask flags
type Permission uint8
const (
    Read    Permission = 1 << iota // 1 (00000001)
    Write                          // 2 (00000010)
    Execute                        // 4 (00000100)
    Delete                         // 8 (00001000)
)

func (p Permission) String() string {
    result := ""
    if p&Read != 0 { result += "r" } else { result += "-" }
    if p&Write != 0 { result += "w" } else { result += "-" }
    if p&Execute != 0 { result += "x" } else { result += "-" }
    if p&Delete != 0 { result += "d" } else { result += "-" }
    return result
}

func main() {
    fmt.Println(KB, MB, GB) // 1024 1048576 1073741824
    
    // Combining bit flags
    userPerm := Read | Write
    fmt.Println(userPerm)     // prints "rw--" (via String() method)
    
    // Checking flags
    if userPerm&Execute == 0 {
        fmt.Println("No execute permission")
    }
    
    // Adding a flag
    userPerm |= Execute
    fmt.Println(userPerm) // "rwx-"
    
    // Removing a flag
    userPerm &^= Write // AND NOT   IMPORTANT
    fmt.Println(userPerm) // "r-x-"
}
```

### Testing Constants

```go
package main

import (
    "math"
    "testing"
)

func TestByteSizes(t *testing.T) {
    tests := []struct {
        name     string
        size     uint64
        expected uint64
    }{
        {"kilobyte", uint64(KB), 1024},
        {"megabyte", uint64(MB), 1024 * 1024},
        {"gigabyte", uint64(GB), 1024 * 1024 * 1024},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.size != tt.expected {
                t.Errorf("%s: got %d, want %d", tt.name, tt.size, tt.expected)
            }
        })
    }
}

func TestPermissionBitmask(t *testing.T) {
    // Test individual flags
    if Read != 1 { t.Errorf("Read should be 1, got %d", Read) }
    if Write != 2 { t.Errorf("Write should be 2, got %d", Write) }
    if Execute != 4 { t.Errorf("Execute should be 4, got %d", Execute) }
    
    // Test combinations
    rw := Read | Write
    if rw&Read == 0 { t.Error("Combined permission should have Read") }
    if rw&Execute != 0 { t.Error("Combined permission should not have Execute") }
    
    // Test removal
    rw &^= Write
    if rw&Write != 0 { t.Error("Write should be removed") }
    if rw&Read == 0 { t.Error("Read should still be present") }
}

func TestIotaWeekdays(t *testing.T) {
    days := []Weekday{Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday}
    
    if len(days) != 7 {
        t.Errorf("expected 7 days, got %d", len(days))
    }
    
    // Verify they're sequential from 0
    for i, day := range days {
        if int(day) != i {
            t.Errorf("day[%d] = %d, want %d", i, day, i)
        }
    }
    
    // Verify they fit in int (for use with arrays)
    if Sunday < 0 || Saturday > math.MaxInt32 {
        t.Error("weekday values out of reasonable range")
    }
}
```

---

## Chapter 8: Zero Values — Why They Matter

Zero values are Go's guarantee: **every variable is always valid**. This eliminates entire classes of bugs (undefined behavior, use of uninitialized memory).

```go
// Every type's zero value:
var i int          // 0
var f float64      // 0.0
var b bool         // false
var s string       // "" — NOT nil, actually empty string
var p *int         // nil
var sl []int       // nil (but len=0, cap=0, safe to range and append)
var m map[string]int // nil (UNSAFE to write to!)
var ch chan int     // nil (send/receive on nil channel blocks forever!)
var fn func()      // nil (calling nil func panics)
var iface interface{} // nil

// Useful zero values:
type Config struct {
    Debug   bool   // false by default — good!
    MaxRetry int   // 0 by default — often you'd want > 0, check for this!
    Timeout time.Duration // 0 by default — means "no timeout", might be OK
}

// Common zero-value gotchas:
func getUser(id int) map[string]interface{} {
    m := map[string]interface{}{} // WRONG: returns non-nil empty map always
    // ...
    return m
}
// Callers can't distinguish "function returned empty" from "user has no data"

// Better: return nil explicitly for "nothing"
func getUser2(id int) (map[string]interface{}, error) {
    // if not found:
    return nil, nil
    // callers check: if result == nil { /* not found */ }
}
```

### Testing Zero Values

```go
func TestZeroValues(t *testing.T) {
    var i int
    var f float64
    var b bool
    var s string
    var sl []int
    var m map[string]int
    
    if i != 0 { t.Errorf("int zero value: got %d", i) }
    if f != 0.0 { t.Errorf("float64 zero value: got %f", f) }
    if b { t.Error("bool zero value should be false") }
    if s != "" { t.Errorf("string zero value: got %q", s) }
    
    // Nil slice is safe to use
    if len(sl) != 0 { t.Errorf("nil slice len: got %d", len(sl)) }
    sl = append(sl, 1) // OK on nil slice
    if len(sl) != 1 { t.Errorf("appended slice len: got %d", len(sl)) }
    
    // Nil map read returns zero value (safe)
    val := m["key"] // returns 0, doesn't panic
    if val != 0 { t.Errorf("nil map read: got %d", val) }
    
    // But nil map write panics — test that it panics correctly
    func() {
        defer func() {
            if r := recover(); r == nil {
                t.Error("expected panic on nil map write")
            }
        }()
        m["key"] = 1 // should panic
    }()
}
```

---

**Summary of Part 2:**
- Go has NO implicit type conversions — prevents entire bug classes
- Integers: use `int` generally, `int64` for timestamps/IDs, `byte` for raw data, `rune` for unicode
- Strings are immutable byte sequences — iterate with `range` for Unicode correctness
- Type definitions create genuinely new types (not aliases)
- Untyped constants are more flexible than typed ones
- `iota` is powerful for enums and bit flags
- Zero values make every variable always valid — but watch out for nil maps/channels
