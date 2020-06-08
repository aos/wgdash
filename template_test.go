package main

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"text/template"
)

var client = Peer{
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
PostDown = iptables -D FORWARD -i %i -o %i -j ACCEPT
SaveConfig = false
`
	clientOutput := `
# Louie
[Peer]
PublicKey = abcdefg0==
AllowedIPs = 10.11.32.87/32
`
	var serverTemplates = []struct {
		numPeers int
		out      string
	}{
		{0, baseOutput},
		{1, baseOutput + clientOutput},
		{2, baseOutput + clientOutput + clientOutput},
	}

	for _, tt := range serverTemplates {
		t.Run(fmt.Sprintf("%d peers in template", tt.numPeers), func(t *testing.T) {
			tmpl := template.Must(template.ParseFiles("templates/server.conf.tmpl"))
			var b bytes.Buffer
			serv := makeTestServerConfig()
			for i := 0; i < tt.numPeers; i++ {
				serv.Peers = append(serv.Peers, client)
			}
			err := tmpl.Execute(&b, *serv)
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
		PublicIP:     "188.272.271.04",
		Port:         "4566",
		VirtualIP:    "10.22.65.87",
		CIDR:         "16",
		PublicKey:    "helloworld==",
		PrivateKey:   "topsecret==",
		DNS:          "1.1.12.1",
		WgConfigPath: "/etc/hello/wg0.conf",
		mux:          http.NewServeMux(),
	}
}
