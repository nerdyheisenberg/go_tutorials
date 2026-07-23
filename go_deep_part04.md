# Go Deep Dive — Part 4: Functions — Everything You Need to Know

---

## Chapter 1: Functions as First-Class Citizens

In Go, functions are **values**. They can be:
- Assigned to variables
- Passed as arguments
- Returned from other functions
- Stored in data structures

This is the foundation of Go's composability — Go doesn't have generics-based algorithms like C++ STL. Instead, it passes functions.

### Why First-Class Functions Matter

```go
// Without first-class functions (C-style):
// if condition == "ascending"  { sortAscending(data) }
// if condition == "descending" { sortDescending(data) }

// With first-class functions (Go-style):
type SortFn func(a, b int) bool

func sortWith(data []int, less SortFn) {
    for i := 0; i < len(data)-1; i++ {
        for j := 0; j < len(data)-i-1; j++ {
            if !less(data[j], data[j+1]) {
                data[j], data[j+1] = data[j+1], data[j]
            }
        }
    }
}

func main() {
    data := []int{5, 2, 8, 1, 9, 3}
    
    sortWith(data, func(a, b int) bool { return a < b }) // ascending
    fmt.Println(data) // [1 2 3 5 8 9]
    
    sortWith(data, func(a, b int) bool { return a > b }) // descending
    fmt.Println(data) // [9 8 5 3 2 1]
}
```

---

## Chapter 2: Function Signatures — All Forms

### Basic Functions

```go
// No parameters, no return
func sayHello() {
    fmt.Println("Hello!")
}

// Parameters, no return
func greet(name string) {
    fmt.Printf("Hello, %s!\n", name)
}

// Shortened parameter list (same type)
func add(a, b int) int {            // equivalent to: a int, b int
    return a + b
}

func processData(x, y, z float64) float64 {
    return (x + y + z) / 3
}

// Mixed types (can't shorten)
func configure(host string, port int, debug bool) {
    ...
}
```

### Multiple Return Values — The Go Pattern

```go
// Go's standard: return (result, error)
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// Multiple non-error returns
func minMax(nums []int) (min, max int) {  // named returns
    if len(nums) == 0 {
        return 0, 0
    }
    min, max = nums[0], nums[0]
    for _, n := range nums[1:] {
        if n < min { min = n }
        if n > max { max = n }
    }
    return // naked return — returns named values
}

// Three returns (less common)
func parseAddress(addr string) (host string, port int, err error) {
    parts := strings.Split(addr, ":")
    if len(parts) != 2 {
        err = fmt.Errorf("invalid address: %s", addr)
        return // naked return: host="", port=0, err is set
    }
    host = parts[0]
    port, err = strconv.Atoi(parts[1])
    return // naked return with all named values
}
```

**Named return values — When to use:**

```go
// Good: Short functions where naming aids documentation
func bounds(data []float64) (min, max float64) {
    // the names "min" and "max" are self-documenting
    ...
}

// Bad: Long complex functions (naked returns make code hard to understand)
func complexOperation(...) (result int, data []byte, err error) {
    // ... 100 lines ...
    return // which values are returned? must scroll up to find out!
}

// Better for long functions: explicit returns
func complexOperation(...) (int, []byte, error) {
    // ... 100 lines ...
    return result, data, err // you can see all values at return site
}
```

---

## Chapter 3: Variadic Functions

```go
// ... before the type makes the last parameter variadic
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

// Calling variadic functions
sum()               // 0 — zero arguments is valid!
sum(1)              // 1
sum(1, 2, 3, 4, 5)  // 15

// Spreading a slice
nums := []int{1, 2, 3, 4, 5}
sum(nums...)        // spread the slice — the ... unpacks it

// Common example: fmt.Printf
fmt.Printf("Hello, %s! You are %d years old.\n", "Rohit", 30)
// func Printf(format string, a ...any) (n int, err error)

// Why variadic works with slices internally:
func printAll(args ...string) {
    // args is a []string inside the function
    for i, arg := range args {
        fmt.Printf("[%d] %s\n", i, arg)
    }
}
```

### Variadic vs Slice Parameter — Design Choice

```go
// Variadic: convenient for literal calls
func joinStrings(sep string, parts ...string) string {
    return strings.Join(parts, sep)
}
joinStrings(", ", "Alice", "Bob", "Charlie") // natural to call

// Slice: better when caller already has a collection
func processItems(items []string) {
    // ...
}
existing := []string{"Alice", "Bob"}
processItems(existing) // clean — no ... needed

// Rule: use variadic when you expect literal calls
// Use slice when you expect to receive an existing collection
```

---

## Chapter 4: Closures — Deep Dive

A **closure** is a function that captures variables from its surrounding scope. The captured variables are shared between the closure and the enclosing function.

### How Closures Work Internally

```go
func makeCounter() func() int {
    count := 0  // count lives on the HEAP (escapes to heap via closure)
    return func() int {
        count++  // this closure CLOSES OVER count
        return count
    }
}

// What happens in memory:
// - count is allocated on the heap (not stack!)
// - The returned function holds a reference to count
// - Multiple closures can share the SAME count
```

### Closure Patterns

```go
// Pattern 1: Stateful function
func makeCounter() func() int {
    count := 0
    return func() int { count++; return count }
}

c1 := makeCounter()
c2 := makeCounter() // SEPARATE counter, separate closure

c1() // 1
c1() // 2
c2() // 1 — independent counter
c1() // 3

// Pattern 2: Middleware / Decorator
func withLogging(fn func(int) int, name string) func(int) int {
    return func(x int) int {
        fmt.Printf("Calling %s(%d)\n", name, x)
        result := fn(x)
        fmt.Printf("%s returned %d\n", name, result)
        return result
    }
}

doubler := func(x int) int { return x * 2 }
loggedDoubler := withLogging(doubler, "doubler")
loggedDoubler(5) // logs "Calling doubler(5)" then "doubler returned 10"

// Pattern 3: Memoization
func memoize(fn func(int) int) func(int) int {
    cache := make(map[int]int)
    return func(n int) int {
        if val, ok := cache[n]; ok {
            return val // cache hit
        }
        result := fn(n)
        cache[n] = result
        return result
    }
}

fib := memoize(func(n int) int {
    // simplified — in real code this is recursive
    return n
})

// Pattern 4: Functional options (common Go pattern)
type Server struct {
    host    string
    port    int
    timeout time.Duration
}

type ServerOption func(*Server)

func WithPort(port int) ServerOption {
    return func(s *Server) { s.port = port }  // closure capturing 'port'
}

func WithTimeout(d time.Duration) ServerOption {
    return func(s *Server) { s.timeout = d }
}

func NewServer(host string, opts ...ServerOption) *Server {
    s := &Server{host: host, port: 8080}
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Clean call site:
s := NewServer("localhost", WithPort(9090), WithTimeout(30*time.Second))
```

### The Classic Goroutine Closure Bug

```go
// BUG: All goroutines capture the same variable 'i'
for i := 0; i < 5; i++ {
    go func() {
        fmt.Println(i) // Might print "5 5 5 5 5" — race condition + closure
    }()
}

// FIX 1: Pass as argument (creates a copy)
for i := 0; i < 5; i++ {
    go func(n int) {
        fmt.Println(n) // captures a COPY of i at call time
    }(i)
}

// FIX 2: Create a new variable (old idiom before Go 1.22)
for i := 0; i < 5; i++ {
    i := i // shadows outer i with a new variable
    go func() {
        fmt.Println(i) // captures the inner i
    }()
}

// FIX 3: Go 1.22+ fixes this automatically
// In Go 1.22+, each iteration of a for loop creates a new variable
// so the original buggy code works correctly!
```

---

## Chapter 5: defer — Complete Guide

### What defer Does Under the Hood

`defer` pushes a function call onto a LIFO stack. When the enclosing function returns (for any reason), the deferred calls execute in reverse order.

```
main() called
  ↓
defer A pushed    → stack: [A]
defer B pushed    → stack: [A, B]  
defer C pushed    → stack: [A, B, C]
  ↓
function body completes (or panic, or return)
  ↓
C executes (top of stack)
B executes
A executes (bottom of stack)
  ↓
caller receives return value
```

### Defer Evaluation: WHEN Arguments Are Captured

```go
func example() {
    x := 10
    
    defer fmt.Println("deferred:", x) // x evaluated NOW (captures 10)
    
    x = 20
    fmt.Println("current:", x)
}
// Output:
// current: 20
// deferred: 10  ← NOT 20, because 10 was captured at defer time

// To capture the LATEST value, use a closure:
func example2() {
    x := 10
    defer func() {
        fmt.Println("deferred:", x) // x evaluated WHEN CALLED (at defer time)
    }()
    x = 20
    fmt.Println("current:", x)
}
// Output:
// current: 20
// deferred: 20  ← the latest value, because closure captures by reference
```

### defer and Named Return Values

```go
// Normally, defer runs AFTER return value is set
// But with named returns, defer can MODIFY the return value!

func withDefer() (result int) {
    defer func() {
        result++ // modifies the return value!
    }()
    return 5 // sets result = 5, then defer runs: result becomes 6
}

fmt.Println(withDefer()) // 6, not 5!

// Practical use: wrapping errors in defer
func doWork() (err error) {
    defer func() {
        if err != nil {
            err = fmt.Errorf("doWork: %w", err) // wrap the error
        }
    }()
    
    if err = step1(); err != nil {
        return // err is wrapped by defer
    }
    if err = step2(); err != nil {
        return // err is wrapped by defer
    }
    return
}
```

### defer for Resource Cleanup — The RAII Pattern

```go
// File handling
func processFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close() // guaranteed cleanup

    // ... process file
    return nil
}

// Mutex locking
func (c *Cache) Get(key string) string {
    c.mu.Lock()
    defer c.mu.Unlock() // guaranteed unlock
    return c.data[key]
}

// Database transaction
func (r *Repo) Transfer(from, to int, amount float64) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer func() {
        if err != nil {
            tx.Rollback() // rollback if error
        }
    }()
    
    if err = r.debit(tx, from, amount); err != nil {
        return err
    }
    if err = r.credit(tx, to, amount); err != nil {
        return err
    }
    return tx.Commit()
}

// HTTP response body
func fetchJSON(url string) (map[string]interface{}, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close() // CRITICAL — memory/connection leak without this!
    
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return result, nil
}
```

### defer Performance Considerations

```go
// defer has a small overhead (~20-100ns per call)
// For hot paths called millions of times, this matters

// Hot path without defer (maximum performance):
func hotPath(mu *sync.Mutex, m map[string]int, key string) int {
    mu.Lock()
    val := m[key]
    mu.Unlock() // explicit unlock — faster but riskier (easy to forget)
    return val
}

// Normal path with defer (safer, recommended for most code):
func normalPath(mu *sync.Mutex, m map[string]int, key string) int {
    mu.Lock()
    defer mu.Unlock()
    return m[key]
}

func BenchmarkDefer(b *testing.B) {
    var mu sync.Mutex
    m := map[string]int{"key": 42}
    
    b.Run("with defer", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            normalPath(&mu, m, "key")
        }
    })
    
    b.Run("without defer", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            hotPath(&mu, m, "key")
        }
    })
}
```

---

## Chapter 6: panic and recover

### When to Use panic

```go
// RULE: panic is for programmer errors, NOT runtime errors
// Runtime errors → return error
// Programmer errors (shouldn't happen) → panic

// GOOD use of panic:
func mustParse(s string) int {
    n, err := strconv.Atoi(s)
    if err != nil {
        panic(fmt.Sprintf("mustParse: invalid integer %q: %v", s, err))
        // Caller of mustParse KNOWS the string should be valid
        // If it's invalid, that's a programming error
    }
    return n
}

// In init() — can't continue if this fails
func init() {
    config, err := loadConfig("required-config.json")
    if err != nil {
        panic(fmt.Sprintf("startup failed: cannot load config: %v", err))
    }
    globalConfig = config
}

// BAD use of panic:
func getUser(id int) *User {
    user, err := db.Find(id)
    if err != nil {
        panic(err) // WRONG: this is a normal runtime error, not a programmer error
    }
    return user
}
```

### recover — Catching panics

```go
// recover() only works inside a deferred function
// It returns the value passed to panic(), or nil if no panic

func safeDivide(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic in divide: %v", r)
        }
    }()
    return a / b, nil // panics if b == 0
}

// Common pattern: Server that recovers from handler panics
func recoverMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                // Log the full stack trace
                buf := make([]byte, 64*1024)
                n := runtime.Stack(buf, false)
                log.Printf("PANIC in handler: %v\n%s", err, buf[:n])
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### The Panic-Recover Anti-Pattern

```go
// ANTI-PATTERN: Using panic/recover as a control flow mechanism
// (like exceptions in Java/C++)

// WRONG:
func findUser(id int) *User {
    users := loadUsers()
    for _, u := range users {
        if u.ID == id {
            return u
        }
    }
    panic("user not found") // DON'T DO THIS
}

func main() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Caught:", r) // messy, not idiomatic
        }
    }()
    
    user := findUser(999)
    fmt.Println(user)
}

// RIGHT: Return an error
func findUser(id int) (*User, error) {
    users := loadUsers()
    for _, u := range users {
        if u.ID == id {
            return u, nil
        }
    }
    return nil, fmt.Errorf("user %d not found", id)
}
```

---

## Chapter 7: Function as Method Receiver — Preview

(Full coverage in structs/methods chapter, but here's the key concept)

```go
type Stack struct {
    items []int
}

// Method: function with a receiver
func (s *Stack) Push(item int) {
    s.items = append(s.items, item)
}

// vs free function doing the same
func StackPush(s *Stack, item int) {
    s.items = append(s.items, item)
}

// Both work the same internally
// Method call:     s.Push(1)
// Free function:   StackPush(s, 1)
// Go translates methods → functions internally anyway
```

---

## Chapter 8: init() Function — Deep Dive

```go
package database

import (
    "database/sql"
    _ "github.com/lib/pq" // side-effect import
)

var db *sql.DB

// init() runs automatically, before main(), after package vars are initialized
func init() {
    var err error
    db, err = sql.Open("postgres", "postgres://localhost/mydb?sslmode=disable")
    if err != nil {
        // In init, we panic because the program can't function without a DB
        panic(fmt.Sprintf("cannot open database: %v", err))
    }
    
    if err = db.Ping(); err != nil {
        panic(fmt.Sprintf("cannot connect to database: %v", err))
    }
}

// RULES about init():
// 1. A single file can have multiple init() functions
// 2. They run in the order they appear
// 3. You cannot call init() explicitly
// 4. init() can read package-level variables
// 5. If init() panics, the program crashes with a stack trace
```

---

## Chapter 9: Comprehensive Function Testing

```go
package functions_test

import (
    "errors"
    "fmt"
    "testing"
)

// Testing multiple return values
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

func TestDivide(t *testing.T) {
    tests := []struct {
        name      string
        a, b      float64
        want      float64
        wantErr   bool
        errString string
    }{
        // Positive cases
        {"normal division", 10, 2, 5, false, ""},
        {"fraction result", 10, 3, 3.333, false, ""},
        {"divide by 1", 5, 1, 5, false, ""},
        {"negative dividend", -10, 2, -5, false, ""},
        {"negative divisor", 10, -2, -5, false, ""},
        {"both negative", -10, -2, 5, false, ""},
        {"zero dividend", 0, 5, 0, false, ""},
        
        // Negative cases
        {"divide by zero", 10, 0, 0, true, "division by zero"},
        {"negative divide by zero", -5, 0, 0, true, "division by zero"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := divide(tt.a, tt.b)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("divide(%v, %v) error = %v, wantErr %v", 
                    tt.a, tt.b, err, tt.wantErr)
                return
            }
            
            if tt.wantErr && err.Error() != tt.errString {
                t.Errorf("divide() error = %q, want %q", err.Error(), tt.errString)
            }
            
            if !tt.wantErr {
                const epsilon = 0.001
                diff := got - tt.want
                if diff < 0 { diff = -diff }
                if diff > epsilon {
                    t.Errorf("divide(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
                }
            }
        })
    }
}

// Testing closures
func makeAdder(n int) func(int) int {
    return func(x int) int { return x + n }
}

func TestClosure(t *testing.T) {
    add5 := makeAdder(5)
    add10 := makeAdder(10)
    
    // Each closure is independent
    if got := add5(3); got != 8 {
        t.Errorf("add5(3) = %d, want 8", got)
    }
    if got := add10(3); got != 13 {
        t.Errorf("add10(3) = %d, want 13", got)
    }
    
    // Closure state is persistent
    counter := makeCounter()
    for i := 1; i <= 5; i++ {
        if got := counter(); got != i {
            t.Errorf("counter() = %d, want %d", got, i)
        }
    }
}

func makeCounter() func() int {
    count := 0
    return func() int { count++; return count }
}

// Testing panic/recover
func mustParseInt(s string) int {
    n, err := fmt.Sscanf(s, "%d", new(int))
    if err != nil || n != 1 {
        panic(fmt.Sprintf("invalid integer: %q", s))
    }
    var result int
    fmt.Sscan(s, &result)
    return result
}

func TestPanicRecovery(t *testing.T) {
    // Test that valid input doesn't panic
    defer func() {
        if r := recover(); r != nil {
            t.Errorf("unexpected panic: %v", r)
        }
    }()
    
    if got := mustParseInt("42"); got != 42 {
        t.Errorf("mustParseInt(\"42\") = %d, want 42", got)
    }
}

func TestExpectedPanic(t *testing.T) {
    // Helper to test that a function panics
    assertPanics := func(t *testing.T, name string, fn func()) {
        t.Helper()
        defer func() {
            if recover() == nil {
                t.Errorf("%s: expected panic, but none occurred", name)
            }
        }()
        fn()
    }
    
    assertPanics(t, "invalid string", func() {
        mustParseInt("not-a-number")
    })
    
    assertPanics(t, "empty string", func() {
        mustParseInt("")
    })
}

// Testing variadic functions
func sumAll(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

func TestVariadic(t *testing.T) {
    tests := []struct {
        name  string
        input []int
        want  int
    }{
        {"no args", nil, 0},
        {"single", []int{5}, 5},
        {"multiple", []int{1, 2, 3, 4, 5}, 15},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := sumAll(tt.input...) // spread input for variadic
            if got != tt.want {
                t.Errorf("sumAll(%v) = %d, want %d", tt.input, got, tt.want)
            }
        })
    }
    
    // Also test literal call
    if got := sumAll(1, 2, 3); got != 6 {
        t.Errorf("sumAll(1,2,3) = %d, want 6", got)
    }
}

// Benchmarking functions
func BenchmarkDivide(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _, err := divide(float64(i+1), float64(i+2))
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkClosure(b *testing.B) {
    counter := makeCounter()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        counter()
    }
}
```

---

**Summary of Part 4:**
- Functions are first-class values — store, pass, return them
- Multiple return values are idiomatic — always `(result, error)`
- Named returns aid documentation but avoid naked returns in long functions
- Variadic functions: use `...` for literal calls, slice for existing collections
- Closures capture variables BY REFERENCE — beware the goroutine loop bug
- `defer` runs in LIFO order, captures arguments at defer time (not execution time)
- Use defer for all resource cleanup — locks, files, DB transactions
- `panic` is for programmer errors; `recover` is for frameworks recovering from panics
- Never use panic/recover as a control flow mechanism
