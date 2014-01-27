package main

import (
	"fmt"
	"time"

	"github.com/xiangli-cmu/raft"
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
		Key:   "testKey",
		Value: "testValue",
	}

	for i := 0; i < num; i++ {
		_, err := s.Do(command)
		if err != nil {
			fmt.Println(err)
			return
			done <- true
		}
	}
	fmt.Println("finished!")
	done <- true
}
