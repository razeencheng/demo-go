package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {

	var (
		start bool
		e     bool
	)

	flag.BoolVar(&start, "s", false, "start output")
	flag.BoolVar(&e, "e", false, "output err")
	flag.Parse()

	if start {
		for i := 5; i > 0; i-- {
			fmt.Fprintln(os.Stdout, "test cmd output", i)
			time.Sleep(1 * time.Second)
		}
	}

	if e {
		fmt.Fprintln(os.Stderr, "a err occur")
		os.Exit(1)
	}

	fmt.Fprintln(os.Stdout, "test cmd stdout")
}
