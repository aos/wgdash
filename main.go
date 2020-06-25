package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var port = os.Getenv("PORT")

func main() {
	if err := run(os.Args, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if port == "" {
		port = "3100"
	}

	s := NewWgServer()

	log.Printf("Started server on port: %s...\n", port)
	if err := http.ListenAndServe(":"+port, s); err != nil {
		return err
	}
	return nil
}
