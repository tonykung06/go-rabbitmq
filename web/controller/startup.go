package controller

import "net/http"

var wsc = newWebsocketController()

func Initialize() {
	registerRoutes()
	registerFileServers()
}

func registerRoutes() {
	http.HandleFunc("/ws", wsc.handleMessage)
}

func registerFileServers() {
	http.Handle("/public/", http.FileServer(http.Dir("assets")))
	http.Handle("/public/lib/", http.StripPrefix("/public/lib/", http.FileServer(http.Dir("node_modules"))))
}
