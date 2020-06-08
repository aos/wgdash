package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
)

// Peer is any client device added that connects to the wg server
type Peer struct {
	Active    bool
	Name      string
	PublicKey string
	QRcode    string
	VirtualIP string
}

// WgServer holds all configuration of our server, including the router
type WgServer struct {
	PublicIP     string
	Port         string
	VirtualIP    string
	CIDR         string
	DNS          string
	PublicKey    string
	PrivateKey   string
	WgConfigPath string
	Peers      []Peer

	mux *http.ServeMux
}

// NewWgServer instantiates the server
func NewWgServer() *WgServer {
	wgServer := LoadServerConfig()
	wgServer.mux = http.NewServeMux()
	wgServer.Routes()
	return wgServer
}

func (s *WgServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *WgServer) renderTemplatePage(tmplFname string, data interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("templates/" + tmplFname)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = t.Execute(w, data)
		if err != nil {
			panic(err)
		}
	})
}

func (s *WgServer) handleAPI() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/api/" {
			http.Error(w, "Couldn't find anything here :(", http.StatusNotFound)
			return
		}

		urlParts := strings.Split(r.URL.Path, "/")
		fmt.Printf("URL parts: %# v\n", urlParts)
		switch urlParts[2] {
		case "peers":
			s.handlePeers(w, r)
		}
	})
}

func (s *WgServer) handlePeers(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%# v\n", r.Body)

	switch r.Method {
	case "GET":
		http.Error(w, "Not implemented", http.StatusNotImplemented)

	case "POST":
		var c Peer
		err := json.NewDecoder(r.Body).Decode(&c)
		fmt.Printf("client: %# v\n", c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}
