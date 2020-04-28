package main

import (
	"net/http"
	"text/template"
)

// Client is any client device added and connects to the wg server
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
	VirtualIP    string
	CIDR         string
	DNS          string
	PublicKey    string
	WgConfigPath string
	Clients      []Client

	mux *http.ServeMux
}

// NewWgServer instantiates the server
func NewWgServer() *WgServer {
	// parse our config file
	return &WgServer{mux: http.NewServeMux()}
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
