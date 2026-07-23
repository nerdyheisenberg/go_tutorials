# Go Deep Dive — Part 19: Advanced Patterns & Idiomatic Go

---

## Chapter 1: The Repository Pattern

```go
// Classic layered architecture in Go:
// Handler → Service → Repository → Database

// Domain types (business entities)
type User struct {
    ID        int
    Name      string
    Email     string
    CreatedAt time.Time
}

// Repository interface (defined in domain layer — NOT in DB layer)
type UserRepository interface {
    GetByID(ctx context.Context, id int) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Create(ctx context.Context, u *User) error
    Update(ctx context.Context, u *User) error
    Delete(ctx context.Context, id int) error
    List(ctx context.Context, limit, offset int) ([]*User, error)
}

// PostgreSQL implementation
type postgresUserRepo struct {
    db *sql.DB
}

func NewPostgresUserRepo(db *sql.DB) UserRepository {
    return &postgresUserRepo{db: db}
}

func (r *postgresUserRepo) GetByID(ctx context.Context, id int) (*User, error) {
    const q = `SELECT id, name, email, created_at FROM users WHERE id = $1`
    u := &User{}
    err := r.db.QueryRowContext(ctx, q, id).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("GetByID(%d): %w", id, err)
    }
    return u, nil
}

func (r *postgresUserRepo) Create(ctx context.Context, u *User) error {
    const q = `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, created_at`
    return r.db.QueryRowContext(ctx, q, u.Name, u.Email).Scan(&u.ID, &u.CreatedAt)
}

// Service layer
type UserService struct {
    repo  UserRepository
    email EmailSender
}

func NewUserService(repo UserRepository, email EmailSender) *UserService {
    return &UserService{repo: repo, email: email}
}

func (s *UserService) Register(ctx context.Context, name, email string) (*User, error) {
    // Business logic validation
    if name == "" { return nil, ValidationError{Field: "name", Msg: "required"} }
    if !isValidEmail(email) { return nil, ValidationError{Field: "email", Msg: "invalid"} }
    
    // Check uniqueness
    existing, err := s.repo.GetByEmail(ctx, email)
    if err != nil && !errors.Is(err, ErrNotFound) {
        return nil, fmt.Errorf("Register: check email: %w", err)
    }
    if existing != nil {
        return nil, ErrConflict
    }
    
    u := &User{Name: name, Email: email}
    if err := s.repo.Create(ctx, u); err != nil {
        return nil, fmt.Errorf("Register: create: %w", err)
    }
    
    // Send welcome email (non-blocking)
    go func() {
        if err := s.email.Send(email, "Welcome!", "Hello "+name); err != nil {
            log.Printf("Register: welcome email to %s: %v", email, err)
        }
    }()
    
    return u, nil
}
```

---

## Chapter 2: The Options Pattern — Extended

```go
// The canonical Go pattern for configurable objects
// Used by standard library and major frameworks

type DatabaseConfig struct {
    host            string
    port            int
    name            string
    user            string
    password        string
    maxOpenConns    int
    maxIdleConns    int
    connMaxLifetime time.Duration
    sslMode         string
}

type DBOption func(*DatabaseConfig) error

// Options that can fail (return error)
func WithHost(host string) DBOption {
    return func(c *DatabaseConfig) error {
        if host == "" {
            return errors.New("host cannot be empty")
        }
        c.host = host
        return nil
    }
}

func WithPort(port int) DBOption {
    return func(c *DatabaseConfig) error {
        if port < 1 || port > 65535 {
            return fmt.Errorf("invalid port: %d", port)
        }
        c.port = port
        return nil
    }
}

func WithMaxConns(max, idle int) DBOption {
    return func(c *DatabaseConfig) error {
        if max < idle {
            return fmt.Errorf("max open conns (%d) must be >= max idle (%d)", max, idle)
        }
        c.maxOpenConns = max
        c.maxIdleConns = idle
        return nil
    }
}

func WithSSLMode(mode string) DBOption {
    valid := map[string]bool{"disable": true, "require": true, "verify-full": true}
    return func(c *DatabaseConfig) error {
        if !valid[mode] {
            return fmt.Errorf("invalid ssl mode: %s", mode)
        }
        c.sslMode = mode
        return nil
    }
}

func Connect(opts ...DBOption) (*sql.DB, error) {
    cfg := &DatabaseConfig{  // defaults
        host:            "localhost",
        port:            5432,
        maxOpenConns:    25,
        maxIdleConns:    5,
        connMaxLifetime: 5 * time.Minute,
        sslMode:         "disable",
    }
    
    for _, opt := range opts {
        if err := opt(cfg); err != nil {
            return nil, fmt.Errorf("Connect: invalid option: %w", err)
        }
    }
    
    dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
        cfg.host, cfg.port, cfg.name, cfg.user, cfg.password, cfg.sslMode)
    
    db, err := sql.Open("postgres", dsn)
    if err != nil { return nil, err }
    
    db.SetMaxOpenConns(cfg.maxOpenConns)
    db.SetMaxIdleConns(cfg.maxIdleConns)
    db.SetConnMaxLifetime(cfg.connMaxLifetime)
    
    return db, nil
}

// Usage:
db, err := Connect(
    WithHost("db.example.com"),
    WithPort(5432),
    WithMaxConns(50, 10),
    WithSSLMode("require"),
)
```

---

## Chapter 3: The Worker Pool Pattern

```go
// Classic pattern: fixed number of goroutines processing a job queue

type Job struct {
    ID   int
    Data interface{}
}

type Result struct {
    JobID int
    Value interface{}
    Err   error
}

type WorkerPool struct {
    workers int
    jobs    chan Job
    results chan Result
    done    chan struct{}
    wg      sync.WaitGroup
}

func NewWorkerPool(workers, queueSize int) *WorkerPool {
    pool := &WorkerPool{
        workers: workers,
        jobs:    make(chan Job, queueSize),
        results: make(chan Result, queueSize),
        done:    make(chan struct{}),
    }
    pool.start()
    return pool
}

func (p *WorkerPool) start() {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go func(workerID int) {
            defer p.wg.Done()
            for {
                select {
                case job, ok := <-p.jobs:
                    if !ok { return }
                    result := p.processJob(job)
                    p.results <- result
                case <-p.done:
                    return
                }
            }
        }(i)
    }
}

func (p *WorkerPool) processJob(job Job) Result {
    // In real code, this would do actual work
    // This is just a placeholder
    return Result{JobID: job.ID, Value: job.Data}
}

func (p *WorkerPool) Submit(job Job) {
    p.jobs <- job
}

func (p *WorkerPool) Results() <-chan Result {
    return p.results
}

func (p *WorkerPool) Shutdown() {
    close(p.jobs)       // signal: no more jobs
    p.wg.Wait()         // wait for all workers
    close(p.results)    // signal: no more results
}
```

---

## Chapter 4: Event-Driven / Pub-Sub Pattern

```go
// Simple in-process event bus
type Event struct {
    Type    string
    Payload interface{}
}

type Handler func(Event)

type EventBus struct {
    mu       sync.RWMutex
    handlers map[string][]Handler
}

func NewEventBus() *EventBus {
    return &EventBus{handlers: make(map[string][]Handler)}
}

func (b *EventBus) Subscribe(eventType string, handler Handler) func() {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    b.handlers[eventType] = append(b.handlers[eventType], handler)
    
    // Return unsubscribe function
    idx := len(b.handlers[eventType]) - 1
    return func() {
        b.mu.Lock()
        defer b.mu.Unlock()
        handlers := b.handlers[eventType]
        b.handlers[eventType] = append(handlers[:idx], handlers[idx+1:]...)
    }
}

func (b *EventBus) Publish(ctx context.Context, event Event) {
    b.mu.RLock()
    handlers := make([]Handler, len(b.handlers[event.Type]))
    copy(handlers, b.handlers[event.Type])
    b.mu.RUnlock()
    
    for _, h := range handlers {
        h := h  // capture
        go func() {
            defer func() {
                if r := recover(); r != nil {
                    log.Printf("EventBus: handler panicked for %s: %v", event.Type, r)
                }
            }()
            h(event)
        }()
    }
}

// Usage:
bus := NewEventBus()

unsubscribe := bus.Subscribe("user.created", func(e Event) {
    user := e.Payload.(*User)
    sendWelcomeEmail(user)
})
defer unsubscribe()

bus.Subscribe("user.created", func(e Event) {
    user := e.Payload.(*User)
    createUserProfile(user)
})

// When a user is created:
bus.Publish(ctx, Event{Type: "user.created", Payload: newUser})
```

---

## Chapter 5: Circuit Breaker Pattern

```go
// Prevent cascading failures when a downstream service is down

type CircuitState int

const (
    StateClosed   CircuitState = iota // Normal: requests flow through
    StateOpen                          // Failing: requests fail immediately
    StateHalfOpen                      // Testing: allow one request through
)

type CircuitBreaker struct {
    mu            sync.Mutex
    state         CircuitState
    failures      int
    successCount  int
    lastFailure   time.Time
    maxFailures   int
    resetTimeout  time.Duration
    halfOpenLimit int
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        maxFailures:   maxFailures,
        resetTimeout:  resetTimeout,
        halfOpenLimit: 1,
    }
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    if err := cb.beforeRequest(); err != nil { return err }
    
    err := fn()
    cb.afterRequest(err)
    return err
}

func (cb *CircuitBreaker) beforeRequest() error {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    switch cb.state {
    case StateClosed:
        return nil  // proceed
    case StateOpen:
        if time.Since(cb.lastFailure) > cb.resetTimeout {
            cb.state = StateHalfOpen
            cb.successCount = 0
            return nil  // allow one through
        }
        return errors.New("circuit breaker open")
    case StateHalfOpen:
        if cb.successCount < cb.halfOpenLimit {
            return nil
        }
        return errors.New("circuit breaker half-open: limit reached")
    }
    return nil
}

func (cb *CircuitBreaker) afterRequest(err error) {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        if cb.failures >= cb.maxFailures {
            cb.state = StateOpen
        }
    } else {
        if cb.state == StateHalfOpen {
            cb.successCount++
            if cb.successCount >= cb.halfOpenLimit {
                cb.state = StateClosed
                cb.failures = 0
            }
        } else {
            cb.failures = 0
        }
    }
}
```

---

## Chapter 6: Builder Pattern with Validation

```go
// Useful for complex objects with validation requirements

type QueryBuilder struct {
    table      string
    conditions []string
    orderBy    string
    limit      int
    offset     int
    joins      []string
    errs       []error
}

func Query(table string) *QueryBuilder {
    return &QueryBuilder{table: table}
}

func (q *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
    if condition == "" {
        q.errs = append(q.errs, errors.New("condition cannot be empty"))
        return q
    }
    q.conditions = append(q.conditions, fmt.Sprintf(condition, args...))
    return q
}

func (q *QueryBuilder) OrderBy(field, direction string) *QueryBuilder {
    if direction != "ASC" && direction != "DESC" {
        q.errs = append(q.errs, fmt.Errorf("invalid direction %q: must be ASC or DESC", direction))
        return q
    }
    q.orderBy = field + " " + direction
    return q
}

func (q *QueryBuilder) Limit(n int) *QueryBuilder {
    if n <= 0 {
        q.errs = append(q.errs, fmt.Errorf("limit must be positive: %d", n))
        return q
    }
    q.limit = n
    return q
}

func (q *QueryBuilder) Offset(n int) *QueryBuilder {
    if n < 0 {
        q.errs = append(q.errs, fmt.Errorf("offset cannot be negative: %d", n))
        return q
    }
    q.offset = n
    return q
}

func (q *QueryBuilder) Build() (string, error) {
    if len(q.errs) > 0 {
        return "", errors.Join(q.errs...)
    }
    
    sql := "SELECT * FROM " + q.table
    if len(q.conditions) > 0 {
        sql += " WHERE " + strings.Join(q.conditions, " AND ")
    }
    if q.orderBy != "" { sql += " ORDER BY " + q.orderBy }
    if q.limit > 0 { sql += fmt.Sprintf(" LIMIT %d", q.limit) }
    if q.offset > 0 { sql += fmt.Sprintf(" OFFSET %d", q.offset) }
    
    return sql, nil
}

// Usage:
query, err := Query("users").
    Where("age > %d", 18).
    Where("active = true").
    OrderBy("name", "ASC").
    Limit(10).
    Offset(0).
    Build()
```

---

## Chapter 7: Middleware Chain with Context Values

```go
// Advanced middleware: uses context to pass data through the chain

type contextKey string

const (
    userKey      contextKey = "user"
    requestIDKey contextKey = "requestID"
    traceKey     contextKey = "trace"
)

// Rich context helpers
func WithUser(ctx context.Context, user *User) context.Context {
    return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context) (*User, bool) {
    u, ok := ctx.Value(userKey).(*User)
    return u, ok
}

// Structured middleware chain
type Chain struct {
    middlewares []func(http.Handler) http.Handler
}

func NewChain(middlewares ...func(http.Handler) http.Handler) Chain {
    return Chain{middlewares: middlewares}
}

func (c Chain) Then(h http.Handler) http.Handler {
    for i := len(c.middlewares) - 1; i >= 0; i-- {
        h = c.middlewares[i](h)
    }
    return h
}

func (c Chain) ThenFunc(fn http.HandlerFunc) http.Handler {
    return c.Then(fn)
}

// Usage:
chain := NewChain(
    requestIDMiddleware,
    loggingMiddleware,
    authMiddleware,
)
mux.Handle("/api/", chain.Then(apiHandler))
```

---

## Chapter 8: The Functional Option Pattern for Middleware

```go
// Configurable middleware using functional options

type LoggingConfig struct {
    IncludeBody    bool
    IncludeHeaders bool
    SlowThreshold  time.Duration
    Logger         *slog.Logger
}

type LoggingOption func(*LoggingConfig)

func WithBodyLogging() LoggingOption {
    return func(c *LoggingConfig) { c.IncludeBody = true }
}

func WithSlowThreshold(d time.Duration) LoggingOption {
    return func(c *LoggingConfig) { c.SlowThreshold = d }
}

func Logger(opts ...LoggingOption) func(http.Handler) http.Handler {
    cfg := &LoggingConfig{
        SlowThreshold: 1 * time.Second,
        Logger:        slog.Default(),
    }
    for _, o := range opts { o(cfg) }
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            rw := &responseWriter{ResponseWriter: w, statusCode: 200}
            
            next.ServeHTTP(rw, r)
            
            elapsed := time.Since(start)
            level := slog.LevelInfo
            if rw.statusCode >= 500 { level = slog.LevelError }
            if elapsed > cfg.SlowThreshold { level = slog.LevelWarn }
            
            cfg.Logger.Log(r.Context(), level, "request",
                "method", r.Method,
                "path", r.URL.Path,
                "status", rw.statusCode,
                "duration", elapsed,
            )
        })
    }
}

// Usage:
handler := Logger(
    WithBodyLogging(),
    WithSlowThreshold(500*time.Millisecond),
)(mux)
```

---

## Chapter 9: Idiomatic Go — Do's and Don'ts

### Do's

```go
// ✅ Return errors, don't panic for expected failure cases
func getUser(id int) (*User, error) { ... }

// ✅ Use multiple return values for results + errors
func divide(a, b float64) (float64, error) { ... }

// ✅ Accept interfaces, return concrete types
func processData(r io.Reader) *ProcessedData { ... }

// ✅ Use defer for all cleanup
func openAndProcess(path string) error {
    f, err := os.Open(path)
    if err != nil { return err }
    defer f.Close()
    ...
}

// ✅ Use context.Context as first parameter
func doWork(ctx context.Context, data []byte) error { ... }

// ✅ Short variable names in small scopes
for i, v := range items { ... }

// ✅ Descriptive names in larger scopes
maxRetries := 3
userRepository := NewPostgresUserRepo(db)

// ✅ Error wrapping with %w
return fmt.Errorf("processOrder: %w", err)

// ✅ Table-driven tests
tests := []struct{ in int; want int }{ ... }
for _, tt := range tests { t.Run(...) }

// ✅ Use make for slices/maps with known capacity
results := make([]int, 0, len(input))
cache := make(map[string]int, expectedSize)

// ✅ Consistent receiver names
func (u *User) Name() string { ... }   // u, NOT user, NOT self, NOT this
func (u *User) Email() string { ... }  // same receiver name across all methods!
```

### Don'ts

```go
// ❌ Don't panic for expected failures
func getUser(id int) *User {
    if id == 0 { panic("invalid id") }  // bad!
    ...
}

// ❌ Don't use global mutable state
var globalDB *sql.DB  // avoid — makes testing hard, causes races

// ❌ Don't store context in structs
type Service struct {
    ctx context.Context  // NO! Context is per-request
}

// ❌ Don't return *interface{} or *error
func process() *error { ... }   // always wrong!
func get() *fmt.Stringer { ... } // almost always wrong

// ❌ Don't use different receiver names
func (u *User) DoX() {}
func (user *User) DoY() {}  // inconsistent! Pick one

// ❌ Don't use named returns in long complex functions
func complex() (result int, data []byte, err error) {
    // ... 100 lines ...
    return  // reader can't tell what's being returned!
}

// ❌ Don't use init() for non-trivial initialization
func init() {
    populateDatabase()  // too much in init! Errors can't be returned
}

// ❌ Don't ignore errors
json.Unmarshal(data, &result)  // silently corrupts result!
os.Remove("temp.txt")          // might not actually delete!

// ❌ Don't re-wrap errors without adding context
if err != nil {
    return fmt.Errorf("error: %w", err)  // "error:" adds nothing
}

// ❌ Don't use goroutines without a plan for completion
go func() { doSomething() }()  // will it finish? can it be cancelled?
```

---

## Chapter 10: Testing Advanced Patterns

```go
package patterns_test

import (
    "context"
    "errors"
    "testing"
    "time"
)

// ==================== CIRCUIT BREAKER TESTS ====================

func TestCircuitBreaker(t *testing.T) {
    cb := NewCircuitBreaker(3, 100*time.Millisecond)
    
    t.Run("starts closed (allows requests)", func(t *testing.T) {
        err := cb.Execute(func() error { return nil })
        if err != nil { t.Errorf("closed state: expected nil, got %v", err) }
    })
    
    t.Run("opens after max failures", func(t *testing.T) {
        failErr := errors.New("service down")
        
        for i := 0; i < 3; i++ {
            cb.Execute(func() error { return failErr })
        }
        
        // Now circuit is open — should fail fast
        err := cb.Execute(func() error { return nil })
        if err == nil { t.Error("open circuit should reject requests") }
    })
    
    t.Run("transitions to half-open after timeout", func(t *testing.T) {
        time.Sleep(150 * time.Millisecond)  // wait for reset timeout
        
        var called bool
        err := cb.Execute(func() error {
            called = true
            return nil  // success
        })
        if err != nil { t.Errorf("half-open should allow one request: %v", err) }
        if !called { t.Error("function should have been called") }
    })
    
    t.Run("closes after successful half-open", func(t *testing.T) {
        // Should now be closed (after successful half-open)
        err := cb.Execute(func() error { return nil })
        if err != nil { t.Errorf("should be closed: %v", err) }
    })
}

// ==================== EVENT BUS TESTS ====================

func TestEventBus(t *testing.T) {
    bus := NewEventBus()
    
    var received []Event
    var mu sync.Mutex
    
    unsubscribe := bus.Subscribe("test.event", func(e Event) {
        mu.Lock()
        received = append(received, e)
        mu.Unlock()
    })
    
    // Publish an event
    bus.Publish(context.Background(), Event{Type: "test.event", Payload: "hello"})
    
    // Wait for async delivery
    time.Sleep(50 * time.Millisecond)
    
    mu.Lock()
    count := len(received)
    mu.Unlock()
    
    if count != 1 { t.Errorf("expected 1 event, got %d", count) }
    
    // Unsubscribe and verify no more events received
    unsubscribe()
    bus.Publish(context.Background(), Event{Type: "test.event", Payload: "world"})
    time.Sleep(50 * time.Millisecond)
    
    mu.Lock()
    count2 := len(received)
    mu.Unlock()
    
    if count2 != 1 { t.Errorf("after unsubscribe: expected still 1 event, got %d", count2) }
}

// ==================== BUILDER PATTERN TESTS ====================

func TestQueryBuilder(t *testing.T) {
    tests := []struct {
        name    string
        builder func() (string, error)
        want    string
        wantErr bool
    }{
        {
            "simple query",
            func() (string, error) {
                return Query("users").Build()
            },
            "SELECT * FROM users",
            false,
        },
        {
            "with where clause",
            func() (string, error) {
                return Query("users").Where("active = true").Build()
            },
            "SELECT * FROM users WHERE active = true",
            false,
        },
        {
            "full query",
            func() (string, error) {
                return Query("users").
                    Where("age > 18").
                    OrderBy("name", "ASC").
                    Limit(10).
                    Offset(20).
                    Build()
            },
            "SELECT * FROM users WHERE age > 18 ORDER BY name ASC LIMIT 10 OFFSET 20",
            false,
        },
        {
            "invalid order direction",
            func() (string, error) {
                return Query("users").OrderBy("name", "RANDOM").Build()
            },
            "",
            true,
        },
        {
            "invalid limit",
            func() (string, error) {
                return Query("users").Limit(-1).Build()
            },
            "",
            true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := tt.builder()
            
            if (err != nil) != tt.wantErr {
                t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if !tt.wantErr && got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}

// ==================== WORKER POOL TESTS ====================

func TestWorkerPool(t *testing.T) {
    // Create a pool with 3 workers
    pool := NewWorkerPool(3, 100)
    
    // Submit 10 jobs
    for i := 0; i < 10; i++ {
        pool.Submit(Job{ID: i, Data: i * 2})
    }
    
    // Collect results
    var results []Result
    go func() {
        for r := range pool.Results() {
            results = append(results, r)
        }
    }()
    
    pool.Shutdown()
    time.Sleep(100 * time.Millisecond)  // let collector finish
    
    if len(results) != 10 {
        t.Errorf("expected 10 results, got %d", len(results))
    }
}
```

---

**Summary of Part 19:**
- Repository pattern: interface in domain layer, implementations in infrastructure — testable by design
- Functional options with error returns: validate during option application, not in constructor
- Worker pool: fixed goroutine count, job channel as work queue, shutdown via close+Wait
- Circuit breaker: three-state state machine prevents cascading failures
- Event bus: fan-out with unsubscribe functions returned; recover panics in async handlers
- Builder pattern: accumulate errors, build at the end — fluent API is idiomatic in Go
- Middleware: consistent `func(http.Handler) http.Handler` signature enables composability
- Do: accept interfaces, return concretes; error wrapping with context; defer for cleanup
- Don't: store context in structs; panic for expected failures; ignore errors; global state
- Always design for testability: interfaces at every boundary, functional options, mock-friendly
