package handlers

import (
	"io"
	"net/http"
	//"os/exec"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin: func(r *http.Request) bool {
		// same-origin; adjust if you front this behind another host
		return true
	},
}

type wsAuthMsg struct {
	Op       string `json:"op"`       // expect "auth"
	Password string `json:"password"` // plain for DEV; move to WSS in prod
	Cols     int    `json:"cols"`
	Rows     int    `json:"rows"`
}

func SSHWS(c *gin.Context) {
	// Query: /ws/ssh?host=127.0.0.1&port=22&user=fahrizal
	host := c.Query("host")
	if host == "" {
		host = "127.0.0.1"
	}
	port := c.DefaultQuery("port", "22")
	user := c.Query("user")
	if user == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 1st message must be auth info
	var auth wsAuthMsg
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	if err := conn.ReadJSON(&auth); err != nil || auth.Op != "auth" {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("auth error\n"))
		return
	}
	conn.SetReadDeadline(time.Time{})

	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(auth.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // DEV ONLY
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", host+":"+port, cfg)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("ssh dial error: "+err.Error()+"\n"))
		return
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("ssh session error: "+err.Error()+"\n"))
		return
	}
	defer session.Close()

	// PTY
	cols := auth.Cols
	rows := auth.Rows
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
		_ = conn.WriteMessage(websocket.TextMessage, []byte("pty error: "+err.Error()+"\n"))
		return
	}

	stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()
	stderr, _ := session.StderrPipe()

	// Start user shell (login shell if available)
	if err := session.Start("/bin/bash -l"); err != nil {
		// fallback to /bin/sh
		_ = session.Start("/bin/sh")
	}

	// Copy SSH -> WS
	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				_ = conn.WriteMessage(websocket.BinaryMessage, buf[:n])
			}
			if err != nil {
				return
			}
		}
	}()
	go func() { io.Copy(websocketWriter{conn}, stderr) }()

	// Copy WS -> SSH
	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if mt == websocket.TextMessage || mt == websocket.BinaryMessage {
			_, _ = stdin.Write(data)
		}
	}

	_ = session.Close()
}

type websocketWriter struct{ *websocket.Conn }

func (w websocketWriter) Write(p []byte) (int, error) {
	return len(p), w.Conn.WriteMessage(websocket.BinaryMessage, p)
}

// Optional: tiny helper to parse ints safely (unused but handy)
func atoi(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

// Optional: server-side ping to keep idle sessions alive (uncomment if you need)
// func keepAlive(conn *websocket.Conn) {
// 	ticker := time.NewTicker(20 * time.Second)
// 	defer ticker.Stop()
// 	for range ticker.C {
// 		_ = conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second))
// 	}
// }
