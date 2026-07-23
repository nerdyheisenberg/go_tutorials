# Go Deep Dive — Part 16: Standard Library — Key Packages

---

## Chapter 1: fmt — Formatting and Printing

```go
import "fmt"

// Printing functions
fmt.Print("hello")           // no newline, uses default format
fmt.Println("hello", "world") // space between args, newline at end
fmt.Printf("Hello, %s! You are %d.\n", "Rohit", 30)

// Printing to writers
fmt.Fprint(os.Stderr, "error!")
fmt.Fprintln(w, "HTTP response")
fmt.Fprintf(w, "status: %s", status)

// Building strings (without printing)
s1 := fmt.Sprint(1, 2, 3)           // "1 2 3"
s2 := fmt.Sprintf("%.2f", 3.14159)  // "3.14"
s3 := fmt.Sprintln("hi")            // "hi\n"

// Scanning input
var name string
var age int
fmt.Scan(&name, &age)              // reads space-separated
fmt.Scanf("%s %d", &name, &age)    // formatted read
fmt.Sscanf("Rohit 30", "%s %d", &name, &age)  // from string

// FORMAT VERBS — Complete Reference:
// General:
// %v   — default format for any value
// %+v  — struct with field names
// %#v  — Go-syntax representation
// %T   — type name
// %p   — pointer address

// Integer:
// %d   — decimal
// %b   — binary
// %o   — octal
// %x   — hex lowercase
// %X   — hex uppercase
// %c   — character (Unicode code point)
// %q   — quoted character

// Floating point:
// %f   — decimal (no exponent)
// %e   — scientific notation
// %g   — shorter of %f and %e
// %.2f — 2 decimal places
// %8.2f — width 8, 2 decimal places (right-aligned)
// %-8.2f — left-aligned

// String:
// %s   — string
// %q   — quoted string
// %10s — right-padded to width 10
// %-10s — left-padded to width 10

// Boolean:
// %t   — true/false

// Example: debug printing
type User struct { Name string; Age int }
u := User{"Rohit", 30}
fmt.Printf("%v\n", u)   // {Rohit 30}
fmt.Printf("%+v\n", u)  // {Name:Rohit Age:30}
fmt.Printf("%#v\n", u)  // main.User{Name:"Rohit", Age:30}
```

---

## Chapter 2: strings — String Manipulation

```go
import "strings"

s := "Hello, World! Hello, Go!"

// Search
strings.Contains(s, "World")    // true
strings.ContainsAny(s, "aeiou") // true (contains any vowel)
strings.ContainsRune(s, '!')    // true
strings.Count(s, "Hello")       // 2
strings.HasPrefix(s, "Hello")   // true
strings.HasSuffix(s, "Go!")     // true
strings.Index(s, "World")       // 7  (-1 if not found)
strings.LastIndex(s, "Hello")   // 14
strings.IndexAny(s, "aeiou")    // 1 (first vowel 'e')
strings.IndexRune(s, '!')       // 12

// Transform
strings.ToUpper(s)
strings.ToLower(s)
strings.Title(s)           // deprecated, use golang.org/x/text
strings.TrimSpace("  hi ") // "hi"
strings.Trim(s, "!H")      // trims leading/trailing '!' or 'H'
strings.TrimLeft(s, "He")  // trims leading H or e
strings.TrimRight(s, "!")
strings.TrimPrefix(s, "Hello, ")  // "World! Hello, Go!"
strings.TrimSuffix(s, "Go!")      // "Hello, World! Hello, "
strings.Replace(s, "Hello", "Hi", 1)   // replace first
strings.ReplaceAll(s, "Hello", "Hi")   // replace all
strings.Map(func(r rune) rune { if r == 'l' { return 'r' }; return r }, "hello")  // "herro"

// Split
strings.Split("a,b,c", ",")         // ["a" "b" "c"]
strings.SplitN("a,b,c,d", ",", 2)   // ["a" "b,c,d"]
strings.SplitAfter("a,b,c", ",")    // ["a," "b," "c"]
strings.Fields("  a  b  c  ")       // ["a" "b" "c"]
strings.Cut("a=b=c", "=")           // "a", "b=c", true (like SplitN(,,2) but safer)

// Join
strings.Join([]string{"a", "b", "c"}, ", ")  // "a, b, c"
strings.Repeat("ab", 3)                       // "ababab"

// Builder — for efficient concatenation
var b strings.Builder
b.Grow(100)  // pre-allocate
for _, word := range words {
    b.WriteString(word)
    b.WriteByte(' ')
}
result := b.String()

// Reader — io.Reader from string (no allocation)
r := strings.NewReader("hello world")
io.Copy(os.Stdout, r)

// Replacer — efficient multi-replacement
r2 := strings.NewReplacer(
    "Hello", "Hi",
    "World", "Earth",
    "Go!", "Gopher!",
)
r2.Replace(s)  // all replacements in one pass
```

---

## Chapter 3: strconv — Type Conversions

```go
import "strconv"

// int ↔ string
n := 42
s := strconv.Itoa(n)       // "42" (int to ASCII)
n2, err := strconv.Atoi(s) // 42, nil (ASCII to int)
// strconv.Atoi error: *strconv.NumError with .Err = strconv.ErrSyntax or ErrRange

// Parse numbers from strings (more flexible than Atoi)
i64, err := strconv.ParseInt("FF", 16, 64)  // base 16 (hex) → 255
u64, err := strconv.ParseUint("11111111", 2, 8)  // base 2 (binary) → 255
f64, err := strconv.ParseFloat("3.14159", 64)   // float64

// Format numbers to strings
s1 := strconv.FormatInt(255, 16)   // "ff" (decimal 255 in hex)
s2 := strconv.FormatInt(255, 2)    // "11111111" (in binary)
s3 := strconv.FormatFloat(3.14159, 'f', 2, 64)  // "3.14" (f=decimal, 2=precision)

// bool ↔ string
b := true
s4 := strconv.FormatBool(b)  // "true"
b2, err := strconv.ParseBool("true")   // true, nil
b3, err := strconv.ParseBool("1")      // true, nil
b4, err := strconv.ParseBool("false")  // false, nil

// Quoting strings (useful for debugging)
s5 := strconv.Quote("hello\nworld")     // "\"hello\\nworld\""
s6, err := strconv.Unquote(`"hello"`)   // "hello", nil

// AppendXxx variants — avoid allocation by appending to existing []byte
buf := make([]byte, 0, 20)
buf = strconv.AppendInt(buf, 255, 16)   // appends "ff" to buf
buf = strconv.AppendFloat(buf, 3.14, 'f', 2, 64)
```

---

## Chapter 4: os — Operating System Interface

```go
import "os"

// File operations
file, err := os.Open("data.txt")          // read-only: os.O_RDONLY
file, err = os.Create("output.txt")       // create/truncate: O_RDWR|O_CREATE|O_TRUNC
file, err = os.OpenFile("data.txt", os.O_APPEND|os.O_WRONLY, 0644) // flags + permissions

defer file.Close()

// Read/Write
data := make([]byte, 1024)
n, err := file.Read(data)                 // read up to len(data) bytes
n, err = file.Write([]byte("hello\n"))    // write bytes

// Seek
offset, err := file.Seek(0, io.SeekStart)  // seek to start
offset, err = file.Seek(0, io.SeekEnd)     // seek to end
offset, err = file.Seek(-10, io.SeekCurrent) // seek backwards 10

// Stat
info, err := os.Stat("data.txt")
if os.IsNotExist(err) {
    fmt.Println("file doesn't exist")
}
fmt.Println(info.Name(), info.Size(), info.ModTime(), info.IsDir())

// Directory operations
entries, err := os.ReadDir("./")          // list directory
for _, e := range entries {
    fmt.Println(e.Name(), e.IsDir())
}
os.Mkdir("newdir", 0755)
os.MkdirAll("a/b/c", 0755)              // mkdir -p
os.Remove("file.txt")                   // delete file or empty dir
os.RemoveAll("dir/")                    // rm -rf

// Rename / Move
os.Rename("old.txt", "new.txt")

// Read entire file in one shot
data2, err := os.ReadFile("data.txt")   // returns []byte
err = os.WriteFile("output.txt", data2, 0644)

// Environment
val := os.Getenv("HOME")
os.Setenv("MY_VAR", "value")
os.Unsetenv("MY_VAR")
allEnv := os.Environ()                   // []string of KEY=value pairs

// Process
os.Args                                   // []string of command-line arguments
os.Exit(1)                               // exit with code (no defers run!)
pid := os.Getpid()
```

---

## Chapter 5: io — I/O Primitives

```go
import "io"

// Key functions
data, err := io.ReadAll(r)               // read until EOF or error
n, err := io.Copy(dst, src)             // copy from src to dst reader→writer
n, err = io.CopyN(dst, src, 1024)      // copy at most N bytes
n, err = io.WriteString(w, "hello")    // write a string to a Writer

// Composing readers and writers
limited := io.LimitReader(r, 1024)     // read at most 1024 bytes
tee := io.TeeReader(r, w)             // read from r, also write to w
multi := io.MultiReader(r1, r2, r3)   // read from r1, then r2, then r3
mw := io.MultiWriter(w1, w2, w3)     // write to all writers simultaneously

// Pipes — synchronous in-memory pipe
pr, pw := io.Pipe()
go func() {
    fmt.Fprintln(pw, "hello from pipe")
    pw.Close()
}()
io.Copy(os.Stdout, pr)

// io.ReadFull — read EXACTLY len(buf) bytes
buf := make([]byte, 10)
n, err = io.ReadFull(r, buf)  // error if fewer than 10 bytes available

// io.NopCloser — wrap a Reader with a no-op Close
rc := io.NopCloser(strings.NewReader("hello"))  // now has Close() method
```

---

## Chapter 6: bufio — Buffered I/O

```go
import "bufio"

// Buffered reader — reduces system calls
file, _ := os.Open("large.txt")
br := bufio.NewReader(file)
// or: br = bufio.NewReaderSize(file, 65536)  // 64KB buffer

// Read line by line (most common use)
scanner := bufio.NewScanner(file)
for scanner.Scan() {
    line := scanner.Text()  // each line WITHOUT newline
    fmt.Println(line)
}
if err := scanner.Err(); err != nil {
    log.Fatal(err)
}

// Scan with larger buffer (for long lines)
scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

// Custom split function
scanner.Split(bufio.ScanWords)   // scan word by word
scanner.Split(bufio.ScanBytes)   // scan byte by byte
scanner.Split(bufio.ScanRunes)   // scan rune by rune

// Buffered writer — reduces system calls for many small writes
bw := bufio.NewWriter(file)
bw.WriteString("hello\n")
bw.WriteString("world\n")
bw.Flush()  // MUST flush — unflushed data is lost!
// OR: use defer bw.Flush()

// bufio.ReadWriter — both read and write
brw := bufio.NewReadWriter(br, bw)

// Read with specific delimiters
line, err := br.ReadString('\n')  // read until newline (includes delimiter)
line2, err := br.ReadBytes('\n')  // same but returns []byte
```

---

## Chapter 7: encoding/json — JSON Handling

```go
import "encoding/json"

// Marshal: Go value → JSON bytes
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Password  string    `json:"-"`           // never include in JSON
    CreatedAt time.Time `json:"created_at"`
    Points    *int      `json:"points,omitempty"` // omit if nil
    Tags      []string  `json:"tags,omitempty"`   // omit if empty slice
}

u := User{ID: 1, Name: "Rohit", Email: "rohit@example.com"}
data, err := json.Marshal(u)
// {"id":1,"name":"Rohit","email":"rohit@example.com","created_at":"..."}

// Pretty-print
prettyData, err := json.MarshalIndent(u, "", "  ")

// Unmarshal: JSON bytes → Go value
var u2 User
err = json.Unmarshal(data, &u2)  // note: pointer to u2!

// Streaming encode/decode (preferred for HTTP)
encoder := json.NewEncoder(w)    // writes to io.Writer
encoder.SetIndent("", "  ")     // optional pretty-print
err = encoder.Encode(u)          // appends newline after JSON

decoder := json.NewDecoder(r)    // reads from io.Reader
decoder.DisallowUnknownFields()  // return error if JSON has extra fields
err = decoder.Decode(&u2)

// Dynamic JSON (when structure is unknown)
var raw map[string]interface{}
json.Unmarshal(data, &raw)
name := raw["name"].(string)   // type assertion needed

// json.RawMessage — delay parsing
type Response struct {
    Status string          `json:"status"`
    Data   json.RawMessage `json:"data"`  // keep as raw JSON
}

var resp Response
json.Unmarshal(data, &resp)
// Parse resp.Data later based on resp.Status

// Custom JSON marshaling
type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
    return json.Marshal(time.Duration(d).String())  // "5m30s"
}

func (d *Duration) UnmarshalJSON(b []byte) error {
    var s string
    if err := json.Unmarshal(b, &s); err != nil { return err }
    dur, err := time.ParseDuration(s)
    if err != nil { return err }
    *d = Duration(dur)
    return nil
}
```

---

## Chapter 8: time — Time and Duration

```go
import "time"

// Current time
now := time.Now()           // local time  
utc := time.Now().UTC()     // UTC

// Creating specific times
t := time.Date(2024, time.January, 15, 10, 30, 0, 0, time.UTC)

// Durations
d := 2*time.Hour + 30*time.Minute + 15*time.Second
fmt.Println(d)  // 2h30m15s

d2 := 100 * time.Millisecond
d3 := 5 * time.Microsecond

// Duration arithmetic
until := time.Until(t)          // Duration until a time (negative if past)
since := time.Since(t)          // Duration since a time
t2 := t.Add(24 * time.Hour)    // time + duration
t3 := t.Add(-time.Hour)         // time - duration
diff := t2.Sub(t)               // Duration between two times

// Comparison
t.Before(t2)  // true
t.After(t2)   // false
t.Equal(t)    // true

// Formatting and parsing
// Go uses a REFERENCE TIME: Mon Jan 2 15:04:05 MST 2006
// (= 01/02 03:04:05PM '06 -0700)
s := now.Format("2006-01-02 15:04:05")  // "2024-01-15 10:30:00"
s2 := now.Format(time.RFC3339)           // "2024-01-15T10:30:00Z"
s3 := now.Format("Jan 2, 2006")         // "Jan 15, 2024"

parsed, err := time.Parse("2006-01-02", "2024-01-15")
parsed2, err := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")

// Unix timestamps
unix := now.Unix()      // int64 seconds
unixNano := now.UnixNano()  // int64 nanoseconds
back := time.Unix(unix, 0)  // from Unix timestamp

// Timezones
loc, _ := time.LoadLocation("Asia/Kolkata")
ist := now.In(loc)
fmt.Println(ist.Format("2006-01-02 15:04:05 MST"))

// Timer and Ticker (for goroutines — see Part 10/11)
timer := time.NewTimer(5 * time.Second)
defer timer.Stop()
<-timer.C  // fires once after 5s

ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()
<-ticker.C  // fires every second

// sleep
time.Sleep(100 * time.Millisecond)

// Measuring elapsed time
start := time.Now()
doWork()
elapsed := time.Since(start)
fmt.Printf("took: %v\n", elapsed)
```

---

## Chapter 9: log — Standard Logging

```go
import "log"

// Default logger writes to stderr with date/time prefix
log.Println("server started")
log.Printf("listening on port %d", 8080)
log.Fatal("critical error!")    // calls os.Exit(1)
log.Fatalf("error: %v", err)
log.Panic("this always panics") // panics after logging
log.Panicf("panic: %v", err)

// Custom logger
logger := log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lshortfile)
// flags: log.Ldate, log.Ltime, log.Lmicroseconds, log.Lshortfile, log.Llongfile, log.LUTC

// Structured logging (Go 1.21+)
import "log/slog"

slog.Info("user logged in", "userID", 42, "ip", "192.168.1.1")
slog.Error("database error", "err", err, "query", "SELECT * FROM users")
slog.Debug("processing", "items", 100)
slog.Warn("slow query", "duration", 2*time.Second)

// Custom slog handler
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
})
logger2 := slog.New(handler)
logger2.Info("structured log", "key", "value")
// Output: {"time":"...","level":"INFO","msg":"structured log","key":"value"}

// With logger for a request
childLogger := logger2.With("requestID", "abc123")
childLogger.Info("processing request")  // always includes requestID
```

---

## Chapter 10: regexp — Regular Expressions

```go
import "regexp"

// Compile once, use many times (panics on bad pattern — use for constants)
re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)  // date pattern

// Safe compile (returns error)
re2, err := regexp.Compile(`\d+`)
if err != nil { log.Fatal(err) }

// Match check
re.MatchString("2024-01-15")   // true
re.MatchString("not-a-date")   // false

// Find first match
re2.FindString("abc 123 def")           // "123"
re2.FindStringIndex("abc 123")          // [4 7] (start, end)
re2.FindStringSubmatch(`Say (\w+)!`)   // with capture group

// Find all matches
re2.FindAllString("1 22 333", -1)    // ["1" "22" "333"]
re2.FindAllString("1 22 333", 2)     // ["1" "22"] (max 2)

// Replace
re2.ReplaceAllString("I have 3 cats and 2 dogs", "N")  // "I have N cats and N dogs"
re2.ReplaceAllStringFunc("123", func(s string) string { return "[" + s + "]" })

// Split
re2.Split("one2two3three", -1)  // ["one" "two" "three"]

// Named capture groups
reNamed := regexp.MustCompile(`(?P<year>\d{4})-(?P<month>\d{2})-(?P<day>\d{2})`)
match := reNamed.FindStringSubmatch("2024-01-15")
names := reNamed.SubexpNames()
for i, name := range names {
    if name != "" && i < len(match) {
        fmt.Printf("%s: %s\n", name, match[i])
    }
}
// year: 2024, month: 01, day: 15
```

---

## Chapter 11: Complete Testing of Stdlib Usage

```go
package stdlib_test

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
    "strings"
    "strconv"
    "testing"
    "time"
)

func TestStringOperations(t *testing.T) {
    tests := []struct {
        name   string
        input  string
        fn     func(string) string
        want   string
    }{
        {"ToUpper", "hello", strings.ToUpper, "HELLO"},
        {"TrimSpace", "  hi  ", strings.TrimSpace, "hi"},
        {"ReplaceAll", "aababcabc",
            func(s string) string { return strings.ReplaceAll(s, "a", "X") },
            "XXbXbcXbc"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.fn(tt.input)
            if got != tt.want { t.Errorf("got %q, want %q", got, tt.want) }
        })
    }
}

func TestStrconvRoundTrip(t *testing.T) {
    nums := []int{0, 1, -1, 42, -42, 1000000}
    for _, n := range nums {
        t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
            s := strconv.Itoa(n)
            got, err := strconv.Atoi(s)
            if err != nil { t.Fatalf("Atoi(%q) err: %v", s, err) }
            if got != n { t.Errorf("roundtrip: got %d, want %d", got, n) }
        })
    }
}

func TestJSONRoundTrip(t *testing.T) {
    type Person struct {
        Name  string `json:"name"`
        Age   int    `json:"age"`
        Email string `json:"email,omitempty"`
    }
    
    original := Person{Name: "Rohit", Age: 30, Email: "rohit@example.com"}
    
    data, err := json.Marshal(original)
    if err != nil { t.Fatalf("Marshal: %v", err) }
    
    var decoded Person
    if err := json.Unmarshal(data, &decoded); err != nil {
        t.Fatalf("Unmarshal: %v", err)
    }
    
    if original != decoded {
        t.Errorf("roundtrip: got %+v, want %+v", decoded, original)
    }
    
    // Test omitempty
    empty := Person{Name: "Bob", Age: 25}
    data2, _ := json.Marshal(empty)
    if strings.Contains(string(data2), "email") {
        t.Error("omitempty: email field should not appear when empty")
    }
}

func TestTimeFormatting(t *testing.T) {
    t1 := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
    
    tests := []struct {
        format string
        want   string
    }{
        {"2006-01-02", "2024-01-15"},
        {"15:04:05", "10:30:00"},
        {time.RFC3339, "2024-01-15T10:30:00Z"},
    }
    
    for _, tt := range tests {
        t.Run(tt.format, func(t *testing.T) {
            got := t1.Format(tt.format)
            if got != tt.want { t.Errorf("Format(%q) = %q, want %q", tt.format, got, tt.want) }
        })
    }
}

func TestBufioScanner(t *testing.T) {
    content := "line1\nline2\nline3\n"
    reader := strings.NewReader(content)
    
    scanner := bufio.NewScanner(reader)
    var lines []string
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        t.Fatalf("scanner error: %v", err)
    }
    
    if len(lines) != 3 { t.Errorf("expected 3 lines, got %d", len(lines)) }
    if lines[0] != "line1" { t.Errorf("line1 = %q, want 'line1'", lines[0]) }
}

func TestFileOperations(t *testing.T) {
    // Use a temp file
    f, err := os.CreateTemp("", "test-*.txt")
    if err != nil { t.Fatal(err) }
    defer os.Remove(f.Name())
    defer f.Close()
    
    // Write
    content := "hello, world!"
    if _, err := f.WriteString(content); err != nil { t.Fatal(err) }
    f.Close()
    
    // Read back
    data, err := os.ReadFile(f.Name())
    if err != nil { t.Fatal(err) }
    
    if string(data) != content {
        t.Errorf("read = %q, want %q", string(data), content)
    }
}
```

---

**Summary of Part 16:**
- `fmt`: `Printf` with `%v/%+v/%#v/%T` for debugging; `Sprintf` for string building
- `strings`: `Cut` for splitting on first delimiter; `Builder` for efficient concatenation
- `strconv`: `Atoi`/`Itoa` for int↔string; `ParseFloat`/`FormatFloat` for floats
- `os`: `ReadFile`/`WriteFile` for simple, `OpenFile` with flags for advanced
- `io`: `ReadAll`, `Copy`, `TeeReader`, `MultiWriter` — compose I/O pipelines
- `bufio`: Always wrap file I/O with `bufio.Scanner` for line-by-line reading; `Flush()` required
- `encoding/json`: Struct tags control marshaling; streaming with `Encoder`/`Decoder` preferred
- `time`: Go's reference time is `Mon Jan 2 15:04:05 MST 2006` — memorize it!
- `log/slog` (Go 1.21+): structured logging with key-value pairs; `With()` for context
- `regexp`: Compile once with `MustCompile`; use named groups for complex patterns
