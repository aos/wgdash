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

// AddPeer adds a new peer to an active server
func AddPeer(pubkey, ip string) error {
	cmd := exec.Command("wg", "set", "wg0", "peer", pubkey, "allowed-ips", ip+"/32")
	res, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(fmt.Sprint(err) + ": " + string(res))
	}
	return nil
}

// DeletePeer deletes a given peer from an active server
func DeletePeer(pubkey string) error {
	cmd := exec.Command("wg", "set", "wg0", "peer", pubkey, "remove")
	res, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(fmt.Sprint(err) + ": " + string(res))
	}
	return nil
}
