# Go Deep Dive — Part 14: Generics (Go 1.18+) — Type Parameters

---

## Chapter 1: Why Generics Were Added

Before Go 1.18, writing generic code required one of three ugly choices:
1. Use `interface{}` — lost type safety, required type assertions everywhere
2. Code generation (like `go:generate` with `genny`) — fragile, bloated output
3. Write the same function for every type — massive code duplication

```go
// Pre-generics — write for every type:
func SumInts(nums []int) int       { /* ... */ }
func SumFloat64s(nums []float64) float64 { /* ... */ }
func SumInt64s(nums []int64) int64 { /* ... */ }
// ...

// Pre-generics with interface{} — type-unsafe:
func Sum(nums []interface{}) interface{} {
    // Which type is it? int? float64? Have to assert...
    // And: cannot do nums[0] + nums[1] without knowing the type!
}

// Post Go 1.18 — one function works for all:
func Sum[T int | float64 | int64](nums []T) T {
    var total T
    for _, n := range nums { total += n }
    return total
}
```

---

## Chapter 2: Type Parameters — Syntax and Semantics

```go
// Single type parameter
func Map[T any, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

// Two type parameters
func Filter[T any](slice []T, fn func(T) bool) []T {
    var result []T
    for _, v := range slice {
        if fn(v) {
            result = append(result, v)
        }
    }
    return result
}

// With constraint
func Max[T Ordered](a, b T) T {
    if a > b { return a }
    return b
}

// Type inference — compiler often infers type parameters:
nums := []int{1, 2, 3, 4, 5}
doubled := Map(nums, func(n int) int { return n * 2 })
// Map[int, int] — type params inferred from arguments!

// Sometimes you must be explicit:
result := Map[string, int](words, func(s string) int { return len(s) })
```

---

## Chapter 3: Type Constraints — Complete Reference

### Built-in Constraints

```go
import "golang.org/x/exp/constraints"  // or define your own

// 'any' — accepts every type (alias for interface{})
func PrintAll[T any](items []T) {
    for _, item := range items { fmt.Println(item) }
}

// 'comparable' — types that support == and != (for maps, sets)
func Contains[T comparable](slice []T, target T) bool {
    for _, v := range slice {
        if v == target { return true }
    }
    return false
}

// Union constraints — type MUST be one of these
type Integer interface {
    int | int8 | int16 | int32 | int64 |
    uint | uint8 | uint16 | uint32 | uint64
}

type Float interface {
    float32 | float64
}

type Number interface {
    Integer | Float  // combine with |
}

// Ordered — types that support < > <= >= (the most common constraint)
type Ordered interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
    ~float32 | ~float64 | ~string
}
// This is what golang.org/x/exp/constraints.Ordered looks like
```

### The `~` Tilde — Underlying Type

```go
// Without ~: only the exact type matches
type Integer interface { int }

type MyInt int
var x MyInt = 5

// func Sum[T Integer](nums []T) T  →  Sum([]MyInt{1,2,3}) FAILS!
// MyInt is NOT int (even though underlying type is int)

// With ~: any type whose underlying type is int matches
type Integer interface { ~int }

// Sum[T Integer](nums []T) T  →  Sum([]MyInt{1,2,3}) WORKS!
// Because ~int includes int AND all types with int as underlying type

// Practical example:
type Celsius float64
type Kelvin float64

type Temperature interface { ~float64 }

func Average[T Temperature](temps []T) T {
    var sum T
    for _, t := range temps { sum += t }
    return sum / T(len(temps))
}

avgC := Average([]Celsius{36.5, 37.0, 36.8})   // works!
avgK := Average([]Kelvin{309.65, 310.15, 309.95}) // works!
```

### Interface Constraints with Methods

```go
// Constraints can combine method requirements with type unions:
type Stringer interface {
    String() string  // must have String method
}

// Or combine type sets with methods:
type Printable interface {
    ~int | ~float64 | ~string  // allowed types
    fmt.Stringer               // must implement Stringer
}
// Note: A basic int literal doesn't satisfy Printable (no String method)
// You'd need a named type based on int that has String()
```

---

## Chapter 4: Generic Data Structures

### Generic Stack

```go
package stack

type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T          // zero value of T
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

func (s *Stack[T]) Size() int { return len(s.items) }
func (s *Stack[T]) IsEmpty() bool { return len(s.items) == 0 }

// Usage:
intStack := &Stack[int]{}
intStack.Push(1); intStack.Push(2); intStack.Push(3)
val, ok := intStack.Pop()  // 3, true

strStack := &Stack[string]{}
strStack.Push("hello"); strStack.Push("world")
```

### Generic Queue

```go
type Queue[T any] struct {
    items []T
}

func (q *Queue[T]) Enqueue(item T) {
    q.items = append(q.items, item)
}

func (q *Queue[T]) Dequeue() (T, bool) {
    if len(q.items) == 0 {
        var zero T
        return zero, false
    }
    item := q.items[0]
    q.items = q.items[1:]
    return item, true
}

func (q *Queue[T]) Size() int { return len(q.items) }
```

### Generic Set

```go
type Set[T comparable] struct {
    items map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
    return &Set[T]{items: make(map[T]struct{})}
}

func (s *Set[T]) Add(item T) {
    s.items[item] = struct{}{}
}

func (s *Set[T]) Remove(item T) {
    delete(s.items, item)
}

func (s *Set[T]) Contains(item T) bool {
    _, ok := s.items[item]
    return ok
}

func (s *Set[T]) Size() int { return len(s.items) }

func (s *Set[T]) ToSlice() []T {
    result := make([]T, 0, len(s.items))
    for k := range s.items {
        result = append(result, k)
    }
    return result
}

func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
    result := NewSet[T]()
    for k := range s.items {
        if other.Contains(k) {
            result.Add(k)
        }
    }
    return result
}

func (s *Set[T]) Union(other *Set[T]) *Set[T] {
    result := NewSet[T]()
    for k := range s.items { result.Add(k) }
    for k := range other.items { result.Add(k) }
    return result
}
```

### Generic Result Type (Option/Result pattern)

```go
// Similar to Rust's Result<T, E> or Haskell's Either
type Result[T any] struct {
    value T
    err   error
}

func OK[T any](value T) Result[T] {
    return Result[T]{value: value}
}

func Err[T any](err error) Result[T] {
    return Result[T]{err: err}
}

func (r Result[T]) IsOK() bool { return r.err == nil }
func (r Result[T]) Unwrap() T {
    if r.err != nil { panic(r.err) }
    return r.value
}
func (r Result[T]) UnwrapOr(defaultVal T) T {
    if r.err != nil { return defaultVal }
    return r.value
}
func (r Result[T]) Error() error { return r.err }

// Usage:
func divide(a, b float64) Result[float64] {
    if b == 0 { return Err[float64](errors.New("division by zero")) }
    return OK(a / b)
}

r := divide(10, 3)
if r.IsOK() {
    fmt.Println(r.Unwrap())   // 3.333...
}
fmt.Println(r.UnwrapOr(0.0)) // safe default
```

---

## Chapter 5: Generic Functions — Standard Patterns

```go
// Map: transform each element
func Map[T, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice { result[i] = fn(v) }
    return result
}

// Filter: keep elements matching predicate
func Filter[T any](slice []T, fn func(T) bool) []T {
    var result []T
    for _, v := range slice {
        if fn(v) { result = append(result, v) }
    }
    return result
}

// Reduce: aggregate to single value
func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
    result := initial
    for _, v := range slice { result = fn(result, v) }
    return result
}

// Count: count elements matching predicate
func Count[T any](slice []T, fn func(T) bool) int {
    count := 0
    for _, v := range slice {
        if fn(v) { count++ }
    }
    return count
}

// Any/All: check predicate against slice
func Any[T any](slice []T, fn func(T) bool) bool {
    for _, v := range slice { if fn(v) { return true } }
    return false
}

func All[T any](slice []T, fn func(T) bool) bool {
    for _, v := range slice { if !fn(v) { return false } }
    return true
}

// Keys/Values from map
func Keys[K comparable, V any](m map[K]V) []K {
    keys := make([]K, 0, len(m))
    for k := range m { keys = append(keys, k) }
    return keys
}

func Values[K comparable, V any](m map[K]V) []V {
    vals := make([]V, 0, len(m))
    for _, v := range m { vals = append(vals, v) }
    return vals
}

// GroupBy: group slice elements by key
func GroupBy[T any, K comparable](slice []T, key func(T) K) map[K][]T {
    result := make(map[K][]T)
    for _, v := range slice {
        k := key(v)
        result[k] = append(result[k], v)
    }
    return result
}

// Usage examples:
nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

squares := Map(nums, func(n int) int { return n * n })
evens := Filter(nums, func(n int) bool { return n%2 == 0 })
sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })

people := []struct{ Name string; Dept string }{
    {"Alice", "Eng"}, {"Bob", "Sales"}, {"Carol", "Eng"},
}
byDept := GroupBy(people, func(p struct{ Name, Dept string }) string {
    return p.Dept
})
// {"Eng": [Alice, Carol], "Sales": [Bob]}
```

---

## Chapter 6: Generics Limitations

```go
// 1. Cannot use type parameters in type switches
func process[T any](v T) {
    switch v.(type) {  // COMPILE ERROR: cannot use type switch on type parameter
    case int:
    case string:
    }
    
    // Workaround: use any
    switch any(v).(type) {
    case int:
    case string:
    }
}

// 2. Cannot have generic methods (only generic types)
type MyType struct{}

// COMPILE ERROR: method cannot have type parameters
// func (m MyType) Process[T any](v T) {}

// Workaround: use generic function
func Process[T any](m MyType, v T) {}

// 3. No specialization — you can't have different implementations per type
// Unlike C++ templates, Go generics use a single implementation
// for all types (with interface-boxing under the hood)

// 4. Inference limitations — sometimes you must be explicit
func Pair[T, U any](t T, u U) (T, U) { return t, u }
// a, b := Pair(1, "hello")  // works: inferred [int, string]
// a, b := Pair[int, string](1, "hello")  // explicit
```

---

## Chapter 7: When to Use Generics — Guidelines

```go
// USE generics when:
// 1. Writing container/collection data structures (Stack, Queue, Set, Tree)
// 2. Writing functions that work on slices of any type (Map, Filter, Reduce)
// 3. Writing functions that work on maps of any type (Keys, Values, GroupBy)
// 4. Type-safe wrappers for existing type-unsafe code

// DON'T use generics when:
// 1. Interfaces already solve the problem — prefer interfaces!
//    If all you need is a method call, use an interface.
// 2. Only one or two types are involved — just write two functions!
// 3. The function body is different per type — use interface or separate functions

// FAMOUS QUOTE from Go team:
// "Write code, don't design types." — if you're spending time on a complex
// generic type hierarchy, you might be over-engineering.

// Example: prefer interface over generics when you have behavior
type Sizer interface { Size() int }  // works for any type with Size()

// vs. generics (unnecessary if you just need Size()):
func PrintSize[T Sizer](v T) { fmt.Println(v.Size()) }  // overkill
func PrintSize(v Sizer) { fmt.Println(v.Size()) }         // simpler!
```

---

## Chapter 8: Comprehensive Generics Tests

```go
package generics_test

import (
    "testing"
    "reflect"
    "errors"
)

// Tests for Map
func TestMap(t *testing.T) {
    tests := []struct {
        name  string
        input []int
        fn    func(int) string
        want  []string
    }{
        {
            "int to string",
            []int{1, 2, 3},
            func(n int) string { return fmt.Sprintf("%d", n) },
            []string{"1", "2", "3"},
        },
        {
            "empty input",
            []int{},
            func(n int) string { return "x" },
            []string{},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Map(tt.input, tt.fn)
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Map() = %v, want %v", got, tt.want)
            }
        })
    }
}

// Tests for Filter
func TestFilter(t *testing.T) {
    nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    
    evens := Filter(nums, func(n int) bool { return n%2 == 0 })
    if !reflect.DeepEqual(evens, []int{2, 4, 6, 8, 10}) {
        t.Errorf("Filter evens: got %v", evens)
    }
    
    // Filter all out
    none := Filter(nums, func(int) bool { return false })
    if len(none) != 0 { t.Errorf("Filter none: got %v", none) }
    
    // Filter all in
    all := Filter(nums, func(int) bool { return true })
    if len(all) != 10 { t.Errorf("Filter all: got %v", all) }
}

// Tests for Reduce
func TestReduce(t *testing.T) {
    nums := []int{1, 2, 3, 4, 5}
    
    sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
    if sum != 15 { t.Errorf("sum = %d, want 15", sum) }
    
    product := Reduce(nums, 1, func(acc, n int) int { return acc * n })
    if product != 120 { t.Errorf("product = %d, want 120", product) }
    
    // Empty slice returns initial value
    empty := Reduce([]int{}, 42, func(acc, n int) int { return acc + n })
    if empty != 42 { t.Errorf("empty reduce: got %d, want 42", empty) }
}

// Tests for Stack
func TestStack(t *testing.T) {
    s := &Stack[int]{}
    
    if !s.IsEmpty() { t.Error("new stack should be empty") }
    
    // Push and peek
    s.Push(1); s.Push(2); s.Push(3)
    if s.Size() != 3 { t.Errorf("size = %d, want 3", s.Size()) }
    
    top, ok := s.Peek()
    if !ok || top != 3 { t.Errorf("Peek() = %d,%v, want 3,true", top, ok) }
    if s.Size() != 3 { t.Error("Peek should not remove element") }
    
    // Pop in LIFO order
    for _, want := range []int{3, 2, 1} {
        got, ok := s.Pop()
        if !ok || got != want {
            t.Errorf("Pop() = %d,%v, want %d,true", got, ok, want)
        }
    }
    
    // Pop from empty stack
    _, ok = s.Pop()
    if ok { t.Error("Pop from empty stack should return false") }
}

// Tests for Set
func TestSet(t *testing.T) {
    s := NewSet[string]()
    
    s.Add("apple"); s.Add("banana"); s.Add("cherry")
    s.Add("apple")  // duplicate — should not increase size
    
    if s.Size() != 3 { t.Errorf("size = %d, want 3", s.Size()) }
    if !s.Contains("apple") { t.Error("should contain apple") }
    if s.Contains("grape") { t.Error("should not contain grape") }
    
    s.Remove("banana")
    if s.Contains("banana") { t.Error("should not contain banana after remove") }
    if s.Size() != 2 { t.Errorf("size after remove = %d, want 2", s.Size()) }
    
    // Intersection
    s2 := NewSet[string]()
    s2.Add("apple"); s2.Add("grape"); s2.Add("cherry")
    
    inter := s.Intersection(s2)
    if !inter.Contains("apple") || !inter.Contains("cherry") {
        t.Error("intersection should contain apple and cherry")
    }
    if inter.Contains("grape") { t.Error("intersection should not contain grape") }
    if inter.Size() != 2 { t.Errorf("intersection size = %d, want 2", inter.Size()) }
}

// Tests for Result type
func TestResult(t *testing.T) {
    // OK result
    r := OK(42)
    if !r.IsOK() { t.Error("OK result should be OK") }
    if r.Unwrap() != 42 { t.Errorf("Unwrap = %d, want 42", r.Unwrap()) }
    if r.Error() != nil { t.Error("OK result should have no error") }
    
    // Err result
    myErr := errors.New("something failed")
    r2 := Err[int](myErr)
    if r2.IsOK() { t.Error("Err result should not be OK") }
    if !errors.Is(r2.Error(), myErr) { t.Error("Err result should carry the error") }
    if r2.UnwrapOr(99) != 99 { t.Error("UnwrapOr should return default for err result") }
    
    // Panic on Unwrap of err result
    defer func() {
        if r := recover(); r == nil {
            t.Error("Unwrap on err result should panic")
        }
    }()
    r2.Unwrap()  // should panic
}

// Benchmark generics vs interface{}
func BenchmarkGenericMap(b *testing.B) {
    input := make([]int, 1000)
    for i := range input { input[i] = i }
    
    b.Run("generic", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = Map(input, func(n int) int { return n * 2 })
        }
    })
    
    b.Run("interface{}", func(b *testing.B) {
        iface := make([]interface{}, len(input))
        for i, v := range input { iface[i] = v }
        
        for i := 0; i < b.N; i++ {
            result := make([]interface{}, len(iface))
            for j, v := range iface {
                result[j] = v.(int) * 2
            }
            _ = result
        }
    })
}
```

---

**Summary of Part 14:**
- Generics solve code duplication for type-safe containers and utility functions
- Type parameters use `[T constraint]` syntax in function/type declarations
- `any` = no constraint (any type), `comparable` = supports == (for maps/sets)
- `~T` means "any type whose underlying type is T" — enables named types like `type Celsius float64`
- Generic data structures: Stack, Queue, Set, Result are the classics
- Generic utility functions: Map, Filter, Reduce, GroupBy are the most useful
- Type inference usually works — only explicit when inference fails
- Cannot use type switch on type parameter directly — use `any(v).(type)` workaround
- Methods cannot have their own type parameters — use generic functions instead
- Prefer interfaces over generics when the operations are behavior-based (not type-structural)
