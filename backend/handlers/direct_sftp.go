package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"web-ssh/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type DirectSFTPHandler struct{}

func NewDirectSFTPHandler() *DirectSFTPHandler {
	return &DirectSFTPHandler{}
}

type DirectSFTPClient struct {
	Client    *sftp.Client
	SSHClient *ssh.Client
	Token     string
	CreatedAt time.Time
	mu        sync.RWMutex
}

var directSFTPClients = make(map[string]*DirectSFTPClient)
var directSFTPMutex sync.RWMutex

func (h *DirectSFTPHandler) HandleDirectSFTP(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "缺少token"})
		return
	}

	log.Printf("Direct SFTP request for token: %s", token)

	directSFTPMutex.RLock()
	ds, exists := directSFTPClients[token]
	directSFTPMutex.RUnlock()

	if !exists {
		log.Printf("SFTP client not found, creating new one for token: %s", token)

		directSessionsMutex.RLock()
		session, sessionExists := directSessions[token]
		directSessionsMutex.RUnlock()

		if !sessionExists {
			log.Printf("Session not found for token: %s", token)
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "会话不存在或已过期"})
			return
		}

		log.Printf("Creating SSH/SFTP client for host: %s:%s, user: %s", session.Host, session.Port, session.Username)

		portInt := 22
		if session.Port != "" {
			fmt.Sscanf(session.Port, "%d", &portInt)
		}

		sshClient, err := createSSHClient(session.Host, portInt, session.Username, session.Password, "")
		if err != nil {
			log.Printf("SSH connection failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": fmt.Sprintf("SSH连接失败: %s", err.Error())})
			return
		}

		sftpClient, err := sftp.NewClient(sshClient)
		if err != nil {
			sshClient.Close()
			log.Printf("SFTP client creation failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": fmt.Sprintf("SFTP连接失败: %s", err.Error())})
			return
		}

		ds = &DirectSFTPClient{
			Client:    sftpClient,
			SSHClient: sshClient,
			Token:     token,
			CreatedAt: time.Now(),
		}

		directSFTPMutex.Lock()
		directSFTPClients[token] = ds
		directSFTPMutex.Unlock()

		log.Printf("SFTP client created successfully for token: %s", token)
	}

	upgrader := getUpgrader()
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Direct SFTP WebSocket upgrade failed: %v", err)
		return
	}
	defer ws.Close()

	writeWSMessage(ws, models.WSMessage{Type: "connected", Data: "SFTP连接就绪"})

	done := make(chan struct{})
	var closeOnce sync.Once
	closeDone := func() { closeOnce.Do(func() { close(done) }) }

	go func() {
		defer closeDone()
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				return
			}
			var req models.SFTPRequest
			if err := jsonUnmarshal(message, &req); err != nil {
				writeWSMessage(ws, models.WSMessage{Type: "error", Data: "无效的请求格式"})
				continue
			}
			h.handleSFTPRequest(ws, ds.Client, &req)
		}
	}()

	<-done
}

func (h *DirectSFTPHandler) handleSFTPRequest(ws *websocket.Conn, client *sftp.Client, req *models.SFTPRequest) {
	var response models.SFTPResponse

	switch req.Action {
	case "list":
		response = h.listDirectory(client, req.Path)
	case "stat":
		response = h.getFileInfo(client, req.Path)
	case "download":
		response = h.downloadFile(client, req.Path)
	case "upload":
		response = h.uploadFile(client, req.Path, req.Content)
	case "mkdir":
		response = h.makeDir(client, req.Path)
	case "rm":
		response = h.removeItem(client, req.Path)
	case "rename":
		response = h.renameItem(client, req.Path, req.Dst)
	case "read":
		response = h.readFile(client, req.Path)
	default:
		response = models.SFTPResponse{Success: false, Message: "未知操作"}
	}

	data, _ := json.Marshal(response)
	ws.WriteMessage(websocket.TextMessage, data)
}

func (h *DirectSFTPHandler) listDirectory(client *sftp.Client, path string) models.SFTPResponse {
	if path == "" { path = "/" }
	log.Printf("Listing directory: %s", path)
	entries, err := client.ReadDir(path)
	if err != nil {
		log.Printf("ReadDir failed: %v", err)
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("读取目录失败: %s", err.Error())}
	}

	files := make([]models.FileInfo, 0, len(entries))
	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		info := models.FileInfo{
			Name: entry.Name(), Path: fullPath, Size: entry.Size(),
			Mode: entry.Mode().String(), ModTime: entry.ModTime(), IsDir: entry.IsDir(),
		}
		if entry.Mode()&0o120000 != 0 {
			info.IsSymlink = true
			if target, err := client.ReadLink(fullPath); err == nil {
				info.Name = entry.Name() + " -> " + target
			}
		}
		files = append(files, info)
	}

	return models.SFTPResponse{Success: true, Files: files}
}

func (h *DirectSFTPHandler) getFileInfo(client *sftp.Client, path string) models.SFTPResponse {
	info, err := client.Stat(path)
	if err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("获取文件信息失败: %s", err.Error())}
	}

	return models.SFTPResponse{Success: true, Files: []models.FileInfo{{
		Name: info.Name(), Path: path, Size: info.Size(),
		Mode: info.Mode().String(), ModTime: info.ModTime(), IsDir: info.IsDir(),
	}}}
}

func (h *DirectSFTPHandler) downloadFile(client *sftp.Client, path string) models.SFTPResponse {
	rf, err := client.OpenFile(path, syscall.O_RDONLY)
	if err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("打开文件失败: %s", err.Error())}
	}
	defer rf.Close()

	data, err := io.ReadAll(rf)
	if err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("读取文件失败: %s", err.Error())}
	}

	return models.SFTPResponse{Success: true, Data: base64.StdEncoding.EncodeToString(data)}
}

func (h *DirectSFTPHandler) uploadFile(client *sftp.Client, path string, content string) models.SFTPResponse {
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("文件内容解码失败: %s", err.Error())}
	}

	rf, err := client.Create(path)
	if err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("创建文件失败: %s", err.Error())}
	}
	defer rf.Close()

	if _, err := rf.Write(decoded); err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("写入文件失败: %s", err.Error())}
	}

	return models.SFTPResponse{Success: true, Message: "文件上传成功"}
}

func (h *DirectSFTPHandler) makeDir(client *sftp.Client, path string) models.SFTPResponse {
	if err := client.MkdirAll(path); err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("创建目录失败: %s", err.Error())}
	}
	return models.SFTPResponse{Success: true, Message: "目录创建成功"}
}

func (h *DirectSFTPHandler) removeItem(client *sftp.Client, path string) models.SFTPResponse {
	info, err := client.Stat(path)
	if err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("获取文件信息失败: %s", err.Error())}
	}

	var errDel error
	if info.IsDir() {
		errDel = client.RemoveDirectory(path)
	} else {
		errDel = client.Remove(path)
	}

	if errDel != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("删除失败: %s", errDel.Error())}
	}
	return models.SFTPResponse{Success: true, Message: "删除成功"}
}

func (h *DirectSFTPHandler) renameItem(client *sftp.Client, oldPath, newPath string) models.SFTPResponse {
	if err := client.Rename(oldPath, newPath); err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("重命名失败: %s", err.Error())}
	}
	return models.SFTPResponse{Success: true, Message: "重命名成功"}
}

func (h *DirectSFTPHandler) readFile(client *sftp.Client, path string) models.SFTPResponse {
	rf, err := client.OpenFile(path, syscall.O_RDONLY)
	if err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("打开文件失败: %s", err.Error())}
	}
	defer rf.Close()

	data, err := io.ReadAll(rf)
	if err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("读取文件失败: %s", err.Error())}
	}

	return models.SFTPResponse{Success: true, Data: base64.StdEncoding.EncodeToString(data)}
}

func CleanupDirectSFTPSessions() {
	for {
		directSFTPMutex.Lock()
		now := time.Now()
		for token, client := range directSFTPClients {
			if now.Sub(client.CreatedAt) > 30*time.Minute {
				client.Client.Close()
				client.SSHClient.Close()
				delete(directSFTPClients, token)
			}
		}
		directSFTPMutex.Unlock()
		time.Sleep(5 * time.Minute)
	}
}

func (h *DirectSFTPHandler) HandleDirectSFTPDownload(c *gin.Context) {
	token := c.Query("token")
	path := c.Query("path")

	if token == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "缺少参数"})
		return
	}

	directSFTPMutex.RLock()
	ds, exists := directSFTPClients[token]
	directSFTPMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "会话不存在或已过期"})
		return
	}

	rf, err := ds.Client.OpenFile(path, syscall.O_RDONLY)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "打开文件失败"})
		return
	}
	defer rf.Close()

	data, err := io.ReadAll(rf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "读取文件失败"})
		return
	}

	filename := filepath.Base(path)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Data(http.StatusOK, "application/octet-stream", data)
}
