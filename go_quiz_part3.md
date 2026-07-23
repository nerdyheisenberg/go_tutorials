# Go Quiz — Part 3 of 5 (Q201–Q300)
### Covers: Parts 9–12 (Error Handling, Goroutines, Channels, sync Package)
> Answer is hidden after `> ✅`. Try to answer before looking!

---

## PART 9 — Error Handling (Q201–Q225)

**Q201.** What is the zero value of the `error` interface?

- a) `""`
- b) `0`
- c) `nil`
- d) A zero-value error struct

> ✅ **c** — `nil` means no error. Always check `if err != nil`.

---

**Q202.** What does `fmt.Errorf("context: %w", err)` do differently from `fmt.Errorf("context: %v", err)`?

- a) `%w` makes the error prettier; `%v` is plain text
- b) `%w` wraps the error (preserving it for `errors.Is`/`errors.As`); `%v` embeds the string only (loses the original error)
- c) `%w` is only valid in Go 1.20+
- d) No difference — both preserve the original error

> ✅ **b** — `%w` calls `Unwrap()` on the result, enabling the error chain traversal by `errors.Is` and `errors.As`.

---

**Q203.** What is a sentinel error?

- a) An error that automatically recovers from panics
- b) A package-level `var ErrXxx = errors.New("...")` used as a known, named error value
- c) An error that only appears at security boundaries
- d) An error that wraps another error

> ✅ **b** — Sentinels like `io.EOF`, `sql.ErrNoRows` let callers check with `errors.Is(err, io.EOF)`.

---

**Q204.** What is the output?
```go
var ErrNotFound = errors.New("not found")
wrapped := fmt.Errorf("layer: %w", ErrNotFound)
fmt.Println(errors.Is(wrapped, ErrNotFound))
fmt.Println(wrapped == ErrNotFound)
```
- a) `true true`
- b) `true false`
- c) `false false`
- d) `false true`

> ✅ **b** — `errors.Is` traverses the chain and finds `ErrNotFound`. Direct `==` compares the wrapped error to the original — different values.

---

**Q205.** True or False: `errors.As(err, &target)` can extract a typed error from anywhere in the error chain.

> ✅ **True** — `errors.As` unwraps the chain and checks if any error in it matches the target type.

---

**Q206.** What does this code print?
```go
type MyErr struct{ Code int }
func (e *MyErr) Error() string { return fmt.Sprintf("code %d", e.Code) }

err := fmt.Errorf("wrap: %w", &MyErr{Code: 42})
var myErr *MyErr
if errors.As(err, &myErr) {
    fmt.Println(myErr.Code)
}
```
- a) Nothing
- b) `42`
- c) `code 42`
- d) Compile error

> ✅ **b** — `errors.As` finds `*MyErr` in the chain and assigns it to `myErr`. Then `myErr.Code` is `42`.

---

**Q207.** True or False: `errors.New("msg") == errors.New("msg")` returns `true`.

> ✅ **False** — `errors.New` returns a pointer to a new struct each time. Pointer equality (`==`) compares addresses, not content. Always use `errors.Is` for sentinel error checks.

---

**Q208.** When should you use `panic` in Go?

- a) For any error that is unexpected at runtime
- b) Never — always use errors
- c) For programmer errors (invariant violations, bugs), never for expected runtime failures
- d) For performance-critical code paths

> ✅ **c** — Panic is for "this should never happen" situations. HTTP servers should recover panics. User-facing errors should be returned as `error` values.

---

**Q209.** What does `_ = err` mean?

- a) Swaps `err` with nil
- b) Intentionally discards the error — used to silence "declared and not used" in specific situations
- c) Resets `err` to nil
- d) Compile error

> ✅ **b** — The blank identifier `_` discards a value. Using `_ = err` is sometimes used to intentionally acknowledge you're ignoring an error. Still bad practice in production code.

---

**Q210.** What is the correct anti-pattern to avoid?
```go
if err != nil {
    return fmt.Errorf("error: %w", err)
}
```
- a) This code is correct
- b) `"error:"` adds no context — the prefix is meaningless
- c) Should use `%v` instead of `%w`
- d) Should `panic` instead

> ✅ **b** — "error: original message" — the prefix "error:" is redundant. Use meaningful context: `return fmt.Errorf("getUserByID(%d): %w", id, err)`.

---

**Q211.** True or False: `errors.Join` (Go 1.20+) combines multiple errors into one, preserving all for `errors.Is`/`errors.As`.

> ✅ **True** — `errors.Join(err1, err2)` returns an error whose `Unwrap() []error` returns both. `errors.Is/As` traverses all of them.

---

**Q212.** What is the `must` pattern used for?
```go
func Must[T any](v T, err error) T {
    if err != nil { panic(err) }
    return v
}
```
- a) For production error handling in hot paths
- b) For errors that represent programming bugs — values that MUST be valid (package-level init, hardcoded URLs, compiled regexps)
- c) For user-facing error messages
- d) It's an anti-pattern — never use panics

> ✅ **b** — `Must(url.Parse("https://api.example.com"))` — if this URL is wrong, it's a code bug, not a runtime error. Panic at startup is acceptable.

---

**Q213.** What is the proper way to add a custom `Is()` method to your error type?

```go
type TimeoutError struct{ Duration time.Duration }
func (e *TimeoutError) Is(target error) bool {
    _, ok := target.(*TimeoutError)
    return ok
}
```
- a) This is wrong syntax
- b) This allows `errors.Is(err, &TimeoutError{})` to return true for ANY `*TimeoutError` in the chain
- c) This replaces the `Error()` method requirement
- d) This only works for sentinel errors

> ✅ **b** — Custom `Is()` lets you match by type rather than exact value — useful for `errors.Is(err, new(TimeoutError))`.

---

**Q214.** True or False: Wrapping an error with `%v` instead of `%w` still allows `errors.Is` to traverse the chain.

> ✅ **False** — `%v` embeds the error's string representation only. The resulting error has no `Unwrap()` method. `errors.Is` cannot traverse it.

---

**Q215.** What is the output?
```go
err1 := errors.New("error one")
err2 := errors.New("error two")
combined := errors.Join(err1, err2)
fmt.Println(errors.Is(combined, err1))
fmt.Println(errors.Is(combined, err2))
```
- a) `false false`
- b) `true false`
- c) `true true`
- d) Compile error — errors.Join doesn't exist

> ✅ **c** — `errors.Join` (Go 1.20+) preserves both errors. `errors.Is` finds both in the joined error.

---

**Q216.** What is wrong with ignoring errors from `defer`?
```go
defer file.Close() // error ignored
```
- a) Nothing — defer errors are always safe to ignore
- b) If Close() fails (e.g., flushing data to disk fails), you silently lose data
- c) Compile error — defer must capture return values
- d) Close() cannot return an error

> ✅ **b** — `Close()` on a file can fail (e.g., network filesystem). For write operations, check the error with a named return or explicit defer function.

---

**Q217.** True or False: An error type can implement both `Error() string` and `Unwrap() error` simultaneously.

> ✅ **True** — A custom error type can wrap another error AND have its own message. This is how rich error chains work.

---

**Q218.** What is the correct way to return an error from a deferred function?
```go
func process() (err error) {
    f, _ := os.Open("file")
    defer func() {
        if cerr := f.Close(); cerr != nil && err == nil {
            err = cerr
        }
    }()
    // ...
    return nil
}
```
- a) This is wrong — you cannot modify return values from defer
- b) This is correct — named return `err` can be modified in defer
- c) Use `log.Fatal` instead
- d) Compile error

> ✅ **b** — Named return values are in scope for deferred functions. This is a common pattern for capturing close errors.

---

**Q219.** What does `errors.Unwrap(err)` return if `err` was not wrapped?

- a) `err` itself
- b) `nil`
- c) An empty error
- d) Panics

> ✅ **b** — If the error doesn't implement `Unwrap()`, `errors.Unwrap` returns `nil`.

---

**Q220.** True or False: In Go, it is idiomatic to return `(result, error)` from functions that can fail.

> ✅ **True** — This is the core error handling pattern in Go. The caller explicitly checks `if err != nil`.

---

**Q221.** What is the output?
```go
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}
result, err := divide(10, 0)
fmt.Println(result, err)
```
- a) `0 <nil>`
- b) `0 division by zero`
- c) Compile error
- d) Runtime panic

> ✅ **b** — Returns `(0, error)`. Println formats `err` by calling `err.Error()`.

---

**Q222.** True or False: `recover()` can catch any panic, including out-of-memory panics.

> ✅ **False** — Out-of-memory conditions (like running out of memory for a goroutine stack) may terminate the program before `recover()` can run. `recover()` works for user panics and most runtime panics.

---

**Q223.** What is the idiomatic way to handle errors in Go when the caller doesn't care about the specific error type?

- a) Return `interface{}`
- b) Return `error` and let the caller use `err.Error()` for the message
- c) Log and continue
- d) Use `panic`

> ✅ **b** — Return `error`. The caller checks `err != nil` and handles appropriately. For specific behavior, the caller can use `errors.Is`/`errors.As`.

---

**Q224.** What is `ValidationError` used for in this pattern?
```go
type ValidationError struct {
    Field   string
    Message string
}
func (e *ValidationError) Error() string { ... }
```
- a) Only for JSON validation
- b) A structured error that carries field-specific information extractable via `errors.As`
- c) A replacement for Go's built-in error
- d) Only valid in HTTP handlers

> ✅ **b** — Structured errors let callers extract specific info (`field`, `message`) via `errors.As(err, &ve)`.

---

**Q225.** True or False: It is acceptable to return a nil error interface from a function that returns `*MyError` when no error occurred.

> ✅ **False** — Return `nil` typed as `error`, not `(*MyError)(nil)`. Returning a typed nil pointer as `error` creates a non-nil interface value (the nil interface bug). Use: `return nil` not `return (*MyError)(nil)`.

---

## PART 10 — Goroutines (Q226–Q250)

**Q226.** What is a goroutine?

- a) A wrapper around an OS thread
- b) A lightweight concurrent execution unit managed by the Go runtime (not the OS)
- c) A scheduled task in the OS scheduler
- d) A Go function that returns an error

> ✅ **b** — Goroutines are ~2KB initial stack, scheduled by Go's runtime GMP scheduler, NOT the OS.

---

**Q227.** What does GMP stand for in Go's scheduler?

- a) Goroutine, Memory, Processor
- b) Goroutine, Machine (OS thread), Processor (logical CPU with run queue)
- c) Global, Memory, Parallel
- d) Goroutine, Mutex, Pool

> ✅ **b** — G=Goroutine, M=Machine (OS thread), P=Processor (owns a run queue of goroutines).

---

**Q228.** What is the initial stack size of a goroutine?

- a) ~1MB (same as an OS thread)
- b) ~8MB
- c) ~2KB (grows dynamically)
- d) Fixed 64KB

> ✅ **c** — Goroutines start with ~2KB and grow as needed. This enables millions of goroutines concurrently.

---

**Q229.** What is the output?
```go
var wg sync.WaitGroup
for i := 0; i < 3; i++ {
    wg.Add(1)
    go func(n int) {
        defer wg.Done()
        fmt.Println(n)
    }(i)
}
wg.Wait()
```
- a) `0 1 2` in order
- b) `0 1 2` in any order
- c) `3 3 3`
- d) Compile error

> ✅ **b** — `i` is passed as argument `n` (copied), so values are 0, 1, 2. But goroutine scheduling is non-deterministic, so ORDER is random.

---

**Q230.** What is the most common mistake with `sync.WaitGroup`?

- a) Calling `Wait()` before goroutines start
- b) Calling `Add(1)` INSIDE the goroutine — `Wait()` might return before `Add()` runs
- c) Not calling `Done()` at the end
- d) Using `defer wg.Done()`

> ✅ **b** — Always call `wg.Add(n)` BEFORE starting the goroutine, not inside it.

---

**Q231.** True or False: A goroutine leak is when a goroutine is blocked forever, consuming memory and preventing GC of captured variables.

> ✅ **True** — Leaked goroutines accumulate over time and eventually exhaust memory.

---

**Q232.** What is the primary tool to detect goroutine leaks in tests?

- a) `go test -race`
- b) `runtime.NumGoroutine()` comparison before/after
- c) `go.uber.org/goleak` package's `VerifyNone(t)`
- d) Both b and c

> ✅ **d** — Both approaches work. `goleak` is more robust. Manual goroutine count comparison also catches leaks.

---

**Q233.** What is a data race in Go?

- a) When two goroutines read the same variable simultaneously
- b) When two goroutines access the same memory location concurrently and at least one is a write, without synchronization
- c) When a goroutine accesses freed memory
- d) When a goroutine runs faster than expected

> ✅ **b** — Concurrent reads are safe. One write (unsynchronized) + any other access = data race.

---

**Q234.** How do you detect data races in Go?

- a) `go test -check-races ./...`
- b) `go test -race ./...` or `go run -race main.go`
- c) `go vet -race ./...`
- d) Enable GODEBUG=race=1

> ✅ **b** — The `-race` flag instruments the binary. Races are detected at **runtime** when the code actually runs with concurrent access.

---

**Q235.** True or False: `runtime.GOMAXPROCS(1)` prevents all data races.

> ✅ **False** — Even on one OS thread, goroutines can interleave at scheduling points. `GOMAXPROCS(1)` just reduces parallelism but doesn't eliminate concurrency.

---

**Q236.** What is the output?
```go
counter := 0
var wg sync.WaitGroup
for i := 0; i < 1000; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        counter++  // no synchronization!
    }()
}
wg.Wait()
fmt.Println(counter)
```
- a) Always `1000`
- b) Exactly `500`
- c) Undefined — any value from 1 to 1000 due to data race
- d) Compile error

> ✅ **c** — This is a data race. The increment `counter++` is not atomic (read + write). Result is unpredictable.

---

**Q237.** What is `runtime.Gosched()` used for?

- a) Schedules a new goroutine
- b) Voluntarily yields the CPU to other goroutines
- c) Stops the current goroutine permanently
- d) Sets GOMAXPROCS

> ✅ **b** — Rarely needed but useful in benchmarks or tests to give other goroutines a chance to run.

---

**Q238.** True or False: When you `go func()...` a goroutine, the goroutine starts executing immediately on another OS thread.

> ✅ **False** — The goroutine is placed on a run queue. It runs when a Processor (P) picks it up, which may or may not be immediate.

---

**Q239.** What is the best way to stop a long-running background goroutine?

- a) `kill goroutine(id)`
- b) Use `context.Context` with cancellation — the goroutine checks `ctx.Done()`
- c) Use `runtime.Goexit()`
- d) Close the goroutine with a mutex

> ✅ **b** — Pass a `context.Context`; cancel it to signal the goroutine to stop.

---

**Q240.** What is `sync.WaitGroup` used for?

- a) Mutual exclusion between goroutines
- b) Waiting for a collection of goroutines to finish
- c) Limiting the number of concurrent goroutines
- d) Sharing data between goroutines

> ✅ **b** — Add(n) before spawning, Done() when each goroutine finishes, Wait() blocks until count reaches 0.

---

**Q241.** True or False: After `wg.Wait()` returns, you can safely reuse the same `WaitGroup`.

> ✅ **True** — A WaitGroup can be reused after `Wait()` returns. Just call `Add()` again.

---

**Q242.** What is the output when this is compiled with `-race`?
```go
m := make(map[string]int)
go func() { m["a"] = 1 }()
go func() { m["b"] = 2 }()
time.Sleep(time.Second)
```
- a) No issues
- b) "fatal error: concurrent map writes"
- c) Compile error
- d) Data race detector triggers, program continues

> ✅ **b** — Concurrent map writes cause a runtime panic in Go (even without `-race`): "fatal error: concurrent map writes".

---

**Q243.** What does `runtime.NumGoroutine()` return?

- a) Maximum allowed goroutines
- b) Number of currently live goroutines including the current one
- c) Number of OS threads currently in use
- d) Number of goroutines created since program start

> ✅ **b** — Returns the current count of live goroutines.

---

**Q244.** True or False: Goroutines can be preempted in Go (since 1.14), so a tight loop won't hog the scheduler.

> ✅ **True** — Before Go 1.14, goroutines were only preempted at function calls. Since 1.14, signals-based preemption allows interrupting tight loops.

---

**Q245.** What does `time.After(d)` return?

- a) A `time.Time`
- b) A `<-chan time.Time` that receives after duration `d`
- c) A boolean
- d) A `time.Timer`

> ✅ **b** — `time.After(d)` is shorthand for `time.NewTimer(d).C`. Use in `select` for timeouts. Note: the timer goroutine leaks if unused; prefer `time.NewTimer` + `defer timer.Stop()`.

---

**Q246.** True or False: `go func() { }()` — the `()` at the end is required.

> ✅ **True** — `go func() { }` without `()` is a compile error. `go` takes a function CALL expression, not a function value.

---

**Q247.** What is the goroutine stack growth model in Go?

- a) Fixed stack — panics if exceeded
- b) Dynamically growing stack — doubles until heap is exhausted
- c) Starts small (~2KB), grows by copying to larger stack as needed; shrinks when no longer needed
- d) Uses the same stack as the calling goroutine

> ✅ **c** — Go goroutine stacks are "segmented" (early) then "copying" (since Go 1.4). They grow and shrink automatically.

---

**Q248.** What is the purpose of `t.Parallel()` in tests?

- a) Runs the test using multiple CPU cores
- b) Marks the test to run in parallel with other tests that also call `t.Parallel()`
- c) Disables test timeout
- d) Runs the test in a separate goroutine automatically

> ✅ **b** — `t.Parallel()` releases the current test's execution slot so other parallel tests can run. They resume together.

---

**Q249.** True or False: `defer wg.Done()` is safer than `wg.Done()` at the end of a goroutine function.

> ✅ **True** — If the goroutine panics, `defer wg.Done()` still runs (as panics unwind deferred calls). Non-deferred `wg.Done()` would be skipped, causing `wg.Wait()` to block forever.

---

**Q250.** What is the typical use case for `runtime.GOMAXPROCS(n)`?

- a) Setting it in production to improve performance
- b) Setting it in benchmarks/tests to control parallelism for reproducibility
- c) Required for goroutines to use multiple cores
- d) Limiting memory usage

> ✅ **b** — Default is `runtime.NumCPU()`. Changing it is rarely needed in production; used in tests to expose race conditions.

---

## PART 11 — Channels (Q251–Q275)

**Q251.** What is an unbuffered channel?

- a) A channel with no synchronization
- b) A channel with capacity 0 — sender blocks until receiver is ready, and vice versa
- c) A channel that drops messages if full
- d) A channel used only for signaling

> ✅ **b** — Unbuffered channels are a synchronization point. Both sides must be ready simultaneously.

---

**Q252.** What is the output?
```go
ch := make(chan int, 3)
ch <- 1; ch <- 2; ch <- 3
fmt.Println(len(ch), cap(ch))
```
- a) `3, 3`
- b) `0, 3`
- c) `3, 0`
- d) Compile error

> ✅ **a** — `len(ch)` = number of items in buffer = 3. `cap(ch)` = buffer capacity = 3.

---

**Q253.** True or False: Sending to a closed channel panics.

> ✅ **True** — `ch <- value` on a closed channel: panic "send on closed channel".

---

**Q254.** What does receiving from a closed, empty channel return?

- a) Panics
- b) Blocks forever
- c) Returns the zero value and `ok=false` immediately
- d) Returns an error

> ✅ **c** — `v, ok := <-ch` after close and drain: `ok = false`, `v = zero value`. Never blocks.

---

**Q255.** What is the `chan<-` type?

- a) Receive-only channel
- b) Send-only channel
- c) Bidirectional channel with buffer
- d) Channel that never blocks

> ✅ **b** — `chan<- T` is a send-only channel type. You can send but not receive. `<-chan T` is receive-only.

---

**Q256.** Who should close a channel — sender or receiver?

- a) The receiver — it knows when it's done
- b) The sender — it knows when there's no more data to send
- c) Both — close from either side
- d) Neither — channels auto-close when garbage collected

> ✅ **b** — Only the sender should close. Receivers signal "done" through other means (e.g., a separate done channel or context cancellation).

---

**Q257.** What is the output?
```go
ch := make(chan int, 2)
ch <- 1
ch <- 2
close(ch)
for v := range ch {
    fmt.Println(v)
}
```
- a) Infinite loop
- b) `1` then `2` then loop exits
- c) `1` then `2` then `0`
- d) Compile error

> ✅ **b** — `range` on channel reads until channel is closed AND empty. Exits cleanly after reading 1 and 2.

---

**Q258.** What does `select` do when multiple cases are ready simultaneously?

- a) Executes the first case listed
- b) Picks a case at random
- c) Executes all ready cases
- d) Blocks until only one case is ready

> ✅ **b** — Go's `select` picks randomly from ready cases to prevent starvation.

---

**Q259.** What is the `default` case in `select` used for?

- a) Error handling when all channels fail
- b) Makes the select non-blocking — executes if no other case is ready
- c) The fallback for type switches
- d) Catches panics from channel operations

> ✅ **b** — `select { case v := <-ch: ...; default: fmt.Println("no data") }` — non-blocking receive.

---

**Q260.** True or False: Closing a nil channel panics.

> ✅ **True** — `var ch chan int; close(ch)` → panic: "close of nil channel".

---

**Q261.** What is a goroutine leak via channel?

- a) A goroutine that uses too many channels
- b) A goroutine blocked on a channel send/receive that will never happen
- c) A channel that was never closed
- d) A channel with no buffer

> ✅ **b** — If a producer writes to a full buffered channel and nobody reads, the goroutine blocks forever — it's leaked.

---

**Q262.** True or False: `make(chan int)` and `make(chan int, 0)` create equivalent channels.

> ✅ **True** — Both create unbuffered channels with capacity 0.

---

**Q263.** What is a pipeline pattern in Go channels?

- a) A series of stages where each stage receives from a channel, transforms data, and sends to another channel
- b) A pattern for reading files sequentially
- c) Multiple goroutines writing to the same channel
- d) A channel connected to stdin/stdout

> ✅ **a** — Classic concurrent pipeline: `gen → square → filter → print`, each stage in its own goroutine.

---

**Q264.** What is the fan-out pattern?

- a) One channel receives data and distributes to multiple goroutines (workers)
- b) Multiple channels merge into one
- c) A channel that duplicates every message
- d) One goroutine reads from multiple channels

> ✅ **a** — Fan-out: one input channel, N worker goroutines all reading from it. Fan-in: N channels merged into one output.

---

**Q265.** What is the output?
```go
ch := make(chan int, 1)
ch <- 42
v, ok := <-ch
fmt.Println(v, ok)
close(ch)
v2, ok2 := <-ch
fmt.Println(v2, ok2)
```
- a) `42 true` then `0 false`
- b) `42 true` then `0 true`
- c) `42 false` then `0 false`
- d) Compile error

> ✅ **a** — First receive: `v=42, ok=true` (from buffer). After close, second receive: `v2=0, ok2=false`.

---

**Q266.** True or False: A channel variable itself can be nil, which is different from a closed channel.

> ✅ **True** — `var ch chan int` is nil. Sending to nil blocks forever. Receiving from nil blocks forever. Closing nil panics. A closed non-nil channel returns zero values.

---

**Q267.** What is the semaphore pattern with channels?
```go
sem := make(chan struct{}, N)
```
- a) Limits concurrent goroutines to N by acquiring (send) and releasing (receive)
- b) Creates N goroutines
- c) Signals N goroutines to start
- d) Pipes N channels together

> ✅ **a** — `sem <- struct{}{}` acquires (blocks if N slots full). `<-sem` releases. Classic bounded concurrency.

---

**Q268.** What does `close(ch)` broadcast to all receivers?

- a) An error value
- b) The zero value of the channel type with `ok=false`
- c) A special "CLOSED" signal struct
- d) It only notifies one receiver

> ✅ **b** — ALL goroutines blocked on `<-ch` unblock immediately with `(zero, false)` when closed. This makes channels a broadcast primitive when closed.

---

**Q269.** True or False: You can use `select` with only one case — it still compiles.

> ✅ **True** — `select { case v := <-ch: ... }` with one case always blocks until that case is ready. Uncommon but valid.

---

**Q270.** What is the timeout pattern in Go?
```go
select {
case result := <-work:
    handle(result)
case <-time.After(5 * time.Second):
    fmt.Println("timeout!")
}
```
- a) Compile error — time.After can't be used in select
- b) Waits up to 5 seconds for a result; prints "timeout!" if none arrives in time
- c) Always times out after exactly 5 seconds
- d) Runs both cases simultaneously

> ✅ **b** — `time.After` returns a channel that receives after the duration. Classic timeout with select.

---

**Q271.** True or False: Receiving from a nil channel blocks forever (never panics).

> ✅ **True** — `var ch chan int; v := <-ch` blocks forever. Combined with `select`, this effectively disables a case.

---

**Q272.** Why is a buffered channel with capacity 1 useful as a "mutex replacement"?
```go
mu := make(chan struct{}, 1)
mu <- struct{}{} // acquire
// critical section
<-mu             // release
```
- a) It's not — always use sync.Mutex
- b) Allows the pattern but is less efficient than sync.Mutex; still conceptually valid
- c) Equivalent and equally efficient to sync.Mutex
- d) Creates a deadlock

> ✅ **b** — Works correctly but `sync.Mutex` is more efficient and idiomatic. Channel mutex is occasionally used in educational contexts.

---

**Q273.** What is the "done channel" pattern?
```go
done := make(chan struct{})
go func() {
    defer close(done)
    doWork()
}()
<-done // wait for completion
```
- a) An anti-pattern — use WaitGroup instead
- b) A valid way to wait for a goroutine AND signal cancellation (close done to cancel)
- c) Only valid for goroutines that return errors
- d) Requires a buffered channel

> ✅ **b** — Closing `done` broadcasts to ALL goroutines waiting on it — more powerful than WaitGroup for signaling. Use when you need both waiting AND cancellation signaling.

---

**Q274.** True or False: You can close the same channel twice without panicking.

> ✅ **False** — Closing a channel that is already closed: panic "close of closed channel". Use `sync.Once` to safely close once.

---

**Q275.** What is the difference between `fmt.Println(<-ch)` and reading the channel separately?
```go
// Option A:
fmt.Println(<-ch)
// Option B:
v := <-ch
fmt.Println(v)
```
- a) Option A is always better — fewer allocations
- b) With Option A, if `ch` is closed, `fmt.Println` sees the zero value; both approaches behave identically
- c) Option A panics on closed channel; B doesn't
- d) Option A doesn't work with bidirectional channels

> ✅ **b** — Both are equivalent in behavior. Closed channels return zero value in both cases.

---

## PART 12 — sync Package (Q276–Q300)

**Q276.** True or False: `sync.Mutex` is reentrant — the same goroutine can lock it twice.

> ✅ **False** — `sync.Mutex` is NOT reentrant. If a goroutine calls `Lock()` twice without unlocking, it deadlocks permanently.

---

**Q277.** What is the best practice for using `sync.Mutex`?

- a) Lock and unlock in different functions
- b) Always use `defer mu.Unlock()` immediately after `mu.Lock()`
- c) Use global mutexes for simplicity
- d) Pair each Lock with a separate Unlock at the end of the function

> ✅ **b** — `defer mu.Unlock()` guarantees unlock even if the function panics.

---

**Q278.** True or False: You can copy a `sync.Mutex` by value (e.g., assign a struct containing a Mutex).

> ✅ **False** — Copying a mutex that is locked (or has been locked) corrupts its state. Always use a pointer to the struct, or pass the struct by pointer.

---

**Q279.** When is `sync.RWMutex` faster than `sync.Mutex`?

- a) Always
- b) When reads significantly outnumber writes (multiple concurrent readers allowed)
- c) When writes outnumber reads
- d) When GOMAXPROCS > 4

> ✅ **b** — `RWMutex` allows N concurrent readers OR 1 exclusive writer. Only beneficial in read-heavy workloads. With equal reads/writes, overhead may make it SLOWER.

---

**Q280.** What is the output?
```go
var once sync.Once
count := 0
for i := 0; i < 5; i++ {
    once.Do(func() { count++ })
}
fmt.Println(count)
```
- a) `5`
- b) `1`
- c) `0`
- d) Compile error

> ✅ **b** — `sync.Once` guarantees the function runs exactly once, regardless of how many times `Do` is called.

---

**Q281.** True or False: If the function passed to `sync.Once.Do` panics, future calls to `Do` will not retry it.

> ✅ **True** — If fn panics inside `Do`, the Once is considered "done". Future calls to `Do` are no-ops. The panic propagates to the goroutine that called `Do`.

---

**Q282.** What is `sync.Pool` designed for?

- a) Goroutine pool management
- b) Connection pooling for databases
- c) Reusing short-lived temporary objects to reduce GC pressure
- d) Memory pool for large allocations

> ✅ **c** — `sync.Pool` caches objects that would otherwise be frequently allocated and GC'd. Common use: `bytes.Buffer` objects.

---

**Q283.** True or False: Objects in `sync.Pool` are guaranteed to persist between GC cycles.

> ✅ **False** — The GC can clear the pool at any time. Pool objects may be collected. Never store objects you NEED to keep in a pool.

---

**Q284.** What must you do with a `sync.Pool` object before using it?

- a) Lock it with a mutex
- b) Reset/clear it — pooled objects may have leftover state from previous use
- c) Check if it's nil
- d) Increase its reference count

> ✅ **b** — `buf := pool.Get().(*bytes.Buffer); buf.Reset()` — always reset before use.

---

**Q285.** What is `sync.Map` optimized for?

- a) General-purpose concurrent map operations
- b) Maps where entries are written once and read many times, or where many goroutines write DIFFERENT keys
- c) Sorted map operations
- d) Maps with string keys only

> ✅ **b** — For mixed read/write with same keys, `sync.Mutex` + regular map is often better.

---

**Q286.** What does `sync.Map.LoadOrStore(key, value)` do?

- a) Loads if key exists, stores if not — atomically
- b) Conditionally stores only if a condition is met
- c) Loads and then stores the same value
- d) Stores first, then loads to verify

> ✅ **a** — Returns `(actual, loaded)`. If key existed: `actual=existingValue, loaded=true`. If new: `actual=value, loaded=false`.

---

**Q287.** True or False: `atomic.Int64` (Go 1.19+) is faster than a `sync.Mutex` for a single counter.

> ✅ **True** — Atomic operations are single CPU instructions. Mutex involves locking/unlocking overhead. For simple scalar values, atomics are significantly faster.

---

**Q288.** What does `atomic.CompareAndSwap(old, new)` return?

- a) The current value
- b) `true` if the swap was performed (current value == old), `false` otherwise
- c) The old value
- d) The new value

> ✅ **b** — CAS: atomically "if current == old, set to new, return true; else return false". Foundation of lock-free data structures.

---

**Q289.** What is the Lock ordering rule for preventing deadlocks?

- a) Always lock the mutex with the smallest memory address first
- b) Always lock multiple mutexes in the SAME consistent order across ALL goroutines
- c) Never hold more than one mutex at a time
- d) Use a global lock ordering protocol

> ✅ **b** — If goroutine 1 always locks A then B, and goroutine 2 also locks A then B (never B then A), deadlock cannot occur.

---

**Q290.** What does `sync.WaitGroup.Add(-1)` do?

- a) Compile error — negative values not allowed
- b) Decrements the counter by 1 (equivalent to Done())
- c) Sets the counter to -1
- d) Runtime panic

> ✅ **b** — `wg.Add(-1)` and `wg.Done()` are equivalent. `Done()` is defined as `wg.Add(-1)`.

---

**Q291.** True or False: Reading a map from multiple goroutines simultaneously (no writes) is safe in Go.

> ✅ **True** — Concurrent reads are safe. Only concurrent writes (or read + write) cause problems.

---

**Q292.** What is the `atomic.Pointer[T]` type used for?

- a) Storing a pointer that the GC tracks
- b) Atomically swappable pointer — lock-free configuration updates
- c) Unsafe pointer with GC bypass
- d) A typed wrapper around `unsafe.Pointer`

> ✅ **b** — Pattern: `var config atomic.Pointer[Config]; config.Store(newConfig)` — readers call `config.Load()` with no locking overhead.

---

**Q293.** What happens if `wg.Wait()` is called when the counter is already 0?

- a) Panics
- b) Returns immediately
- c) Blocks forever
- d) Resets the counter

> ✅ **b** — `Wait()` returns immediately if the count is 0.

---

**Q294.** True or False: `sync.Cond.Wait()` atomically releases the associated lock and suspends the goroutine.

> ✅ **True** — `cond.Wait()` = unlock mutex + suspend goroutine atomically. When woken, it re-acquires the lock before returning.

---

**Q295.** What does `sync.Cond.Broadcast()` do vs `Signal()`?

- a) `Signal()` wakes all waiters; `Broadcast()` wakes one
- b) `Broadcast()` wakes ALL waiting goroutines; `Signal()` wakes ONE
- c) They are identical
- d) `Broadcast()` is deprecated in Go 1.21+

> ✅ **b** — Use `Signal()` when only one waiter needs to react. Use `Broadcast()` when all waiters should wake up (e.g., "server is ready").

---

**Q296.** Why must `sync.Cond.Wait()` always be called inside a `for` loop instead of an `if` statement?

- a) It doesn't have to — `if` is fine
- b) To handle spurious wakeups — `Wait()` can return without `Signal/Broadcast` being called
- c) Performance requirement
- d) The condition might be violated by another goroutine between wakeup and lock re-acquisition

> ✅ **d** (and b) — Both are correct: spurious wakeups AND another goroutine might consume the "ready" state between when Signal is called and when this goroutine re-acquires the lock.

---

**Q297.** What is the recommended way to limit concurrent goroutines to N?

- a) `sync.WaitGroup` with a counter
- b) A buffered channel of size N (semaphore pattern)
- c) `runtime.GOMAXPROCS(N)`
- d) `atomic.Int32` counter with spin-wait

> ✅ **b** — Semaphore with buffered channel: `sem := make(chan struct{}, N)`. Acquire: `sem <- struct{}{}`. Release: `<-sem`.

---

**Q298.** True or False: `sync.Map.Range` is safe to call concurrently with `Store` and `Delete`.

> ✅ **True** — All operations on `sync.Map` are safe for concurrent use. That's its entire purpose.

---

**Q299.** What is the purpose of the `mu sync.Mutex` embedded in a struct (not as a named field)?

- a) It's an anti-pattern
- b) Promotes `Lock()` and `Unlock()` methods to the struct for convenient use
- c) Prevents the struct from being copied
- d) Both b and c

> ✅ **d** — Embedding `sync.Mutex` promotes `Lock`/`Unlock` to the struct AND makes the struct non-copyable (vet catches copies of locked mutexes).

---

**Q300.** What does `go test -race ./...` measure?

- a) Performance benchmarks
- b) Race conditions in concurrent code detected at runtime during test execution
- c) Static analysis of potential race conditions
- d) Correctness of goroutine ordering

> ✅ **b** — The race detector instruments memory accesses. It detects actual concurrent unsynchronized accesses when they OCCUR during test runs. Static analysis cannot find all races.

---

*End of Quiz Part 3 (Q201–Q300)*
*Continue with go_quiz_part4.md for Q301–Q400*
