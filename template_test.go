package main

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"testing"
	"text/template"
)

var client = Peer{
	Active:     true,
	Name:       "Louie",
	ID:         3,
	PublicKey:  "abcdefg0==",
	PrivateKey: "shh==secret",
	VirtualIP:  "10.11.32.87",
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
			s := makeTestServerConfig()
			for i := 0; i < tt.numPeers; i++ {
				s.Peers = append(s.Peers, client)
			}
			err := tmpl.Execute(&b, *s)
			if err != nil {
				t.Fatalf("error opening template: %s", err)
			}
			if b.String() != tt.out {
				t.Errorf("got: %s, want: %s\n", b.String(), tt.out)
			}
		})
	}
}

func TestClientConfigTemplate(t *testing.T) {
	out := `[Interface]
Address = 10.11.32.87/32
PrivateKey = shh==secret
PostUp = iptables -A FORWARD -i %i -o %i -j ACCEPT
PostDown = iptables -D FORWARD -i %i -o %i -j ACCEPT
SaveConfig = false

[Peer]
PublicKey = helloworld==
Endpoint = 188.272.271.04:4566
AllowedIPs = 10.22.0.0/16
`

	tmpl := template.Must(template.ParseFiles("templates/client.conf.tmpl"))
	var b bytes.Buffer
	s := makeTestServerConfig()
	_, ipNet, err := net.ParseCIDR(s.VirtualIP + "/" + s.CIDR)
	if err != nil {
		t.Errorf("error parsing CIDR: %s", err)
	}

	err = tmpl.Execute(&b, struct {
		VirtualIP       string
		PrivateKey      string
		ServerPublicKey string
		PublicIP        string
		Port            string
		AllowedIPs      string
	}{
		VirtualIP:       client.VirtualIP,
		PrivateKey:      client.PrivateKey,
		ServerPublicKey: s.PublicKey,
		PublicIP:        s.PublicIP,
		Port:            s.Port,
		AllowedIPs:      ipNet.String(),
	})

	if err != nil {
		t.Fatalf("error opening template: %s", err)
	}
	if b.String() != out {
		t.Errorf("got: %s, want: %s\n", b.String(), out)
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
