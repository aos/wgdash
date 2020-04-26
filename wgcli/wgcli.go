package wgcli

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
)

// GenerateKeyPair uses wg cli to create a private and public key pair
func GenerateKeyPair() map[string]string {
	cmd := exec.Command("wg", "genkey")
	priv, err := cmd.Output()
	if err != nil {
		log.Printf("wg genkey: %s", err)
	}

	cmd = exec.Command("wg", "pubkey")
	cmd.Stdin = bytes.NewReader(priv)
	pub, err := cmd.Output()
	if err != nil {
		log.Printf("wg pubkey: %s", err)
	}

	return map[string]string{
		"privateKey": strings.TrimSpace(string(priv)),
		"publicKey":  strings.TrimSpace(string(pub)),
	}
}
