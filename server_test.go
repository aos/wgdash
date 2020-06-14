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

		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()
		s.ServeHTTP(response, request)

		got := response.Body.String()
		want := "Server Settings"

		if !strings.Contains(got, want) {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func mockNewWgServer() *WgServer {
	wgServer := makeTestServerConfig()
	wgServer.Routes()
	return wgServer
}
