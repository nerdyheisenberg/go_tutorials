# Complete Go Programming Guide — Part 2: Interfaces, Error Handling & Generics

---

# Chapter 8: Interfaces (The Heart of Go)

Interfaces are THE most important concept in Go. They enable polymorphism, decoupling, and testability.

## Interface Basics

```go
// An interface defines a SET OF METHOD SIGNATURES
type Shape interface {
    Area() float64
    Perimeter() float64
}

// Any type that implements ALL methods of an interface
// automatically satisfies it — NO "implements" keyword needed!

type Circle struct {
    Radius float64
}

func (c Circle) Area() float64 {
    return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
    return 2 * math.Pi * c.Radius
}

type Rectangle struct {
    Width, Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
    return 2 * (r.Width + r.Height)
}

// Both Circle and Rectangle satisfy the Shape interface
func PrintShapeInfo(s Shape) {
    fmt.Printf("Area: %.2f, Perimeter: %.2f\n", s.Area(), s.Perimeter())
}

func main() {
    c := Circle{Radius: 5}
    r := Rectangle{Width: 10, Height: 3}

    PrintShapeInfo(c) // Works!
    PrintShapeInfo(r) // Works!

    // Slice of interfaces
    shapes := []Shape{c, r}
    for _, s := range shapes {
        PrintShapeInfo(s)
    }
}
```

## The Empty Interface `interface{}` and `any`

```go
// interface{} has zero methods — EVERY type satisfies it
// Since Go 1.18, 'any' is an alias for interface{}

func printAnything(v any) {
    fmt.Println(v)
}

func main() {
    printAnything(42)
    printAnything("hello")
    printAnything([]int{1, 2, 3})

    // Container of anything
    var stuff []any
    stuff = append(stuff, 1, "two", 3.0, true)
}
```

## Type Assertions

```go
// Extract the concrete type from an interface
var s Shape = Circle{Radius: 5}

// Type assertion
c := s.(Circle) // panics if s is not a Circle
fmt.Println(c.Radius)

// Safe type assertion (comma-ok pattern)
c, ok := s.(Circle)
if ok {
    fmt.Println("It's a circle with radius:", c.Radius)
} else {
    fmt.Println("Not a circle")
}

// Type switch
func describe(i interface{}) string {
    switch v := i.(type) {
    case int:
        return fmt.Sprintf("integer: %d", v)
    case string:
        return fmt.Sprintf("string: %q", v)
    case Circle:
        return fmt.Sprintf("circle with radius: %.2f", v.Radius)
    case nil:
        return "nil"
    default:
        return fmt.Sprintf("unknown type: %T", v)
    }
}
```

## Interface Composition

```go
// Small, focused interfaces are idiomatic Go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}

// Compose interfaces from smaller ones
type ReadWriter interface {
    Reader
    Writer
}

type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}

// This is exactly how the standard library does it!
// io.Reader, io.Writer, io.ReadWriter, io.ReadWriteCloser
```

## Common Standard Library Interfaces

```go
// You MUST know these:

// fmt.Stringer — like toString() in Java or operator<< in C++
type Stringer interface {
    String() string
}

type User struct {
    Name string
    Age  int
}

func (u User) String() string {
    return fmt.Sprintf("%s (age %d)", u.Name, u.Age)
}
// Now fmt.Println(user) calls user.String() automatically

// error interface — the foundation of error handling
type error interface {
    Error() string
}

// io.Reader — THE most important interface in Go
type Reader interface {
    Read(p []byte) (n int, err error)
}
// Files, HTTP bodies, strings, buffers all implement io.Reader

// io.Writer
type Writer interface {
    Write(p []byte) (n int, err error)
}
// Files, HTTP response writers, buffers all implement io.Writer

// sort.Interface
type Interface interface {
    Len() int
    Less(i, j int) bool
    Swap(i, j int)
}

// http.Handler
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

## Interface Best Practices

```go
// 1. Keep interfaces SMALL (1-3 methods)
// Good:
type Reader interface { Read(p []byte) (n int, err error) }

// Bad:
type DoEverything interface {
    Read() error
    Write() error
    Parse() error
    Validate() error
    Transform() error
    // ... too many methods
}

// 2. Define interfaces where they're USED, not where they're implemented
// The consumer defines the interface, not the producer

// In package "service":
type UserStore interface {  // defined by the consumer
    GetUser(id int) (*User, error)
    SaveUser(u *User) error
}

type UserService struct {
    store UserStore // depends on interface, not concrete type
}

// In package "postgres":
type PostgresStore struct { db *sql.DB }
func (s *PostgresStore) GetUser(id int) (*User, error) { ... }
func (s *PostgresStore) SaveUser(u *User) error { ... }
// Implicitly satisfies UserStore

// 3. Accept interfaces, return structs
// Good:
func ProcessData(r io.Reader) (*Result, error) { ... }
// Bad:
func ProcessData(f *os.File) (*Result, error) { ... } // too specific
```

## Interface Internals (Interview Question!) VVIMPORTANT

An interface value holds TWO things:
1. A pointer to the type information (type descriptor)
2. A pointer to the actual data

```go
var s Shape          // (nil, nil) — nil interface
var c *Circle        // nil pointer
s = c                // (Circle, nil) — NOT a nil interface!

fmt.Println(s == nil) // FALSE! The type is set even though value is nil

// This is a common source of bugs! IMPORTANT
// Lesson: Never assign a typed nil to an interface
```

---

# Chapter 9: Error Handling (Deep Dive)

## The Error Interface

```go
// error is just an interface:
type error interface {
    Error() string
}

// Creating errors
import "errors"

err1 := errors.New("something went wrong")
err2 := fmt.Errorf("failed to process item %d: %w", 42, err1) // %w wraps the error
```

## The Idiomatic Pattern

```go
func readFile(path string) ([]byte, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("readFile: %w", err) // wrap with context
    }
    defer file.Close()

    data, err := io.ReadAll(file)
    if err != nil {
        return nil, fmt.Errorf("readFile: reading %s: %w", path, err)
    }

    return data, nil
}

func main() {
    data, err := readFile("config.json")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(data))
}
```

## Custom Error Types

```go
// Custom error with extra information
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: field %s: %s", e.Field, e.Message)
}

func validateAge(age int) error {
    if age < 0 || age > 150 {
        return &ValidationError{
            Field:   "age",
            Message: fmt.Sprintf("must be between 0 and 150, got %d", age),
        }
    }
    return nil
}

func main() {
    err := validateAge(200)
    if err != nil {
        // Attempt to extract the custom ValidationError type
        var vErr *ValidationError
        if errors.As(err, &vErr) {
            fmt.Printf("Custom Handling -> Field: %s, Message: %s\n", vErr.Field, vErr.Message)
        } else {
            // Generic error print using the Error() method automatically
            fmt.Println("Generic Error:", err)
        }
    }
}

```

## Error Wrapping and Unwrapping (Go 1.13+)

```go
import (
    "errors"
    "fmt"
)

var ErrNotFound = errors.New("not found")
var ErrPermission = errors.New("permission denied")

func getUser(id int) (*User, error) {
    // ... db lookup
    return nil, fmt.Errorf("getUser(%d): %w", id, ErrNotFound) // wraps ErrNotFound
}

func main() {
    _, err := getUser(42)

    // errors.Is — checks if ANY error in the chain matches
    if errors.Is(err, ErrNotFound) {
        fmt.Println("User not found") // This matches!
    }

    // errors.As — extracts a specific error type from the chain
    var valErr *ValidationError
    if errors.As(err, &valErr) {
        fmt.Println("Validation failed on field:", valErr.Field)
    }
}
```

## Multiple Error Handling Patterns

```go
// Pattern 1: Early return (most common)
func process() error {
    data, err := fetchData()
    if err != nil {
        return fmt.Errorf("process: %w", err)
    }

    result, err := transform(data)
    if err != nil {
        return fmt.Errorf("process transform: %w", err)
    }

    return save(result)
}

// Pattern 2: errors.Join (Go 1.20+) — multiple errors
func validateUser(u User) error {
    var errs []error
    if u.Name == "" {
        errs = append(errs, errors.New("name is required"))
    }
    if u.Email == "" {
        errs = append(errs, errors.New("email is required"))
    }
    if u.Age < 0 {
        errs = append(errs, errors.New("age must be positive"))
    }
    return errors.Join(errs...) // returns nil if errs is empty
}
```

---

# Chapter 10: Generics (Go 1.18+)

## Type Parameters

```go
// Before generics — had to write separate functions or use interface{}
func maxInt(a, b int) int {
    if a > b { return a }
    return b
}
func maxFloat(a, b float64) float64 {
    if a > b { return a }
    return b
}

// With generics — one function for all comparable types
func Max[T int | float64 | string](a, b T) T {
    if a > b {
        return a
    }
    return b
}

func main() {
    fmt.Println(Max(3, 5))         // 5 (int)
    fmt.Println(Max(3.14, 2.71))   // 3.14 (float64)
    fmt.Println(Max("hello", "world")) // "world" (string)
}
```

## Type Constraints

```go
// Built-in constraints (from the "constraints" package concept)
// comparable — types that support == and !=
// any — alias for interface{}

// Custom constraint using interface
type Number interface {
    int | int8 | int16 | int32 | int64 |
    float32 | float64
}

func Sum[T Number](nums []T) T {
    var total T
    for _, n := range nums {
        total += n
    }
    return total
}

// Using ~ for underlying types IMPORTANT
type MyInt int

type Numeric interface {
    ~int | ~int32 | ~int64 | ~float32 | ~float64
    // ~int means "any type whose underlying type is int" 
    // This allows MyInt to satisfy Numeric
}

// Ordered constraint (very common)
type Ordered interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
    ~float32 | ~float64 | ~string
}

func Min[T Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}
```

## Generic Data Structures

```go
// Generic Stack
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}

func (s *Stack[T]) Peek() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Size() int {
    return len(s.items)
}

// Usage
func main() {
    intStack := Stack[int]{}
    intStack.Push(1)
    intStack.Push(2)
    val, _ := intStack.Pop() // 2

    strStack := Stack[string]{}
    strStack.Push("hello")

    _ = val
}

// Generic Map/Filter/Reduce
func Map[T any, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

func Filter[T any](slice []T, fn func(T) bool) []T {
    var result []T
    for _, v := range slice {
        if fn(v) {
            result = append(result, v)
        }
    }
    return result
}

func Reduce[T any, U any](slice []T, initial U, fn func(U, T) U) U {
    result := initial
    for _, v := range slice {
        result = fn(result, v)
    }
    return result
}

// Usage
nums := []int{1, 2, 3, 4, 5}
doubled := Map(nums, func(n int) int { return n * 2 })
evens := Filter(nums, func(n int) bool { return n%2 == 0 })
sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
```

---

# Chapter 11: Packages, Modules & Visibility

## Package System

```go
// RULE 1: One package per directory
// RULE 2: All files in a directory must have the same package name
// RULE 3: Package name = last element of import path (by convention)

// File: mathutil/calc.go
package mathutil

// RULE 4: CAPITALIZED names are exported (public)
// lowercase names are unexported (private to the package)
func Add(a, b int) int { return a + b }      // Public
func subtract(a, b int) int { return a - b } // Private

type User struct {
    Name  string // Public field
    email string // Private field
}

// Using in main.go:
import "github.com/yourname/project/mathutil"

result := mathutil.Add(1, 2)       // OK
// result := mathutil.subtract(1, 2) // COMPILE ERROR — unexported
```

## Module Management

```bash
# Initialize a module
go mod init github.com/yourname/project

# Add a dependency
go get github.com/gin-gonic/gin@latest

# Remove unused dependencies
go mod tidy

# Vendor dependencies (copy into project)
go mod vendor

# Update all dependencies
go get -u ./...

# View dependency tree
go mod graph
```

## Internal Packages

```
project/
├── internal/        # Cannot be imported by code outside 'project'
│   ├── auth/
│   └── database/
├── pkg/             # Can be imported by anyone
│   └── utils/
└── main.go
```

---

# Chapter 12: Testing

## Basic Tests

```go
// File: calc.go
package calc

func Add(a, b int) int { return a + b }
func Multiply(a, b int) int { return a * b }

// File: calc_test.go (test files end with _test.go)
package calc

import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }
}

// Table-driven tests (THE Go testing pattern)
func TestMultiply(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 2, 3, 6},
        {"with zero", 5, 0, 0},
        {"negative numbers", -2, 3, -6},
        {"both negative", -2, -3, 6},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Multiply(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Multiply(%d, %d) = %d; want %d",
                    tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

```bash
# Run tests
go test ./...

# With verbose output
go test -v ./...

# Run specific test
go test -run TestMultiply ./...

# With coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # visual coverage report
```

## Benchmarks

```go
func BenchmarkAdd(b *testing.B) {  IMPORTANT
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}
```

```bash
go test -bench=. ./...
```

## Test Helpers and Mocking with Interfaces

```go
// The interface makes testing easy — mock the dependency
type UserRepository interface {
    GetByID(id int) (*User, error)
}

// Production implementation
type PostgresUserRepo struct { db *sql.DB }
func (r *PostgresUserRepo) GetByID(id int) (*User, error) { /* real DB */ }

// Test mock
type MockUserRepo struct {
    users map[int]*User
}
func (m *MockUserRepo) GetByID(id int) (*User, error) {
    u, ok := m.users[id]
    if !ok {
        return nil, errors.New("not found")
    }
    return u, nil
}

// Service uses the interface
type UserService struct {
    repo UserRepository
}

func (s *UserService) GetUserName(id int) (string, error) {
    user, err := s.repo.GetByID(id)
    if err != nil {
        return "", err
    }
    return user.Name, nil
}

// Test
func TestGetUserName(t *testing.T) {
    mock := &MockUserRepo{
        users: map[int]*User{
            1: {Name: "Alice"},
        },
    }
    service := &UserService{repo: mock}

    name, err := service.GetUserName(1)
    if err != nil {
        t.Fatal(err)
    }
    if name != "Alice" {
        t.Errorf("got %s, want Alice", name)
    }
}
```
