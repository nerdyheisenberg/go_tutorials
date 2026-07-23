# Go Deep Dive — Part 20: Interview Preparation & Complete Reference

---

## Chapter 1: The 50 Questions Interviewers Ask

### Category 1: Language Fundamentals

**Q1: What is the zero value in Go? Why does it matter?**

Every variable is always initialized to its zero value. `int` → 0, `bool` → false, `string` → `""`, pointer/slice/map/chan/func → `nil`. This eliminates "used before initialization" bugs and makes Go programs predictable. A zero-value struct is always valid to use.

**Q2: What's the difference between nil slice and empty slice?**

```go
var s1 []int          // nil slice: pointer=nil, len=0, cap=0
s2 := []int{}         // empty slice: pointer≠nil, len=0, cap=0
s3 := make([]int, 0) // same as s2

s1 == nil  // true
s2 == nil  // false
// BUT: both work with len(), cap(), range, append
// DIFFERENCE: json.Marshal(s1) = "null", json.Marshal(s2) = "[]"
```

**Q3: How does `append` work? What is the growth strategy?**

`append` returns a new slice that may or may not share the underlying array. When `len == cap`, it allocates a larger array and copies. Growth strategy: roughly double for small slices, ~25% for large ones. **Always assign the result of append**.

**Q4: Explain the interface value representation.**

An interface value is `(concrete_type, concrete_value)`. A nil interface has both fields nil. A non-nil interface can hold a nil pointer — causing the infamous nil interface bug.

```go
var err error       // nil — both fields nil
var p *MyError = nil
err = p             // NON-nil — type field = *MyError, value field = nil
err == nil          // FALSE! Type field is set
```

**Q5: What is the difference between `make` and `new`?**

- `new(T)`: allocates zeroed T, returns `*T`. Works with any type.
- `make(T, ...)`: allocates AND initializes. Only for slice, map, chan. Returns T (not *T).

```go
p := new([]int)   // *[]int pointing to nil slice — rarely useful
s := make([]int, 5) // []int with 5 zeros — what you want
```

---

### Category 2: Concurrency

**Q6: What is a goroutine? How is it different from a thread?**

Goroutine is a lightweight concurrent function managed by Go runtime, NOT the OS. Initial stack ~2KB (vs OS thread ~1-8MB), millions can run concurrently. Scheduled by Go's M:N scheduler (goroutines on OS threads), not the OS kernel directly.

**Q7: What is a data race? How do you detect it?**

A data race is when two goroutines access the same memory concurrently and at least one write, without synchronization. Detect with `go test -race ./...` or `go run -race main.go`. Prevent with `sync.Mutex`, `sync/atomic`, or channels.

**Q8: When do you use channels vs mutexes?**

- **Channels**: Passing ownership of data between goroutines, signaling events, pipeline processing
- **Mutex**: Protecting shared state that's read/modified by multiple goroutines (cache, counter, shared map)

**Q9: What happens when you send to a closed channel? Receive from a closed channel?**

- Send to closed: **panic**
- Receive from closed: returns immediately with zero value and `ok=false`. After buffer drained, all future receives return `(zero, false)` — never blocks.

**Q10: Explain the select statement. What if multiple cases are ready?**

`select` waits on multiple channel operations. If **multiple cases are ready simultaneously, Go picks one at random** — this prevents starvation. A `default` case makes select non-blocking.

**Q11: What is a goroutine leak?**

A goroutine that runs forever without being able to exit — typically blocked on a channel that nobody sends to (or reads from). Always pair goroutines with a cancellation mechanism (`context.Context` or `done` channel).

**Q12: What is `sync.Once` used for?**

Ensures a function runs exactly once, even if called from multiple goroutines concurrently. Common use: lazy singleton initialization. The function runs while all other callers block.

---

### Category 3: Interfaces and Types

**Q13: How does Go implement polymorphism without classes?**

Via interfaces. Any type that has all the methods an interface requires automatically satisfies it — no explicit declaration. This is structural (duck) typing. Interfaces are defined by the CONSUMER, not the implementation.

**Q14: What is method set? Why does it matter?**

The set of methods callable on a type. Value type `T` has only value receiver methods. Pointer type `*T` has both value and pointer receiver methods. This determines which interfaces a type satisfies.

```go
type I interface { M() }
type T struct{}
func (T) M() {}   // value receiver

var _ I = T{}    // OK: T's method set includes M()
var _ I = &T{}   // OK: *T's method set also includes M()

func (t *T) M2() {} // pointer receiver added

// Now: 
var _ = T{}.M2()   // still compiles (Go auto-takes address: (&T{}).M2())
// BUT: T{} no longer satisfies I if M() is changed to pointer receiver!
```

**Q15: What is the empty interface? When should you use it?**

`interface{}` (or `any` in 1.18+) accepts any type. Use when you genuinely don't know the type (JSON unmarshaling, generic containers). Avoid when a specific interface or generics would work — you lose type safety.

---

### Category 4: Error Handling

**Q16: What is the idiomatic way to handle errors in Go?**

Return `(result, error)`. Check `if err != nil`. Wrap with context using `fmt.Errorf("context: %w", err)`. Use `errors.Is()` to check specific errors, `errors.As()` to extract typed errors. Never ignore errors.

**Q17: What is the difference between `errors.Is` and `errors.As`?**

- `errors.Is(err, target)`: checks if `err` OR any wrapped error **is** the target (by identity or `.Is()` method)
- `errors.As(err, &target)`: checks if `err` OR any wrapped error **can be typed as** `*target` — and assigns it

**Q18: When should you use panic and recover?**

`panic`: for programmer errors (assertion failures, broken invariants), never for expected runtime errors. `recover`: in frameworks/servers to catch panics from handlers and return 500 instead of crashing. Never use panic/recover as a control flow substitute for errors.

---

### Category 5: Memory and Performance

**Q19: What is escape analysis?**

The compiler determines whether a variable can stay on the stack (faster, auto-freed) or must move to the heap (GC-managed). Variables returned by pointer, captured by closures, or stored in interfaces typically escape to the heap. Check with `go build -gcflags="-m"`.

**Q20: How does Go's garbage collector work?**

Tri-color mark-sweep, running concurrently with your program. Stop-the-world pauses are < 1ms since Go 1.8. Tune with `GOGC` environment variable. Reduce GC pressure with `sync.Pool`, pre-allocated slices, avoiding unnecessary allocations.

---

## Chapter 2: Code Patterns — Quick Reference

### All Go Control Flow
```go
// if with init statement
if err := doSomething(); err != nil { ... }

// switch forms
switch x { case 1, 2: ... }             // value
switch { case x > 0: ... }              // no condition (replaces if/else)
switch v := x.(type) { case int: ... }  // type switch

// for forms
for {}                          // infinite
for condition {}                // while-style
for i := 0; i < n; i++ {}      // C-style
for i, v := range slice {}      // range over slice
for k, v := range m {}          // range over map
for v := range ch {}            // range over channel (blocks until close)
for i := range 10 {}            // Go 1.22+: range over integer
```

### Channel Patterns at a Glance
```go
// Unbuffered: sync handshake
ch := make(chan T)

// Buffered: async up to capacity
ch := make(chan T, n)

// Directional types
func producer(out chan<- T) {}    // send-only
func consumer(in <-chan T) {}     // receive-only

// Close: sender closes
close(ch)                         // signals receivers
for v := range ch { }            // reads until close

// Select
select {
case v := <-ch1: ...
case ch2 <- x: ...
case <-time.After(d): ...       // timeout
default: ...                      // non-blocking
}
```

### Error Handling Quick Reference
```go
// Create
var ErrNotFound = errors.New("not found")  // sentinel
fmt.Errorf("ctx: %w", err)                 // wrap
fmt.Errorf("ctx: %v", err)                 // wrap (NOT unwrappable — avoid)

// Check
errors.Is(err, ErrNotFound)               // identity/sentinel check
errors.As(err, &myErr)                    // type extraction

// Error type
type MyError struct { Field string }
func (e *MyError) Error() string { ... }
func (e *MyError) Is(target error) bool { _, ok := target.(*MyError); return ok }
```

### Interface Quick Reference
```go
// Define
type Doer interface { Do() error }

// Compile-time check
var _ Doer = (*MyType)(nil)  // if MyType doesn't implement Doer → compile error

// Type assertion
v, ok := ifaceVal.(ConcreteType)   // safe form (comma-ok)
v := ifaceVal.(ConcreteType)        // panics if wrong

// Type switch
switch v := x.(type) {
case int:     ...
case string:  ...
case Doer:    v.Do()    // v is typed as Doer here
default:      ...
}
```

---

## Chapter 3: Slice and Map Quick Reference

```go
// ===== SLICE =====
// Create
var s []int                          // nil slice
s := []int{1, 2, 3}                  // literal
s := make([]int, length)             // all zeros
s := make([]int, length, capacity)   // preallocate

// Key operations
append(s, v)                         // always assign result!
copy(dst, src)                       // min(len(dst), len(src)) elements
s[low:high]                          // subslice, shares array
s[low:high:max]                      // limit capacity

// Built-ins
len(s), cap(s)

// ===== MAP =====
m := make(map[K]V)
m := map[K]V{"key": value}

m["key"] = value                     // write
v := m["key"]                        // read (zero if missing)
v, ok := m["key"]                    // comma-ok (check existence)
delete(m, "key")                     // delete (safe on missing key)
len(m)                               // size

// Iteration (RANDOM ORDER)
for k, v := range m { }
for k := range m { }               // keys only
```

---

## Chapter 4: Goroutine Patterns Quick Reference

```go
// ===== SYNC =====
var wg sync.WaitGroup
wg.Add(1)
go func() { defer wg.Done(); work() }()
wg.Wait()

var mu sync.Mutex
mu.Lock(); defer mu.Unlock()

var rmu sync.RWMutex
rmu.RLock(); defer rmu.RUnlock()  // multiple readers
rmu.Lock(); defer rmu.Unlock()    // single writer

var once sync.Once
once.Do(func() { /* runs once */ })

// ===== ATOMIC =====
var n atomic.Int64
n.Add(1); n.Load(); n.Store(42)
n.CompareAndSwap(old, new)

// ===== CONTEXT =====
ctx := context.Background()
ctx, cancel := context.WithCancel(parent)
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()

select {
case <-ctx.Done():
    return ctx.Err()  // Canceled or DeadlineExceeded
default:
}
```

---

## Chapter 5: Testing Quick Reference

```bash
# Run tests
go test ./...
go test -v ./...              # verbose
go test -run TestFoo ./...    # specific test
go test -run TestFoo/subtest  # specific subtest
go test -race ./...           # race detection (ALWAYS IN CI)
go test -cover ./...          # coverage
go test -bench=. ./...        # benchmarks
go test -benchmem ./...       # with memory stats
go test -count=1 ./...        # disable cache
go test -short ./...          # skip slow tests
```

```go
// ===== BASIC TEST =====
func TestFoo(t *testing.T) {
    t.Run("subtest", func(t *testing.T) {
        t.Parallel()
        if got := foo(); got != want {
            t.Errorf("foo() = %v, want %v", got, want)
        }
    })
}

// ===== TABLE-DRIVEN =====
tests := []struct{ in, want int }{
    {1, 1}, {0, 0}, {-1, -1},
}
for _, tt := range tests {
    t.Run(fmt.Sprint(tt.in), func(t *testing.T) {
        if got := fn(tt.in); got != tt.want {
            t.Errorf("fn(%d) = %d, want %d", tt.in, got, tt.want)
        }
    })
}

// ===== SETUP/TEARDOWN =====
func TestMain(m *testing.M) {          // package-level
    setup()
    code := m.Run()
    teardown()
    os.Exit(code)
}

t.Cleanup(func() { cleanup() })        // per-test

// ===== HELPERS =====
func require(t *testing.T, err error) {
    t.Helper()  // critical!
    if err != nil { t.Fatal(err) }
}
```

---

## Chapter 6: Standard Library Cheat Sheet

```go
// fmt
fmt.Sprintf("%v %+v %#v %T", x, x, x, x)  // debug verbs
fmt.Printf("%d %s %f %.2f %t %p\n", ...)   // common verbs

// strings
strings.Contains/HasPrefix/HasSuffix
strings.Split/SplitN/Fields/Join
strings.TrimSpace/Trim/TrimPrefix/TrimSuffix
strings.Replace/ReplaceAll/ToUpper/ToLower
strings.Builder — efficient concatenation
strings.NewReader — io.Reader from string

// strconv
strconv.Itoa(n) / strconv.Atoi(s)           // int ↔ string
strconv.FormatInt(n, base) / ParseInt(s, base, bits)
strconv.FormatFloat(f, 'f', prec, 64) / ParseFloat(s, 64)

// io
io.ReadAll(r) / io.Copy(dst, src)
io.LimitReader / io.TeeReader / io.MultiWriter
io.ReadFull / io.NopCloser

// os
os.ReadFile / os.WriteFile                   // simple file I/O
os.Open / os.Create / os.OpenFile           // file handles
os.Stat / os.IsNotExist / os.MkdirAll      // file info
os.Getenv / os.Setenv / os.Environ         // environment

// encoding/json
json.Marshal / json.Unmarshal              // []byte ↔ Go
json.NewEncoder(w).Encode(v)               // streaming (preferred for HTTP)
json.NewDecoder(r).Decode(&v)

// time
time.Now() / time.Since(t) / time.Until(t)
t.Format("2006-01-02 15:04:05")            // reference time!
time.Parse(format, s)
time.NewTicker(d) / time.NewTimer(d)
```

---

## Chapter 7: Common Interview Exercises — Solutions

### Exercise 1: FizzBuzz

```go
func fizzBuzz(n int) []string {
    result := make([]string, n)
    for i := 1; i <= n; i++ {
        switch {
        case i%15 == 0: result[i-1] = "FizzBuzz"
        case i%3 == 0:  result[i-1] = "Fizz"
        case i%5 == 0:  result[i-1] = "Buzz"
        default:         result[i-1] = strconv.Itoa(i)
        }
    }
    return result
}
```

### Exercise 2: Concurrent Map with Locking

```go
type SafeMap[K comparable, V any] struct {
    mu sync.RWMutex
    m  map[K]V
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
    return &SafeMap[K, V]{m: make(map[K]V)}
}

func (sm *SafeMap[K, V]) Set(k K, v V) {
    sm.mu.Lock(); defer sm.mu.Unlock()
    sm.m[k] = v
}

func (sm *SafeMap[K, V]) Get(k K) (V, bool) {
    sm.mu.RLock(); defer sm.mu.RUnlock()
    v, ok := sm.m[k]
    return v, ok
}
```

### Exercise 3: Implement a Semaphore

```go
type Semaphore struct { ch chan struct{} }

func NewSemaphore(n int) *Semaphore { return &Semaphore{ch: make(chan struct{}, n)} }
func (s *Semaphore) Acquire() { s.ch <- struct{}{} }
func (s *Semaphore) Release() { <-s.ch }

func (s *Semaphore) AcquireCtx(ctx context.Context) error {
    select {
    case s.ch <- struct{}{}: return nil
    case <-ctx.Done(): return ctx.Err()
    }
}
```

### Exercise 4: Fan-out / Fan-in (Pipeline)

```go
func merge(cs ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    out := make(chan int)
    
    output := func(c <-chan int) {
        defer wg.Done()
        for n := range c { out <- n }
    }
    
    wg.Add(len(cs))
    for _, c := range cs { go output(c) }
    
    go func() { wg.Wait(); close(out) }()
    return out
}
```

### Exercise 5: LRU Cache

```go
import "container/list"

type LRUCache struct {
    cap   int
    mu    sync.Mutex
    items map[int]*list.Element
    list  *list.List
}

type entry struct{ key, val int }

func NewLRU(cap int) *LRUCache {
    return &LRUCache{cap: cap, items: make(map[int]*list.Element), list: list.New()}
}

func (c *LRUCache) Get(key int) (int, bool) {
    c.mu.Lock(); defer c.mu.Unlock()
    if el, ok := c.items[key]; ok {
        c.list.MoveToFront(el)
        return el.Value.(*entry).val, true
    }
    return 0, false
}

func (c *LRUCache) Put(key, value int) {
    c.mu.Lock(); defer c.mu.Unlock()
    if el, ok := c.items[key]; ok {
        c.list.MoveToFront(el)
        el.Value.(*entry).val = value
        return
    }
    if c.list.Len() == c.cap {
        // Evict LRU
        back := c.list.Back()
        c.list.Remove(back)
        delete(c.items, back.Value.(*entry).key)
    }
    el := c.list.PushFront(&entry{key, value})
    c.items[key] = el
}
```

---

## Chapter 8: C++ → Go Mental Models — Final Summary

| C++ Thinking | → | Go Thinking |
|---|---|---|
| "How do I manage this memory?" | → | "GC handles it — just use `*T`" |
| "Which exception might this throw?" | → | "Which error does this return?" |
| "What's the base class?" | → | "Which interface does this implement?" |
| "I'll use templates for this" | → | "Can an interface solve this? If not, generics" |
| "Add a virtual method" | → | "Add a method to the interface" |
| "RAII for cleanup" | → | "`defer` for cleanup" |
| "Launch a thread" | → | "`go func()`" |
| "Condition variable" | → | "Channel or `sync.Cond`" |
| "Make it header-only" | → | "Put it in `internal/`" |
| "Multiple inheritance" | → | "Embed multiple structs, implement multiple interfaces" |
| "Friend class" | → | "Same package (package-level access)" |
| "Explicit constructor" | → | "`NewXxx()` factory function" |
| "Copy constructor" | → | "Value semantics — Go copies by default" |
| "Move semantics" | → | "Slices/maps/channels are reference types — passing is cheap" |

---

## Chapter 9: Final Checklist Before the Interview

### Core Concepts (must-know)
- [ ] Goroutines vs OS threads (GMP model)
- [ ] Channels — buffered vs unbuffered, close rules, select
- [ ] Interfaces — implicit implementation, method sets, nil interface bug
- [ ] Error handling — wrapping, errors.Is, errors.As, sentinel errors
- [ ] defer — LIFO, argument capture, named return modification
- [ ] Slice internals — pointer/len/cap, append growth, sharing
- [ ] Map — hash table, nil map rules, concurrency (not safe!)
- [ ] Context — cancellation tree, timeout, values

### Common Gotchas (differentiate you from beginners)
- [ ] Nil interface ≠ nil typed pointer stored in interface
- [ ] Range variable captures (goroutine loop bug — fixed in Go 1.22)
- [ ] Map iteration is random order (by design)
- [ ] Modifying slice/struct in range loop copies value, not original
- [ ] defer arguments evaluated at defer time, not execution time
- [ ] sync.Mutex must NOT be copied — always use pointer
- [ ] Never store context.Context in a struct

### Patterns (show your experience)
- [ ] Table-driven tests with t.Run subtests
- [ ] Functional options for constructors
- [ ] Repository + Service + Handler architecture
- [ ] Worker pool with graceful shutdown
- [ ] Context cancellation through goroutine trees
- [ ] Timeout with context.WithTimeout

---

**Congratulations! You have completed the 20-part Go Deep Dive.**

You now understand:
- **Why** Go makes each design decision (philosophy, not just syntax)
- **What** every language feature does (theory + internals)
- **How** to write idiomatic Go (practical code + patterns)
- **When** things go wrong (negative cases + edge cases)
- **How to test** everything properly (positive + negative + benchmarks)
- **How to migrate** C++ patterns to Go equivalents

Go build something awesome! 🚀
