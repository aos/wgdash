package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/aos/wgdash/wgcli"
	"github.com/apparentlymart/go-cidr/cidr"
)

var fileName = "server_config.json"

// LoadServerConfig looks for the server config and if it can't find it,
// will make a new one and return the struct
func LoadServerConfig() *WgServer {
	f, err := ioutil.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return CreateServerConfig()
		}
		log.Fatalf("Unable to open server config file: %s", err)
	}

	var wgServer WgServer
	err = json.Unmarshal(f, &wgServer)
	if err != nil {
		log.Fatalf("Incorrectly formated JSON server config: %s", err)
	}

	// TODO: this could potentially break if we find a template that does
	// not have port, or other values filled out. Need to initialize with
	// defaults
	if wgServer.PublicKey == "" || wgServer.PrivateKey == "" {
		keys, err := wgcli.GenerateKeyPair()
		if err != nil {
			log.Fatalf("unable to generate wg key pair: %s", err)
		}

		wgServer.PublicKey = keys["publicKey"]
		wgServer.PrivateKey = keys["privateKey"]
		err = wgServer.saveBothConfigs()
		if err != nil {
			log.Fatalf("unable to save server configs: %s", err)
		}
	}

	if _, err := os.Stat(wgServer.WgConfigPath); os.IsNotExist(err) {
		wgServer.saveBothConfigs()
		if err != nil {
			log.Fatalf("unable to save server configs: %s", err)
		}
	}

	return &wgServer
}

// CreateServerConfig creates a new server config file and returns the struct
func CreateServerConfig() *WgServer {
	keys, err := wgcli.GenerateKeyPair()
	if err != nil {
		log.Fatalf("unable to generate wg key pair: %s", err)
	}
	// TODO: read from a template
	wgServer := &WgServer{
		Port:         "58210",
		VirtualIP:    "10.22.0.1",
		CIDR:         "16",
		DNS:          "1.1.1.1",
		WgConfigPath: "/etc/wireguard/wg0.conf",
		PublicKey:    keys["publicKey"],
		PrivateKey:   keys["privateKey"],
		Peers:        []Peer{},
	}
	err = wgServer.getPublicIPAddr()
	if err != nil {
		log.Fatalf("unable to get public IP address: %s", err)
	}

	err = wgServer.saveBothConfigs()
	if err != nil {
		log.Fatalf("unable to save server and wg configs: %s", err)
	}
	return wgServer
}

func (s *WgServer) saveBothConfigs() error {
	j, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		log.Printf("unable to save server config: %s", err)
		return err
	}

	err = ioutil.WriteFile(fileName, j, 0600)
	if err != nil {
		log.Printf("unable to write server config JSON file: %s", err)
		return err
	}

	tmpl := template.Must(template.ParseFiles("templates/server.conf.tmpl"))
	f, err := os.OpenFile(s.WgConfigPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("unable to open wireguard config file: %s", err)
		return err
	}
	defer f.Close()

	err = tmpl.Execute(f, s)
	if err != nil {
		log.Printf("error writing template to file: %s", err)
		return err
	}
	return nil
}

func (s *WgServer) getPublicIPAddr() error {
	// Alternative: ip -4 a show wlp2s0 | grep -oP '(?<=inet\s)\d+(\.\d+){3}'
	// Note: this does not make an actual connection and can be used offline
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		return err
	}
	defer conn.Close()
	s.PublicIP = conn.LocalAddr().(*net.UDPAddr).IP.String()

	return nil
}

func (s *WgServer) nextAvailableIP(assignedIP string) (string, error) {
	usedIPs := make(map[string]struct{})
	usedIPs[s.VirtualIP] = struct{}{}
	for _, p := range s.Peers {
		usedIPs[p.VirtualIP] = struct{}{}
	}

	_, ipNet, err := net.ParseCIDR(s.VirtualIP + "/" + s.CIDR)
	if err != nil {
		return "", errors.New("nextAvailableIP: server IP address incorrect")
	}

	if assignedIP != "" {
		ip, _, err := net.ParseCIDR(assignedIP + "/" + s.CIDR)
		if err != nil {
			return "", errors.New("addPeer: incorrectly formatted IP address")
		}

		if !ipNet.Contains(ip) {
			return "", errors.New("addPeer: assigned peer IP not in server subnet")
		}

		if _, ok := usedIPs[assignedIP]; !ok {
			return assignedIP, nil
		}
	}

	networkIP, broadcastIP := cidr.AddressRange(ipNet)
	// Don't use network address and broadcast address
	firstIP := cidr.Inc(networkIP)
	lastIP := cidr.Dec(broadcastIP)

	for i := firstIP; !lastIP.Equal(i); i = cidr.Inc(i) {
		if _, ok := usedIPs[i.To4().String()]; !ok {
			return i.To4().String(), nil
		}
	}

	return "", errors.New("nextAvailableIP: no available IPs")
}

// CheckServerActive queries systemd to check that wg server is up
func (s *WgServer) CheckServerActive() {
	cmd := exec.Command("systemctl", "is-active", "--quiet", "wg-quick@wg0")
	if err := cmd.Run(); err != nil {
		s.Active = false
	}
	s.Active = true
}

// ActivateServer starts wg server through systemd wg-quick unit
func (s *WgServer) ActivateServer() {
	cmd := exec.Command("systemctl", "start", "wg-quick@wg0")
	if err := cmd.Run(); err != nil {
		s.Active = false
	}
	s.Active = true
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func gzipHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		h.ServeHTTP(gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
	})
}
