# Go Deep Dive — Part 8: Interfaces — The Soul of Go

---

## Chapter 1: What Makes Go Interfaces Unique

### Implicit Implementation = Duck Typing

In Java/C++, you explicitly declare that a class implements an interface:
```java
// Java: explicit
class Dog implements Animal { ... }
```

In Go, a type implements an interface **automatically** if it has all the required methods:
```go
// Go: implicit — no "implements" keyword
type Animal interface { Speak() string }
type Dog struct { Name string }
func (d Dog) Speak() string { return "Woof!" }
// Dog automatically satisfies Animal — no declaration needed!
```

**Why this matters:**
1. **Decoupling**: The `Dog` type doesn't need to know about the `Animal` interface
2. **Retroactive implementation**: You can make existing types satisfy new interfaces without modifying them
3. **Cross-package interfaces**: Define interfaces in the *consumer* package, not the *producer*

### The Dependency Inversion Principle in Go

```go
// WRONG way (tight coupling):
// Package: payment
func ProcessPayment(db *postgres.Database, amount float64) error {
    // directly depends on PostgreSQL! Can't test without a DB!
    return db.Execute("INSERT INTO payments...")
}

// RIGHT way (using interface):
// Package: payment
type PaymentStore interface {
    SavePayment(amount float64) error
}

func ProcessPayment(store PaymentStore, amount float64) error {
    return store.SavePayment(amount) // depends on behavior, not implementation
}

// Now you can:
// 1. Use PostgresStore in production
// 2. Use MockStore in tests
// 3. Switch to MySQL without changing ProcessPayment
```

---

## Chapter 2: Interface Mechanics — Complete Reference

### Defining Interfaces

```go
// Single method interface (most common — keep them small!)
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// No methods — the empty interface
type Empty interface{} // any type satisfies this

// Since Go 1.18, 'any' is a built-in alias
type Empty = any  // equivalent

// Composed interface (embedding other interfaces)
type ReadWriter interface {
    Reader
    Writer
}

type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}
```

### What Makes a Type Satisfy an Interface?

```go
type Stringer interface {
    String() string
}

// Satisfied by value receiver:
type Temperature float64
func (t Temperature) String() string {
    return fmt.Sprintf("%.1f°C", float64(t))
}
// Both Temperature AND *Temperature satisfy Stringer

// Satisfied only by pointer receiver:
type Counter struct { n int }
func (c *Counter) String() string {  // pointer receiver!
    return fmt.Sprintf("count: %d", c.n)
}
// *Counter satisfies Stringer, but Counter does NOT
var s1 Stringer = &Counter{}  // OK
// var s2 Stringer = Counter{}   // COMPILE ERROR
```

### Interface Values = (Type, Value) Pair

This is the most important thing to understand about interfaces:

```go
// An interface value is internally: (concrete_type, concrete_value)
var s Stringer

fmt.Println(s == nil)  // true — both type and value are nil

s = Temperature(37.5)
// Now: s = (Temperature, 37.5)
fmt.Println(s == nil)  // false — type is Temperature, value is 37.5
fmt.Println(s)         // "37.5°C" — calls Temperature.String()

// The nil interface trap:
var t *Temperature = nil
s = t
// Now: s = (*Temperature, nil) — type is set!
fmt.Println(s == nil)  // FALSE! The type field is non-nil
// This is a common bug — see below for the full explanation
```

### The Nil Interface Bug — Most Common Go Gotcha

```go
// BUG: This function has a subtle bug
func createUser(admin bool) error {
    var err *ValidationError  // typed nil
    
    if admin {
        err = &ValidationError{Field: "role"}
    }
    
    // BUG: returning typed nil converts to non-nil interface!
    return err  // Returns (*ValidationError, nil) — not a nil interface!
}

func main() {
    err := createUser(false)  // expecting nil
    if err != nil {
        fmt.Println("Error:", err) // This prints! Bug!
    }
}

// FIX: Only return nil for error
func createUserFixed(admin bool) error {
    if admin {
        return &ValidationError{Field: "role"}
    }
    return nil  // returns (nil, nil) — truly nil interface
}

// General rule: Never return a typed nil where an interface is expected
```

---

## Chapter 3: Type Assertions and Type Switches

### Type Assertions

```go
var i interface{} = "hello"

// Single-return form (panics if wrong type)
s := i.(string)  // OK: s = "hello"

// i.(int) would panic: interface conversion: interface {} is string, not int

// Two-return form (safe, comma-ok pattern)
s, ok := i.(string)  // s = "hello", ok = true
n, ok := i.(int)     // n = 0, ok = false (doesn't panic!)

// Always use comma-ok unless you're certain of the type
if s, ok := i.(string); ok {
    fmt.Println("String:", s)
}

// Type assertion is NOT type conversion:
var f float64 = 3.14
var ii interface{} = f

// This is assertion (checking the interface's concrete type):
v := ii.(float64)   // 3.14 — OK, concrete type is float64

// This would fail — the concrete type is float64, not int:
// n := ii.(int)    // PANIC
```

### Type Switches

```go
// Type switch is the idiomatic way to handle multiple types
func describe(i interface{}) string {
    switch v := i.(type) {
    case nil:
        return "nil"
    case int:
        return fmt.Sprintf("int(%d)", v)
    case float64:
        return fmt.Sprintf("float64(%.2f)", v)
    case string:
        return fmt.Sprintf("string(%q, len=%d)", v, len(v))
    case bool:
        return fmt.Sprintf("bool(%t)", v)
    case []int:
        return fmt.Sprintf("[]int(len=%d)", len(v))
    case map[string]interface{}:
        return fmt.Sprintf("map(len=%d)", len(v))
    case fmt.Stringer:  // check interface implementation!
        return fmt.Sprintf("Stringer(%s)", v.String())
    default:
        return fmt.Sprintf("unknown(%T)", v)
    }
}

// The case clauses are checked in ORDER
// For interface cases (like fmt.Stringer), the first matching one wins
```

---

## Chapter 4: Interface Patterns in Practice

### Pattern 1: Accept interfaces, Return concrete types

```go
// Accept the most general interface your function needs
// This maximizes callers that can use it

// BAD: accepts a specific type, limits reuse
func writeToFile(file *os.File, data []byte) error {
    _, err := file.Write(data)
    return err
}
// Can only write to files

// GOOD: accepts io.Writer, works with anything
func writeTo(w io.Writer, data []byte) error {
    _, err := w.Write(data)
    return err
}
// Can write to files, buffers, HTTP responses, strings, etc.

// Return concrete types (not interfaces) from constructors/factories
// BAD:
func NewDB() DBInterface { return &PostgresDB{...} }
// Hides type information, can't use PostgresDB-specific methods

// GOOD:
func NewDB() *PostgresDB { return &PostgresDB{...} }
// Callers can use it as *PostgresDB or as DBInterface
```

### Pattern 2: Defining Interfaces Where They're Used

```go
// Package: orderservice (consumer)
// Define the interface HERE (where you need it)
type InventoryChecker interface {
    IsInStock(productID int, qty int) (bool, error)
}

type OrderService struct {
    inventory InventoryChecker  // depends on the interface
}

func (s *OrderService) PlaceOrder(productID, qty int) error {
    ok, err := s.inventory.IsInStock(productID, qty)
    if err != nil { return err }
    if !ok { return errors.New("out of stock") }
    // ... place order
    return nil
}

// Package: inventory (producer)
// This type HAPPENS to satisfy InventoryChecker — no import needed!
type Service struct { db *sql.DB }

func (s *Service) IsInStock(productID int, qty int) (bool, error) {
    // ... check database
    return true, nil
}
```

### Pattern 3: Small, Focused Interfaces

```go
// AVOID big interfaces — they're hard to satisfy and mock
type BigStore interface { // Bad
    GetUser(id int) (*User, error)
    SaveUser(u *User) error
    DeleteUser(id int) error
    ListUsers() ([]*User, error)
    GetOrder(id int) (*Order, error)
    SaveOrder(o *Order) error
    // ...
}

// PREFER small focused interfaces
type UserGetter interface {
    GetUser(id int) (*User, error)
}

type UserSaver interface {
    SaveUser(u *User) error
}

type UserDeleter interface {
    DeleteUser(id int) error
}

// Compose when needed
type UserStore interface {
    UserGetter
    UserSaver
    UserDeleter
}

// Each function takes only what it needs
func sendWelcomeEmail(g UserGetter, id int) error {
    user, err := g.GetUser(id)
    // ...
}
// sendWelcomeEmail only needs GetUser — UserStore satisfies this
```

### Pattern 4: io.Reader / io.Writer Everywhere

```go
// These two interfaces are the backbone of Go I/O

// EVERYTHING can be an io.Reader:
os.File          // files
strings.NewReader("hello")   // in-memory string
bytes.NewReader([]byte{1,2}) // in-memory bytes
http.Request.Body            // HTTP request body
bufio.NewReader(r)           // buffered reader
gzip.NewReader(r)            // decompressing reader
io.LimitReader(r, n)         // size-limited reader
io.TeeReader(r, w)           // reads r, writes copy to w

// EVERYTHING can be an io.Writer:
os.File          // files
os.Stdout        // standard output
http.ResponseWriter  // HTTP response
bytes.Buffer     // in-memory buffer
gzip.NewWriter(w)   // compressing writer
io.MultiWriter(w1, w2)  // fan-out writer

// Functions that accept io.Reader work with ALL of the above:
func process(r io.Reader) error {
    data, err := io.ReadAll(r)
    if err != nil { return err }
    fmt.Println("Got:", len(data), "bytes")
    return nil
}

// In production:
file, _ := os.Open("data.txt")
process(file)  // reading a file

// In tests:
process(strings.NewReader("test data"))  // no file needed!
```

### Pattern 5: Interface for Testability

```go
// Production code
type EmailSender interface {
    Send(to, subject, body string) error
}

type UserService struct {
    emailer EmailSender
}

func (s *UserService) Register(email, name string) error {
    // ... create user in DB
    return s.emailer.Send(email, "Welcome!", "Hello "+name)
}

// Production implementation
type SMTPSender struct { server string }
func (s *SMTPSender) Send(to, subject, body string) error {
    // ... real SMTP code
    return nil
}

// Test mock (no external dependencies!)
type MockEmailSender struct {
    SentMessages []struct{ To, Subject, Body string }
    ShouldFail   bool
}

func (m *MockEmailSender) Send(to, subject, body string) error {
    if m.ShouldFail {
        return errors.New("send failed")
    }
    m.SentMessages = append(m.SentMessages, struct{ To, Subject, Body string }{to, subject, body})
    return nil
}

// Test:
func TestRegister(t *testing.T) {
    mock := &MockEmailSender{}
    svc := &UserService{emailer: mock}
    
    err := svc.Register("test@example.com", "Alice")
    if err != nil { t.Fatal(err) }
    
    if len(mock.SentMessages) != 1 {
        t.Errorf("expected 1 email, got %d", len(mock.SentMessages))
    }
    if mock.SentMessages[0].To != "test@example.com" {
        t.Error("email sent to wrong address")
    }
}

func TestRegisterEmailFailure(t *testing.T) {
    mock := &MockEmailSender{ShouldFail: true}
    svc := &UserService{emailer: mock}
    
    err := svc.Register("test@example.com", "Alice")
    if err == nil {
        t.Error("expected error when email fails")
    }
}
```

---

## Chapter 5: Standard Library Interface Guide

These are the interfaces you MUST know:

```go
// 1. fmt.Stringer — String representation
type Stringer interface {
    String() string
}
// Implement on any type to control how it prints

// 2. fmt.GoStringer — Go syntax representation
type GoStringer interface {
    GoString() string
}
// For %#v format

// 3. error — Error representation
type error interface {
    Error() string
}

// 4. io.Reader — Read data from source
type Reader interface {
    Read(p []byte) (n int, err error)
}

// 5. io.Writer — Write data to destination
type Writer interface {
    Write(p []byte) (n int, err error)
}

// 6. io.Closer — Close a resource
type Closer interface {
    Close() error
}

// 7. io.ReadWriter, io.ReadCloser, io.WriteCloser, io.ReadWriteCloser
// — Compositions of the above

// 8. io.Seeker — Seek to position in stream
type Seeker interface {
    Seek(offset int64, whence int) (int64, error)
}

// 9. sort.Interface — Sort any collection
type Interface interface {
    Len() int
    Less(i, j int) bool
    Swap(i, j int)
}

// 10. http.Handler — Handle HTTP requests
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}

// 11. http.ResponseWriter — Write HTTP response
type ResponseWriter interface {
    Header() Header
    Write([]byte) (int, error)
    WriteHeader(statusCode int)
}
```

### Implementing sort.Interface

```go
type Person struct {
    Name string
    Age  int
}

// Sort by age
type ByAge []Person

func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Less(i, j int) bool { return a[i].Age < a[j].Age }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// Sort by name
type ByName []Person

func (b ByName) Len() int           { return len(b) }
func (b ByName) Less(i, j int) bool { return b[i].Name < b[j].Name }
func (b ByName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

// Modern alternative with sort.Slice (Go 1.8+):
people := []Person{{"Alice", 30}, {"Bob", 25}, {"Carol", 35}}

sort.Slice(people, func(i, j int) bool {
    return people[i].Age < people[j].Age
})

// Multi-criteria sort:
sort.Slice(people, func(i, j int) bool {
    if people[i].Age != people[j].Age {
        return people[i].Age < people[j].Age
    }
    return people[i].Name < people[j].Name
})
```

---

## Chapter 6: Comprehensive Interface Testing

```go
package interfaces_test

import (
    "errors"
    "testing"
    "strings"
    "fmt"
)

// ==================== INTERFACE SATISFACTION ====================

type Shape interface {
    Area() float64
    Perimeter() float64
}

type Circle struct{ Radius float64 }
func (c Circle) Area() float64      { return 3.14159 * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * 3.14159 * c.Radius }

type Rectangle struct{ Width, Height float64 }
func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

func TestInterfaceSatisfaction(t *testing.T) {
    // Compile-time checks using blank interface assignments
    var _ Shape = Circle{}      // panics at compile if Circle doesn't implement Shape
    var _ Shape = Rectangle{}
    
    // Runtime test: both satisfy Shape
    shapes := []Shape{
        Circle{Radius: 5},
        Rectangle{Width: 4, Height: 3},
    }
    
    for i, s := range shapes {
        if s.Area() <= 0 {
            t.Errorf("shapes[%d].Area() = %v, want > 0", i, s.Area())
        }
    }
}

// ==================== NIL INTERFACE BUG ====================

type MyError struct{ msg string }
func (e *MyError) Error() string { return e.msg }

// Buggy function
func buggyGetError(fail bool) error {
    var err *MyError
    if fail {
        err = &MyError{msg: "something failed"}
    }
    return err  // BUG: returns non-nil interface when fail=false
}

// Fixed function
func fixedGetError(fail bool) error {
    if fail {
        return &MyError{msg: "something failed"}
    }
    return nil  // truly nil interface
}

func TestNilInterfaceBug(t *testing.T) {
    // Demonstrate the bug
    err := buggyGetError(false)
    if err != nil {
        // This is reached! The bug is confirmed
        t.Log("Bug confirmed: returned 'nil' *MyError became non-nil error interface:", err)
    }
    
    // Fixed version
    errFixed := fixedGetError(false)
    if errFixed != nil {
        t.Error("Fixed version should return nil error")
    }
}

// ==================== TYPE ASSERTION ====================

func TestTypeAssertion(t *testing.T) {
    var i interface{} = "hello world"
    
    // Safe assertion
    if s, ok := i.(string); ok {
        if !strings.Contains(s, "world") {
            t.Errorf("expected 'world' in string, got %q", s)
        }
    } else {
        t.Error("expected string type assertion to succeed")
    }
    
    // Failed assertion (safe form)
    n, ok := i.(int)
    if ok {
        t.Errorf("int assertion should fail, got %d", n)
    }
    if n != 0 {
        t.Errorf("failed assertion should return zero value, got %d", n)
    }
}

func TestTypeAssertionPanic(t *testing.T) {
    var i interface{} = "hello"
    
    defer func() {
        if r := recover(); r == nil {
            t.Error("expected panic on wrong type assertion")
        }
    }()
    
    _ = i.(int)  // should panic
}

// ==================== INTERFACE MOCK TESTING ====================

type Logger interface {
    Log(level, message string)
}

type Service struct{ logger Logger }

func (s *Service) DoWork() error {
    s.logger.Log("info", "starting work")
    // ... do work
    s.logger.Log("info", "work complete")
    return nil
}

type MockLogger struct {
    Messages []struct{ Level, Msg string }
}

func (m *MockLogger) Log(level, message string) {
    m.Messages = append(m.Messages, struct{ Level, Msg string }{level, message})
}

func TestServiceLogging(t *testing.T) {
    mock := &MockLogger{}
    svc := &Service{logger: mock}
    
    if err := svc.DoWork(); err != nil {
        t.Fatal(err)
    }
    
    if len(mock.Messages) != 2 {
        t.Errorf("expected 2 log messages, got %d", len(mock.Messages))
    }
    if mock.Messages[0].Level != "info" {
        t.Errorf("first message level = %q, want 'info'", mock.Messages[0].Level)
    }
}

// ==================== SORT.INTERFACE ====================

import "sort"

func TestCustomSort(t *testing.T) {
    type Person struct{ Name string; Age int }
    people := []Person{
        {"Charlie", 35},
        {"Alice", 25},
        {"Bob", 25},
        {"Dave", 30},
    }
    
    // Sort by age, then name
    sort.Slice(people, func(i, j int) bool {
        if people[i].Age != people[j].Age {
            return people[i].Age < people[j].Age
        }
        return people[i].Name < people[j].Name
    })
    
    expected := []Person{
        {"Alice", 25}, {"Bob", 25},  // same age, sorted by name
        {"Dave", 30}, {"Charlie", 35},
    }
    
    for i, p := range people {
        if p != expected[i] {
            t.Errorf("people[%d] = %+v, want %+v", i, p, expected[i])
        }
    }
}

// Compile-time interface check (use in non-test code too!)
// This gives a better compile error than "missing method X"
var _ Shape = (*Circle)(nil)    // verifies *Circle implements Shape
var _ Shape = Circle{}          // verifies Circle implements Shape
var _ Logger = (*MockLogger)(nil) // verifies *MockLogger implements Logger
```

---

**Summary of Part 8:**
- Interfaces are satisfied implicitly — no `implements` keyword
- An interface value = (type, value) pair — nil interface only when both are nil
- The nil interface bug: returning a typed nil pointer as an interface returns non-nil interface
- Type assertions: single-return panics, two-return (comma-ok) is safe
- Type switches: check `v := i.(type)`, first matching case wins
- Keep interfaces small (1-3 methods) — compose larger ones from smaller ones
- Define interfaces where they're USED (consumer), not where they're implemented
- `io.Reader` and `io.Writer` are the most important interfaces to know
- Use interfaces to make code testable — mock the interface in tests
- Use `var _ Interface = (*ConcreteType)(nil)` for compile-time interface checks
