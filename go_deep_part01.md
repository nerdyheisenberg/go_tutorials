# Go Deep Dive — Part 1: Philosophy, Setup & The Go Mental Model

---

## Chapter 1: Why Go Exists — The Philosophy

### The Problem Go Was Built to Solve

By 2007, Google faced a unique engineering problem. Their C++ codebases took **45 minutes to compile**. Python was too slow at runtime. Java required heavy tooling. The team — Robert Griesemer, Rob Pike, and Ken Thompson — asked: "What if we designed a language that kept the speed of C++ but was as productive as Python?"

Go's design decisions flow from three core tensions:
1. **Fast compilation** vs **Type safety** (Go chose both)
2. **Performance** vs **Garbage collection** (Go chose both with a low-latency GC)
3. **Simplicity** vs **Expressiveness** (Go aggressively chose simplicity)

### The Opinionated Choices

Every "missing" feature in Go is a deliberate choice: IMPORTANT

| Feature | Go's answer | Why |
|---|---|---|
| No `while` loop | Use `for` | One way to loop = less debate |
| No `try/catch` | Return errors as values | Errors are normal, not exceptions |
| No inheritance | Composition + Interfaces | Avoids fragile base class problem |
| No method overloading | Not allowed | Prevents confusing APIs |
| No default function args | Use option structs | Makes call sites explicit |
| No generics (before 1.18) | Use `interface{}` or code gen | Added later when patterns were clear |
| Mandatory braces `{}` | Not optional | Prevents Apple SSL bug class |
| Unused imports/vars = compile error | Keeps code clean | Prevents dead code accumulation |
| `gofmt` built in | One true style | No style debates |


Composition -->> struct inside a struct without a name
Inteface -->> interface inside a interface without a name does the same
Using all the functionality of previous struct or interface is composition
Code reuse and polymorphism(normally we cannot even allow this like 
type bla float64 , type b float64 ) , now we cannot do bla = b or b = bla, even though type is same, but in case of interface we can cause of polymorphism





### Go's Mental Model: Simplicity IS the Feature

```
"Complexity is multiplicative: fixing a problem by making
 the design more complex permanently adds to the complexity
 of the system." — Rob Pike
```

When learning Go from C++, the hardest thing to unlearn is the instinct to add complexity. In Go:
- If your design needs a complex class hierarchy → you're doing it wrong
- If you need runtime type inspection everywhere → your interfaces are too big
- If your error handling has layers of wrapping → review your package boundaries

---

## Chapter 2: Installation, Project Structure & The Module System

### Installation

```bash
# Linux (x86-64)
wget https://go.dev/dl/go1.22.2.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz

# Add to ~/.bashrc or ~/.zshrc
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go              # where go tools install to
export PATH=$PATH:$GOPATH/bin       # so installed tools are accessible

# Verify
go version    # go version go1.22.2 linux/amd64
go env        # shows all Go environment variables
```

### Key Environment Variables

```bash
go env GOPATH    # ~/go — where 'go install' puts binaries
go env GOROOT    # /usr/local/go — where the Go SDK lives
go env GOPROXY   # proxy.golang.org,direct — module proxy
go env GOMODCACHE # ~/.go/pkg/mod — cached downloaded modules
go env GONOSUMDB # modules that skip checksum verification
go env GOOS      # current OS (linux, darwin, windows)
go env GOARCH    # current arch (amd64, arm64, 386)
```

### The Module System (Deep Dive) IMPORTANT

A **module** is a collection of related Go packages versioned together. Modules replaced `GOPATH` as the primary project organization mechanism in Go 1.11.

```bash
# Create a new module
mkdir myproject && cd myproject
go mod init github.com/rohit/myproject

# This creates go.mod:
# module github.com/rohit/myproject
# go 1.22
```

**What `go.mod` does:**
- Declares the module's identity (its import path)
- Declares the minimum Go version required
- Lists direct dependencies and their minimum required versions
- Uses **Minimum Version Selection (MVS)** — deterministic, reproducible builds

```
module github.com/rohit/myproject

go 1.22

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/stretchr/testify v1.8.4
)

require (
    // indirect dependencies
    github.com/bytedance/sonic v1.9.1 // indirect
)
```

**`go.sum` — The Security File** 

```bash
# go.sum contains cryptographic hashes of every dependency version
# Do NOT add it to .gitignore — it MUST be committed
# It prevents supply chain attacks (downloading tampered modules)
```

### Essential Module Commands IMPORTANT

```bash
# Add a dependency
go get github.com/gin-gonic/gin@v1.9.1    # specific version
go get github.com/gin-gonic/gin@latest    # latest
go get github.com/gin-gonic/gin@main      # main branch (be careful!)

# Remove unused dependencies, add missing ones
go mod tidy

# Download all dependencies locally
go mod download

# Show dependency graph
go mod graph | head -20

# Check for dependency updates
go list -u -m all

# Replace a module (useful for local development)
# In go.mod:
replace github.com/mylib v1.0.0 => ../mylib

# Vendor dependencies (for offline builds or auditing)
go mod vendor
go build -mod=vendor ./...

# Verify downloaded modules match go.sum
go mod verify
```

### Project Structure — The Standard Layout

```
myproject/
├── go.mod                    # module definition
├── go.sum                    # dependency checksums (commit this!)
├── Makefile                  # build automation
├── README.md
│
├── cmd/                      # main packages (entry points)
│   ├── server/
│   │   └── main.go           # binary: server
│   └── migrate/
│       └── main.go           # binary: migrate
│
├── internal/                 # PRIVATE — cannot be imported outside this module
│   ├── auth/
│   │   ├── auth.go
│   │   └── auth_test.go
│   ├── handler/
│   └── repository/
│       ├── user_repo.go
│       └── user_repo_test.go
│
├── pkg/                      # PUBLIC — can be imported by other modules
│   └── validator/
│       ├── validator.go
│       └── validator_test.go
│
├── api/                      # API definitions (proto, OpenAPI)
│   └── v1/
│       └── user.proto
│
├── configs/                  # Configuration files
│   ├── config.yaml
│   └── config_test.yaml
│
├── scripts/                  # Build, test, deployment scripts
│   └── migrate.sh
│
└── testdata/                 # Test fixtures (special directory — not a package)
    └── sample.json
```

**The `internal` package rule (critical!):**  IMPORTANT

The `internal` directory is enforced by the Go compiler. Any package inside `internal/` can ONLY be imported by packages rooted at the parent of the `internal` directory.

```
myproject/internal/auth/auth.go

# Can be imported by:
# myproject/cmd/server
# myproject/internal/handler
# myproject (the root)

# CANNOT be imported by:
# github.com/otherproject/something  → COMPILE ERROR
```

### Your First Real Project

```bash
mkdir myapp && cd myapp
go mod init github.com/rohit/myapp

mkdir -p cmd/server internal/config internal/handler pkg/utils

# cmd/server/main.go — the entry point
cat > cmd/server/main.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "github.com/rohit/myapp/internal/config"
    "github.com/rohit/myapp/internal/handler"
)

func main() {
    cfg, err := config.Load("configs/config.yaml")
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }
    
    h := handler.New(cfg)
    fmt.Printf("Starting server on port %d\n", cfg.Port)
    log.Fatal(h.Start())
}
EOF
```

### go.work — Multi-Module Workspaces (Go 1.18+)

When working on multiple modules simultaneously (like you have in `my_project`):

```bash
# In /home/rohit/jira_prediction/my_project/go.work:
go 1.22.2

use (
    ./hello    # the hello module
    ./util     # the util module
)
```

```bash
# Commands
go work init ./hello ./util    # create go.work
go work use ./newmodule        # add a module
go work sync                   # sync work dependencies
```

**Why use workspaces?**
- Work on multiple interdependent modules without publishing
- Replace `replace` directives in go.mod (cleaner approach)
- The workspace file is NOT committed (add to .gitignore)

---

## Chapter 3: How Go Programs Execute

### Compilation Pipeline IMPORTANT

```
Source Files (.go)
    ↓
Lexer/Scanner        → tokens
    ↓
Parser               → Abstract Syntax Tree (AST)
    ↓
Type Checker         → verified AST + type info
    ↓
Intermediate Code    → SSA (Static Single Assignment)
    ↓
Code Generator       → machine code
    ↓
Linker               → single binary with runtime embedded
```

Go links the **entire runtime** into each binary. This is why: //IMPORTANT 
- Go binaries are larger (typically 5-15MB) than C++ binaries
- But they have **zero runtime dependencies** — copy the binary anywhere and run
- No need for `libstdc++.so`, no `LD_LIBRARY_PATH` issues

### Build Commands

```bash
# Run without building a file
go run main.go
go run cmd/server/main.go

# Build a binary
go build -o bin/server ./cmd/server/
go build ./...    # build everything (catches compile errors)

# Install to $GOPATH/bin
go install github.com/some/tool@latest

# Cross compilation (Go's killer feature!)
GOOS=linux   GOARCH=amd64   go build -o server-linux   ./cmd/server/
GOOS=darwin  GOARCH=arm64   go build -o server-mac     ./cmd/server/
GOOS=windows GOARCH=amd64   go build -o server.exe     ./cmd/server/

# Build with race detector (for testing/staging, not production) IMPORTANT
go build -race ./...
go run -race main.go

# Reduce binary size
go build -ldflags="-s -w" ./cmd/server/ IMPORTANT
# -s: strip symbol table
# -w: strip DWARF debug info

# Show what would be compiled (verbose) IMPORTANT
go build -v ./...

# List all packages
go list ./...
```

### Init Order (Critical for Interview)

```go
package main IMPORTANT

import "fmt"

// Order of initialization:
// 1. Package-level variables (in declaration order, dependencies first)
// 2. init() functions (all of them, in source file order)
// 3. main()

var (
    a = compute()   // runs first
    b = a * 2       // runs second (depends on a)
)

func compute() int {
    fmt.Println("computing a")
    return 10
}

func init() {
    fmt.Println("init 1 — a:", a, "b:", b)
}

func init() {
    fmt.Println("init 2 — running setup")
}

func main() {
    fmt.Println("main — a:", a, "b:", b)
}

// Output:
// computing a
// init 1 — a: 10 b: 20
// init 2 — running setup
// main — a: 10 b: 20
```

**Cross-package init order:** IMPORTANT

```
If package A imports package B:
1. B's package-level vars are initialized
2. B's init() functions run
3. A's package-level vars are initialized
4. A's init() functions run
5. main() runs
```

### Testing: Is Your Setup Working?

```go
// File: setup_test.go
package main

import (
    "os"
    "os/exec"
    "testing"
)

func TestGoInstalled(t *testing.T) {
    cmd := exec.Command("go", "version")
    if err := cmd.Run(); err != nil {
        t.Fatalf("Go not installed or not in PATH: %v", err)
    }
}

func TestModuleInitialized(t *testing.T) {
    if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
        t.Fatal("go.mod not found — run 'go mod init <module-name>'")
    }
}

func TestGoEnv(t *testing.T) {
    gopath := os.Getenv("GOPATH")
    if gopath == "" {
        // GOPATH has a default if not set — this isn't necessarily an error
        t.Log("GOPATH not explicitly set, using default")
    }
    t.Logf("GOPATH: %s", gopath)
}
```
test with go test -v 
---

## Chapter 4: Package Naming and Import System

### Package Naming Rules

```go
// Rules:
// 1. Package name = lowercase single word (no underscores, no camelCase)
// 2. Package name is the LAST element of the import path (by convention)
// 3. The directory name and package name should match

// Good:
package auth
package database
package httputil
package json

// Avoid:
package Auth          // no uppercase
package my_auth       // no underscores
package authenticator // too long (use 'auth')
```

### Import Paths

```go
// Standard library (no module path needed)
import "fmt"
import "net/http"
import "encoding/json"

// Third party (full module path)
import "github.com/gin-gonic/gin"
import "go.uber.org/zap"

// Your own packages
import "github.com/rohit/myapp/internal/config"
import "github.com/rohit/myapp/pkg/validator"

// Multiple imports (ALWAYS use grouped import)
import (
    // Standard library first
    "context"
    "fmt"
    "net/http"
    
    // Then third party (separated by blank line)
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    
    // Then your own packages
    "github.com/rohit/myapp/internal/config"
)
```

### Import Aliases

```go
import (
    "fmt"
    
    // Alias when package names conflict
    nethttp "net/http"
    myhttp "github.com/rohit/myapp/internal/http"
    
    // Blank import — import for SIDE EFFECTS only (runs init())
    _ "github.com/lib/pq"          // registers PostgreSQL driver
    _ "image/png"                   // registers PNG decoder
    
    // Dot import — brings all exported names into current scope
    // (AVOID: makes it unclear where names come from)
    . "fmt"    // now you can write Println instead of fmt.Println
)
```

**Why blank imports?**

```go
package main

import (
    "database/sql"
    _ "github.com/lib/pq" // Run pq's init() which calls sql.Register("postgres", ...)
)

func main() {
    // Without the blank import, this would fail at runtime
    // because the postgres driver was never registered
    db, err := sql.Open("postgres", "host=localhost user=postgres...")
}
```

### Testing Package Naming Conventions

```go
// Two conventions for test package names:

// 1. Same package (white-box testing — can access unexported names)
// File: auth/auth_test.go
package auth

func TestInternalHelper(t *testing.T) {
    // can call unexported functions like internalHelper()
}

// 2. External package (black-box testing — tests the public API only)
// File: auth/auth_test.go  
package auth_test   // note the _test suffix

import "github.com/rohit/myapp/internal/auth"

func TestPublicAPI(t *testing.T) {
    // can only use exported names from auth package
    auth.Login(...)
}
```

---

## Chapter 5: Practical Exercises for Part 1

### Exercise 1: Set Up a Real Module

```bash
mkdir -p ~/go-practice/exercise1
cd ~/go-practice/exercise1
go mod init github.com/practice/exercise1

cat > main.go << 'EOF'
package main

import (
    "fmt"
    "runtime"
)

func main() {
    fmt.Printf("Go Version : %s\n", runtime.Version())
    fmt.Printf("OS/Arch    : %s/%s\n", runtime.GOOS, runtime.GOARCH)
    fmt.Printf("CPU Cores  : %d\n", runtime.NumCPU())
    fmt.Printf("Goroutines : %d\n", runtime.NumGoroutine())
}
EOF

go run main.go
```

### Exercise 2: Explore Build Flags

```bash
# Build with version information embedded
VERSION="1.0.0"
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S')
IMPORTANT
go build -ldflags="\
  -X 'main.Version=${VERSION}' \
  -X 'main.Commit=${COMMIT}' \
  -X 'main.BuildTime=${BUILD_TIME}'" \
  -o myapp main.go

# In main.go:
var (
    Version   = "dev"
    Commit    = "unknown"
    BuildTime = "unknown"
)

func main() {
    fmt.Printf("Version %s (%s) built at %s\n", Version, Commit, BuildTime)
}
```

### Test File for Chapter Setup Verification

```go
// File: setup_test.go
package main

import (
    "os"
    "path/filepath"
    "runtime"
    "strings"
    "testing"
)

func TestGoVersion(t *testing.T) {
    version := runtime.Version()
    major := strings.TrimPrefix(version, "go")
    parts := strings.Split(major, ".")
    
    if len(parts) < 2 {
        t.Fatalf("unexpected version format: %s", version)
    }
    
    t.Logf("Running Go %s", version)
    
    // Ensure we're on Go 1.18+ for generics support
    if parts[0] < "1" || (parts[0] == "1" && parts[1] < "18") {
        t.Errorf("need Go 1.18+, got %s", version)
    }
}

func TestProjectStructure(t *testing.T) {
    required := []string{"go.mod"}
    for _, path := range required {
        if _, err := os.Stat(path); os.IsNotExist(err) {
            t.Errorf("required file/dir missing: %s", path)
        }
    }
}

func TestBuildConstraints(t *testing.T) {
    goos := runtime.GOOS
    goarch := runtime.GOARCH
    
    validOS := map[string]bool{
        "linux": true, "darwin": true, "windows": true,
    }
    validArch := map[string]bool{
        "amd64": true, "arm64": true, "386": true,
    }
    
    if !validOS[goos] {
        t.Logf("Unusual OS: %s (may require special build flags)", goos)
    }
    if !validArch[goarch] {
        t.Logf("Unusual arch: %s", goarch)
    }
}

func TestModFile(t *testing.T) {
    data, err := os.ReadFile("go.mod")
    if err != nil {
        t.Fatal("Cannot read go.mod:", err)
    }
    
    content := string(data)
    if !strings.HasPrefix(content, "module ") {
        t.Error("go.mod should start with 'module' declaration")
    }
    if !strings.Contains(content, "go ") {
        t.Error("go.mod should contain Go version declaration")
    }
    
    t.Log("go.mod contents:")
    t.Log(content)
    _ = filepath.Abs("go.mod") // just ensure filepath import is used
}
```

Run with: `go test -v ./...` IMPORTANT

---

**Summary of Part 1:**
- Go's simplicity is intentional — "missing" features are deliberate choices
- Every Go project is a module with a `go.mod` file
- `internal/` packages are compiler-enforced private code
- `go.sum` must be committed — it's your security file
- Cross-compilation is trivially easy in Go
- Init order: package vars → init() functions → main()
- Import blank `_` runs `init()` for side effects only
