package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"teleport_lite/internal/models"
)

// RegisterAgent registers a remote agent in the database
// and automatically adds its public key to authorized_keys
func RegisterAgent(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Hostname  string `json:"hostname"`
			IP        string `json:"ip"`
			OS        string `json:"os"`
			PublicKey string `json:"public_key"`
			PrivateKey string `json:"private_key"`
			Role      string `json:"role"`
		}

		// ✅ Parse incoming JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		// ✅ Build metadata
		meta := map[string]interface{}{
			"hostname": req.Hostname,
			"os":       req.OS,
			"role":     req.Role,
		}
		
		metaJSON, _ := json.Marshal(meta)

		// ✅ Prepare resource struct
		resource := models.Resource{
			OrgID:         1,
			Name:          req.Hostname,
			Type:          "SSH",
			Host:          req.IP,
			Port:          22,
			ExternalRef:   "Remote Agent",
			PublicKey:     req.PublicKey,
			PrivateKey:    req.PrivateKey,
			Status:        "online",
			LastHeartbeat: time.Now(),
			Metadata:      datatypes.JSON(metaJSON),
		}

		// ✅ Save or update record (same style as local_agent.go)
		if err := gdb.Where("host = ?", req.IP).
			Assign(resource).FirstOrCreate(&resource).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error: " + err.Error()})
			return
		}

		// ✅ Append agent key to authorized_keys
		if err := addAgentKey(req.PublicKey); err != nil {
			log.Printf("⚠️ Failed to append agent key: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "agent registered but failed to append key: " + err.Error(),
			})
			return
		}

		log.Printf("✅ Agent %s (%s) registered and key added.", req.Hostname, req.IP)
		c.JSON(http.StatusOK, gin.H{
			"message":     "agent registered successfully",
			"resource_id": resource.ID,
		})
	}
}

// AgentHeartbeat updates the agent’s heartbeat timestamp
func AgentHeartbeat(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			IP string `json:"ip"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		if err := gdb.Model(&models.Resource{}).
			Where("host = ?", req.IP).
			Updates(map[string]interface{}{
				"last_heartbeat": time.Now(),
				"status":         "online",
				"updated_at":     time.Now(),
			}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
			return
		}

		log.Printf("💓 Heartbeat received from agent %s", req.IP)
		c.JSON(http.StatusOK, gin.H{"status": "heartbeat ok"})
	}
}

// -----------------------------------------------------------
// Helper: addAgentKey appends agent’s public key to controller’s ~/.ssh/authorized_keys
// -----------------------------------------------------------
func addAgentKey(pubKey string) error {
	if strings.TrimSpace(pubKey) == "" {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	authFile := filepath.Join(home, ".ssh", "authorized_keys")

	// Ensure .ssh directory exists
	if err := os.MkdirAll(filepath.Dir(authFile), 0700); err != nil {
		return err
	}

	// Check if the key already exists
	data, _ := os.ReadFile(authFile)
	if strings.Contains(string(data), strings.TrimSpace(pubKey)) {
		log.Printf("ℹ️ Agent key already present in %s", authFile)
		return nil
	}

	// Append new key
	f, err := os.OpenFile(authFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(pubKey + "\n"); err != nil {
		return err
	}

	log.Printf("🔐 Added agent public key to %s", authFile)
	return nil
}
