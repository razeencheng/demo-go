package main

import (
	"os"
	"os/signal"
	"runtime/trace"
	"sync/atomic"
	"syscall"
	"time"
)

func main(){

	trace.Start(os.Stderr)
	defer trace.Stop()

	sigs := make(chan os.Signal,1)
	signal.Notify(sigs,syscall.SIGINT,syscall.SIGTERM)

	//time.AfterFunc(time.Second, func() {
	//	println("done")
	//})


	var num int64

	for i:=0; i< 1e3 ; i++ {
		time.AfterFunc(time.Second, func() {
			atomic.AddInt64(&num,1)
		})
	}

	t:= 0
	for i:=0;i<1e10; i++ {
		t ++
	}
	_ = t

	<- sigs

	// println(num,"timers created,",t,"iterations done")
}
