package agent

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"teleport_lite/internal/models"
)

// RunLocalAgent starts the local registration & heartbeat process
func RunLocalAgent(gdb *gorm.DB) {
	log.Println("🧠 Starting local agent registration...")

	if gdb == nil {
		log.Println("❌ Database not initialized for local agent")
		return
	}

	// Detect current non-root user
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("❌ Unable to get current user: %v", err)
	}

	homeDir := currentUser.HomeDir
	keyDir := filepath.Join(homeDir, ".teleport-lite")
	privateKeyPath := filepath.Join(keyDir, "id_rsa")
	publicKeyPath := privateKeyPath + ".pub"
	authKeysPath := filepath.Join(homeDir, ".ssh", "authorized_keys")

	hostname, _ := os.Hostname()
	osVersion := detectOS()
	ip := getLocalIP()

	// ✅ Generate SSH keypair if not exists
	if err := generateSSHKeyPair(privateKeyPath, publicKeyPath); err != nil {
		log.Printf("❌ Failed to generate SSH key: %v", err)
		return
	}

	pubKeyBytes, _ := ioutil.ReadFile(publicKeyPath)
	privKeyBytes, _ := ioutil.ReadFile(privateKeyPath)

	pubKeyStr := strings.TrimSpace(string(pubKeyBytes))
	privKeyStr := string(privKeyBytes)

	// ✅ Ensure authorized_keys includes public key
	appendKeyIfMissing(authKeysPath, pubKeyStr)

	meta := map[string]string{
		"hostname": hostname,
		"os":       osVersion,
		"role":     "controller",
	}
	metaJSON, _ := json.Marshal(meta)

	resource := models.Resource{
		OrgID:         1,
		Name:          hostname,
		Type:          "SSH",
		Host:          ip,
		Port:          22,
		ExternalRef:   "Local Controller",
		PublicKey:     pubKeyStr,
		PrivateKey:    privKeyStr, // ✅ now also stored
		Status:        "online",
		LastHeartbeat: time.Now(),
		Metadata:      datatypes.JSON(metaJSON),
	}

	// ✅ Save/update resource (exact same method as before)
	if err := gdb.Where("host = ?", resource.Host).
		Assign(resource).FirstOrCreate(&resource).Error; err != nil {
		log.Printf("❌ Failed to register resource: %v", err)
		return
	}

	log.Printf("✅ Local controller registered as resource id=%d host=%s user=%s", resource.ID, resource.Host, currentUser.Username)

	// ✅ Start heartbeat updater
	go startHeartbeat(gdb, resource.Host)
}

// startHeartbeat keeps updating resource status every 60s
func startHeartbeat(gdb *gorm.DB, host string) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		err := gdb.Model(&models.Resource{}).
			Where("host = ?", host).
			Updates(map[string]interface{}{
				"last_heartbeat": time.Now(),
				"status":         "online",
				"updated_at":     time.Now(),
			}).Error
		if err != nil {
			log.Printf("⚠️ Heartbeat update failed: %v", err)
		} 
	}
}

// ---------------------- Helpers ---------------------- //

func detectOS() string {
	if data, err := ioutil.ReadFile("/etc/os-release"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				return strings.Trim(line[13:], `"`)
			}
		}
	}
	return runtime.GOOS
}

func generateSSHKeyPair(privPath, pubPath string) error {
	if _, err := os.Stat(privPath); err == nil {
		log.Println("🔑 Existing SSH keypair found — skipping generation")
		return nil
	}

	log.Println("🔑 Generating new SSH keypair for Teleport Lite agent...")
	if err := os.MkdirAll(filepath.Dir(privPath), 0700); err != nil {
		return err
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

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

	pub, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		return err
	}

	pubBytes := ssh.MarshalAuthorizedKey(pub)
	if err := ioutil.WriteFile(pubPath, pubBytes, 0644); err != nil {
		return err
	}

	log.Printf("✅ SSH keypair created at %s", privPath)
	return nil
}

func appendKeyIfMissing(authPath, pubKey string) {
	data, _ := ioutil.ReadFile(authPath)
	if strings.Contains(string(data), pubKey) {
		log.Printf("ℹ️ Controller key already present in %s", authPath)
		return
	}
	if err := os.MkdirAll(filepath.Dir(authPath), 0700); err != nil {
		log.Printf("⚠️ Failed to create .ssh directory: %v", err)
		return
	}
	f, err := os.OpenFile(authPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("⚠️ Failed to open authorized_keys: %v", err)
		return
	}
	defer f.Close()
	f.WriteString(pubKey + "\n")
	log.Printf("🔐 Added controller public key to %s", authPath)
}

// Always return local IP (127.0.0.1)
func getLocalIP() string {
	return "127.0.0.1"
}
