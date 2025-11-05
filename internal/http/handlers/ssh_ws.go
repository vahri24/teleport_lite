package handlers

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
	"gorm.io/datatypes"
	"teleport_lite/internal/models"
	"teleport_lite/internal/auth"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin: func(r *http.Request) bool {
		return true // local-only, adjust later if needed
	},
}

type wsAuthMsg struct {
	Op   string `json:"op"`   // operation: "auth"
	Cols int    `json:"cols"` // terminal width
	Rows int    `json:"rows"` // terminal height
}

// SSHWS establishes SSH session via WebSocket using private key from DB
func SSHWS(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.DefaultQuery("host", "127.0.0.1")
		port := c.DefaultQuery("port", "22")
		user := c.DefaultQuery("user", "root")

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// ✅ Extract user agent from request
		userAgent := c.Request.Header.Get("User-Agent")


		// ✅ Extract user info from JWT
		var userID, orgID int64
		var webUserName, webUserEmail string
		if claimsVal, ok := c.Get("claims"); ok {
			if cl, ok := claimsVal.(*auth.Claims); ok {
				userID = int64(cl.UserID)
				orgID = int64(cl.OrgID)
			}
		}

		var dbUser models.User
		if err := gdb.First(&dbUser, userID).Error; err == nil {
		    webUserName = dbUser.Name
		    webUserEmail = dbUser.Email
		} else {
		    webUserName = "Unknown"
}

		// ✅ Detect client IP
		clientIP, _, _ := net.SplitHostPort(c.Request.RemoteAddr)

		// ✅ Prepare metadata JSON
		meta := map[string]string{
			"ssh_user": user,
			"host":     host,
			"initiator":  webUserName,
    		"initiator_email": webUserEmail,
		}
		metaJSON, _ := json.Marshal(meta)

		// ✅ Record SSH connect
		connectLog := models.AuditLog{
			OrgID:        	orgID,
			UserID:       	userID,
			Action:       	"ssh_connect",
			ResourceType: 	"SSH",
			IP:           	clientIP,
			UserAgent:     	userAgent,
			InitiatorName: 	webUserName,
			Metadata:     	datatypes.JSON(metaJSON),
			CreatedAt:    	time.Now(),
		}
		_ = gdb.Create(&connectLog).Error

		// Wait for initial auth message from client
		var auth wsAuthMsg
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		if err := conn.ReadJSON(&auth); err != nil || auth.Op != "auth" {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("auth error\n"))
			return
		}
		conn.SetReadDeadline(time.Time{})

		// ✅ Fetch resource from DB using the same method as local_agent.go
		var resource models.Resource
		if err := gdb.Where("host = ?", host).First(&resource).Error; err != nil {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("resource not found for host "+host+"\n"))
			return
		}

		if resource.PrivateKey == "" {
			_ = conn.WriteMessage(websocket.TextMessage,
				[]byte("❌ No private key found in DB for host "+host+"\n"))
			return
		}

		// ✅ Parse private key from DB
		signer, err := ssh.ParsePrivateKey([]byte(resource.PrivateKey))
		if err != nil {
			_ = conn.WriteMessage(websocket.TextMessage,
				[]byte("❌ Invalid private key in DB: "+err.Error()+"\n"))
			return
		}

		cfg := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         10 * time.Second,
		}

		client, err := ssh.Dial("tcp", host+":"+port, cfg)
		if err != nil {
			_ = conn.WriteMessage(websocket.TextMessage,
				[]byte("ssh dial error: "+err.Error()+"\n"))
			return
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			_ = conn.WriteMessage(websocket.TextMessage,
				[]byte("ssh session error: "+err.Error()+"\n"))
			return
		}
		defer session.Close()

		cols, rows := auth.Cols, auth.Rows
		if cols == 0 {
			cols = 120
		}
		if rows == 0 {
			rows = 32
		}

		if err := session.RequestPty("xterm-256color", rows, cols, ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}); err != nil {
			_ = conn.WriteMessage(websocket.TextMessage,
				[]byte("pty error: "+err.Error()+"\n"))
			return
		}

		stdin, _ := session.StdinPipe()
		stdout, _ := session.StdoutPipe()
		stderr, _ := session.StderrPipe()

		if err := session.Start("/bin/bash -l"); err != nil {
			_ = session.Start("/bin/sh")
		}

		// SSH → WebSocket
		go io.Copy(websocketWriter{conn}, stdout)
		go io.Copy(websocketWriter{conn}, stderr)

		// WebSocket → SSH
		for {
			mt, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if mt == websocket.TextMessage || mt == websocket.BinaryMessage {
				_, _ = stdin.Write(data)
			}
		}

		// ✅ Record SSH disconnect when session ends
		disconnectLog := models.AuditLog{
			OrgID:        	orgID,
			UserID:       	userID,
			Action:       	"ssh_connect",
			ResourceType: 	"SSH",
			IP:           	clientIP,
			UserAgent:     	userAgent,
			InitiatorName: 	webUserName,
			Metadata:     	datatypes.JSON(metaJSON),
			CreatedAt:    	time.Now(),
		}
		_ = gdb.Create(&disconnectLog).Error

		_ = session.Close()
	}
}

type websocketWriter struct{ *websocket.Conn }

func (w websocketWriter) Write(p []byte) (int, error) {
	return len(p), w.Conn.WriteMessage(websocket.BinaryMessage, p)
}

func atoi(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}
