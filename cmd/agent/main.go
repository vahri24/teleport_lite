package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	//"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {
	controllerURL := os.Getenv("CONTROLLER_URL") // e.g., http://192.168.1.10:8080
	if controllerURL == "" {
		log.Fatal("❌ CONTROLLER_URL not set. Example: CONTROLLER_URL=http://192.168.1.10:8080 ./teleport-agent")
	}

	hostname, _ := os.Hostname()
	osVersion := detectOS()
	ip := getLocalIP()

	usr, _ := user.Current()
	keyDir := filepath.Join(usr.HomeDir, ".teleport-agent")
	priv := filepath.Join(keyDir, "id_rsa")
	pub := priv + ".pub"

	// ✅ Generate keypair if missing
	if err := generateSSHKeyPair(priv, pub); err != nil {
		log.Fatalf("❌ keygen failed: %v", err)
	}

	pubBytes, _ := os.ReadFile(pub)
	pubStr := strings.TrimSpace(string(pubBytes))

	privBytes, _ := os.ReadFile(priv)
	privStr := strings.TrimSpace(string(privBytes))

	payload := map[string]string{
		"hostname":   hostname,
		"ip":         ip,
		"os":         osVersion,
		"public_key": pubStr,
		"private_key": privStr,
		"role":       "agent",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(controllerURL+"/agents/register", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("❌ register failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️ Controller responded with %d", resp.StatusCode)
	} else {
		log.Printf("✅ Registered agent: %s (%s)", hostname, ip)
	}

	// ✅ Heartbeat loop
	for {
		time.Sleep(60 * time.Second)
		hb, _ := json.Marshal(map[string]string{"ip": ip})
		http.Post(controllerURL+"/agents/heartbeat", "application/json", bytes.NewReader(hb))
	}
}

// ------------------------------------------------------------
// SSH Key Generation
// ------------------------------------------------------------
func generateSSHKeyPair(privPath, pubPath string) error {
	if _, err := os.Stat(privPath); err == nil {
		log.Println("🔑 Existing keypair found — skipping generation.")
		return nil
	}

	log.Println("🔑 Generating SSH keypair for remote agent...")
	if err := os.MkdirAll(filepath.Dir(privPath), 0700); err != nil {
		return err
	}

	// Generate RSA key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Save private key
	privFile, err := os.Create(privPath)
	if err != nil {
		return err
	}
	defer privFile.Close()

	privBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	if err := pem.Encode(privFile, privBlock); err != nil {
		return err
	}
	os.Chmod(privPath, 0600)

	// Create public key
	pub, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		return err
	}
	pubBytes := ssh.MarshalAuthorizedKey(pub)
	if err := os.WriteFile(pubPath, pubBytes, 0644); err != nil {
		return err
	}

	log.Printf("✅ SSH keypair created at %s", privPath)
	return nil
}

// ------------------------------------------------------------
// Helper functions
// ------------------------------------------------------------
func detectOS() string {
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				return strings.Trim(line[13:], `"`)
			}
		}
	}
	return runtime.GOOS
}

func getLocalIP() string {
	// Detect LAN IP, fallback to loopback if not found
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "127.0.0.1"
}
