# Complete Go Programming Guide — Part 1: Foundations
## From Zero to Expert

---

# Chapter 1: Introduction & Setup

## What is Go?
Go (or Golang) is a statically typed, compiled language designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson. It was created to solve real-world problems at scale: fast compilation, efficient concurrency, and simplicity.

### Key Characteristics
- **Statically typed** — types are checked at compile time > true
- **Compiled** — produces a single static binary (no runtime dependencies) > true
- **Garbage collected** — automatic memory management
- **Concurrent by design** — goroutines and channels are first-class citizens
- **Simple** — only 25 keywords in the entire language

### The 25 Keywords
```
break    default     func    interface  select
case     defer       go      map        struct
chan     else        goto    package    switch
const   fallthrough if      range      type
continue for        import  return     var
```
That's it. Compare this to C++ which has 90+ keywords.

## Installation & First Program

```bash
# Install Go (Linux)
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Verify
go version
```

### Your First Go Program
```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Go!")
}
```

```bash
# Run directly
go run main.go

# Build binary and run
go build -o myapp main.go
./myapp
```

### Go Module System (Project Setup)
Every Go project needs a module. This is like a `package.json` or `CMakeLists.txt`.

```bash
mkdir myproject && cd myproject
go mod init github.com/yourname/myproject
```

This creates a `go.mod` file:
```
module github.com/yourname/myproject

go 1.22
```

### Project Structure Convention
```
myproject/
├── go.mod
├── go.sum          # dependency checksums (auto-generated)
├── main.go         # entry point
├── cmd/            # multiple entry points
│   ├── server/
│   │   └── main.go
│   └── cli/
│       └── main.go
├── internal/       # private packages (cannot be imported by others)
│   ├── handler/
│   └── service/
├── pkg/            # public reusable packages
│   └── utils/
└── api/            # API definitions (protobuf, OpenAPI)
```

---

# Chapter 2: Variables, Types & Constants

## Variable Declaration — 4 Ways

```go
package main

import "fmt"

func main() {
    // Way 1: Full declaration
    var name string = "Rohit"

    // Way 2: Type inference with var
    var age = 30 // compiler infers int

    // Way 3: Short declaration (MOST COMMON — only inside functions)
    city := "Delhi"

    // Way 4: Multiple declarations
    var (
        x    int     = 10
        y    float64 = 3.14
        flag bool    = true
    )

    fmt.Println(name, age, city, x, y, flag)
}
```

### Zero Values (Very Important!)
In Go, every variable has a **zero value** if not explicitly initialized. There is no "undefined" or "garbage" value.

```go
var i int       // 0
var f float64   // 0.0
var b bool      // false
var s string    // "" (empty string)
var p *int      // nil
var sl []int    // nil (nil slice)
var m map[string]int // nil (nil map — CAREFUL: writing to nil map panics!)
```

## Basic Types

```go
// Integers
int     int8    int16   int32   int64
uint    uint8   uint16  uint32  uint64

// Aliases
byte    // alias for uint8
rune    // alias for int32 (represents a Unicode code point)

// Floating point
float32 float64

// Complex numbers
complex64 complex128

// Boolean
bool

// String (immutable sequence of bytes, UTF-8 encoded)
string
```

### Type Conversions (No Implicit Casting!)
Go has NO implicit type conversion. Every conversion must be explicit.

```go
var i int = 42
var f float64 = float64(i)  // explicit conversion required
var u uint = uint(f)        // explicit conversion required

// This WILL NOT compile:
// var f float64 = i   // ERROR: cannot use i (type int) as type float64
```

### String Operations
```go
package main

import (
    "fmt"
    "strings"
)

func main() {
    s := "Hello, World"

    // Length (in bytes, not characters!)
    fmt.Println(len(s)) // 12

    // For character (rune) count:
    fmt.Println(len([]rune(s))) // 12 for ASCII, different for unicode

    // Substring (slicing)
    fmt.Println(s[0:5]) // "Hello"

    // Concatenation
    full := "Hello" + " " + "World"

    // String builder (efficient for many concatenations)
    var builder strings.Builder
    for i := 0; i < 1000; i++ {
        builder.WriteString("hello ")
    }
    result := builder.String()

    // Common string functions
    fmt.Println(strings.Contains(s, "World"))    // true
    fmt.Println(strings.HasPrefix(s, "Hello"))   // true
    fmt.Println(strings.HasSuffix(s, "World"))   // true
    fmt.Println(strings.ToUpper(s))              // "HELLO, WORLD"
    fmt.Println(strings.Split("a,b,c", ","))     // ["a" "b" "c"]
    fmt.Println(strings.Replace(s, "World", "Go", 1)) // "Hello, Go"
    fmt.Println(strings.TrimSpace("  hello  "))  // "hello"

    _ = full
    _ = result
}
```

### Runes and UTF-8
```go
package main

import "fmt"

func main() {
    s := "Hello, 世界" // Chinese characters

    // Iterating by BYTES (wrong for unicode)
    for i := 0; i < len(s); i++ {
        fmt.Printf("%x ", s[i]) // prints raw bytes
    }
    fmt.Println()

    // Iterating by RUNES (correct for unicode)
    for index, runeValue := range s {
        fmt.Printf("index=%d rune=%c unicode=%U\n", index, runeValue, runeValue)
    }

    // Converting
    r := rune('A')
    fmt.Println(r)         // 65
    fmt.Println(string(r)) // "A"
}
```

## Constants

```go
package main

import "fmt"

// Constants are immutable, set at compile time
const Pi = 3.14159
const AppName = "MyApp"

// Grouped constants
const (
    StatusOK    = 200
    StatusError = 500
)

// iota — Auto-incrementing constant generator (very powerful!)
const (
    Sunday    = iota // 0
    Monday           // 1
    Tuesday          // 2
    Wednesday        // 3
    Thursday         // 4
    Friday           // 5
    Saturday         // 6
)

// iota with expressions
const (
    _  = iota             // 0 (skip)
    KB = 1 << (10 * iota) // 1 << 10 = 1024
    MB                    // 1 << 20 = 1048576
    GB                    // 1 << 30
    TB                    // 1 << 40
)

// iota for bit flags
const (
    ReadPermission   = 1 << iota // 1
    WritePermission              // 2
    ExecutePermission            // 4
)

func main() {
    fmt.Println("KB:", KB) // 1024
    fmt.Println("MB:", MB) // 1048576

    // Combining bit flags
    permissions := ReadPermission | WritePermission
    fmt.Println("Can read?", permissions&ReadPermission != 0) // true
    fmt.Println("Can execute?", permissions&ExecutePermission != 0) // false
}
```

### Untyped Constants (Unique to Go)
```go
const x = 5 // untyped constant — has no fixed type yet

var i int = x       // works: x becomes int
var f float64 = x   // works: x becomes float64
var c complex128 = x // works: x becomes complex128

// This flexibility only applies to UNTYPED constants
const y int = 5
// var f2 float64 = y  // ERROR: y is typed as int
```

---

# Chapter 3: Control Flow

## If/Else

```go
// Standard if/else
if x > 10 {
    fmt.Println("big")
} else if x > 5 {
    fmt.Println("medium")
} else {
    fmt.Println("small")
}

// If with initialization statement (VERY COMMON in Go)
if err := doSomething(); err != nil {
    fmt.Println("Error:", err)
    // err is scoped to this if/else block
}
// err is NOT accessible here
```

## Switch

```go
// Basic switch (no need for 'break' — it's implicit!)
switch day {
case "Monday":
    fmt.Println("Start of week")
case "Friday":
    fmt.Println("TGIF!")
case "Saturday", "Sunday": // multiple values
    fmt.Println("Weekend!")
default:
    fmt.Println("Midweek")
}

// Switch with no condition (replaces long if/else chains)
switch {
case score >= 90:
    grade = "A"
case score >= 80:
    grade = "B"
case score >= 70:
    grade = "C"
default:
    grade = "F"
}

func getGrade3(score int) string {
    switch {
    case score >= 90: return "A"
    case score >= 80: return "B"
    case score >= 70: return "C"
    default: return "F"
    }
}

// Switch with initialization
switch os := runtime.GOOS; os {
case "linux":
    fmt.Println("Linux")
case "darwin":
    fmt.Println("macOS")
}

// fallthrough — explicitly falls through to next case (rare)
switch x {
case 1:
    fmt.Println("one")
    fallthrough // explicitly falls through
case 2:
    fmt.Println("two") // this runs even if x == 1
}

// Type switch (used with interfaces — covered later)
switch v := i.(type) {
case int:
    fmt.Println("int:", v)
case string:
    fmt.Println("string:", v)
default:
    fmt.Println("unknown")
}
```

## For Loop (The ONLY loop in Go)

```go
// Traditional for
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// While-style
count := 0
for count < 10 {
    count++
}

// Infinite loop
for {
    // do forever
    if shouldStop {
        break
    }
}

// Range-based iteration (VERY IMPORTANT)
// Over slices
nums := []int{10, 20, 30, 40}
for index, value := range nums {
    fmt.Printf("index=%d value=%d\n", index, value)
}

// Skip index
for _, value := range nums {
    fmt.Println(value)
}

// Skip value (just iterate indices)
for index := range nums {
    fmt.Println(index)
}

// Over maps
m := map[string]int{"a": 1, "b": 2}
for key, value := range m {
    fmt.Printf("%s: %d\n", key, value)
}

// Over strings (iterates runes, not bytes)
for i, ch := range "Hello, 世界" {
    fmt.Printf("%d: %c\n", i, ch)
}

// Over channels (blocks until channel is closed)
ch := make(chan int)
for val := range ch {
    fmt.Println(val)
}
```

### Labels, Break, Continue
```go
// Labeled break (break out of nested loops)
outer:
    for i := 0; i < 5; i++ {
        for j := 0; j < 5; j++ {
            if i*j > 6 {
                break outer // breaks out of BOTH loops
            }
        }
    }

// Continue skips to next iteration
for i := 0; i < 10; i++ {
    if i%2 == 0 {
        continue // skip even numbers
    }
    fmt.Println(i) // only prints odd numbers
}
```

---

# Chapter 4: Functions

## Basic Functions

```go
// Simple function
func add(a int, b int) int {
    return a + b
}

// Shortened parameter list (same type)
func add(a, b int) int {
    return a + b
}

// Multiple return values
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// Named return values (use sparingly — good for short functions)
func divide(a, b float64) (result float64, err error) {
    if b == 0 {
        err = errors.New("division by zero")
        return // "naked return" — returns named values
    }
    result = a / b
    return
}
```

## Variadic Functions

```go
// Accepts any number of int arguments
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

func main() {
    fmt.Println(sum(1, 2, 3))       // 6
    fmt.Println(sum(1, 2, 3, 4, 5)) // 15

    // Passing a slice to variadic function
    numbers := []int{1, 2, 3, 4}
    fmt.Println(sum(numbers...)) // spread operator
}
```

## Functions as First-Class Values

```go
package main

import "fmt"

func main() {
    // Assigning function to variable
    greet := func(name string) string {
        return "Hello, " + name
    }
    fmt.Println(greet("Rohit"))

    // Passing function as argument
    apply := func(nums []int, fn func(int) int) []int {
        result := make([]int, len(nums))
        for i, v := range nums {
            result[i] = fn(v)
        }
        return result
    }

    doubled := apply([]int{1, 2, 3}, func(n int) int {
        return n * 2
    })
    fmt.Println(doubled) // [2 4 6]
}
```

## Closures

```go
package main

import "fmt"

// Function that returns a function (closure) IMPORTANT
func makeCounter() func() int {
    count := 0 // captured by the closure
    return func() int {
        count++
        return count
    }
}

func main() {
    counter := makeCounter()
    fmt.Println(counter()) // 1
    fmt.Println(counter()) // 2
    fmt.Println(counter()) // 3

    // Each call to makeCounter creates a NEW closure
    counter2 := makeCounter()
    fmt.Println(counter2()) // 1 (independent counter)
}

// Practical closure: middleware pattern
func withLogging(fn func(string) string) func(string) string {
    return func(input string) string {
        fmt.Println("Input:", input)
        result := fn(input)
        fmt.Println("Output:", result)
        return result
    }
}
```

## Defer

`defer` schedules a function call to run just before the enclosing function returns. It's Go's replacement for RAII / destructors.

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    // Basic defer
    fmt.Println("start")
    defer fmt.Println("deferred") // runs LAST
    fmt.Println("end")
    // Output: start, end, deferred

    // LIFO order — deferred calls stack
    defer fmt.Println("first defer")
    defer fmt.Println("second defer")
    defer fmt.Println("third defer")
    // Output: third defer, second defer, first defer

    // Practical: File handling
    file, err := os.Open("data.txt")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer file.Close() // guaranteed cleanup, even if panic occurs IMPORTANT

    // Read file...
}

// Defer evaluates arguments immediately, but executes later
func deferTrap() { IMPORTANT
    x := 10
    defer fmt.Println("deferred x =", x) // captures x=10 NOW
    x = 20
    fmt.Println("current x =", x)
    // Output: current x = 20
    //         deferred x = 10  (NOT 20!)
}

// To capture the FINAL value, use a closure
func deferWithClosure() { IMPORTANT
    x := 10
    defer func() {
        fmt.Println("deferred x =", x) // captures x by reference
    }()
    x = 20
    // Output: deferred x = 20
}
```

## Panic and Recover

`panic` is like an unhandled exception. `recover` catches it (but only inside a deferred function).

```go
package main

import "fmt"

func safeDivide(a, b int) (result int, err error) {
    // Recover from panic
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered from panic: %v", r)
        }
    }()

    // This will panic if b == 0
    return a / b, nil
}

func main() {
    result, err := safeDivide(10, 0)
    if err != nil {
        fmt.Println("Error:", err) // "recovered from panic: runtime error: integer divide by zero"
    } else {
        fmt.Println("Result:", result)
    }
}

// When to use panic:
// 1. Truly unrecoverable situations (corrupted state)
// 2. Programming errors that should fail fast
// 3. Init functions that can't proceed
// NEVER use panic for normal error handling — use error returns instead
```

## Init Functions

```go
package main

import "fmt"

// init() runs automatically before main(), used for setup
func init() {
    fmt.Println("Initializing...")
    // Setup database connections, load config, etc.
}

// You can have multiple init() in one file or package
func init() {
    fmt.Println("Second init")
}

func main() {
    fmt.Println("Main function")
}
// Output:
// Initializing...
// Second init
// Main function
```

---

# Chapter 5: Pointers

## Pointer Basics

```go
package main

import "fmt"

func main() {
    x := 42
    p := &x     // p is a pointer to x (type: *int)
    fmt.Println(*p)  // 42 — dereference the pointer
    *p = 100         // modify x through the pointer
    fmt.Println(x)   // 100

    // Zero value of pointer is nil
    var q *int
    fmt.Println(q)       // <nil>
    fmt.Println(q == nil) // true
    // fmt.Println(*q)   // PANIC: nil pointer dereference

    // new() allocates memory and returns a pointer
    n := new(int)   // n is *int, pointing to a zero-valued int
    *n = 42
    fmt.Println(*n) // 42
}
```

## Value vs Pointer — When to Use Which

```go
// Pass by VALUE — function gets a COPY
func doubleValue(x int) {
    x = x * 2 // modifies the copy, not the original
}

// Pass by POINTER — function gets the address
func doublePointer(x *int) {
    *x = *x * 2 // modifies the original
}

func main() {
    a := 10
    doubleValue(a)
    fmt.Println(a) // still 10

    doublePointer(&a)
    fmt.Println(a) // now 20
}
```

### Rules of Thumb for Pointer vs Value: IMPORTANT
1. **Use pointers** when you need to modify the receiver/argument
2. **Use pointers** for large structs (avoids copying)
3. **Use values** for small, immutable data (int, bool, small structs)
4. **Use pointers** if any method needs a pointer receiver (consistency)
5. **Slices, maps, channels** are already reference types — don't need pointers

## No Pointer Arithmetic
```go
p := &x
// p++     // COMPILE ERROR — no pointer arithmetic in Go
// p + 1   // COMPILE ERROR
// This prevents entire classes of bugs from C/C++
```

## Escape Analysis
```go
// In C++, returning a pointer to a local variable = dangling pointer = crash
// In Go, the compiler detects this and allocates on the heap automatically

func createUser() *User {
    u := User{Name: "Rohit"} // would be stack in C++
    return &u                 // Go detects escape → allocates on heap
    // Perfectly safe!
}
```

---

# Chapter 6: Data Structures Deep Dive

## Arrays

```go
// Arrays have FIXED size — size is part of the type
var a [5]int                  // [0, 0, 0, 0, 0]
b := [3]string{"go", "is", "fun"}
c := [...]int{1, 2, 3, 4, 5} // compiler counts the elements

// Arrays are VALUES in Go — assignment copies the entire array
x := [3]int{1, 2, 3}
y := x      // y is a COPY of x
y[0] = 999
fmt.Println(x[0]) // still 1
fmt.Println(y[0]) // 999

// [3]int and [5]int are DIFFERENT types — cannot assign one to the other IMPORTANT
```

## Slices (The Most Important Data Structure)

A slice is a flexible, dynamically-sized view into an underlying array. Internally it's a struct with three fields:

```
type slice struct {
    array unsafe.Pointer  // pointer to underlying array
    len   int            // number of elements
    cap   int            // capacity of underlying array
}
```

```go
package main

import "fmt"

func main() {
    // Creating slices
    s1 := []int{1, 2, 3, 4, 5}          // literal
    s2 := make([]int, 5)                  // length=5, cap=5, all zeros
    s3 := make([]int, 0, 10)              // length=0, cap=10
    var s4 []int                          // nil slice (length=0, cap=0)

    // nil slice vs empty slice
    fmt.Println(s4 == nil)                // true
    fmt.Println(len(s4), cap(s4))         // 0 0
    s5 := []int{}                         // empty slice, NOT nil
    fmt.Println(s5 == nil)                // false
    fmt.Println(len(s5), cap(s5))         // 0 0
    // Both work with append, range — prefer nil slice

    // Append (may allocate new array if capacity exceeded)
    s := []int{1, 2, 3}
    fmt.Println(len(s), cap(s)) // 3 3
    s = append(s, 4)
    fmt.Println(len(s), cap(s)) // 4 6 (capacity doubled!)
    s = append(s, 5, 6, 7)
    fmt.Println(len(s), cap(s)) // 7 12

    // Append one slice to another
    a := []int{1, 2}
    b := []int{3, 4}
    a = append(a, b...) // spread operator
    fmt.Println(a) // [1 2 3 4]

    _ = s1
    _ = s2
    _ = s3
}
```

### Slice Internals — The Gotcha

```go
// Slices share underlying arrays!
original := []int{1, 2, 3, 4, 5}
sub := original[1:3] // [2, 3] — shares the same array!

sub[0] = 999
fmt.Println(original) // [1 999 3 4 5] — original is modified! IMPORTANT

// To avoid this, use copy() or full slice expression
// Method 1: copy
sub2 := make([]int, 2)
copy(sub2, original[1:3])
sub2[0] = 888
fmt.Println(original) // [1 999 3 4 5] — original unchanged IMPORTANT

// Method 2: Full slice expression (limits capacity)
sub3 := original[1:3:3] // [low:high:max] — cap = max - low = 2 IMPORTANT
// append to sub3 will allocate new array instead of overwriting original
```

### Common Slice Operations

```go
// Delete element at index i
s := []int{1, 2, 3, 4, 5}
i := 2
s = append(s[:i], s[i+1:]...) // [1 2 4 5] IMPORTANT

// Insert element at index i
s = []int{1, 2, 4, 5}
i = 2
s = append(s[:i], append([]int{3}, s[i:]...)...) // [1 2 3 4 5]

// Filter
nums := []int{1, 2, 3, 4, 5, 6}
var evens []int
for _, n := range nums {
    if n%2 == 0 {
        evens = append(evens, n)
    }
}

// Reverse
for left, right := 0, len(s)-1; left < right; left, right = left+1, right-1 {
    s[left], s[right] = s[right], s[left]
}
```

## Maps

```go
package main

import "fmt"

func main() {
    // Creation
    m1 := make(map[string]int)          // empty map
    m2 := map[string]int{               // literal
        "alice": 25,
        "bob":   30,
    }

    // IMPORTANT: nil map vs empty map
    var m3 map[string]int    // nil map
    // m3["key"] = 1         // PANIC! Cannot write to nil map
    m3 = make(map[string]int) // now it's initialized
    m3["key"] = 1             // OK

    // Read (returns zero value if key doesn't exist)
    age := m2["alice"] // 25
    missing := m2["charlie"] // 0 (zero value, NOT an error)

    // Check existence (comma ok pattern)
    val, ok := m2["charlie"]
    if !ok {
        fmt.Println("charlie not found")
    }

    // Delete
    delete(m2, "alice")

    // Iterate (order is RANDOM — by design!)
    for key, value := range m2 {
        fmt.Printf("%s: %d\n", key, value)
    }

    // Length
    fmt.Println(len(m2))

    // Maps are NOT safe for concurrent use — use sync.Map or mutex
    _ = age
    _ = missing
    _ = val
    _ = m1
}
```

### Map with Struct Values
```go
type Student struct {
    Name string
    Age  int
}

students := map[int]Student{
    1: {Name: "Alice", Age: 20},
    2: {Name: "Bob", Age: 22},
}

// GOTCHA: Cannot modify struct fields in map directly
// students[1].Age = 21  // COMPILE ERROR

// Solution: Replace the entire value
s := students[1]
s.Age = 21
students[1] = s

// Or use pointer values
students2 := map[int]*Student{
    1: {Name: "Alice", Age: 20},
}
students2[1].Age = 21 // OK with pointers
```

### Sets (Using Maps)
Go has no built-in set. Use a map with empty struct values:

```go
// Set of strings
set := make(map[string]struct{})
set["apple"] = struct{}{}
set["banana"] = struct{}{}

// Check membership
if _, exists := set["apple"]; exists {
    fmt.Println("apple is in the set")
}

// Why struct{} and not bool?
// struct{} takes zero bytes of memory
```

---

# Chapter 7: Structs and Methods

## Struct Basics

```go
// Defining a struct
type User struct {
    ID        int
    FirstName string
    LastName  string
    Email     string
    Age       int
}

func main() {
    // Creating structs
    u1 := User{
        ID:        1,
        FirstName: "Rohit",
        LastName:  "Kumar",
        Email:     "rohit@example.com",
        Age:       30,
    }

    // Positional (not recommended — brittle)
    u2 := User{2, "Alice", "Smith", "alice@example.com", 25}

    // Partial initialization (rest get zero values)
    u3 := User{FirstName: "Bob"}
    fmt.Println(u3.Age) // 0

    // Pointer to struct
    u4 := &User{FirstName: "Charlie"}
    fmt.Println(u4.FirstName) // "Charlie" — auto-dereferenced (no -> needed)

    // Anonymous struct (useful for one-off use)
    config := struct {
        Host string
        Port int
    }{
        Host: "localhost",
        Port: 8080,
    }
    fmt.Println(config.Host)

    _ = u1
    _ = u2
}
```

## Methods (Functions Attached to Types)

```go
type Rectangle struct {
    Width, Height float64
}

// Value receiver — operates on a COPY
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

// Pointer receiver — operates on the ORIGINAL
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}

func main() {
    rect := Rectangle{Width: 10, Height: 5}
    fmt.Println(rect.Area()) // 50

    rect.Scale(2)
    fmt.Println(rect.Area()) // 200

    // Go automatically handles &rect when calling pointer receiver methods
    // rect.Scale(2) is syntactic sugar for (&rect).Scale(2)
}
```

### When to Use Pointer vs Value Receivers:
1. **Pointer receiver** if the method modifies the receiver
2. **Pointer receiver** if the struct is large (avoids copying)
3. **Pointer receiver** if any method has pointer receiver (be consistent)
4. **Value receiver** for small, immutable types (time.Time uses value receivers)

## Struct Embedding (Composition over Inheritance)

```go
// Go does NOT have inheritance. It has EMBEDDING (composition).

type Animal struct {
    Name string
}

func (a Animal) Speak() string {
    return a.Name + " makes a sound"
}

type Dog struct {
    Animal       // embedded — Dog "inherits" all fields and methods of Animal
    Breed string
}

func main() {
    d := Dog{
        Animal: Animal{Name: "Rex"},
        Breed:  "German Shepherd",
    }

    // Access embedded fields directly
    fmt.Println(d.Name)    // "Rex" (promoted from Animal)
    fmt.Println(d.Speak()) // "Rex makes a sound" (promoted method)
    fmt.Println(d.Breed)   // "German Shepherd"

    // Can also access explicitly
    fmt.Println(d.Animal.Name) // "Rex"
}

// Method overriding via shadowing
func (d Dog) Speak() string {
    return d.Name + " barks!"
}
// Now d.Speak() returns "Rex barks!"
// d.Animal.Speak() still returns "Rex makes a sound"
```

### Multiple Embedding

```go
type Logger struct{}
func (l Logger) Log(msg string) { fmt.Println("LOG:", msg) }

type Metrics struct{}
func (m Metrics) Record(name string) { fmt.Println("METRIC:", name) }

type Server struct {
    Logger  // embed Logger
    Metrics // embed Metrics
    Port int
}

func main() {
    s := Server{Port: 8080}
    s.Log("started")         // from Logger
    s.Record("request_count") // from Metrics
}
```

## Struct Tags (JSON, Database, Validation)

```go
type User struct {
    ID        int    `json:"id" db:"user_id"`
    FirstName string `json:"first_name" db:"first_name"`
    LastName  string `json:"last_name" db:"last_name"`
    Password  string `json:"-"` // "-" means exclude from JSON
    Email     string `json:"email,omitempty"` // omit if empty
}

type User struct{
    ID int `json:"id"`
           `json:"-"` not included in json
           `json:"email,omitempty"`
}

// Using with JSON
import "encoding/json"

u := User{ID: 1, FirstName: "Rohit", Email: ""}
data, _ := json.Marshal(u)
fmt.Println(string(data))
// {"id":1,"first_name":"Rohit","last_name":""}
// Note: email is omitted because of omitempty

// Unmarshalling JSON to struct
jsonStr := `{"id":2,"first_name":"Alice"}`
var u2 User
json.Unmarshal([]byte(jsonStr), &u2)
fmt.Println(u2.FirstName) // "Alice"
```

## Constructor Pattern

Go has no constructors. Use factory functions by convention:

```go
type Server struct {
    host string
    port int
}

// Convention: New<TypeName> returns a pointer
func NewServer(host string, port int) *Server {
    return &Server{
        host: host,
        port: port,
    }
}

// With validation
func NewServer(host string, port int) (*Server, error) {
    if port < 1 || port > 65535 {
        return nil, fmt.Errorf("invalid port: %d", port)
    }
    return &Server{host: host, port: port}, nil
}
```

## Functional Options Pattern (Advanced Constructor)

```go
type Server struct {
    host    string
    port    int
    timeout time.Duration
    maxConn int
}

// Option is a function that configures Server
type Option func(*Server)

func WithPort(port int) Option {
    return func(s *Server) { s.port = port }
}

func WithTimeout(timeout time.Duration) Option {
    return func(s *Server) { s.timeout = timeout }
}

func WithMaxConnections(max int) Option {
    return func(s *Server) { s.maxConn = max }
}

func NewServer(host string, opts ...Option) *Server {
    s := &Server{
        host:    host,
        port:    8080,           // default
        timeout: 30 * time.Second, // default
        maxConn: 100,            // default
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Usage — clean, extensible API
s := NewServer("localhost",
    WithPort(9090),
    WithTimeout(60*time.Second),
)
```
