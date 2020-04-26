package main

import (
	"net"
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
	PublicIP         string
	Port             string
	VirtualIP        string
	CIDR             string
	PublicKey        string
	DNS              string
	WgConfigPath     string
	ServerConfigPath string
	Clients          []Client

	mux *http.ServeMux
}

// NewWgServer instantiates the server
func NewWgServer() *WgServer {
	// parse our configuration file
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

func (s *WgServer) addPublicIPAddr() error {
	// Alternative: ip -4 a show wlp2s0 | grep -oP '(?<=inet\s)\d+(\.\d+){3}'
	// Note: this does not make an actual connection and can be used
	// offline
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		return err
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr)
	s.PublicIP = addr.IP.String()
	return nil
}
