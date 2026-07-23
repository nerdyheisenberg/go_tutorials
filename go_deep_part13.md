# Go Deep Dive — Part 13: The context Package — Cancellation & Deadlines

---

## Chapter 1: Why context.Context Exists

Before `context`, Go programs had no standard way to:
1. Cancel a long-running operation
2. Set a deadline for an operation  
3. Pass request-scoped values (request ID, user auth) through call chains

The `context` package was added in Go 1.7 to solve all three.

### The Core Problem

```go
// Without context — how do you cancel this?
func fetchData(url string) ([]byte, error) {
    resp, err := http.Get(url)   // what if this takes 30 seconds? No cancel!
    if err != nil { return nil, err }
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)
}

// With context — full cancellation support
func fetchData(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil { return nil, err }
    resp, err := http.DefaultClient.Do(req)  // respects ctx cancellation/deadline
    if err != nil { return nil, err }
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)
}

// Callers can now control the lifecycle:
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
data, err := fetchData(ctx, "https://api.example.com/data")
```

---

## Chapter 2: The `Context` Interface

```go
// The full interface — only 4 methods!
type Context interface {
    // Deadline returns the time when work done on behalf of this context
    // should be canceled. Zero time means no deadline.
    Deadline() (deadline time.Time, ok bool)
    
    // Done returns a channel that's closed when the context is canceled
    // or its deadline elapses. If Done is nil, context can never be canceled.
    Done() <-chan struct{}
    
    // Err returns why this context was canceled.
    // Returns nil if Done is not yet closed.
    // Returns Canceled if Done is closed via cancel().
    // Returns DeadlineExceeded if Done is closed due to deadline.
    Err() error
    
    // Value returns the value associated with key, or nil.
    Value(key any) any
}
```

---

## Chapter 3: Context Types — When to Use Each

### `context.Background()` — The Root

```go
// Background is never canceled, has no deadline, has no values
// Use as the ROOT context for top-level operations

func main() {
    ctx := context.Background()
    
    // Derive child contexts from Background
    ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    runServer(ctx2)
}

// Also used in:
// - main() function
// - init() functions  
// - Tests at the top level
// - When you genuinely don't know what context to use
```

### `context.TODO()` — The Placeholder

```go
// TODO means "I'll add a real context here later"
// Use when refactoring and you haven't propagated context yet

func legacyFunction() {
    // TODO: accept ctx context.Context as first parameter
    doWork(context.TODO())  // placeholder — must replace with real context!
}

// Key difference:
// Background = "this is correct — there is no parent context"
// TODO = "this is incomplete — I need to wire up a real context"
```

### `context.WithCancel()` — Manual Cancellation

```go
func runWithCancel() {
    ctx, cancel := context.WithCancel(context.Background())
    
    // ALWAYS defer cancel to release resources!
    // Even if the context is already cancelled, calling cancel again is safe
    defer cancel()
    
    // Start workers that respect context
    go worker(ctx, "worker-1")
    go worker(ctx, "worker-2")
    
    // Do some work...
    time.Sleep(2 * time.Second)
    
    // Manually cancel — signals all goroutines watching ctx.Done()
    cancel()  // all workers see ctx.Done() close and return
}

func worker(ctx context.Context, name string) {
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("%s: stopping (%v)\n", name, ctx.Err())
            return
        default:
            doWork(name)
        }
    }
}
```

### `context.WithTimeout()` — Auto-Cancel After Duration

```go
func fetchWithTimeout(url string) ([]byte, error) {
    // Automatically cancelled after 5 seconds
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()  // clean up even if we return early
    
    return fetchData(ctx, url)
}

// Common pattern: accept parent context, add tighter timeout
func processItem(parent context.Context, item Item) error {
    // Each item gets at most 1 second, regardless of parent's remaining time
    ctx, cancel := context.WithTimeout(parent, 1*time.Second)
    defer cancel()
    return doProcess(ctx, item)
}
```

### `context.WithDeadline()` — Auto-Cancel at Specific Time

```go
func runUntilMidnight() error {
    // Calculate time until midnight
    now := time.Now()
    midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
    
    ctx, cancel := context.WithDeadline(context.Background(), midnight)
    defer cancel()
    
    return doWork(ctx)
}

// WithTimeout is just syntactic sugar for WithDeadline:
// context.WithTimeout(parent, d) ==
//     context.WithDeadline(parent, time.Now().Add(d))

// Getting the deadline:
if deadline, ok := ctx.Deadline(); ok {
    remaining := time.Until(deadline)
    fmt.Printf("Deadline in %v\n", remaining)
}
```

### `context.WithValue()` — Request-Scoped Data

```go
// WithValue attaches request-scoped data to contexts
// Use for data that CROSSES API boundaries, not for optional function params

// KEY RULE: Define your own key type to prevent collisions!
// NEVER use built-in types (string, int) as context keys

type contextKey string  // private type — prevents collision with other packages

const (
    requestIDKey contextKey = "requestID"
    userIDKey    contextKey = "userID"
    traceIDKey   contextKey = "traceID"
)

// Middleware that adds request ID to context
func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        reqID := generateRequestID()
        ctx := context.WithValue(r.Context(), requestIDKey, reqID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Helper functions for type-safe access (BEST PRACTICE)
func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(requestIDKey).(string)
    return id, ok
}

// Usage:
func handleRequest(ctx context.Context) {
    if reqID, ok := RequestIDFromContext(ctx); ok {
        log.Print("request:", reqID)
    }
}
```

---

## Chapter 4: Context Propagation — The Golden Rule

**The golden rule: context.Context must be the FIRST parameter of every function that crosses API boundaries or does I/O.**

```go
// CORRECT: ctx is first parameter
func GetUser(ctx context.Context, id int) (*User, error)
func ProcessOrder(ctx context.Context, order Order) error
func SendEmail(ctx context.Context, to, subject, body string) error
func QueryDB(ctx context.Context, query string, args ...any) (*sql.Rows, error)

// WRONG: ctx stored in struct (anti-pattern)
type Service struct {
    ctx context.Context  // NEVER store context in a struct!
    // ...
}

// WHY? Context represents a SINGLE request's lifetime.
// Structs represent PERSISTENT state that outlives requests.
// Storing context in a struct means the struct either:
// 1. Carries a stale/cancelled context for later requests
// 2. Or is incorrectly shared between requests

// Exception: if the struct IS the request itself
type Request struct {
    ctx context.Context
    // ... request-specific fields
}
func (r *Request) Context() context.Context { return r.ctx }
// This is what *http.Request does — it's the request OBJECT, not a service
```

### Context Tree

```go
// Contexts form a tree — cancelling a parent cancels all children

background              ← context.Background()
    |
    ├── mainCtx         ← context.WithCancel(background)
    │       |
    │       ├── reqCtx1 ← context.WithTimeout(mainCtx, 5s)
    │       └── reqCtx2 ← context.WithValue(mainCtx, key, val)
    │
    └── (other)

// Cancelling mainCtx automatically cancels reqCtx1 and reqCtx2
// Cancelling reqCtx1 does NOT cancel mainCtx or reqCtx2
```

---

## Chapter 5: Checking Context Cancellation

```go
// Pattern 1: At the start of every long operation
func longOperation(ctx context.Context) error {
    // Check before starting expensive work
    if ctx.Err() != nil {
        return ctx.Err()  // already cancelled — skip work
    }
    
    // ... do expensive work
    return nil
}

// Pattern 2: In loops — check each iteration
func processItems(ctx context.Context, items []Item) error {
    for _, item := range items {
        // Check before each item
        if err := ctx.Err(); err != nil {
            return fmt.Errorf("processItems: cancelled after %d items: %w", processed, err)
        }
        if err := process(ctx, item); err != nil {
            return err
        }
        processed++
    }
    return nil
}

// Pattern 3: select in loops — check while waiting
func listen(ctx context.Context, ch <-chan Message) error {
    for {
        select {
        case msg, ok := <-ch:
            if !ok {
                return nil  // channel closed
            }
            handle(msg)
        case <-ctx.Done():
            return ctx.Err()  // context cancelled
        }
    }
}

// Pattern 4: Non-blocking check with default
func tryWithContext(ctx context.Context) bool {
    select {
    case <-ctx.Done():
        return false  // context cancelled
    default:
        return true   // not cancelled
    }
}
```

---

## Chapter 6: Context in HTTP Servers

```go
package main

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
)

// http.Request already carries a context — use it!
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()  // get the request's context
    
    // The request context is automatically cancelled when:
    // - Client disconnects
    // - Request is done
    // - Server receives shutdown signal
    
    result, err := slowDatabaseQuery(ctx)
    if err != nil {
        if ctx.Err() != nil {
            // Client disconnected — don't bother sending response
            return
        }
        http.Error(w, "Internal error", 500)
        return
    }
    
    json.NewEncoder(w).Encode(result)
}

func slowDatabaseQuery(ctx context.Context) (interface{}, error) {
    // Pass context to ALL downstream calls
    rows, err := db.QueryContext(ctx, "SELECT * FROM large_table")
    // If client disconnects, ctx is cancelled, query is aborted!
    // ...
    return nil, nil
}

// Detecting client disconnect in streaming response
func streamHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }
    
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            // Client disconnected
            return
        case t := <-ticker.C:
            fmt.Fprintf(w, "data: %v\n\n", t)
            flusher.Flush()
        }
    }
}
```

---

## Chapter 7: errgroup — Structured Concurrency with Context

```go
// errgroup is context-aware concurrency — the modern replacement for WaitGroup+channels
// go get golang.org/x/sync/errgroup

import "golang.org/x/sync/errgroup"

func fetchMultiple(ctx context.Context, urls []string) ([][]byte, error) {
    g, ctx := errgroup.WithContext(ctx)  // New context that cancels if any goroutine errors
    
    results := make([][]byte, len(urls))
    
    for i, url := range urls {
        i, url := i, url
        g.Go(func() error {
            data, err := fetchData(ctx, url)
            if err != nil {
                return fmt.Errorf("fetch %s: %w", url, err)
            }
            results[i] = data
            return nil
        })
    }
    
    // Wait for ALL goroutines to complete
    // Returns first non-nil error; cancels context for others
    if err := g.Wait(); err != nil {
        return nil, err
    }
    return results, nil
}

// With concurrency limit
func fetchWithLimit(ctx context.Context, urls []string, limit int) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(limit)  // max N goroutines at once
    
    for _, url := range urls {
        url := url
        g.Go(func() error {
            return fetch(ctx, url)
        })
    }
    
    return g.Wait()
}
```

---

## Chapter 8: Comprehensive Context Tests

```go
package context_test

import (
    "context"
    "errors"
    "testing"
    "time"
)

// Test cancellation propagates to children
func TestCancellationPropagates(t *testing.T) {
    parent, cancel := context.WithCancel(context.Background())
    child, _ := context.WithCancel(parent)  // child of parent
    
    cancel()  // cancel parent
    
    // Child should also be cancelled
    select {
    case <-child.Done():
        // Expected
        if !errors.Is(child.Err(), context.Canceled) {
            t.Errorf("child.Err() = %v, want Canceled", child.Err())
        }
    case <-time.After(100 * time.Millisecond):
        t.Error("child should be cancelled when parent is cancelled")
    }
}

// Test cancelling child doesn't affect parent
func TestChildCancelDoesNotAffectParent(t *testing.T) {
    parent, parentCancel := context.WithCancel(context.Background())
    defer parentCancel()
    
    child, childCancel := context.WithCancel(parent)
    childCancel()  // cancel child
    
    // Parent should NOT be cancelled
    select {
    case <-parent.Done():
        t.Error("parent should not be cancelled when child is cancelled")
    default:
        // Expected — parent is still alive
    }
}

// Test WithTimeout
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    select {
    case <-ctx.Done():
        if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
            t.Errorf("ctx.Err() = %v, want DeadlineExceeded", ctx.Err())
        }
    case <-time.After(500 * time.Millisecond):
        t.Error("context should have timed out by now")
    }
}

// Test that calling cancel multiple times is safe
func TestCancelIdempotent(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    
    cancel()  // first cancel
    cancel()  // second cancel — should not panic
    cancel()  // third — also safe
    
    if ctx.Err() != context.Canceled {
        t.Errorf("ctx.Err() = %v, want Canceled", ctx.Err())
    }
}

// Test WithValue
func TestWithValue(t *testing.T) {
    type key string
    const k key = "testKey"
    
    ctx := context.WithValue(context.Background(), k, "testValue")
    
    // Correct key retrieves value
    val, ok := ctx.Value(k).(string)
    if !ok || val != "testValue" {
        t.Errorf("ctx.Value(k) = %v, want 'testValue'", ctx.Value(k))
    }
    
    // Wrong key returns nil
    if ctx.Value("wrongKey") != nil {
        t.Error("wrong key should return nil")
    }
    
    // Child inherits parent values
    child := context.WithValue(ctx, key("childKey"), "childVal")
    if child.Value(k).(string) != "testValue" {
        t.Error("child should inherit parent values")
    }
}

// Test context in a function that respects cancellation
func operationWithContext(ctx context.Context, duration time.Duration) error {
    timer := time.NewTimer(duration)
    defer timer.Stop()
    
    select {
    case <-timer.C:
        return nil  // completed normally
    case <-ctx.Done():
        return ctx.Err()  // cancelled
    }
}

func TestOperationRespectsCancellation(t *testing.T) {
    t.Run("completes before cancellation", func(t *testing.T) {
        ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
        defer cancel()
        
        err := operationWithContext(ctx, 100*time.Millisecond)
        if err != nil {
            t.Errorf("expected no error, got %v", err)
        }
    })
    
    t.Run("cancelled before completion", func(t *testing.T) {
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
        defer cancel()
        
        err := operationWithContext(ctx, 500*time.Millisecond)
        if !errors.Is(err, context.DeadlineExceeded) {
            t.Errorf("expected DeadlineExceeded, got %v", err)
        }
    })
    
    t.Run("manual cancellation", func(t *testing.T) {
        ctx, cancel := context.WithCancel(context.Background())
        
        // Cancel after 50ms
        go func() {
            time.Sleep(50 * time.Millisecond)
            cancel()
        }()
        
        err := operationWithContext(ctx, 500*time.Millisecond)
        if !errors.Is(err, context.Canceled) {
            t.Errorf("expected Canceled, got %v", err)
        }
    })
}

// Test Deadline
func TestDeadline(t *testing.T) {
    deadline := time.Now().Add(5 * time.Second)
    ctx, cancel := context.WithDeadline(context.Background(), deadline)
    defer cancel()
    
    d, ok := ctx.Deadline()
    if !ok {
        t.Error("expected ok=true for context with deadline")
    }
    if !d.Equal(deadline) {
        t.Errorf("deadline = %v, want %v", d, deadline)
    }
    
    // Background context has no deadline
    bgCtx := context.Background()
    _, ok2 := bgCtx.Deadline()
    if ok2 {
        t.Error("background context should have no deadline")
    }
}

// Benchmark context propagation
func BenchmarkContextValue(b *testing.B) {
    type key string
    ctx := context.WithValue(context.Background(), key("k"), "v")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = ctx.Value(key("k"))
    }
}
```

---

**Summary of Part 13:**
- `context.Context` solves cancellation, timeouts, and request-scoped data propagation
- Four creation functions: `Background()` (root), `TODO()` (placeholder), `WithCancel`, `WithTimeout/Deadline`, `WithValue`
- `ctx` must be the **first parameter** in any function doing I/O or crossing API boundaries
- **Never store context in a struct** — context represents a single request's lifetime
- Cancelling a parent automatically cancels all children (tree structure)
- Always `defer cancel()` immediately after `WithCancel/WithTimeout/WithDeadline`
- Use custom private key types for `WithValue` — prevents package collisions
- `ctx.Done()` is a `<-chan struct{}` — use in `select` for responsive cancellation
- `errgroup.WithContext` is the modern structured concurrency primitive
- `*http.Request.Context()` carries the request context — use it for all downstream calls
