package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"
)

// Settings are the settings for the wireguard server
type Settings struct {
	PublicIP  string
	Port      string
	Iface     string
	VirtualIP string
	CIDR      string
	PublicKey string
}

// Peer is any client device connecting to the wg server
type Peer struct {
	QRcode    string
	Active    bool
	Device    string
	PublicKey string
	VirtualIP string
}

var templates = template.Must(
	// ParseFiles uses the base name
	template.ParseFiles("templates/settings.html", "templates/peers.html"),
)

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := "settings"
	if strings.HasSuffix(r.URL.Path, "peers") {
		title = "peers"
	}
	page := &Page{Title: title}
	renderTemplate(w, title, page)
}

func main() {
	http.HandleFunc("/", viewHandler)
	log.Fatal(http.ListenAndServe(":3100", nil))
}
