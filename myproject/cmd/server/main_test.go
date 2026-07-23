package main

import (
	"fmt"
	"testing"
)

func TestAdd(t *testing.T) {
	result := Max(10, 12)
	if result != 22 {
		t.Errorf("Adding wrong data")
		return
	}
	fmt.Println("result :", result)
}

func TestBig(t *testing.T) {
	result := Max(10, 12)
	if result != 12 {
		t.Errorf("Adding wrong data")
		return
	}
	fmt.Println("result :", result)
}
