package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"web-ssh/models"
	"web-ssh/store"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"github.com/pkg/sftp"
)

type SFTPHandler struct {
	store   store.Store
	clients map[string]*sftpClient
	mu      sync.RWMutex
}

type sftpClient struct {
	Client    *sftp.Client
	SSHClient *ssh.Client
	LastUsed  time.Time
}

func NewSFTPHandler(s store.Store) *SFTPHandler {
	handler := &SFTPHandler{
		store:   s,
		clients: make(map[string]*sftpClient),
	}
	go handler.cleanup()
	return handler
}

func (h *SFTPHandler) getSFTPClient(serverID string) (*sftpClient, error) {
	h.mu.RLock()
	client, ok := h.clients[serverID]
	h.mu.RUnlock()

	if ok {
		client.LastUsed = time.Now()
		return client, nil
	}

	server, ok := h.store.GetServerByID(serverID)
	if !ok {
		return nil, fmt.Errorf("服务器不存在")
	}

	password, err := h.store.DecryptServerPassword(serverID)
	if err != nil {
		return nil, fmt.Errorf("解密密码失败")
	}

	sshClient, err := createSSHClient(server.Host, server.Port, server.Username, password, server.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("SSH连接失败: %w", err)
	}

	sftpConn, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("SFTP连接失败: %w", err)
	}

	client = &sftpClient{
		Client:    sftpConn,
		SSHClient: sshClient,
		LastUsed:  time.Now(),
	}

	h.mu.Lock()
	h.clients[serverID] = client
	h.mu.Unlock()

	return client, nil
}

func (h *SFTPHandler) HandleSFTPWebSocket(c *gin.Context) {
	serverID := c.Query("server_id")
	if serverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "缺少服务器ID"})
		return
	}

	upgrader := getUpgrader()
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("SFTP WebSocket upgrade failed: %v", err)
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
			h.handleSFTPRequest(ws, serverID, &req)
		}
	}()

	<-done
}

func (h *SFTPHandler) handleSFTPRequest(ws *websocket.Conn, serverID string, req *models.SFTPRequest) {
	client, err := h.getSFTPClient(serverID)
	if err != nil {
		sendSFTPError(ws, fmt.Sprintf("获取SFTP连接失败: %s", err))
		return
	}

	var response models.SFTPResponse

	switch req.Action {
	case "list":
		response = h.listDirectory(client.Client, req.Path)
	case "stat":
		response = h.getFileInfo(client.Client, req.Path)
	case "mkdir":
		response = h.makeDirectory(client.Client, req.Path)
	case "rm":
		response = h.removePath(client.Client, req.Path)
	case "rename":
		response = h.renamePath(client.Client, req.Path, req.Dst)
	case "download":
		response = h.downloadFile(client.Client, req.Path)
	case "upload":
		response = h.uploadFile(client.Client, req.Path, req.Content, req.Name)
	case "read":
		response = h.readFile(client.Client, req.Path)
	case "write":
		response = h.writeFile(client.Client, req.Path, req.Content)
	case "search":
		response = h.searchFiles(client.Client, req.Path, req.Name)
	default:
		response = models.SFTPResponse{Success: false, Message: "不支持的操作"}
	}

	data, _ := json.Marshal(response)
	ws.WriteMessage(websocket.TextMessage, data)
}

func (h *SFTPHandler) listDirectory(client *sftp.Client, path string) models.SFTPResponse {
	if path == "" { path = "/" }
	entries, err := client.ReadDir(path)
	if err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("读取目录失败: %s", err)}
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

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir { return files[i].IsDir }
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	return models.SFTPResponse{Success: true, Files: files}
}

func (h *SFTPHandler) getFileInfo(client *sftp.Client, path string) models.SFTPResponse {
	info, err := client.Stat(path)
	if err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("获取文件信息失败: %s", err)}
	}
	return models.SFTPResponse{
		Success: true,
		Files: []models.FileInfo{{Name: info.Name(), Path: path, Size: info.Size(),
			Mode: info.Mode().String(), ModTime: info.ModTime(), IsDir: info.IsDir()}},
	}
}

func (h *SFTPHandler) makeDirectory(client *sftp.Client, path string) models.SFTPResponse {
	if err := client.MkdirAll(path); err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("创建目录失败: %s", err)}
	}
	return models.SFTPResponse{Success: true, Message: "目录创建成功"}
}

func (h *SFTPHandler) removePath(client *sftp.Client, path string) models.SFTPResponse {
	info, err := client.Stat(path)
	if err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("文件不存在: %s", err)}
	}
	if info.IsDir() {
		if err := h.removeDirectory(client, path); err != nil {
			return models.SFTPResponse{Success: false, Message: fmt.Sprintf("删除目录失败: %s", err)}
		}
	} else {
		if err := client.Remove(path); err != nil {
			return models.SFTPResponse{Success: false, Message: fmt.Sprintf("删除文件失败: %s", err)}
		}
	}
	return models.SFTPResponse{Success: true, Message: "删除成功"}
}

func (h *SFTPHandler) removeDirectory(client *sftp.Client, path string) error {
	entries, err := client.ReadDir(path)
	if err != nil { return err }
	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		if entry.IsDir() { h.removeDirectory(client, fullPath) } else { client.Remove(fullPath) }
	}
	return client.RemoveDirectory(path)
}

func (h *SFTPHandler) renamePath(client *sftp.Client, oldPath, newPath string) models.SFTPResponse {
	if err := client.Rename(oldPath, newPath); err != nil {
		return models.SFTPResponse{Success: false, Message: fmt.Sprintf("重命名失败: %s", err)}
	}
	return models.SFTPResponse{Success: true, Message: "重命名成功"}
}

func (h *SFTPHandler) downloadFile(client *sftp.Client, path string) models.SFTPResponse {
	info, err := client.Stat(path)
	if err != nil { return models.SFTPResponse{Success: false, Message: fmt.Sprintf("文件不存在: %s", err)} }
	if info.IsDir() { return models.SFTPResponse{Success: false, Message: "不能下载目录，请选择文件"} }
	if info.Size() > 50*1024*1024 { return models.SFTPResponse{Success: false, Message: "文件太大，最大支持50MB"} }

	file, err := client.Open(path)
	if err != nil { return models.SFTPResponse{Success: false, Message: fmt.Sprintf("打开文件失败: %s", err)} }
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil { return models.SFTPResponse{Success: false, Message: fmt.Sprintf("读取文件失败: %s", err)} }
	return models.SFTPResponse{Success: true, Data: base64.StdEncoding.EncodeToString(data)}
}

func (h *SFTPHandler) uploadFile(client *sftp.Client, dirPath, content, filename string) models.SFTPResponse {
	if filename == "" { return models.SFTPResponse{Success: false, Message: "文件名不能为空"} }
	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil { return models.SFTPResponse{Success: false, Message: fmt.Sprintf("文件内容解码失败: %s", err)} }

	fullPath := filepath.Join(dirPath, filename)
	file, err := client.Create(fullPath)
	if err != nil { return models.SFTPResponse{Success: false, Message: fmt.Sprintf("创建文件失败: %s", err)} }
	defer file.Close()

	if _, err := file.Write(data); err != nil { return models.SFTPResponse{Success: false, Message: fmt.Sprintf("写入文件失败: %s", err)} }
	return models.SFTPResponse{Success: true, Message: "文件上传成功"}
}

func (h *SFTPHandler) readFile(client *sftp.Client, path string) models.SFTPResponse {
	info, err := client.Stat(path)
	if err != nil { return models.SFTPResponse{Success: false, Message: fmt.Sprintf("文件不存在: %s", err)} }
	if info.IsDir() { return models.SFTPResponse{Success: false, Message: "不能读取目录"} }
	if info.Size() > 2*1024*1024 { return models.SFTPResponse{Success: false, Message: "文件太大，文本编辑最大支持2MB"} }

	file, err := client.Open(path)
	if err != nil { return models.SFTPResponse{Success: false, Message: fmt.Sprintf("打开文件失败: %s", err)} }
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil { return models.SFTPResponse{Success: false, Message: fmt.Sprintf("读取文件失败: %s", err)} }
	return models.SFTPResponse{Success: true, Data: base64.StdEncoding.EncodeToString(data)}
}

func (h *SFTPHandler) writeFile(client *sftp.Client, path, content string) models.SFTPResponse {
	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil { return models.SFTPResponse{Success: false, Message: fmt.Sprintf("内容解码失败: %s", err)} }

	file, err := client.Create(path)
	if err != nil { return models.SFTPResponse{Success: false, Message: fmt.Sprintf("创建文件失败: %s", err)} }
	defer file.Close()

	if _, err := file.Write(data); err != nil { return models.SFTPResponse{Success: false, Message: fmt.Sprintf("写入文件失败: %s", err)} }
	return models.SFTPResponse{Success: true, Message: "文件保存成功"}
}

func (h *SFTPHandler) searchFiles(client *sftp.Client, dirPath, pattern string) models.SFTPResponse {
	var results []models.FileInfo
	walker := client.Walk(dirPath)
	for walker.Step() {
		if walker.Err() != nil { continue }
		entry := walker.Stat()
		if entry == nil { continue }
		if strings.Contains(strings.ToLower(entry.Name()), strings.ToLower(pattern)) {
			results = append(results, models.FileInfo{
				Name: entry.Name(), Path: walker.Path(), Size: entry.Size(),
				Mode: entry.Mode().String(), ModTime: entry.ModTime(), IsDir: entry.IsDir(),
			})
		}
		if len(results) >= 100 { break }
	}
	return models.SFTPResponse{Success: true, Files: results}
}

func (h *SFTPHandler) HandleFileUpload(c *gin.Context) {
	serverID := c.PostForm("server_id")
	path := c.PostForm("path")
	if serverID == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "缺少参数"})
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil { c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "获取上传文件失败"}); return }
	defer file.Close()

	client, err := h.getSFTPClient(serverID)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()}); return }

	fullPath := filepath.Join(path, header.Filename)
	dst, err := client.Client.Create(fullPath)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建文件失败"}); return }
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "写入文件失败"}); return }
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "文件上传成功", "path": fullPath, "size": header.Size})
}

func (h *SFTPHandler) HandleFileDownload(c *gin.Context) {
	serverID := c.Query("server_id")
	path := c.Query("path")
	if serverID == "" || path == "" { c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "缺少参数"}); return }

	client, err := h.getSFTPClient(serverID)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()}); return }

	info, err := client.Client.Stat(path)
	if err != nil { c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "文件不存在"}); return }
	if info.IsDir() { c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "不能下载目录"}); return }

	file, err := client.Client.Open(path)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "打开文件失败"}); return }
	defer file.Close()

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(path)))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
	io.Copy(c.Writer, file)
}

func (h *SFTPHandler) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		h.mu.Lock()
		for id, client := range h.clients {
			if time.Since(client.LastUsed) > 30*time.Minute {
				client.Client.Close(); client.SSHClient.Close(); delete(h.clients, id)
			}
		}
		h.mu.Unlock()
	}
}

func sendSFTPError(ws *websocket.Conn, message string) {
	response := models.SFTPResponse{Success: false, Message: message}
	data, _ := json.Marshal(response)
	ws.WriteMessage(websocket.TextMessage, data)
}
