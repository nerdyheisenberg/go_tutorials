# Go Deep Dive — Part 6: Arrays, Slices & Maps — The Deep Internals

---

## Chapter 1: Arrays — The Foundation

### Why Arrays Exist (and Why You Rarely Use Them)

An **array** is a fixed-size sequence of elements of the same type. The size is part of the **type**:

```go
[3]int  // array of 3 ints
[5]int  // DIFFERENT type from [3]int
```

Arrays are **value types** in Go — assigning or passing them copies the entire array.

```go
// Declaration forms
var a [5]int                    // [0, 0, 0, 0, 0]
b := [3]string{"go", "is", "fast"}
c := [...]int{1, 2, 3, 4, 5}   // size inferred from literal (5)

// Array size is compile-time constant
n := 5
// var d [n]int  // COMPILE ERROR: non-constant array bound
var d [5]int     // OK: constant

// Comparison
x := [3]int{1, 2, 3}
y := [3]int{1, 2, 3}
fmt.Println(x == y) // true — arrays are comparable if element type is comparable

// [3]int and [4]int cannot be compared — different types
```

### When to Use Arrays

```go
// 1. Fixed-size buffers
var buf [4096]byte  // 4KB buffer on the stack (fast!)

// 2. Known-size data (SHA, UUIDs)
var hash [32]byte  // SHA-256 hash
var uuid [16]byte  // UUID

// 3. 2D grids, game boards
var chessBoard [8][8]string  // 8x8 grid

// 4. Crypto keys (fixed known size)
type AES256Key [32]byte
```

---

## Chapter 2: Slices — The Most Important Data Structure

### What IS a Slice Internally?

A slice is a **descriptor** (not the data itself) — a small struct with three fields:

```go
// Internal representation (from Go runtime):
type slice struct {
    array uintptr  // pointer to underlying array element (8 bytes)
    len   int      // number of elements (8 bytes)
    cap   int      // capacity = len(underlying array) - offset (8 bytes)
}
// Total: 24 bytes — always, regardless of how many elements
```

```go
// Visualizing slice internals:
original := []int{10, 20, 30, 40, 50}
//                      ↑ pointer
//                      len=5, cap=5

sub := original[1:3]   // slice of original from index 1 to 3 (exclusive)
// sub = {array=&original[1], len=2, cap=4}
// The pointer points to original[1] (=20)
// cap = 5-1 = 4 (distance from start of sub to end of underlying array)

fmt.Println(sub)        // [20 30]
fmt.Println(len(sub))   // 2
fmt.Println(cap(sub))   // 4

// CRITICAL: sub and original SHARE the same underlying array!
sub[0] = 999
fmt.Println(original)  // [10 999 30 40 50] — original is MODIFIED!
```

### Creating Slices — All Forms

```go
// Form 1: Literal
s1 := []int{1, 2, 3, 4, 5}
s2 := []string{"a", "b", "c"}
s3 := []int{}           // empty slice (NOT nil, length=0, cap=0)

// Form 2: make(type, length, capacity)
s4 := make([]int, 5)       // len=5, cap=5, all zeros
s5 := make([]int, 0, 100)  // len=0, cap=100 — pre-allocated!
s6 := make([]int, 5, 10)   // len=5, cap=10

// Form 3: nil slice
var s7 []int               // len=0, cap=0, pointer=nil

// Form 4: Slice of array
arr := [5]int{1, 2, 3, 4, 5}
s8 := arr[1:4]            // slice of arr: [2, 3, 4]
s9 := arr[:]              // entire array as slice

// nil vs empty — important distinction
fmt.Println(s7 == nil)    // true
fmt.Println(s3 == nil)    // false (empty, not nil)
fmt.Println(len(s7))      // 0
fmt.Println(len(s3))      // 0 — both have len 0!

// Both nil and empty slices:
// - Work with len(), cap(), append(), range
// - JSON encode to null vs []
import "encoding/json"
d1, _ := json.Marshal(s7) // "null"
d2, _ := json.Marshal(s3) // "[]"
// This matters for APIs!
```

### append — The Growth Algorithm

```go
// append returns a new slice (possibly with new underlying array)
s := make([]int, 0, 3)
fmt.Println(len(s), cap(s))  // 0 3

s = append(s, 1)
fmt.Println(len(s), cap(s))  // 1 3

s = append(s, 2)
fmt.Println(len(s), cap(s))  // 2 3

s = append(s, 3)
fmt.Println(len(s), cap(s))  // 3 3

s = append(s, 4) // exceeds capacity!
// Go allocates a new larger array and copies data
fmt.Println(len(s), cap(s))  // 4 6 (doubled!)

// Growth strategy (as of Go 1.18+):
// - Small slices (cap < 256): double
// - Large slices: grow by ~25%
// - Exact formula is complex, subject to change — don't depend on it!

// CRITICAL: append may or may not return the same underlying array
// ALWAYS assign the result of append:
s = append(s, 5)  // CORRECT

// WRONG:
append(s, 5)  // result discarded — compiler will warn but it's valid Go
              // (you'd never see this in real code)
```

### The append Gotcha — Slice Mutation

```go
// This is the #1 source of slice-related bugs

a := []int{1, 2, 3}    // a: {array, len=3, cap=3}
b := a                  // b: {SAME array, len=3, cap=3}

b = append(b, 4)        // b exceeds cap → NEW array allocated
// Now: a still points to old array, b points to new array
b[0] = 99
fmt.Println(a)  // [1 2 3] — a unchanged (different underlying array)
fmt.Println(b)  // [99 2 3 4]

// But with capacity:
a2 := make([]int, 3, 10) // plenty of capacity!
copy(a2, []int{1, 2, 3})
b2 := a2                  // SAME underlying array

b2 = append(b2, 4)       // no reallocation (cap=10 > len=3)
// b2[3]=4 was written to the SHARED array
b2[0] = 99
fmt.Println(a2)  // [99 2 3] — a2 IS affected! (shared array)
fmt.Println(b2)  // [99 2 3 4]

// Lesson: after subslicing or sharing a slice,
// you can't predict next append will allocate a new array or not.
// ALWAYS use copy() to get a truly independent slice:
c := make([]int, len(a))
copy(c, a)  // truly independent copy
```

### Using copy Correctly

```go
// copy(dst, src) copies min(len(dst), len(src)) elements
// Returns number of elements copied

src := []int{1, 2, 3, 4, 5}
dst := make([]int, 3)      // only 3 slots

n := copy(dst, src)
fmt.Println(dst, n)         // [1 2 3] 3

// Copying between slices (overlapping is ok!)
s := []int{1, 2, 3, 4, 5}
copy(s[1:], s)              // shift right by 1: [1 1 2 3 4]

// Deep copy pattern
func copySlice(s []int) []int {
    if s == nil {
        return nil
    }
    result := make([]int, len(s))
    copy(result, s)
    return result
}
```

### Full-Slice Expression — Controlling Capacity After Subslicing

```go
// s[low : high : max]
// Creates slice: pointer=&s[low], len=high-low, cap=max-low

original := []int{1, 2, 3, 4, 5}

// Regular subslice (shares capacity with original)
s1 := original[1:3]        // len=2, cap=4 (can see original[3] and original[4])

// Full slice expression (limits capacity)
s2 := original[1:3:3]      // len=2, cap=2 (max-low = 3-1 = 2)
// Now append to s2 WILL allocate a new array (cap exhausted)
s2 = append(s2, 99)        // new allocation!
fmt.Println(original)       // [1 2 3 4 5] — original untouched
fmt.Println(s2)             // [2 3 99]
```

### Common Slice Operations — How and Why

```go
package main

import (
    "sort"
    "fmt"
)

func main() {
    s := []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3}
    
    // Sort (in-place)
    sort.Ints(s)
    fmt.Println(s) // [1 1 2 3 3 4 5 5 6 9]
    
    // Filter (build new slice)
    var odds []int
    for _, v := range s {
        if v%2 != 0 {
            odds = append(odds, v)
        }
    }
    
    // Map (transform)
    doubled := make([]int, len(s))
    for i, v := range s {
        doubled[i] = v * 2
    }
    
    // Reduce
    sum := 0
    for _, v := range s {
        sum += v
    }
    
    // Remove element at index i (preserving order)
    i := 3
    s = append(s[:i], s[i+1:]...)
    
    // Remove element at index i (fast, order not preserved)
    s[i] = s[len(s)-1]
    s = s[:len(s)-1]
    
    // Insert at index i
    i = 2
    val := 99
    s = append(s[:i], append([]int{val}, s[i:]...)...) // creates intermediate slice
    
    // Better insert (no intermediate allocation):
    s = append(s, 0)         // grow by 1
    copy(s[i+1:], s[i:])    // shift right
    s[i] = val              // set value
    
    // Contains check
    target := 5
    found := false
    for _, v := range s {
        if v == target {
            found = true
            break
        }
    }
    fmt.Println("contains 5:", found)
    
    // Binary search (on sorted slice)
    sorted := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    idx := sort.SearchInts(sorted, 7) // 6 (index of 7)
    fmt.Println("index of 7:", idx)
    
    _ = odds
    _ = doubled
    _ = sum
}
```

---

## Chapter 3: Maps — Complete Reference

### Map Internals

A Go map is a hash table. Internally it uses **buckets** — each bucket holds 8 key-value pairs. When a bucket is full, a new one is chained.

```go
// Internal representation (simplified):
type hmap struct {
    count     int        // number of live keys
    B         uint8      // log_2 of # of buckets
    buckets   unsafe.Pointer  // array of 2^B buckets
    oldbuckets unsafe.Pointer // for growing maps
    // ...
}
```

**Map lookup**: O(1) average, O(n) worst case (with hash collisions).
**Map ordering**: INTENTIONALLY random — Go randomizes map iteration to prevent programs from depending on insertion order.

```go
// Why random map order?
// Programs should not depend on map iteration order.
// Making it random in every run forces programmers to not rely on it.
// This enables the runtime to freely reorganize maps for performance.
```

### Creating Maps — All Forms

```go
// Form 1: make
m1 := make(map[string]int)        // empty map, ready to use
m2 := make(map[string]int, 100)   // pre-allocated for ~100 entries (hint, not cap)

// Form 2: Literal
m3 := map[string]int{
    "alice": 25,
    "bob":   30,
    "carol": 22,
}

// Form 3: nil map (DANGEROUS — read is OK, write panics)
var m4 map[string]int  // nil!
fmt.Println(m4["key"]) // 0 — reading nil map returns zero value (OK)
// m4["key"] = 1       // PANIC: assignment to entry in nil map

// Always initialize!
m4 = make(map[string]int)
m4["key"] = 1  // Now OK
```

### Map Operations

```go
m := make(map[string]int)

// Write
m["alice"] = 25
m["bob"] = 30

// Read (always safe, returns zero value if key missing)
age := m["alice"]     // 25
missing := m["carol"] // 0 — no error, no panic

// Check existence (comma-ok idiom)
age, exists := m["alice"]
if exists {
    fmt.Println("alice's age:", age)
}

// The single-value form is ambiguous:
age2 := m["carol"]  // 0 — but is carol 0, or does carol not exist?
// Solution: always use comma-ok when you need to distinguish

// Delete
delete(m, "bob")         // safe even if key doesn't exist
delete(m, "nonexistent") // no error

// Length
fmt.Println(len(m)) // 1 (only alice now)

// Iterate (random order!)
for key, value := range m {
    fmt.Printf("%s: %d\n", key, value)
}

// Iterate only keys
for key := range m {
    fmt.Println(key)
}

// Sorted iteration
keys := make([]string, 0, len(m))
for k := range m {
    keys = append(keys, k)
}
sort.Strings(keys)
for _, k := range keys {
    fmt.Printf("%s: %d\n", k, m[k])
}
```

### Map with Complex Keys and Values

```go
// Any comparable type can be a MAP KEY
// int, string, bool, pointer, array, struct (if all fields comparable)
// CANNOT be key: slice, map, function

// Struct as key
type Point struct{ X, Y int }
distances := map[Point]float64{
    {0, 0}: 0.0,
    {3, 4}: 5.0,  // Pythagorean triple
}
d := distances[Point{3, 4}] // 5.0

// Array as key (fixed size makes it comparable unlike slice!)
type RGB [3]uint8
colorNames := map[RGB]string{
    {255, 0, 0}:   "red",
    {0, 255, 0}:   "green",
    {0, 0, 255}:   "blue",
}

// Map as value
graph := map[string][]string{
    "A": {"B", "C"},
    "B": {"D"},
    "C": {"D", "E"},
    "D": {},
    "E": {},
}

// Nested maps
matrix := map[string]map[string]int{
    "users":   {"count": 100, "active": 80},
    "orders":  {"count": 500, "pending": 50},
}
fmt.Println(matrix["users"]["count"]) // 100

// IMPORTANT: Nested map must be initialized at each level
data := make(map[string]map[string]int)
// data["users"]["count"] = 100  // PANIC: data["users"] is nil!

// Correct:
data["users"] = make(map[string]int)
data["users"]["count"] = 100  // OK
```

### Map Patterns

```go
// Pattern 1: Grouping / Group-by
type Person struct { Name, City string }
people := []Person{
    {"Alice", "Delhi"}, {"Bob", "Mumbai"}, 
    {"Charlie", "Delhi"}, {"Dave", "Mumbai"},
}

// Group by city
byCity := make(map[string][]Person)
for _, p := range people {
    byCity[p.City] = append(byCity[p.City], p)  // append to nil slice is OK!
}
// {"Delhi": [Alice, Charlie], "Mumbai": [Bob, Dave]}

// Pattern 2: Counting / Frequency
words := []string{"go", "is", "great", "go", "is", "amazing", "go"}
freq := make(map[string]int)
for _, w := range words {
    freq[w]++  // zero value of int is 0, so this works for first occurrence
}
// {"go": 3, "is": 2, "great": 1, "amazing": 1}

// Pattern 3: Set (using empty struct for zero memory)
seen := make(map[string]struct{})
unique := []string{}
for _, w := range words {
    if _, ok := seen[w]; !ok {
        seen[w] = struct{}{}
        unique = append(unique, w)
    }
}
// unique = ["go", "is", "great", "amazing"]

// Pattern 4: Cache / Memoization
type cache struct {
    mu   sync.RWMutex
    data map[string]int
}

func (c *cache) Get(key string) (int, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    v, ok := c.data[key]
    return v, ok
}

func (c *cache) Set(key string, value int) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value
}

// Pattern 5: Lookup table (replacing switch)
var dayName = map[int]string{
    0: "Sunday", 1: "Monday", 2: "Tuesday",
    3: "Wednesday", 4: "Thursday", 5: "Friday", 6: "Saturday",
}

func getDayName(day int) (string, bool) {
    name, ok := dayName[day]
    return name, ok
}
```

### Maps and Concurrency — Critical!

```go
// Maps are NOT safe for concurrent access!
// Concurrent reads: OK
// Concurrent writes: DATA RACE → crash
// Concurrent read + write: DATA RACE → crash

// OPTION 1: sync.Mutex (simple, general)
type SafeMap struct {
    mu sync.RWMutex
    data map[string]int
}

func NewSafeMap() *SafeMap {
    return &SafeMap{data: make(map[string]int)}
}

func (m *SafeMap) Get(k string) (int, bool) {
    m.mu.RLock()   // multiple readers can hold this simultaneously
    defer m.mu.RUnlock()
    v, ok := m.data[k]
    return v, ok
}

func (m *SafeMap) Set(k string, v int) {
    m.mu.Lock()    // exclusive write lock
    defer m.mu.Unlock()
    m.data[k] = v
}

// OPTION 2: sync.Map (built-in, optimized for specific patterns)
// Use sync.Map when:
// - Read-heavy (many more reads than writes)
// - Many goroutines write different keys (low contention)
// NOT when: you need iteration over all items efficiently

import "sync"

var m sync.Map

m.Store("key", 42)
val, ok := m.Load("key")       // ok = true
m.Delete("key")
actual, loaded := m.LoadOrStore("key", 99) // atomic check-and-set
m.Range(func(k, v interface{}) bool {
    fmt.Println(k, v)
    return true // return false to stop iteration
})

// DETECT DATA RACES:
// go test -race ./...
// go run -race main.go
```

---

## Chapter 4: Comprehensive Testing for Arrays, Slices, Maps

```go
package collections_test

import (
    "sort"
    "sync"
    "testing"
    "reflect"
)

// ==================== SLICE TESTS ====================

func TestSliceInternals(t *testing.T) {
    // Test that subslice shares memory
    original := []int{1, 2, 3, 4, 5}
    sub := original[1:3]
    
    if len(sub) != 2 { t.Errorf("sub len: got %d, want 2", len(sub)) }
    if cap(sub) != 4 { t.Errorf("sub cap: got %d, want 4", cap(sub)) }
    
    // Modifying through sub modifies original
    sub[0] = 99
    if original[1] != 99 {
        t.Error("modifying sub should modify original (shared memory)")
    }
}

func TestAppendGrowth(t *testing.T) {
    s := make([]int, 0, 3)
    
    // Under capacity — same backing array
    prevArray := &s[0:cap(s)][0] // hacky way to get underlying pointer
    s = append(s, 1, 2, 3)
    // Still same array (no growth)
    
    // Over capacity — new array
    oldLen, oldCap := len(s), cap(s)
    s = append(s, 4) // exceeds cap
    
    if len(s) != oldLen+1 {
        t.Errorf("len after grow: got %d", len(s))
    }
    if cap(s) <= oldCap {
        t.Errorf("cap should grow after exceeding: was %d, now %d", oldCap, cap(s))
    }
    
    _ = prevArray
}

func TestCopyIndependence(t *testing.T) {
    original := []int{1, 2, 3}
    copied := make([]int, len(original))
    copy(copied, original)
    
    // Modify copy — original should be unaffected
    copied[0] = 99
    if original[0] == 99 {
        t.Error("copy should be independent from original")
    }
}

func TestNilVsEmptySlice(t *testing.T) {
    var nilSlice []int
    emptySlice := []int{}
    
    if nilSlice != nil {
        t.Error("var []int should be nil")
    }
    // Note: can't compare slices with ==, only with nil
    
    // Both work with append
    nilSlice = append(nilSlice, 1)
    emptySlice = append(emptySlice, 1)
    
    if len(nilSlice) != 1 || len(emptySlice) != 1 {
        t.Error("append should work on both nil and empty slice")
    }
}

// Filter function test
func filter(s []int, fn func(int) bool) []int {
    var result []int
    for _, v := range s {
        if fn(v) {
            result = append(result, v)
        }
    }
    return result
}

func TestFilter(t *testing.T) {
    tests := []struct {
        name   string
        input  []int
        fn     func(int) bool
        want   []int
    }{
        {"keep evens", []int{1,2,3,4,5}, func(n int) bool { return n%2==0 }, []int{2,4}},
        {"keep positives", []int{-1,2,-3,4}, func(n int) bool { return n>0 }, []int{2,4}},
        {"keep all", []int{1,2,3}, func(int) bool { return true }, []int{1,2,3}},
        {"keep none", []int{1,2,3}, func(int) bool { return false }, nil},
        {"empty input", []int{}, func(int) bool { return true }, nil},
        {"nil input", nil, func(int) bool { return true }, nil},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := filter(tt.input, tt.fn)
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("filter() = %v, want %v", got, tt.want)
            }
        })
    }
}

// ==================== MAP TESTS ====================

func TestMapBasics(t *testing.T) {
    m := make(map[string]int)
    
    // Set and get
    m["alice"] = 25
    if got := m["alice"]; got != 25 {
        t.Errorf("m[alice] = %d, want 25", got)
    }
    
    // Missing key returns zero value
    if got := m["nobody"]; got != 0 {
        t.Errorf("missing key should return 0, got %d", got)
    }
    
    // Comma-ok pattern
    _, exists := m["alice"]
    if !exists { t.Error("alice should exist") }
    
    _, exists = m["nobody"]
    if exists { t.Error("nobody should not exist") }
    
    // Delete
    delete(m, "alice")
    _, exists = m["alice"]
    if exists { t.Error("alice should be deleted") }
    
    // Delete non-existent key (should not panic)
    delete(m, "alice") // OK, no-op
}

func TestNilMapRead(t *testing.T) {
    var m map[string]int
    // Reading from nil map is OK
    v := m["key"]
    if v != 0 {
        t.Errorf("nil map read: got %d, want 0", v)
    }
}

func TestNilMapWrite(t *testing.T) {
    var m map[string]int
    
    defer func() {
        if r := recover(); r == nil {
            t.Error("expected panic on nil map write")
        }
    }()
    
    m["key"] = 1 // should panic
}

func TestMapConcurrency(t *testing.T) {
    // Test that sync.Map is safe for concurrent access
    var m sync.Map
    var wg sync.WaitGroup
    
    // Concurrent writers
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            key := fmt.Sprintf("key%d", n%10) // multiple goroutines write same keys
            m.Store(key, n)
        }(i)
    }
    
    wg.Wait()
    
    // Count stored items
    count := 0
    m.Range(func(_, _ interface{}) bool { count++; return true })
    
    if count > 10 {
        t.Errorf("expected at most 10 unique keys, got %d", count)
    }
}

func TestGroupBy(t *testing.T) {
    type Item struct{ Category, Name string }
    items := []Item{
        {"fruit", "apple"}, {"veggie", "carrot"},
        {"fruit", "banana"}, {"veggie", "pea"},
    }
    
    grouped := make(map[string][]Item)
    for _, item := range items {
        grouped[item.Category] = append(grouped[item.Category], item)
    }
    
    if len(grouped["fruit"]) != 2 {
        t.Errorf("expected 2 fruits, got %d", len(grouped["fruit"]))
    }
    if len(grouped["veggie"]) != 2 {
        t.Errorf("expected 2 veggies, got %d", len(grouped["veggie"]))
    }
}

func TestFrequencyCount(t *testing.T) {
    words := []string{"go", "is", "great", "go", "is", "go"}
    
    freq := make(map[string]int)
    for _, w := range words {
        freq[w]++
    }
    
    if freq["go"] != 3 { t.Errorf("go freq: got %d, want 3", freq["go"]) }
    if freq["is"] != 2 { t.Errorf("is freq: got %d, want 2", freq["is"]) }
    if freq["great"] != 1 { t.Errorf("great freq: got %d, want 1", freq["great"]) }
}

// ==================== BENCHMARKS ====================

func BenchmarkSliceAppendPreallocated(b *testing.B) {
    for i := 0; i < b.N; i++ {
        s := make([]int, 0, 1000)
        for j := 0; j < 1000; j++ {
            s = append(s, j)
        }
        _ = s
    }
}

func BenchmarkSliceAppendNoPrealloc(b *testing.B) {
    for i := 0; i < b.N; i++ {
        var s []int
        for j := 0; j < 1000; j++ {
            s = append(s, j)
        }
        _ = s
    }
}

func BenchmarkMapLookup(b *testing.B) {
    m := make(map[int]int, 1000)
    for i := 0; i < 1000; i++ {
        m[i] = i * 2
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = m[i%1000]
    }
}
```

---

**Summary of Part 6:**
- Slices are descriptors (pointer + len + cap) — 24 bytes regardless of size
- Subslicing creates a new descriptor but SHARES the underlying array
- `append` may allocate a new array — always assign its return value
- nil slice vs empty slice: both work with range/len/append; JSON encodes differently
- Full slice expression `[low:high:max]` limits capacity to prevent accidental sharing
- Maps are hash tables: O(1) average lookup, NOT safe for concurrent access
- Map iteration order is RANDOM by design — never depend on it
- Zero-value int in maps enables `freq[word]++` without initialization
- `sync.Map` for concurrent access, but prefer `sync.RWMutex` + regular map for most cases
- Race detector (`-race`) is your friend — always run it in CI
