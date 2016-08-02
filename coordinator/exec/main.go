package main

import (
	"fmt"

	"github.com/go-rabbitmq/coordinator"
)

func main() {
	ql := coordinator.NewQueueListener()
	go ql.ListenForNewSouce()
	fmt.Println("a coordinator is started")
	var a string
	fmt.Scanln(&a)
}
