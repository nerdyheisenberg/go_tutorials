# Go Deep Dive — Part 18: C++ to Go Migration — Complete Reference

---

## Chapter 1: Mental Model Shift

The most important thing when migrating from C++ to Go is accepting that **Go is a completely different language with a different philosophy** — not just "C++ without templates."

| C++ Concept | Go Equivalent | Key Difference |
|---|---|---|
| Class | Struct + Methods | No access specifiers on methods |
| Inheritance | Struct Embedding | NOT IS-A, just HAS-A |
| Virtual functions | Interfaces | Implicit, not declared |
| Templates | Generics (1.18+) | Limited but simpler |
| Smart pointers | GC | No RAII for memory |
| RAII | defer | Manual, but explicit |
| Exceptions | error return values | Explicit, visible |
| std::vector | []slice | Built-in, dynamic |
| std::map | map[K]V | Built-in |
| std::thread | goroutine | Lightweight, GC-managed |
| std::mutex | sync.Mutex | Same concept |
| std::unique_ptr | *T (GC) | No ownership semantics |
| Namespace | Package | One package per directory |
| `#include` | `import` | No header files |
| Overloading | Named functions | No function overloading |
| Default params | Options/variadic | Explicit |

---

## Chapter 2: Class → Struct + Methods

### C++ Class → Go Struct

```cpp
// C++ Class
class BankAccount {
private:
    double balance;
    std::string owner;
    std::vector<std::string> transactions;
    
public:
    explicit BankAccount(const std::string& owner, double initial_balance = 0.0)
        : owner(owner), balance(initial_balance) {}
    
    bool deposit(double amount) {
        if (amount <= 0) return false;
        balance += amount;
        transactions.push_back("deposit: " + std::to_string(amount));
        return true;
    }
    
    bool withdraw(double amount) {
        if (amount <= 0 || amount > balance) return false;
        balance -= amount;
        transactions.push_back("withdraw: " + std::to_string(amount));
        return true;
    }
    
    double getBalance() const { return balance; }
    const std::string& getOwner() const { return owner; }
};
```

```go
// Go Equivalent
package account

import (
    "errors"
    "fmt"
)

// Exported type, unexported fields
type BankAccount struct {
    owner        string   // unexported = private (package-level, not file-level)
    balance      float64
    transactions []string
}

// Constructor function (Go convention: NewXxx)
func NewBankAccount(owner string, initialBalance float64) (*BankAccount, error) {
    if owner == "" {
        return nil, errors.New("owner cannot be empty")
    }
    if initialBalance < 0 {
        return nil, fmt.Errorf("initial balance cannot be negative: %f", initialBalance)
    }
    return &BankAccount{
        owner:   owner,
        balance: initialBalance,
    }, nil
}

// Methods on pointer receiver (can modify state)
func (a *BankAccount) Deposit(amount float64) error {
    if amount <= 0 {
        return fmt.Errorf("deposit amount must be positive: %f", amount)
    }
    a.balance += amount
    a.transactions = append(a.transactions, fmt.Sprintf("deposit: %.2f", amount))
    return nil
}

func (a *BankAccount) Withdraw(amount float64) error {
    if amount <= 0 {
        return fmt.Errorf("withdraw amount must be positive: %f", amount)
    }
    if amount > a.balance {
        return fmt.Errorf("insufficient funds: balance=%.2f, requested=%.2f", a.balance, amount)
    }
    a.balance -= amount
    a.transactions = append(a.transactions, fmt.Sprintf("withdraw: %.2f", amount))
    return nil
}

// Methods on value receiver (read-only, no state change)
func (a BankAccount) Balance() float64 { return a.balance }
func (a BankAccount) Owner() string    { return a.owner }

// Stringer interface (like toString() in Java)
func (a BankAccount) String() string {
    return fmt.Sprintf("BankAccount{owner:%s, balance:%.2f}", a.owner, a.balance)
}
```

---

## Chapter 3: Inheritance → Embedding + Interfaces

### Single Inheritance → Embedding

```cpp
// C++ Inheritance
class Animal {
protected:
    std::string name;
public:
    explicit Animal(std::string name) : name(std::move(name)) {}
    virtual std::string speak() const = 0;  // pure virtual
    virtual std::string move() const { return name + " moves"; }
    std::string getName() const { return name; }
};

class Dog : public Animal {
private:
    std::string breed;
public:
    Dog(std::string name, std::string breed)
        : Animal(std::move(name)), breed(std::move(breed)) {}
    
    std::string speak() const override { return name + " says: Woof!"; }
    std::string fetch() const { return name + " fetches the ball"; }
};
```

```go
// Go: Composition instead of inheritance

// The "base" type (not a base class — just a regular struct)
type Animal struct {
    Name string  // exported — embedding gives access to embedded fields
}

func (a Animal) Move() string { return a.Name + " moves" }

// Interface for the behavior (replaces pure virtual)
type Speaker interface {
    Speak() string
}

// Dog EMBEDS Animal + adds its own fields and methods
type Dog struct {
    Animal                   // embedded — Dog "inherits" Name and Move()
    Breed string
}

// Dog must implement Speak() to satisfy Speaker interface
func (d Dog) Speak() string { return d.Name + " says: Woof!" }
func (d Dog) Fetch() string { return d.Name + " fetches the ball" }

// Cat also embeds Animal
type Cat struct {
    Animal
    Indoor bool
}

func (c Cat) Speak() string { return c.Name + " says: Meow!" }

// Polymorphism via interface:
func makeNoise(s Speaker) string { return s.Speak() }

// Usage:
dog := Dog{Animal: Animal{Name: "Rex"}, Breed: "Lab"}
cat := Cat{Animal: Animal{Name: "Whiskers"}, Indoor: true}

fmt.Println(dog.Speak())    // "Rex says: Woof!"
fmt.Println(dog.Move())     // "Rex moves" (promoted from Animal)
fmt.Println(dog.Name)       // "Rex" (promoted field)
fmt.Println(makeNoise(dog)) // polymorphic!
fmt.Println(makeNoise(cat)) // also polymorphic!

// Key difference: Dog is NOT an Animal in Go's type system
var a Animal = dog.Animal    // explicit: must access embedded type
// var a Animal = dog        // COMPILE ERROR
```

### Multiple Inheritance → Multiple Embedding

```cpp
// C++ Multiple Inheritance
class Flyable {
public:
    virtual std::string fly() const = 0;
};

class Swimmable {
public:
    virtual std::string swim() const = 0;
};

class Duck : public Animal, public Flyable, public Swimmable {
public:
    std::string speak() const override { return "Quack!"; }
    std::string fly() const override { return name + " flies"; }
    std::string swim() const override { return name + " swims"; }
};
```

```go
// Go: Embed multiple types + implement multiple interfaces

type Flyer interface { Fly() string }
type Swimmer interface { Swim() string }
type AnimalBehavior interface {
    Speaker
    Flyer
    Swimmer
}

type Duck struct {
    Animal  // embed Animal for Name and Move() promotion
}

func (d Duck) Speak() string { return "Quack!" }
func (d Duck) Fly() string   { return d.Name + " flies" }
func (d Duck) Swim() string  { return d.Name + " swims" }

// Duck satisfies ALL three interfaces automatically
var _ Speaker = Duck{}
var _ Flyer = Duck{}
var _ Swimmer = Duck{}
var _ AnimalBehavior = Duck{}
```

---

## Chapter 4: RAII → defer

### Resource Management

```cpp
// C++: RAII — resource acquired in constructor, released in destructor
class FileRAII {
    FILE* file;
public:
    explicit FileRAII(const char* path) : file(fopen(path, "r")) {
        if (!file) throw std::runtime_error("Cannot open file");
    }
    ~FileRAII() { if (file) fclose(file); }  // Destructor guarantees cleanup
    
    std::string read() { /* ... */ }
};

void processFile(const std::string& path) {
    FileRAII file(path);    // acquires
    // ... use file
}   // FileRAII destructor runs here — even if exception thrown!
```

```go
// Go: defer — explicit but guaranteed
func processFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("processFile: %w", err)
    }
    defer file.Close()  // guaranteed cleanup — even if panic!
    
    // ... use file
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        fmt.Println(scanner.Text())
    }
    return scanner.Err()
}

// More complex RAII-like pattern
type Resource struct {
    conn *sql.DB
    mu   sync.Mutex
}

func (r *Resource) AcquireLock() func() {
    r.mu.Lock()
    return r.mu.Unlock  // returns the unlock function
}

// Usage:
release := resource.AcquireLock()
defer release()
// ... critical section
```

---

## Chapter 5: Smart Pointers → GC + Careful Design

```cpp
// C++: Ownership with smart pointers
std::unique_ptr<User> createUser(int id) {
    return std::make_unique<User>(id, "Alice");  // sole ownership
}

std::shared_ptr<Config> getConfig() {
    static auto config = std::make_shared<Config>();  // shared ownership
    return config;  // ref count incremented
}

void process(std::weak_ptr<User> weakUser) {
    if (auto user = weakUser.lock()) {  // safe access
        user->process();
    }
    // User might be deleted after lock() expires
}
```

```go
// Go: GC handles memory — no ownership semantics needed
// Use *T (pointer) when you need to share/modify
// Use T (value) when you want a copy

func createUser(id int) *User {
    return &User{ID: id, Name: "Alice"}  // GC handles lifetime
    // Caller can use it as long as it holds the reference
}

// "Unique ownership" pattern: Use direct access, not copy
func withUser(id int, fn func(*User)) {
    user := &User{ID: id}
    fn(user)
    // user is GC'd when no references remain
}

// "Shared ownership": Just keep multiple references
config := &Config{MaxConns: 100}
server1 := &Server{config: config}
server2 := &Server{config: config}  // both share the SAME config
// GC won't collect config while either server1 or server2 references it

// "Weak reference" pattern: Don't hold a reference when idle
type Pool struct {
    mu    sync.Mutex
    items []*Item 
}

// Items in pool are only weakly reachable when in the pool
// GC-friendly: sync.Pool does this automatically!
var pool = sync.Pool{New: func() interface{} { return new(Item) }}
```

---

## Chapter 6: Templates → Generics

```cpp
// C++ Templates — very powerful, complex
template<typename T>
T max(T a, T b) {
    return a > b ? a : b;
}

template<typename Container, typename Predicate>
Container filter(const Container& c, Predicate pred) {
    Container result;
    std::copy_if(c.begin(), c.end(), std::back_inserter(result), pred);
    return result;
}

// Template specialization
template<>
std::string max<std::string>(std::string a, std::string b) {
    return a.length() > b.length() ? a : b;
}
```

```go
// Go Generics — simpler, less powerful
func Max[T constraints.Ordered](a, b T) T {
    if a > b { return a }
    return b
}

func Filter[T any](slice []T, pred func(T) bool) []T {
    var result []T
    for _, v := range slice {
        if pred(v) { result = append(result, v) }
    }
    return result
}

// No specialization in Go — must handle via type switch or interfaces
// If you need different behavior per type, use an interface:
type MaxSelector interface {
    IsGreaterThan(other MaxSelector) bool
}

// Or: use separate functions (Go's preferred approach)
func MaxString(a, b string, byLen bool) string {
    if byLen {
        if len(a) > len(b) { return a }
        return b
    }
    if a > b { return a }
    return b
}
```

---

## Chapter 7: std::thread → goroutine

```cpp
// C++: std::thread
#include <thread>
#include <mutex>
#include <future>

std::mutex mtx;
int counter = 0;

void increment(int n) {
    for (int i = 0; i < n; i++) {
        std::lock_guard<std::mutex> lock(mtx);
        counter++;
    }
}

int main() {
    std::vector<std::thread> threads;
    for (int i = 0; i < 10; i++) {
        threads.emplace_back(increment, 1000);
    }
    for (auto& t : threads) {
        t.join();
    }
    
    // Async task
    auto future = std::async(std::launch::async, []() {
        return compute();
    });
    auto result = future.get();  // blocks until done
}
```

```go
// Go: goroutines
var (
    mu      sync.Mutex
    counter int
)

func increment(n int, wg *sync.WaitGroup) {
    defer wg.Done()
    for i := 0; i < n; i++ {
        mu.Lock()
        counter++
        mu.Unlock()
    }
}

func main() {
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go increment(1000, &wg)
    }
    wg.Wait()
    
    // Async task with channel (like std::future)
    resultCh := make(chan int, 1)
    go func() { resultCh <- compute() }()
    result := <-resultCh  // blocks until done
    
    // More idiomatic: errgroup
    g, _ := errgroup.WithContext(context.Background())
    g.Go(func() error {
        result = compute()
        return nil
    })
    g.Wait()
}
```

---

## Chapter 8: Exceptions → Error Returns

```cpp
// C++: exceptions
double divide(double a, double b) {
    if (b == 0) throw std::invalid_argument("division by zero");
    return a / b;
}

void process(int userId) {
    try {
        auto user = getUser(userId);
        auto order = getOrder(user);
        sendNotification(order);
    } catch (const DatabaseException& e) {
        log("DB error: " + e.what());
    } catch (const std::exception& e) {
        log("Error: " + e.what());
        throw;  // rethrow
    }
}
```

```go
// Go: error values
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

func process(userID int) error {
    user, err := getUser(userID)
    if err != nil {
        // Specific error type check (like catch DatabaseException):
        var dbErr *DatabaseError
        if errors.As(err, &dbErr) {
            log.Printf("DB error: %v", dbErr)
            return nil  // handled
        }
        return fmt.Errorf("process: getUser(%d): %w", userID, err)
    }
    
    order, err := getOrder(user)
    if err != nil {
        return fmt.Errorf("process: getOrder: %w", err)
    }
    
    if err := sendNotification(order); err != nil {
        return fmt.Errorf("process: sendNotification: %w", err)
    }
    
    return nil
}
```

---

## Chapter 9: std::vector / std::map → slice / map

```cpp
// C++: vector
std::vector<int> v = {1, 2, 3, 4, 5};
v.push_back(6);                    // append
v.erase(v.begin() + 2);           // remove at index
v.insert(v.begin() + 1, 99);      // insert at index
std::sort(v.begin(), v.end());     // sort
auto it = std::find(v.begin(), v.end(), 99); // find
v.reserve(100);                    // pre-allocate

// std::map
std::map<std::string, int> m;
m["key"] = 42;
auto it2 = m.find("key");
if (it2 != m.end()) { int val = it2->second; }
m.erase("key");
```

```go
// Go: slice (like vector)
v := []int{1, 2, 3, 4, 5}
v = append(v, 6)                    // push_back → append
v = append(v[:2], v[3:]...)         // erase at index 2
v = append(v[:1], append([]int{99}, v[1:]...)...) // insert at index 1
sort.Ints(v)                        // sort

// Find (no built-in find — use slices.Index in Go 1.21+):
import "slices"
idx := slices.Index(v, 99)          // -1 if not found

v = make([]int, 0, 100)             // reserve

// Go: map (like unordered_map)
m := make(map[string]int)
m["key"] = 42
val, ok := m["key"]                 // find: comma-ok idiom
if ok { fmt.Println(val) }
delete(m, "key")                    // erase

// Iteration (random order unlike C++ std::map)
for k, v := range m { fmt.Println(k, v) }
```

---

## Chapter 10: namespaces → packages

```cpp
// C++: namespaces
namespace database {
namespace postgres {
    class Connection { ... };
    Connection connect(const std::string& dsn);
}
}

// Usage:
database::postgres::Connection conn = database::postgres::connect(dsn);
// or with 'using':
using namespace database::postgres;
Connection conn = connect(dsn);
```

```go
// Go: packages
// File: database/postgres/connection.go
package postgres  // package name is the LAST element of the import path

type Connection struct { ... }
func Connect(dsn string) (*Connection, error) { ... }

// Usage:
import "github.com/myapp/database/postgres"
conn, err := postgres.Connect(dsn)
// The package qualifier (postgres.) is always required — no 'using namespace' equivalent

// Package alias (for conflicts or verbosity):
import pg "github.com/myapp/database/postgres"
conn, err := pg.Connect(dsn)
```

---

## Chapter 11: Complete Migration Test

```go
package migration_test

import (
    "errors"
    "testing"
    "sync"
)

// ==================== CLASS MIGRATION TEST ====================

type BankAccount struct {
    owner   string
    balance float64
    mu      sync.Mutex  // needed if used concurrently
}

func NewBankAccount(owner string, balance float64) (*BankAccount, error) {
    if owner == "" { return nil, errors.New("empty owner") }
    if balance < 0 { return nil, errors.New("negative balance") }
    return &BankAccount{owner: owner, balance: balance}, nil
}

func (a *BankAccount) Deposit(amount float64) error {
    if amount <= 0 { return errors.New("amount must be positive") }
    a.mu.Lock()
    defer a.mu.Unlock()
    a.balance += amount
    return nil
}

func (a *BankAccount) Withdraw(amount float64) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    if amount <= 0 { return errors.New("amount must be positive") }
    if amount > a.balance { return errors.New("insufficient funds") }
    a.balance -= amount
    return nil
}

func (a *BankAccount) Balance() float64 {
    a.mu.Lock()
    defer a.mu.Unlock()
    return a.balance
}

func TestBankAccount(t *testing.T) {
    // Positive cases
    t.Run("create valid account", func(t *testing.T) {
        acc, err := NewBankAccount("Alice", 100.0)
        if err != nil { t.Fatalf("unexpected error: %v", err) }
        if acc.Balance() != 100.0 { t.Errorf("balance = %f, want 100", acc.Balance()) }
    })
    
    t.Run("deposit increases balance", func(t *testing.T) {
        acc, _ := NewBankAccount("Bob", 50.0)
        if err := acc.Deposit(25.0); err != nil { t.Fatalf("deposit: %v", err) }
        if acc.Balance() != 75.0 { t.Errorf("balance = %f, want 75", acc.Balance()) }
    })
    
    t.Run("withdraw decreases balance", func(t *testing.T) {
        acc, _ := NewBankAccount("Carol", 100.0)
        if err := acc.Withdraw(30.0); err != nil { t.Fatalf("withdraw: %v", err) }
        if acc.Balance() != 70.0 { t.Errorf("balance = %f, want 70", acc.Balance()) }
    })
    
    // Negative cases
    t.Run("cannot create with empty owner", func(t *testing.T) {
        _, err := NewBankAccount("", 100.0)
        if err == nil { t.Error("expected error for empty owner") }
    })
    
    t.Run("cannot create with negative balance", func(t *testing.T) {
        _, err := NewBankAccount("Dave", -10.0)
        if err == nil { t.Error("expected error for negative balance") }
    })
    
    t.Run("cannot deposit negative amount", func(t *testing.T) {
        acc, _ := NewBankAccount("Eve", 100.0)
        if err := acc.Deposit(-5.0); err == nil { t.Error("expected error") }
    })
    
    t.Run("cannot withdraw more than balance", func(t *testing.T) {
        acc, _ := NewBankAccount("Frank", 100.0)
        if err := acc.Withdraw(200.0); err == nil { t.Error("expected insufficient funds error") }
        if acc.Balance() != 100.0 { t.Error("balance should be unchanged on failed withdraw") }
    })
    
    // Concurrent access test
    t.Run("concurrent access is safe", func(t *testing.T) {
        acc, _ := NewBankAccount("Grace", 0.0)
        var wg sync.WaitGroup
        for i := 0; i < 100; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                acc.Deposit(1.0)
            }()
        }
        wg.Wait()
        if acc.Balance() != 100.0 {
            t.Errorf("concurrent deposits: got %f, want 100", acc.Balance())
        }
    })
}

// ==================== INTERFACE MIGRATION TEST ====================

type Speaker interface { Speak() string }

type Dog struct{ Name string }
func (d Dog) Speak() string { return d.Name + ": Woof!" }

type Cat struct{ Name string }
func (c Cat) Speak() string { return c.Name + ": Meow!" }

func makeAllSpeak(speakers []Speaker) []string {
    results := make([]string, len(speakers))
    for i, s := range speakers { results[i] = s.Speak() }
    return results
}

func TestPolymorphism(t *testing.T) {
    animals := []Speaker{
        Dog{Name: "Rex"},
        Cat{Name: "Whiskers"},
        Dog{Name: "Buddy"},
    }
    
    results := makeAllSpeak(animals)
    
    expected := []string{
        "Rex: Woof!",
        "Whiskers: Meow!",
        "Buddy: Woof!",
    }
    
    for i, r := range results {
        if r != expected[i] {
            t.Errorf("results[%d] = %q, want %q", i, r, expected[i])
        }
    }
}
```

---

**Summary of Part 18:**
- C++ class → Go struct + methods + constructor function (`NewXxx`)
- C++ inheritance → Go embedding (HAS-A) + interfaces (IS-A for behavior)
- C++ virtual functions → Go interfaces (implicit implementation)
- C++ RAII → Go `defer` (explicit but guaranteed even on panic)
- C++ smart pointers → Go GC (just use `*T`, GC manages lifetime)
- C++ templates → Go Generics (simpler, no specialization)
- C++ `std::thread` → Go goroutines + `sync.WaitGroup` or `errgroup`
- C++ exceptions → Go `(result, error)` return values
- C++ `std::vector` → Go slice (built-in, `append` for push_back)
- C++ `std::map` → Go map (built-in, comma-ok for lookup)
- C++ namespaces → Go packages (always qualified — no `using namespace`)
- The biggest shift: **everything is explicit in Go** — errors, memory patterns, interfaces
