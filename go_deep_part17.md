# Go Deep Dive — Part 17: HTTP Server & Client

---

## Chapter 1: The net/http Package — Architecture

Go's `net/http` package is production-ready out of the box. It's used by companies like Google, Netflix, and Uber in production without wrapping in another framework.

```
Request lifecycle:
  Client → TCP → http.Server.ListenAndServe()
                    → Reads HTTP request  
                    → Parses to *http.Request
                    → Routes to http.Handler
                       → Your code runs
                    → Writes http.ResponseWriter
  Response → TCP → Client
```

---

## Chapter 2: HTTP Server — All Patterns

### Basic Server

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
)

// Simplest possible server
func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello, World!")
    })
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Production Server with Proper Config

```go
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /users", listUsers)       // Go 1.22+: method in pattern
    mux.HandleFunc("POST /users", createUser)
    mux.HandleFunc("GET /users/{id}", getUser)    // Go 1.22+: path variables
    mux.HandleFunc("PUT /users/{id}", updateUser)
    mux.HandleFunc("DELETE /users/{id}", deleteUser)
    
    // Chain middleware
    handler := loggingMiddleware(rateLimitMiddleware(mux))
    
    server := &http.Server{
        Addr:         ":8080",
        Handler:      handler,
        ReadTimeout:  15 * time.Second,  // time to read the full request
        WriteTimeout: 15 * time.Second,  // time to write the full response
        IdleTimeout:  60 * time.Second,  // keep-alive timeout
        MaxHeaderBytes: 1 << 20,         // 1MB max headers
    }
    
    // Graceful shutdown
    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("ListenAndServe: %v", err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }
    log.Println("Server stopped")
}
```

### Handler — Complete Reference

```go
// http.ResponseWriter interface:
// Header() http.Header     — set response headers
// Write([]byte) (int, error) — write response body
// WriteHeader(int)          — set HTTP status code (must be before Write!)

func handler(w http.ResponseWriter, r *http.Request) {
    // Reading the request
    method := r.Method                  // "GET", "POST", etc.
    path := r.URL.Path                  // "/users/42"
    query := r.URL.Query().Get("name")  // ?name=Rohit
    rawQuery := r.URL.RawQuery          // "name=Rohit&page=1"
    
    // Headers
    auth := r.Header.Get("Authorization")
    contentType := r.Header.Get("Content-Type")
    
    // Body
    body, err := io.ReadAll(r.Body)     // read entire body
    defer r.Body.Close()                // always close!
    
    // Path variables (Go 1.22+)
    id := r.PathValue("id")            // from pattern /users/{id}
    
    // Context (with request lifecycle)
    ctx := r.Context()
    
    // Writing the response
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Request-ID", "abc123")
    w.WriteHeader(http.StatusOK)       // must be AFTER setting headers!
    w.Write([]byte(`{"status":"ok"}`))
    
    // Shortcut: all-in-one
    http.Error(w, "Bad Request", http.StatusBadRequest)
    
    // JSON response helper pattern
    writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// Reusable JSON response helper
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(v); err != nil {
        // Can't change status code now — it's already sent
        log.Printf("writeJSON: encode: %v", err)
    }
}

// Reading JSON request body
func readJSON(r *http.Request, v interface{}) error {
    defer r.Body.Close()
    decoder := json.NewDecoder(r.Body)
    decoder.DisallowUnknownFields()  // reject extra fields
    return decoder.Decode(v)
}
```

---

## Chapter 3: Middleware — The Request Pipeline

```go
// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Capture status code by wrapping ResponseWriter
        rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        next.ServeHTTP(rw, r)
        
        log.Printf(
            "method=%s path=%s status=%d duration=%v",
            r.Method, r.URL.Path, rw.statusCode, time.Since(start),
        )
    })
}

// ResponseWriter wrapper to capture status code
type responseWriter struct {
    http.ResponseWriter
    statusCode int
    written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
    if !rw.written {
        rw.statusCode = code
        rw.written = true
        rw.ResponseWriter.WriteHeader(code)
    }
}

// Auth middleware
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        user, err := validateToken(token)
        if err != nil {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        
        // Add user to context
        ctx := context.WithValue(r.Context(), userContextKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Recovery middleware (panic recovery)
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                buf := make([]byte, 64*1024)
                n := runtime.Stack(buf, false)
                log.Printf("PANIC: %v\n%s", err, buf[:n])
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// Chaining middleware (right-to-left order: first in the list = outermost)
func chain(mws ...Middleware) Middleware {
    return func(next http.Handler) http.Handler {
        for i := len(mws) - 1; i >= 0; i-- {
            next = mws[i](next)
        }
        return next
    }
}

// Usage:
middleware := chain(
    recoveryMiddleware,  // outermost: catches panics from all others
    loggingMiddleware,
    corsMiddleware,
    authMiddleware,      // innermost: runs just before handler
)
handler := middleware(mux)
```

---

## Chapter 4: HTTP Client — Complete Reference

```go
// Default client: uses http.DefaultClient (no timeouts — DANGEROUS!)
resp, err := http.Get("https://api.example.com/data")
// If server hangs, this blocks forever! Never use in production.

// Always create a client with timeouts
client := &http.Client{
    Timeout: 30 * time.Second,  // total request timeout
    Transport: &http.Transport{
        MaxIdleConns:        100,              // max idle connections in pool
        MaxIdleConnsPerHost: 10,               // max idle per host
        IdleConnTimeout:     90 * time.Second, // how long idle connections live
        TLSHandshakeTimeout: 10 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second, // wait for first byte of headers
        DisableCompression:  false,            // accept gzip
    },
}

// GET request
resp, err := client.Get("https://api.example.com/users/1")
if err != nil { return err }
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("unexpected status: %d", resp.StatusCode)
}

var user User
if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
    return err
}

// POST request with JSON body
body := &CreateUserRequest{Name: "Alice", Email: "alice@example.com"}
data, _ := json.Marshal(body)

resp2, err := client.Post(
    "https://api.example.com/users",
    "application/json",
    bytes.NewReader(data),
)
defer resp2.Body.Close()

// Builder pattern for requests (more control)
req, err := http.NewRequestWithContext(ctx, "POST", "https://api.example.com/users", bytes.NewReader(data))
if err != nil { return err }
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+token)
req.Header.Set("X-Request-ID", requestID)

resp3, err := client.Do(req)
if err != nil { return err }
defer resp3.Body.Close()
```

### Resilient HTTP Client

```go
type ResilientClient struct {
    client     *http.Client
    maxRetries int
    baseDelay  time.Duration
}

func (c *ResilientClient) Do(req *http.Request) (*http.Response, error) {
    var lastErr error
    
    for attempt := 0; attempt <= c.maxRetries; attempt++ {
        if attempt > 0 {
            // Exponential backoff with jitter
            delay := c.baseDelay * time.Duration(1<<uint(attempt-1))
            jitter := time.Duration(rand.Int63n(int64(delay / 2)))
            time.Sleep(delay + jitter)
        }
        
        // Clone request for retry (Body can only be read once)
        cloned, err := cloneRequest(req)
        if err != nil { return nil, err }
        
        resp, err := c.client.Do(cloned)
        if err != nil {
            lastErr = err
            continue
        }
        
        // Retry on server errors (5xx) but not client errors (4xx)
        if resp.StatusCode >= 500 {
            resp.Body.Close()
            lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
            continue
        }
        
        return resp, nil
    }
    
    return nil, fmt.Errorf("all %d attempts failed: %w", c.maxRetries+1, lastErr)
}

func cloneRequest(req *http.Request) (*http.Request, error) {
    clone := req.Clone(req.Context())
    if req.Body != nil {
        // Re-read the body — requires GetBody to be set
        if req.GetBody == nil {
            return nil, errors.New("cannot retry request: no GetBody function")
        }
        body, err := req.GetBody()
        if err != nil { return nil, err }
        clone.Body = body
    }
    return clone, nil
}
```

---

## Chapter 5: Testing HTTP Handlers

```go
package handlers_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

type UserHandler struct {
    store UserStore
}

func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
    var req struct { Name, Email string }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }
    
    if req.Name == "" || req.Email == "" {
        http.Error(w, "Name and email required", http.StatusUnprocessableEntity)
        return
    }
    
    user, err := h.store.Create(r.Context(), req.Name, req.Email)
    if err != nil {
        http.Error(w, "Internal Error", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func TestCreateUser(t *testing.T) {
    tests := []struct {
        name           string
        body           interface{}
        storeBehavior  func(*MockStore)
        wantStatus     int
        checkResponse  func(t *testing.T, resp *http.Response)
    }{
        {
            name: "valid creation",
            body: map[string]string{"name": "Alice", "email": "alice@test.com"},
            storeBehavior: func(s *MockStore) {
                s.CreateFunc = func(_ context.Context, name, email string) (*User, error) {
                    return &User{ID: 1, Name: name, Email: email}, nil
                }
            },
            wantStatus: http.StatusCreated,
            checkResponse: func(t *testing.T, resp *http.Response) {
                var user User
                json.NewDecoder(resp.Body).Decode(&user)
                if user.Name != "Alice" { t.Errorf("name = %q", user.Name) }
            },
        },
        {
            name:       "missing name",
            body:       map[string]string{"email": "alice@test.com"},
            wantStatus: http.StatusUnprocessableEntity,
        },
        {
            name:       "invalid JSON",
            body:       "not JSON",
            wantStatus: http.StatusBadRequest,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            store := &MockStore{}
            if tt.storeBehavior != nil { tt.storeBehavior(store) }
            handler := &UserHandler{store: store}
            
            var bodyReader bytes.Reader
            switch v := tt.body.(type) {
            case string:
                bodyReader = *bytes.NewReader([]byte(v))
            default:
                data, _ := json.Marshal(v)
                bodyReader = *bytes.NewReader(data)
            }
            
            req := httptest.NewRequest("POST", "/users", &bodyReader)
            req.Header.Set("Content-Type", "application/json")
            w := httptest.NewRecorder()
            
            handler.createUser(w, req)
            resp := w.Result()
            
            if resp.StatusCode != tt.wantStatus {
                t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
            }
            if tt.checkResponse != nil {
                tt.checkResponse(t, resp)
            }
        })
    }
}

// Test middleware
func TestLoggingMiddleware(t *testing.T) {
    var logged []string
    
    testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    
    // Custom logging middleware for test
    loggingMW := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            rw := &responseWriter{ResponseWriter: w, statusCode: 200}
            next.ServeHTTP(rw, r)
            logged = append(logged, fmt.Sprintf("%s %s %d", r.Method, r.URL.Path, rw.statusCode))
        })
    }
    
    handler := loggingMW(testHandler)
    req := httptest.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    
    if len(logged) != 1 { t.Errorf("expected 1 log entry, got %d", len(logged)) }
    if logged[0] != "GET /test 200" {
        t.Errorf("log = %q, want 'GET /test 200'", logged[0])
    }
}
```

---

## Chapter 6: Building a Complete REST API

```go
package main

import (
    "context"
    "encoding/json"
    "errors"
    "log"
    "net/http"
    "strconv"
    "time"
)

// Domain types
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

// In-memory store (replace with real DB in production)
type InMemoryUserStore struct {
    users  map[int]*User
    nextID int
}

func NewInMemoryUserStore() *InMemoryUserStore {
    return &InMemoryUserStore{users: make(map[int]*User), nextID: 1}
}

func (s *InMemoryUserStore) Get(id int) (*User, error) {
    u, ok := s.users[id]
    if !ok { return nil, ErrNotFound }
    return u, nil
}

func (s *InMemoryUserStore) Create(name, email string) (*User, error) {
    u := &User{ID: s.nextID, Name: name, Email: email, CreatedAt: time.Now()}
    s.users[s.nextID] = u
    s.nextID++
    return u, nil
}

// Handler struct
type UserAPI struct {
    store *InMemoryUserStore
}

func (api *UserAPI) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("GET /api/users/{id}", api.getUser)
    mux.HandleFunc("POST /api/users", api.createUser)
}

func (api *UserAPI) getUser(w http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid user ID")
        return
    }
    
    user, err := api.store.Get(id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            writeError(w, http.StatusNotFound, "user not found")
            return
        }
        writeError(w, http.StatusInternalServerError, "internal error")
        return
    }
    
    writeJSON(w, http.StatusOK, user)
}

func (api *UserAPI) createUser(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid JSON")
        return
    }
    defer r.Body.Close()
    
    if req.Name == "" || req.Email == "" {
        writeError(w, http.StatusUnprocessableEntity, "name and email are required")
        return
    }
    
    user, err := api.store.Create(req.Name, req.Email)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "failed to create user")
        return
    }
    
    writeJSON(w, http.StatusCreated, user)
}

// Helpers
type ErrorResponse struct {
    Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
    writeJSON(w, status, ErrorResponse{Error: msg})
}

var ErrNotFound = errors.New("not found")

func main() {
    store := NewInMemoryUserStore()
    api := &UserAPI{store: store}
    
    mux := http.NewServeMux()
    api.RegisterRoutes(mux)
    
    server := &http.Server{
        Addr:         ":8080",
        Handler:      mux,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
    }
    
    log.Println("Server starting on :8080")
    log.Fatal(server.ListenAndServe())
}
```

---

**Summary of Part 17:**
- Always create `*http.Server` with timeouts — never use `http.ListenAndServe` in production
- Go 1.22 added method routing (`GET /users/{id}`) and `r.PathValue()` to stdlib
- Middleware: wrap `http.Handler` with `http.HandlerFunc` — chain right-to-left
- Always wrap the response writer to capture status codes for logging
- `http.Client` must have `Timeout` set — never use `http.DefaultClient` in production
- `Transport` settings control connection pooling — tune for production workloads
- `httptest.NewRecorder()` for unit-testing handlers without a real server
- `httptest.NewServer()` for integration tests with a real (in-process) server
- Graceful shutdown: `server.Shutdown(ctx)` after catching `SIGINT`/`SIGTERM`
- Pattern: struct-based handlers with `RegisterRoutes` method for clean organization
