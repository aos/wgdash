package main

import (
	"net/http"
	"text/template"
)

// Client is any client device added that connects to the wg server
type Client struct {
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
	Clients      []Client

	mux *http.ServeMux
}

// NewWgServer instantiates the server
func NewWgServer() *WgServer {
	wgServer := LoadServerConfig()
	wgServer.mux = http.NewServeMux()
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
