package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/aos/wgdash/wgcli"
)

// Peer is any client device added that connects to the wg server
type Peer struct {
	Active     bool
	ID         int
	Name       string
	PrivateKey string
	PublicKey  string
	VirtualIP  string
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
	Peers        []Peer

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
			return
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
		switch urlParts[2] {
		case "peers":
			s.handlePeers(w, r)
		}
	})
}

func (s *WgServer) handlePeers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		urlParts := strings.Split(r.URL.Path, "/")
		if len(urlParts) < 4 {
			http.Error(w, "Did not specify peer ID", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(urlParts[3])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, p := range s.Peers {
			if p.ID == id {
				_, ipNet, err := net.ParseCIDR(s.VirtualIP + "/" + s.CIDR)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				tmpl := template.Must(template.ParseFiles("templates/peer.conf.tmpl"))
				tmpl.Execute(w, struct {
					VirtualIP       string
					PrivateKey      string
					ServerPublicKey string
					PublicIP        string
					Port            string
					AllowedIPs      string
				}{
					VirtualIP:       p.VirtualIP,
					PrivateKey:      p.PrivateKey,
					ServerPublicKey: s.PublicKey,
					PublicIP:        s.PublicIP,
					Port:            s.Port,
					AllowedIPs:      ipNet.String(),
				})

				return
			}
		}

		http.Error(w, fmt.Sprintf("Peer %d not found", id), http.StatusNotFound)

	case "POST":
		var p Peer
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		keys, err := wgcli.GenerateKeyPair()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		peerIP, err := s.nextAvailableIP(p.VirtualIP)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		p.VirtualIP = peerIP
		p.PrivateKey = keys["privateKey"]
		p.PublicKey = keys["publicKey"]
		p.ID = len(s.Peers) + 1

		// We want to make sure that wg is actually running here
		err = wgcli.AddPeer(p.PublicKey, p.VirtualIP)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.Peers = append(s.Peers, p)
		err = s.saveBothConfigs()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		js, err := json.Marshal(struct {
			StatusCode    int
			Message       string
			PeerID        int
			PeerVirtualIP string
		}{
			200,
			"Peer added successfully",
			p.ID,
			p.VirtualIP,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
	return
}
