package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"web-ssh/store"

	"github.com/gin-gonic/gin"
)

type DirectConnectHandler struct {
	store store.Store
}

func NewDirectConnectHandler(s store.Store) *DirectConnectHandler {
	return &DirectConnectHandler{store: s}
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (h *DirectConnectHandler) Connect(c *gin.Context) {
	var req struct {
		Host       string `json:"host"`
		Port       int    `json:"port"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		PrivateKey string `json:"private_key"`
		ShareToken string `json:"share_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		host := c.Query("host")
		port := c.Query("port")
		username := c.Query("username")
		password := c.Query("password")

		if host == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "host参数不能为空"})
			return
		}
		if username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "username参数不能为空"})
			return
		}
		if password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "password参数不能为空"})
			return
		}

		if port == "" {
			port = "22"
		}

		token := generateToken()
		session := &DirectSession{
			Token:     token,
			Host:      host,
			Port:      port,
			Username:  username,
			Password:  password,
			CreatedAt: time.Now(),
		}

		directSessionsMutex.Lock()
		directSessions[token] = session
		directSessionsMutex.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"token":   token,
			"message": "会话已创建",
		})
		return
	}

	if req.Host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "host参数不能为空"})
		return
	}
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "username参数不能为空"})
		return
	}
	if req.Password == "" && req.PrivateKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "password或private_key参数不能为空"})
		return
	}

	if req.ShareToken != "" && h.store != nil {
		shareToken, err := h.store.GetShareToken(req.ShareToken)
		if err == nil && shareToken != nil {
			if !shareToken.Used && time.Now().Before(shareToken.ExpiresAt) {
				h.store.UseShareToken(req.ShareToken)
			}
		}
	}

	port := "22"
	if req.Port > 0 {
		port = fmt.Sprintf("%d", req.Port)
	}

	token := generateToken()
	session := &DirectSession{
		Token:      token,
		Host:       req.Host,
		Port:       port,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: req.PrivateKey,
		CreatedAt:  time.Now(),
	}

	directSessionsMutex.Lock()
	directSessions[token] = session
	directSessionsMutex.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"message": "会话已创建",
	})
}

func (h *DirectConnectHandler) GetSession(c *gin.Context) {
	token := c.Param("token")

	directSessionsMutex.Lock()
	session, exists := directSessions[token]
	directSessionsMutex.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "会话不存在或已过期"})
		return
	}

	if time.Since(session.CreatedAt) > 5*time.Minute {
		directSessionsMutex.Lock()
		delete(directSessions, token)
		directSessionsMutex.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "会话已过期"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"host":        session.Host,
		"port":        session.Port,
		"username":    session.Username,
		"password":    session.Password,
		"private_key": session.PrivateKey,
	})
}

type DirectSession struct {
	Token      string
	Host       string
	Port       string
	Username   string
	Password   string
	PrivateKey string
	CreatedAt  time.Time
}

func GetDirectSession(token string) *DirectSession {
	directSessionsMutex.Lock()
	defer directSessionsMutex.Unlock()
	return directSessions[token]
}
