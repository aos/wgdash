package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var port = os.Getenv("PORT")

func main() {
	if port == "" {
		port = "3100"
	}

	s := NewServer()
	s.Routes()
	ip, _ := s.GetPublicIPAddr()
	fmt.Println(ip)

	log.Printf("Starting server on port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, s))
}
