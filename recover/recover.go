package main

import (
	"fmt"
	"log"
	"time"
)

func main() {

	ch := make(chan int, 10)

	for i := 2; i > 0; i-- {
		go func(i int) {
			defer func() {
				err := recover()
				if err != nil {
					log.Println(err)
				}
			}()
			for val := range ch {
				fmt.Println("---->", val, "Go", i)
				if val%2 == 1 && i == 1 {
					panic("BOOM BOOM")
				}
				time.Sleep(2 * time.Second)
			}
		}(i)
	}

	var i int
	for {
		ch <- i
		time.Sleep(1 * time.Second)
		fmt.Println(i, "<---")
		i++
	}
}
