package main

import (
	server "base-app/server"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/hello", server.Hello)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
