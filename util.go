package main

import (
	"log"
	"os"
)

// CheckServerConfig looks for the server config and if it can't find it,
// will make a new one.
func (s *WgServer) CheckServerConfig() {
	_, err := os.Open("config.json")
	if err != nil {
		// File not found -- let's create it
		if os.IsNotExist(err) {
			log.Println("Server config does not exist. Creating...")
			//tmpl := template.Must(template.ParseFiles("templates/server.conf.tmpl"))
			// 1. generate private/public key pair
			// 2. create server config struct
			// 3. write out to template
		}
	}
}
