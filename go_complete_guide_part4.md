# Complete Go Programming Guide — Part 4: Advanced Topics & C++ Migration

---

# Chapter 18: Advanced Patterns

## Embedding Interfaces in Structs

```go
// You can embed interfaces inside structs — useful for partial implementation
type Animal interface {
    Speak() string
    Move() string
}

// BaseAnimal provides default implementations via an embedded interface
type BaseAnimal struct {
    Animal // embed the interface
    Name   string
}

// Override specific methods
func (b BaseAnimal) Speak() string {
    return b.Name + " makes a generic sound"
}

// Must still provide Move() or it'll panic at runtime if called
```

## Enum Pattern (Go has no enums)

```go
// Use typed constants with iota
type Status int

const (
    StatusPending  Status = iota // 0
    StatusActive                 // 1
    StatusInactive               // 2
    StatusDeleted                // 3
)

// String representation
func (s Status) String() string {
    switch s {
    case StatusPending:
        return "PENDING"
    case StatusActive:
        return "ACTIVE"
    case StatusInactive:
        return "INACTIVE"
    case StatusDeleted:
        return "DELETED"
    default:
        return "UNKNOWN"
    }
}

// Validation
func (s Status) IsValid() bool {
    return s >= StatusPending && s <= StatusDeleted
}

func main() {
    s := StatusActive
    fmt.Println(s) // "ACTIVE" (calls String() method)
}
```

## The Stringer Interface & Custom Formatting

```go
type Point struct {
    X, Y float64
}

// Implement fmt.Stringer
func (p Point) String() string {
    return fmt.Sprintf("(%0.2f, %0.2f)", p.X, p.Y)
}

// Implement fmt.GoStringer (for %#v format)
func (p Point) GoString() string {
    return fmt.Sprintf("Point{X: %f, Y: %f}", p.X, p.Y)
}

func main() {
    p := Point{3.14, 2.71}
    fmt.Println(p)        // (3.14, 2.71)
    fmt.Printf("%v\n", p) // (3.14, 2.71)
    fmt.Printf("%#v\n", p) // Point{X: 3.140000, Y: 2.710000}
}
```

## Builder Pattern

```go
type QueryBuilder struct {
    table      string
    conditions []string
    orderBy    string
    limit      int
}

func NewQueryBuilder(table string) *QueryBuilder {
    return &QueryBuilder{table: table}
}

func (qb *QueryBuilder) Where(condition string) *QueryBuilder {
    qb.conditions = append(qb.conditions, condition)
    return qb // return self for chaining
}

func (qb *QueryBuilder) OrderBy(field string) *QueryBuilder {
    qb.orderBy = field
    return qb
}

func (qb *QueryBuilder) Limit(n int) *QueryBuilder {
    qb.limit = n
    return qb
}

func (qb *QueryBuilder) Build() string {
    query := "SELECT * FROM " + qb.table
    if len(qb.conditions) > 0 {
        query += " WHERE " + strings.Join(qb.conditions, " AND ")
    }
    if qb.orderBy != "" {
        query += " ORDER BY " + qb.orderBy
    }
    if qb.limit > 0 {
        query += fmt.Sprintf(" LIMIT %d", qb.limit)
    }
    return query
}

// Usage
query := NewQueryBuilder("users").
    Where("age > 18").
    Where("status = 'active'").
    OrderBy("name").
    Limit(10).
    Build()
```

## Strategy Pattern with Interfaces

```go
// Define strategy interface
type SortStrategy interface {
    Sort(data []int) []int
}

// Concrete strategies
type BubbleSort struct{}
func (b BubbleSort) Sort(data []int) []int { /* ... */ return data }

type QuickSort struct{}
func (q QuickSort) Sort(data []int) []int { /* ... */ return data }

type MergeSort struct{}
func (m MergeSort) Sort(data []int) []int { /* ... */ return data }

// Context
type Sorter struct {
    strategy SortStrategy
}

func (s *Sorter) SetStrategy(strategy SortStrategy) {
    s.strategy = strategy
}

func (s *Sorter) Sort(data []int) []int {
    return s.strategy.Sort(data)
}

// Usage
sorter := &Sorter{}
sorter.SetStrategy(QuickSort{})
result := sorter.Sort([]int{5, 3, 1, 4, 2})
```

## Middleware Pattern (HTTP)

```go
type Middleware func(http.Handler) http.Handler

// Logging middleware
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        log.Printf("Started %s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
        log.Printf("Completed in %v", time.Since(start))
    })
}

// Auth middleware
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// Chain middlewares
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        handler = middlewares[i](handler)
    }
    return handler
}

// Usage
mux := http.NewServeMux()
mux.Handle("/api/data", Chain(
    http.HandlerFunc(handleData),
    LoggingMiddleware,
    AuthMiddleware,
))
```

---

# Chapter 19: Reflection (Advanced)

```go
package main

import (
    "fmt"
    "reflect"
)

type User struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"email"`
    Age   int    `json:"age" validate:"min=0,max=150"`
}

func inspectStruct(v interface{}) {
    val := reflect.ValueOf(v)
    typ := reflect.TypeOf(v)

    // Handle pointer
    if typ.Kind() == reflect.Ptr {
        val = val.Elem()
        typ = typ.Elem()
    }

    fmt.Printf("Type: %s\n", typ.Name())
    fmt.Printf("Kind: %s\n", typ.Kind())
    fmt.Printf("Fields:\n")

    for i := 0; i < typ.NumField(); i++ {
        field := typ.Field(i)
        value := val.Field(i)
        tag := field.Tag.Get("json")
        fmt.Printf("  %s (%s) = %v [json:%s]\n",
            field.Name, field.Type, value.Interface(), tag)
    }
}

func main() {
    u := User{Name: "Rohit", Email: "rohit@test.com", Age: 30}
    inspectStruct(u)
    // Output:
    // Type: User
    // Kind: struct
    // Fields:
    //   Name (string) = Rohit [json:name]
    //   Email (string) = rohit@test.com [json:email]
    //   Age (int) = 30 [json:age]
}

// When to use reflection:
// - JSON/XML marshaling (encoding/json uses it)
// - ORMs and database mapping
// - Validation frameworks
// - Plugin systems
// When NOT to use:
// - Regular business logic (it's slow and not type-safe)
// - Avoid if generics can solve the problem
```

---

# Chapter 20: Go Tools You Must Know

```bash
# Format code (enforced style — no debates!)
go fmt ./...
# Or the stricter version:
gofmt -w .

# Vet — find suspicious code patterns
go vet ./...

# Lint — more comprehensive analysis
# Install: go install golang.org/x/lint/golint@latest
golint ./...

# Modern alternative: staticcheck
# Install: go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...

# Race detector — finds data races at runtime
go run -race main.go
go test -race ./...

# Profiling
go test -cpuprofile cpu.prof -memprofile mem.prof -bench .
go tool pprof cpu.prof

# Documentation
go doc fmt.Println
go doc -all net/http

# Generate (code generation)
# Put this in your code:
//go:generate stringer -type=Status
go generate ./...

# Build for different platforms (cross-compilation!)
GOOS=linux GOARCH=amd64 go build -o myapp-linux
GOOS=darwin GOARCH=arm64 go build -o myapp-mac
GOOS=windows GOARCH=amd64 go build -o myapp.exe
```

---

# Chapter 21: C++ to Go Translation Guide

This is the most critical chapter for your interview. Here are side-by-side translations of common C++ patterns.

## Classes → Structs + Methods

```cpp
// C++
class User {
private:
    std::string name;
    int age;
public:
    User(const std::string& name, int age) : name(name), age(age) {}
    std::string getName() const { return name; }
    void setAge(int newAge) { age = newAge; }
    virtual std::string describe() const {
        return name + " is " + std::to_string(age);
    }
    virtual ~User() = default;
};
```

```go
// Go
type User struct {
    name string // lowercase = private to package
    age  int
}

func NewUser(name string, age int) *User {
    return &User{name: name, age: age}
}

func (u *User) GetName() string { return u.name }
func (u *User) SetAge(newAge int) { u.age = newAge }
func (u *User) Describe() string {
    return fmt.Sprintf("%s is %d", u.name, u.age)
}
// No destructor needed — garbage collected
```

## Inheritance → Composition + Interfaces

```cpp
// C++
class Admin : public User {
    std::string role;
public:
    Admin(const std::string& n, int a, const std::string& r) 
        : User(n, a), role(r) {}
    std::string describe() const override {
        return User::describe() + " [" + role + "]";
    }
};
```

```go
// Go — Composition
type Admin struct {
    User           // embed User
    role string
}

func NewAdmin(name string, age int, role string) *Admin {
    return &Admin{
        User: User{name: name, age: age},
        role: role,
    }
}

// Override Describe
func (a *Admin) Describe() string {
    return a.User.Describe() + " [" + a.role + "]"
}

// Use interface for polymorphism
type Describable interface {
    Describe() string
}

func PrintDescription(d Describable) {
    fmt.Println(d.Describe())
}
```

## Templates → Generics

```cpp
// C++
template<typename T>
T max(T a, T b) {
    return (a > b) ? a : b;
}
```

```go
// Go
type Ordered interface {
    ~int | ~float64 | ~string
}

func Max[T Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}
```

## std::thread → Goroutines

```cpp
// C++
#include <thread>
#include <mutex>
#include <condition_variable>
#include <queue>

class ThreadSafeQueue {
    std::queue<int> queue;
    std::mutex mtx;
    std::condition_variable cv;
public:
    void push(int val) {
        std::lock_guard<std::mutex> lock(mtx);
        queue.push(val);
        cv.notify_one();
    }
    int pop() {
        std::unique_lock<std::mutex> lock(mtx);
        cv.wait(lock, [this]{ return !queue.empty(); });
        int val = queue.front();
        queue.pop();
        return val;
    }
};

// Usage
ThreadSafeQueue q;
std::thread producer([&q]{
    for (int i = 0; i < 10; i++) q.push(i);
});
std::thread consumer([&q]{
    for (int i = 0; i < 10; i++) std::cout << q.pop() << "\n";
});
producer.join();
consumer.join();
```

```go
// Go — The channel IS the thread-safe queue!
func main() {
    queue := make(chan int, 10) // buffered channel = thread-safe queue

    // Producer
    go func() {
        for i := 0; i < 10; i++ {
            queue <- i // blocks if buffer is full (like cv.wait!)
        }
        close(queue)
    }()

    // Consumer
    for val := range queue { // blocks if empty, exits when closed
        fmt.Println(val)
    }
}
// That's it. No mutex, no condition_variable, no lock_guard.
// The channel handles ALL of it.
```

## RAII → defer

```cpp
// C++
class FileHandler {
    std::ofstream file;
public:
    FileHandler(const std::string& path) : file(path) {}
    ~FileHandler() { file.close(); } // RAII cleanup
    void write(const std::string& data) { file << data; }
};

{
    FileHandler fh("output.txt"); // opens file
    fh.write("hello");
} // destructor called automatically — file closed
```

```go
// Go
func writeToFile() error {
    file, err := os.Create("output.txt")
    if err != nil {
        return err
    }
    defer file.Close() // cleanup when function returns

    _, err = file.WriteString("hello")
    return err
}
```

## std::unique_ptr / std::shared_ptr → Just Use Values or Pointers

```cpp
// C++
auto user = std::make_unique<User>("Rohit", 30);
auto shared = std::make_shared<User>("Alice", 25);
```

```go
// Go — no smart pointers needed, garbage collector handles it
user := &User{name: "Rohit", age: 30}  // pointer
shared := &User{name: "Alice", age: 25} // also pointer — GC tracks references
// When no references exist, GC frees the memory. Done.
```

## Exception Handling → Error Returns

```cpp
// C++
try {
    auto result = riskyOperation();
    process(result);
} catch (const std::runtime_error& e) {
    std::cerr << "Error: " << e.what() << std::endl;
} catch (const std::exception& e) {
    std::cerr << "General error: " << e.what() << std::endl;
}
```

```go
// Go
result, err := riskyOperation()
if err != nil {
    // Handle specific error types
    var notFoundErr *NotFoundError
    if errors.As(err, &notFoundErr) {
        fmt.Println("Not found:", notFoundErr.ID)
        return
    }
    // General error
    return fmt.Errorf("riskyOperation failed: %w", err)
}
process(result)
```

## Singleton Pattern

```cpp
// C++ (Meyer's Singleton)
class Database {
    Database() {} // private constructor
public:
    static Database& getInstance() {
        static Database instance;
        return instance;
    }
};
```

```go
// Go (sync.Once)
var (
    dbInstance *Database
    dbOnce    sync.Once
)

func GetDatabase() *Database {
    dbOnce.Do(func() {
        dbInstance = &Database{
            // initialization
        }
    })
    return dbInstance
}
```

## Observer Pattern

```cpp
// C++ summary: Base Observer class, vector of observers, notify loop
```

```go
// Go — Use channels instead of callbacks
type EventBus struct {
    subscribers map[string][]chan string
    mu          sync.RWMutex
}

func NewEventBus() *EventBus {
    return &EventBus{
        subscribers: make(map[string][]chan string),
    }
}

func (eb *EventBus) Subscribe(topic string) <-chan string {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    ch := make(chan string, 10)
    eb.subscribers[topic] = append(eb.subscribers[topic], ch)
    return ch
}

func (eb *EventBus) Publish(topic, message string) {
    eb.mu.RLock()
    defer eb.mu.RUnlock()
    for _, ch := range eb.subscribers[topic] {
        go func(c chan string) {
            c <- message
        }(ch)
    }
}

func main() {
    bus := NewEventBus()

    // Subscriber
    events := bus.Subscribe("user.created")
    go func() {
        for msg := range events {
            fmt.Println("Got event:", msg)
        }
    }()

    // Publisher
    bus.Publish("user.created", "New user: Rohit")
}
```

---

# Chapter 22: Interview Preparation Checklist

## Must-Know Topics
- [ ] Variables, types, zero values, type conversion
- [ ] Slices (internal structure, append behavior, gotchas)
- [ ] Maps (nil map vs empty map, comma-ok pattern)
- [ ] Structs, methods, pointer vs value receivers
- [ ] Interfaces (implicit implementation, empty interface, type assertions)
- [ ] Error handling (wrapping, errors.Is, errors.As)
- [ ] Goroutines and WaitGroup
- [ ] Channels (buffered vs unbuffered, direction, closing)
- [ ] Select statement
- [ ] Context (cancel, timeout, deadline)
- [ ] sync.Mutex, sync.RWMutex, sync.Once
- [ ] Defer (execution order, argument evaluation)
- [ ] Generics (type parameters, constraints)
- [ ] Testing (table-driven tests, benchmarks)

## Common Interview Questions

**Q: What's the difference between a nil slice and an empty slice?**
A: A nil slice has no underlying array (pointer=nil, len=0, cap=0). An empty slice has an allocated array but length 0. Both work with `len()`, `cap()`, `append()`, and `range`. Prefer nil slices.

**Q: What happens if you write to a nil map?**
A: It panics. Always initialize maps with `make()` before writing.

**Q: Explain how interfaces work internally.**
A: An interface value contains two pointers: one to a type descriptor and one to the data. A nil pointer stored in an interface makes the interface non-nil (common bug!).

**Q: What's the difference between buffered and unbuffered channels?**
A: Unbuffered: sender blocks until receiver is ready (synchronous). Buffered: sender only blocks when buffer is full (async up to buffer size).

**Q: When do you use Mutex vs Channel?**
A: Mutex for protecting shared state (cache, counters). Channel for passing ownership of data between goroutines, orchestrating workflows, signaling events.

**Q: What is a goroutine leak and how do you prevent it?**
A: A goroutine that runs forever because it's blocked on a channel or waiting for something that never happens. Prevent with: context cancellation, done channels, timeouts, and ensuring channels are eventually closed.

**Q: Explain defer execution order.**
A: Deferred calls are pushed onto a stack (LIFO). Arguments are evaluated immediately but the call is executed when the enclosing function returns. Multiple defers execute in reverse order.

**Q: How would you convert a C++ class hierarchy to Go?**
A: Replace inheritance with struct embedding (composition). Define interfaces for polymorphic behavior. Use method overriding via shadowing. Replace virtual functions with interface dispatch.

---

# Chapter 23: Quick Reference Card

```go
// Variable declaration
x := 42                         // short declaration
var x int = 42                   // explicit
var x int                        // zero value

// Slice operations
s := make([]int, 0, 10)          // create
s = append(s, 1, 2, 3)           // append
copy(dst, src)                   // copy
s[1:3]                           // sub-slice

// Map operations
m := make(map[string]int)        // create
m["key"] = val                   // set
val, ok := m["key"]              // get + check
delete(m, "key")                 // delete

// Goroutine
go func() { /* ... */ }()        // start

// Channel
ch := make(chan int)              // unbuffered
ch := make(chan int, 10)          // buffered
ch <- val                        // send
val := <-ch                      // receive
close(ch)                        // close
for v := range ch { }            // iterate

// Select
select {
case v := <-ch1:
case ch2 <- val:
case <-time.After(1*time.Second):
default:
}

// Error handling
if err != nil { return err }
errors.Is(err, target)
errors.As(err, &target)
fmt.Errorf("context: %w", err)

// Struct + Method
type T struct { Field int }
func (t *T) Method() { }

// Interface
type I interface { Method() }

// Generics
func F[T any](x T) T { return x }

// Testing
func TestX(t *testing.T) { }
func BenchmarkX(b *testing.B) { }
```

---

**Good luck with your interview, Rohit! 🚀**

Key things to remember:
1. Go is about SIMPLICITY — don't try to write C++ in Go syntax
2. Composition over inheritance, always
3. Channels replace condition variables and queues
4. `defer` replaces RAII
5. Error returns replace exceptions
6. Keep interfaces small (1-3 methods)
7. Accept interfaces, return structs
