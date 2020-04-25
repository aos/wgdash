package main

import (
	"bytes"
	"net/http"
	"testing"
	"text/template"
)

func TestConfigTemplates(t *testing.T) {
	t.Run("server config", func(t *testing.T) {
		tmpl := template.Must(template.ParseFiles("templates/server.conf.tmpl"))
		var b bytes.Buffer
		wantOutput := `[Interface]
Address = 10.22.65.87
ListenPort = 4566
PrivateKey = topsecret==
PostUp = iptables -A FORWARD -i %i -o %i -j ACCEPT
PostDown = iptables -D FORWARD -i %i -i %i -j ACCEPT
SaveConfig = false

[Peer]
PublicKey = abcdefg0==
AllowedIPs = 10.11.32.87/32
`

		serv := makeTestServerConfig()
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
		if b.String() != wantOutput {
			t.Errorf("got: %s, want: %s\n", b.String(), wantOutput)
		}
	})
}

func makeTestServerConfig() *WgServer {
	peers := []Peer{
		{
			Active:    true,
			Name:      "Louie",
			PublicKey: "abcdefg0==",
			QRcode:    "andefdf==",
			VirtualIP: "10.11.32.87",
		},
	}
	return &WgServer{
		PublicIP:         "188.272.271.04",
		Port:             "4566",
		VirtualIP:        "10.22.65.87",
		CIDR:             "16",
		PublicKey:        "helloworld==",
		DNS:              "1.1.12.1",
		WgConfigPath:     "/etc/hello/wg0.conf",
		ServerConfigPath: "/home/louie/vpn",
		Peers:            peers,
		mux:              http.NewServeMux(),
	}
}
