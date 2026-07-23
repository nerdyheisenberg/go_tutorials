# Go Deep Dive — Part 12: The sync Package — Low-Level Synchronization

---

## Chapter 1: When to Use sync vs Channels

The Go FAQ says: "Use channels when data is passed between goroutines. Use sync primitives when protecting shared state."

| Situation | Use |
|---|---|
| Passing data ownership between goroutines | Channel |
| Signaling events (done, ready, shutdown) | Channel |
| Protecting a shared cache/map | `sync.Mutex` or `sync.RWMutex` |
| Ensuring code runs exactly once | `sync.Once` |
| Reusing pools of objects | `sync.Pool` |
| Waiting for multiple goroutines | `sync.WaitGroup` |
| Coordinating goroutine reads/writes | `sync.RWMutex` |
| Concurrent-safe map access | `sync.Map` |
| Low-level atomic operations | `sync/atomic` |

---

## Chapter 2: `sync.Mutex` — Mutual Exclusion

```go
// Mutex guarantees that only ONE goroutine executes the critical section at a time

type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()  // ALWAYS use defer for unlock — panic-safe
    c.count++
}

func (c *SafeCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}

// Critical rules:
// 1. NEVER copy a Mutex — copy a pointer to the struct instead
// 2. Lock and Unlock must be from the same goroutine (unlike std::mutex)
// 3. Mutexes are not reentrant — same goroutine locking twice = DEADLOCK
// 4. Always unlock (use defer unless performance is critical)
```

### Mutex Deadlock Patterns

```go
// DEADLOCK 1: Forgotten unlock
func (s *Store) badGet(key string) string {
    s.mu.Lock()
    defer s.mu.Unlock()
    val := s.data[key]
    if val == "" {
        // s.mu.Unlock() // programmer forgot this before 'defer' habit
        return s.fetchAndCache(key)   // tries to lock mu again inside → DEADLOCK!
    }
    return val
}

// FIX: Never call a function that locks the same mutex inside a locked section
// Or use a different (unexported) function that doesn't lock:
func (s *Store) fetchAndCacheUnlocked(key string) string {
    // assumes mu is ALREADY locked by caller
    val := fetchFromDB(key)
    s.data[key] = val
    return val
}

// DEADLOCK 2: Lock ordering inconsistency
func transfer(from, to *Account, amount float64) {
    from.mu.Lock()
    to.mu.Lock()    // What if another goroutine locks to first, then from?
    // ...         // DEADLOCK!
    to.mu.Unlock()
    from.mu.Unlock()
}

// FIX: Always lock in the same order — lock lower-ID account first
func transferSafe(a, b *Account, amount float64) {
    // Canonical ordering prevents deadlock
    first, second := a, b
    if a.ID > b.ID {
        first, second = b, a
    }
    first.mu.Lock()
    defer first.mu.Unlock()
    second.mu.Lock()
    defer second.mu.Unlock()
    // ...
}
```

### Mutex Performance — When it Matters

```go
// Mutex contention (many goroutines competing for the same lock) is expensive
// Strategies to reduce contention:

// 1. Sharding — multiple mutexes for different key ranges
type ShardedMap struct {
    shards [256]struct {
        sync.RWMutex
        data map[string]int
    }
}

func (m *ShardedMap) shard(key string) int {
    h := fnv.New32()
    h.Write([]byte(key))
    return int(h.Sum32() % 256)
}

func (m *ShardedMap) Get(key string) (int, bool) {
    s := m.shard(key)
    m.shards[s].RLock()
    defer m.shards[s].RUnlock()
    v, ok := m.shards[s].data[key]
    return v, ok
}

// 2. Lock-free reads with occasional writes
// → Use sync.RWMutex (see next section)

// 3. Copy-on-write (advanced)
// Keep an atomic pointer to an immutable snapshot
// Writers create a new snapshot and atomically replace the pointer
type COWMap struct {
    ptr atomic.Pointer[map[string]int]
}
```

---

## Chapter 3: `sync.RWMutex` — Reader-Writer Lock

```go
// RWMutex allows:
// - Multiple readers simultaneously
// - Only ONE writer (exclusive, blocks all readers and other writers)

type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()         // multiple goroutines can hold RLock simultaneously
    defer c.mu.RUnlock()
    val, ok := c.data[key]
    return val, ok
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()          // exclusive: blocks all readers AND writers
    defer c.mu.Unlock()
    c.data[key] = value
}

func (c *Cache) Delete(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.data, key)
}

// When is RWMutex better than Mutex?
// - Read-heavy workloads (>10x more reads than writes)
// - RWMutex has more overhead than Mutex for pure writes
// - With equal reads/writes, RWMutex can be SLOWER due to overhead!
// - Benchmark before choosing!
```

---

## Chapter 4: `sync.Once` — Guaranteed Single Execution

```go
// sync.Once ensures a function runs exactly once, even across goroutines
// Primary use case: lazy initialization of shared resources

type Database struct {
    conn *sql.DB
}

var (
    dbInstance *Database
    dbOnce     sync.Once
)

func GetDB() *Database {
    dbOnce.Do(func() {
        db, err := sql.Open("postgres", connectionString)
        if err != nil {
            panic("failed to connect to database: " + err.Error())
        }
        dbInstance = &Database{conn: db}
    })
    return dbInstance
}

// ALL goroutines calling GetDB() concurrently will:
// - One goroutine executes the initialization function
// - All others BLOCK until initialization completes
// - All get the same initialized instance

// IMPORTANT: If the function panics, the Once is considered "done"
// Future calls to Do() will NOT call fn again!
// If fn panics and leaves half-initialized state, you have a problem.
// Consider: recover in fn and set a sentinel error state.

var (
    globalConfig *Config
    configOnce   sync.Once
    configErr    error
)

func initConfig() {
    configOnce.Do(func() {
        globalConfig, configErr = loadConfig()
    })
}

func GetConfig() (*Config, error) {
    initConfig()
    return globalConfig, configErr
}
```

---

## Chapter 5: `sync.WaitGroup` — Waiting for Goroutines

```go
// WaitGroup: a counter that blocks Wait() until it reaches zero

var wg sync.WaitGroup

// Add BEFORE starting goroutine (critical!)
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(n int) {
        defer wg.Done()  // decrement when done
        process(n)
    }(i)
}

wg.Wait()  // block until counter = 0

// Common mistake: Add inside goroutine
for i := 0; i < 10; i++ {
    go func(n int) {
        wg.Add(1)    // BUG: goroutine might not run before Wait() returns!
        defer wg.Done()
        process(n)
    }(i)
}
wg.Wait()  // might return immediately!

// Advanced: WaitGroup can be reused after Wait() returns
// as long as Add() calls happen-before Wait() calls

// Pattern: Fire and collect multiple operations
type Result struct{ ID int; Data string; Err error }

func parallelFetch(ids []int) []Result {
    results := make([]Result, len(ids))
    var wg sync.WaitGroup
    
    for i, id := range ids {
        wg.Add(1)
        go func(i, id int) {
            defer wg.Done()
            data, err := fetch(id)
            results[i] = Result{ID: id, Data: data, Err: err}
        }(i, id)
    }
    
    wg.Wait()
    return results
}
```

---

## Chapter 6: `sync.Pool` — Reducing GC Pressure

```go
// sync.Pool is a cache of temporary objects that can be reused
// Use when you create many short-lived objects to reduce GC pressure
// GC can reclaim pool objects at any time — don't put things you need to keep!

var bufferPool = sync.Pool{
    New: func() interface{} {
        // Called when pool is empty — create a new object
        return &bytes.Buffer{}
    },
}

func processRequest(data []byte) string {
    // Get a buffer from the pool (or create a new one if empty)
    buf := bufferPool.Get().(*bytes.Buffer)
    
    // ALWAYS reset before use (pool objects may have leftover state)
    buf.Reset()
    
    // Use the buffer
    buf.Write(data)
    buf.WriteString("\n---\n")
    result := buf.String()
    
    // Return to pool for reuse
    bufferPool.Put(buf)
    
    return result
}

// Without pool: each request creates a new *bytes.Buffer → GC pressure
// With pool: buffers are reused → massive GC reduction under load

// Real benchmark difference:
// Without pool: 500ns/op, 1024 B/op, 2 allocs/op
// With pool:    180ns/op, 0 B/op,    0 allocs/op (from pool)

// Pool guidelines:
// - Only store pointer types (not values)
// - Reset before use (item might have leftover state)
// - Don't put items back if they're in a bad state
// - Don't rely on pool for correctness — it's a performance optimization only
```

---

## Chapter 7: `sync.Map` — Concurrent Map

```go
// sync.Map is optimized for two patterns:
// 1. Key-value pairs written ONCE and read many times
// 2. Multiple goroutines writing DIFFERENT keys (low contention)
//
// For general concurrent map access with mixed read/write, 
// sync.RWMutex + regular map is often better.

var m sync.Map

// Store
m.Store("key1", "value1")
m.Store(42, []int{1, 2, 3})  // any key/value types (interface{})

// Load
val, ok := m.Load("key1")
if ok {
    fmt.Println(val.(string))  // type assertion needed
}

// LoadOrStore — atomic: load existing OR store new
actual, loaded := m.LoadOrStore("key2", "new_value")
// loaded=true means "key2" already existed (actual is old value)
// loaded=false means we stored "new_value" (actual is the new value)

// LoadAndDelete — atomic load+delete
val2, ok2 := m.LoadAndDelete("key1")

// Delete
m.Delete("key1")

// Range — iterate (no guaranteed order)
m.Range(func(key, value interface{}) bool {
    fmt.Printf("%v: %v\n", key, value)
    return true  // return false to stop iteration
})

// sync.Map vs RWMutex+map benchmark:
// sync.Map excels when many goroutines read the same keys
// RWMutex+map often wins when writes are frequent
// → Always benchmark your actual workload!
```

---

## Chapter 8: `sync/atomic` — Lock-Free Operations

```go
package main

import (
    "fmt"
    "sync"
    "sync/atomic"
)

// Atomic operations are single-instruction, thread-safe
// Faster than mutex for simple scalar values

// Old style (Go < 1.19)
var counter int64

func increment() {
    atomic.AddInt64(&counter, 1)
}

func getCount() int64 {
    return atomic.LoadInt64(&counter)
}

// New style (Go 1.19+) — typed atomic types
var (
    requestCount atomic.Int64
    errorCount   atomic.Int64
    isShutdown   atomic.Bool
)

func handleRequest() {
    requestCount.Add(1)
    if isShutdown.Load() {
        errorCount.Add(1)
        return
    }
    // ... process request
}

// Atomic pointer — lock-free swappable value
type Config struct{ MaxConns int; Timeout time.Duration }

var currentConfig atomic.Pointer[Config]

func updateConfig(c *Config) {
    currentConfig.Store(c)  // atomic swap
}

func getConfig() *Config {
    return currentConfig.Load()  // atomic read
}
// This is the "read-copy-update" (RCU) pattern — zero-cost reads!

// Compare-and-Swap (CAS) — the building block of lock-free algorithms
var version atomic.Int64

func incrementVersion() bool {
    current := version.Load()
    // Only increment if version hasn't changed since we read it
    return version.CompareAndSwap(current, current+1)
}
```

---

## Chapter 9: `sync.Cond` — Condition Variable (Rare but Powerful)

```go
// sync.Cond is the Go equivalent of std::condition_variable
// Use when goroutines need to wait for a specific condition to become true

type BoundedQueue struct {
    mu       sync.Mutex
    cond     *sync.Cond
    items    []int
    capacity int
}

func NewBoundedQueue(cap int) *BoundedQueue {
    bq := &BoundedQueue{capacity: cap}
    bq.cond = sync.NewCond(&bq.mu)
    return bq
}

func (q *BoundedQueue) Enqueue(item int) {
    q.mu.Lock()
    defer q.mu.Unlock()
    
    // Wait while queue is full
    for len(q.items) >= q.capacity {
        q.cond.Wait()  // atomically unlocks mu and suspends goroutine
        // When woken, mu is re-locked before Wait returns
    }
    q.items = append(q.items, item)
    q.cond.Broadcast()  // wake ALL waiting goroutines (Signal wakes just one)
}

func (q *BoundedQueue) Dequeue() int {
    q.mu.Lock()
    defer q.mu.Unlock()
    
    // Wait while queue is empty
    for len(q.items) == 0 {
        q.cond.Wait()  // wait until not-empty
    }
    
    item := q.items[0]
    q.items = q.items[1:]
    q.cond.Broadcast()  // wake producers that might be waiting
    return item
}

// Why use for not if with Wait():
// Spurious wakeups: cond.Wait() can return without Signal/Broadcast
// Another goroutine might have consumed the item between Signal and this goroutine waking
// Always re-check the condition in a for loop!

// WHEN to use sync.Cond vs channels:
// sync.Cond: when the condition depends on complex state (not just data availability)
// channel: when you're passing data (most cases)
// sync.Cond: when you need Broadcast() (wake ALL waiters at once)
// channels are generally preferred — sync.Cond is low-level
```

---

## Chapter 10: Comprehensive sync Tests

```go
package sync_test

import (
    "sync"
    "sync/atomic"
    "testing"
    "time"
    "bytes"
)

// ==================== MUTEX TESTS ====================

func TestMutexProtectsData(t *testing.T) {
    var mu sync.Mutex
    count := 0
    var wg sync.WaitGroup
    
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            mu.Lock()
            count++
            mu.Unlock()
        }()
    }
    
    wg.Wait()
    if count != 1000 {
        t.Errorf("count = %d, want 1000 (data race!)", count)
    }
}

// Run with: go test -race ./... to confirm NO races

// ==================== RWMUTEX TESTS ====================

func TestRWMutexAllowsConcurrentReads(t *testing.T) {
    var mu sync.RWMutex
    var readersDone atomic.Int32
    
    // Multiple readers should proceed concurrently
    start := make(chan struct{})
    var wg sync.WaitGroup
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            <-start          // wait for signal
            mu.RLock()       // should not block other readers
            time.Sleep(50 * time.Millisecond)
            readersDone.Add(1)
            mu.RUnlock()
        }()
    }
    
    close(start)  // start all readers simultaneously
    
    startTime := time.Now()
    wg.Wait()
    elapsed := time.Since(startTime)
    
    // If readers ran concurrently: elapsed ≈ 50ms
    // If readers ran sequentially: elapsed ≈ 500ms
    if elapsed > 200*time.Millisecond {
        t.Errorf("concurrent reads took %v, expected ~50ms (serialized?)", elapsed)
    }
    if readersDone.Load() != 10 {
        t.Errorf("expected 10 readers to complete, got %d", readersDone.Load())
    }
}

// ==================== ONCE TESTS ====================

func TestOnceRunsExactlyOnce(t *testing.T) {
    var once sync.Once
    var count atomic.Int32
    var wg sync.WaitGroup
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            once.Do(func() { count.Add(1) })
        }()
    }
    
    wg.Wait()
    if count.Load() != 1 {
        t.Errorf("once.Do ran %d times, want 1", count.Load())
    }
}

// ==================== POOL TESTS ====================

func TestPoolReuse(t *testing.T) {
    allocations := atomic.Int32{}
    
    pool := sync.Pool{
        New: func() interface{} {
            allocations.Add(1)
            return new(bytes.Buffer)
        },
    }
    
    // First get: allocates
    b1 := pool.Get().(*bytes.Buffer)
    b1.Reset()
    pool.Put(b1)
    
    // Second get: might reuse (not guaranteed, but usually does in test)
    b2 := pool.Get().(*bytes.Buffer)
    b2.Reset()
    pool.Put(b2)
    
    // In practice, pool helps under load
    // The important test is: does Get() return a valid usable object?
    b3 := pool.Get().(*bytes.Buffer)
    b3.Reset()
    b3.WriteString("hello")
    if b3.String() != "hello" {
        t.Error("pool object should be usable")
    }
    pool.Put(b3)
}

// ==================== ATOMIC TESTS ====================

func TestAtomicCounter(t *testing.T) {
    var counter atomic.Int64
    var wg sync.WaitGroup
    
    const goroutines = 100
    const increments = 1000
    
    for i := 0; i < goroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < increments; j++ {
                counter.Add(1)
            }
        }()
    }
    
    wg.Wait()
    
    expected := int64(goroutines * increments)
    if got := counter.Load(); got != expected {
        t.Errorf("counter = %d, want %d", got, expected)
    }
}

func TestAtomicCAS(t *testing.T) {
    var val atomic.Int64
    val.Store(10)
    
    // CAS: swap only if current value matches
    swapped := val.CompareAndSwap(10, 20)
    if !swapped || val.Load() != 20 {
        t.Errorf("CAS should succeed: swapped=%v, val=%d", swapped, val.Load())
    }
    
    // CAS with wrong old value: should fail
    swapped = val.CompareAndSwap(10, 30)  // 10 != current 20
    if swapped || val.Load() != 20 {
        t.Errorf("CAS should fail: swapped=%v, val=%d", swapped, val.Load())
    }
}

// ==================== BENCHMARKS ====================

func BenchmarkMutexVsAtomic(b *testing.B) {
    b.Run("mutex", func(b *testing.B) {
        var mu sync.Mutex
        count := 0
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                mu.Lock()
                count++
                mu.Unlock()
            }
        })
    })
    
    b.Run("atomic", func(b *testing.B) {
        var count atomic.Int64
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                count.Add(1)
            }
        })
    })
}

func BenchmarkRWMutexReadHeavy(b *testing.B) {
    data := map[string]int{"key": 42}
    
    b.Run("mutex_read_heavy", func(b *testing.B) {
        var mu sync.Mutex
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                mu.Lock()
                _ = data["key"]
                mu.Unlock()
            }
        })
    })
    
    b.Run("rwmutex_read_heavy", func(b *testing.B) {
        var mu sync.RWMutex
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                mu.RLock()
                _ = data["key"]
                mu.RUnlock()
            }
        })
    })
}
```

Run: `go test -v -race -bench=. ./...`

---

**Summary of Part 12:**
- Mutex: mutual exclusion — only one goroutine in critical section; always use defer Unlock
- Never re-lock the same Mutex from the same goroutine (not reentrant) — deadlock!
- Lock ordering must be consistent across all goroutines — prevent lock-order deadlock
- RWMutex: many concurrent readers OR one exclusive writer; benchmarks before assuming it's faster
- sync.Once: guaranteed single execution — ideal for lazy singleton initialization
- sync.Pool: reuse short-lived objects; reduces GC pressure under load; always Reset before use
- sync.Map: optimized for read-heavy or per-key-write patterns; not a general concurrent map
- sync/atomic: fast lock-free operations for scalars; Go 1.19+ typed atomic types preferred
- sync.Cond: condition variable for complex state waiting; channels usually preferred
