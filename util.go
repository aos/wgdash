package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"text/template"

	"github.com/aos/wgdash/wgcli"
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
	return &wgServer
}

// CreateServerConfig creates a new server config file and returns the struct
func CreateServerConfig() *WgServer {
	keys, err := wgcli.GenerateKeyPair()
	if err != nil {
		log.Fatalf("unable to generate wg key pair: %s", err)
	}
	pubIP, err := getPublicIPAddr()
	if err != nil {
		log.Fatalf("CreateServerConfig: %s", err)
	}
	wgServer := &WgServer{
		PublicIP:     pubIP,
		Port:         "51820",
		VirtualIP:    "10.22.0.1",
		CIDR:         "16",
		DNS:          "1.1.1.1",
		WgConfigPath: "/etc/wireguard/wg0.conf",
		PublicKey:    keys["publicKey"],
		PrivateKey:   keys["privateKey"],
		Clients:      []Client{},
	}
	err = saveBothConfigs(wgServer)
	if err != nil {
		log.Fatalf("Unable to save server and wg configs: %s.\nShutting down.", err)
	}
	return wgServer
}

func saveBothConfigs(conf *WgServer) error {
	j, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		log.Printf("unable to save server config: %s", err)
		return err
	}

	err = ioutil.WriteFile(fileName, j, 0600)
	if err != nil {
		log.Printf("Unable to write server config JSON file: %s", err)
		return err
	}

	tmpl := template.Must(template.ParseFiles("templates/server.conf.tmpl"))
	f, err := os.OpenFile("wg0.conf", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		log.Printf("Unable to open wireguard config file: %s", err)
		return err
	}

	err = tmpl.Execute(f, conf)
	if err != nil {
		log.Printf("error writing template: %s", err)
		return err
	}
	return nil
}

func getPublicIPAddr() (string, error) {
	// Alternative: ip -4 a show wlp2s0 | grep -oP '(?<=inet\s)\d+(\.\d+){3}'
	// Note: this does not make an actual connection and can be used offline
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String(), nil
}
