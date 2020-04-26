package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aos/wgdash/wgcli"
)

var port = os.Getenv("PORT")

func main() {
	if port == "" {
		port = "3100"
	}

	s := NewWgServer()
	s.Routes()

	m := wgcli.GenerateKeyPair()
	fmt.Println(m)

	log.Printf("Starting server on port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, s))
}
