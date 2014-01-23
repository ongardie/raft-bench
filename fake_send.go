package main

import (
	"fmt"
	"time"

	"github.com/goraft/raft"
)

func send(s raft.Server, num int) {
	fmt.Println("start sending requests...")
	start := time.Now()

	done := make(chan bool, 10)

	for i := 0; i < 1000; i++ {
		go currentSend(s, num, done)
	}

	for i := 0; i < 1000; i++ {
		<-done
	}

	fmt.Println(time.Now().Sub(start))
}

func currentSend(s raft.Server, num int, done chan bool) {
	command := &SetCommand{
		key:   "testKey",
		value: "testValue",
	}

	for i := 0; i < num; i++ {
		_, err := s.Do(command)
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("finished!")
	done <- true
}
