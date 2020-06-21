package main

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"text/template"
)

var peer = Peer{
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
	peerOutput := `
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
		{1, baseOutput + peerOutput},
		{2, baseOutput + peerOutput + peerOutput},
	}

	for _, tt := range serverTemplates {
		t.Run(fmt.Sprintf("%d peers in template", tt.numPeers), func(t *testing.T) {
			tmpl := template.Must(template.ParseFiles("templates/server.conf.tmpl"))
			var b bytes.Buffer
			s := MakeTestServerConfig()
			for i := 0; i < tt.numPeers; i++ {
				s.Peers = append(s.Peers, peer)
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

func TestPeerConfigTemplate(t *testing.T) {
	var peerTemplates = []struct {
		qr     bool
		output string
	}{
		{false, `[Interface]
Address = 10.11.32.87/32
PrivateKey = shh==secret
PostUp = iptables -A FORWARD -i %i -o %i -j ACCEPT
PostDown = iptables -D FORWARD -i %i -o %i -j ACCEPT
SaveConfig = false

[Peer]
PublicKey = helloworld==
Endpoint = 188.272.271.04:4566
AllowedIPs = 10.22.0.0/16
`},
		{true, `[Interface]
Address = 10.11.32.87/32
PrivateKey = shh==secret

[Peer]
PublicKey = helloworld==
Endpoint = 188.272.271.04:4566
AllowedIPs = 10.22.0.0/16
`},
	}

	for _, tt := range peerTemplates {
		t.Run(fmt.Sprintf("QR code requested: %v", tt.qr), func(t *testing.T) {
			tmpl := template.Must(template.ParseFiles("templates/peer.conf.tmpl"))
			var b bytes.Buffer
			s := MakeTestServerConfig()
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
				QRCode          bool
			}{
				VirtualIP:       peer.VirtualIP,
				PrivateKey:      peer.PrivateKey,
				ServerPublicKey: s.PublicKey,
				PublicIP:        s.PublicIP,
				Port:            s.Port,
				AllowedIPs:      ipNet.String(),
				QRCode:          tt.qr,
			})

			if err != nil {
				t.Fatalf("error opening template: %s", err)
			}
			if b.String() != tt.output {
				t.Errorf("got: %s, want: %s\n", b.String(), tt.output)
			}
		})
	}
}
