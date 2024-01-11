package main

import (
	"fmt"
	"sync"
)

var limit = make(chan int, 3)

func main() {
	wg := sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			limit <- 1
			fmt.Println("send")
			<-limit
			fmt.Println("receive")
			wg.Done()
		}()
	}
	wg.Wait()
}
