# Complete Go Programming Guide — Part 3: Concurrency

---

# Chapter 13: Goroutines

Goroutines are the foundation of Go's concurrency model. They are lightweight, user-space threads managed by the Go runtime scheduler — NOT OS threads.

## Goroutine vs OS Thread
- **OS Thread**: ~1MB stack, expensive to create, managed by the OS kernel
- **Goroutine**: ~2KB initial stack (grows dynamically), cheap to create, managed by Go runtime
- Go uses M:N scheduling: M goroutines are multiplexed onto N OS threads

You can easily run **hundreds of thousands** of goroutines in a single program.

## Starting Goroutines

```go
package main

import (
    "fmt"
    "time"
)

func sayHello(name string) {
    for i := 0; i < 3; i++ {
        fmt.Printf("Hello from %s (%d)\n", name, i)
        time.Sleep(100 * time.Millisecond)
    }
}

func main() {
    // Start goroutines with the 'go' keyword
    start := time.Now()
    fmt.Println(start)
    go sayHello("goroutine-1")
    go sayHello("goroutine-2")
    go sayHello("goroutine-3")

    // IMPORTANT: main() is itself a goroutine.
    // If main() exits, ALL goroutines are killed immediately!
    time.Sleep(1 * time.Second) // bad way to wait — use sync.WaitGroup
    fmt.Println("Main done")
    fmt.Println(time.Since(start))
}
```

## sync.WaitGroup (Proper Way to Wait)

```go
package main

import (
    "fmt"
    "sync"
)

func worker(id int, wg *sync.WaitGroup) {
    defer wg.Done() // Decrement counter when function returns
    fmt.Printf("Worker %d starting\n", id)
    // ... do work
    fmt.Printf("Worker %d done\n", id)
}

func main() {
    var wg sync.WaitGroup

    for i := 1; i <= 5; i++ {
        wg.Add(1) // Increment counter
        go worker(i, &wg)
    }

    wg.Wait() // Block until counter reaches 0
    fmt.Println("All workers done")
}
```

## Common Goroutine Gotcha — Loop Variable Capture

```go
// BUG: All goroutines capture the SAME variable 'i'
for i := 0; i < 5; i++ {
    go func() {
        fmt.Println(i) // might print "5 5 5 5 5"
    }()
}

// FIX 1: Pass as argument
for i := 0; i < 5; i++ {
    go func(n int) {
        fmt.Println(n) // correct: prints 0-4 in some order
    }(i)
}

// FIX 2 (Go 1.22+): Loop variables are per-iteration by default
// In Go 1.22+, the first version works correctly!
```

---

# Chapter 14: Channels

Channels are typed conduits for sending and receiving values between goroutines. They provide synchronization and communication.

## Unbuffered Channels

```go
package main

import "fmt"

func main() {
    // Create an unbuffered channel
    ch := make(chan string)

    // Send in a goroutine (sending blocks until someone receives)
    go func() {
        ch <- "hello" // BLOCKS until main goroutine receives
    }()

    // Receive (blocks until someone sends)
    msg := <-ch // BLOCKS until the goroutine sends
    fmt.Println(msg) // "hello"
}

// Unbuffered channel = synchronous handshake
// Sender blocks until receiver is ready
// Receiver blocks until sender sends
```

## Buffered Channels

```go
// Buffered channel — can hold N items before blocking
ch := make(chan int, 3) // buffer size 3

ch <- 1 // doesn't block
ch <- 2 // doesn't block
ch <- 3 // doesn't block
// ch <- 4 // BLOCKS — buffer is full

fmt.Println(<-ch) // 1 (FIFO)
fmt.Println(<-ch) // 2
```

## Channel Direction (Restricting Sends/Receives)

```go
// Send-only channel
func producer(ch chan<- int) {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch) // signal that no more values will be sent
}

// Receive-only channel
func consumer(ch <-chan int) {
    for val := range ch { // range over channel until it's closed
        fmt.Println("Received:", val)
    }
}

func main() {
    ch := make(chan int, 5)
    go producer(ch)
    consumer(ch)
}
```

## Closing Channels

```go
ch := make(chan int)

go func() {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch) // IMPORTANT: only the SENDER should close a channel
}()

// Detecting closed channel IMPORTANT
for {
    val, ok := <-ch
    if !ok {
        break // channel is closed and drained
    }
    fmt.Println(val)
}

// Simpler: range automatically handles closed channels
for val := range ch {
    fmt.Println(val)
}

// Rules about close():
// 1. Only the SENDER closes the channel, never the receiver
// 2. Sending on a closed channel PANICS
// 3. Receiving from a closed channel returns the zero value immediately
// 4. You don't have to close channels — only close when the receiver needs to know
```

## Select Statement (Multiplexing Channels)

`select` is like a `switch` but for channels. It waits on multiple channel operations simultaneously.

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch1 := make(chan string)
    ch2 := make(chan string)

    go func() {
        time.Sleep(1 * time.Second)
        ch1 <- "from channel 1"
    }()

    go func() {
        time.Sleep(2 * time.Second)
        ch2 <- "from channel 2"
    }()

    // Wait for whichever channel is ready first
    for i := 0; i < 2; i++ {
        select {
        case msg1 := <-ch1:
            fmt.Println(msg1)
        case msg2 := <-ch2:
            fmt.Println(msg2)
        }
    }
}

// Select with timeout
select {
case result := <-ch:
    fmt.Println("Got:", result)
case <-time.After(3 * time.Second):
    fmt.Println("Timeout!")
}

// Select with default (non-blocking)
select {
case msg := <-ch:
    fmt.Println("Received:", msg)
default:
    fmt.Println("No message available") // doesn't block
}
```

---

# Chapter 15: Concurrency Patterns

## Pattern 1: Worker Pool

```go
package main

import (
    "fmt"
    "sync"
)

func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
    defer wg.Done()
    for job := range jobs {
        fmt.Printf("Worker %d processing job %d\n", id, job)
        results <- job * 2 // simulate processing
    }
}

func main() {
    const numJobs = 20
    const numWorkers = 5

    jobs := make(chan int, numJobs)
    results := make(chan int, numJobs)

    // Start workers
    var wg sync.WaitGroup
    for w := 1; w <= numWorkers; w++ {
        wg.Add(1)
        go worker(w, jobs, results, &wg)
    }

    // Send jobs
    for j := 1; j <= numJobs; j++ {
        jobs <- j
    }
    close(jobs) // no more jobs

    // Wait for all workers to finish, then close results
    go func() {
        wg.Wait()
        close(results)
    }()

    // Collect results
    for result := range results {
        fmt.Println("Result:", result)
    }
}
```

## Pattern 2: Fan-Out / Fan-In

```go
// Fan-out: Multiple goroutines reading from the same channel
// Fan-in: Multiple channels merged into one

func fanIn(channels ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    merged := make(chan int)

    // Start a goroutine for each input channel
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for val := range c {
                merged <- val
            }
        }(ch)
    }

    // Close merged channel when all inputs are done
    go func() {
        wg.Wait()
        close(merged)
    }()

    return merged
}
```

## Pattern 3: Pipeline

```go
// Stage 1: Generate numbers
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

// Stage 2: Square each number
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

// Stage 3: Filter even numbers
func filterEven(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            if n%2 == 0 {
                out <- n
            }
        }
        close(out)
    }()
    return out
}

func main() {
    // Compose the pipeline
    ch := generate(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
    squared := square(ch)
    evens := filterEven(squared)

    for val := range evens {
        fmt.Println(val) // 4, 16, 36, 64, 100
    }
}
```

## Pattern 4: Context for Cancellation

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func longRunningTask(ctx context.Context, name string) {
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("%s: cancelled (%v)\n", name, ctx.Err())
            return
        default:
            fmt.Printf("%s: working...\n", name)
            time.Sleep(500 * time.Millisecond)
        }
    }
}

func main() {
    // Cancel after 2 seconds
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel() // always call cancel to release resources

    go longRunningTask(ctx, "worker-1")
    go longRunningTask(ctx, "worker-2")

    // Wait for timeout
    <-ctx.Done()
    fmt.Println("Main: all workers cancelled")
    time.Sleep(100 * time.Millisecond) // let goroutines print their cancel message
}

// Context types:
// context.Background()                — root context
// context.TODO()                      — placeholder when unsure which context to use
// context.WithCancel(parent)          — manual cancellation
// context.WithTimeout(parent, dur)    — auto-cancel after duration
// context.WithDeadline(parent, time)  — auto-cancel at specific time
// context.WithValue(parent, key, val) — pass request-scoped data (use sparingly)
```

## Pattern 5: Semaphore (Limiting Concurrency)

```go
// Use a buffered channel as a semaphore
func main() {
    maxConcurrent := 3
    sem := make(chan struct{}, maxConcurrent)

    var wg sync.WaitGroup
    for i := 0; i < 20; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            sem <- struct{}{} // acquire (blocks if 3 goroutines are running)
            defer func() { <-sem }() // release

            fmt.Printf("Task %d running\n", id)
            time.Sleep(1 * time.Second)
        }(i)
    }
    wg.Wait()
}
```

## Pattern 6: Done Channel (Graceful Shutdown)

```go
func main() {
    done := make(chan struct{})

    go func() {
        defer close(done) // signal completion
        // do work...
        fmt.Println("Worker finished")
    }()

    // Wait for worker or timeout
    select {
    case <-done:
        fmt.Println("Worker completed successfully")
    case <-time.After(5 * time.Second):
        fmt.Println("Timeout waiting for worker")
    }
}
```

## Pattern 7: Rate Limiter

```go
func main() {
    // Allow 1 request per 200ms
    limiter := time.NewTicker(200 * time.Millisecond)
    defer limiter.Stop()

    requests := make(chan int, 5)
    for i := 1; i <= 5; i++ {
        requests <- i
    }
    close(requests)

    for req := range requests {
        <-limiter.C // wait for the tick
        fmt.Println("Processing request", req, time.Now())
    }
}

// Burst rate limiter
func burstLimiter() {
    burstyLimiter := make(chan time.Time, 3) // allow burst of 3

    // Pre-fill the burst
    for i := 0; i < 3; i++ {
        burstyLimiter <- time.Now()
    }

    // Then refill every 200ms
    go func() {
        for t := range time.Tick(200 * time.Millisecond) {
            burstyLimiter <- t
        }
    }()

    for i := 1; i <= 10; i++ {
        <-burstyLimiter
        fmt.Println("Request", i, time.Now())
    }
}
```

---

# Chapter 16: Sync Package (Low-Level Concurrency)

## sync.Mutex

```go
type SafeMap struct {
    mu sync.Mutex
    data map[string]int
}

func (m *SafeMap) Set(key string, val int) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.data[key] = val
}

func (m *SafeMap) Get(key string) (int, bool) {
    m.mu.Lock()
    defer m.mu.Unlock()
    val, ok := m.data[key]
    return val, ok
}
```

## sync.RWMutex (Read-Write Lock)

```go
type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()         // multiple readers can hold this simultaneously
    defer c.mu.RUnlock()
    val, ok := c.data[key]
    return val, ok
}

func (c *Cache) Set(key, val string) {
    c.mu.Lock()          // exclusive lock — no readers or writers
    defer c.mu.Unlock()
    c.data[key] = val
}
```

## sync.Once

```go
// Guarantees a function runs exactly once, even from multiple goroutines
var (
    instance *Database
    once     sync.Once
)

func GetDatabase() *Database {
    once.Do(func() {
        // This runs exactly ONCE, no matter how many goroutines call it
        instance = &Database{
            // ... expensive initialization
        }
    })
    return instance
}
```

## sync.Map (Concurrent Map)

```go
// For cases where the map is read-heavy and written rarely
var m sync.Map

// Store
m.Store("key1", "value1")

// Load
val, ok := m.Load("key1")

// LoadOrStore — returns existing value or stores new one
actual, loaded := m.LoadOrStore("key2", "value2")

// Delete
m.Delete("key1")

// Range
m.Range(func(key, value any) bool {
    fmt.Printf("%v: %v\n", key, value)
    return true // return false to stop
})
```

## sync.Pool (Object Reuse)

```go
// Pool of reusable objects — reduces GC pressure
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func process() {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()

    buf.WriteString("hello")
    // use buf...
}
```

## Atomic Operations

```go
import "sync/atomic"

var counter int64

func increment() {
    atomic.AddInt64(&counter, 1) // thread-safe without mutex
}

func getCount() int64 {
    return atomic.LoadInt64(&counter)
}

// Go 1.19+ atomic types
var atomicCounter atomic.Int64

func main() {
    atomicCounter.Add(1)
    atomicCounter.Add(5)
    fmt.Println(atomicCounter.Load()) // 6
}
```

## errgroup (Structured Concurrency)

```go
import "golang.org/x/sync/errgroup"

func main() {
    g, ctx := errgroup.WithContext(context.Background())

    urls := []string{
        "https://api.example.com/users",
        "https://api.example.com/posts",
        "https://api.example.com/comments",
    }

    results := make([]string, len(urls))

    for i, url := range urls {
        i, url := i, url // capture loop variables
        g.Go(func() error {
            // If any goroutine returns an error, ctx is cancelled
            req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
            if err != nil {
                return err
            }
            resp, err := http.DefaultClient.Do(req)
            if err != nil {
                return err
            }
            defer resp.Body.Close()
            body, err := io.ReadAll(resp.Body)
            if err != nil {
                return err
            }
            results[i] = string(body)
            return nil
        })
    }

    // Wait for all goroutines. Returns first error (if any).
    if err := g.Wait(); err != nil {
        log.Fatal(err)
    }

    for _, r := range results {
        fmt.Println(r[:100]) // print first 100 chars
    }
}
```

---

# Chapter 17: Common Standard Library Usage

## I/O Operations

```go
package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "strings"
)

func main() {
    // Reading a file entirely
    data, err := os.ReadFile("config.txt")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println(string(data))

    // Writing a file
    err = os.WriteFile("output.txt", []byte("hello world"), 0644)

    // Reading line by line
    file, err := os.Open("data.txt")
    if err != nil {
        return
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        fmt.Println(line)
    }
    if err := scanner.Err(); err != nil {
        fmt.Println("Error reading:", err)
    }

    // io.Reader abstraction — works with files, strings, HTTP bodies, etc.
    reader := strings.NewReader("Hello from a string reader")
    buf := make([]byte, 5)
    for {
        n, err := reader.Read(buf)
        if err == io.EOF {
            break
        }
        fmt.Print(string(buf[:n]))
    }

    // Copying between readers and writers
    src := strings.NewReader("copy this data")
    dst := os.Stdout
    io.Copy(dst, src) // prints "copy this data"
}
```

## JSON Handling

```go
package main

import (
    "encoding/json"
    "fmt"
)

type User struct {
    ID       int      `json:"id"`
    Name     string   `json:"name"`
    Email    string   `json:"email,omitempty"`
    Tags     []string `json:"tags"`
    Password string   `json:"-"` // never include in JSON
}

func main() {
    // Marshal (struct → JSON)
    user := User{
        ID:   1,
        Name: "Rohit",
        Tags: []string{"admin", "dev"},
    }
    jsonBytes, _ := json.Marshal(user)
    fmt.Println(string(jsonBytes))
    // {"id":1,"name":"Rohit","tags":["admin","dev"]}

    // Pretty print
    jsonPretty, _ := json.MarshalIndent(user, "", "  ")
    fmt.Println(string(jsonPretty))

    // Unmarshal (JSON → struct)
    jsonStr := `{"id":2,"name":"Alice","tags":["user"]}`
    var u2 User
    json.Unmarshal([]byte(jsonStr), &u2)
    fmt.Println(u2.Name) // "Alice"

    // Dynamic JSON (when structure is unknown)
    var result map[string]interface{}
    json.Unmarshal([]byte(jsonStr), &result)
    fmt.Println(result["name"]) // "Alice"

    // Streaming JSON (for large data)
    // encoder := json.NewEncoder(os.Stdout)
    // encoder.Encode(user)
    // decoder := json.NewDecoder(reader)
    // decoder.Decode(&u2)
}
```

## HTTP Server and Client

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "time"
)

// --- HTTP SERVER ---
func handleHello(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}

func handleJSON(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    response := map[string]string{"message": "hello", "status": "ok"}
    json.NewEncoder(w).Encode(response)
}

func startServer() {
    mux := http.NewServeMux()
    mux.HandleFunc("/hello", handleHello)
    mux.HandleFunc("/api/status", handleJSON)

    server := &http.Server{
        Addr:         ":8080",
        Handler:      mux,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    log.Println("Server starting on :8080")
    log.Fatal(server.ListenAndServe())
}

// --- HTTP CLIENT ---
func fetchURL(url string) (string, error) {
    client := &http.Client{
        Timeout: 10 * time.Second,
    }

    resp, err := client.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(body), nil
}
```

## Time and Duration

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    // Current time
    now := time.Now()
    fmt.Println(now)

    // Creating specific times
    t := time.Date(2024, time.March, 14, 12, 0, 0, 0, time.UTC)

    // Duration
    d := 5 * time.Second
    fmt.Println(d) // 5s

    // Sleep
    time.Sleep(100 * time.Millisecond)

    // Formatting (Go uses a REFERENCE TIME: Mon Jan 2 15:04:05 MST 2006)
    // This is January 2, 2006 at 3:04:05 PM
    fmt.Println(now.Format("2006-01-02 15:04:05"))
    fmt.Println(now.Format(time.RFC3339))

    // Parsing
    parsed, _ := time.Parse("2006-01-02", "2024-03-14")
    fmt.Println(parsed)

    // Timer and Ticker
    timer := time.NewTimer(2 * time.Second)
    <-timer.C // blocks for 2 seconds

    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()
    for i := 0; i < 5; i++ {
        <-ticker.C
        fmt.Println("Tick", i)
    }

    // Measuring elapsed time
    start := time.Now()
    // ... do work
    elapsed := time.Since(start)
    fmt.Printf("Took %s\n", elapsed)

    _ = t
}
```
