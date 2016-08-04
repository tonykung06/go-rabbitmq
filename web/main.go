package main

import (
	"net/http"

	"github.com/go-rabbitmq/web/controller"
)

func main() {
	controller.Initialize()
	http.ListenAndServe(":3000", nil)
}
