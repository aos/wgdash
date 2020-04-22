package main

import (
	"net"
	"net/http"
	"text/template"
)

// Peer is any client device added and connects to the wg server
type Peer struct {
	Active    bool
	Name      string
	PublicKey string
	QRcode    string
	VirtualIP string
}

// Server holds all configuration of our server, including the router
type Server struct {
	PublicIP         string
	Port             string
	VirtualIP        string
	CIDR             string
	PublicKey        string
	DNS              string
	WgConfigPath     string
	ServerConfigPath string
	Peers            []Peer

	mux *http.ServeMux
}

// NewServer instantiates the server
func NewServer() *Server {
	// parse our configuration file
	return &Server{mux: http.NewServeMux()}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) renderTemplatePage(tmplFname string, data interface{}) http.Handler {
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

func (s *Server) GetPublicIPAddr() (string, error) {
	// ip -4 a show wlp2s0 | grep -oP '(?<=inet\s)\d+(\.\d+){3}'
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String(), nil
}
