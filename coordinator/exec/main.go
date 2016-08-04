package main

import (
	"fmt"

	"github.com/go-rabbitmq/coordinator"
)

var dc *coordinator.DatabaseConsumer
var wc *coordinator.WebappConsumer

func main() {
	ea := coordinator.NewEventAggregator()
	coordinator.NewDatabaseConsumer(ea)
	coordinator.NewWebappConsumer(ea)
	ql := coordinator.NewQueueListener(ea)
	go ql.ListenForNewSouce()
	fmt.Println("a coordinator is started")
	var a string
	fmt.Scanln(&a)
}
