package main

import (
	"log"
	"net/http"
	"os"
)

var port = os.Getenv("PORT")

func main() {
	if port == "" {
		port = "3100"
	}

	s := NewWgServer()

	log.Printf("Started server on port: %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, s))
}
