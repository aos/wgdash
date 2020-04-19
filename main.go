package main

import "net/http"

func main() {
	s := NewServer()
	s.Routes()

	http.ListenAndServe(":3100", s)
}
