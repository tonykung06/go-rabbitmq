package main

import (
	"fmt"

	"github.com/go-rabbitmq/coordinator"
)

func main() {
	ql := coordinator.NewQueueListener()
	go ql.ListenForNewSouce()
	var a string
	fmt.Scanln(&a)
}
