package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/aos/wgdash/wgcli"
)

var fileName = "server_config.json"

// LoadWriteServerConfig looks for the server config and
// if it can't find it, will make a new one.
func LoadWriteServerConfig() {
	var wgServer *WgServer
	_, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Server config does not exist. Creating...")
			keys, err := wgcli.GenerateKeyPair()
			if err != nil {
				log.Fatalf("unable to create server config: %s\nShutting down.", err)
			}
			pubIP, err := getPublicIPAddr()
			if err != nil {
				log.Fatalf("LoadWriteServerConfig: %s", err)
			}
			// 2. create default server config struct
			wgServer = &WgServer{
				PublicIP:     pubIP,
				VirtualIP:    "10.22.0.1",
				CIDR:         "16",
				DNS:          "1.1.1.1",
				WgConfigPath: "/etc/wireguard/wg0.conf",
				PublicKey:    keys["publicKey"],
				Clients:      []Client{},
			}
			// 3. save json config -- this also acts as the "db",
			// storing our private key
			f, err := json.MarshalIndent(struct {
				WgServer
				PrivateKey string
			}{
				WgServer:   *wgServer,
				PrivateKey: keys["privateKey"],
			}, "", "    ")
			if err != nil {
				log.Fatalf("unable to create server config: %s\nShutting down.", err)
			}
			err = ioutil.WriteFile(fileName, f, 0600)
			if err != nil {
				log.Fatal("Unable to write server config JSON file")
			}
		}
	}
	//tmpl := template.Must(template.ParseFiles("templates/server.conf.tmpl"))
	// 3. write out to template
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
