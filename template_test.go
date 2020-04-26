package main

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"text/template"
)

var client = Client{
	Active:    true,
	Name:      "Louie",
	PublicKey: "abcdefg0==",
	QRcode:    "andefdf==",
	VirtualIP: "10.11.32.87",
}

func TestServerConfigTemplates(t *testing.T) {
	baseOutput := `[Interface]
Address = 10.22.65.87/16
ListenPort = 4566
PrivateKey = topsecret==
PostUp = iptables -A FORWARD -i %i -o %i -j ACCEPT
PostDown = iptables -D FORWARD -i %i -i %i -j ACCEPT
SaveConfig = false
`
	clientOutput := `
# Louie
[Client]
PublicKey = abcdefg0==
AllowedIPs = 10.11.32.87/32
`
	var serverTemplates = []struct {
		numClients int
		out        string
	}{
		{0, baseOutput},
		{1, baseOutput + clientOutput},
		{2, baseOutput + clientOutput + clientOutput},
	}

	for _, tt := range serverTemplates {
		t.Run(fmt.Sprintf("%d clients in template", tt.numClients), func(t *testing.T) {
			tmpl := template.Must(template.ParseFiles("templates/server.conf.tmpl"))
			var b bytes.Buffer
			serv := makeTestServerConfig()
			for i := 0; i < tt.numClients; i++ {
				serv.Clients = append(serv.Clients, client)
			}
			err := tmpl.Execute(&b, struct {
				WgServer
				PrivateKey string
			}{
				WgServer:   *serv,
				PrivateKey: "topsecret==",
			})
			if err != nil {
				t.Fatalf("error opening template: %s", err)
			}
			if b.String() != tt.out {
				t.Errorf("got: %s, want: %s\n", b.String(), tt.out)
			}
		})
	}
}

func makeTestServerConfig() *WgServer {
	return &WgServer{
		PublicIP:         "188.272.271.04",
		Port:             "4566",
		VirtualIP:        "10.22.65.87",
		CIDR:             "16",
		PublicKey:        "helloworld==",
		DNS:              "1.1.12.1",
		WgConfigPath:     "/etc/hello/wg0.conf",
		ServerConfigPath: "/home/louie/vpn",
		mux:              http.NewServeMux(),
	}
}
