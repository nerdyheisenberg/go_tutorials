# Go Deep Dive — Part 9: Error Handling — The Go Way

---

## Chapter 1: Philosophy — Why Not Exceptions?

### The C++ Exception Model

In C++, exceptions create an invisible control flow path:

```cpp
int getUser(int id) {
    auto row = db.query(id); 
    // ^ This might throw. From here, execution can jump ANYWHERE.
    // The reader must mentally track all possible throw paths.
    return row.id;
}

void processOrder(int userId) {
    try {
        auto user = getUser(userId);  // might throw from 3 levels deep
        processUserData(user);
        sendConfirmation(user);
    } catch (DatabaseException& e) {
        // ...
    } catch (NetworkException& e) {
        // ...
    }
}
```

**Problems with exceptions:**
1. **Invisible control flow**: You can't tell by reading `getUser()` that it might throw
2. **Exception safety is hard**: RAII helps, but it's easy to get wrong
3. **Performance**: Exception handling adds overhead to every `try` block
4. **Overused**: In practice, exceptions are often used for non-exceptional situations

### Go's Error Values

Go treats errors as **ordinary values** — they're returned like any other value:

```go
func getUser(id int) (*User, error) {
    row, err := db.QueryRow(id)
    if err != nil {
        return nil, err  // error is VISIBLE — caller must handle it
    }
    return row.ToUser(), nil
}

func processOrder(userID int) error {
    user, err := getUser(userID)  // error handling is EXPLICIT
    if err != nil {
        return fmt.Errorf("processOrder: %w", err)  // add context, propagate
    }
    
    if err := processUserData(user); err != nil {
        return fmt.Errorf("processOrder: processUserData: %w", err)
    }
    
    if err := sendConfirmation(user); err != nil {
        return fmt.Errorf("processOrder: sendConfirmation: %w", err)
    }
    
    return nil
}
```

**Benefits:**
1. **Explicit**: Every possible error is visible at the call site
2. **Composable**: Errors are values — you can store, transform, and wrap them
3. **Predictable**: No hidden control flow jumps
4. **Simple model**: The `error` interface has one method: `Error() string`

---

## Chapter 2: The `error` Interface

```go
// The entire error interface:
type error interface {
    Error() string
}

// That's it. Any type with Error() string satisfies it.
```

### Creating Errors — Four Ways

```go
// 1. errors.New — simple static error
import "errors"

var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")

// These are SENTINEL errors — compare with errors.Is()
// Note: var (not const) — error values are pointer-like (comparable by address)

// 2. fmt.Errorf — formatted error with optional wrapping
err := fmt.Errorf("user %d not found in region %s", userID, region)
wrappedErr := fmt.Errorf("handler: %w", originalErr)  // %w wraps the error

// 3. Custom error type — for structured error information
type NotFoundError struct {
    Resource string
    ID       int
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s with ID %d not found", e.Resource, e.ID)
}

// 4. errors.Join — combine multiple errors (Go 1.20+)
err1 := errors.New("field 'name' is required")
err2 := errors.New("field 'email' is invalid")
combined := errors.Join(err1, err2)
// combined.Error() = "field 'name' is required\nfield 'email' is invalid"
```

---

## Chapter 3: Error Wrapping and Unwrapping

### Why Wrap Errors?

Wrapping adds **context** while **preserving the original error** for inspection:

```go
// Without wrapping:
return err  // "connection refused"
// Caller has no idea WHERE the connection was refused!

// With wrapping:
return fmt.Errorf("connecting to database at %s: %w", dsn, err)
// "connecting to database at postgres://localhost:5432/mydb: connection refused"
// Now caller knows WHERE and WHY

// Chain of wrapped errors:
// "processOrder: getUser: queryDatabase: connection refused"
```

### errors.Is — Checking Error Identity

```go
// errors.Is checks the ENTIRE error chain
var ErrNotFound = errors.New("not found")

err := fmt.Errorf("getUserByEmail: %w",
    fmt.Errorf("db.Query: %w", ErrNotFound))

// Simple == comparison:
err == ErrNotFound  // false! err is a wrapped error, not ErrNotFound directly

// errors.Is traverses the chain:
errors.Is(err, ErrNotFound)  // true! Found ErrNotFound in the chain

// errors.Is uses the Is() method if defined:
type TimeoutError struct{ Duration time.Duration }
func (e *TimeoutError) Is(target error) bool {
    _, ok := target.(*TimeoutError)
    return ok  // true for any *TimeoutError target
}
```

### errors.As — Extracting Error Types

```go
// errors.As finds a specific type in the error chain and extracts it
type ValidationError struct {
    Field   string
    Message string
}
func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s: %s", e.Field, e.Message)
}

err := fmt.Errorf("handler: %w", &ValidationError{Field: "email", Message: "invalid format"})

var valErr *ValidationError
if errors.As(err, &valErr) {
    // valErr is now the *ValidationError, fully typed
    fmt.Println("Field:", valErr.Field)    // "email"
    fmt.Println("Message:", valErr.Message) // "invalid format"
    
    // Can now return appropriate HTTP status, provide field-specific feedback, etc.
}

// errors.As also traverses the chain just like errors.Is
```

---

## Chapter 4: Error Patterns in Detail

### Pattern 1: Sentinel Errors (Package-Level Errors)

```go
// Define sentinel errors at package level
var (
    // Exported — callers can check with errors.Is
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrConflict     = errors.New("already exists")
    
    // Unexported — internal use only
    errInvalidState = errors.New("invalid internal state")
)

func GetUser(id int) (*User, error) {
    if id <= 0 {
        return nil, fmt.Errorf("GetUser: invalid id %d: %w", id, ErrNotFound)
    }
    // ...
    return nil, nil
}

// Caller:
user, err := GetUser(0)
if errors.Is(err, ErrNotFound) {
    // Handle "not found" specifically
    http.Error(w, "User not found", http.StatusNotFound)
    return
}
if err != nil {
    // Handle other errors generically
    http.Error(w, "Internal error", http.StatusInternalServerError)
    return
}
```

### Pattern 2: Structured Error Types

```go
// When you need to carry more information with an error
type APIError struct {
    Code       int    // HTTP status code
    Message    string // Human-readable message
    Details    string // Debug info (don't send to users!)
    RequestID  string // For tracking in logs
}

func (e *APIError) Error() string {
    return fmt.Sprintf("[%d] %s (request: %s)", e.Code, e.Message, e.RequestID)
}

// Constructor for common cases
func ErrAPI(code int, msg string) *APIError {
    return &APIError{Code: code, Message: msg}
}

func handleRequest(r *http.Request) error {
    if r.Header.Get("Authorization") == "" {
        return &APIError{
            Code:      401,
            Message:   "Authentication required",
            Details:   "No Authorization header present",
            RequestID: r.Header.Get("X-Request-ID"),
        }
    }
    return nil
}

// Middleware that handles APIError:
func errorMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            // let next.ServeHTTP run first (in actual code, you'd use different approach)
        }()
        next.ServeHTTP(w, r)
    })
}
```

### Pattern 3: Error Context Chain

```go
// The standard pattern for adding context to errors:
// func f() error {
//     if err := g(); err != nil {
//         return fmt.Errorf("f: %w", err)
//     }
//     return nil
// }
//
// This creates: "topLevel: middleLayer: bottomLayer: original error message"
// Which gives a complete call-path context

package repository

func (r *Repo) GetOrder(ctx context.Context, id int) (*Order, error) {
    row := r.db.QueryRowContext(ctx, "SELECT * FROM orders WHERE id = $1", id)
    var o Order
    if err := row.Scan(&o.ID, &o.CustomerID, &o.Amount); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, fmt.Errorf("GetOrder(%d): %w", id, ErrNotFound)
        }
        return nil, fmt.Errorf("GetOrder(%d): scanning: %w", id, err)
    }
    return &o, nil
}

// service layer:
func (s *OrderService) GetOrder(ctx context.Context, id int) (*Order, error) {
    order, err := s.repo.GetOrder(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("OrderService.GetOrder: %w", err)  // add layer
    }
    return order, nil
}

// handler layer:
func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
    id, _ := strconv.Atoi(r.PathValue("id"))
    order, err := h.service.GetOrder(r.Context(), id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            http.Error(w, "Order not found", 404)
            return
        }
        log.Printf("getOrder: %v", err)  // full chain logged
        http.Error(w, "Internal error", 500)
        return
    }
    json.NewEncoder(w).Encode(order)
}
// Final error: "OrderService.GetOrder: GetOrder(42): not found"
```

### Pattern 4: Multiple Error Aggregation

```go
// When you want to collect ALL errors (not just the first):
func validateUser(u User) error {
    var errs []error
    
    if u.Name == "" {
        errs = append(errs, errors.New("name is required"))
    } else if len(u.Name) > 100 {
        errs = append(errs, fmt.Errorf("name too long: %d chars (max 100)", len(u.Name)))
    }
    
    if u.Email == "" {
        errs = append(errs, errors.New("email is required"))
    } else if !isValidEmail(u.Email) {
        errs = append(errs, fmt.Errorf("invalid email: %s", u.Email))
    }
    
    if u.Age < 0 || u.Age > 150 {
        errs = append(errs, fmt.Errorf("age out of range: %d", u.Age))
    }
    
    return errors.Join(errs...)  // nil if errs is empty
}

// Custom multi-error type (before Go 1.20):
type MultiError struct {
    Errors []error
}

func (m *MultiError) Error() string {
    msgs := make([]string, len(m.Errors))
    for i, err := range m.Errors {
        msgs[i] = err.Error()
    }
    return strings.Join(msgs, "; ")
}

func (m *MultiError) Unwrap() []error {
    return m.Errors  // enables errors.Is/As to traverse all errors
}
```

### Pattern 5: Retry with Error Classification

```go
type RetryableError struct{ Err error }
func (e *RetryableError) Error() string { return "retryable: " + e.Err.Error() }
func (e *RetryableError) Unwrap() error { return e.Err }

func isRetryable(err error) bool {
    var re *RetryableError
    return errors.As(err, &re)
}

func callWithRetry(maxTries int, fn func() error) error {
    var lastErr error
    for i := 0; i < maxTries; i++ {
        if err := fn(); err != nil {
            if !isRetryable(err) {
                return err  // non-retryable: fail immediately
            }
            lastErr = err
            time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
            continue
        }
        return nil  // success
    }
    return fmt.Errorf("all %d attempts failed: %w", maxTries, lastErr)
}
```

---

## Chapter 5: The `must` Pattern

```go
// For cases where you KNOW the error can't happen
// (program logic guarantees it), use a must helper

func must(v interface{}, err error) interface{} {
    if err != nil {
        panic(err)
    }
    return v
}

// Example: parsing a URL that's hardcoded in source
var baseURL = must(url.Parse("https://api.example.com")).(*url.URL)

// Typed version (Go generics):
func Must[T any](v T, err error) T {
    if err != nil {
        panic(err)
    }
    return v
}

baseURL := Must(url.Parse("https://api.example.com"))
re := Must(regexp.Compile(`^\d{4}-\d{2}-\d{2}$`))
// If these panic, it means your source code has a bug (wrong URL/regex)
// These would fail at program startup, not in production for a user
```

---

## Chapter 6: Error Handling Anti-Patterns

### Anti-Pattern 1: Ignoring errors

```go
// NEVER DO THIS:
os.Remove("tempfile.txt")     // ignoring error — file might not be deleted!
json.Unmarshal(data, &result) // ignoring error — result might be garbage!

// ALWAYS handle:
if err := os.Remove("tempfile.txt"); err != nil {
    log.Printf("failed to remove temp file: %v", err)
    // decide: return error? continue? log and continue?
}

// Exception: when you truly don't care (document WHY)
_ = os.Remove("tempfile.txt") // intentional ignore: cleanup, failure is non-fatal
```

### Anti-Pattern 2: Wrapping without context

```go
// BAD: adds no information
return fmt.Errorf("error: %w", err)  // "error: original message" — redundant prefix!

// BAD: loses the original error (no %w, just %v)
return fmt.Errorf("database error: %v", err)  // callers can't use errors.Is/As!

// GOOD: meaningful context + wrapped
return fmt.Errorf("GetUser(id=%d): %w", id, err)
```

### Anti-Pattern 3: Panic for control flow

```go
// BAD: using panic as exception
func findUser(id int) *User {
    if id <= 0 {
        panic("invalid id")  // never do this for business logic errors
    }
    ...
}

// GOOD: return error
func findUser(id int) (*User, error) {
    if id <= 0 {
        return nil, fmt.Errorf("invalid id: %d", id)
    }
    ...
}
```

### Anti-Pattern 4: Swallowing errors in defers

```go
// BAD: error from Close is lost
defer file.Close()

// BETTER: capture error (though in main flow, the original error matters more)
defer func() {
    if err := file.Close(); err != nil {
        log.Printf("failed to close file: %v", err)
    }
}()

// BEST for functions that return error: use named return
func process() (err error) {
    file, err := os.Open("data.txt")
    if err != nil {
        return
    }
    defer func() {
        if cerr := file.Close(); cerr != nil && err == nil {
            err = cerr  // only override if no other error
        }
    }()
    // ... process
    return
}
```

---

## Chapter 7: Comprehensive Error Handling Tests

```go
package errors_test

import (
    "errors"
    "fmt"
    "testing"
)

// ==================== SENTINEL ERRORS ====================

var ErrNotFound = errors.New("not found")
var ErrForbidden = errors.New("forbidden")

func findItem(id int) error {
    if id == 0 {
        return fmt.Errorf("findItem(%d): %w", id, ErrNotFound)
    }
    if id < 0 {
        return fmt.Errorf("findItem(%d): %w", id, ErrForbidden)
    }
    return nil
}

func TestSentinelErrors(t *testing.T) {
    tests := []struct {
        name    string
        id      int
        isNotFound bool
        isForbidden bool
        wantErr bool
    }{
        {"valid id", 1, false, false, false},
        {"not found", 0, true, false, true},
        {"forbidden", -1, false, true, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := findItem(tt.id)
            
            if (err != nil) != tt.wantErr {
                t.Fatalf("findItem(%d) err = %v, wantErr %v", tt.id, err, tt.wantErr)
            }
            
            if errors.Is(err, ErrNotFound) != tt.isNotFound {
                t.Errorf("errors.Is(ErrNotFound) = %v, want %v", 
                    errors.Is(err, ErrNotFound), tt.isNotFound)
            }
            
            if errors.Is(err, ErrForbidden) != tt.isForbidden {
                t.Errorf("errors.Is(ErrForbidden) = %v, want %v",
                    errors.Is(err, ErrForbidden), tt.isForbidden)
            }
        })
    }
}

// ==================== STRUCTURED ERRORS ====================

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s: %s", e.Field, e.Message)
}

func validateAge(age int) error {
    if age < 0 {
        return fmt.Errorf("validateAge: %w", 
            &ValidationError{Field: "age", Message: "must be non-negative"})
    }
    if age > 150 {
        return fmt.Errorf("validateAge: %w",
            &ValidationError{Field: "age", Message: "unrealistically large"})
    }
    return nil
}

func TestStructuredErrors(t *testing.T) {
    tests := []struct {
        name         string
        age          int
        wantErr      bool
        wantField    string
    }{
        {"valid age", 25, false, ""},
        {"zero age", 0, false, ""},
        {"negative age", -1, true, "age"},
        {"age too large", 200, true, "age"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateAge(tt.age)
            
            if (err != nil) != tt.wantErr {
                t.Fatalf("validateAge(%d) error = %v, wantErr = %v", tt.age, err, tt.wantErr)
            }
            
            if tt.wantErr {
                var ve *ValidationError
                if !errors.As(err, &ve) {
                    t.Fatalf("expected *ValidationError, got %T", err)
                }
                if ve.Field != tt.wantField {
                    t.Errorf("ValidationError.Field = %q, want %q", ve.Field, tt.wantField)
                }
            }
        })
    }
}

// ==================== ERROR WRAPPING CHAIN ====================

func layer3() error { return ErrNotFound }
func layer2() error { return fmt.Errorf("layer2: %w", layer3()) }
func layer1() error { return fmt.Errorf("layer1: %w", layer2()) }

func TestErrorChain(t *testing.T) {
    err := layer1()
    
    // Should be able to find ErrNotFound anywhere in chain
    if !errors.Is(err, ErrNotFound) {
        t.Error("errors.Is should find ErrNotFound in chain")
    }
    
    // Error message should contain all layers
    msg := err.Error()
    if !contains(msg, "layer1") { t.Error("error should contain 'layer1'") }
    if !contains(msg, "layer2") { t.Error("error should contain 'layer2'") }
    if !contains(msg, "not found") { t.Error("error should contain 'not found'") }
}

// ==================== MULTI-ERROR ====================

func validate(name string, age int) error {
    var errs []error
    if name == "" { errs = append(errs, errors.New("name required")) }
    if age < 0 { errs = append(errs, errors.New("age must be positive")) }
    if age > 150 { errs = append(errs, errors.New("age too large")) }
    return errors.Join(errs...)
}

func TestMultiError(t *testing.T) {
    // No errors
    if err := validate("Alice", 25); err != nil {
        t.Errorf("valid input: unexpected error: %v", err)
    }
    
    // Single error
    err := validate("", 25)
    if err == nil { t.Fatal("empty name: expected error") }
    
    // Multiple errors
    err2 := validate("", -1)
    if err2 == nil { t.Fatal("multiple invalid: expected error") }
    
    msg := err2.Error()
    if !contains(msg, "name required") { t.Error("should mention name error") }
    if !contains(msg, "age") { t.Error("should mention age error") }
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
    for i := 0; i <= len(s)-len(sub); i++ {
        if s[i:i+len(sub)] == sub {
            return true
        }
    }
    return false
}
```

Run: `go test -v ./...`

---

**Summary of Part 9:**
- Errors are values — return them, don't throw them
- The `error` interface has ONE method: `Error() string`
- Wrap with `%w` to add context while preserving the original error
- `errors.Is` traverses the chain (checks identity/sentinel errors)
- `errors.As` traverses the chain to extract typed errors
- Sentinel errors: package-level `var ErrXxx = errors.New("...")` for specific conditions
- Custom error types: when you need structured data with the error
- Error chain pattern: each layer adds context with `fmt.Errorf("layer: %w", err)`
- `errors.Join` (Go 1.20+) combines multiple errors
- Anti-patterns: ignoring errors, wrapping without context, panic for control flow
- The `must` pattern: for errors that represent programming bugs, not runtime errors
