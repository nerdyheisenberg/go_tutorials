# Go Deep Dive — Part 11: Channels — The Communication Primitive

---

## Chapter 1: What IS a Channel?

A channel is a **typed conduit** through which goroutines communicate. The Go philosophy:

> "Do not communicate by sharing memory; instead, share memory by communicating."
> — Rob Pike

This is the inversion of the traditional mutex model:
- **Mutex model (C++)**: Multiple goroutines/threads share memory, coordinate access with locks
- **Channel model (Go)**: Pass data ownership between goroutines through channels

Channels solve three problems simultaneously:
1. **Communication**: Transfer data between goroutines
2. **Synchronization**: Sender and receiver rendezvous
3. **Signaling**: Signal events (goroutine done, shutdown, etc.)

### Channel Internals

```go
// A channel is a reference type (like a map)
// Internally it's a circular ring buffer + mutex + goroutine wait queues

// The runtime's hchan struct (simplified):
type hchan struct {
    qcount   uint           // number of elements in queue
    dataqsiz uint           // size of the circular buffer
    buf      unsafe.Pointer // pointer to the circular buffer
    elemsize uint16         // size of each element in bytes
    closed   uint32         // 0 = open, 1 = closed
    sendq    waitq          // goroutines waiting to send
    recvq    waitq          // goroutines waiting to receive
    lock     mutex          // protects all fields
}
```

---

## Chapter 2: Unbuffered Channels — Synchronous Communication

An unbuffered channel has **zero capacity**. Every send blocks until a receiver is ready. Every receive blocks until a sender sends.

```go
// Creating an unbuffered channel
ch := make(chan int)      // cap = 0 (default)
fmt.Println(cap(ch))      // 0

// SEND blocks until RECEIVE is ready
go func() {
    ch <- 42              // blocks here until someone receives
    fmt.Println("sent 42")
}()

v := <-ch                // blocks here until someone sends
fmt.Println("received:", v)

// Execution order is guaranteed:
// 1. <-ch blocks in main goroutine
// 2. goroutine sends 42
// 3. main goroutine unblocks, receives 42
// 4. goroutine unblocks after send, prints "sent 42"
// 5. main goroutine prints "received: 42"

// But print order of 3-4 vs 5 is not guaranteed!
```

### What Unbuffered Channels Replace from C++

```go
// C++ with condition variable:
// std::mutex mu;
// std::condition_variable cv;
// bool ready = false;
//
// // producer thread:
// {
//     std::lock_guard<std::mutex> lk(mu);
//     ready = true;
// }
// cv.notify_one();
//
// // consumer thread:
// std::unique_lock<std::mutex> lk(mu);
// cv.wait(lk, [&]{ return ready; });

// Go equivalent — unbuffered channel IS the condition variable:
ready := make(chan struct{}) // signal channel (no data, just signal)

go func() {
    time.Sleep(time.Second)  // do some work
    close(ready)             // signal: ready!
    // OR: ready <- struct{}{} // send once
}()

<-ready  // wait for signal (blocks until closed/sent)
fmt.Println("Go!")
```

---

## Chapter 3: Buffered Channels — Asynchronous Communication

A buffered channel has capacity > 0. Sends only block when the buffer is **full**. Receives only block when the buffer is **empty**.

```go
// Creating buffered channels
ch := make(chan int, 5)     // cap = 5
fmt.Println(len(ch), cap(ch)) // 0 5

// Sends don't block until buffer is full
ch <- 1     // OK, buffer: [1]
ch <- 2     // OK, buffer: [1, 2]
ch <- 3     // OK, buffer: [1, 2, 3]
ch <- 4     // OK, buffer: [1, 2, 3, 4]
ch <- 5     // OK, buffer: [1, 2, 3, 4, 5]
// ch <- 6  // BLOCKS — buffer is full!

// Receives from buffered channel
fmt.Println(<-ch)  // 1 (FIFO!)
fmt.Println(<-ch)  // 2
fmt.Println(len(ch), cap(ch)) // 3 5
```

### Buffered Channel as a Semaphore

```go
// Limit concurrency to N simultaneous operations
const maxConcurrent = 5
sem := make(chan struct{}, maxConcurrent)

var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        
        sem <- struct{}{}    // acquire — blocks if 5 goroutines already running
        defer func() { <-sem }()  // release when done
        
        // Only 5 goroutines can be here simultaneously
        doExpensiveWork(id)
    }(i)
}
wg.Wait()
```

---

## Chapter 4: Channel Direction — Type Safety

Channel direction types restrict how a channel can be used. This enforces correct use at compile time.

```go
// Bidirectional (read and write)
ch := make(chan int)   // type: chan int

// Send-only (can only send, not receive)
var snd chan<- int = ch   // type: chan<- int
snd <- 42  // OK
// v := <-snd  // COMPILE ERROR: receive from send-only channel

// Receive-only (can only receive, not send)
var rcv <-chan int = ch   // type: <-chan int
v := <-rcv  // OK
// rcv <- 42  // COMPILE ERROR: send to receive-only channel

// Directional channels used in function signatures:
func producer(out chan<- int, n int) {  // can only send to 'out'
    for i := 0; i < n; i++ {
        out <- i
    }
    close(out)
}

func consumer(in <-chan int) {          // can only receive from 'in'
    for v := range in {
        fmt.Println(v)
    }
}

func main() {
    ch := make(chan int, 10)
    go producer(ch, 5)  // Go automatically converts chan int → chan<- int
    consumer(ch)        // Go automatically converts chan int → <-chan int
}
```

---

## Chapter 5: Closing Channels

```go
// Rules for closing channels:
// 1. Only the SENDER should close (not the receiver)
// 2. Closing a closed channel PANICS
// 3. Sending to a closed channel PANICS
// 4. Receiving from a closed channel is ALWAYS safe
//    → returns zero values immediately after buffer is drained

ch := make(chan int, 3)
ch <- 1; ch <- 2; ch <- 3
close(ch)

// Receiving from closed channel:
fmt.Println(<-ch)   // 1 (still in buffer)
fmt.Println(<-ch)   // 2 (still in buffer)
fmt.Println(<-ch)   // 3 (last in buffer)
fmt.Println(<-ch)   // 0 (zero value — channel closed and empty)
fmt.Println(<-ch)   // 0 (always zero value for closed channel)

// Comma-ok pattern — detect closed channel:
v, ok := <-ch
if !ok {
    fmt.Println("channel closed")
}

// Range over channel — cleanest way (auto-detects close):
for v := range ch {
    fmt.Println(v)
}
// Loop exits when channel is closed AND empty

// Safe channel close (prevent panic on double-close):
var once sync.Once
safeClose := func(ch chan int) {
    once.Do(func() { close(ch) })
}
```

### Multiple Producers with `sync.WaitGroup`

```go
// When multiple goroutines write to the same channel,
// who closes it? Use WaitGroup + dedicated closer:

func mergeResults(sources []<-chan int) <-chan int {
    out := make(chan int, 100)
    var wg sync.WaitGroup
    
    for _, src := range sources {
        wg.Add(1)
        go func(ch <-chan int) {
            defer wg.Done()
            for v := range ch {
                out <- v
            }
        }(src)
    }
    
    // Close out when all sources are done
    go func() {
        wg.Wait()
        close(out)  // only this goroutine closes
    }()
    
    return out
}
```

---

## Chapter 6: The `select` Statement — Multiplex Channels

`select` waits on multiple channel operations simultaneously. Like a `switch` but for channels.

```go
// Basic select
select {
case msg1 := <-ch1:      // receive from ch1
    fmt.Println("ch1:", msg1)
case msg2 := <-ch2:      // receive from ch2
    fmt.Println("ch2:", msg2)
case ch3 <- "hello":     // send to ch3
    fmt.Println("sent to ch3")
}

// If MULTIPLE cases are ready, Go picks one at RANDOM
// (not the first one — random! Prevents starvation)

// Default case — makes select non-blocking
select {
case msg := <-ch:
    fmt.Println("received:", msg)
default:
    fmt.Println("no message available")  // runs immediately if ch is empty
}

// Timeout pattern
select {
case result := <-work:
    fmt.Println("result:", result)
case <-time.After(5 * time.Second):
    fmt.Println("timeout!")
}

// Done channel pattern (stop a goroutine)
for {
    select {
    case job := <-jobs:
        process(job)
    case <-done:
        return  // exit goroutine
    }
}

// Priority select (workaround — Go doesn't have native priority)
// Process high-priority messages before low-priority
for {
    select {
    case msg := <-criticalCh:
        handleCritical(msg)
        continue  // check critical again before falling to normal
    default:
    }
    
    select {
    case msg := <-criticalCh:
        handleCritical(msg)
    case msg := <-normalCh:
        handleNormal(msg)
    }
}
```

---

## Chapter 7: Channel Patterns — Detailed

### Pattern 1: Pipeline

```go
// Each stage has an input channel and an output channel
// Data flows through the stages like a Unix pipe

func integers(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

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

func filter(in <-chan int, pred func(int) bool) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            if pred(n) {
                out <- n
            }
        }
        close(out)
    }()
    return out
}

// Compose the pipeline:
func main() {
    nums := integers(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
    squares := square(nums)
    smallSquares := filter(squares, func(n int) bool { return n < 50 })
    
    for v := range smallSquares {
        fmt.Println(v)  // 1, 4, 9, 16, 25, 36, 49
    }
}
```

### Pattern 2: Fan-Out / Fan-In

```go
// Fan-out: One producer, many consumers (parallelism)
func fanOut(in <-chan Work, workers int) []<-chan Result {
    channels := make([]<-chan Result, workers)
    for i := 0; i < workers; i++ {
        channels[i] = worker(in)    // each worker reads from SAME in channel
    }
    return channels
}

func worker(in <-chan Work) <-chan Result {
    out := make(chan Result)
    go func() {
        for w := range in {         // multiple workers compete for jobs
            out <- process(w)
        }
        close(out)
    }()
    return out
}

// Fan-in: Many producers, one consumer
func fanIn(channels ...<-chan Result) <-chan Result {
    var wg sync.WaitGroup
    merged := make(chan Result)
    
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan Result) {
            defer wg.Done()
            for r := range c {
                merged <- r         // funnel all results into merged
            }
        }(ch)
    }
    
    go func() {
        wg.Wait()
        close(merged)
    }()
    
    return merged
}
```

### Pattern 3: Stop Channel (Cancellation Before context.Context)

```go
// Before context.Context was added, "done" channels were the pattern
func producer(done <-chan struct{}) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        i := 0
        for {
            select {
            case <-done:
                return  // stop producing
            case out <- i:
                i++
            }
        }
    }()
    return out
}

func main() {
    done := make(chan struct{})
    numbers := producer(done)
    
    for i := 0; i < 10; i++ {
        fmt.Println(<-numbers)
    }
    
    close(done)  // stop the producer
}

// Today, prefer context.Context (next part)
```

### Pattern 4: Channel for One-Time Events

```go
// A closed channel can be received from by MULTIPLE goroutines simultaneously
// → perfect for broadcast signals

type Server struct {
    ready chan struct{}  // signals when server is ready for traffic
}

func (s *Server) waitForReady() {
    <-s.ready  // blocks until channel is closed
}

// All goroutines unblock simultaneously when closed:
close(s.ready)  // broadcasts to ALL waiters at once!

// This is the classic "start-line" pattern:
// All goroutines wait at <-s.ready, then all start at once
```

---

## Chapter 8: Common Channel Mistakes

### Mistake 1: Deadlock

```go
// DEADLOCK: Nobody is sending or receiving
func badDeadlock() {
    ch := make(chan int)
    ch <- 1       // BLOCKS — nobody is receiving!
    <-ch          // Never reached
    // fatal error: all goroutines are asleep - deadlock!
}

// DEADLOCK: Both goroutines waiting for each other
func badDeadlock2() {
    ch1 := make(chan int)
    ch2 := make(chan int)
    
    go func() {
        v := <-ch1    // waits for main to send
        ch2 <- v
    }()
    
    v := <-ch2        // waits for goroutine to send, but goroutine waits for us!
    ch1 <- v
    // Deadlock!
}

// FIX: Always have a goroutine for sending AND another for receiving
// or use buffered channels
```

### Mistake 2: Sending on Closed Channel

```go
// PANIC: send on closed channel
ch := make(chan int)
close(ch)
ch <- 1   // panic: send on closed channel

// This commonly happens with fan-out patterns where multiple goroutines try to close
// FIX: Only one goroutine closes, use sync.Once for safety
```

### Mistake 3: Goroutine leak via channel

```go
// LEAK: Producer keeps running but consumer stopped reading
func badLeak() {
    ch := make(chan int, 10)
    
    // Producer — runs forever
    go func() {
        for i := 0; ; i++ {
            ch <- i  // blocks when buffer is full — goroutine is stuck!
        }
    }()
    
    // Consumer — only reads 5 items then returns
    for i := 0; i < 5; i++ {
        <-ch
    }
    // Producer goroutine is now permanently blocked! LEAK!
}

// FIX: Use context.Context or done channel
func goodNoLeak() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()  // signals producer to stop
    
    ch := make(chan int, 10)
    go func() {
        defer close(ch)
        for i := 0; ; i++ {
            select {
            case <-ctx.Done():
                return  // stop when cancelled
            case ch <- i:
            }
        }
    }()
    
    for i := 0; i < 5; i++ {
        <-ch
    }
    // defer cancel() runs → goroutine exits cleanly
}
```

---

## Chapter 9: Comprehensive Channel Testing

```go
package channels_test

import (
    "context"
    "sync"
    "testing"
    "time"
)

// Test unbuffered channel blocking behavior
func TestUnbufferedBlocking(t *testing.T) {
    ch := make(chan int)
    
    // Without goroutine, send would deadlock
    go func() { ch <- 42 }()
    
    select {
    case v := <-ch:
        if v != 42 {
            t.Errorf("received %d, want 42", v)
        }
    case <-time.After(1 * time.Second):
        t.Error("timed out waiting for value")
    }
}

// Test buffered channel non-blocking behavior
func TestBufferedNonBlocking(t *testing.T) {
    ch := make(chan int, 3)
    
    // Can send without goroutine (non-blocking up to capacity)
    ch <- 1
    ch <- 2
    ch <- 3
    
    if len(ch) != 3 { t.Errorf("len = %d, want 3", len(ch)) }
    
    if v := <-ch; v != 1 { t.Errorf("FIFO: got %d, want 1", v) }
    if v := <-ch; v != 2 { t.Errorf("FIFO: got %d, want 2", v) }
}

// Test close and range behavior
func TestCloseAndRange(t *testing.T) {
    ch := make(chan int, 5)
    for i := 1; i <= 5; i++ { ch <- i }
    close(ch)
    
    var received []int
    for v := range ch {
        received = append(received, v)
    }
    
    if len(received) != 5 {
        t.Errorf("expected 5 values, got %d", len(received))
    }
    for i, v := range received {
        if v != i+1 { t.Errorf("index %d: got %d, want %d", i, v, i+1) }
    }
}

// Test comma-ok on closed channel
func TestClosedChannelCommaOk(t *testing.T) {
    ch := make(chan int, 1)
    ch <- 42
    close(ch)
    
    v1, ok1 := <-ch
    if !ok1 || v1 != 42 { t.Errorf("first receive: ok=%v val=%d", ok1, v1) }
    
    v2, ok2 := <-ch
    if ok2 || v2 != 0 { t.Errorf("second receive (closed): ok=%v val=%d", ok2, v2) }
}

// Test select — verify fair selection
func TestSelectFairness(t *testing.T) {
    ch1 := make(chan int, 100)
    ch2 := make(chan int, 100)
    
    for i := 0; i < 100; i++ { ch1 <- 1 }
    for i := 0; i < 100; i++ { ch2 <- 2 }
    
    counts := map[int]int{}
    for i := 0; i < 200; i++ {
        select {
        case v := <-ch1: counts[v]++
        case v := <-ch2: counts[v]++
        }
    }
    
    // Both channels should be selected roughly equally (random)
    // Allow 30% variance from perfect 50/50
    if counts[1] < 70 || counts[1] > 130 {
        t.Logf("select fairness: ch1=%d, ch2=%d (should be ~100 each)", counts[1], counts[2])
        // Note: this is a probabilistic test, may rarely fail
    }
}

// Test pipeline
func makePipeline(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

func squarePipeline(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in { out <- n * n }
        close(out)
    }()
    return out
}

func TestPipeline(t *testing.T) {
    nums := makePipeline(1, 2, 3, 4, 5)
    squares := squarePipeline(nums)
    
    expected := []int{1, 4, 9, 16, 25}
    var got []int
    for v := range squares {
        got = append(got, v)
    }
    
    if len(got) != len(expected) {
        t.Fatalf("got %d results, want %d", len(got), len(expected))
    }
    for i := range expected {
        if got[i] != expected[i] {
            t.Errorf("index %d: got %d, want %d", i, got[i], expected[i])
        }
    }
}

// Test semaphore pattern
func TestSemaphore(t *testing.T) {
    const maxConcurrent = 3
    sem := make(chan struct{}, maxConcurrent)
    var currentConcurrent atomic.Int32
    var maxObserved atomic.Int32
    var wg sync.WaitGroup
    
    for i := 0; i < 20; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()
            
            n := currentConcurrent.Add(1)
            if n > maxObserved.Load() { maxObserved.Store(n) }
            
            time.Sleep(10 * time.Millisecond)
            currentConcurrent.Add(-1)
        }()
    }
    
    wg.Wait()
    
    if m := maxObserved.Load(); m > maxConcurrent {
        t.Errorf("max concurrent = %d, want <= %d", m, maxConcurrent)
    }
}

// Test context cancellation through channel
func TestContextCancellation(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
    defer cancel()
    
    results := make(chan int)
    
    go func() {
        defer close(results)
        for i := 0; ; i++ {
            select {
            case <-ctx.Done():
                return
            case results <- i:
            }
        }
    }()
    
    var count int
    for range results { count++ }
    
    if count == 0 {
        t.Error("expected some results before cancellation")
    }
    t.Logf("received %d values before cancellation", count)
}

// Benchmark: buffered vs unbuffered
func BenchmarkUnbufferedChannel(b *testing.B) {
    ch := make(chan int)
    go func() {
        for v := range ch {
            _ = v
        }
    }()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ch <- i
    }
    close(ch)
}

func BenchmarkBufferedChannel(b *testing.B) {
    ch := make(chan int, b.N)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ch <- i
    }
}
```

---

**Summary of Part 11:**
- Channels are typed conduits for goroutine communication — not just mutexes with extra steps
- Unbuffered: synchronous handshake (sender AND receiver must be ready)
- Buffered: async up to capacity; use as semaphores, rate limiters, work queues
- Channel direction (`chan<-`, `<-chan`) enforces correct send/receive at compile time
- Only the SENDER closes a channel — closing twice panics; sending to closed panics
- `range` over channel: cleanest iteration, auto-exits when closed
- `select` randomly picks from ready cases — prevents starvation
- Pipeline, fan-out/fan-in, semaphore, broadcast are the core channel patterns
- Goroutine leaks: producer stuck on full channel — always pair with cancellation
- Use `-race` and watch for channel-based deadlocks
