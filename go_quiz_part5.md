# Go Quiz — Part 5 of 5 (Q401–Q500)
### Covers: Parts 17–20 (HTTP, C++ Migration, Advanced Patterns, Interview Prep)
> Answer is hidden after `> ✅`. Try to answer before looking!

---

## PART 17 — HTTP Server & Client (Q401–Q420)

**Q401.** True or False: Using `http.ListenAndServe(":8080", nil)` with `http.DefaultServeMux` is safe for production.

> ✅ **False** — `http.DefaultServeMux` has no timeouts. A slow or hung client can hold connections open forever. Always use `&http.Server{Addr: ":8080", ReadTimeout: ..., WriteTimeout: ...}`.

---

**Q402.** What does `http.Server.ReadTimeout` control?

- a) Time to read the entire request body
- b) Time from when the connection is accepted to when the full request (including body) has been read
- c) Time between request headers
- d) Time to establish the TCP connection

> ✅ **b** — `ReadTimeout` includes header reading. `ReadHeaderTimeout` is specifically for headers only.

---

**Q403.** True or False: In Go 1.22+, you can route by HTTP method in `http.ServeMux`: `mux.HandleFunc("GET /users", handler)`.

> ✅ **True** — Go 1.22 added method+path routing to the standard library. Previously needed a third-party router.

---

**Q404.** How do you get a path variable `{id}` in a Go 1.22+ handler?

- a) `r.PathValue("id")`
- b) `r.URL.Query().Get("id")`
- c) `r.Header.Get("id")`
- d) `mux.Vars(r)["id"]` (gorilla/mux style)

> ✅ **a** — `r.PathValue("id")` returns the value matched by `{id}` in the route pattern.

---

**Q405.** What is the correct order of operations for writing an HTTP response?

- a) Write body → WriteHeader → Header.Set
- b) Header.Set → WriteHeader → Write body
- c) WriteHeader → Header.Set → Write body
- d) Any order — Go buffers and sends correctly

> ✅ **b** — Headers must be set BEFORE `WriteHeader`. Once `WriteHeader` is called (or `Write` is called, which implicitly calls it), headers are sent and cannot be changed.

---

**Q406.** Why should you always `defer resp.Body.Close()` when making HTTP GET requests?

- a) To prevent goroutine leaks
- b) To return the connection to the pool for reuse; unclosed body leaks the connection
- c) The body is not needed after reading
- d) Because `io.ReadAll` holds a reference

> ✅ **b** — Not closing response body keeps the connection open and prevents pooling. Even for errors, close the body.

---

**Q407.** True or False: `http.DefaultClient` is sufficient for production HTTP requests.

> ✅ **False** — `http.DefaultClient` has NO timeout. A hanging server causes your goroutine to block forever. Always create `&http.Client{Timeout: 30*time.Second}`.

---

**Q408.** What is middleware in Go HTTP context?

- a) Software between OS and Go runtime
- b) A function of type `func(http.Handler) http.Handler` that wraps a handler to add behavior
- c) The routing layer
- d) A third-party package required for HTTP servers

> ✅ **b** — Classic pattern: logging, auth, recovery, CORS — all implemented as `func(http.Handler) http.Handler` functions.

---

**Q409.** What does `http.HandlerFunc` do?

- a) Registers a handler function with the default mux
- b) Converts a function of signature `func(w http.ResponseWriter, r *http.Request)` into an `http.Handler`
- c) Calls a handler function with a fake request for testing
- d) Creates a new mux

> ✅ **b** — `http.HandlerFunc` is a type adapter: `type HandlerFunc func(ResponseWriter, *Request)` with a `ServeHTTP` method.

---

**Q410.** True or False: `server.Shutdown(ctx)` immediately closes all active connections.

> ✅ **False** — `Shutdown` is **graceful**: it stops accepting new connections and waits for existing connections to finish (within the context deadline). Active connections can complete their requests.

---

**Q411.** How do you get the request body as a string in a handler?

- a) `r.Body.String()`
- b) `body, _ := io.ReadAll(r.Body); defer r.Body.Close(); s := string(body)`
- c) `r.BodyString()`
- d) `fmt.Fscan(r.Body)`

> ✅ **b** — `io.ReadAll(r.Body)` reads all bytes. Remember to `defer r.Body.Close()`.

---

**Q412.** What does `http.Transport` control?

- a) HTTP protocol version switching (HTTP/1.1 vs HTTP/2)
- b) Low-level connection management: connection pooling, TLS, timeouts, keep-alive
- c) Request/response compression
- d) URL routing

> ✅ **b** — Tune `MaxIdleConns`, `MaxIdleConnsPerHost`, `IdleConnTimeout`, etc. in `Transport` for production performance.

---

**Q413.** True or False: A Go HTTP handler runs in its own goroutine for each request.

> ✅ **True** — Each HTTP request gets its own goroutine. This is why Go servers handle high concurrency well.

---

**Q414.** What is the purpose of wrapping `http.ResponseWriter` in tests?

- a) To add compression
- b) To capture the status code and body written by the handler for assertions
- c) To implement HTTPS
- d) Required for middleware to work

> ✅ **b** — `httptest.NewRecorder()` implements `http.ResponseWriter` and records what the handler writes. Check `recorder.Code`, `recorder.Body`.

---

**Q415.** True or False: `json.NewDecoder(r.Body).Decode(&v)` is preferred over `json.Unmarshal` for HTTP handlers.

> ✅ **True** — `Decoder` streams directly from the body without loading all bytes into memory first. Better for large bodies.

---

**Q416.** What does `json.Decoder.DisallowUnknownFields()` do?

- a) Rejects JSON with null values
- b) Returns an error if the JSON contains fields not present in the target struct
- c) Ignores extra JSON fields silently
- d) Rejects non-string JSON keys

> ✅ **b** — Useful for strict API validation — reject requests with unexpected fields (typos, wrong API version).

---

**Q417.** True or False: You can call `w.WriteHeader` multiple times in the same handler.

> ✅ **False** — Only the FIRST `WriteHeader` call has effect. Subsequent calls are logged and ignored. (This is a common bug when writing error responses after partial work.)

---

**Q418.** What HTTP status code should a successful resource creation return?

- a) 200 OK
- b) 201 Created
- c) 202 Accepted
- d) 204 No Content

> ✅ **b** — `201 Created` for POST that creates a resource. `200` for successful GET/PUT. `204` for successful DELETE.

---

**Q419.** What is the retry pattern for HTTP clients?

- a) Retry all failed requests immediately
- b) Retry only server errors (5xx) with exponential backoff and jitter; never retry client errors (4xx)
- c) Retry all errors exactly 3 times
- d) Retry only timeout errors

> ✅ **b** — 4xx = client's fault (wrong request), no point retrying. 5xx = server's fault, retry with backoff. Jitter prevents thundering herd.

---

**Q420.** What signal should a production Go HTTP server listen for to trigger graceful shutdown?

- a) `SIGKILL` and `SIGTERM`
- b) `SIGINT` and `SIGTERM`
- c) `SIGHUP` only
- d) Use a special HTTP endpoint `/shutdown`

> ✅ **b** — `signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)`. `SIGKILL` cannot be caught.

---

## PART 18 — C++ to Go Migration (Q421–Q440)

**Q421.** What is the Go equivalent of a C++ class constructor?

- a) The struct's `init()` function
- b) A convention function named `NewXxx()` that returns `(*Xxx, error)` or `*Xxx`
- c) A method named `Xxx.New()`
- d) The `make` function

> ✅ **b** — `func NewUser(name, email string) (*User, error)` — validates and returns initialized struct. No constructor syntax in Go.

---

**Q422.** True or False: Go's struct embedding is equivalent to C++ public inheritance.

> ✅ **False** — Embedding is **composition** (HAS-A). In Go's type system, a `Dog` that embeds `Animal` is NOT assignable to an `Animal` variable. It's syntactic sugar for field/method promotion.

---

**Q423.** What is the Go equivalent of C++ RAII (Resource Acquisition Is Initialization)?

- a) `sync.Pool`
- b) `defer` — resource acquired manually; cleanup scheduled with defer
- c) Smart pointers via the `unsafe` package
- d) Finalizers via `runtime.SetFinalizer`

> ✅ **b** — `defer file.Close()` is the Go RAII equivalent. Cleanup is explicit and guaranteed even on panic, but not automatic on construction.

---

**Q424.** What is the Go equivalent of `std::unique_ptr<T>`?

- a) `*T` — Go's GC handles the lifetime automatically
- b) `unsafe.Pointer`
- c) `sync.Pool`
- d) `new(T)` — the allocator

> ✅ **a** — Go has no ownership semantics. A `*T` lives as long as something references it. GC collects when unreachable.

---

**Q425.** True or False: Go's goroutine is equivalent to `std::thread`.

> ✅ **False** — Goroutines are much lighter: ~2KB vs ~8MB for `std::thread`. Millions of goroutines can coexist. They're scheduled by Go's M:N runtime scheduler, not the OS.

---

**Q426.** What is the Go equivalent of C++ `std::vector<T>`?

- a) `array[T]`
- b) `[]T` (Go slice)
- c) `list.List`
- d) `container/vector`

> ✅ **b** — Go slices are dynamically resizable like `std::vector`. `append` is the equivalent of `push_back`.

---

**Q427.** What is the Go equivalent of C++ exceptions?

- a) `panic` and `recover`
- b) Multiple return values `(result, error)` — errors are values
- c) An `error` channel pattern
- d) `defer` with error assignment

> ✅ **b** — In C++: `throw`/`catch`. In Go: return `error`, check `if err != nil`. `panic`/`recover` exist but are not for normal error flow.

---

**Q428.** How do you express C++ pure virtual functions in Go?

- a) Use `abstract` keyword
- b) Define a method in an interface — any type implementing the method satisfies the interface
- c) Use `//go:noinline` directive
- d) Embed an empty struct

> ✅ **b** — Go interfaces are pure abstract. There's no "virtual" keyword. Any type with matching method signatures satisfies the interface automatically.

---

**Q429.** What is the Go equivalent of C++ `std::map<K,V>` (ordered)?

- a) `map[K]V` (same as C++ std::map)
- b) There is NO ordered map in Go's standard library — use `map[K]V` (unordered) or third-party btree
- c) `sync.Map`
- d) `container/heap`

> ✅ **b** — Go's `map[K]V` is a hash map (like `std::unordered_map`). There is no `std::map` equivalent (red-black tree) in stdlib. Use `golang.org/x/exp/maps` or a third-party btree.

---

**Q430.** True or False: C++ namespaces directly correspond to Go packages.

> ✅ **True (mostly)** — Both organize code into named scopes. Key difference: in Go you ALWAYS qualify names (`pkg.Func()`); there's no `using namespace`. Package name = last element of import path.

---

**Q431.** What is the Go equivalent of C++ template specialization?

- a) Generic type constraints with `~` (tilde)
- b) There is NO template specialization in Go — use interfaces or separate functions
- c) Type aliases
- d) Interface embedding

> ✅ **b** — Go generics don't support specialization. Use interfaces for type-specific behavior, or just write separate functions.

---

**Q432.** True or False: C++ copy constructors have a direct equivalent in Go.

> ✅ **False** — Go structs are copied by value assignment (`b := a`). Slices/maps/channels within structs are shallow-copied. Go has no user-defined copy constructors.

---

**Q433.** What is the Go equivalent of C++ `static` local variables?

- a) `var` at package level
- b) `sync.Once` + package-level variable for lazy initialization
- c) `const` block
- d) `init()` function

> ✅ **b** — C++ `static` local = initialize once, persist across calls. In Go: `var instance *Type; var once sync.Once; once.Do(func() { instance = ... })`.

---

**Q434.** How do you express C++ multiple inheritance in Go?

- a) `type Dog struct { Animal, Swimmer }` — embed multiple structs
- b) Multiple interfaces implemented by one type
- c) Both a and b — embedding for fields, interfaces for behavior
- d) Not possible in Go

> ✅ **c** — Embed multiple structs for promotion. Implement multiple interfaces for polymorphism. Both are valid approaches.

---

**Q435.** True or False: Go's `map[K]V` is thread-safe by default, unlike C++'s `std::unordered_map`.

> ✅ **False** — Both are NOT thread-safe. C++ `std::unordered_map` and Go `map[K]V` require external synchronization for concurrent use.

---

**Q436.** What is the Go equivalent of C++ `const` member function (read-only)?

- a) Method with value receiver: `func (t T) ReadOnly() int`
- b) Method with pointer receiver but no writes
- c) Interface method
- d) Function with `const` keyword

> ✅ **a** — A value receiver gets a copy; modifications don't affect the original. Semantically equivalent to C++ const member functions.

---

**Q437.** True or False: C++ friend classes have a direct equivalent in Go.

> ✅ **False** — Go uses package-level access control instead. Unexported (lowercase) identifiers are accessible within the same **package** — similar to "all types in the same package are friends." There's no `friend` keyword.

---

**Q438.** What is the Go equivalent of C++ `std::shared_ptr<T>` (reference counting)?

- a) `&T{}`
- b) `*T` — Go's GC tracks all references automatically; reference counting is implicit
- c) `atomic.Pointer[T]`
- d) `sync.Pool`

> ✅ **b** — Go's GC (using a mark-sweep algorithm) handles shared ownership automatically. No reference counting exposed to users.

---

**Q439.** What is the Go pattern for C++ CRTP (Curiously Recurring Template Pattern)?

- a) Type embedding with interface
- b) Generics with type constraints
- c) Go doesn't need CRTP — most CRTP use cases in C++ are solved by interfaces in Go
- d) `reflect` package

> ✅ **c** — CRTP in C++ enables "static polymorphism" to avoid virtual function overhead. Go interfaces + compiler inlining achieve similar results without the complexity.

---

**Q440.** True or False: In Go, you MUST use `make` to initialize a struct, similar to C++'s `new`.

> ✅ **False** — `make` is ONLY for slices, maps, and channels. Structs are initialized with struct literals `User{Name: "Alice"}` or `new(User)`. Zero-value structs are also valid directly: `var u User`.

---

## PART 19 — Advanced Patterns (Q441–Q460)

**Q441.** What is the Repository pattern?

- a) A design for storing files in Git
- b) An abstraction layer (interface) over data storage — separates domain logic from storage details
- c) A pattern for HTTP routing
- d) A factory for creating domain objects

> ✅ **b** — `type UserRepository interface { GetByID(ctx, id) (*User, error); Create(ctx, *User) error }` — implemented by PostgresUserRepo, MockUserRepo for tests, etc.

---

**Q442.** True or False: The functional options pattern allows adding new configuration options without breaking existing callers.

> ✅ **True** — `func Connect(opts ...Option)` — adding a new `WithSSLMode` option doesn't require changing callers that don't use it.

---

**Q443.** What is the purpose of the Circuit Breaker pattern?

- a) Protect electrical circuits in server rooms
- b) Prevent cascading failures by failing fast when a downstream service is consistently failing
- c) Limit API rate from clients
- d) Buffer requests during high load

> ✅ **b** — When errors exceed threshold, circuit "opens" and rejects requests immediately instead of waiting for timeouts. Allows downstream service to recover.

---

**Q444.** What are the three states of a Circuit Breaker?

- a) Open, Closed, Broken
- b) Closed (normal), Open (failing), Half-Open (testing recovery)
- c) Active, Inactive, Pending
- d) Working, Failed, Recovering

> ✅ **b** — Closed = requests pass through. Open = fail fast. Half-Open = allow one request through to test if service recovered.

---

**Q445.** What does the Worker Pool pattern solve?

- a) Memory management for workers
- b) Bounds concurrency — prevents unlimited goroutine creation for unlimited incoming work
- c) Load balancing across servers
- d) Thread management in the OS

> ✅ **b** — Fixed N worker goroutines process a job queue channel. N goroutines instead of N goroutines per job.

---

**Q446.** True or False: In the Event Bus pattern, subscribers should be called synchronously to maintain ordering.

> ✅ **False** — Event bus subscribers are typically called in goroutines (asynchronously). Always `defer recover()` in event handlers to prevent one handler's panic from crashing all handlers.

---

**Q447.** What is the Builder pattern with accumulated errors?
```go
qb := Query("users").Where("age > 18").Limit(-1).Build()
```
- a) Returns the query ignoring invalid options
- b) Accumulates errors during chaining; `Build()` returns the first (or all) errors
- c) Panics immediately on invalid options
- d) A compile-time pattern

> ✅ **b** — Each method accumulates errors in the builder. `Build()` returns them all. Allows fluent chaining while still surfacing all validation errors.

---

**Q448.** What is the key property of idiomatic Go error handling?

- a) All errors must be wrapped
- b) Errors are values — returned explicitly, checked explicitly, wrapped with context
- c) Errors must be logged before returning
- d) Errors should be converted to panics at package boundaries

> ✅ **b** — Errors are values (not exceptions). They are returned, checked, and propagated with `fmt.Errorf("ctx: %w", err)`.

---

**Q449.** True or False: Storing `context.Context` in a struct is the recommended way to propagate context in Go.

> ✅ **False** — NEVER store context in a struct. Pass it as the first parameter to each function: `func (s *Service) GetUser(ctx context.Context, id int)`.

---

**Q450.** What is the idiomatic Go principle "Accept interfaces, return structs"?

- a) Function parameters should be concrete types; return interfaces
- b) Function parameters should be interfaces (maximum flexibility for callers); return concrete types (caller knows exactly what they get)
- c) Always use interfaces for both input and output
- d) Only HTTP handlers follow this rule

> ✅ **b** — `func processData(r io.Reader) *ProcessedData` — accepts any reader (file, string, HTTP body); returns a specific type (not an interface).

---

**Q451.** What does the Graceful Shutdown pattern do?

- a) Deletes logs cleanly on exit
- b) Stops accepting new connections, waits for in-flight requests to complete, then exits cleanly
- c) Saves application state to disk
- d) Sends SIGTERM to all goroutines

> ✅ **b** — `server.Shutdown(ctx)` after catching SIGINT/SIGTERM. Essential for zero-downtime deployments.

---

**Q452.** True or False: `sync.Pool` is safe for storing database connections.

> ✅ **False** — The GC can evict pool objects at any time. Database connections need managed lifetimes. Use `database/sql.DB` which manages its own connection pool.

---

**Q453.** What is the difference between `errors.Is` and a type assertion in error handling?

- a) Same thing
- b) `errors.Is` traverses the error chain; type assertion only checks the outermost error type
- c) `errors.Is` is faster
- d) `errors.Is` is for interfaces; type assertion is for concrete types

> ✅ **b** — `errors.Is(wrapped, ErrNotFound)` finds `ErrNotFound` even 5 wraps deep. `wrapped.(*NotFoundError)` only works if `wrapped` directly IS `*NotFoundError`.

---

**Q454.** What is the `io.ReadCloser` interface used for?

- a) Reading and closing files only
- b) Combining `io.Reader` and `io.Closer` — the type for HTTP response bodies, request bodies
- c) Reading and writing
- d) Closing channels

> ✅ **b** — `http.Response.Body` is `io.ReadCloser`. Must be closed after reading to return the connection to the pool.

---

**Q455.** True or False: `errgroup.Group.Wait()` returns all errors from all goroutines.

> ✅ **False** — `Wait()` returns the **first non-nil error**. Use a custom multi-error approach if you need all errors.

---

**Q456.** What is the `Do's and Don'ts` of using global state in Go?

- a) Global state is required for singleton services
- b) Avoid global mutable state — it makes testing hard, causes data races. Use dependency injection instead.
- c) Global state is fine if protected by a mutex
- d) Use `sync.Map` for all global state

> ✅ **b** — Testability and concurrency safety both suffer with global mutable state. Pass dependencies explicitly.

---

**Q457.** True or False: The `internal/` package visibility rule means changes to internal packages never break downstream users.

> ✅ **True** — Since external packages cannot import `internal/`, changing internal packages is a safe, backward-compatible refactoring.

---

**Q458.** What is the main benefit of the "table-driven test + interface dependency injection" combination?

- a) Reduces test file count
- b) Tests multiple scenarios without duplicating setup code, and mocks allow testing each scenario in isolation
- c) Required by the Go compiler
- d) Eliminates the need for integration tests

> ✅ **b** — Table-driven tests enumerate cases; mock dependencies control environment per-case. Clean, maintainable, comprehensive.

---

**Q459.** What makes Go code "production-ready" for concurrency?

- a) Using goroutines for everything
- b) All shared state protected (mutex/atomic/channel), goroutines have exit conditions, contexts propagated, races detected with `-race`
- c) Setting GOMAXPROCS to the right value
- d) Using sync.Map everywhere

> ✅ **b** — Comprehensive list: synchronization, lifecycle management, context cancellation, race detector in CI.

---

**Q460.** True or False: In Go, you should prefer `goroutines + channels` over `mutex + shared memory` in ALL cases.

> ✅ **False** — "Share memory by communicating" is a guideline, not a rule. For simple counters/flags, `sync/atomic` or `sync.Mutex` are simpler and more efficient. Use channels when passing ownership or coordinating complex workflows.

---

## PART 20 — Interview Preparation Final Questions (Q461–Q500)

**Q461.** What does "the bigger the interface, the weaker the abstraction" mean?

- a) Small interfaces with many methods are easier to implement
- b) Interfaces with fewer methods (like io.Reader with 1 method) are satisfied by more types, making them more powerful and reusable
- c) Large interfaces expose more functionality
- d) Interfaces should never have more than 5 methods

> ✅ **b** — `io.Reader` (1 method) is satisfied by files, strings, HTTP bodies, custom buffers. `DatabaseClient` (30 methods) must be satisfied entirely — much harder to mock.

---

**Q462.** True or False: Go guarantees that goroutines scheduled with `go f()` will execute in the order they are started.

> ✅ **False** — Goroutine scheduling is non-deterministic. Order depends on the scheduler, available Ps, and system load.

---

**Q463.** What is the time complexity of Go's `map[K]V` Get and Set operations?

- a) O(log n)
- b) O(n)
- c) O(1) average case
- d) O(1) worst case

> ✅ **c** — Hash map operations are O(1) average. Worst case (hash collision) is O(n) but practically negligible.

---

**Q464.** What is the output?
```go
x := []int{1, 2, 3}
y := x[1:]
y[0] = 99
fmt.Println(x)
```
- a) `[1 2 3]`
- b) `[1 99 3]`
- c) `[99 2 3]`
- d) Compile error

> ✅ **b** — `y` shares the backing array with `x`. `y[0]` = `x[1]`. Modifying `y[0]` changes `x[1]` to 99.

---

**Q465.** True or False: `sync.Mutex` has a `TryLock()` method (Go 1.18+).

> ✅ **True** — `mu.TryLock()` returns `true` if the lock was acquired, `false` without blocking if it's already held. Rarely needed but useful for non-blocking scenarios.

---

**Q466.** What is the purpose of `runtime.Goexit()`?

- a) Terminates the entire program
- b) Terminates the current goroutine, running all deferred functions
- c) Pauses the current goroutine
- d) Removes the goroutine from the scheduler

> ✅ **b** — `runtime.Goexit()` is like returning from the goroutine's top-level function. Deferred functions run. Other goroutines continue. (Used by `t.Fatal` internally.)

---

**Q467.** True or False: Every Go program has at least 2 goroutines running.

> ✅ **True** — The main goroutine + the GC goroutine (at minimum). In practice, many more (HTTP server, runtime monitoring, etc.).

---

**Q468.** What problem does this code have?
```go
ch := make(chan int)
go func() { ch <- 42 }()
```
- a) No problem
- b) Goroutine leak if nobody reads from `ch`
- c) Compile error — goroutine cannot send to channel
- d) The goroutine exits immediately

> ✅ **b** — If nobody reads from the unbuffered channel, the goroutine blocks forever — a goroutine leak. Solution: buffered channel or ensure a receiver.

---

**Q469.** True or False: `[]byte("hello")` and `string("hello")` both allocate new memory.

> ✅ **True** — Converting between `string` and `[]byte` typically allocates. Use `strings.Builder` or `bytes.Buffer` for efficiency in loops.

---

**Q470.** What is `http.StatusUnprocessableEntity`?

- a) 400 — Bad syntax
- b) 404 — Not found
- c) 422 — Semantically invalid (syntactically correct but logically wrong — e.g., age=-1)
- d) 500 — Server error

> ✅ **c** — 422 is for valid JSON but invalid content (validation failure). 400 is for malformed syntax.

---

**Q471.** What is the output?
```go
m := map[string]int{"a": 1}
m["b"] += 10
fmt.Println(m["b"])
```
- a) Compile error
- b) `0` — key "b" doesn't exist
- c) `10`
- d) Runtime panic

> ✅ **c** — `m["b"]` returns 0 (zero value). `0 + 10 = 10`. Map creates the key with value 10.

---

**Q472.** True or False: `for i, v := range s` creates a copy of each element `v`.

> ✅ **True** — `v` is a copy of `s[i]`. Modifying `v` does not modify `s[i]`. To modify, use `s[i]` directly or range over a slice of pointers.

---

**Q473.** In a goroutine pool, when should you close the jobs channel?

- a) As soon as you submit all jobs
- b) After all workers have finished
- c) When the first job fails
- d) Never — let the GC handle it

> ✅ **a** — Close the jobs channel AFTER submitting all jobs to signal workers there's no more work. Workers' `for job := range jobs` loop exits when channel is closed.

---

**Q474.** What does `go:embed` do?

- a) Embeds Go code into another Go file
- b) Embeds files from the filesystem into the compiled binary at build time
- c) Embeds Go binaries into shell scripts
- d) An import alias directive

> ✅ **b** — `//go:embed static/*` embeds all files in `static/` into the binary. No runtime file system dependency.

---

**Q475.** True or False: A method on a nil pointer receiver can execute without panicking.

> ✅ **True** — In Go, nil receiver methods are valid IF the method checks for nil before dereferencing.
```go
func (u *User) Name() string {
    if u == nil { return "unknown" }
    return u.name
}
```

---

**Q476.** What is the difference between `context.WithTimeout` and HTTP client's `Timeout` field?

- a) They are equivalent
- b) `http.Client.Timeout` covers the ENTIRE request (connect + send + receive). Context timeout can be more granular (per-request, per-operation)
- c) Context timeout only covers server processing time
- d) Client Timeout only covers TCP connect

> ✅ **b** — Both can be used together. Context provides finer grained control (can cancel mid-request). Client Timeout is simpler for whole-request limits.

---

**Q477.** True or False: `go test -bench=. -benchtime=10s` runs each benchmark for 10 seconds.

> ✅ **True** — Default is 1 second. `benchtime=10s` gives more iterations for more stable results.

---

**Q478.** What is the output?
```go
type S struct{}
func (s S) M() { fmt.Println("value receiver") }

var i interface{ M() } = S{}
i.M()
```
- a) Compile error — interface requires pointer methods
- b) `value receiver`
- c) Panic
- d) `value receiver` but with extra memory allocation

> ✅ **b** — `S` (value type) satisfies the interface since `M()` has a value receiver. Works fine.

---

**Q479.** What is `go generate` used for?

- a) Generates test data
- b) Runs commands specified in `//go:generate` directives (e.g., for code generation: stringer, mockgen, protoc)
- c) Generates go.mod from existing code
- d) Generates documentation

> ✅ **b** — `go generate ./...` runs all `//go:generate` commands. Common for generating Mock interfaces, string methods for enums, protobuf code.

---

**Q480.** True or False: Go allows you to return different error types from the same function signature `(int, error)`.

> ✅ **True** — The `error` interface can hold any type implementing `Error() string`. A function can return `*DatabaseError`, `*ValidationError`, or `*NetworkError` all as `error`.

---

**Q481.** What is the proper way to cancel all goroutines in a group when one fails?

- a) Kill all goroutines with `runtime.Goexit()`
- b) Use `errgroup.WithContext` — the context is cancelled when any goroutine returns an error
- c) Use a global done channel and close it
- d) Use `os.Exit(1)` from the failing goroutine

> ✅ **b** — `errgroup.WithContext` returns a Context that is cancelled when any goroutine in the group returns non-nil error.

---

**Q482.** True or False: The blank identifier `_` is a valid variable that stores the discarded value.

> ✅ **False** — `_` is not a variable. It's a special "discard" identifier. You cannot read from `_`. Multiple assignments to `_` are all valid independently.

---

**Q483.** What happens with this code?
```go
var wg sync.WaitGroup
wg.Add(1)
go func() {
    panic("oops")
    wg.Done()
}()
wg.Wait()
```
- a) Prints "oops" and continues
- b) `wg.Done()` is never called — `wg.Wait()` blocks forever (and the program crashes from panic)
- c) The WaitGroup handles panics
- d) `wg.Add(-1)` is called automatically on panic

> ✅ **b** — The panic crashes the program before `wg.Done()` runs. Use `defer wg.Done()` to ensure it runs even on panic.

---

**Q484.** What is the `strings.Builder.Grow(n)` method for?

- a) Increases the string length by n
- b) Pre-allocates n bytes of capacity to avoid repeated re-allocations during WriteString
- c) Writes n zero bytes
- d) Limits maximum string size to n

> ✅ **b** — If you know the approximate final size, `b.Grow(n)` avoids multiple reallocations. Optimization.

---

**Q485.** True or False: `make([]int, 5)` and `make([]int, 0, 5)` result in slices with the same `cap` but different `len`.

> ✅ **True** — First: `len=5, cap=5` (5 zeros initialized). Second: `len=0, cap=5` (no elements yet, but space pre-allocated for 5).

---

**Q486.** What does it mean for an interface to be "satisfied implicitly" in Go?

- a) The interface is automatically satisfied by all types
- b) No explicit `implements` declaration — if a type has all the required methods, it satisfies the interface
- c) Implicit means the interface is optional
- d) Implicit satisfaction requires `reflect`

> ✅ **b** — This is Go's "duck typing" or structural typing. Contrast with Java's `implements Interface`.

---

**Q487.** When would you use `atomic.Value` over `sync.RWMutex` for a config struct?

- a) When config changes frequently
- b) When reads vastly outnumber writes and you want lock-free reads (no blocking at all)
- c) For single integers only
- d) Never — RWMutex is always better

> ✅ **b** — `atomic.Value.Load()` is entirely lock-free. `RWMutex.RLock` still requires a lock operation. For very high-frequency reads, atomic wins.

---

**Q488.** True or False: The `defer` keyword evaluates method receiver at defer time, not at execution time.

> ✅ **True** — Both the receiver and arguments are evaluated at the `defer` statement, not when the defer executes.

---

**Q489.** What is the `//nolint` comment used for?

- a) Native Go compiler directive to skip linting
- b) Suppresses linter warnings from tools like `golangci-lint` for known acceptable violations
- c) Disables the race detector for a specific line
- d) Marks code as safe for `unsafe` operations

> ✅ **b** — `//nolint:errcheck` suppresses specific linter rules. Use sparingly and with a comment explaining why.

---

**Q490.** What is the fastest way to check if a byte slice contains a specific byte?
- a) Convert to string and use `strings.Contains`
- b) `bytes.IndexByte(b, byte)` — returns index of first occurrence
- c) Loop manually
- d) `strings.Index(string(b), string([]byte{target}))`

> ✅ **b** — `bytes.IndexByte` is implemented in assembly on most platforms. Extremely fast.

---

**Q491.** True or False: You can have an anonymous struct in Go: `x := struct{ Name string }{"Alice"}`.

> ✅ **True** — Anonymous structs are valid and useful for one-off data structures in tests (table-driven test cases), JSON decoding of specific shapes, etc.

---

**Q492.** What is the recommended approach for database transactions in Go?

- a) Use `db.Begin()`, `tx.Exec()`, `tx.Commit()` or `tx.Rollback()` — always rollback on error
- b) Use global transactions
- c) Use `sql.DB` directly (transactions are automatic)
- d) Use `sync.Mutex` around database calls

> ✅ **a** — Pattern: `defer tx.Rollback()` immediately after Begin (no-op after Commit), then `tx.Commit()` at the end of success path.

---

**Q493.** True or False: `fmt.Println` and `fmt.Printf` are safe to call from multiple goroutines simultaneously.

> ✅ **True** — `fmt` package writes are individually atomic (a single Println call won't be interleaved with another). However, the ORDER of output between goroutines is non-deterministic.

---

**Q494.** What is the `expvar` package used for?

- a) Experimental variables (to be removed)
- b) Exporting runtime metrics via HTTP `/debug/vars` endpoint — counters, gauges, maps
- c) Environment variables management
- d) Expanded variable declarations

> ✅ **b** — Register counters, strings, or maps with `expvar.NewInt("requests")`. Accessible at `http://localhost:6060/debug/vars`.

---

**Q495.** True or False: In Go, `nil` maps and `nil` slices are treated identically in all operations.

> ✅ **False** — Both have `len() = 0` and range-safe. BUT: nil map panics on WRITE; nil slice is fine for append. `json.Marshal` difference: nil slice = "null", nil map = "null", empty slice = "[]", empty map = "{}".

---

**Q496.** What is `pprof` used for?

- a) Protocol Buffers profiling syntax
- b) Go's built-in profiling tool — CPU, memory, goroutine, block, mutex profiles
- c) A testing profiler for catching flaky tests
- d) An HTTP request profiler

> ✅ **b** — `import _ "net/http/pprof"` adds endpoints: `/debug/pprof/heap`, `/cpu`, `/goroutine`. Analyze with `go tool pprof`.

---

**Q497.** True or False: A `panic` in one goroutine can be recovered by `recover()` in a different goroutine.

> ✅ **False** — `recover()` only catches panics in the **same goroutine's** deferred functions. A panic in goroutine G2 terminates the ENTIRE program if not recovered within G2.

---

**Q498.** What is the proper way to write a Go API package for external consumption?

- a) Export everything — let users decide what to use
- b) Export minimal, well-documented interfaces. Keep implementation unexported. Use `internal/` for helpers. Version with semantic versioning.
- c) Use `interface{}` for all parameters for maximum flexibility
- d) Don't write packages — use a monolith

> ✅ **b** — Minimal surface area = easier to maintain. Follow Go API design principles: small interfaces, clear naming, explicit errors.

---

**Q499.** What does this line check at compile time?
```go
var _ http.Handler = (*MyServer)(nil)
```
- a) That `MyServer` is not nil at runtime
- b) That `*MyServer` implements the `http.Handler` interface — fails at compile time if not
- c) That `MyServer` implements `http.Handler` (value, not pointer)
- d) That `http.Handler` has been imported

> ✅ **b** — The canonical compile-time interface check. Zero runtime cost.

---

**Q500.** What is the single most important thing to remember about error handling in Go?

- a) Always use `panic` for unexpected errors
- b) Never ignore errors — every `(result, error)` return must have its `error` checked with `if err != nil`
- c) Wrap every error with a new error
- d) Use `log.Fatal` on all errors

> ✅ **b** — "Errors are values" — they must be handled explicitly. Unanswered errors hide bugs, cause silent data corruption, and make debugging impossible. The Go philosophy: explicit is better than implicit.

---

# 🏆 Quiz Complete — All 500 Questions

## Score Guide
| Score | Level |
|-------|-------|
| 450–500 | 🥇 Go Expert — Ready for senior interviews |
| 380–449 | 🥈 Go Proficient — Strong junior/mid-level |
| 280–379 | 🥉 Go Competent — Review weak areas |
| < 280 | 📚 Keep Studying — Re-read the relevant parts |

## Quick Review Map
| Q Range | Topic | Part |
|---------|-------|------|
| Q1–Q25 | Philosophy & Setup | Part 1 |
| Q26–Q50 | Variables & Types | Part 2 |
| Q51–Q75 | Control Flow | Part 3 |
| Q76–Q100 | Functions | Part 4 |
| Q101–Q125 | Pointers & Memory | Part 5 |
| Q126–Q150 | Slices & Maps | Part 6 |
| Q151–Q175 | Structs & Methods | Part 7 |
| Q176–Q200 | Interfaces | Part 8 |
| Q201–Q225 | Error Handling | Part 9 |
| Q226–Q250 | Goroutines | Part 10 |
| Q251–Q275 | Channels | Part 11 |
| Q276–Q300 | sync Package | Part 12 |
| Q301–Q325 | context.Context | Part 13 |
| Q326–Q350 | Generics | Part 14 |
| Q351–Q375 | Testing | Part 15 |
| Q376–Q400 | Standard Library | Part 16 |
| Q401–Q420 | HTTP | Part 17 |
| Q421–Q440 | C++ Migration | Part 18 |
| Q441–Q460 | Advanced Patterns | Part 19 |
| Q461–Q500 | Interview Prep | Part 20 |

Good luck with your interview! 🚀
