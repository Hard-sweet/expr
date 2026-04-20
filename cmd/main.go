package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(3)
	defer wg.Wait()
	chA := make(chan struct{}, 1)
	chB := make(chan struct{}, 1)
	chC := make(chan struct{}, 1)
	chA <- struct{}{}
	go func() {
		defer wg.Done()
		<-chA
		fmt.Println("a")
		chB <- struct{}{}
	}()
	go func() {
		defer wg.Done()
		<-chB
		fmt.Println("b")
		chC <- struct{}{}
	}()
	go func() {
		defer wg.Done()
		<-chC
		fmt.Println("c")
	}()
}
