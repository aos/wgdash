package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	htmlTemp "html/template"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/aos/wgdash/wgcli"
	"github.com/skip2/go-qrcode"
)

// Peer is any client device added that connects to the wg server
type Peer struct {
	Active     bool
	ID         int
	Name       string
	PrivateKey string
	PublicKey  string
	VirtualIP  string
	KeepAlive  int
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
	Active       bool
	Peers        []Peer

	mux *http.ServeMux
}

// NewWgServer instantiates the server
func NewWgServer() *WgServer {
	wgServer := LoadServerConfig()
	wgServer.mux = http.NewServeMux()
	wgServer.Routes()
	wgServer.ActivateServer()

	return wgServer
}

func (s *WgServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *WgServer) renderTemplatePage(tmplFname string, data interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, err := htmlTemp.ParseFiles("templates/base.html.tmpl", "templates/"+tmplFname)
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
		switch urlParts[2] {
		case "peers":
			s.handlePeersAPI(w, r)
		}
	})
}

func (s *WgServer) handlePeersAPI(w http.ResponseWriter, r *http.Request) {
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

				params := r.URL.Query()
				qrCode := false
				if v, ok := params["qr"]; ok && v[0] == "true" {
					qrCode = true
				}

				var buf bytes.Buffer
				tmpl := template.Must(template.ParseFiles("templates/peer.conf.tmpl"))
				tmpl.Execute(&buf, struct {
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

				if qrCode {
					qr, err := qrcode.Encode(buf.String(), qrcode.Medium, 256)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						log.Printf("error (get_peer - generate qr code): %s\n", err.Error())
					}
					encoder := base64.NewEncoder(base64.StdEncoding, w)
					encoder.Write(qr)
				} else {
					w.Header().Set("Content-Disposition", "attachment; filename=wg0.conf")
					w.Header().Set("Content-Type", "application/octet-stream")
					w.Write(buf.Bytes())
				}

				return
			}
		}

		http.Error(w, fmt.Sprintf("Peer %d not found", id), http.StatusNotFound)

	case "POST":
		var p Peer
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Printf("error (add_peer - decode JSON): %s\n", err.Error())
			return
		}
		defer r.Body.Close()

		keys, err := wgcli.GenerateKeyPair()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("error (add_peer - generate key pair): %s\n", err.Error())
			return
		}

		peerIP, err := s.nextAvailableIP(p.VirtualIP)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Printf("error: (add_peer - available IP): %s\n", err.Error())
			return
		}

		p.VirtualIP = peerIP
		p.PrivateKey = keys["privateKey"]
		p.PublicKey = keys["publicKey"]
		if len(s.Peers) <= 0 {
			p.ID = len(s.Peers) + 1
		} else {
			p.ID = s.Peers[len(s.Peers)-1].ID + 1 // always increment IDs
		}

		// We want to make sure that wg is actually running here
		if s.Active {
			err = wgcli.AddPeer(p.PublicKey, p.VirtualIP, p.KeepAlive)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Printf("error: (add_peer - wg add peer): %s\n", err.Error())
				return
			}
		}

		s.Peers = append(s.Peers, p)
		err = s.saveBothConfigs()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("error: (add_peer - save configs): %s\n", err.Error())
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

		log.Printf("success (add_peer): %s\n", js)
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)

	case "DELETE":
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

		for i, p := range s.Peers {
			if p.ID == id {
				if s.Active {
					err = wgcli.RemovePeer(p.PublicKey)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}
				s.Peers = append(s.Peers[:i], s.Peers[i+1:]...)

				err = s.saveBothConfigs()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				return
			}
		}

		http.Error(w, fmt.Sprintf("Peer %d not found", id), http.StatusNotFound)

	case "PUT":
		http.Error(w, "Not implemented yet", http.StatusNotImplemented)
	}
}
