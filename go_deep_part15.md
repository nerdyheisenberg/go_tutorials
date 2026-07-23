# Go Deep Dive — Part 15: Testing — Complete Guide

---

## Chapter 1: Go's Testing Philosophy

Go has testing built into the standard toolchain — no external framework needed. The `testing` package is opinionated:
1. Tests live in `_test.go` files in the same package
2. Test functions must be `func TestXxx(t *testing.T)`
3. Benchmark functions must be `func BenchmarkXxx(b *testing.B)`
4. Example functions must be `func ExampleXxx()`

Philosophy: **Tests are just Go code.** No DSL, no annotations, no magic. This makes tests:
- Easy to debug (set breakpoints, use `fmt.Println`)
- Easy to refactor (the IDE understands them)
- Fast to compile and run

---

## Chapter 2: Test Anatomy — Complete Reference

### Test Functions

```go
package mypackage_test  // external test package (can only use exported symbols)
// OR:
package mypackage       // same package (can access unexported symbols)

import (
    "testing"
    "fmt"
)

// Test function naming: TestXxx where Xxx starts with capital letter
func TestAdd(t *testing.T) {
    // t.Error/t.Errorf: mark test as failed, continue running
    // t.Fatal/t.Fatalf: mark test as failed, STOP running immediately
    // t.Log/t.Logf:     log message (only shown if test fails or -v flag)
    // t.Helper():       marks this function as a test helper (affects line reporting)
    // t.Skip/t.Skipf:   mark test as skipped
    // t.Parallel():     run this test in parallel with other parallel tests
    // t.Cleanup(fn):    register cleanup to run when test ends
    
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }
}
```

### Table-Driven Tests — The Standard Pattern

```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name      string
        a, b      float64
        want      float64
        wantErr   bool
    }{
        // Normal cases
        {"positive/positive", 10, 2, 5, false},
        {"negative/positive", -10, 2, -5, false},
        {"zero dividend", 0, 5, 0, false},
        
        // Boundary cases
        {"divide by 1", 7, 1, 7, false},
        {"divide by -1", 7, -1, -7, false},
        
        // Error cases
        {"divide by zero", 10, 0, 0, true},
        {"negative divide by zero", -5, 0, 0, true},
    }
    
    for _, tt := range tests {
        // Use t.Run for subtests — each gets its own pass/fail status
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()  // run subtests in parallel if safe
            
            got, err := Divide(tt.a, tt.b)
            
            if (err != nil) != tt.wantErr {
                t.Fatalf("Divide(%v, %v) error = %v, wantErr %v", tt.a, tt.b, err, tt.wantErr)
            }
            
            if !tt.wantErr {
                const eps = 1e-9
                diff := got - tt.want
                if diff < 0 { diff = -diff }
                if diff > eps {
                    t.Errorf("Divide(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
                }
            }
        })
    }
}
```

### Setup and Teardown

```go
// TestMain: setup before ALL tests, teardown after ALL tests
func TestMain(m *testing.M) {
    // Setup
    setupTestDatabase()
    
    // Run all tests
    exitCode := m.Run()
    
    // Teardown (always runs)
    teardownTestDatabase()
    
    // Must call os.Exit with the result code
    os.Exit(exitCode)
}

// t.Cleanup: per-test teardown
func TestWithCleanup(t *testing.T) {
    db := setupDB(t)
    t.Cleanup(func() {
        db.Close()  // called when test (and all subtests) finish
    })
    
    // Test code...
}

// Helper function for per-test setup with automatic cleanup
func newTestDB(t *testing.T) *sql.DB {
    t.Helper()  // Makes line numbers point to caller, not this helper
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("failed to open test db: %v", err)
    }
    t.Cleanup(func() { db.Close() })
    return db
}

// Subtests share parent's cleanup:
func TestSubtest(t *testing.T) {
    shared := createSharedResource()
    t.Cleanup(func() { shared.Cleanup() })  // runs after ALL subtests finish
    
    t.Run("case1", func(t *testing.T) {
        t.Cleanup(func() { /* runs after case1 only */ })
        use(shared)
    })
    
    t.Run("case2", func(t *testing.T) {
        use(shared)
    })
}
```

---

## Chapter 3: Subtests and Sub-benchmarks

```go
// Subtests enable:
// 1. Running individual tests: go test -run TestAdd/positive
// 2. Parallel subtests: t.Parallel() inside t.Run
// 3. Shared setup with per-case teardown

func TestHTTPHandler(t *testing.T) {
    // Setup shared once
    handler := NewUserHandler(mockDB)
    server := httptest.NewServer(handler)
    defer server.Close()
    
    t.Run("GET existing user", func(t *testing.T) {
        resp, err := http.Get(server.URL + "/users/1")
        if err != nil { t.Fatal(err) }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusOK {
            t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
        }
    })
    
    t.Run("GET nonexistent user", func(t *testing.T) {
        resp, err := http.Get(server.URL + "/users/999")
        if err != nil { t.Fatal(err) }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusNotFound {
            t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
        }
    })
    
    t.Run("POST create user", func(t *testing.T) {
        body := strings.NewReader(`{"name":"Alice","email":"alice@example.com"}`)
        resp, err := http.Post(server.URL+"/users", "application/json", body)
        if err != nil { t.Fatal(err) }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusCreated {
            t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
        }
    })
}

// Running individual subtests:
// go test -run TestHTTPHandler/GET_existing_user
// Note: spaces become underscores in test names for matching purposes
```

---

## Chapter 4: Benchmarks — Deep Dive

```go
// Benchmark naming: BenchmarkXxx
func BenchmarkAppend(b *testing.B) {
    // b.N is set by the testing framework — runs until timing stabilizes
    for i := 0; i < b.N; i++ {
        s := make([]int, 0)
        for j := 0; j < 1000; j++ {
            s = append(s, j)
        }
        _ = s
    }
}

// Including setup time:
func BenchmarkSort(b *testing.B) {
    // Reset timer AFTER setup (don't count setup time)
    data := generateData(10000)
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        // Make a copy so each iteration sorts unsorted data
        input := make([]int, len(data))
        copy(input, data)
        sort.Ints(input)
    }
}

// Sub-benchmarks: test different sizes
func BenchmarkSearch(b *testing.B) {
    for _, size := range []int{100, 1000, 10000, 100000} {
        data := generateSortedData(size)
        
        b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                sort.SearchInts(data, size/2)  // binary search
            }
        })
    }
}

// Parallel benchmarks:
func BenchmarkMapGet(b *testing.B) {
    m := make(map[int]int, 1000)
    for i := 0; i < 1000; i++ { m[i] = i * 2 }
    
    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            _ = m[i%1000]
            i++
        }
    })
}

// Memory benchmarks:
func BenchmarkAlloc(b *testing.B) {
    b.ReportAllocs()  // show allocations in benchmark output
    for i := 0; i < b.N; i++ {
        s := make([]int, 0, 100)
        for j := 0; j < 100; j++ {
            s = append(s, j)
        }
        _ = s
    }
}
```

### Running Benchmarks

```bash
# Run benchmarks (not tests)
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkSort ./...

# Run with memory stats
go test -bench=. -benchmem ./...

# Run for specific duration
go test -bench=. -benchtime=10s ./...

# Compare benchmarks (with benchstat tool):
go test -bench=. -count=5 ./... > old.txt
# make changes...
go test -bench=. -count=5 ./... > new.txt
benchstat old.txt new.txt
```

### Benchmark Output Interpretation

```
BenchmarkAppend-8   1000000   1245 ns/op   8192 B/op   12 allocs/op
     ^              ^          ^            ^            ^
     name           iterations time/op      bytes/op     allocs/op
     (-8 = 8 CPUs)
```

---

## Chapter 5: Test Helpers — The Right Way

```go
// t.Helper() makes error messages point to the CALLER, not the helper
func assertEqual(t *testing.T, got, want interface{}) {
    t.Helper()  // CRITICAL — without this, line numbers are wrong!
    if !reflect.DeepEqual(got, want) {
        t.Errorf("got %v, want %v", got, want)
    }
}

func assertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

func assertError(t *testing.T, err error) {
    t.Helper()
    if err == nil {
        t.Fatal("expected error, got nil")
    }
}

// Using testify (popular library — be aware of it for interviews):
// import "github.com/stretchr/testify/assert"
// assert.Equal(t, expected, actual)
// assert.NoError(t, err)
// assert.Error(t, err)
// assert.Contains(t, slice, element)

// Without testify (standard library only):
func assertContains[T comparable](t *testing.T, slice []T, elem T) {
    t.Helper()
    for _, v := range slice {
        if v == elem { return }
    }
    t.Errorf("%v does not contain %v", slice, elem)
}
```

---

## Chapter 6: Mocking — Manual and Generated

### Manual Mocking

```go
// Define the interface you want to mock
type UserStore interface {
    GetUser(ctx context.Context, id int) (*User, error)
    SaveUser(ctx context.Context, u *User) error
    DeleteUser(ctx context.Context, id int) error
}

// Manual mock — gives you full control for tests
type MockUserStore struct {
    GetUserFunc    func(ctx context.Context, id int) (*User, error)
    SaveUserFunc   func(ctx context.Context, u *User) error
    DeleteUserFunc func(ctx context.Context, id int) error
    
    // Record calls for assertions
    GetUserCalls    []int
    SaveUserCalls   []*User
    DeleteUserCalls []int
}

func (m *MockUserStore) GetUser(ctx context.Context, id int) (*User, error) {
    m.GetUserCalls = append(m.GetUserCalls, id)
    if m.GetUserFunc != nil {
        return m.GetUserFunc(ctx, id)
    }
    return nil, nil
}

func (m *MockUserStore) SaveUser(ctx context.Context, u *User) error {
    m.SaveUserCalls = append(m.SaveUserCalls, u)
    if m.SaveUserFunc != nil {
        return m.SaveUserFunc(ctx, u)
    }
    return nil
}

func (m *MockUserStore) DeleteUser(ctx context.Context, id int) error {
    m.DeleteUserCalls = append(m.DeleteUserCalls, id)
    if m.DeleteUserFunc != nil {
        return m.DeleteUserFunc(ctx, id)
    }
    return nil
}

// Test using the mock:
func TestUserService_GetUser(t *testing.T) {
    t.Run("user found", func(t *testing.T) {
        store := &MockUserStore{
            GetUserFunc: func(_ context.Context, id int) (*User, error) {
                if id == 42 {
                    return &User{ID: 42, Name: "Alice"}, nil
                }
                return nil, ErrNotFound
            },
        }
        svc := NewUserService(store)
        
        user, err := svc.GetUser(context.Background(), 42)
        assertNoError(t, err)
        if user.Name != "Alice" { t.Errorf("name = %q, want Alice", user.Name) }
        if len(store.GetUserCalls) != 1 { t.Error("GetUser should be called once") }
    })
    
    t.Run("user not found", func(t *testing.T) {
        store := &MockUserStore{
            GetUserFunc: func(_ context.Context, id int) (*User, error) {
                return nil, ErrNotFound
            },
        }
        svc := NewUserService(store)
        
        _, err := svc.GetUser(context.Background(), 999)
        assertError(t, err)
        if !errors.Is(err, ErrNotFound) {
            t.Errorf("error = %v, want ErrNotFound", err)
        }
    })
}
```

---

## Chapter 7: HTTP Testing — httptest Package

```go
package handlers_test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

// Testing an HTTP handler WITHOUT starting a real server
func TestGetUserHandler(t *testing.T) {
    handler := func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]string{"name": "Alice"})
    }
    
    // Create a request
    req := httptest.NewRequest("GET", "/users/1", nil)
    req.Header.Set("Accept", "application/json")
    
    // Create a response recorder
    w := httptest.NewRecorder()
    
    // Call the handler directly
    handler(w, req)
    
    // Check the response
    resp := w.Result()
    if resp.StatusCode != http.StatusOK {
        t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
    }
    
    var body map[string]string
    json.NewDecoder(resp.Body).Decode(&body)
    if body["name"] != "Alice" {
        t.Errorf("name = %q, want Alice", body["name"])
    }
}

// Testing with a real (in-process) HTTP server
func TestUserAPIIntegration(t *testing.T) {
    // Create your router/handler
    mux := http.NewServeMux()
    mux.HandleFunc("/users/", handleGetUser)
    
    // Start a test server (same process, random port)
    server := httptest.NewServer(mux)
    defer server.Close()  // shuts down after test
    
    // Make HTTP request to the test server
    resp, err := http.Get(server.URL + "/users/1")
    if err != nil { t.Fatal(err) }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
    }
}

// Testing with TLS
func TestUserAPITLS(t *testing.T) {
    server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()
    
    // server.Client() returns an http.Client configured to trust the test TLS cert
    client := server.Client()
    resp, err := client.Get(server.URL)
    if err != nil { t.Fatal(err) }
    if resp.StatusCode != http.StatusOK { t.Errorf("status = %d", resp.StatusCode) }
}
```

---

## Chapter 8: Golden Files — Testing Complex Output

```go
// Golden file tests: compare output against a "golden" reference file
// Useful for testing complex output (JSON, HTML, generated code)

func TestRenderTemplate(t *testing.T) {
    input := TemplateInput{Name: "Rohit", Items: []string{"Go", "C++"}}
    
    got, err := RenderTemplate(input)
    if err != nil { t.Fatal(err) }
    
    goldenFile := "testdata/render_template.golden"
    
    // Update golden file when output changes (run with -update flag)
    if *update {
        os.WriteFile(goldenFile, []byte(got), 0644)
    }
    
    want, err := os.ReadFile(goldenFile)
    if err != nil { t.Fatal(err) }
    
    if got != string(want) {
        t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
        // Or for structured diff:
        diff := computeDiff(string(want), got)
        t.Errorf("diff (-want +got):\n%s", diff)
    }
}

var update = flag.Bool("update", false, "update golden files")
```

---

## Chapter 9: Running Tests — Complete Command Reference

```bash
# Run all tests
go test ./...

# Run with verbose output (see each test name)
go test -v ./...

# Run specific test function
go test -run TestAdd ./...

# Run tests matching pattern (regex)
go test -run "TestUser.*" ./...

# Run subtests
go test -run "TestHTTPHandler/GET" ./...

# Run with race detector (ALWAYS use in CI!)
go test -race ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # open in browser

# Run benchmarks 
go test -bench=. ./...
go test -bench=BenchmarkSort -benchmem ./...

# Run tests with timeout
go test -timeout 30s ./...

# Run tests N times (for flaky test detection)
go test -count=10 ./...

# Skip caching (always re-run)
go test -count=1 ./...

# Run with multiple CPUs
go test -cpu=1,2,4,8 ./...

# Short mode (skip slow tests)
go test -short ./...

# Parallel test count
go test -parallel=4 ./...

# Build only (no run)
go test -build ./...
```

### Controlling Test Execution

```go
// Skip slow tests in short mode
func TestSlowOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping slow test in short mode")
    }
    // ... slow test
}

// Skip on specific platforms
func TestLinuxOnly(t *testing.T) {
    if runtime.GOOS != "linux" {
        t.Skipf("skipping on %s", runtime.GOOS)
    }
    // ... linux-specific test
}

// Run subtest pattern matching
func TestSuite(t *testing.T) {
    cases := []struct{ name string }{ /* ... */ }
    for _, tc := range cases {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()  // subtests run in parallel
        })
    }
}
```

---

## Chapter 10: Example Functions — Runnable Documentation

```go
// Example functions:
// 1. Appear in godoc as runnable examples
// 2. Are compiled and run during 'go test'
// 3. Output comment is checked against actual output

// Example of a function
func ExampleAdd() {
    result := Add(2, 3)
    fmt.Println(result)
    // Output:
    // 5
}

// Example of a method
func ExampleStack_Push() {
    s := &Stack[int]{}
    s.Push(1)
    s.Push(2)
    s.Push(3)
    v, _ := s.Pop()
    fmt.Println(v)
    // Output:
    // 3
}

// Unordered output (for maps, goroutines)
func ExampleKeys() {
    m := map[string]int{"a": 1, "b": 2}
    keys := Keys(m)
    sort.Strings(keys)  // sort for deterministic output
    fmt.Println(keys)
    // Output:
    // [a b]
}

// Example without output check (for functions with no deterministic output)
func ExampleGenerateID() {
    id := GenerateID()
    fmt.Println(len(id) > 0)  // just verify it's non-empty
    // Output:
    // true
}
```

---

**Summary of Part 15:**
- Tests are Go code in `_test.go` files — no external framework needed
- `TestXxx(t *testing.T)` — test functions; `BenchmarkXxx(b *testing.B)` — benchmarks
- Table-driven tests with `t.Run` subtests are the idiomatic pattern
- `t.Helper()` is critical in test helpers — makes error lines point to the actual caller
- `t.Fatal/t.Fatalf` — stop current test immediately; `t.Error/t.Errorf` — continue
- `t.Cleanup(fn)` — registers teardown that runs after the test, even on failure
- `TestMain(m *testing.M)` — global setup/teardown for the entire package
- `httptest.NewRecorder()` for unit testing handlers; `httptest.NewServer()` for integration
- `-race` flag detects data races— always run in CI
- `-coverprofile` produces coverage data; `go tool cover -html` renders it visually
- `go test -count=1` disables caching — use when tests have external state
