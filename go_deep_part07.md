# Go Deep Dive — Part 7: Structs, Methods & Embedding — OOP in Go

---

## Chapter 1: Why No Classes?

Go deliberately omits classes and classical inheritance. The Go team believed:

1. **Inheritance creates tight coupling** — changing the base class breaks all subclasses
2. **Deep hierarchies are hard to understand** — where does this method actually come from?
3. **Composition is more flexible** — you can change behavior by changing embedded types

Instead of `class B extends A`, Go uses:
- **Struct embedding** for composition
- **Interfaces** for polymorphism
- **Methods** on any named type for behavior

The mantra: **Composition over Inheritance** (a GoF principle that Go takes seriously).

---

## Chapter 2: Structs — Complete Reference

### Defining Structs

```go
// Basic struct
type User struct {
    ID        int
    FirstName string
    LastName  string
    Email     string
    CreatedAt time.Time
    IsActive  bool
}

// Exported (public) vs unexported (private)
type Config struct {
    Host     string  // exported — uppercase first letter
    Port     int     // exported
    password string  // unexported — lowercase (private to package)
}

// Struct with comments (document each field)
type Order struct {
    // ID is the unique order identifier.
    ID int

    // CustomerID references the customer who placed this order.
    CustomerID int

    // Items contains the line items in this order.
    Items []OrderItem

    // TotalAmount is the total in USD cents (avoids float precision issues).
    TotalAmount int64

    // Status is the current order status.
    Status OrderStatus

    // CreatedAt is when the order was placed.
    CreatedAt time.Time
}
```

### Initialization Forms

```go
// Form 1: Named fields (PREFERRED — self-documenting, order-independent)
u := User{
    ID:        1,
    FirstName: "Rohit",
    LastName:  "Kumar",
    Email:     "rohit@example.com",
    IsActive:  true,
}
// Unspecified fields get zero values: CreatedAt = zero time

// Form 2: Positional (AVOID — breaks if fields are added/reordered)
u2 := User{1, "Alice", "Smith", "alice@example.com", time.Now(), true}

// Form 3: Pointer to struct
u3 := &User{
    ID:    2,
    Email: "bob@example.com",
}

// Form 4: new() — zero-value pointer (less common for structs with fields)
u4 := new(User)   // *User, all fields at zero values
u4.ID = 3         // set fields after

// Form 5: Anonymous struct (for one-off use)
config := struct {
    Host string
    Port int
}{
    Host: "localhost",
    Port: 8080,
}
fmt.Println(config.Host) // "localhost"
```

### Field Accessing and Modification

```go
u := User{FirstName: "Rohit"}

// Dot notation (same for value and pointer)
fmt.Println(u.FirstName)    // "Rohit"
u.FirstName = "Raj"         // modify

// Pointer to struct — dot notation works the same!
p := &u
fmt.Println(p.FirstName)    // "Raj" — auto-dereferenced (no -> needed!)
p.FirstName = "Ram"         // auto-dereferenced modification
fmt.Println(u.FirstName)    // "Ram" — original u modified

// Go translates p.FirstName to (*p).FirstName automatically
```

### Struct Comparison

```go
// Structs are comparable if ALL their fields are comparable
type Point struct{ X, Y int }
type Circle struct { Center Point; Radius float64 }

p1 := Point{1, 2}
p2 := Point{1, 2}
p3 := Point{3, 4}

fmt.Println(p1 == p2) // true
fmt.Println(p1 == p3) // false

// Can use as map key since all fields comparable
locations := map[Point]string{
    {0, 0}: "origin",
    {1, 0}: "east",
}

// Struct with slice CANNOT be compared with ==
type Container struct {
    Items []int  // slices are not comparable
}
// c1 == c2 // COMPILE ERROR — cannot compare struct with slice field
// Use reflect.DeepEqual for deep comparison:
import "reflect"
c1 := Container{Items: []int{1, 2, 3}}
c2 := Container{Items: []int{1, 2, 3}}
fmt.Println(reflect.DeepEqual(c1, c2)) // true
```

### Struct Tags — The Metadata System

```go
type Product struct {
    // Tags are read at RUNTIME via reflection
    // Multiple tags separated by spaces
    ID          int     `json:"id" db:"product_id" validate:"required,min=1"`
    Name        string  `json:"name" db:"name" validate:"required,max=100"`
    Price       float64 `json:"price" db:"price_cents" validate:"min=0"`
    Description string  `json:"description,omitempty" db:"description"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    Internal    string  `json:"-"`  // omit from JSON entirely
}

// Tag format: `key:"value" key2:"value2"`
// Key: tool that uses it (json, db, validate, yaml, etc.)
// Value: tool-specific options

// Tags accessed via reflection:
func printTags(v interface{}) {
    t := reflect.TypeOf(v)
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        jsonTag := field.Tag.Get("json")
        fmt.Printf("Field: %s, JSON: %s\n", field.Name, jsonTag)
    }
}
```

---

## Chapter 3: Methods — Comprehensive Guide

### What Are Methods?

A method is a function with a special **receiver** argument that appears before the function name. Methods are associated with a type.

```go
// Free function:
func Greet(name string) string { return "Hello, " + name }

// Method on User type:
func (u User) Greet() string { return "Hello, I'm " + u.FirstName }

// The receiver is just syntactic sugar — these are equivalent:
u.Greet()          // method call
User.Greet(u)      // function call with explicit receiver (rarely used)
```

### Value Receivers vs Pointer Receivers

```go
type Rectangle struct {
    Width, Height float64
}

// VALUE RECEIVER — operates on a COPY of rect
// - Can be called on both value and pointer
// - Does not modify the original
// - Thread-safe to call concurrently (each call gets its own copy)
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func (r Rectangle) String() string {
    return fmt.Sprintf("%.2fx%.2f", r.Width, r.Height)
}

// POINTER RECEIVER — operates on THE ORIGINAL via its address
// - Can be called on both value and pointer (Go auto-takes address)
// - CAN modify the original
// - NOT safe to copy if there's a mutex inside
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}

func main() {
    r := Rectangle{Width: 10, Height: 5}
    
    // Calling value receiver methods
    fmt.Println(r.Area())     // 50
    fmt.Println(r.String())   // "10.00x5.00"
    
    // Calling pointer receiver methods
    r.Scale(2)                // Go converts r.Scale(2) to (&r).Scale(2)
    fmt.Println(r.Area())     // 200
    
    // Via pointer
    p := &r
    p.Scale(0.5)              // works — p is already a pointer
    fmt.Println(p.Area())     // 100 — also works with pointer receiver via value call
}
```

### Method Sets — Critical for Interfaces

```go
// METHOD SET determines which interfaces a type satisfies:

// Value type T's method set:
// - All methods with VALUE receiver (T)

// Pointer type *T's method set:
// - All methods with VALUE receiver (T)
// - All methods with POINTER receiver (*T)

// Consequence:
type Mover interface { Move() }

type Car struct{}
func (c Car) Move() {}    // value receiver

type Boat struct{}
func (b *Boat) Move() {}  // pointer receiver

// Car (value) satisfies Mover — because Move is a value method
var m1 Mover = Car{}      // OK
var m2 Mover = &Car{}     // OK

// Boat (pointer) satisfies Mover — pointer's method set includes pointer receivers
var m3 Mover = &Boat{}    // OK
// var m4 Mover = Boat{}  // COMPILE ERROR: Boat doesn't implement Mover
                         // (only *Boat has Move in its method set)

// This is why: if ANY method has a pointer receiver → always use *T to satisfy interfaces
```

### Methods on Non-Struct Types

```go
// You can define methods on ANY named type, not just structs!

type StringSlice []string

func (s StringSlice) Contains(target string) bool {
    for _, str := range s {
        if str == target {
            return true
        }
    }
    return false
}

func (s StringSlice) Join(sep string) string {
    return strings.Join(s, sep)
}

// Usage:
words := StringSlice{"hello", "world", "go"}
fmt.Println(words.Contains("go"))    // true
fmt.Println(words.Join(", "))        // "hello, world, go"

// Methods on int type
type Celsius float64
type Fahrenheit float64

func (c Celsius) ToFahrenheit() Fahrenheit {
    return Fahrenheit(c*9/5 + 32)
}

func (f Fahrenheit) ToCelsius() Celsius {
    return Celsius((f - 32) * 5 / 9)
}

// Methods on function types!
type HandlerFunc func(http.ResponseWriter, *http.Request)

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    f(w, r) // call the function
}
// This is exactly how net/http.HandlerFunc works in the stdlib!
```

### Promoted Methods — The Embedding Shortcut

```go
type Logger struct{}
func (l Logger) Log(msg string) { fmt.Println("[LOG]", msg) }

type Auditor struct{}
func (a Auditor) Audit(action string) { fmt.Println("[AUDIT]", action) }

type Service struct {
    Logger              // embedded — all Logger methods promoted
    Auditor             // embedded — all Auditor methods promoted
    Name string
}

s := Service{Name: "PaymentService"}

// Call promoted methods directly:
s.Log("starting...")      // same as s.Logger.Log("starting...")
s.Audit("payment_init")   // same as s.Auditor.Audit("payment_init")

// Can also access explicitly:
s.Logger.Log("explicit call")
```

---

## Chapter 4: Struct Embedding — Composition Deep Dive

### Embedding is NOT Inheritance!

The key difference:
- **Inheritance**: `class B extends A` — B IS-A A, B's method can call A's methods
- **Embedding**: `type B struct { A }` — B HAS-A A, B's methods are promoted but B is not A

```go
type Animal struct {
    Name string
}

func (a Animal) Breathe() string { return a.Name + " breathes" }
func (a Animal) Eat() string     { return a.Name + " eats" }

type Dog struct {
    Animal       // Dog HAS-A Animal (not IS-A)
    Breed string
}

func (d Dog) Bark() string { return d.Name + " barks!" } // can access Animal.Name directly

d := Dog{
    Animal: Animal{Name: "Rex"},
    Breed:  "German Shepherd",
}

// Promoted methods:
fmt.Println(d.Breathe()) // "Rex breathes"
fmt.Println(d.Eat())     // "Rex eats"
fmt.Println(d.Bark())    // "Rex barks!"
fmt.Println(d.Name)      // "Rex" — promoted field too!

// Dog is NOT an Animal in Go's type system:
func feed(a Animal) {} 
// feed(d)   // COMPILE ERROR — Dog is not Animal
feed(d.Animal)  // Correct — explicitly pass the embedded Animal
```

### Method Overriding (Shadowing)

```go
type Base struct{}
func (b Base) Process() string { return "base process" }
func (b Base) Helper() string  { return "base helper" }

type Child struct {
    Base
}

// Shadow/override Base's Process
func (c Child) Process() string {
    base := c.Base.Process()       // can still call original!
    return "child process (was: " + base + ")"
}
// Helper is NOT overridden — still comes from Base

c := Child{}
fmt.Println(c.Process())   // "child process (was: base process)"
fmt.Println(c.Helper())    // "base helper"
fmt.Println(c.Base.Process()) // "base process" — explicit original call
```

### Interface Satisfaction via Embedding

```go
// If embedded type satisfies an interface, the outer type satisfies it too

type io.Reader interface {
    Read(p []byte) (n int, err error)
}

type ProgressReader struct {
    io.Reader           // embed io.Reader — satisfies io.Reader interface
    BytesRead int64
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
    n, err = pr.Reader.Read(p)  // delegate to embedded reader
    pr.BytesRead += int64(n)    // add tracking
    return
}

// Creating:
file, _ := os.Open("data.txt")
progress := &ProgressReader{Reader: file}
io.Copy(os.Stdout, progress)  // works — *ProgressReader implements io.Reader
fmt.Println("Read:", progress.BytesRead, "bytes")
```

### Embedding Pointer vs Value

```go
type Logger struct{ prefix string }
func (l *Logger) Log(msg string) { fmt.Println(l.prefix+":", msg) }

// Embedding pointer (*Logger)
type Server struct {
    *Logger           // embed POINTER to Logger
    Port int
}

s := Server{
    Logger: &Logger{prefix: "SERVER"},
    Port:   8080,
}
s.Log("started")  // works even though Log has pointer receiver

// When to embed pointer:
// - The embedded type must be initialized (can't be zero-value)
// - You want to share the Logger between multiple structs
// - The embedded type has pointer receivers

// Embedding value (Logger)
type Client struct {
    Logger            // embed VALUE
    URL string
}

c := Client{
    Logger: Logger{prefix: "CLIENT"},
    URL:    "http://api.example.com",
}
c.Log("connecting") // OK — auto-takes address for pointer receiver
```

---

## Chapter 5: Constructor Patterns Deep Dive

### Pattern 1: Simple Constructor

```go
type User struct {
    id        int    // private (lowercase)
    name      string
    email     string
    createdAt time.Time
}

func NewUser(name, email string) *User {
    return &User{
        id:        generateID(),
        name:      name,
        email:     email,
        createdAt: time.Now(),
    }
}
```

### Pattern 2: Constructor with Validation

```go
var (
    ErrEmptyName  = errors.New("name cannot be empty")
    ErrInvalidEmail = errors.New("invalid email format")
)

func NewUser(name, email string) (*User, error) {
    if name == "" {
        return nil, ErrEmptyName
    }
    if !isValidEmail(email) {
        return nil, ErrInvalidEmail
    }
    return &User{name: name, email: email, createdAt: time.Now()}, nil
}
```

### Pattern 3: Builder Pattern

```go
type UserBuilder struct {
    user User
    errs []error
}

func NewUserBuilder(id int) *UserBuilder {
    return &UserBuilder{user: User{id: id}}
}

func (b *UserBuilder) WithName(name string) *UserBuilder {
    if name == "" {
        b.errs = append(b.errs, errors.New("name cannot be empty"))
    }
    b.user.name = name
    return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
    if !isValidEmail(email) {
        b.errs = append(b.errs, fmt.Errorf("invalid email: %s", email))
    }
    b.user.email = email
    return b
}

func (b *UserBuilder) Build() (*User, error) {
    if len(b.errs) > 0 {
        return nil, errors.Join(b.errs...)
    }
    return &b.user, nil
}

// Usage:
user, err := NewUserBuilder(1).
    WithName("Rohit").
    WithEmail("rohit@example.com").
    Build()
```

### Pattern 4: Functional Options (Most Idiomatic)

```go
type ServerConfig struct {
    host         string
    port         int
    readTimeout  time.Duration
    writeTimeout time.Duration
    maxConns     int
    tls          bool
    certFile     string
}

type Option func(*ServerConfig)

func WithHost(host string) Option {
    return func(c *ServerConfig) { c.host = host }
}

func WithPort(port int) Option {
    return func(c *ServerConfig) {
        if port < 1 || port > 65535 {
            panic(fmt.Sprintf("invalid port: %d", port))
        }
        c.port = port
    }
}

func WithTLS(certFile, keyFile string) Option {
    return func(c *ServerConfig) {
        c.tls = true
        c.certFile = certFile
    }
}

func WithTimeout(read, write time.Duration) Option {
    return func(c *ServerConfig) {
        c.readTimeout = read
        c.writeTimeout = write
    }
}

func NewServer(opts ...Option) *ServerConfig {
    // Sensible defaults
    cfg := &ServerConfig{
        host:         "0.0.0.0",
        port:         8080,
        readTimeout:  30 * time.Second,
        writeTimeout: 30 * time.Second,
        maxConns:     1000,
    }
    for _, opt := range opts {
        opt(cfg)
    }
    return cfg
}

// Clean, extensible usage:
s := NewServer(
    WithHost("localhost"),
    WithPort(9090),
    WithTLS("cert.pem", "key.pem"),
    WithTimeout(10*time.Second, 30*time.Second),
)
```

---

## Chapter 6: Comprehensive Testing

```go
package structs_test

import (
    "testing"
    "time"
    "reflect"
)

type Rectangle struct {
    Width, Height float64
}

func (r Rectangle) Area() float64    { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}

func TestRectangleArea(t *testing.T) {
    tests := []struct {
        name   string
        r      Rectangle
        want   float64
    }{
        {"normal", Rectangle{10, 5}, 50},
        {"square", Rectangle{4, 4}, 16},
        {"zero width", Rectangle{0, 5}, 0},
        {"zero height", Rectangle{5, 0}, 0},
        {"both zero", Rectangle{0, 0}, 0},
        {"small floats", Rectangle{0.5, 0.5}, 0.25},
        {"large values", Rectangle{1e6, 1e6}, 1e12},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.r.Area()
            if got != tt.want {
                t.Errorf("Area() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestRectangleScale(t *testing.T) {
    tests := []struct {
        name     string
        r        Rectangle
        factor   float64
        wantW    float64
        wantH    float64
    }{
        {"scale up", Rectangle{10, 5}, 2, 20, 10},
        {"scale down", Rectangle{10, 5}, 0.5, 5, 2.5},
        {"scale by 1", Rectangle{10, 5}, 1, 10, 5},     // no change
        {"scale by 0", Rectangle{10, 5}, 0, 0, 0},      // collapses
        {"scale by -1", Rectangle{10, 5}, -1, -10, -5}, // negative
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            r := tt.r  // copy to avoid mutation between tests
            r.Scale(tt.factor)
            if r.Width != tt.wantW || r.Height != tt.wantH {
                t.Errorf("After Scale(%v): got {%v, %v}, want {%v, %v}",
                    tt.factor, r.Width, r.Height, tt.wantW, tt.wantH)
            }
        })
    }
}

// Test that value receiver doesn't modify original
func TestValueReceiverNoMutation(t *testing.T) {
    r := Rectangle{Width: 10, Height: 5}
    original := r  // save copy
    
    _ = r.Area()        // value receiver — should not modify r
    _ = r.Perimeter()   // value receiver — should not modify r
    
    if r != original {
        t.Error("value receiver method should not modify the receiver")
    }
}

// Test embedding
type Animal struct{ Name string }
func (a Animal) Speak() string { return a.Name + " speaks" }

type Dog struct {
    Animal
    Breed string
}
func (d Dog) Speak() string { return d.Name + " barks" } // override

func TestEmbeddingAndOverride(t *testing.T) {
    d := Dog{
        Animal: Animal{Name: "Rex"},
        Breed:  "Labrador",
    }
    
    // Overridden method
    if d.Speak() != "Rex barks" {
        t.Errorf("Dog.Speak() = %q, want 'Rex barks'", d.Speak())
    }
    
    // Original method still accessible
    if d.Animal.Speak() != "Rex speaks" {
        t.Errorf("Animal.Speak() = %q, want 'Rex speaks'", d.Animal.Speak())
    }
    
    // Promoted field
    if d.Name != "Rex" {
        t.Errorf("d.Name = %q, want 'Rex'", d.Name)
    }
}

// Test functional options pattern
type Server struct{ host string; port int }
type SrvOption func(*Server)

func WithHost(host string) SrvOption { return func(s *Server) { s.host = host } }
func WithPort(port int) SrvOption    { return func(s *Server) { s.port = port } }
func NewServer(opts ...SrvOption) *Server {
    s := &Server{host: "localhost", port: 8080}
    for _, o := range opts { o(s) }
    return s
}

func TestFunctionalOptions(t *testing.T) {
    // Default values
    s1 := NewServer()
    if s1.host != "localhost" || s1.port != 8080 {
        t.Errorf("defaults: got {%s, %d}", s1.host, s1.port)
    }
    
    // With custom options
    s2 := NewServer(WithHost("example.com"), WithPort(9090))
    if s2.host != "example.com" || s2.port != 9090 {
        t.Errorf("custom opts: got {%s, %d}", s2.host, s2.port)
    }
    
    // Partial options
    s3 := NewServer(WithPort(3000))
    if s3.host != "localhost" || s3.port != 3000 {
        t.Errorf("partial opts: got {%s, %d}", s3.host, s3.port)
    }
    
    // Options are independent
    if reflect.DeepEqual(s1, s2) {
        t.Error("different opts should produce different servers")
    }
}

// Test struct comparison
func TestStructComparison(t *testing.T) {
    type Point struct{ X, Y int }
    
    p1 := Point{1, 2}
    p2 := Point{1, 2}
    p3 := Point{3, 4}
    
    if p1 != p2 { t.Error("identical structs should be equal") }
    if p1 == p3 { t.Error("different structs should not be equal") }
    
    // As map keys
    m := map[Point]string{p1: "origin-ish", p3: "far"}
    if m[p2] != "origin-ish" {
        t.Error("struct as map key: identical structs should look up same entry")
    }
}

// Benchmarks
func BenchmarkValueReceiver(b *testing.B) {
    r := Rectangle{Width: 10, Height: 5}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = r.Area()
    }
}

func BenchmarkPointerReceiver(b *testing.B) {
    r := &Rectangle{Width: 10, Height: 5}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        r.Scale(1.0) // no-op scale
    }
}
```

---

**Summary of Part 7:**
- Go has no classes — structs + methods + interfaces replace everything
- Embedding is composition (HAS-A), NOT inheritance (IS-A)
- Value receiver = copy (thread-safe, non-modifying)
- Pointer receiver = original (modifying, consistent for types with mutexes)
- If ANY method has pointer receiver → use pointer everywhere for that type
- *T's method set includes all value and pointer methods; T's set only includes value methods
- This determines which interfaces a type satisfies
- Functional options is the most idiomatic Go constructor pattern
- Test both value and pointer receiver behaviors explicitly
