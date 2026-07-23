# Go Deep Dive — Part 3: Control Flow — If, Switch, For (Complete)

---

## Chapter 1: The `if` Statement — Deep Dive

### Why Go's `if` is Different from C++

In C++ and Java, `if (condition)` takes a boolean expression. Go's `if` has an additional **initialization statement**. This is not just syntactic sugar — it fundamentally changes how you write Go code by scoping variables tightly.

### The Initialization Statement Pattern

```go
// Syntax: if initialization; condition { ... }

// Most common use: error handling IMPORTANT
file, err := os.Open("data.txt")
if err != nil {
    return err
}
// file and err are accessible here — they leaked out of the if

// Better: scope them tightly
if file, err := os.Open("data.txt"); err != nil {
    return err
} else {
    defer file.Close() // file is scoped to this if/else block
    // ... use file
}
// file and err are NOT accessible here — cleaner!

// Another common use: map lookup IMPORTANT
if val, ok := myMap["key"]; ok {
    fmt.Println("Found:", val)
} else {
    fmt.Println("Not found, default:", val) // val is zero value here
}

// Type assertion check
if s, ok := i.(string); ok {
    fmt.Println("It's a string:", s)
}
```

### If vs Switch — When to Use Which

```go
// Use if when:
// 1. Conditions involve different variables
if err != nil {
    ...
} else if ctx.Err() != nil {
    ...
}

// 2. Conditions are complex boolean expressions
if len(data) > 0 && data[0] == 0xFF {
    ...
}

// Use switch when:
// 1. Comparing one variable against multiple values
switch status {
case "active", "pending":
    ...
case "deleted":
    ...
}

// 2. Type dispatch (type switch)
switch v := i.(type) {
case int:
    ...
}

// 3. Replacing long if/else chains
```

### Negative Cases — Common Mistakes

```go
// MISTAKE 1: Assignment where comparison was intended
// Go prevents this compile-time unlike C++:
if x = calculateValue(); x > 0 {  // assignment in condition — valid Go!
    // This is legal — x is assigned then tested
}
// But this looks like == which can confuse readers
// Prefer the explicit initialization form: if x := ...; x > 0 {}

// MISTAKE 2: Not using else when you should
func getGrade(score int) string {
    if score >= 90 {
        return "A"
    }
    if score >= 80 {    // BAD: if score >= 90 was already returned
        return "B"      // so this is fine, but it's not obvious
    }
    return "C"
}

// BETTER: use else if to show the conditions are mutually exclusive
func getGrade2(score int) string {
    if score >= 90 {
        return "A"
    } else if score >= 80 {
        return "B"
    } else if score >= 70 {
        return "C"
    } else {
        return "F"
    }
}

// BEST: use switch for this pattern IMPORTANT
func getGrade3(score int) string {
    switch {
    case score >= 90: return "A"
    case score >= 80: return "B"
    case score >= 70: return "C"
    default: return "F"
    }
}
```

### Testing if Branches

```go
func TestGradeAssignment(t *testing.T) {
    tests := []struct {
        score    int
        expected string
    }{
        // Positive cases — normal grades
        {95, "A"},
        {85, "B"},
        {75, "C"},
        {55, "F"},
        
        // Boundary cases — most important to test!
        {90, "A"},   // exactly at A threshold
        {89, "B"},   // just below A
        {80, "B"},   // exactly at B threshold
        {79, "C"},   // just below B
        {70, "C"},   // exactly at C threshold
        {69, "F"},   // just below C
        
        // Edge cases
        {0, "F"},
        {100, "A"},
        {-1, "F"},   // negative score
    }
    
    for _, tt := range tests {
        t.Run(fmt.Sprintf("score_%d", tt.score), func(t *testing.T) {
            got := getGrade3(tt.score)
            if got != tt.expected {
                t.Errorf("score %d: got %q, want %q", tt.score, got, tt.expected)
            }
        })
    }
}
```

---

## Chapter 2: The `switch` Statement — Complete Reference

### Why Go's Switch is Better than C++'s

In C++, `switch` has two notorious problems:
1. **Fall-through by default**: Forgetting `break` causes bugs (Apple SSL bug is a famous example)
2. **Limited to integers**: Can't switch on strings or custom comparisons

Go fixes both:
1. **No fall-through by default**: Each case is independent
2. **Works on any comparable type**: strings, structs, interfaces...

```go
// C++ switch (fall-through by default)
// switch (x) {
//     case 1: doA(); // FALLS THROUGH to case 2!
//     case 2: doB(); break;
//     case 3: doC(); break;
// }

// Go switch (no fall-through by default)
switch x {
case 1:
    doA() // execution stops here
case 2:
    doB()
case 3:
    doC()
}
```

### All Switch Forms

```go
// Form 1: Switch on a value
switch status {
case "active":
    fmt.Println("Active")
case "inactive":
    fmt.Println("Inactive")
case "deleted", "archived": // multiple values per case
    fmt.Println("Removed")
default:
    fmt.Println("Unknown status")
}

// Form 2: Switch with initialization
switch status := getStatus(); status {
case "active":
    fmt.Println("Active:", status)
default:
    fmt.Println("Other:", status)
}

// Form 3: Switch with no condition (replaces if/else chains)
score := 85
switch {
case score >= 90:
    fmt.Println("Outstanding")
case score >= 80:
    fmt.Println("Good")
case score >= 70:
    fmt.Println("Average")
default:
    fmt.Println("Needs improvement")
}

// Form 4: Type switch (with interface)
func describe(i interface{}) {
    switch v := i.(type) {
    case int:
        fmt.Printf("int: %d\n", v)
    case float64:
        fmt.Printf("float64: %.2f\n", v)
    case string:
        fmt.Printf("string: %q (len=%d)\n", v, len(v))
    case bool:
        fmt.Printf("bool: %t\n", v)
    case []int:
        fmt.Printf("[]int: %v (len=%d)\n", v, len(v))
    case nil:
        fmt.Println("nil")
    default:
        fmt.Printf("unknown type: %T\n", v) // %T prints the type name
    }
}

// Form 5: fallthrough (explicit, rare)
switch n {
case 1:
    fmt.Println("one")
    fallthrough // explicitly fall through to next case
case 2:
    fmt.Println("two") // runs if n==1 OR n==2
case 3:
    fmt.Println("three") // only runs if n==3
}

// When fallthrough is useful: tagging multiple conditions
// Example: HTTP status code ranges
switch {
case code >= 500:
    log.Errorf("server error: %d", code)
    fallthrough // also do what 4xx does:
case code >= 400:
    metrics.RecordError(code)
    fallthrough // also do what all errors do:
default:
    logRequest(code) // always log the request
}
```

### Type Switch — The Interface Dispatcher

```go
// Most practical use: processing commands in a system
type Command interface{}
type CreateUser struct { Name, Email string }
type DeleteUser struct { ID int }
type UpdateUser struct { ID int; Name string }

func processCommand(cmd Command) error {
    switch c := cmd.(type) {
    case CreateUser:
        return createUser(c.Name, c.Email)
    case DeleteUser:
        return deleteUser(c.ID)
    case UpdateUser:
        return updateUser(c.ID, c.Name)
    default:
        return fmt.Errorf("unknown command type: %T", c)
    }
}

func TestProcessCommand(t *testing.T) {
    tests := []struct {
        name    string
        cmd     Command
        wantErr bool
    }{
        {"create user", CreateUser{"Alice", "alice@example.com"}, false},
        {"delete user", DeleteUser{ID: 1}, false},
        {"update user", UpdateUser{1, "Alice Updated"}, false},
        {"unknown command", struct{}{}, true}, // negative case
        {"nil command", nil, true},             // negative case
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := processCommand(tt.cmd)
            if (err != nil) != tt.wantErr {
                t.Errorf("processCommand() err = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## Chapter 3: The `for` Loop — Every Form Explained

### Why Go Has Only One Loop Keyword

Go's designers believed that multiple loop keywords (`while`, `do-while`, `for`) create choices where there should be one obvious way. The `for` keyword handles all cases cleanly.

```go
// Traditional for (C-style)
for init; condition; post {
    // init: runs once before the loop
    // condition: checked before each iteration
    // post: runs after each iteration
}

// While-style (condition only)
for condition {
    // equivalent to: for ; condition ; { }
}

// Infinite loop
for {
    // equivalent to: for true { }
}

// Range-based (see next section)
for key, value := range collection {
}
```

### Range — The Most Important Loop Form

```go
// Range over SLICE
nums := []int{10, 20, 30, 40, 50}

for i, v := range nums {        // both index and value
    fmt.Println(i, v)
}
for i := range nums {           // index only
    fmt.Println(i)
}
for _, v := range nums {        // value only
    fmt.Println(v)
}

// Range over ARRAY (same as slice)
arr := [3]int{1, 2, 3}
for i, v := range arr { ... }

// Range over MAP (random order! — by design, prevents depending on order)
m := map[string]int{"a": 1, "b": 2, "c": 3}
for k, v := range m {
    fmt.Printf("%s: %d\n", k, v)
}
for k := range m {              // keys only
    fmt.Println(k)
}

// Range over STRING (iterates RUNES, not bytes!)
for i, ch := range "Hello, 世界" {
    fmt.Printf("index=%d char=%c\n", i, ch)
    // index 7 = 世 (first byte of the 3-byte sequence)
    // index 10 = 界
}

// Range over CHANNEL (blocks until channel closed)
ch := make(chan int)
go func() {
    ch <- 1; ch <- 2; ch <- 3
    close(ch)
}()
for n := range ch {
    fmt.Println(n) // 1, 2, 3, then loop ends when ch is closed
}

// Range with only one variable (Go 1.22+, range over integer!)
for i := range 5 {  // Go 1.22+
    fmt.Println(i) // 0, 1, 2, 3, 4
}
```

### Loop Modification Patterns — Positive Cases

```go
// Pattern 1: Modifying a slice while ranging
// Wrong — modifying slice during range is tricky
nums := []int{1, 2, 3, 4, 5}

// WRONG: modifying underlying slice doesn't work as expected when deleting
for i, n := range nums {
    if n%2 == 0 {
        nums = append(nums[:i], nums[i+1:]...) // BUG: i is still the same
    }
}

// CORRECT: Collect then process
var result []int
for _, n := range nums {
    if n%2 != 0 {
        result = append(result, n)
    }
}
nums = result

// Pattern 2: Modifying values in-place
// Range variables are COPIES — modifying them doesn't change the original
type Person struct { Name string; Age int }
people := []Person{{"Alice", 30}, {"Bob", 25}}

for _, p := range people {
    p.Age++ // modifies the COPY, not the original!
}
fmt.Println(people[0].Age) // still 30

// Correct: use index
for i := range people {
    people[i].Age++
}
fmt.Println(people[0].Age) // 31

// Or: slice of pointers
peoplePtrs := []*Person{{"Alice", 30}, {"Bob", 25}}
for _, p := range peoplePtrs {
    p.Age++ // modifies the struct -- p is a pointer copy, still points to original
}
fmt.Println(peoplePtrs[0].Age) // 31
```

### Nested Loops and Labels

```go
// Labels for breaking out of nested loops
matrix := [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
target := 5

var foundRow, foundCol int = -1, -1

outer:
    for row, r := range matrix {
        for col, val := range r {
            if val == target {
                foundRow, foundCol = row, col
                break outer // breaks BOTH loops
            }
        }
    }

if foundRow != -1 {
    fmt.Printf("Found %d at [%d][%d]\n", target, foundRow, foundCol)
}

// Continue with label — go to next iteration of OUTER loop
search:
    for i, row := range matrix {
        for _, val := range row {
            if val < 0 {
                fmt.Printf("Row %d has negative value\n", i)
                continue search // skip rest of this row
            }
        }
    }

// goto — rare but legal
i := 0
loop:
    if i < 5 {
        fmt.Println(i)
        i++
        goto loop // jump to label
    }
// goto is occasionally useful in generated code or state machines
```

### Loop Performance Patterns

```go
// Avoid function calls in loop condition
// BAD: len(slice) is called every iteration (technically O(1) in Go, but still)
for i := 0; i < len(hugeSlice); i++ { ... }

// GOOD: if you're doing complex condition computation
n := len(hugeSlice)
for i := 0; i < n; i++ { ... }

// Avoid unnecessary work in loop body
// BAD: repeated map lookup
for _, id := range userIDs {
    if config["timeout"] > 0 { // map lookup every iteration!
        process(id, config["timeout"])
    }
}

// GOOD: extract constant computations out of loop
timeout := config["timeout"]
if timeout > 0 {
    for _, id := range userIDs {
        process(id, timeout)
    }
}

// GOOD: Use range over index when working with parallel data
for i, user := range users {
    orders[i] = createOrder(user) // parallel arrays
}

// Benchmark: range vs traditional for
func BenchmarkRangeVsTraditional(b *testing.B) {
    data := make([]int, 10000)
    for i := range data {
        data[i] = i
    }
    
    b.Run("range", func(b *testing.B) {
        for n := 0; n < b.N; n++ {
            sum := 0
            for _, v := range data { sum += v }
            _ = sum
        }
    })
    
    b.Run("traditional", func(b *testing.B) {
        for n := 0; n < b.N; n++ {
            sum := 0
            for i := 0; i < len(data); i++ { sum += data[i] }
            _ = sum
        }
    })
}
// In practice, these are equivalent speed — the compiler optimizes both
```

---

## Chapter 4: The `goto` Statement

```go
// goto is legal in Go but rarely needed
// Valid uses: breaking out of deeply nested code, state machines

// Pattern: cleanup on error (rare — prefer defer)
func processLegacy(input []byte) error {
    buf := make([]byte, 1024)
    
    if len(input) == 0 {
        goto cleanup
    }
    
    if _, err := process(buf, input); err != nil {
        goto cleanup
    }
    
    return nil
    
cleanup:
    // cleanup code...
    return nil
}

// goto CANNOT:
// - Jump over variable declarations
// - Jump into a block from outside

// This ERRORS:
// goto end
// x := 10    // ERROR: cannot jump over declaration of x
// end:
// fmt.Println(x)
```

---

## Chapter 5: Comprehensive Testing for Control Flow

```go
package control_test

import (
    "testing"
    "fmt"
)

// Test all branches of if/else
func TestAllBranches(t *testing.T) {
    // Use a helper that counts which branch was taken
    branchCounts := make(map[string]int)
    
    classify := func(n int) string {
        switch {
        case n < 0:
            branchCounts["negative"]++
            return "negative"
        case n == 0:
            branchCounts["zero"]++
            return "zero"
        case n < 100:
            branchCounts["small positive"]++
            return "small positive"
        default:
            branchCounts["large positive"]++
            return "large positive"
        }
    }
    
    tests := []struct { n int; want string }{
        {-5, "negative"}, {-1, "negative"}, {0, "zero"},
        {1, "small positive"}, {50, "small positive"}, {99, "small positive"},
        {100, "large positive"}, {1000, "large positive"},
    }
    
    for _, tt := range tests {
        t.Run(fmt.Sprintf("n=%d", tt.n), func(t *testing.T) {
            got := classify(tt.n)
            if got != tt.want {
                t.Errorf("classify(%d) = %q, want %q", tt.n, got, tt.want)
            }
        })
    }
    
    // Ensure all branches were tested
    if branchCounts["negative"] == 0 { t.Error("negative branch not covered") }
    if branchCounts["zero"] == 0 { t.Error("zero branch not covered") }
    if branchCounts["small positive"] == 0 { t.Error("small positive branch not covered") }
    if branchCounts["large positive"] == 0 { t.Error("large positive branch not covered") }
}

// Table-driven test for loops
func sum(nums []int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

func TestSum(t *testing.T) {
    tests := []struct {
        name     string
        input    []int
        expected int
    }{
        {"empty slice", []int{}, 0},           // edge: empty
        {"single element", []int{5}, 5},        // edge: single
        {"all positive", []int{1, 2, 3}, 6},   // normal
        {"all negative", []int{-1, -2, -3}, -6}, // negative values
        {"mixed", []int{1, -1, 2, -2}, 0},      // sum to zero
        {"large numbers", []int{1000000, 2000000}, 3000000}, // large
        {"nil slice", nil, 0},                  // nil input
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := sum(tt.input)
            if got != tt.expected {
                t.Errorf("sum(%v) = %d, want %d", tt.input, got, tt.expected)
            }
        })
    }
}

// Test break/continue behavior
func TestBreakBehavior(t *testing.T) {
    // find first negative
    findFirst := func(nums []int) int {
        for _, n := range nums {
            if n < 0 {
                return n
            }
        }
        return 0 // not found
    }
    
    if got := findFirst([]int{1, 2, -3, -4}); got != -3 {
        t.Errorf("expected -3, got %d", got)
    }
    if got := findFirst([]int{1, 2, 3}); got != 0 {
        t.Errorf("expected 0, got %d", got)
    }
}

// Test that range over map works correctly (order doesn't matter)
func TestMapRange(t *testing.T) {
    counts := map[string]int{
        "a": 1, "b": 2, "c": 3, "d": 4,
    }
    
    total := 0
    seen := make(map[string]bool)
    
    for k, v := range counts {
        if seen[k] {
            t.Errorf("key %q seen twice in range", k)
        }
        seen[k] = true
        total += v
    }
    
    if total != 10 {
        t.Errorf("total = %d, want 10", total)
    }
    if len(seen) != 4 {
        t.Errorf("saw %d keys, want 4", len(seen))
    }
}

// Test switch fallthrough
func TestFallthrough(t *testing.T) {
    result := ""
    
    switch 2 {
    case 1:
        result += "one:"
        fallthrough
    case 2:
        result += "two:"
        fallthrough
    case 3:
        result += "three"
    case 4:
        result += "four" // should not run
    }
    
    if result != "two:three" {
        t.Errorf("fallthrough: got %q, want %q", result, "two:three")
    }
}
```

Run with: `go test -v -cover ./...`

---

## Chapter 6: Practical Real-World Control Flow

### HTTP Request Router (shows switch + error handling)

```go
package main

import (
    "fmt"
    "net/http"
    "strings"
)

type Router struct {
    routes map[string]http.HandlerFunc
}

func NewRouter() *Router {
    return &Router{routes: make(map[string]http.HandlerFunc)}
}

func (r *Router) Handle(path string, handler http.HandlerFunc) {
    r.routes[path] = handler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    // Clean the path
    path := strings.TrimRight(req.URL.Path, "/")
    if path == "" {
        path = "/"
    }
    
    // Route based on method + path combination
    switch {
    case req.Method == "GET" && path == "/":
        r.routes["/"]
        if handler, ok := r.routes["GET /"]; ok {
            handler(w, req)
        }
    default:
        key := req.Method + " " + path
        if handler, ok := r.routes[key]; ok {
            handler(w, req)
        } else {
            http.NotFound(w, req)
        }
    }
}
```

### Retry Loop with Exponential Backoff

```go
package main

import (
    "fmt"
    "math"
    "time"
)

func withRetry(maxAttempts int, fn func() error) error {
    var lastErr error
    
    for attempt := 0; attempt < maxAttempts; attempt++ {
        // Exponential backoff: 0ms, 100ms, 200ms, 400ms...
        if attempt > 0 {
            backoff := time.Duration(math.Pow(2, float64(attempt-1))) * 100 * time.Millisecond
            time.Sleep(backoff)
        }
        
        if err := fn(); err == nil {
            return nil // success!
        } else {
            lastErr = err
            fmt.Printf("Attempt %d/%d failed: %v\n", attempt+1, maxAttempts, err)
        }
    }
    
    return fmt.Errorf("all %d attempts failed, last error: %w", maxAttempts, lastErr)
}

// Test the retry logic
func TestWithRetry(t *testing.T) {
    t.Run("succeeds on first try", func(t *testing.T) {
        calls := 0
        err := withRetry(3, func() error {
            calls++
            return nil
        })
        if err != nil { t.Errorf("unexpected error: %v", err) }
        if calls != 1 { t.Errorf("expected 1 call, got %d", calls) }
    })
    
    t.Run("succeeds on third try", func(t *testing.T) {
        calls := 0
        err := withRetry(5, func() error {
            calls++
            if calls < 3 {
                return fmt.Errorf("not ready yet")
            }
            return nil
        })
        if err != nil { t.Errorf("unexpected error: %v", err) }
        if calls != 3 { t.Errorf("expected 3 calls, got %d", calls) }
    })
    
    t.Run("all attempts fail", func(t *testing.T) {
        calls := 0
        err := withRetry(3, func() error {
            calls++
            return fmt.Errorf("always fails")
        })
        if err == nil { t.Error("expected error, got nil") }
        if calls != 3 { t.Errorf("expected 3 calls, got %d", calls) }
    })
}
```

---

**Summary of Part 3:**
- `if` with initialization statement scopes variables tightly — prefer it
- Go's switch has no fall-through by default — eliminates a major C++ bug class
- Type switch is the idiomatic way to dispatch on interface types
- `for` is the only loop keyword — it handles all loop patterns
- Range iterates runes (not bytes) over strings — critical for Unicode
- Range over map has random order — never depend on iteration order
- Labels enable breaking/continuing from nested loops
- Testing control flow requires covering all branches including boundary conditions
