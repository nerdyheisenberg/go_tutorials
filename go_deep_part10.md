# Go Deep Dive — Part 10: Goroutines & The Go Scheduler

---

## Chapter 1: What Is a Goroutine?

A goroutine is **not** an OS thread. It's a lightweight unit of concurrent execution managed entirely by the Go runtime.

### Goroutine vs Thread Comparison

| | OS Thread | Goroutine |
|---|---|---|
| Stack size (initial) | ~1-8 MB (fixed) | ~2 KB (growable) |
| Creation cost | ~1ms, ~140 system calls | ~~100ns, no system calls |
| Max concurrent | Thousands (limited by OS) | Hundreds of thousands |
| Scheduling | OS kernel (preemptive) | Go runtime (cooperative + preemptive) |
| Context switch cost | ~1-2μs | ~100-200ns |
| Blocked I/O | Thread sleeps, OS wakes it | Runtime parks goroutine, runs another |

```go
// A Go program can run 100,000 goroutines easily
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    const n = 100_000
    var wg sync.WaitGroup
    
    start := time.Now()
    for i := 0; i < n; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            time.Sleep(1 * time.Second)  // simulate waiting
        }()
    }
    
    wg.Wait()
    fmt.Printf("100,000 goroutines completed in %v\n", time.Since(start))
    // Typically: ~1 second total (all waiting concurrently!)
    // With threads: would need 100,000 OS threads — out of memory
}
```

---

## Chapter 2: The Go Scheduler — GMP Model

Go uses an **M:N scheduler** — M goroutines are multiplexed onto N OS threads:

```
GMP Model:
  G = Goroutine (the user-level task)
  M = Machine (OS thread that actually runs code)
  P = Processor (owns a run queue of Goroutines)

                    ┌─────────┐  ┌─────────┐
                    │    P    │  │    P    │
                    │ [G G G] │  │ [G G G] │   ← Local run queues
                    └────┬────┘  └────┬────┘
                         │            │
                      ┌──┴──┐    ┌──┴──┐
                      │  M  │    │  M  │        ← OS Threads
                      └─────┘    └─────┘
                      
              Global Run Queue: [G G G G G...]
              
GOMAXPROCS = number of P's = number of CPU cores (by default)
```

### How Scheduling Works

```go
// When a goroutine blocks (e.g., waiting for I/O):
// 1. The runtime detects the block
// 2. The M is "detached" from P and handed to the OS for blocking
// 3. P finds another M (or creates one) to continue running
// 4. When I/O completes, the goroutine is put back in a run queue
// 5. When an M is available, it picks up the goroutine and continues

// Goroutine preemption (since Go 1.14):
// Goroutines are preempted at function call boundaries AND at loop iterations
// (previously only at function calls — tight loops could monopolize a P)

// You can see the scheduler in action:
import "runtime"

runtime.GOMAXPROCS(4)    // Use 4 OS threads (default: number of CPU cores)
runtime.NumCPU()         // Query number of CPU cores
runtime.NumGoroutine()   // Current number of goroutines
runtime.Gosched()        // Voluntarily yield the current goroutine's P
```

---

## Chapter 3: Starting and Managing Goroutines

### The `go` Keyword

```go
// Start a goroutine with any function call
go fmt.Println("concurrent!")   // anonymous function call
go someFunction(arg1, arg2)     // named function

// Most common: anonymous function
go func() {
    // ... do something concurrently
}()
// The () at the end immediately calls the anonymous function

// goroutines receive copies of arguments at call time
value := 10
go func(v int) {
    fmt.Println(v)  // uses copy of value
}(value)
value = 20  // doesn't affect the goroutine
```

### sync.WaitGroup — Waiting for Goroutines

```go
// WaitGroup solves: "how do I know when all goroutines are done?"

func downloadAll(urls []string) []string {
    results := make([]string, len(urls))
    
    var wg sync.WaitGroup
    
    for i, url := range urls {
        wg.Add(1)              // increment BEFORE starting goroutine
        
        go func(i int, url string) {
            defer wg.Done()    // decrement when goroutine exits
            
            // do the work
            results[i] = download(url)  // OK: each goroutine writes to different index
        }(i, url)
    }
    
    wg.Wait()  // block until counter reaches 0
    return results
}

// Common mistake: calling wg.Add inside the goroutine
// BAD:
for i, url := range urls {
    go func(i int, url string) {
        wg.Add(1)  // WRONG: goroutine might not start before wg.Wait()
        defer wg.Done()
        // ...
    }(i, url)
}
wg.Wait()  // might return immediately if goroutines haven't run yet!
```

### Goroutine Lifecycle Patterns

```go
// Pattern 1: Fire and forget (careful with goroutine leaks!)
func process(item Item) {
    go func() {
        sendNotification(item)  // run in background, don't wait
    }()
    // But: if sendNotification panics, it crashes the program!
    // And: goroutine might outlive the caller — is that OK?
}

// Pattern 2: Managed with WaitGroup
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    process()
}()
wg.Wait()

// Pattern 3: Managed with channel signal
done := make(chan struct{})
go func() {
    defer close(done)
    process()
}()
<-done  // wait for completion

// Pattern 4: With context for cancellation
func startWorker(ctx context.Context) <-chan Result {
    results := make(chan Result)
    go func() {
        defer close(results)
        for {
            select {
            case <-ctx.Done():
                return  // stop when context cancelled
            default:
                r, err := doWork()
                if err != nil {
                    return
                }
                results <- r
            }
        }
    }()
    return results
}
```

---

## Chapter 4: Goroutine Leaks — Detection and Prevention

A **goroutine leak** is when a goroutine runs forever, never returning, consuming memory and CPU.

```go
// LEAK 1: Goroutine blocked on channel that never receives/sends
func badServer() {
    for {
        conn, err := listener.Accept()
        if err != nil { return }
        
        go func() {
            results <- processConnection(conn)  // blocks if nobody reads results!
        }()
        // If results channel is full and nobody drains it, goroutines pile up → leak!
    }
}

// FIX: Use buffered channel or select with default
go func() {
    select {
    case results <- processConnection(conn):
    case <-time.After(30 * time.Second):
        log.Println("timed out sending result")
    }
}()

// LEAK 2: Goroutine waiting forever on a channel with no writer
func subscribe() <-chan Event {
    ch := make(chan Event)
    go func() {
        // What if the producer never sends? This goroutine leaks!
        for e := range producer.Events() {
            ch <- e
        }
    }()
    return ch
}

// FIX: Context cancellation
func subscribe(ctx context.Context) <-chan Event {
    ch := make(chan Event)
    go func() {
        defer close(ch)
        for {
            select {
            case <-ctx.Done():
                return  // properly exit when cancelled
            case e := <-producer.Events():
                ch <- e
            }
        }
    }()
    return ch
}

// LEAK 3: Goroutine panics silently (program continues with goroutine gone)
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("goroutine panicked: %v", r)
        }
    }()
    riskyOperation()
}()
```

### Detecting Goroutine Leaks in Tests

```go
// Using goleak library:
// go get go.uber.org/goleak

import "go.uber.org/goleak"

func TestNoGoroutineLeak(t *testing.T) {
    defer goleak.VerifyNone(t)  // fails if goroutines are still running after test
    
    // your test code...
    go backgroundWork()
    
    // If backgroundWork goroutine doesn't exit, goleak.VerifyNone fails!
}

// Manual goroutine count check:
func TestGoroutineCount(t *testing.T) {
    before := runtime.NumGoroutine()
    
    // ... code that might leak goroutines ...
    
    time.Sleep(100 * time.Millisecond) // give goroutines time to finish
    
    after := runtime.NumGoroutine()
    if after > before {
        t.Errorf("goroutine leak: started with %d, now have %d", before, after)
    }
}
```

---

## Chapter 5: Common Goroutine Patterns

### Pattern 1: Parallel Execution

```go
// Execute multiple tasks in parallel and wait for all
func parallel(tasks ...func() error) error {
    errs := make(chan error, len(tasks))
    
    for _, task := range tasks {
        task := task  // capture (not needed in Go 1.22+)
        go func() {
            errs <- task()
        }()
    }
    
    var firstErr error
    for range tasks {
        if err := <-errs; err != nil && firstErr == nil {
            firstErr = err
        }
    }
    return firstErr
}

// Usage:
err := parallel(
    func() error { return fetchUserData() },
    func() error { return fetchOrderData() },
    func() error { return fetchInventory() },
)
// All three run concurrently!
```

### Pattern 2: Background Worker

```go
type Worker struct {
    jobs chan Job
    done chan struct{}
    wg   sync.WaitGroup
}

func NewWorker(concurrency int) *Worker {
    w := &Worker{
        jobs: make(chan Job, 100),
        done: make(chan struct{}),
    }
    
    for i := 0; i < concurrency; i++ {
        w.wg.Add(1)
        go func() {
            defer w.wg.Done()
            for {
                select {
                case job, ok := <-w.jobs:
                    if !ok {
                        return  // channel closed
                    }
                    job.Process()
                case <-w.done:
                    return  // shutdown signal
                }
            }
        }()
    }
    
    return w
}

func (w *Worker) Submit(job Job) {
    w.jobs <- job  // blocks if queue is full (back-pressure)
}

func (w *Worker) Shutdown() {
    close(w.jobs)  // signal workers: no more jobs
    w.wg.Wait()    // wait for all workers to finish
}
```

### Pattern 3: Periodic Tasks

```go
func startPeriodicCleanup(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    
    go func() {
        defer ticker.Stop()  // important! Otherwise timer goroutine leaks
        
        for {
            select {
            case <-ticker.C:
                if err := cleanup(); err != nil {
                    log.Printf("cleanup failed: %v", err)
                }
            case <-ctx.Done():
                log.Println("cleanup stopped")
                return
            }
        }
    }()
}

// Usage:
ctx, cancel := context.WithCancel(context.Background())
startPeriodicCleanup(ctx, 5*time.Minute)
// Later:
cancel()  // stops the cleanup goroutine
```

### Pattern 4: Timeout with Goroutine

```go
func withTimeout(timeout time.Duration, fn func() (Result, error)) (Result, error) {
    type outcome struct {
        result Result
        err    error
    }
    
    ch := make(chan outcome, 1)  // buffered! goroutine doesn't block if we timeout
    
    go func() {
        result, err := fn()
        ch <- outcome{result, err}
    }()
    
    select {
    case out := <-ch:
        return out.result, out.err
    case <-time.After(timeout):
        return Result{}, fmt.Errorf("operation timed out after %v", timeout)
    }
    // NOTE: When we timeout, the goroutine is still running!
    // Only acceptable if fn() will eventually complete and clean up itself.
    // For proper cancellation, use context.WithTimeout instead.
}
```

---

## Chapter 6: Data Races — Detection and Prevention

A **data race** occurs when two goroutines access the same memory location concurrently, and at least one access is a write.

```go
// RACE: two goroutines accessing counter without synchronization
var counter int

go func() { counter++ }()  // read + increment + write (3 operations, not atomic!)
go func() { counter++ }()

// Result is undefined! Could be 1 or 2.
// With Go's race detector: "DATA RACE"

// FIX 1: Mutex
var mu sync.Mutex
var counter2 int

go func() {
    mu.Lock()
    counter2++
    mu.Unlock()
}()

// FIX 2: Atomic operation
var counter3 atomic.Int64

go func() { counter3.Add(1) }()

// FIX 3: Channel
counter4Ch := make(chan int, 1)
counter4Ch <- 0

go func() {
    v := <-counter4Ch
    counter4Ch <- v + 1
}()

// RUN THE RACE DETECTOR! It finds races at runtime:
// go test -race ./...
// go run -race main.go
```

---

## Chapter 7: Testing Goroutines

```go
package goroutines_test

import (
    "context"
    "runtime"
    "sync"
    "sync/atomic"
    "testing"
    "time"
)

// Test concurrent increment with race detection
func TestConcurrentIncrement(t *testing.T) {
    const goroutines = 100
    const incrementsPerGoroutine = 1000
    
    var counter atomic.Int64
    var wg sync.WaitGroup
    
    for i := 0; i < goroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < incrementsPerGoroutine; j++ {
                counter.Add(1)
            }
        }()
    }
    
    wg.Wait()
    
    expected := int64(goroutines * incrementsPerGoroutine)
    got := counter.Load()
    if got != expected {
        t.Errorf("counter = %d, want %d (data race?)", got, expected)
    }
}

// Run with: go test -race ./... to detect races

// Test goroutine doesn't leak
func TestNoGoroutineLeak(t *testing.T) {
    before := runtime.NumGoroutine()
    
    ctx, cancel := context.WithCancel(context.Background())
    
    // Start goroutine that responds to cancellation
    done := make(chan struct{})
    go func() {
        defer close(done)
        <-ctx.Done()
    }()
    
    cancel()
    <-done
    
    // Give scheduler time to clean up
    time.Sleep(10 * time.Millisecond)
    runtime.Gosched()
    
    after := runtime.NumGoroutine()
    if after > before+1 {  // +1 for test runner variance
        t.Errorf("goroutine leak: before=%d, after=%d", before, after)
    }
}

// Test WaitGroup correctness
func TestWaitGroup(t *testing.T) {
    var wg sync.WaitGroup
    var results []int
    var mu sync.Mutex  // protect results slice
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            mu.Lock()
            results = append(results, n)
            mu.Unlock()
        }(i)
    }
    
    wg.Wait()
    
    if len(results) != 10 {
        t.Errorf("expected 10 results, got %d", len(results))
    }
    
    // Results should contain all numbers 0-9 (in any order)
    seen := make(map[int]bool)
    for _, v := range results {
        if seen[v] { t.Errorf("duplicate value: %d", v) }
        seen[v] = true
    }
}

// Test timeout behavior
func TestWithTimeout(t *testing.T) {
    t.Run("completes before timeout", func(t *testing.T) {
        ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
        defer cancel()
        
        done := make(chan struct{})
        go func() {
            time.Sleep(100 * time.Millisecond)  // fast
            close(done)
        }()
        
        select {
        case <-done:
            // Success
        case <-ctx.Done():
            t.Error("should not have timed out")
        }
    })
    
    t.Run("times out before completion", func(t *testing.T) {
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
        defer cancel()
        
        done := make(chan struct{})
        go func() {
            time.Sleep(1 * time.Second) // slow
            close(done)
        }()
        
        select {
        case <-done:
            t.Error("should have timed out")
        case <-ctx.Done():
            // Expected timeout
            if ctx.Err() != context.DeadlineExceeded {
                t.Errorf("expected DeadlineExceeded, got %v", ctx.Err())
            }
        }
    })
}

// Test periodic task
func TestPeriodicTask(t *testing.T) {
    var callCount atomic.Int32
    
    ctx, cancel := context.WithTimeout(context.Background(), 550*time.Millisecond)
    defer cancel()
    
    ticker := time.NewTicker(100 * time.Millisecond)
    go func() {
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                callCount.Add(1)
            case <-ctx.Done():
                return
            }
        }
    }()
    
    <-ctx.Done()
    time.Sleep(50 * time.Millisecond)  // let goroutine finish
    
    count := callCount.Load()
    if count < 4 || count > 6 {
        t.Errorf("expected ~5 ticks in 550ms, got %d", count)
    }
}

// Benchmark goroutine creation cost
func BenchmarkGoroutineCreation(b *testing.B) {
    done := make(chan struct{})
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        go func() {
            done <- struct{}{}
        }()
        <-done
    }
}

func BenchmarkParallelVsSerial(b *testing.B) {
    work := func() int {
        sum := 0
        for i := 0; i < 1000; i++ {
            sum += i
        }
        return sum
    }
    
    b.Run("serial", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            work()
            work()
            work()
            work()
        }
    })
    
    b.Run("parallel", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            var wg sync.WaitGroup
            for j := 0; j < 4; j++ {
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    work()
                }()
            }
            wg.Wait()
        }
    })
}
```

---

**Summary of Part 10:**
- Goroutines are ~2KB lightweight, managed by Go's GMP scheduler — NOT OS threads
- GOMAXPROCS = number of P's = number of CPU cores (parallel execution)
- When goroutine blocks on I/O, the M is handed to OS, P picks up another M
- WaitGroup: Add(1) BEFORE go, Done() deferred inside goroutine
- Goroutine leak: goroutine stuck blocking forever — use context cancellation to prevent
- Data race: concurrent unsynchronized access — use `-race` flag to detect
- Always close channels from the sender, not the receiver
- Use `context.Context` for cancellation propagation through goroutine trees
- Prefer `sync/atomic` for single-value counters (faster than mutex)
