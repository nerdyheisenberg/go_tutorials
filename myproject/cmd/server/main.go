package main

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"
)

// typedef type
type Celsius float64    // new type not a alias
type Fahrenheit float64 // new type not a alias
type MyInt = int        // alias

func (c Celsius) ToFahrenheit() Fahrenheit {
	return Fahrenheit(c*9/5 + 32)
}

// interface polymorphism
type Reader interface {
	Read() string
}

type Writer interface {
	Write(content string)
}

type ReaderWriter interface {
	Reader
	Writer
}

type Document struct {
	Title   string
	Content string
}

type SecureNote struct {
	Document
	Password string
}

func (d *Document) Read() string {
	return fmt.Sprintf("Title :[%s], Content:[%s]", d.Title, d.Content)
}

func (sn *SecureNote) Write(content string) {
	sn.Document.Content = content
	fmt.Println("Secure content updated")
}

// error type
func add(a, b int) (int, error) {
	if a == 0 && b == 0 {
		return 0, errors.New("a and b cannot be zero")
	}
	return a + b, nil
}

/* Named return

func add (a , b int) (result int, err error){
	if a == 0 && b == 0 {
		err =  errors.New("a and b cannot be zero")
		return
	}
	result = a+b
	return
}
*/

// variadic function
func suma(num ...int) int {
	total := 0
	for _, n := range num {
		total += n
	}
	return total
}

// closure
func closure(fn func(string) string) func(string) string {
	return func(name string) string {
		fmt.Println("Input : ", name)
		surname := fn(name)
		fmt.Println("Output : ", surname)
		return surname
	}
}

func close_capture() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

// panic and recover
func safeDivide(a, b int) (result int, err error) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	// This will panic if b == 0
	return a / b, nil
}

// struct
type User struct {
	Name string
}

// 2. The function as requested
func createUser() *User {
	u := User{Name: "Rohit"} // Escapes to heap automatically
	return &u
}

//Composition:

type Animal struct {
	Name string
}

func (a Animal) Speak() string {
	return a.Name + " creates a sound"
}

type Dog struct {
	Animal
	Breed string
}

func (d Dog) Speak() string {
	return d.Name + " woofs"
}

// Interfaces

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Circle struct {
	Radius float64
}

type Rectangle struct {
	Width, Height float64
}

func (c Circle) Area() float64 {
	return 3.14 * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * 3.14 * c.Radius
}

func (c Rectangle) Area() float64 {
	return c.Width * c.Height
}

func (c Rectangle) Perimeter() float64 {
	return 2 * (c.Width + c.Height)
}

func print(s Shape) {
	fmt.Println("Area = ", s.Area(), " Perimeter = ", s.Perimeter())
}

func describe(i any) string { // can use i interface{} too
	switch v := i.(type) {
	case int:
		return fmt.Sprintf("integer: %d", v)
	case string:
		return fmt.Sprintf("string: %q", v)
	case Circle:
		return fmt.Sprintf("circle with radius: %.2f", v.Radius)
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("unknown type: %T", v)
	}
}

// generics
func Max[T int | float64 | string](a, b T) T {
	if a < b {
		return b
	}
	return a
}

// type constraint on generics with interface, only case when member variables are allowed
//otherwise always functions

type Normal interface {
	int | float32 | int32 | int64 | int8 | int16 | float64
}

func Mas[T Normal](a, b T) T {
	// whatever
	return a
}

//stack implementation using generics

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(t T) {
	s.items = append(s.items, t)
}

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var temp T
		return temp, false
	}
	temp := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return temp, true
}

// producer consumer
var waitn sync.WaitGroup

func producer(ch chan<- int) {
	defer waitn.Done()
	for i := 0; i < 10; i++ {
		ch <- i
	}
	close(ch)
}

func consumer(ch <-chan int) {
	defer waitn.Done()
	for cha := range ch {
		fmt.Println("Consumer : ", cha)
	}
}

func consumer1(ch <-chan int) {
	defer waitn.Done()
	for {
		if channnel, ok := <-ch; ok {
			fmt.Println("Consumer1 : ", channnel)
		} else {
			break
		}
	}
}

// worker problem
func worker(id int, source <-chan int, result chan<- int, s *sync.WaitGroup) {
	defer s.Done()
	for channels := range source {
		result <- channels * 2
	}
}

type Node struct {
	Data int
	Next *Node
}

type LinkedList struct {
	Head *Node
}

func (n *LinkedList) insert(data int) {
	temp := &Node{data, nil}
	if n.Head == nil {
		n.Head = temp
		return
	}
	rec := n.Head
	for rec.Next != nil {
		rec = rec.Next
	}
	rec.Next = temp
}

func (n *LinkedList) print() {
	rec := n.Head
	for rec != nil {
		fmt.Println(rec.Data)
		rec = rec.Next
	}
}

//fmt.Stringer

type Stringer interface { // you can name it anything , Stringer is not compulsory only thing is String() function
	String() string
}

type User_S struct {
	Name string
	Age  int
}

func (u User_S) String() string {
	return fmt.Sprintf("Name : %s, Age : %d", u.Name, u.Age)
}

//context

func context_check(ctx context.Context) {
	for {
		select {
		case <-time.After(500 * time.Millisecond):
			fmt.Println("Context after 500 milliseconds")
		case <-ctx.Done():
			fmt.Println("Timed out ", ctx.Err())
			return
		}
	}
}

func main() {

	fmt.Println("----------------------------------------")
	fmt.Println("Variables section")
	fmt.Println("----------------------------------------")

	//var a int = 10.8 // failure Statically typed , types checked at compile time.
	var _ string = "Rohit Gupta" //unused but okay
	var distance int64 = 5000

	//distance := 5000 //already declare no new variable then error will come, suppose distance was already declared before
	// also always do it inside a function not at package level, atleast one variable on left must be new. Creates a new scope than outer variable

	//_ := 1000 // error : no new variables on left side of :=

	distance, vector := 10, 20 // atleast one should be new, vector is new there, so okay
	distance = distance + 9000

	fmt.Println(distance)
	nww, _ := 11, 22 //nww is new here.

	/*	var (
		data := 123
		nata := "sola"
	)*/ // cannot use := inside

	var (
		_    = 123
		_    = "sola"
		data = 10 // or data int = 10
	)

	{
		//shadowing concept, inner circle no outer interaction
		distance := 1000
		fmt.Printf("Inner distance : %d  DAta : %d NWW  : %d\n", distance, data, nww)
	}
	fmt.Println("----------------------------------------")
	fmt.Println("Variables section")
	fmt.Println("----------------------------------------")
	fmt.Println("----------------------------------------")
	fmt.Println("Types section")
	fmt.Println("----------------------------------------")

	raw_bytes := []byte{0xFF, 0xA1, 0x12, 0x7E}
	pixel := uint8(255)
	character := rune('A')
	das := int(10) // can be converted like this too
	_ = das

	// var a int // uninitialized variable

	//var x int8 = 127
	//x++              // What happens? Overflow! x becomes -128
	//fmt.Println(x) prints -128

	s := "My name"

	//s[0] = 'H' // error immutable data type, single character cannot be changed
	s = "Nooos" // this will still work , cause this is assigning values not single character change

	bytes_data := []byte(s)
	bytes_data[0] = 'H'

	// or convert it to string back
	s = string(bytes_data)

	fmt.Printf("Changes through byte : %s\n", bytes_data)

	{
		//or use string builder

		var sb strings.Builder
		sb.WriteString("Hello")
		sb.WriteString("workd")

		s := sb.String()
		fmt.Print("String ", s)
		fmt.Println("")
	}

	{
		n := 65
		//s := string(n) // shouldn't do compilation warning , implicit not allowed
		s := string(rune(n))
		// or use
		t := fmt.Sprintf("%c", n)
		fmt.Println(s, t)
	}

	var read string = "Rohit"

	read = "NN"

	fmt.Println("Read ", read)

	array_int := []int{1, 2, 3, 4, 6, 7, 8}
	for i := 0; i < len(array_int); i++ {
		fmt.Println(array_int[i])
	}

	for i := range array_int { // this is a slice, so it will print only the index , you have to take 2 arguments
		fmt.Println(i)
	}

	for i := 0; i < len(s); i++ {
		fmt.Printf("index : %d value: %c\n", i, s[i])
	}

	// rune printing
	for i, ch := range s {
		fmt.Printf("byte index %d: %c (U+%04X)\n", i, ch, ch)
	}

	parts := strings.Split("a,b,c,d", ",") // ["a" "b" "c" "d"] // it returns slice of string

	for _, i := range parts { // that's why 2 values will be returned
		fmt.Println(i)
	}

	ex_temp := []string{"a", "b", "c"}

	for _, i := range ex_temp { // that's why 2 values will be returned
		fmt.Println(i)
	}
	fmt.Println("----------------------------------------")
	fmt.Println("Types section")
	fmt.Println("----------------------------------------")
	fmt.Println("----------------------------------------")
	fmt.Println("typedef section & const")
	fmt.Println("----------------------------------------")

	var c Celsius = Celsius(20.5)
	fmt.Println("Temperature in Fahrenheit : ", c.ToFahrenheit())

	{
		var c Celsius = 2.2
		var f Fahrenheit = 3.4
		//f = c // error new types cannot be assigned although they are internally of same type
		f = Fahrenheit(c) // this is possible though no issues with this
		_ = f

		var typess MyInt = 10
		var typessa int = 10
		typessa = typess // holds good cause of alias, internally MyInt is also int only
		_ = typessa
	}

	const x = 10 // wherever copied will take that type

	var my_value float64 = x

	_ = my_value

	//enum in golang

	const (
		_  = iota             // 0 (skip with _)
		KB = 1 << (10 * iota) // 1 << 10 = 1024
		MB                    // 1 << 20 = 1048576
		GB                    // 1 << 30
		TB                    // 1 << 40
		PB                    // 1 << 50
	)

	const a int = KB // if a const is not used there is no error
	var b = KB       // if a variable is not used there will be error , that's why
	_ = b
	fmt.Println("----------------------------------------")
	fmt.Println("typedef section & const")
	fmt.Println("----------------------------------------")
	fmt.Println("----------------------------------------")
	fmt.Println("interface section")
	fmt.Println("----------------------------------------")

	secure_note := &SecureNote{
		Document: Document{Title: "Mybook"},
		Password: "Mindit",
	}

	var rw ReaderWriter = secure_note
	rw.Write("Horrray")
	fmt.Println(rw.Read())

	var i interface{} = "hello" // interface type //talkes any type
	var j = "qweety"            // it will work previously mistaken

	_ = j

	if s, ok := i.(string); ok { // type check with i.(string) ??
		fmt.Println("It's a string:", s)
	}
	fmt.Println("----------------------------------------")
	fmt.Println("interface section")
	fmt.Println("----------------------------------------")

	fmt.Println("----------------------------------------")
	fmt.Println("slices and container section")
	fmt.Println("----------------------------------------")

	st := "Hello, World" // this is collection of characters in slices

	// For character (rune) count:
	fmt.Println(len([]rune(st)))
	// needs [] in conversion otherwise issues , rune is single, but [] represents slices
	// even len(st) will be same
	fmt.Println(st[:1])

	m := map[string]int{"a": 1, "b": 2}
	_ = m

	sum, err := add(1, 2)
	if err != nil {
		fmt.Println("Error : ", err)
	}
	fmt.Println("Sum : ", sum)

	number := suma(1, 2, 3, 4, 5)
	_ = number

	sl_int := []int{1, 2, 3, 4, 5}
	_ = suma(sl_int...)

	value_returned := func(a, b int) int {
		return (a / b)
	}
	_ = value_returned

	vae := func(sl_int []int, fn func(int) int) []int {
		result := make([]int, len(sl_int))
		for i, value := range sl_int {
			result[i] = fn(value)
		}
		return result
	}

	double := vae([]int{1, 2, 3, 4, 5}, func(n int) int {
		return n * 2
	})

	fmt.Println("Doubled ", double)

	/*var fn func(int)int // crashes , declared nil function , called with value boom!!!!

	erry := fn(10)
	*/

	fmt.Println("----------------------------------------")
	fmt.Println("slices and container section")
	fmt.Println("----------------------------------------")

	fmt.Println("----------------------------------------")
	fmt.Println("closure section")
	fmt.Println("----------------------------------------")

	clos := func(name string) string {
		return name + " Gupta"
	}

	res := closure(clos)
	res1 := res("Rohit")

	fmt.Println(res1)

	vard := close_capture()

	resue := vard()

	fmt.Println(resue)
	fmt.Println(vard())
	fmt.Println(vard())
	fmt.Println(vard())
	fmt.Println(vard())

	fmt.Println("----------------------------------------")
	fmt.Println("closure section")
	fmt.Println("----------------------------------------")

	fmt.Println("----------------------------------------")
	fmt.Println("panic and recover section")
	fmt.Println("----------------------------------------")

	_, erro := safeDivide(1, 0)

	if erro != nil {
		fmt.Println(erro)
	}

	fmt.Println("----------------------------------------")
	fmt.Println("panic and recover section")
	fmt.Println("----------------------------------------")

	fmt.Println("----------------------------------------")
	fmt.Println("struct and slice and maps section")
	fmt.Println("----------------------------------------")

	my_name := createUser()

	fmt.Println(my_name.Name) // not -> for value or pointer even , this . will work

	// [...]int{1, 2, 3, 4, 5} array or [5]int{1, 2, 3, 4, 5} both are same and arrays, cannot be changed after this
	// []int{1,2,3,4,5} this is slice and dynamic and can be changed

	//Slices
	empty_slice := []int{}
	var nil_slice []int
	var slice = make([]int, 0, 6)
	slice = append(slice, 1, 2, 3)

	fmt.Println(len(slice)) //len = 3, cap = 6
	fmt.Println(slice)      // [1,2,3]

	bewatch := []int{1, 2, 3, 4, 5}
	slice = append(slice, bewatch...) // ... is important for appending everything afterwards needed only for right side variables

	fmt.Println(slice) // [1,2,3,1,2,3,4,5]

	//if you change any element in original slice and make a sub slice with eg c := slice[1:3]
	// and modify like c[0] = 999, this will modify original too , cause both are sharing same underlying array

	cd := slice[1:3]

	fmt.Println(cd) // [2,3]
	cd[0] = 999     //

	fmt.Println(cd)    // cd is [999,3]
	fmt.Println(slice) // [1,999,3,1,2,3,4,5]
	d := make([]int, 2)
	copy(d, slice) // this is proper copy , if you change something in source, it will not impact
	fmt.Println(d) // [1,999]

	sub1 := cd[1:3:3]
	//sub2 := d[1:3:3] // wrong cannot take 2nd index which is unavailable

	fmt.Println(sub1) // [3,1] copied from slice variable
	//fmt.Println(sub2) // crash no data available cause we copied.

	_ = empty_slice
	_ = nil_slice

	// reverse a slice
	for left, right := 0, len(slice)-1; left < right; left, right = left+1, right-1 {
		slice[left], slice[right] = slice[right], slice[left]
	}

	fmt.Println(slice)

	//sorting a slice
	slices.Sort(slice)

	my_map := make(map[string]int)

	my_map["A"] = 12
	my_map["B"] = 20

	delete(my_map, "A")
	clear(my_map)

	var map_s map[string]int

	if key, value := map_s["a"]; value {
		fmt.Println(key, value)
	}

	type People struct {
		Name string
		Age  int
	}

	map_var := map[int]People{
		1: {"Rohit", 12},
		2: {"Mohit", 23},
	}
	// map_var[1].Name = "Ehr" // cannot be modified , struct is by value type not reference or pointer ?
	temp := map_var[1]
	temp.Name = "Rod"
	temp.Age = 34

	map_var[1] = temp

	fmt.Println(map_var)

	temp_map := map[int]int{
		1: 10,
		2: 15,
	}

	temp_map[1] = 100 // works with basic types

	fmt.Println(temp_map)

	// set data type in go
	set_m := make(map[int]struct{})

	set_m[99] = struct{}{}
	set_m[2] = struct{}{}
	set_m[1] = struct{}{}

	fmt.Println(set_m)

	ani := Dog{
		Animal: Animal{Name: "Don"},
		Breed:  "German Shepherd",
	}

	fmt.Println(ani.Speak())

	circle := Circle{5}
	rectangle := Rectangle{4, 5}

	print(circle)
	print(rectangle)

	for _, shape := range []Shape{circle, rectangle} {
		print(shape)
	}

	var shapu Shape = Circle{5}

	//type assertion
	type_final := shapu.(Circle) // you can also take , two arguments like type_final, ok

	if type_final1, ok := shapu.(Circle); ok {
		fmt.Println(type_final1.Radius)
	}

	fmt.Println(type_final.Area()) // or can call anything like type_final.Radius

	// slice of different type
	var my_type []interface{} // or replace interface{} with any
	my_type = append(my_type, 1, true, 10.32, "My name")

	resulte := describe(circle)

	_ = resulte

	var waitgroup sync.WaitGroup
	const n = 100_000

	for i := 0; i < n; i++ {
		waitgroup.Add(1)
		go func() {
			defer waitgroup.Done()
			time.Sleep(1 * time.Second)
		}()
	}
	waitgroup.Wait()

	channelw := make(chan string)

	go func() {
		channelw <- "Rohit"
	}()

	my_names := <-channelw

	fmt.Println(my_names)

	channeli := make(chan int, 3)
	waitgroup.Add(1)
	go func() {
		defer waitgroup.Done()
		time.Sleep(2 * time.Second)
		channeli <- 1
		channeli <- 2
		channeli <- 3
		channeli <- 4
		//channeli <- 5 issue this will overflow the buffer, which was 3 earlier and one was received in main
		fmt.Println("I was blocked earlier")
	}()

	fmt.Println(<-channeli)
	waitgroup.Wait()

	ch := make(chan int, 30) // buffered chnanel can be closed and data willl be on belt
	waitn.Add(3)
	go producer(ch)
	go consumer(ch) // we desperately wanted to try go routine consume, otherwise we will not need waitgroup
	consumer1(ch)
	waitn.Wait()

	const number_of_worker = 5
	const num_of_jobs = 20

	jobs := make(chan int, num_of_jobs)
	results := make(chan int, num_of_jobs)

	var wait_g sync.WaitGroup

	for i := 0; i < number_of_worker; i++ {
		wait_g.Add(1)
		go worker(i, jobs, results, &wait_g)
	}

	for i := 0; i < num_of_jobs; i++ {
		jobs <- i
	}
	close(jobs)

	go func() {
		wait_g.Wait()
		close(results)
	}()

	for result := range results {
		fmt.Println(result)
	}

	//queue

	queue := list.New()

	queue.PushBack(1)
	queue.PushBack(2)
	queue.PushBack(3)
	queue.PushBack(4)
	queue.PushBack(5)

	for e := queue.Front(); e != nil; e = e.Next() {
		if e.Value.(int)%2 == 0 {
			queue.Remove(e)
		}
	}

	for e := queue.Front(); e != nil; e = e.Next() {
		fmt.Println(e.Value)
	}

	head := LinkedList{}

	head.insert(1)
	head.print()

	user := User_S{
		Name: "Rohit Gupta",
		Age:  30,
	}

	fmt.Println(user)

	var_read_file, err := func(path string) ([]byte, error) {
		id, err := os.Open(path)
		if err != nil {
			fmt.Errorf("File error  %w", err)
		}
		defer id.Close()
		read, err := io.ReadAll(id)
		if err != nil {
			fmt.Errorf("File error  %w", err)
		}
		return read, nil
	}("/home/rohit/ucc/README.md")

	fmt.Println(string(var_read_file))

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	defer cancel()

	go context_check(ctx)
	go context_check(ctx)

	<-ctx.Done()
	time.Sleep(3 * time.Second)

	go fmt.Printf("Go Version : %s\n", runtime.Version())
	fmt.Printf("OS/Arch    : %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("CPU Cores  : %d\n", runtime.NumCPU())
	fmt.Printf("Goroutines : %d\n", runtime.NumGoroutine())
	fmt.Println("Distance : ", distance, "Vector : ", vector)
	fmt.Println("Raw Byte : ", raw_bytes[0], "Pixel ", pixel)
	fmt.Println("Rune : ", character)

}
