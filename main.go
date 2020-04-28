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
	s.Routes()

	LoadWriteServerConfig()

	log.Printf("Starting server on port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, s))
}
