# Go Quiz — Part 4 of 5 (Q301–Q400)
### Covers: Parts 13–16 (Context, Generics, Testing, Standard Library)
> Answer is hidden after `> ✅`. Try to answer before looking!

---

## PART 13 — context.Context (Q301–Q325)

**Q301.** What is `context.Background()` used for?

- a) It's the context for background goroutines only
- b) The top-level, never-cancelled context — used at the start of a request chain or in main/tests
- c) A context that automatically times out after 30 seconds
- d) A context that does nothing

> ✅ **b** — `context.Background()` is the root of all context trees. Never cancelled, no deadline, no values.

---

**Q302.** True or False: Storing a `context.Context` in a struct is idiomatic Go.

> ✅ **False** — The Go docs explicitly say: "Do not store Contexts inside a struct type; instead, pass a Context explicitly to each function." Context is per-call, not per-object.

---

**Q303.** What does `context.WithCancel(parent)` return?

- a) `(ctx context.Context, err error)`
- b) `(ctx context.Context, cancel context.CancelFunc)`
- c) `(ctx context.Context)`
- d) `(cancel func(), ctx context.Context)`

> ✅ **b** — Always call `defer cancel()` immediately. Not calling cancel leaks context resources.

---

**Q304.** When is `ctx.Done()` closed?

- a) When the function that created the context returns
- b) When `cancel()` is called, or the deadline/timeout expires
- c) Only when `context.Background()` is GC'd
- d) When the last goroutine using it finishes

> ✅ **b** — `Done()` is a channel. It closes when the context is cancelled (via `cancel()`) or deadline expires.

---

**Q305.** What is the output conceptually?
```go
ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()

select {
case <-time.After(100 * time.Millisecond):
    fmt.Println("timer fired")
case <-ctx.Done():
    fmt.Println("context done:", ctx.Err())
}
```
- a) "timer fired"
- b) "context done: context deadline exceeded"
- c) Blocks forever
- d) Compile error

> ✅ **b** — The context times out after 50ms, before the 100ms timer. `ctx.Err()` returns `context.DeadlineExceeded`.

---

**Q306.** True or False: Cancelling a parent context also cancels all child contexts derived from it.

> ✅ **True** — Cancellation propagates DOWN the context tree. Parent cancel → all children cancelled.

---

**Q307.** What does `context.WithValue(ctx, key, value)` do?

- a) Adds a key-value pair accessible to ALL goroutines
- b) Attaches a value to the context propagatable to downstream functions via `ctx.Value(key)`
- c) Stores values in a global registry
- d) Creates a new independent context with the value

> ✅ **b** — Values travel with the context down the call chain. Used for request-scoped data (requestID, user, trace).

---

**Q308.** True or False: Using `string` as a context key is recommended.

> ✅ **False** — Using a custom unexported type as a key prevents key collisions between packages: `type ctxKey string; const userKey ctxKey = "user"`.

---

**Q309.** What is the difference between `context.WithTimeout` and `context.WithDeadline`?

- a) WithTimeout takes a duration; WithDeadline takes an absolute time.Time
- b) WithDeadline is deprecated in favor of WithTimeout
- c) They are identical
- d) WithTimeout is more accurate

> ✅ **a** — `WithTimeout(ctx, 5*time.Second)` is syntactic sugar for `WithDeadline(ctx, time.Now().Add(5*time.Second))`.

---

**Q310.** True or False: `ctx.Err()` returns `nil` if the context has NOT been cancelled.

> ✅ **True** — `ctx.Err()` = nil (active), `context.Canceled` (cancelled), or `context.DeadlineExceeded` (timed out).

---

**Q311.** What should you do if a function receives a cancelled context?

- a) Ignore it and complete the work anyway
- b) Return immediately with `ctx.Err()` as the error
- c) Create a new context to continue
- d) Panic

> ✅ **b** — Honour cancellation. `if ctx.Err() != nil { return ctx.Err() }` — stop doing work immediately.

---

**Q312.** What is the correct first parameter in an HTTP handler in Go 1.22+?

- a) `(r *http.Request, w http.ResponseWriter)`
- b) `(w http.ResponseWriter, r *http.Request)`
- c) `(ctx context.Context, w http.ResponseWriter, r *http.Request)`
- d) `(r http.Request, w http.ResponseWriter)`

> ✅ **b** — Always `(w http.ResponseWriter, r *http.Request)`. Context is accessed via `r.Context()`.

---

**Q313.** True or False: Cancellation only propagates from parent to child — cancelling a child does NOT cancel the parent.

> ✅ **True** — Context cancellation is one-directional: parent → children. Child cancellation is local.

---

**Q314.** What is the purpose of `context.TODO()`?

- a) Marks context-related code as a TODO item for cleanup
- b) A placeholder when you need a context but haven't decided which one to use (testing, initial migration)
- c) A context that executes work in a background goroutine
- d) The same as `context.Background()` in all ways

> ✅ **b** — `TODO()` and `Background()` are identical at runtime, but `TODO()` signals to tools/readers that this needs to be replaced with a real context.

---

**Q315.** What does this check?
```go
select {
case <-ctx.Done():
    return ctx.Err()
default:
}
```
- a) Blocks until context is cancelled
- b) Non-blocking check: if context is already cancelled, return immediately; otherwise continue
- c) Always returns `ctx.Err()`
- d) Compile error in `select`

> ✅ **b** — The `default` case makes this non-blocking. Used to poll cancellation in long loops.

---

**Q316.** True or False: You should always call `defer cancel()` even for `context.WithValue` (which has no cancel).

> ✅ **False** — `context.WithValue` does NOT return a cancel function. Only `WithCancel`, `WithTimeout`, and `WithDeadline` return cancel functions.

---

**Q317.** What is request-scoped context value?

- a) A value that lives only as long as the HTTP request
- b) A global variable
- c) A value stored in the database per request
- d) A session cookie

> ✅ **a** — Values like `requestID`, `userID`, `traceSpan` are stored in context and passed along for the lifetime of one request.

---

**Q318.** What is wrong with this code?
```go
type key string  // exported key type
const UserKey key = "user"
ctx = context.WithValue(ctx, UserKey, user)
```
- a) No problem
- b) The key type `key` is exported — other packages can use the same key type and collide
- c) `key` should be a struct type, not string
- d) `context.WithValue` doesn't accept custom types

> ✅ **b** — Key types should be **unexported** to prevent other packages from accessing or colliding.

---

**Q319.** True or False: `context.WithCancel` returns a function that must be called to free resources even if the parent is cancelled first.

> ✅ **True** — Always `defer cancel()`. The cancel function is idempotent — safe to call multiple times.

---

**Q320.** Which function should all long-running goroutines accept as their first parameter?

- a) `*sync.WaitGroup`
- b) `context.Context`
- c) `chan struct{}`
- d) `*log.Logger`

> ✅ **b** — `func worker(ctx context.Context, ...)` — Context first, following the go.dev convention.

---

**Q321.** What is the error returned by `ctx.Err()` when the deadline has passed?

- a) `context.Canceled`
- b) `context.DeadlineExceeded`
- c) `io.ErrDeadlineExceeded`
- d) `errors.New("deadline exceeded")`

> ✅ **b** — `context.DeadlineExceeded` (for timeout/deadline). `context.Canceled` for explicit cancellation via `cancel()`.

---

**Q322.** True or False: An HTTP request's context is automatically cancelled when the client disconnects.

> ✅ **True** — Go's HTTP server automatically cancels `r.Context()` when the client disconnects, allowing handlers to stop work early.

---

**Q323.** What is `errgroup.WithContext` best used for?

- a) Running a single function that might error
- b) Running multiple goroutines, any one of which can cancel all others via shared context
- c) Collecting multiple errors from a single goroutine
- d) A replacement for `sync.WaitGroup`

> ✅ **b** — `errgroup.Group` + context: if any goroutine returns an error, the context is cancelled, signalling all others to stop. `Wait()` returns the first non-nil error.

---

**Q324.** True or False: Context values should be used for optional, cross-cutting concerns (logging, tracing, auth), not for required function arguments.

> ✅ **True** — Required data (like function parameters) should be explicit parameters. Context values are for metadata that travels with the request.

---

**Q325.** What does `ctx.Value(key)` return if the key doesn't exist?

- a) An error
- b) `nil`
- c) The zero value of whatever type was stored
- d) Panics

> ✅ **b** — Returns `nil` if not found. Always type-assert the result. Check for nil before asserting.

---

## PART 14 — Generics (Q326–Q350)

**Q326.** What Go version introduced generics?

- a) Go 1.15
- b) Go 1.18
- c) Go 1.20
- d) Go 1.21

> ✅ **b** — Go 1.18 (March 2022) introduced generics with type parameters.

---

**Q327.** What is a type parameter in Go generics?

- a) A parameter that accepts a specific type at compile time
- b) A placeholder type in square brackets that gets replaced with a concrete type at instantiation
- c) A runtime type switch optimization
- d) An interface type parameter

> ✅ **b** — `func Min[T constraints.Ordered](a, b T) T` — `T` is the type parameter, filled in at call site.

---

**Q328.** What does `[T any]` mean in a generic function?

- a) `T` must be an interface
- b) `T` can be any type whatsoever — no constraint on operations
- c) `T` is required to be comparable
- d) `T` accepts any non-nil type

> ✅ **b** — `any` = `interface{}` as a constraint. No methods or operations are guaranteed; you can only pass/return values.

---

**Q329.** True or False: Generic functions in Go support specialization — you can provide a different implementation for specific types.

> ✅ **False** — Go generics do NOT support template specialization (unlike C++). One implementation for all types that satisfy the constraint.

---

**Q330.** What does `~int` mean in a type constraint?

- a) Approximately int in size
- b) An unsigned integer
- c) Any type whose **underlying type** is `int` (includes named types like `type MyInt int`)
- d) A bitwise NOT of int

> ✅ **c** — The tilde `~` means "underlying type". `~int` matches `int`, `type Count int`, `type ID int`, etc.

---

**Q331.** What is the output?
```go
func Map[T, U any](s []T, f func(T) U) []U {
    result := make([]U, len(s))
    for i, v := range s { result[i] = f(v) }
    return result
}
doubled := Map([]int{1, 2, 3}, func(n int) int { return n * 2 })
fmt.Println(doubled)
```
- a) `[1 2 3]`
- b) `[2 4 6]`
- c) Compile error
- d) `[2 4 6 0]`

> ✅ **b** — Generic `Map` applies `f` to each element. `n*2` for 1,2,3 gives 2,4,6.

---

**Q332.** True or False: You can use `==` inside a generic function with `[T any]` constraint.

> ✅ **False** — `any` doesn't include comparability. Use `[T comparable]` to allow `==` and `!=`.

---

**Q333.** What is `constraints.Ordered`?

- a) A custom sort order interface
- b) A constraint from `golang.org/x/exp/constraints` now in `slices`/`cmp` packages — types that support `<`, `>`, `<=`, `>=`
- c) An interface for ordered collections
- d) A constraint for integer types only

> ✅ **b** — `Ordered` = ordered scalar types: all integers, floats, and strings.

---

**Q334.** What is the syntax to instantiate a generic function explicitly?
```go
func Identity[T any](v T) T { return v }
```
- a) `Identity<int>(42)`
- b) `Identity[int](42)`
- c) `Identity.(int)(42)`
- d) Both compile — Go accepts both syntaxes

> ✅ **b** — Go uses `[T]` syntax, not `<T>` like C++/Java.

---

**Q335.** True or False: A generic type can have methods with additional type parameters beyond the type's type parameters.

> ✅ **False** — Methods of generic types cannot introduce NEW type parameters. The method can use the type's existing type parameters only.

---

**Q336.** What is a type set?

- a) A set of unique values of a type
- b) The set of types that satisfy an interface constraint in generics
- c) A collection of generic type parameters
- d) A compile-time set data structure

> ✅ **b** — In generics, `interface { int | float64 | string }` defines a type set. Any type in the set satisfies the constraint.

---

**Q337.** What does this generic stack implement?
```go
type Stack[T any] struct { items []T }
func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }
func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 { var zero T; return zero, false }
    top := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return top, true
}
```
- a) A compile error — methods on generic types are invalid
- b) A type-safe LIFO stack that works for any type T
- c) A stack limited to comparable types
- d) An interface

> ✅ **b** — Valid generic data structure. `Stack[int]`, `Stack[string]`, `Stack[*User]` all work.

---

**Q338.** How do you get the zero value of a type parameter `T` inside a generic function?

- a) `T{}`
- b) `nil`
- c) `var zero T; return zero`
- d) `reflect.Zero(reflect.TypeOf((*T)(nil)).Elem())`

> ✅ **c** — `var zero T` is always the zero value for any type T, including primitives, pointers, structs, etc.

---

**Q339.** True or False: Before generics (pre-1.18), the most common workaround was using `interface{}` with type assertions everywhere.

> ✅ **True** — `interface{}` + type switch/assertion was the pre-generics way to write generic-ish code. Less type-safe.

---

**Q340.** What is the `cmp.Ordered` constraint useful for?

- a) Sorting integers only
- b) Writing generic functions that compare values with `<`, `>` — works for integers, floats, strings
- c) Ordering channels by capacity
- d) Comparing interface values

> ✅ **b** — `func Min[T cmp.Ordered](a, b T) T` works for any ordered type.

---

**Q341.** True or False: A type parameter can be constrained to be a pointer type.

> ✅ **True** — `[T interface{ *SomeType }]` or using method constraints implicitly.

---

**Q342.** What is the problem with this generic function?
```go
func Sum[T any](nums []T) T {
    var total T
    for _, n := range nums { total += n }
    return total
}
```
- a) No problem — generics support `+` for any type
- b) Compile error — `+=` requires a numeric constraint like `constraints.Integer | constraints.Float`
- c) Works only for int
- d) Runtime panic for non-numeric types

> ✅ **b** — `any` doesn't guarantee `+`. Need: `[T interface{ ~int | ~int64 | ~float64 ... }]` or `cmp.Ordered`.

---

**Q343.** How do you write a generic function that works only for slice of comparable types?

- a) `func Contains[T any](s []T, v T) bool`
- b) `func Contains[T comparable](s []T, v T) bool`
- c) `func Contains[T interface{}](s []T, v T) bool`
- d) `func Contains(s []interface{}, v interface{}) bool`

> ✅ **b** — `comparable` constraint is required to use `==`.

---

**Q344.** True or False: Generic code in Go is compiled via monomorphization (a separate copy per instantiation, like C++ templates).

> ✅ **False** — Go uses a GCShape-based approach: types with the same GC shape (pointer types, scalar sizes) share implementations. Not full monomorphization. Less code bloat.

---

**Q345.** What package provides `slices.Index`, `slices.Contains`, `slices.Sort`?

- a) `sort`
- b) `slices` (Go 1.21+)
- c) `container/slices`
- d) `golang.org/x/exp/slices`

> ✅ **b** — Go 1.21 added the `slices` package with generic slice utilities. Previously in `golang.org/x/exp/slices`.

---

**Q346.** True or False: You can use a union type constraint like `int | string` to call methods common to both.

> ✅ **False** — Union constraints only allow operations common to ALL types in the union. `int` and `string` share no methods, so you can only use them in assignments or pass to other functions — you cannot call `.ToUpper()` since `int` doesn't have it.

---

**Q347.** What does `[T interface{ String() string }]` mean?

- a) T can only be a string
- b) T is constrained to types that have a `String() string` method (implements `fmt.Stringer`)
- c) T is constrained to the `fmt.Stringer` interface explicitly
- d) Both b and c — they are equivalent

> ✅ **d** — Both forms are equivalent. Can also write `[T fmt.Stringer]`.

---

**Q348.** True or False: Generic functions can be passed as first-class values to other functions.

> ✅ **False** — In Go, you cannot pass a generic function WITHOUT instantiating it first. `foo[int]` returns a `func(int) int` that CAN be passed. But a raw generic `foo` cannot.

---

**Q349.** What is the `maps` package (Go 1.21+) used for?

- a) Third-party map implementation
- b) Generic utility functions for maps: `maps.Keys`, `maps.Values`, `maps.Copy`, `maps.Clone`
- c) Ordered map implementation
- d) Concurrent map utilities

> ✅ **b** — Standard library `maps` package with generic map utilities, complementing the `slices` package.

---

**Q350.** What is the key advantage of generics over `interface{}` for collection types?

- a) Generics are always faster
- b) Type safety at compile time — no runtime type assertions needed
- c) More flexible than interfaces
- d) Allow method specialization

> ✅ **b** — `Stack[int].Pop()` returns `int` directly. `Stack` with `interface{}` returns `interface{}` requiring `.(int)` assertion.

---

## PART 15 — Testing (Q351–Q375)

**Q351.** What is table-driven testing?

- a) Running tests in a table database
- b) Organizing test cases as a slice of structs with inputs and expected outputs, iterating with `t.Run` subtests
- c) A testing framework for HTML tables
- d) Running tests in order from a spreadsheet

> ✅ **b** — Table-driven tests are idiomatic Go. Each row is a test case; `t.Run(tt.name, func(t *testing.T)...)` for subtests.

---

**Q352.** What does `t.Run("name", func(t *testing.T) { ... })` create?

- a) A new test function
- b) A subtest that can be run independently with `-run TestParent/name`
- c) A goroutine for parallel testing
- d) A test that always passes

> ✅ **b** — Subtests appear as `TestFoo/subtest_name` in output. Can be targeted with `-run TestFoo/specific_case`.

---

**Q353.** True or False: `t.Fatal` and `t.Error` both stop the current test immediately.

> ✅ **False** — `t.Fatal` stops the test immediately (calls `runtime.Goexit()`). `t.Error` marks the test as failed but CONTINUES execution.

---

**Q354.** What is the purpose of `t.Helper()`?

- a) Adds test helpers to the test binary
- b) Marks the current function as a test helper — errors show the CALLER's line instead of this function's line
- c) Parallelizes helper functions
- d) Required for `t.Cleanup` to work

> ✅ **b** — Without `t.Helper()`, test failure lines point into the helper function. With it, they point to the test that called the helper — much more useful.

---

**Q355.** What command runs only tests whose name matches `TestUser`:

- a) `go test -match TestUser ./...`
- b) `go test -run TestUser ./...`
- c) `go test TestUser ./...`
- d) `go test -name TestUser ./...`

> ✅ **b** — `-run` accepts a regex. `go test -run TestUser/create` runs only the "create" subtest of TestUser.

---

**Q356.** True or False: Benchmark functions in Go must start with `Benchmark`.

> ✅ **True** — Convention: `func BenchmarkMyFunc(b *testing.B)`. Run with `go test -bench=.`.

---

**Q357.** What does `b.N` represent in a benchmark?

- a) The number of test files
- b) The number of iterations — the testing framework adjusts N until timing is reliable
- c) The maximum allowed allocations
- d) The number of CPU cores used

> ✅ **b** — `for i := 0; i < b.N; i++ { doWork() }` — the framework starts with small N and increases until stable.

---

**Q358.** What does `go test -benchmem` additionally report?

- a) Memory limit used by the test
- b) Allocations per operation (allocs/op) and bytes allocated per operation (B/op)
- c) Memory leaks
- d) GOMAXPROCS used during benchmark

> ✅ **b** — `ns/op  B/op  allocs/op` — critical for optimizing allocations.

---

**Q359.** True or False: `t.Parallel()` should be called at the start of a test function that can safely run concurrently with other parallel tests.

> ✅ **True** — `t.Parallel()` pauses the test until serial tests in the same function finish, then lets parallel tests run concurrently.

---

**Q360.** What does `httptest.NewRecorder()` return?

- a) An `*http.Response` from a real server
- b) A `*httptest.ResponseRecorder` that implements `http.ResponseWriter` for testing
- c) A mock HTTP server
- d) A request body reader

> ✅ **b** — Pass it as `w` to your handler. After the call, inspect `w.Code`, `w.Body.String()`, `w.Header()`.

---

**Q361.** What does `httptest.NewServer(handler)` do differently from `httptest.NewRecorder()`?

- a) No difference
- b) `NewServer` starts a real local HTTP server; `NewRecorder` tests without a server
- c) `NewServer` is for HTTPS; `NewRecorder` is for HTTP
- d) `NewServer` is used for benchmarks

> ✅ **b** — Use `NewServer` for integration tests that make real HTTP requests. Use `NewRecorder` for unit tests of individual handlers.

---

**Q362.** What is a test fixture?

- a) A testing library
- b) Pre-configured test data or environment setup used across multiple tests
- c) A test that always fails
- d) A compiled test binary

> ✅ **b** — Fixtures standardize test setup: database state, files, structs populated with known values.

---

**Q363.** True or False: `go test ./...` caches test results and won't re-run unchanged tests.

> ✅ **True** — Go caches test results based on inputs. Use `go test -count=1 ./...` to bypass the cache.

---

**Q364.** What is `t.Cleanup(func())` used for?

- a) Auto-formatting test files after test runs
- b) Registering a cleanup function that runs when the test (and all subtests) complete
- c) Cleaning test caches
- d) Required for goroutine-based tests

> ✅ **b** — Alternative to `defer` in tests; also runs after subtests complete. Multiple `Cleanup` calls run in LIFO order.

---

**Q365.** What is a golden file test?

- a) A test that always passes
- b) A test where expected output is stored in a file (`testdata/*.golden`) and compared with actual output
- c) A performance benchmark against a golden baseline
- d) A test using the gold package

> ✅ **b** — Use `go test -update` (convention with `-flag`) to regenerate golden files. Prevents large expected strings in test code.

---

**Q366.** True or False: `testing.T` and `testing.B` implement the same `testing.TB` interface.

> ✅ **True** — Common helper functions accept `testing.TB` to work with both tests (`*testing.T`) and benchmarks (`*testing.B`).

---

**Q367.** What does `b.ResetTimer()` do in a benchmark?

- a) Resets N to 1
- b) Stops then resets the benchmark timer — used after expensive setup to not measure setup time
- c) Resets all benchmark results
- d) Causes the benchmark to run for exactly 1 second

> ✅ **b** — `b.ResetTimer()` after setup: only the actual benchmark loop is measured.

---

**Q368.** True or False: The `testdata/` directory is special — its contents are available to tests but not compiled.

> ✅ **True** — Go tools never compile or import from `testdata/`. It's the conventional location for test input files, golden files, etc.

---

**Q369.** What flag runs tests with verbose output (showing test names and results)?

- a) `go test -verbose ./...`
- b) `go test -v ./...`
- c) `go test -output ./...`
- d) `go test -log ./...`

> ✅ **b** — `-v` shows each test name and PASS/FAIL. Without it, only failures are shown.

---

**Q370.** What is dependency injection in Go testing?

- a) Using the `inject` package to automatically wire dependencies
- b) Passing dependencies (like repositories, HTTP clients) as interface parameters so tests can substitute mocks
- c) Injecting test data into production databases
- d) Using build tags to switch implementations

> ✅ **b** — Design code to accept interfaces. In tests, pass fake/mock implementations. No special library needed.

---

**Q371.** True or False: `go test -short ./...` skips tests that call `t.Skip()` unconditionally.

> ✅ **False** — `-short` only skips tests that explicitly check `testing.Short()` and call `t.Skip`. Normal tests still run.

---

**Q372.** What does `b.StartTimer()` and `b.StopTimer()` do?

- a) Start/stop the benchmark goroutine
- b) Include/exclude specific code sections from the benchmark timing
- c) Control the GC during benchmarks
- d) Control CPU profiling

> ✅ **b** — `b.StopTimer()` pauses timing (e.g., for per-iteration setup). `b.StartTimer()` resumes.

---

**Q373.** True or False: Two table-driven test cases with the same `name` in the same test cause a compile error.

> ✅ **False** — Duplicate subtest names don't cause compile errors. Go appends `#01`, `#02`, etc. to disambiguate in output.

---

**Q374.** What does `go test -cover -coverprofile=cov.out ./...` produce?

- a) A list of uncovered functions
- b) A coverage profile file that can be visualized with `go tool cover -html=cov.out`
- c) A build with coverage, but no output file
- d) Fails if coverage is below 80%

> ✅ **b** — The HTML visualization shows which lines are covered (green) and which are not (red).

---

**Q375.** What is the purpose of the `_test.go` file naming convention?

- a) Files named `_test.go` are ignored by the compiler
- b) Files ending in `_test.go` are compiled only during `go test` — not included in production build
- c) These files replace the regular `.go` files during testing
- d) The `_` prefix makes all identifiers in them unexported

> ✅ **b** — `_test.go` files are exclusively for tests. They can also use package `foo_test` (external/black-box testing) or `foo` (white-box/internal testing).

---

## PART 16 — Standard Library (Q376–Q400)

**Q376.** What does `fmt.Sprintf("%v", user)` print when `user` is a struct?

- a) The JSON representation
- b) Default format: `{FieldValue1 FieldValue2 ...}`
- c) `<User object>`
- d) Compile error

> ✅ **b** — `%v`: default. `%+v`: with field names. `%#v`: Go syntax representation.

---

**Q377.** What format verb prints Go-syntax representation of a value?

- a) `%v`
- b) `%+v`
- c) `%#v`
- d) `%T`

> ✅ **c** — `%#v` prints `main.User{Name:"Alice", Age:30}`. Useful for debugging.

---

**Q378.** True or False: `strings.Builder` is the most efficient way to build a string via repeated concatenation.

> ✅ **True** — `strings.Builder` uses a `[]byte` buffer internally. String concatenation with `+` creates a new string each time — O(n²) total.

---

**Q379.** What does `strings.Cut("user@example.com", "@")` return?

- a) `("user", "example.com", true)`
- b) `"user"` and `"example.com"` as one string
- c) `["user", "example.com"]`
- d) `("user", "example.com", nil)`

> ✅ **a** — `Cut` returns `(before, after, found)`. More idiomatic than `strings.SplitN(s, "@", 2)`.

---

**Q380.** What does `os.ReadFile("data.txt")` return?

- a) `(*os.File, error)`
- b) `([]byte, error)` — reads entire file into memory
- c) `(io.Reader, error)`
- d) `(string, error)`

> ✅ **b** — Convenience function. Opens, reads all, closes. For large files, use `os.Open` + streaming.

---

**Q381.** True or False: `io.Copy(dst, src)` reads from `src` until EOF or error, writing all data to `dst`.

> ✅ **True** — `io.Copy` is the idiomatic way to stream data between an `io.Reader` and `io.Writer` without loading all into memory.

---

**Q382.** What does `bufio.Scanner.Scan()` return when EOF is reached?

- a) `true` with empty text
- b) `false` — loop ends; check `scanner.Err()` for actual errors
- c) `io.EOF` error
- d) Panics

> ✅ **b** — `for scanner.Scan() { }` — when `Scan()` returns false, either EOF or error. Always check `scanner.Err()` after the loop.

---

**Q383.** True or False: `bufio.Writer.Flush()` is optional — unflushed data is automatically written when the writer is garbage collected.

> ✅ **False** — ALWAYS call `Flush()` (or `defer bw.Flush()`). Unflushed data in the buffer is LOST if not explicitly flushed.

---

**Q384.** What JSON struct tag makes a field excluded from marshaling entirely?

- a) `json:"omit"`
- b) `json:"-"`
- c) `json:"hidden"`
- d) `json:",ignore"`

> ✅ **b** — `json:"-"` means: never include this field in JSON output, even if it has a value.

---

**Q385.** What does `json:",omitempty"` do?

- a) Omits the field if it has any value
- b) Omits the field if it has the zero value (0, false, "", nil, empty slice)
- c) Omits the field from JSON input only
- d) Requires the field in JSON input or returns error

> ✅ **b** — `omitempty` skips zero-valued fields. Useful for optional fields in API responses.

---

**Q386.** True or False: `json.NewEncoder(w).Encode(v)` appends a newline after the JSON.

> ✅ **True** — Unlike `json.Marshal`, `Encoder.Encode` always appends `\n`. Useful for NDJSON (newline-delimited JSON).

---

**Q387.** What is Go's time formatting reference time?

- a) `2000-01-01T00:00:00Z`
- b) `Mon Jan 2 15:04:05 MST 2006` (= 01/02 03:04:05 PM '06 -0700)
- c) `1970-01-01T00:00:00Z` (Unix epoch)
- d) `1999-12-31T23:59:59Z`

> ✅ **b** — MEMORIZE IT. Go uses this specific reference time. `time.Now().Format("2006-01-02")` formats current time as date.

---

**Q388.** What does `time.Since(t)` compute?

- a) The time until `t`
- b) `time.Now().Sub(t)` — duration since time `t`
- c) Wall clock time in seconds
- d) The monotonic clock reading at `t`

> ✅ **b** — `time.Since(start)` is idiomatic for elapsed time measurement.

---

**Q389.** True or False: `log.Fatal` calls `os.Exit(1)` — deferred functions do NOT run.

> ✅ **True** — `log.Fatal` = log + `os.Exit(1)`. `os.Exit` bypasses deferred functions. Never use `log.Fatal` inside goroutines or in functions with important defers.

---

**Q390.** What is the difference between `log` (standard) and `log/slog` (Go 1.21+)?

- a) No difference
- b) `slog` provides structured logging with key-value pairs; `log` is unstructured text-only
- c) `slog` is faster than `log`
- d) `slog` supports log levels; `log` doesn't

> ✅ **b** — `slog.Info("request", "method", "GET", "path", "/users")` — structured, greppable logs. Also supports JSON handler.

---

**Q391.** What does `regexp.MustCompile` do differently from `regexp.Compile`?

- a) `MustCompile` is faster
- b) `MustCompile` panics on invalid regex; `Compile` returns an error
- c) `MustCompile` is case-insensitive
- d) `MustCompile` caches the regex

> ✅ **b** — Use `MustCompile` for package-level `var` regex (compile-time constants). Use `Compile` for runtime user input.

---

**Q392.** True or False: `strconv.Itoa` is faster than `fmt.Sprintf("%d", n)` for int-to-string conversion.

> ✅ **True** — `strconv.Itoa` is specialized. `fmt.Sprintf` parses the format string at runtime. For int conversion, `strconv` is significantly faster.

---

**Q393.** What does `strings.NewReplacer` do better than multiple `strings.Replace` calls?

- a) Nothing — they are equivalent
- b) Replaces all patterns in a single pass through the string — more efficient for many replacements
- c) Supports regex patterns
- d) Works on bytes, not strings

> ✅ **b** — `strings.NewReplacer("a","1","b","2","c","3")` does all replacements in one O(n) pass. Multiple `Replace` calls are O(n*k) where k = number of replacements.

---

**Q394.** What does `io.TeeReader(r, w)` do?

- a) Reads from two readers alternately
- b) Returns a reader that reads from `r` and simultaneously writes everything read to `w` (like Unix `tee`)
- c) Duplicates data read from `r`
- d) Returns a writer that writes to both `r` and `w`

> ✅ **b** — Use for logging/debugging: `io.TeeReader(resp.Body, os.Stderr)` prints response body while also reading it normally.

---

**Q395.** True or False: `os.OpenFile("f", os.O_APPEND|os.O_WRONLY, 0644)` creates the file if it doesn't exist.

> ✅ **False** — `O_APPEND|O_WRONLY` without `O_CREATE` does NOT create the file. Add `os.O_CREATE` flag to create if missing.

---

**Q396.** What does `strings.Fields("  a  b  c  ")` return?

- a) `["  a", "  b", "  c  "]`
- b) `["a", "b", "c"]`
- c) `["a  b  c"]`
- d) `["", "a", "b", "c", ""]`

> ✅ **b** — `strings.Fields` splits by any whitespace and trims leading/trailing — equivalent to Python's `str.split()`.

---

**Q397.** What is the purpose of `io.LimitReader(r, n)`?

- a) Limits writes to `r`
- b) Returns a reader that reads at most `n` bytes from `r`
- c) Reads exactly `n` bytes or errors
- d) Rate-limits reads from `r`

> ✅ **b** — Use to prevent reading more data than expected (e.g., limiting HTTP request body size).

---

**Q398.** True or False: `fmt.Sscanf` reads from a string (like `fmt.Scanf` from stdin).

> ✅ **True** — `fmt.Sscanf("Alice 30", "%s %d", &name, &age)` — parses a string as formatted input.

---

**Q399.** What does `time.NewTicker(d)` return, and when should you use `defer ticker.Stop()`?

- a) Returns `<-chan time.Time`; Stop is optional
- b) Returns `*time.Ticker` with a `.C` channel; always `defer ticker.Stop()` to prevent goroutine leak
- c) Returns a `time.Timer`; Stop permanently stops the ticker
- d) Returns an `io.Reader` of time values

> ✅ **b** — The ticker has a background goroutine. Not stopping it leaks resources.

---

**Q400.** Which standard library function is the PREFERRED way to read an HTTP response body?

- a) `resp.Body.Read(buf)`
- b) `io.ReadAll(resp.Body)`
- c) `fmt.Fscan(resp.Body, &result)`
- d) `bufio.NewScanner(resp.Body).Scan()`

> ✅ **b** — `io.ReadAll(resp.Body)` reads all content. Always `defer resp.Body.Close()`. For large responses, prefer streaming with `json.NewDecoder(resp.Body)`.

---

*End of Quiz Part 4 (Q301–Q400)*
*Continue with go_quiz_part5.md for Q401–Q500*
