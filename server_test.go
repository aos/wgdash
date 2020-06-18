package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleIndex(t *testing.T) {
	t.Run("returns the homepage", func(t *testing.T) {
		s := mockNewWgServer()

		request, err := http.NewRequest(http.MethodGet, "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		response := httptest.NewRecorder()
		s.ServeHTTP(response, request)

		got := response.Body.String()
		want := "Server Settings"

		if !strings.Contains(got, want) {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestHandlePeersAPI(t *testing.T) {
	t.Run("Returns peer information", func(t *testing.T) {
		s := mockNewWgServer()
		s.Peers = append(s.Peers, Peer{
			Active:     true,
			Name:       "Louie",
			ID:         3,
			PublicKey:  "abcdefg0==",
			PrivateKey: "shh==secret",
			VirtualIP:  "10.11.32.87",
		})

		req, err := http.NewRequest(http.MethodGet, "/api/peers/3", nil)
		if err != nil {
			t.Fatal(err)
		}

		res := httptest.NewRecorder()
		s.ServeHTTP(res, req)

		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v, want %v",
				status, http.StatusOK)
		}

		got := res.Body.String()
		want := "PrivateKey = shh==secret"

		if !strings.Contains(got, want) {
			t.Errorf("handler returned unexpected output: got %v, want %v",
				got, want)
		}
	})
}

func mockNewWgServer() *WgServer {
	wgServer := MakeTestServerConfig()
	wgServer.Routes()
	return wgServer
}

func MakeTestServerConfig() *WgServer {
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
