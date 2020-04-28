package wgcli

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GenerateKeyPair uses wg cli to create a private and public key pair
func GenerateKeyPair() (map[string]string, error) {
	cmd := exec.Command("wg", "genkey")
	priv, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("wg genkey: %s", err)
	}

	cmd = exec.Command("wg", "pubkey")
	cmd.Stdin = bytes.NewReader(priv)
	pub, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("wg pubkey: %s", err)
	}

	return map[string]string{
		"privateKey": strings.TrimSpace(string(priv)),
		"publicKey":  strings.TrimSpace(string(pub)),
	}, nil
}
