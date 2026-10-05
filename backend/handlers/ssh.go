package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"web-ssh/models"
	"web-ssh/store"
	"web-ssh/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
)

type SSHHandler struct {
	store    store.Store
	sessions map[string]*SSHSession
	mu       sync.RWMutex
}

type SSHSession struct {
	ID        string
	ServerID  string
	Client    *ssh.Client
	Session   *ssh.Session
	Stdin     io.WriteCloser
	Connected bool
	StartTime time.Time
	CloseChan chan struct{}
}

func NewSSHHandler(s store.Store) *SSHHandler {
	return &SSHHandler{store: s, sessions: make(map[string]*SSHSession)}
}

func (h *SSHHandler) HandleTerminal(c *gin.Context) {
	serverID := c.Query("server_id")
	if serverID == "" { c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "缺少服务器ID"}); return }

	server, ok := h.store.GetServerByID(serverID)
	if !ok { c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "服务器不存在"}); return }

	password, err := h.store.DecryptServerPassword(serverID)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "解密密码失败"}); return }

	cols, _ := strconv.Atoi(c.DefaultQuery("cols", "80"))
	rows, _ := strconv.Atoi(c.DefaultQuery("rows", "24"))

	upgrader := getUpgrader()
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil { log.Printf("WebSocket upgrade failed: %v", err); return }
	defer ws.Close()

	sshClient, err := createSSHClient(server.Host, server.Port, server.Username, password, server.PrivateKey)
	if err != nil {
		writeWSMessage(ws, models.WSMessage{Type: "error", Data: fmt.Sprintf("连接失败: %s", err.Error())})
		return
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		writeWSMessage(ws, models.WSMessage{Type: "error", Data: fmt.Sprintf("创建会话失败: %s", err.Error())})
		return
	}
	defer session.Close()

	termWidth := cols; if termWidth == 0 { termWidth = 80 }
	termHeight := rows; if termHeight == 0 { termHeight = 24 }

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
		writeWSMessage(ws, models.WSMessage{Type: "error", Data: fmt.Sprintf("请求终端失败: %s", err.Error())})
		return
	}

	stdin, err := session.StdinPipe()
	if err != nil { writeWSMessage(ws, models.WSMessage{Type: "error", Data: fmt.Sprintf("获取输入流失败: %s", err)}); return }
	stdout, err := session.StdoutPipe()
	if err != nil { writeWSMessage(ws, models.WSMessage{Type: "error", Data: fmt.Sprintf("获取输出流失败: %s", err)}); return }
	stderr, err := session.StderrPipe()
	if err != nil { writeWSMessage(ws, models.WSMessage{Type: "error", Data: fmt.Sprintf("获取错误流失败: %s", err)}); return }

	if err := session.Shell(); err != nil {
		writeWSMessage(ws, models.WSMessage{Type: "error", Data: fmt.Sprintf("启动Shell失败: %s", err.Error())})
		return
	}

	writeWSMessage(ws, models.WSMessage{
		Type: "connected",
		Data: fmt.Sprintf("已连接到 %s (%s)", server.Name, server.Host),
	})

	sessionID := utils.GenerateID()
	sshSession := &SSHSession{
		ID: sessionID, ServerID: serverID, Client: sshClient, Session: session,
		Stdin: stdin, Connected: true, StartTime: time.Now(), CloseChan: make(chan struct{}),
	}
	h.mu.Lock(); h.sessions[sessionID] = sshSession; h.mu.Unlock()
	defer func() { h.mu.Lock(); delete(h.sessions, sessionID); h.mu.Unlock(); sshSession.Connected = false; close(sshSession.CloseChan) }()

	done := make(chan struct{})
	var closeOnce sync.Once
	closeDone := func() { closeOnce.Do(func() { close(done) }) }

	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := stdout.Read(buf)
			if err != nil { if err != io.EOF { log.Printf("SSH stdout read error: %v", err) }; closeDone(); return }
			if n > 0 { writeWSMessage(ws, models.WSMessage{Type: "output", Data: string(buf[:n])}) }
		}
	}()

	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := stderr.Read(buf)
			if err != nil { return }
			if n > 0 { writeWSMessage(ws, models.WSMessage{Type: "output", Data: string(buf[:n])}) }
		}
	}()

	go func() {
		for {
			_, message, err := ws.ReadMessage()
			if err != nil { closeDone(); return }
			var msg models.WSMessage
			if err := jsonUnmarshal(message, &msg); err != nil { continue }
			switch msg.Type {
			case "input":
				if _, err := stdin.Write([]byte(msg.Data)); err != nil { log.Printf("SSH stdin write error: %v", err); closeDone(); return }
			case "resize":
				if msg.Cols > 0 && msg.Rows > 0 { _ = session.WindowChange(msg.Rows, msg.Cols) }
			case "ping":
				writeWSMessage(ws, models.WSMessage{Type: "pong", Data: "ok"})
			}
		}
	}()

	select {
	case <-done:
	case <-sshSession.CloseChan:
	}
	writeWSMessage(ws, models.WSMessage{Type: "disconnected", Data: "连接已断开"})
}

func (h *SSHHandler) GetActiveSessions(c *gin.Context) {
	h.mu.RLock(); defer h.mu.RUnlock()
	type SessionInfo struct { ID string `json:"id"`; ServerID string `json:"server_id"`; ConnectedAt string `json:"connected_at"`; Duration string `json:"duration"` }
	var sessions []SessionInfo
	for id, s := range h.sessions {
		if s.Connected {
			sessions = append(sessions, SessionInfo{
				ID: id, ServerID: s.ServerID,
				ConnectedAt: s.StartTime.Format("2006-01-02 15:04:05"),
				Duration: utils.FormatUptime(time.Since(s.StartTime)),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "sessions": sessions})
}

func (h *SSHHandler) HandleDirectTerminal(c *gin.Context) {
	token := c.Query("token")

	var directSession *DirectSession
	if token != "" {
		directSessionsMutex.RLock()
		ds, exists := directSessions[token]
		directSessionsMutex.RUnlock()
		if exists {
			directSession = ds
		}
	}

	cols, _ := strconv.Atoi(c.DefaultQuery("cols", "80"))
	rows, _ := strconv.Atoi(c.DefaultQuery("rows", "24"))

	upgrader := getUpgrader()
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil { log.Printf("WebSocket upgrade failed: %v", err); return }
	defer ws.Close()

	if directSession == nil {
		writeWSMessage(ws, models.WSMessage{Type: "error", Data: "无效或已过期的会话"})
		return
	}

	port, _ := strconv.Atoi(directSession.Port)
	if port == 0 {
		port = 22
	}
	sshClient, err := createSSHClient(directSession.Host, port, directSession.Username, directSession.Password, directSession.PrivateKey)
	if err != nil {
		writeWSMessage(ws, models.WSMessage{Type: "error", Data: fmt.Sprintf("连接失败: %s", err.Error())})
		return
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		writeWSMessage(ws, models.WSMessage{Type: "error", Data: fmt.Sprintf("创建会话失败: %s", err.Error())})
		return
	}
	defer session.Close()

	termWidth := cols; if termWidth == 0 { termWidth = 80 }
	termHeight := rows; if termHeight == 0 { termHeight = 24 }

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
		writeWSMessage(ws, models.WSMessage{Type: "error", Data: fmt.Sprintf("请求终端失败: %s", err.Error())})
		return
	}

	stdin, err := session.StdinPipe()
	if err != nil { writeWSMessage(ws, models.WSMessage{Type: "error", Data: fmt.Sprintf("获取输入流失败: %s", err)}); return }
	stdout, err := session.StdoutPipe()
	if err != nil { writeWSMessage(ws, models.WSMessage{Type: "error", Data: fmt.Sprintf("获取输出流失败: %s", err)}); return }
	stderr, err := session.StderrPipe()
	if err != nil { writeWSMessage(ws, models.WSMessage{Type: "error", Data: fmt.Sprintf("获取错误流失败: %s", err)}); return }

	if err := session.Shell(); err != nil {
		writeWSMessage(ws, models.WSMessage{Type: "error", Data: fmt.Sprintf("启动Shell失败: %s", err.Error())})
		return
	}

	writeWSMessage(ws, models.WSMessage{
		Type: "connected",
		Data: fmt.Sprintf("已连接到 %s", directSession.Host),
	})

	newSessionID := utils.GenerateID()
	sshSession := &SSHSession{
		ID: newSessionID, ServerID: "direct", Client: sshClient, Session: session,
		Stdin: stdin, Connected: true, StartTime: time.Now(), CloseChan: make(chan struct{}),
	}
	h.mu.Lock(); h.sessions[newSessionID] = sshSession; h.mu.Unlock()
	defer func() { h.mu.Lock(); delete(h.sessions, newSessionID); h.mu.Unlock(); sshSession.Connected = false; close(sshSession.CloseChan) }()

	done := make(chan struct{})
	var closeOnce sync.Once
	closeDone := func() { closeOnce.Do(func() { close(done) }) }

	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := stdout.Read(buf)
			if err != nil { if err != io.EOF { log.Printf("SSH stdout read error: %v", err) }; closeDone(); return }
			if n > 0 { writeWSMessage(ws, models.WSMessage{Type: "output", Data: string(buf[:n])}) }
		}
	}()

	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := stderr.Read(buf)
			if err != nil { return }
			if n > 0 { writeWSMessage(ws, models.WSMessage{Type: "output", Data: string(buf[:n])}) }
		}
	}()

	go func() {
		for {
			_, message, err := ws.ReadMessage()
			if err != nil { closeDone(); return }
			var msg models.WSMessage
			if err := jsonUnmarshal(message, &msg); err != nil { continue }
			switch msg.Type {
			case "input":
				if _, err := stdin.Write([]byte(msg.Data)); err != nil { log.Printf("SSH stdin write error: %v", err); closeDone(); return }
			case "resize":
				if msg.Cols > 0 && msg.Rows > 0 { _ = session.WindowChange(msg.Rows, msg.Cols) }
			case "ping":
				writeWSMessage(ws, models.WSMessage{Type: "pong", Data: "ok"})
			}
		}
	}()

	select {
	case <-done:
	case <-sshSession.CloseChan:
	}
	writeWSMessage(ws, models.WSMessage{Type: "disconnected", Data: "连接已断开"})
}
