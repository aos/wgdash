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
		}
	}
}
