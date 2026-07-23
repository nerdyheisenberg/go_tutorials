# Go Deep Dive — Part 5: Pointers & Memory Model

---

## Chapter 1: Go's Memory Model Philosophy

Go has pointers but deliberately removes the most dangerous pointer operations from C++:

| Feature | C/C++ | Go |
|---------|-------|-----|
| Pointers | ✅ | ✅ |
| Pointer arithmetic (`p++`) | ✅ | ❌ (compile error) |
| Dangling pointers | ✅ (UB) | ❌ (GC prevents) |
| Raw memory access | ✅ (`malloc`) | ❌ (GC manages heap) |
| Null pointer dereference | UB/crash | panic (controlled) |
| Stack vs heap choice | Manual | Automatic (escape analysis) |

The result: Go pointers are safe but have ~90% of the power of C++ pointers.

---

## Chapter 2: Pointer Basics — Complete Reference

### What IS a pointer in Go?

A pointer is a variable that holds the **memory address** of another variable.

```go
package main

import "fmt"

func main() {
    // x is an int with value 42, stored at some memory address
    x := 42
    
    // p is a pointer to int (*int) — it holds x's address
    p := &x          // & is the "address-of" operator
    
    fmt.Printf("x  = %d\n", x)        // 42
    fmt.Printf("&x = %p\n", &x)       // 0xc0000b4000 (hex address)
    fmt.Printf("p  = %p\n", p)        // same address
    fmt.Printf("*p = %d\n", *p)       // 42 (dereferencing)
    
    // Modifying through pointer
    *p = 100         // * is the "dereference" operator (access value at address)
    fmt.Println(x)   // 100 — x was changed through p!
    
    // Pointer to pointer (**int)
    pp := &p
    fmt.Println(**pp) // 100
    
    // Zero value of a pointer is nil
    var q *int
    fmt.Println(q == nil)  // true
    fmt.Println(q)         // <nil>
    
    // CRASH: dereferencing nil pointer
    // fmt.Println(*q)     // panic: runtime error: invalid memory address
}
```

### Creating Pointers — Two Ways

```go
// Way 1: Take address of existing variable
x := 42
p := &x

// Way 2: new() allocates memory and returns pointer
// new(T) allocates a zero-initialized T and returns *T
p2 := new(int)   // *int pointing to int with value 0
*p2 = 42

// Way 3: Address of struct literal (most common for structs)
type User struct { Name string; Age int }
u := &User{Name: "Rohit", Age: 30}  // *User — creates User, returns its address
```

### When nil pointer Panics

```go
type Node struct {
    Value int
    Next  *Node
}

func printList(n *Node) {
    for n != nil {        // ALWAYS check before dereferencing
        fmt.Println(n.Value)
        n = n.Next        // move to next (could be nil)
    }
    // fmt.Println(n.Value) after loop would be panic: n is nil here
}

// Safe pattern: nil check before method call
func (n *Node) String() string {
    if n == nil {          // method with nil receiver — perfectly valid!
        return "<nil>"
    }
    return fmt.Sprintf("%d", n.Value)
}

// Methods on nil pointers are fine as long as you check:
var n *Node
fmt.Println(n.String()) // "<nil>" — no panic!
```

---

## Chapter 3: Pass by Value vs Pass by Pointer

This is one of the most important design decisions in Go code.

### Pass by Value — Function Gets a Copy

```go
type Point struct {
    X, Y float64
}

// Value receiver — gets a COPY of p
func (p Point) Translate(dx, dy float64) Point {
    // We can choose to return a new Point (functional style)
    return Point{p.X + dx, p.Y + dy}
    // p.X += dx; p.Y += dy; return none — would only modify the COPY!
}

func doubleValue(n int) {
    n *= 2  // modifies only the local copy!
}

func main() {
    p := Point{3, 4}
    p2 := p.Translate(1, 1) // p unchanged, p2 is new Point{4, 5}
    fmt.Println(p)  // {3 4}
    fmt.Println(p2) // {4 5}
    
    x := 5
    doubleValue(x)
    fmt.Println(x) // still 5!
}
```

### Pass by Pointer — Function Gets the Original

```go
// Pointer receiver — gets the ACTUAL p (via its address)
func (p *Point) Scale(factor float64) {
    p.X *= factor  // modifies the original
    p.Y *= factor  // note: no -> needed, Go auto-dereferences
}

func doublePointer(n *int) {
    *n *= 2  // dereference and modify
}

func main() {
    p := Point{3, 4}
    p.Scale(2)      // p is now {6, 8} — MODIFIED!
    fmt.Println(p)
    
    x := 5
    doublePointer(&x)
    fmt.Println(x)  // 10!
}
```

### Choosing Value vs Pointer — The Rules

```go
// RULE 1: If the method needs to modify the receiver → pointer
func (c *Counter) Increment() { c.count++ }

// RULE 2: If the struct is large → pointer (avoid copying)
// How large is "large"? Generally > 64 bytes, but use judgment
type LargeStruct struct {
    data [1024]byte  // definitely use pointer
}

// RULE 3: If the type has a mutex or sync primitives → pointer (must not copy)
type SafeMap struct {
    mu   sync.Mutex  // NEVER copy a mutex!
    data map[string]int
}
func (m *SafeMap) Get(k string) int { ... } // must be pointer

// RULE 4: If any method of the type has pointer receiver → ALL methods pointer 
// (for consistency; mixed receivers are confusing)

// RULE 5: Small immutable data → value
type Money struct { Amount int; Currency string }
func (m Money) Add(other Money) Money { /* returns new Money */ }

// ANTI-PATTERN: pointers to basic types (usually unnecessary)
func processValue(n *int) {  // why pass *int when you have int?
    fmt.Println(*n)
}
// Better: pass int directly
func processValue(n int) {
    fmt.Println(n)
}
```

---

## Chapter 4: Escape Analysis — Where Do Variables Live?

Go uses **escape analysis** to decide whether a variable lives on the **stack** or **heap**:

- **Stack**: Fast, automatic cleanup when function returns
- **Heap**: Slower allocation, GC manages cleanup

The programmer doesn't decide — the **compiler** decides based on whether the variable "escapes" the function scope.

```go
// Doesn't escape — stays on stack
func noEscape() {
    x := 42  // x lives on the stack
    fmt.Println(x)
    // x is cleaned up when noEscape returns
}

// Escapes to heap — because we return a pointer to it
func escapes() *int {
    x := 42  // x MUST live on the heap (returned pointer must outlive function)
    return &x  // SAFE in Go! In C++ this would be a dangling pointer!
}

// Escapes because it's used in a closure
func closureEscape() func() int {
    x := 42  // x escapes to heap (closure outlives this function)
    return func() int { return x }
}

// View escape analysis with:
// go build -gcflags="-m" ./...
// Output shows: "x escapes to heap"
```

```bash
# How to check escape analysis:
go build -gcflags="-m -m" ./main.go 2>&1 | head -30

# Output example:
# ./main.go:12:2: x escapes to heap
# ./main.go:8:2: moved to heap: x
# ./main.go:21:9: func literal escapes to heap
```

### Heap vs Stack Performance

```go
// Stack allocation is faster (pointer bump) and GC-free
// Heap allocation is slower (needs synchronization) and GC pressure

// For performance-critical code, try to avoid heap escapes

// BAD: always allocates on heap
func badAlloc() *int {
    n := new(int)  // or: n := 42; return &n
    return n
}

// GOOD: no heap allocation for many cases
func goodAlloc(buf *int) {
    *buf = 42  // caller provides storage, no allocation
}

// BENCHMARK to see the difference
func BenchmarkStackVsHeap(b *testing.B) {
    b.Run("stack", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            x := i  // stack allocated
            _ = x
        }
    })
    
    b.Run("heap", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            x := &i           // forces heap allocation
            _ = x
        }
    })
}
```

---

## Chapter 5: The Garbage Collector — How It Works

Go uses a tri-color mark-sweep garbage collector with low latency goals (<1ms pauses since Go 1.8).

### Garbage Collector Phases

```
1. Mark Start (STW — stop the world, very brief ~100μs)
   - Scans goroutine stacks
   - Enables write barriers

2. Concurrent Marking
   - GC goroutines scan the heap, marking reachable objects
   - Your goroutines continue running concurrently
   - New allocations are handled by write barriers

3. Mark Termination (STW — brief)
   - Ensures all objects are properly marked

4. Sweep (concurrent)
   - Reuses memory from unreachable objects
   - Background goroutines sweep during normal execution
```

```go
// You can interact with the GC:
import "runtime"

// Force a GC cycle (don't do this in production)
runtime.GC()

// Check GC stats
var stats runtime.MemStats
runtime.ReadMemStats(&stats)
fmt.Printf("Total allocated: %d bytes\n", stats.TotalAlloc)
fmt.Printf("Heap in use: %d bytes\n", stats.HeapInuse)
fmt.Printf("GC cycles: %d\n", stats.NumGC)
fmt.Printf("GC pause total: %s\n", time.Duration(stats.PauseTotalNs))

// tune GC with GOGC environment variable
// GOGC=100 (default): GC when heap doubles
// GOGC=50: GC more frequently (less memory, more CPU)
// GOGC=200: GC less frequently (more memory, less CPU)
// GOGC=off: disable GC (dangerous!)
```

### Avoiding GC Pressure

```go
// 1. Reuse objects with sync.Pool
var bufPool = sync.Pool{
    New: func() interface{} { return new(bytes.Buffer) },
}

func processRequest(data []byte) string {
    buf := bufPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufPool.Put(buf)
    }()
    
    buf.Write(data)
    buf.WriteString("\n---processed---")
    return buf.String()
}

// 2. Pre-allocate slices with known capacity
func collectResults(n int) []int {
    results := make([]int, 0, n) // capacity=n, avoids re-allocations
    for i := 0; i < n; i++ {
        results = append(results, i*i)
    }
    return results
}

// 3. Avoid allocations in hot paths
func hotPath(data []byte, result []byte) {
    // Writes to caller-provided result slice — zero allocations
    for i, b := range data {
        result[i] = b ^ 0xFF // XOR each byte
    }
}
```

---

## Chapter 6: Pointers to Interface — A Common Trap

```go
type Stringer interface {
    String() string
}

type User struct{ Name string }
func (u User) String() string { return u.Name }

func printIt(s Stringer) {
    fmt.Println(s.String())
}

// Works:
u := User{Name: "Rohit"}
printIt(u)    // ✅ User implements Stringer

// Also works:
p := &User{Name: "Alice"}
printIt(p)    // ✅ *User also implements Stringer (method set includes value methods)

// TRAP: Pointer to interface (almost always wrong!)
pi := &u       // pi is *User, not *Stringer — this is fine
// psi := &s   // psi would be *Stringer — almost never what you want
// printIt(psi) // ERROR — *Stringer doesn't satisfy Stringer

// Rule: NEVER pass a *Interface. Pass the interface directly.
// If you see *io.Reader, *error, *fmt.Stringer — it's almost always a bug.

// EXCEPTION: json.Unmarshal takes interface{} because it needs to set the value
var result User
json.Unmarshal(data, &result)  // &result is *User, passed as interface{}
```

---

## Chapter 7: unsafe.Pointer — The Escape Hatch

```go
// unsafe.Pointer bypasses Go's type safety
// Use ONLY when absolutely necessary (low-level systems code, performance)

import "unsafe"

// Convert between incompatible pointer types
type MyInt int32
x := int32(42)
myX := (*MyInt)(unsafe.Pointer(&x)) // reinterpret as *MyInt
fmt.Println(*myX) // 42

// Access struct internals (hacky, avoid in production)
type T struct {
    a int32
    b int32
}
t := T{1, 2}
// Get pointer to field b
pb := (*int32)(unsafe.Pointer(uintptr(unsafe.Pointer(&t)) + unsafe.Offsetof(t.b)))
fmt.Println(*pb) // 2

// WHY you might see this:
// - Cgo interfaces (C interop)
// - Very performance-critical serialization
// - Custom garbage-collected data structures
// - NEVER do this in normal application code
```

---

## Chapter 8: Comprehensive Pointer Testing

```go
package pointers_test

import (
    "sync"
    "testing"
)

// Test pass-by-value vs pass-by-pointer
func TestPassByValue(t *testing.T) {
    x := 42
    
    // Pass by value — original unchanged
    modifyValue := func(n int) { n = 100 }
    modifyValue(x)
    if x != 42 {
        t.Errorf("pass by value: expected 42, got %d", x)
    }
    
    // Pass by pointer — original changed
    modifyPointer := func(n *int) { *n = 100 }
    modifyPointer(&x)
    if x != 100 {
        t.Errorf("pass by pointer: expected 100, got %d", x)
    }
}

// Test escape to heap via pointer return
func TestPointerReturn(t *testing.T) {
    makeInt := func(n int) *int {
        // n would normally be on stack, but because we return &n,
        // Go's escape analysis moves it to the heap
        return &n
    }
    
    p := makeInt(42)
    if *p != 42 {
        t.Errorf("heap-escaped pointer: got %d, want 42", *p)
    }
    // p is still valid here — no dangling pointer
}

// Test nil pointer safety
func TestNilPointerSafety(t *testing.T) {
    var p *int
    
    if p != nil {
        t.Error("zero value of pointer should be nil")
    }
    
    // Test that nil pointer dereference panics
    defer func() {
        if r := recover(); r == nil {
            t.Error("expected panic on nil deref")
        }
    }()
    _ = *p // should panic
}

// Test struct with mutex — must not be copied
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}

func TestCounter(t *testing.T) {
    c := &Counter{} // always use pointer for types with mutex
    
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            c.Increment()
        }()
    }
    wg.Wait()
    
    if v := c.Value(); v != 100 {
        t.Errorf("expected 100, got %d (race condition?)", v)
    }
}

// Run with: go test -race ./... to detect race conditions
```

---

## Chapter 9: Memory Layout and Struct Alignment

```go
package main

import (
    "fmt"
    "unsafe"
)

// Struct field alignment matters for memory and performance
type BadAlignment struct {
    a bool    // 1 byte
    // 7 bytes padding inserted by compiler
    b float64 // 8 bytes (needs 8-byte alignment)
    c bool    // 1 byte
    // 7 bytes padding
    d float64 // 8 bytes
}
// size: 1+7+8+1+7+8 = 32 bytes

type GoodAlignment struct {
    b float64 // 8 bytes (aligned)
    d float64 // 8 bytes
    a bool    // 1 byte
    c bool    // 1 byte
    // 6 bytes padding at end
}
// size: 8+8+1+1+6 = 24 bytes

func main() {
    fmt.Println("BadAlignment size:", unsafe.Sizeof(BadAlignment{}))  // 32
    fmt.Println("GoodAlignment size:", unsafe.Sizeof(GoodAlignment{})) // 24
    
    // For large datasets, this 25% size difference is significant!
    
    // Check each field's alignment requirement
    var b BadAlignment
    fmt.Printf("a offset: %d\n", unsafe.Offsetof(b.a)) // 0
    fmt.Printf("b offset: %d\n", unsafe.Offsetof(b.b)) // 8
    fmt.Printf("c offset: %d\n", unsafe.Offsetof(b.c)) // 16
    fmt.Printf("d offset: %d\n", unsafe.Offsetof(b.d)) // 24
}
```

---

**Summary of Part 5:**
- Go pointers are safe — no arithmetic, no dangling pointers (GC prevents them)
- Use `&` to get address, `*` to dereference
- Pass by value for small immutable data; by pointer to modify or for large structs
- Escape analysis: compiler decides stack vs heap — returning `&localVar` is safe!
- Go's GC is concurrent, low-latency — most code needn't worry about it
- Never use `*SomeInterface` — pass interfaces directly
- Struct field order matters for memory size (group same-sized fields)
- `unsafe` exists but almost never belongs in application code
