package main

import (
	"fmt"

	"github.com/go-rabbitmq/coordinator"
)

var dc *coordinator.DatabaseConsumer

func main() {
	ea := coordinator.NewEventAggregator()
	coordinator.NewDatabaseConsumer(ea)
	ql := coordinator.NewQueueListener(ea)
	go ql.ListenForNewSouce()
	fmt.Println("a coordinator is started")
	var a string
	fmt.Scanln(&a)
}
