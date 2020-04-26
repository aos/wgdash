package main

import (
	"log"
	"os"
)

func (s *WgServer) CheckServerConfig() {
	_, err := os.Open(s.ServerConfigPath)
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
