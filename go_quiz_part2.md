# Go Quiz — Part 2 of 5 (Q101–Q200)
### Covers: Parts 5–8 (Pointers & Memory, Slices & Maps, Structs & Methods, Interfaces)
> Answer is hidden after `> ✅`. Try to answer before looking!

---

## PART 5 — Pointers & Memory (Q101–Q125)

**Q101.** What is the zero value of a pointer?
- a) `0`
- b) A pointer to the zero value of the type
- c) `nil`
- d) Invalid memory address

> ✅ **c** — All pointers zero to `nil` in Go.

---

**Q102.** True or False: Go supports pointer arithmetic like `ptr + 1` (without using `unsafe`).

> ✅ **False** — Pointer arithmetic is not allowed in Go outside of `unsafe.Pointer`. This is a deliberate safety choice.

---

**Q103.** What is the output?
```go
a := 10
b := &a
*b = 20
fmt.Println(a)
```
- a) `10`
- b) `20`
- c) Compile error
- d) Memory address

> ✅ **b** — `b` is a pointer to `a`. Writing `*b = 20` modifies `a` directly.

---

**Q104.** What does escape analysis determine?
- a) Whether a goroutine can be escaped from its context
- b) Whether a variable must be allocated on the heap (vs. stack)
- c) Whether an interface value escapes its type assertion
- d) Whether a channel message can be lost

> ✅ **b** — The compiler runs escape analysis to decide: if a variable's address is taken and used after the function returns, it must live on the heap.

---

**Q105.** Which function allocates a zeroed value and returns a pointer to it?
- a) `make`
- b) `alloc`
- c) `new`
- d) `init`

> ✅ **c** — `new(T)` allocates a zeroed `T` and returns `*T`. `make` is for slices, maps, and channels only.

---

**Q106.** True or False: All local variables in Go are stack-allocated.

> ✅ **False** — Variables that **escape** (e.g., their address is returned, they're captured by a closure, stored in interface) are heap-allocated. The compiler decides via escape analysis.

---

**Q107.** What is the output?
```go
func addOne(x int) { x++ }
n := 10
addOne(n)
fmt.Println(n)
```
- a) `10`
- b) `11`
- c) Compile error
- d) `0`

> ✅ **a** — Go passes by **value**. `addOne` gets a copy of `n`; the original is unchanged.

---

**Q108.** What is the output?
```go
func modify(s []int) { s[0] = 99 }
arr := []int{1, 2, 3}
modify(arr)
fmt.Println(arr[0])
```
- a) `1`
- b) `99`
- c) Compile error
- d) Runtime panic

> ✅ **b** — The slice header is copied, but the **underlying array is shared**. Modifying `s[0]` modifies `arr[0]`.

---

**Q109.** What does `go build -gcflags="-m"` show?
- a) Missing symbols
- b) Memory usage statistics
- c) Escape analysis: which variables move to the heap
- d) Global variable declarations

> ✅ **c** — Use this flag to audit allocations. Typical output: `main.go:5:2: moved to heap: x`.

---

**Q110.** True or False: Returning a pointer to a local variable from a Go function is safe (unlike C++).

> ✅ **True** — The variable escapes to the heap. The GC tracks it. In C++, this would be a dangling pointer.

---

**Q111.** What can the `unsafe` package do?
- a) Run untrusted user code safely
- b) Bypass Go's type safety — e.g., convert between unrelated pointer types, read struct offsets
- c) Access private fields in other packages without interfaces
- d) Both b and c

> ✅ **d** — `unsafe.Pointer` allows arbitrary pointer casts. `unsafe.Offsetof` reads struct field positions. Both bypass normal type safety.

---

**Q112.** How do you minimize struct size in Go?
- a) Use pointer fields instead of value fields
- b) Order fields from largest to smallest alignment (largest types first)
- c) Use `compact` struct tag
- d) Use `sync.Pool` for the struct

> ✅ **b** — Struct alignment padding is minimized by grouping fields of the same alignment together. `int64`(8-byte) before `int32`(4-byte) before `bool`(1-byte).

---

**Q113.** True or False: When you pass a slice to a function, the underlying array is shared between caller and callee.

> ✅ **True** — The slice header (pointer, len, cap) is copied, but the pointer points to the same backing array.

---

**Q114.** What does `GOGC=200` do?
- a) Enables the GC only every 200 seconds
- b) Triggers GC after heap grows to 200% of previous live heap (less frequent GC)
- c) Uses 200 goroutines for GC
- d) Limits heap to 200MB

> ✅ **b** — Higher `GOGC` = less frequent GC = higher memory usage but lower CPU overhead. Default is 100 (trigger at 2x heap).

---

**Q115.** What is the difference between `new(User)` and `&User{}`?
- a) `new` allocates on the stack; `&User{}` on the heap
- b) Both produce `*User` with zero-value fields; `&User{}` allows field initialization
- c) `new` is deprecated in Go 1.18+
- d) `&User{}` returns a value, not a pointer

> ✅ **b** — Both return `*User`. In practice, use `&User{field: val}` since it allows initializing fields. `new` is rarely used.

---

**Q116.** True or False: `sync.Pool` prevents the GC from ever collecting pooled objects.

> ✅ **False** — The GC **can** collect pooled objects at any time. `sync.Pool` is a **hint** to reuse, not a guarantee. Never put objects you need to keep in a pool.

---

**Q117.** What does "escape to the heap" mean?
- a) The variable causes a memory leak
- b) The runtime moves the variable from the stack to the heap because its lifetime exceeds the function's
- c) The variable becomes globally accessible
- d) The GC cannot collect the variable

> ✅ **b** — Escape analysis: if the compiler determines a variable's lifetime extends beyond the current function (e.g., its address is returned), it allocates on heap.

---

**Q118.** True or False: Go's garbage collector is stop-the-world and can pause your program for seconds.

> ✅ **False** — Go's GC runs **concurrently** with your program since Go 1.5. Stop-the-world pauses are sub-millisecond since Go 1.8.

---

**Q119.** Which comparison is valid in Go?
- a) `p1 == p2` where both are `*int` — compares addresses
- b) `*p1 == *p2` — compares pointed-to values
- c) `p1 == nil` — checks if pointer is nil
- d) All of the above

> ✅ **d** — All three pointer comparisons are valid.

---

**Q120.** True or False: Go maps are always heap-allocated, even if declared inside a function.

> ✅ **True** — Maps always live on the heap because they have a complex internal structure that the runtime manages.

---

**Q121.** What is the output?
```go
type T struct{ x int }
a := T{x: 1}
b := a
b.x = 99
fmt.Println(a.x, b.x)
```
- a) `99 99`
- b) `1 99`
- c) Compile error
- d) `1 1`

> ✅ **b** — Struct assignment in Go copies by VALUE. `b` is an independent copy; modifying `b.x` doesn't affect `a`.

---

**Q122.** Which of these does NOT cause a variable to escape to the heap?
- a) Returning a pointer to it: `return &x`
- b) Capturing it in a closure that lives longer
- c) Storing it in a local variable that's never used again
- d) Storing it in an interface value

> ✅ **c** — An unused local variable stays on the stack (and may even be optimized away). The others cause heap escape.

---

**Q123.** True or False: You can read from a nil map without panicking.

> ✅ **True** — Reading from a nil map returns the zero value of the value type. Only **writing** to a nil map panics.

---

**Q124.** How many bytes does this struct use (on 64-bit)?
```go
type S struct {
    a bool    // 1 byte
    b int64   // 8 bytes
    c bool    // 1 byte
}
```
- a) 10 bytes
- b) 24 bytes (with padding)
- c) 16 bytes
- d) 12 bytes

> ✅ **b** — `a` (1) + 7 padding + `b` (8) + `c` (1) + 7 padding = 24 bytes. Reorder to `b, a, c` to get 16 bytes.

---

**Q125.** True or False: `unsafe.Sizeof(T{})` returns the size including padding.

> ✅ **True** — `unsafe.Sizeof` includes alignment padding in the size.

---

## PART 6 — Slices & Maps (Q126–Q150)

**Q126.** What is the internal structure of a Go slice?
- a) Just a pointer to the backing array
- b) Pointer + length
- c) Pointer + length + capacity
- d) start index + end index + backing array

> ✅ **c** — A slice is a 3-field header: `(ptr, len, cap)`.

---

**Q127.** What is the output?
```go
a := []int{1, 2, 3, 4, 5}
b := a[1:3]
fmt.Println(len(b), cap(b))
```
- a) `2, 2`
- b) `2, 4`
- c) `2, 5`
- d) `3, 4`

> ✅ **b** — `b` starts at index 1. `len = 3-1 = 2`. `cap = len(a) - start = 5-1 = 4`.

---

**Q128.** True or False: Appending to a sub-slice can overwrite elements of the original slice.

> ✅ **True** — If the sub-slice still shares the backing array (cap allows), appending writes into the original array. Use `a[1:3:3]` (three-index) to limit capacity and prevent this.

---

**Q129.** What is the output?
```go
a := []int{1, 2, 3}
b := a
b = append(b, 4)
fmt.Println(len(a), cap(a))
```
- a) `4, 4`
- b) `3, 3`
- c) `3, 6`
- d) `4, 6`

> ✅ **b** — `a` had cap=3, so appending to `b` caused a new allocation. `a` is unchanged: len=3, cap=3.

---

**Q130.** What happens when you read a missing key from a Go map?
- a) Runtime panic
- b) Returns `nil`
- c) Returns the zero value of the value type
- d) Compile error

> ✅ **c** — `m["missingKey"]` returns `0` for `map[string]int`, `""` for `map[string]string`, etc.

---

**Q131.** True or False: Writing to a nil map panics at runtime.

> ✅ **True** — `var m map[string]int; m["key"] = 1` → panic: assignment to entry in nil map.

---

**Q132.** What is the output?
```go
m := map[string]int{}
m["x"]++
m["x"]++
fmt.Println(m["x"])
```
- a) Compile error
- b) `0`
- c) `2`
- d) Runtime panic

> ✅ **c** — Missing key returns `0`; `0+1=1`, `1+1=2`. The `++` on map values works this way.

---

**Q133.** True or False: Go's built-in `map` is safe for concurrent access from multiple goroutines.

> ✅ **False** — Concurrent map reads/writes cause a **runtime panic**: "concurrent map read and map write". Use `sync.Mutex` or `sync.Map`.

---

**Q134.** What does `copy(dst, src)` return?
- a) A new slice with copied elements
- b) The number of elements copied (`min(len(dst), len(src))`)
- c) Nothing
- d) A boolean indicating success

> ✅ **b** — `copy` returns the number of elements actually copied.

---

**Q135.** When does `append` NOT allocate a new backing array?
- a) Never — it always allocates
- b) When `len(s) < cap(s)` — there's capacity in the existing array
- c) When you provide a `make` hint
- d) When the element type is small (< 8 bytes)

> ✅ **b** — If there's room (`len < cap`), append writes into the existing array. Zero allocation!

---

**Q136.** What is the output?
```go
a := []int{1, 2, 3}
b := a
b[0] = 99
fmt.Println(a[0], b[0])
```
- a) `1 99`
- b) `99 99`
- c) Compile error
- d) `99 1`

> ✅ **b** — Slice assignment copies the header, NOT the backing array. Both `a` and `b` point to the same underlying array.

---

**Q137.** True or False: The iteration order of a Go map is guaranteed to be insertion order.

> ✅ **False** — Go deliberately randomizes map iteration order since Go 1.0.

---

**Q138.** True or False: `delete(m, key)` called on a key that doesn't exist panics.

> ✅ **False** — `delete` is safe on missing keys. It's a no-op.

---

**Q139.** What is the idiomatic way to check if a key exists in a map?
- a) `if m[k] != nil { }`
- b) `if val, ok := m[k]; ok { }`
- c) `if m.contains(k) { }`
- d) `if m[k] != zero { }`

> ✅ **b** — The comma-ok idiom. `ok` is `true` if the key exists, `false` if not (even if value is zero).

---

**Q140.** What does `make(map[string]int, 100)` do differently from `make(map[string]int)`?
- a) Limits the map to 100 entries
- b) Pre-allocates capacity for ~100 entries (hint to avoid rehashing)
- c) Creates 100 zero-value entries
- d) No difference

> ✅ **b** — The second argument is a size hint. It may reduce rehashing overhead but does NOT limit size.

---

**Q141.** What is the three-index slice `a[2:4:6]`?
- a) Slices from 2 to 6 (length 4)
- b) length = `4-2 = 2`, capacity = `6-2 = 4`
- c) length = `4-2 = 2`, capacity = `6`
- d) Compile error

> ✅ **b** — Three-index slice: `[low:high:max]`. len = high-low, cap = max-low.

---

**Q142.** True or False: `len(nil)` panics when called on a nil slice.

> ✅ **False** — `len(nil) == 0` and `cap(nil) == 0`. Nil slices are valid for these operations.

---

**Q143.** What does `append([]int{}, a...)` do?
- a) Appends an empty slice to `a`
- b) Creates a new independent copy of `a`
- c) Modifies `a` in place
- d) Returns nil

> ✅ **b** — This is the idiomatic way to shallow-copy a slice: new backing array, same values.

---

**Q144.** True or False: `json.Marshal([]string{})` returns `null`.

> ✅ **False** — Empty (non-nil) slice marshals to `[]`. Only a **nil** slice marshals to `null`.

---

**Q145.** What struct field order minimizes padding? Fields: `bool`, `int64`, `int32`, `bool`
- a) `bool, bool, int32, int64`
- b) Any order — Go auto-reorganizes
- c) `int64, int32, bool, bool`
- d) `bool, int64, int32, bool`

> ✅ **c** — Largest first: `int64`(8B) + `int32`(4B) + `bool`(1B) + `bool`(1B) + 2B padding = 16B. Option a: `bool`(1) + 1pad + ... is also good.

---

**Q146.** What is the output?
```go
s := make([]int, 3, 5)
fmt.Println(len(s), cap(s))
```
- a) `3, 3`
- b) `5, 5`
- c) `3, 5`
- d) `0, 5`

> ✅ **c** — `make([]T, len, cap)` — initialized with len=3 zeros, capacity=5.

---

**Q147.** Can you use a slice as a map key?
- a) Yes
- b) No — slices are not comparable (`==` is not defined for slices)
- c) Yes, but only `[]byte`
- d) Yes, in Go 1.21+

> ✅ **b** — Map keys must be **comparable** types. Slices, maps, and functions cannot be map keys.

---

**Q148.** What is the output?
```go
a := []int{1, 2, 3, 4, 5}
a = append(a[:2], a[3:]...)
fmt.Println(a)
```
- a) `[1 2 4 5]`
- b) `[1 2 3 4 5]`
- c) `[1 2]`
- d) Compile error

> ✅ **a** — This idiom removes element at index 2 (value 3). `a[:2]` = [1,2], `a[3:]` = [4,5], combined = [1,2,4,5].

---

**Q149.** True or False: `for k := range m { delete(m, k) }` is safe in Go.

> ✅ **True** — Go explicitly allows deleting from a map during range iteration. The spec guarantees this.

---

**Q150.** What is the output?
```go
m := map[string][]int{}
m["a"] = append(m["a"], 1)
m["a"] = append(m["a"], 2)
fmt.Println(m["a"])
```
- a) `[2]`
- b) `[1 2]`
- c) `[1] [2]`
- d) Runtime panic

> ✅ **b** — `m["a"]` returns nil slice on first call (zero value). `append` to nil slice is valid. Result: `[1 2]`.

---

## PART 7 — Structs & Methods (Q151–Q175)

**Q151.** True or False: In Go, a method is just a function with an explicit receiver.

> ✅ **True** — `func (u *User) Name() string` is syntactic sugar; the receiver is the first parameter.

---

**Q152.** When should you use a **pointer receiver** vs a **value receiver**?
- a) Always use pointer receivers — they're always faster
- b) Pointer receiver: when method must modify the struct or struct is large; value receiver: for read-only operations on small structs
- c) Value receiver for exported methods, pointer receiver for unexported
- d) It doesn't matter — Go automatically dereferences

> ✅ **b** — Pointer receiver allows mutation and avoids copying. Value receiver creates a copy. Be consistent across a type's methods.

---

**Q153.** What is the output?
```go
type Counter struct{ n int }
func (c Counter) Inc()  { c.n++ }  // value receiver
func (c *Counter) PInc() { c.n++ } // pointer receiver

ct := Counter{}
ct.Inc()
ct.PInc()
fmt.Println(ct.n)
```
- a) `0`
- b) `1`
- c) `2`
- d) Compile error

> ✅ **b** — `Inc()` operates on a copy (value receiver) — no effect on `ct`. `PInc()` modifies `ct` directly. Result: `ct.n = 1`.

---

**Q154.** What is struct embedding in Go?
- a) Including another struct's fields in a struct by name
- b) Including another struct's fields AND methods by declaring the type without a field name
- c) Inheritance — the embedded struct becomes the parent class
- d) A way to create a union type

> ✅ **b** — `type Dog struct { Animal }` embeds `Animal`. All fields/methods of `Animal` are **promoted** to `Dog` (accessible directly).

---

**Q155.** True or False: Struct embedding is the same as inheritance in C++.

> ✅ **False** — Embedding is composition (HAS-A), not inheritance (IS-A). A `Dog` that embeds `Animal` is NOT an `Animal` in Go's type system.

---

**Q156.** What is the output?
```go
type Animal struct{ Name string }
func (a Animal) Speak() string { return a.Name + " speaks" }

type Dog struct{ Animal }
func (d Dog) Speak() string { return d.Name + " says Woof" }

d := Dog{Animal: Animal{Name: "Rex"}}
fmt.Println(d.Speak())
fmt.Println(d.Animal.Speak())
```
- a) `Rex says Woof` and `Rex says Woof`
- b) `Rex says Woof` and `Rex speaks`
- c) Compile error — ambiguous Speak()
- d) `Rex speaks` and `Rex speaks`

> ✅ **b** — `d.Speak()` calls Dog's version. `d.Animal.Speak()` explicitly calls Animal's. Dog's method **shadows** Animal's.

---

**Q157.** What is the Functional Options pattern used for?
- a) Functional programming in Go
- b) Configuring objects with optional parameters in a flexible, extendable way
- c) Creating pure functions without side effects
- d) Option types like Rust's `Option<T>`

> ✅ **b** — `type Option func(*Config)` — each option is a function that modifies config. Allows adding new options without breaking existing callers.

---

**Q158.** True or False: You can define methods on a type from a different package.

> ✅ **False** — Methods must be defined in the same package as the type. You can use a wrapper/embedded struct instead.

---

**Q159.** What is the output?
```go
type Point struct{ X, Y int }
p1 := Point{1, 2}
p2 := Point{1, 2}
fmt.Println(p1 == p2)
```
- a) Compile error — structs can't be compared
- b) `false`
- c) `true`
- d) Runtime panic

> ✅ **c** — Structs with all comparable fields can be compared with `==`. All fields must match.

---

**Q160.** True or False: A struct with a slice field can be used as a map key.

> ✅ **False** — If any field is not comparable (slices, maps, functions), the struct is not comparable and cannot be a map key.

---

**Q161.** How do you access a promoted field from an embedded struct?
```go
type Engine struct{ HP int }
type Car struct{ Engine }
c := Car{Engine: Engine{HP: 200}}
```
- a) `c.Engine.HP` only
- b) `c.HP` only (promoted)
- c) Both `c.HP` and `c.Engine.HP` work
- d) Compile error

> ✅ **c** — Promoted fields are accessible both directly (`c.HP`) and via the embedded type (`c.Engine.HP`).

---

**Q162.** What is the `String() string` method used for?
- a) It's required for all exported types
- b) If a type implements `String() string`, it's called automatically by `fmt` functions
- c) It converts a struct to JSON
- d) It's part of the `Stringer` interface in `strconv`

> ✅ **b** — `fmt.Stringer` interface: `String() string`. If your type implements it, `fmt.Println(x)` uses your method.

---

**Q163.** True or False: You must explicitly declare that a type implements an interface in Go.

> ✅ **False** — Interface satisfaction is **implicit** (structural/duck typing). If you have the methods, you implement it.

---

**Q164.** What does `var _ io.Writer = (*MyWriter)(nil)` do?
- a) Creates a nil MyWriter variable
- b) Compile-time check that `*MyWriter` implements `io.Writer`; panics at runtime if not
- c) Compile-time check — fails at compile time if `*MyWriter` doesn't implement `io.Writer`
- d) Runtime type assertion

> ✅ **c** — This is the canonical compile-time interface check idiom. Zero cost at runtime (blank identifier discards).

---

**Q165.** What is the method set of type `T` (value, not pointer)?
- a) Methods with both value and pointer receivers
- b) Only methods with value receivers
- c) No methods — methods are only on `*T`
- d) Whatever is exported

> ✅ **b** — Value type `T` only has value receiver methods in its method set. Pointer type `*T` has both.

---

**Q166.** True or False: You can call a pointer receiver method on an addressable value in Go.

> ✅ **True** — Go automatically takes the address: `t.PtrMethod()` becomes `(&t).PtrMethod()` if `t` is addressable. But `T{}` (non-addressable) cannot do this.

---

**Q167.** What is the output?
```go
type S struct{ val int }
func (s S) Get() int  { return s.val }
func (s *S) Set(n int) { s.val = n }

var x interface{} = S{val: 5}
```
Does `x.(S).Set(10)` work?
- a) Yes — it calls Set on the concrete value
- b) No — `x.(S)` is not addressable, so can't call pointer receiver method

> ✅ **b** — A type-asserted interface value is not addressable. You need `*S` in the interface: `var x interface{} = &S{val: 5}`, then `x.(*S).Set(10)`.

---

**Q168.** True or False: Embedding an interface in a struct gives the struct all the interface's methods.

> ✅ **True** — But the method implementations will panic at runtime if not overridden (the embedded interface is nil). Used for partial interface implementation and mocking.

---

**Q169.** What does the `sync.Locker` interface require?
- a) `Lock()` only
- b) `Lock()` and `Unlock()`
- c) `Lock()`, `Unlock()`, and `TryLock()`
- d) Just `Acquire()`

> ✅ **b** — `sync.Locker` = `{ Lock(); Unlock() }`. Both `sync.Mutex` and `sync.RWMutex` implement it.

---

**Q170.** What is the output?
```go
type Animal struct{ Name string }
type Dog struct {
    Animal
    Breed string
}
d := Dog{Animal: Animal{Name: "Rex"}, Breed: "Lab"}
fmt.Println(d.Name, d.Breed)
```
- a) Compile error — must access via `d.Animal.Name`
- b) `Rex Lab`
- c) ` Lab` (Name is inaccessible)
- d) Runtime panic

> ✅ **b** — `d.Name` is promoted from the embedded `Animal`. Both `d.Name` and `d.Animal.Name` work.

---

**Q171.** What is a common use of struct embedding in the standard library?
- a) `http.Request` embeds `url.URL`
- b) `sync.Mutex` embedded in structs to make them lockable
- c) `io.Reader` embedded in `bufio.Reader`
- d) Both b and c

> ✅ **d** — Embedding `sync.Mutex` promotes `Lock/Unlock` to the outer struct. `bufio.Reader` embeds `io.Reader`.

---

**Q172.** True or False: Method sets affect which interfaces a type satisfies.

> ✅ **True** — `T` with only value receivers doesn't satisfy interfaces with pointer receiver methods. `*T` satisfies all of `T`'s interface requirements plus any `*T`-specific ones.

---

**Q173.** What is the output?
```go
type MyError struct{ Code int }
func (e *MyError) Error() string { return fmt.Sprintf("error %d", e.Code) }

var err error = &MyError{Code: 42}
fmt.Println(err)
```
- a) `&{42}`
- b) `error 42`
- c) `{42}`
- d) Compile error

> ✅ **b** — `*MyError` implements the `error` interface. `fmt` calls `Error()` which returns `"error 42"`.

---

**Q174.** What is a receiver name convention in Go?
- a) Use `self` like Python
- b) Use `this` like Java/C++
- c) Use a short abbreviation of the type name (1-2 letters), consistently across all methods
- d) Use the full type name

> ✅ **c** — Go convention: `func (u *User)`, `func (s *Server)`. Same receiver name for ALL methods of a type.

---

**Q175.** True or False: In Go, you can define a method on `int` directly.

> ✅ **False** — You can only define methods on types defined in the current package. To add methods to `int`, create a named type: `type MyInt int`.

---

## PART 8 — Interfaces (Q176–Q200)

**Q176.** What is the internal representation of an interface value?
- a) Just a pointer to the value
- b) A (type, value) pair — two words in memory
- c) A vtable pointer
- d) The concrete value itself

> ✅ **b** — An interface value is `(concrete_type_descriptor, concrete_value_or_pointer)`. Two machine words.

---

**Q177.** What is the "nil interface" bug?
```go
func check() error {
    var p *MyError = nil
    return p  // what does this return?
}
err := check()
fmt.Println(err == nil)
```
- a) `true`
- b) `false` — the error interface holds a non-nil type descriptor even though the value is nil
- c) Compile error
- d) Runtime panic

> ✅ **b** — The `error` interface value is NOT nil — its type field is `*MyError`. Only the value field is nil. `err == nil` is `false`.

---

**Q178.** True or False: A nil interface and an interface holding a nil pointer are the same.

> ✅ **False** — This is the famous nil interface bug. A nil interface = both type and value are nil. An interface holding a nil pointer = type is set, value is nil — NOT equal to nil interface.

---

**Q179.** What does a type assertion `v, ok := x.(Type)` do?
- a) Converts `x` to `Type` (may fail silently)
- b) Extracts the concrete value of type `Type` from interface `x`; `ok` is false if wrong type
- c) Creates a new variable of type `Type`
- d) Checks if `x` is comparable to `Type`

> ✅ **b** — Safe type assertion with comma-ok. Without `ok`, a wrong type causes a panic.

---

**Q180.** True or False: Large interfaces (many methods) are better than small interfaces in Go.

> ✅ **False** — Go advocates small, focused interfaces. "The bigger the interface, the weaker the abstraction." — Rob Pike. `io.Reader` (1 method) is more powerful than `io.ReadCloser` (2 methods) in terms of how many types satisfy it.

---

**Q181.** Which of these is NOT a valid Go interface?
- a) `type Writer interface { Write([]byte) (int, error) }`
- b) `type Empty interface {}`
- c) `type Stringer interface { String() (string, error) }`
- d) `type Doer interface { int | string }`

> ✅ **d** — Type union constraints (`int | string`) are only valid in **generic constraints** (`interface { int | string }`), not as regular interfaces for polymorphism.

---

**Q182.** What is the output?
```go
type Animal interface{ Speak() string }
type Dog struct{}
func (d Dog) Speak() string { return "Woof" }

var a Animal = Dog{}
fmt.Println(a.Speak())
```
- a) Compile error — Dog must explicitly implement Animal
- b) `Woof`
- c) `Dog.Speak()`
- d) Runtime panic

> ✅ **b** — Dog implicitly satisfies Animal (structural typing). Interface satisfaction is automatic.

---

**Q183.** What is `interface{}` (or `any`) useful for?
- a) Making functions that work on all types
- b) Bypassing the type system when you genuinely don't know the type (JSON, generics substitute before 1.18)
- c) Creating discriminated unions
- d) Both a and b

> ✅ **d** — `any` = no constraint. Useful for containers/utilities where you need to handle any type, but you lose type safety.

---

**Q184.** What is the difference between type assertion and type conversion?
- a) They are the same thing
- b) Type assertion: extract concrete type from interface. Type conversion: convert between compatible types (e.g., `int32` to `int64`)
- c) Type assertion is for primitives; conversion is for interfaces
- d) Type assertion is runtime; conversion is compile-time only

> ✅ **b** — `x.(T)` is an assertion (interface → concrete). `int64(x)` is a conversion (compatible types).

---

**Q185.** True or False: An interface can embed other interfaces.

> ✅ **True** — `type ReadWriter interface { Reader; Writer }` embeds both. A type must implement all methods of all embedded interfaces.

---

**Q186.** What is the output?
```go
type I interface{ M() }
type T struct{}
func (T) M() {}

var i I
fmt.Println(i == nil)
i = T{}
fmt.Println(i == nil)
```
- a) `true` then `false`
- b) `false` then `true`
- c) Compile error
- d) `true` then `true`

> ✅ **a** — Before assignment, `i` is a nil interface (both type and value nil). After assignment, it holds `(T, T{})` — not nil.

---

**Q187.** What is idiomatic Go for accepting any writer?
- a) `func Write(w *bytes.Buffer)`
- b) `func Write(w io.Writer)`
- c) `func Write(w interface{})`
- d) `func Write[W any](w W)`

> ✅ **b** — "Accept interfaces, return concrete types." `io.Writer` works for files, buffers, HTTP responses, etc.

---

**Q188.** True or False: `fmt.Println` accepts `...interface{}` (or `...any`) parameters.

> ✅ **True** — That's why it can print any type.

---

**Q189.** What is the purpose of an interface in Go's context of mocking for tests?
- a) Interfaces are not used for testing
- b) Define the interface for the dependency; inject a mock implementation in tests
- c) Use `reflect` to mock any concrete type
- d) Interfaces can't be mocked — use test doubles instead

> ✅ **b** — This is Go's primary testing strategy: `type UserStore interface { GetUser(id int) (*User, error) }` — production uses DB impl, tests use mock impl.

---

**Q190.** What is the output?
```go
var x interface{} = 42
switch v := x.(type) {
case int:
    fmt.Printf("int: %d\n", v)
case string:
    fmt.Printf("string: %s\n", v)
default:
    fmt.Println("unknown")
}
```
- a) `int: 42`
- b) `string: 42`
- c) `unknown`
- d) Compile error

> ✅ **a** — The type switch matches `int`. In the `case int:` branch, `v` is of type `int`.

---

**Q191.** True or False: A type can implement multiple interfaces simultaneously.

> ✅ **True** — A type implements all interfaces whose method sets are a subset of the type's method set. No explicit declaration needed.

---

**Q192.** What does `var _ io.Writer = (*os.File)(nil)` do?
- a) Creates a nil `*os.File` and discards it
- b) Compile-time assertion that `*os.File` implements `io.Writer`
- c) Runtime check that fails if `*os.File` doesn't implement `io.Writer`
- d) Marks `os.File` as unused

> ✅ **b** — This is the canonical compile-time interface check. Fails at compile time if the assertion is invalid.

---

**Q193.** Which interface does `error` satisfy?
- a) `fmt.Stringer` (has `String() string`)
- b) Any interface with `Error() string`
- c) Only the built-in `error` interface
- d) `fmt.Formatter`

> ✅ **b** — Any type with method `Error() string` satisfies the `error` interface, which is defined as `interface { Error() string }`.

---

**Q194.** True or False: You can assign a `*Dog` to a variable of type `Animal` (interface) if `Dog` implements `Animal`.

> ✅ **True** — If `Dog` has the methods (or `*Dog` does), the assignment is valid.

---

**Q195.** What is the difference between `x.(T)` and `x.(*T)`?
- a) First asserts to value type, second to pointer type — both valid
- b) `x.(T)` is wrong syntax; must always use pointer `x.(*T)`
- c) Same result — type assertions always return a pointer
- d) `x.(T)` panics if wrong; `x.(*T)` returns nil

> ✅ **a** — Both are valid type assertions. Use whichever concrete type was stored in the interface.

---

**Q196.** True or False: Calling a method on a nil interface value panics.

> ✅ **True** — A nil interface has no type descriptor and no method dispatch table. Calling any method panics.

---

**Q197.** What is the "accept interfaces, return structs" principle?
- a) Functions should return interface types for flexibility
- b) Function parameters should be interfaces (maximum flexibility); return concrete types (caller knows exactly what they get)
- c) Only exported functions should use interfaces
- d) It's the opposite — return interfaces, accept structs

> ✅ **b** — Returning interfaces forces callers to deal with interface indirection. Returning concrete types is more transparent and efficient.

---

**Q198.** What happens with this code?
```go
type I interface{ M() }
type T struct{}
func (t *T) M() {}

var t T
var i I = t  // ← this line
```
- a) Compiles fine
- b) Compile error — T (value) doesn't implement I because M() has pointer receiver

> ✅ **b** — `T` (value)'s method set only includes value receiver methods. `M()` has pointer receiver, so only `*T` implements `I`. Fix: `var i I = &t`.

---

**Q199.** True or False: `interface{}` and `any` are completely interchangeable in Go 1.18+.

> ✅ **True** — `any` is declared as `type any = interface{}` in the builtin package. They are identical.

---

**Q200.** What is a mock in Go testing?
- a) A concrete type that simulates a dependency for testing — implements the same interface
- b) A copy of the production code used for benchmarking
- c) A test that mocks time.Sleep to run faster
- d) An external mocking library (required to write tests)

> ✅ **a** — Mocks implement interfaces. No library required — just write a struct with the interface methods and configurable behavior.

---

*End of Quiz Part 2 (Q101–Q200)*
*Continue with go_quiz_part3.md for Q201–Q300*
