package handlers

import (
	"log"
	"net/http"
	"time"

	"web-ssh/store"

	"github.com/gin-gonic/gin"
)

type ShareHandler struct {
	store store.Store
}

func NewShareHandler(s store.Store) *ShareHandler {
	return &ShareHandler{store: s}
}

func (h *ShareHandler) CreateShareToken(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未登录"})
		return
	}

	serverID := c.Param("id")
	if serverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "服务器ID无效"})
		return
	}

	server, exists := h.store.GetServer(serverID, userID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "服务器不存在"})
		return
	}

	password, err := h.store.DecryptServerPassword(serverID)
	if err != nil {
		password = server.Password
	}

	shareToken, err := h.store.CreateShareTokenWithCredentials(server.ID, userID, password, server.PrivateKey)
	if err != nil {
		log.Printf("CreateShareToken failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建分享链接失败: " + err.Error()})
		return
	}

	shareURL := "/share/" + shareToken.Token

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"share_url": shareURL,
		"token":     shareToken.Token,
		"expires":   shareToken.ExpiresAt.Format("2006-01-02 15:04:05"),
	})
}

func (h *ShareHandler) GetSharedServer(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的分享链接"})
		return
	}

	shareToken, err := h.store.GetShareToken(token)
	if err != nil || shareToken == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "分享链接不存在"})
		return
	}

	if shareToken.Used {
		c.JSON(http.StatusGone, gin.H{"success": false, "message": "此分享链接已被使用"})
		return
	}

	if time.Now().After(shareToken.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"success": false, "message": "此分享链接已过期"})
		return
	}

	server, exists := h.store.GetServerByID(shareToken.ServerID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "服务器不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"server_id":   server.ID,
		"server_name": server.Name,
		"host":        server.Host,
		"port":        server.Port,
		"username":    server.Username,
		"password":    shareToken.Password,
		"private_key": shareToken.PrivateKey,
		"share_token": token,
	})
}
