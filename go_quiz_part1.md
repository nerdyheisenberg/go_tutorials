# Go Quiz — Part 1 of 5 (Q1–Q100)
### Covers: Parts 1–4 (Philosophy, Variables & Types, Control Flow, Functions)
> Format: Answer is hidden after `> ✅`. Try to answer before looking!

---

## PART 1 — Philosophy, Setup, Execution (Q1–Q25)

**Q1.** What command initializes a new Go module named `github.com/user/app`?
- a) `go new github.com/user/app`
- b) `go mod init github.com/user/app`
- c) `go init github.com/user/app`
- d) `go create github.com/user/app`

> ✅ **b** — `go mod init <module-path>` creates `go.mod` with the given module path.

---

**Q2.** Which of the following is NOT true about Go?
- a) Go is statically typed
- b) Go supports generics (since 1.18)
- c) Go supports function overloading
- d) Go has a garbage collector

> ✅ **c** — Go does NOT support function overloading. You cannot have two functions with the same name but different parameter types.

---

**Q3.** What does `GOMAXPROCS` control?
- a) Maximum memory the Go runtime can use
- b) Maximum number of goroutines
- c) Number of OS threads that can execute Go code simultaneously
- d) Number of packages compiled in parallel

> ✅ **c** — GOMAXPROCS = number of Processors (P's) in the GMP model = max OS threads running Go code in parallel. Default = number of CPU cores.

---

**Q4.** True or False: A Go binary statically includes the Go runtime — it has zero external runtime dependencies.

> ✅ **True** — Unlike C++ that needs `libstdc++.so`, a Go binary embeds the runtime and runs anywhere without installing anything.

---

**Q5.** What is the purpose of `go.sum`?
- a) Lists all installed Go versions
- b) Records cryptographic hashes of dependencies to prevent tampering
- c) Summarizes the module's public API
- d) Replaces `go.mod` in new projects

> ✅ **b** — `go.sum` contains hash of each module version's file tree, preventing supply-chain attacks.

---

**Q6.** What does `go mod tidy` do?
- a) Formats all Go files in the module
- b) Updates all dependencies to their latest versions
- c) Adds missing dependencies and removes unused ones from `go.mod`/`go.sum`
- d) Verifies that all tests pass

> ✅ **c** — `go mod tidy` is idempotent; it syncs `go.mod` and `go.sum` with actual imports.

---

**Q7.** True or False: The `internal` package can be imported by any package in any module.

> ✅ **False** — Only packages rooted at the **parent** of the `internal/` directory can import it. The Go compiler enforces this.

---

**Q8.** What is the correct format for a valid Go module path?
- a) `myapp`
- b) `github.com/user/myapp`
- c) `user/myapp`
- d) All of the above are valid

> ✅ **d** — All are syntactically valid module paths. Simple names work locally; domain-prefixed names are conventional for published modules.

---

**Q9.** What happens to unused imports in Go?
- a) A warning is printed but compilation succeeds
- b) The import is automatically removed
- c) Compilation fails with an error
- d) Nothing — they are silently ignored

> ✅ **c** — Unused imports are a **compile error** in Go. This enforces clean code.

---

**Q10.** What is the initialization order in a Go program?
- a) `main()` → `init()` → package-level variables
- b) `init()` → package-level variables → `main()`
- c) package-level variables → `init()` → `main()`
- d) package-level variables → `main()` → `init()`

> ✅ **c** — Package-level vars are initialized first (in dependency order), then all `init()` functions, then `main()`.

---

**Q11.** How many `init()` functions can a single Go **file** have?
- a) Exactly one
- b) Zero or one
- c) Unlimited — all will execute
- d) Up to 10

> ✅ **c** — A single file can have multiple `init()` functions. They all run in order of declaration.

---

**Q12.** True or False: Go supports circular imports between packages.

> ✅ **False** — The Go compiler rejects circular imports at compile time.

---

**Q13.** What does `go build -race ./...` do?
- a) Runs tests with the race detector
- b) Builds the binary with race-detection code instrumented
- c) Checks for races statically without building
- d) Fails if any data races are found in the source

> ✅ **b** — `-race` instruments the binary with race detection. The actual races are found at **runtime** when the binary runs.

---

**Q14.** What does `-ldflags="-s -w"` achieve when building?
- a) Enables strict linking and warnings
- b) Strips the symbol table (`-s`) and DWARF debug info (`-w`), reducing binary size
- c) Links all shared libraries statically
- d) Enables link-time optimization

> ✅ **b** — Common in production builds to reduce binary size by ~30%.

---

**Q15.** Which command shows escape analysis decisions during compilation?
- a) `go build -v ./...`
- b) `go build -gcflags="-m" ./...`
- c) `go build -escape ./...`
- d) `go vet ./...`

> ✅ **b** — `-gcflags="-m"` prints each variable's allocation decision (stack vs. heap).

---

**Q16.** True or False: `go run main.go` creates a binary in the current directory.

> ✅ **False** — `go run` compiles to a temp directory and runs; no binary is left in the current directory.

---

**Q17.** Cross-compilation in Go requires:
- a) Installing a cross-compiler toolchain (like with C++)
- b) Only setting `GOOS` and `GOARCH` environment variables and running `go build`
- c) A Docker container with the target OS
- d) CGO being enabled

> ✅ **b** — `GOOS=windows GOARCH=amd64 go build ./...` produces a Windows binary from Linux with no extra tools.

---

**Q18.** What does `go vet ./...` do?
- a) Verifies test coverage
- b) Formats Go code
- c) Reports suspicious constructs that compile but are likely bugs
- d) Checks module dependency versions

> ✅ **c** — `go vet` catches things like wrong `Printf` format args, unreachable code, and other common mistakes.

---

**Q19.** What does `import _ "github.com/lib/pq"` mean?
- a) The import is commented out
- b) Import for side effects only — runs `init()` but exports nothing into this file
- c) The package is optional
- d) Imports with alias `_`

> ✅ **b** — The blank identifier discards the package name. Used to register database drivers, etc., via `init()`.

---

**Q20.** True or False: `go install ./cmd/app/` places the binary in `$GOPATH/bin` (or `$GOBIN`).

> ✅ **True** — `go install` is for installing tools/binaries; `go build` just creates a binary in the current directory.

---

**Q21.** What is a Go **workspace** (`go.work` file)?
- a) Replaces `go.mod` for large projects
- b) Allows developing multiple modules simultaneously using local versions
- c) Configures IDE settings
- d) Defines build tags for different environments

> ✅ **b** — `go.work` lets you use local (unreleased) versions of multiple modules together without editing `go.mod`.

---

**Q22.** What does `go list -m all` show?
- a) All installed Go binaries
- b) All modules in the current build (direct + transitive dependencies)
- c) All packages in the current module
- d) All exported symbols in the current package

> ✅ **b** — `go list -m all` lists every module your project depends on.

---

**Q23.** True or False: Packages in Go are identified by their directory path, not the `package` name declaration.

> ✅ **True** — The import path is the directory path. The `package` name is what you use locally (and can differ from the last path element by convention, but shouldn't).

---

**Q24.** What is the role of the `//go:build` directive?
- a) Marks a function for inlining
- b) Specifies build constraints — which OS/arch/tags this file applies to
- c) Embeds files into the binary
- d) Disables the garbage collector for that file

> ✅ **b** — E.g., `//go:build linux && amd64` means this file only compiles for 64-bit Linux.

---

**Q25.** What does `go env GOMODCACHE` show?
- a) The location of the current module's source
- b) The directory where downloaded module source code is cached
- c) The Go installation directory
- d) The workspace config

> ✅ **b** — Typically `$GOPATH/pkg/mod`. All downloaded modules live here.

---

## PART 2 — Variables & Types (Q26–Q50)

**Q26.** What is the zero value of `*int` (pointer to int)?
- a) `0`
- b) A pointer to `0`
- c) `nil`
- d) Compile error — can't have zero value for pointers

> ✅ **c** — All pointers zero to `nil` in Go.

---

**Q27.** Does this compile? `var x int = 3.14`
- a) Yes
- b) No — cannot use untyped float constant 3.14 as int
- c) Yes, x becomes 3 (truncated)
- d) Yes, but only with `//go:noescape`

> ✅ **b** — Go does not truncate. The compiler rejects this since 3.14 is not representable as `int`.

---

**Q28.** What is the output?
```go
var x int8 = 127
x++
fmt.Println(x)
```
- a) 128
- b) -128
- c) Compile error
- d) Runtime panic

> ✅ **b** — Integer overflow wraps silently in Go. `int8` max is 127; incrementing wraps to -128.

---

**Q29.** True or False: `:=` (short variable declaration) can be used at package level.

> ✅ **False** — `:=` is only valid inside functions. Package-level must use `var`.

---

**Q30.** What does `iota` represent in a `const` block?
- a) The current loop index
- b) A sequential integer starting at 0, incrementing per constant in the block
- c) The memory address of the constant
- d) A random unique value

> ✅ **b** — `iota` = 0 for first const, 1 for second, etc. Resets to 0 in each new `const` block.

---

**Q31.** What is the output?
```go
const (
    A = iota  // 0
    B         // 1
    C = 10    // 10
    D = iota  // ?
)
fmt.Println(A, B, C, D)
```
- a) `0 1 10 3`
- b) `0 1 10 10`
- c) `0 1 10 2`
- d) `0 1 10 0`

> ✅ **a** — `iota` continues counting regardless of explicit values. D is the 4th const (index 3), so `iota = 3`.

---

**Q32.** What is the default type of `x` in `x := 3.14`?
- a) `float32`
- b) `float64`
- c) `double`
- d) Untyped float until assigned

> ✅ **b** — Floating-point literals default to `float64`.

---

**Q33.** What does `len("hello, 世界")` return?
- a) 9 (character count)
- b) 7 (visible characters)
- c) 13 (bytes — each Chinese char is 3 UTF-8 bytes)
- d) 11

> ✅ **c** — `len()` on strings returns **byte count**, not character count. "hello, " = 7 bytes + "世界" = 6 bytes = 13.

---

**Q34.** How do you correctly iterate over Unicode **characters** (runes) in a string?
- a) `for i := 0; i < len(s); i++ { ch := s[i] }`
- b) `for i, ch := range s { }`
- c) Both are equivalent
- d) `for ch := range s { }`

> ✅ **b** — `range` on a string iterates over runes (Unicode code points). `s[i]` returns a `byte`, not a rune.

---

**Q35.** True or False: In Go, you can implicitly convert `int32` to `int64`.

> ✅ **False** — Go requires explicit conversion: `int64(myInt32Var)`. No implicit numeric conversions.

---

**Q36.** What is the type of `s[0]` when `s` is a `string`?
- a) `rune` (int32)
- b) `byte` (uint8)
- c) `string` of length 1
- d) `int`

> ✅ **b** — String indexing returns a `byte`. Use `range` or `[]rune(s)[0]` for a rune.

---

**Q37.** Is `var x interface{}` different from `var x any`?
- a) Yes — `any` is a different, stricter type
- b) No — `any` is an alias for `interface{}` (introduced in Go 1.18)
- c) `any` is only valid in generic constraints
- d) `any` requires an import

> ✅ **b** — `any` = `interface{}` exactly. Just a more readable alias.

---

**Q38.** True or False: Constants in Go can have struct types.

> ✅ **False** — Constants must be boolean, numeric (integer/float/complex), or string. Struct constants are not allowed.

---

**Q39.** What is the output?
```go
a, b := 1, 2
a, b = b, a
fmt.Println(a, b)
```
- a) `1 2`
- b) `2 1`
- c) Compile error
- d) Undefined behavior

> ✅ **b** — Go evaluates all right-hand expressions BEFORE any assignment. This is a safe swap.

---

**Q40.** Given `type Celsius float64`, can you add `Celsius(30) + float64(5)` directly?
- a) Yes — same underlying type
- b) No — requires explicit conversion: `Celsius(30) + Celsius(5)`
- c) Yes, in Go 1.18+
- d) Yes, if using `+` with `any`

> ✅ **b** — Even with the same underlying type, named types in Go are distinct. Requires `Celsius(30) + Celsius(float64(5))` or `Celsius(30+5)`.

---

**Q41.** What is the size of `int` on a 64-bit machine?
- a) Always 4 bytes
- b) Always 8 bytes
- c) 8 bytes on 64-bit, 4 bytes on 32-bit (platform-dependent)
- d) Specified in `go.mod`

> ✅ **c** — `int` (and `uint`) is platform-native width. Use `int32`/`int64` when you need a specific size.

---

**Q42.** What is the output?
```go
var s string
fmt.Println(s == "")
fmt.Println(len(s))
```
- a) `false` then `0`
- b) `true` then `0`
- c) Compile error
- d) `false` then `nil`

> ✅ **b** — Zero value of `string` is `""`. `"" == ""` is true, `len("")` is 0.

---

**Q43.** True or False: `var x = 10` and `x := 10` are always exactly equivalent inside a function.

> ✅ **False** — They behave the same for new variables, but `:=` requires at least one **new** variable on the left side (in multi-assignment). `var` can re-declare existing variables only in specific scopes.

---

**Q44.** What happens with `json.Marshal([]string(nil))`?
- a) Returns `[]`, `nil`
- b) Returns `null`, `nil`
- c) Returns `""`, `nil`
- d) Returns an error

> ✅ **b** — A nil slice marshals to JSON `null`. An empty (non-nil) slice `[]string{}` marshals to `[]`.

---

**Q45.** What is the output?
```go
const x = 1 << 10
fmt.Println(x)
```
- a) `10`
- b) `1024`
- c) `1010` (binary)
- d) Compile error

> ✅ **b** — `1 << 10` = 2^10 = 1024. Bit-shift works in constant expressions.

---

**Q46.** True or False: Untyped integer constants can be assigned to any integer type without explicit conversion, as long as the value fits.

> ✅ **True** — `const x = 42` can be used as `int8`, `int64`, `uint32`, etc. as long as 42 fits the range.

---

**Q47.** What is `rune` in Go?
- a) An alias for `byte` (uint8)
- b) An alias for `int32`, representing a Unicode code point
- c) A special Unicode package type
- d) `int64` on 64-bit systems

> ✅ **b** — `rune` = `int32`. It holds a Unicode code point (0 to 0x10FFFF).

---

**Q48.** Does this compile? `_ := 42`
- a) Yes
- b) No — `:=` with only `_` on the left side is invalid

> ✅ **b** — `_` cannot be the only variable on the left of `:=`. "no new variables on left side of :=".

---

**Q49.** What is the zero value of `map[string]int`?
- a) `map[string]int{}` (empty map)
- b) `nil`
- c) `0`
- d) An empty initialized map

> ✅ **b** — Zero value is `nil`. A nil map is **readable** (returns zero values) but **panics on write**.

---

**Q50.** True or False: `var x, y int = 1, 2` is valid Go.

> ✅ **True** — Multiple variables can be declared and assigned in one `var` statement.

---

## PART 3 — Control Flow (Q51–Q75)

**Q51.** True or False: In Go, `switch` falls through to the next case by default (like C/C++).

> ✅ **False** — Go's `switch` does NOT fall through by default. Use `fallthrough` keyword explicitly.

---

**Q52.** What is the output?
```go
x := 5
switch {
case x > 3:
    fmt.Println("big")
    fallthrough
case x > 4:
    fmt.Println("bigger")
case x > 10:
    fmt.Println("biggest")
}
```
- a) `big`
- b) `big` then `bigger`
- c) `big` then `bigger` then `biggest`
- d) Compile error

> ✅ **b** — `fallthrough` forces execution into the NEXT case body regardless of its condition. It does NOT evaluate the next condition.

---

**Q53.** What is `for {}` equivalent to?
- a) `for true {}`
- b) `while (true) {}`
- c) An infinite loop
- d) All of the above (a and c; b is not Go syntax)

> ✅ **d** — `for {}` is the idiomatic infinite loop in Go. `for true {}` also works.

---

**Q54.** What is the output?
```go
for i := 0; i < 3; i++ {
    defer fmt.Println(i)
}
```
- a) `0 1 2`
- b) `2 1 0`
- c) `2 2 2`
- d) Nothing — defer in a loop doesn't work

> ✅ **b** — `defer` is LIFO (stack order). Arguments are captured at defer time, so i=0,1,2 are captured correctly and print in reverse order.

---

**Q55.** Does this compile?
```go
y := 10
if x := y * 2; x > 15 {
    fmt.Println(x)
}
fmt.Println(x) // ← this line
```
- a) Yes
- b) No — `x` is scoped to the `if` block

> ✅ **b** — Variables declared in an `if` init statement are scoped to the `if`/`else` block only.

---

**Q56.** True or False: Go has a `while` keyword.

> ✅ **False** — Go only has `for`. The while-style loop is `for condition { }`.

---

**Q57.** True or False: `goto` is valid in Go.

> ✅ **True** — Go has `goto` with one rule: you cannot jump forward over a variable declaration.

---

**Q58.** What does `select {}` (empty select) do?
- a) Returns immediately
- b) Blocks forever
- c) Compile error
- d) Yields the current goroutine once

> ✅ **b** — An empty `select` has no cases to match, so it blocks forever. Used in server `main()` functions.

---

**Q59.** What does `range` return when iterating over a channel?
- a) `(index, value)`
- b) Just `value` (one variable)
- c) `(key, value)` like a map
- d) Compile error — range doesn't work on channels

> ✅ **b** — `for v := range ch` receives one value at a time. Exits when channel is closed.

---

**Q60.** What is the output?
```go
m := map[string]int{"a": 1, "b": 2}
for k, v := range m {
    fmt.Println(k, v)
}
```
- a) Always `a 1` then `b 2`
- b) Always `b 2` then `a 1`  
- c) Iteration order is random
- d) Compile error

> ✅ **c** — Go deliberately randomizes map iteration order. Never rely on map order.

---

**Q61.** What is a type switch?
- a) `switch x { case int: }`  
- b) `switch x.(type) { case int: }` 
- c) `switch typeof(x) { case int: }`  
- d) `switch reflect.TypeOf(x) { }`

> ✅ **b** — The `.( type)` syntax is the type switch assertion. Only valid in a switch statement.

---

**Q62.** True or False: In a type switch `switch v := x.(type)`, the variable `v` has the specific concrete type in each case branch.

> ✅ **True** — In `case int: v` is `int`, in `case string: v` is `string`, etc.

---

**Q63.** What does `continue` do in a `for range` loop?
- a) Exits the loop
- b) Skips to the next iteration
- c) Restarts the loop from index 0
- d) Continues to the outer loop

> ✅ **b** — `continue` skips the rest of the current iteration body and goes to the next element.

---

**Q64.** Can `break` target an outer loop using a label in Go?
- a) No — `break` only exits the immediately enclosing loop
- b) Yes — `break LABEL` exits the labeled statement
- c) Yes, but only for `for` loops, not `switch`
- d) Labels are not supported in Go

> ✅ **b** — `break LABEL` exits the labeled `for`, `switch`, or `select`.

---

**Q65.** What is the output?
```go
n := 0
for n < 5 {
    n++
}
fmt.Println(n)
```
- a) `4`
- b) `5`
- c) Compile error — no C-style while loops
- d) Infinite loop

> ✅ **b** — `for condition {}` is the while-loop form. Exits when n=5.

---

**Q66.** True or False: `switch` can match multiple values in one case using commas: `case 1, 2, 3:`.

> ✅ **True** — This is idiomatic Go and more readable than `fallthrough`.

---

**Q67.** If you modify the loop variable `i` inside a `for i := range slice` loop, does it affect the next iteration?
- a) Yes — `i` is the actual index variable
- b) No — range evaluates next index independently
- c) Compile error — can't modify range variable
- d) Yes, but only for slices not maps

> ✅ **a** — You can modify `i` inside the loop, and it does affect the next iteration (like any regular variable).

---

**Q68.** What is the output?
```go
for i := 0; i < 5; i++ {
    if i == 3 {
        break
    }
    fmt.Print(i, " ")
}
```
- a) `0 1 2 3 4 `
- b) `0 1 2 `
- c) `0 1 2 3 `
- d) Compile error

> ✅ **b** — breaks when i=3, printing 0, 1, 2 before.

---

**Q69.** True or False: You can use `range` over an integer in Go 1.22+: `for i := range 5 { }`.

> ✅ **True** — Go 1.22 added ranging over integers. `range 5` iterates i from 0 to 4.

---

**Q70.** What happens if you call `delete(m, key)` on a key that doesn't exist?
- a) Runtime panic
- b) Returns false
- c) No-op — silently does nothing
- d) Compile error

> ✅ **c** — `delete` is safe on missing keys. No-op.

---

**Q71.** True or False: A `switch` without a condition is equivalent to `switch true`.

> ✅ **True** — `switch { case x > 0: }` is the same as `switch true { case x > 0: }`. Cases are evaluated as boolean expressions.

---

**Q72.** What does `for k := range m` (one variable) iterate for a map?
- a) Values only
- b) Keys only
- c) Indices only
- d) Both key and value (k is a pair)

> ✅ **b** — With one variable, `range` on map yields only keys.

---

**Q73.** True or False: Modifying a slice while ranging over it is safe and the changes are visible to the range.

> ✅ **False** — `range` captures the slice header (pointer, len, cap) once at start. Appending (which may create new backing array) won't be seen. Modifications within existing len ARE visible.

---

**Q74.** What does `fallthrough` do at the end of the last `case` in a switch?
- a) Compile error
- b) Exits the switch (falls through to after switch)
- c) Runtime panic
- d) Goes to the `default` case

> ✅ **a** — `fallthrough` cannot appear in the last case of a switch — compile error.

---

**Q75.** What is the output?
```go
x := 2
switch x {
case 1:
    fmt.Println("one")
case 2:
    fmt.Println("two")
case 2:
    fmt.Println("two again")
}
```
- a) `two`
- b) `two` then `two again`
- c) Compile error — duplicate case
- d) `two again`

> ✅ **c** — Duplicate case values are a **compile error** in Go.

---

## PART 4 — Functions (Q76–Q100)

**Q76.** True or False: Go supports function overloading (same name, different parameter types).

> ✅ **False** — Go does NOT support overloading. Each function in a package must have a unique name.

---

**Q77.** What is the output?
```go
func f() (int, int) { return 1, 2 }
a, b := f()
fmt.Println(a, b)
```
- a) `1 2`
- b) `[1 2]`
- c) Compile error
- d) `1, 2`

> ✅ **a** — Multiple return values are separated by spaces in Println (default formatting).

---

**Q78.** What is the output of this closure?
```go
x := 10
add := func(n int) { x += n }
add(5)
fmt.Println(x)
```
- a) `10`
- b) `15`
- c) Compile error
- d) `5`

> ✅ **b** — The closure captures `x` by **reference** (not by copy). Modifying it inside the closure modifies the outer `x`.

---

**Q79.** What is the (classic) output of this code in Go <= 1.21?
```go
funcs := make([]func(), 3)
for i := 0; i < 3; i++ {
    funcs[i] = func() { fmt.Println(i) }
}
for _, f := range funcs { f() }
```
- a) `0 1 2`
- b) `3 3 3`
- c) Compile error
- d) `2 2 2`

> ✅ **b** — All closures capture the **same** variable `i`. By the time they run, `i = 3`. Fixed in Go 1.22 where loop variables get per-iteration scope.

---

**Q80.** How do you fix the loop closure bug in Go <= 1.21?
- a) `go func(i int) { fmt.Println(i) }(i)` — pass as argument (copy)
- b) `i := i; funcs[i] = func() { fmt.Println(i) }` — shadow with new variable
- c) Both work
- d) Use a channel

> ✅ **c** — Both A (pass as arg) and B (create new variable shadowing i) correctly capture the current value.

---

**Q81.** True or False: Named return values in Go are zero-initialized.

> ✅ **True** — `func f() (result int, err error)` — result is 0, err is nil at function start.

---

**Q82.** What is the output?
```go
func f() (result int) {
    defer func() { result++ }()
    return 10
}
fmt.Println(f())
```
- a) `10`
- b) `11`
- c) `0`
- d) Compile error

> ✅ **b** — `return 10` sets the named return `result = 10`. Then the deferred function runs and increments it to 11. Named returns + defer interact!

---

**Q83.** In what order do multiple deferred functions execute?
- a) FIFO — first deferred runs first
- b) LIFO — last deferred runs first (stack order)
- c) In parallel
- d) Random order

> ✅ **b** — Defers are a stack. Last `defer` call is the first to execute.

---

**Q84.** What happens when `panic` is called and no `recover` is set up?
- a) The current goroutine pauses and waits
- b) The program terminates, printing the panic value and stack trace
- c) Only the current goroutine dies; others continue
- d) An error is returned to the caller

> ✅ **b** — An unrecovered panic terminates the entire program (all goroutines).

---

**Q85.** When does `recover()` return a non-nil value?
- a) Any time it's called
- b) Only when called directly inside a deferred function during a panic
- c) Only in the main goroutine
- d) Only for runtime panics

> ✅ **b** — `recover()` only catches panics when called **directly** from a deferred function. Calling it from a function called by a defer does NOT work.

---

**Q86.** What is a variadic function?
- a) A function that can be overloaded
- b) A function that accepts a variable number of arguments as its last parameter (`...T`)
- c) A function returning multiple values
- d) A function with optional parameters

> ✅ **b** — `func sum(nums ...int) int` accepts 0 or more ints. Inside the function, `nums` is a `[]int`.

---

**Q87.** What does `func(s []int, fn func(int) int)` describe?
- a) Invalid syntax
- b) A function accepting a slice and a function as parameters
- c) A generic function
- d) An interface method

> ✅ **b** — Functions are first-class in Go and can be passed as arguments.

---

**Q88.** True or False: Deferred functions still run when a goroutine panics.

> ✅ **True** — Defers in the panicking goroutine run as the panic unwinds the stack. This is how `recover()` works.

---

**Q89.** What does `append(a, b...)` do when `a` and `b` are both `[]int`?
- a) Compile error — can't append slices
- b) Appends each element of `b` to `a`
- c) Appends slice `b` as a single element to `a`
- d) Merges `a` and `b` in sorted order

> ✅ **b** — The `...` "unpacks" `b`. Equivalent to `append(a, b[0], b[1], ...)`.

---

**Q90.** True or False: A function that returns a pointer to a local variable causes undefined behavior in Go (like in C++).

> ✅ **False** — It's completely safe. The variable **escapes** to the heap. The GC ensures it lives as long as there are references.

---

**Q91.** What is the type of a function `func f(int) bool`?
- a) Not a type — it's just a signature
- b) `func(int) bool`
- c) `Func`
- d) `func[int, bool]`

> ✅ **b** — Function types in Go are exactly their signature: `func(int) bool`.

---

**Q92.** True or False: `recover()` can be called from a function that is called by a deferred function (two levels deep from defer) and still catch the panic.

> ✅ **False** — `recover()` only works when called **directly** from a deferred function, not transitively.

---

**Q93.** What does this print?
```go
func safeDiv(a, b int) (r int, err error) {
    defer func() {
        if v := recover(); v != nil {
            err = fmt.Errorf("panic: %v", v)
        }
    }()
    return a / b, nil
}
r, e := safeDiv(10, 0)
fmt.Println(r, e)
```
- a) `0 <nil>`
- b) `0 panic: runtime error: integer divide by zero`
- c) Program crashes
- d) `10 <nil>`

> ✅ **b** — Division by zero panics; the deferred recover catches it and sets the named return `err`.

---

**Q94.** True or False: A method in Go is just a function with an explicit receiver parameter.

> ✅ **True** — `func (u *User) Name() string` is syntactic sugar for a function with a receiver.

---

**Q95.** What is the output?
```go
func double(n int) int { return n * 2 }
nums := []int{1, 2, 3}
result := make([]int, len(nums))
for i, v := range nums {
    result[i] = double(v)
}
fmt.Println(result)
```
- a) `[1 2 3]`
- b) `[2 4 6]`
- c) Compile error
- d) `[2 4 6]` but incorrect approach

> ✅ **b** — Straightforward function application in a loop.

---

**Q96.** What is `panic("message")` equivalent to in C++?
- a) `throw std::exception("message")`
- b) `assert(false)`
- c) `std::terminate()`
- d) It has no C++ equivalent

> ✅ **a** — Conceptually similar to throwing, but Go's `panic` propagates by unwinding deferred functions, not the exception stack.

---

**Q97.** True or False: An anonymous function (closure) can call itself recursively.

> ✅ **True** — But you must assign it to a variable first: `var f func(int) int; f = func(n int) int { if n == 0 { return 1 }; return n * f(n-1) }`.

---

**Q98.** What does `defer f()` capture about `f()`'s arguments?
- a) Arguments are evaluated lazily when defer runs
- b) Arguments are evaluated immediately at the `defer` statement
- c) Arguments are not evaluated until the function returns
- d) No arguments are captured

> ✅ **b** — `defer fmt.Println(x)` evaluates `x` immediately. The call itself is deferred, but argument values are fixed at the defer statement.

---

**Q99.** What is the output?
```go
x := 1
defer fmt.Println(x)
x++
```
- a) `1`
- b) `2`
- c) Compile error
- d) `0`

> ✅ **a** — `x` is evaluated at the `defer` call, capturing value `1`. The later `x++` doesn't affect the captured value.

---

**Q100.** True or False: You can use `fmt.Println(f())` directly if `f()` returns multiple values.

> ✅ **True** — When a multi-return function is the **only** argument to another function that accepts `...interface{}`, Go automatically expands them. `fmt.Println(f())` works.

---

*End of Quiz Part 1 (Q1–Q100)*
*Continue with go_quiz_part2.md for Q101–Q200*
