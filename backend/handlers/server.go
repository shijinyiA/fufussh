package handlers

import (
	"net/http"

	"web-ssh/models"
	"web-ssh/store"
	"web-ssh/utils"

	"github.com/gin-gonic/gin"
)

type ServerHandler struct {
	store store.Store
}

func NewServerHandler(s store.Store) *ServerHandler {
	return &ServerHandler{store: s}
}

func (h *ServerHandler) ListServers(c *gin.Context) {
	userID := getUserID(c)
	servers := h.store.GetServers(userID)

	type ServerResponse struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Host        string   `json:"host"`
		Port        int      `json:"port"`
		Username    string   `json:"username"`
		HasPassword bool     `json:"has_password"`
		HasKey      bool     `json:"has_key"`
		Group       string   `json:"group"`
		Tags        []string `json:"tags"`
		Description string   `json:"description"`
		Color       string   `json:"color"`
		CreatedAt   string   `json:"created_at"`
		UpdatedAt   string   `json:"updated_at"`
	}

	response := make([]ServerResponse, 0, len(servers))
	for _, s := range servers {
		response = append(response, ServerResponse{
			ID:          s.ID,
			Name:        s.Name,
			Host:        s.Host,
			Port:        s.Port,
			Username:    s.Username,
			HasPassword: s.Password != "",
			HasKey:      s.PrivateKey != "",
			Group:       s.Group,
			Tags:        s.Tags,
			Description: s.Description,
			Color:       s.Color,
			CreatedAt:   s.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   s.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"servers": response,
	})
}

func (h *ServerHandler) GetServer(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)
	server, ok := h.store.GetServer(id, userID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "服务器不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"server": gin.H{
			"id":          server.ID,
			"name":        server.Name,
			"host":        server.Host,
			"port":        server.Port,
			"username":    server.Username,
			"has_password": server.Password != "",
			"has_key":     server.PrivateKey != "",
			"group":       server.Group,
			"tags":        server.Tags,
			"description": server.Description,
			"color":       server.Color,
			"created_at":  server.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":  server.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

func (h *ServerHandler) CreateServer(c *gin.Context) {
	var req struct {
		Name        string   `json:"name" binding:"required"`
		Host        string   `json:"host" binding:"required"`
		Port        int      `json:"port"`
		Username    string   `json:"username" binding:"required"`
		Password    string   `json:"password"`
		PrivateKey  string   `json:"private_key"`
		Group       string   `json:"group"`
		Tags        []string `json:"tags"`
		Description string   `json:"description"`
		Color       string   `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}

	if req.Port == 0 {
		req.Port = 22
	}
	if req.Color == "" {
		req.Color = utils.RandomColor()
	}

	server := &models.Server{
		ID:         utils.GenerateID(),
		UserID:     getUserID(c),
		Name:       req.Name,
		Host:       req.Host,
		Port:       req.Port,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: req.PrivateKey,
		Group:      req.Group,
		Tags:       req.Tags,
		Description: req.Description,
		Color:      req.Color,
	}

	if err := h.store.CreateServer(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建服务器失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "服务器创建成功",
		"id":      server.ID,
	})
}

func (h *ServerHandler) UpdateServer(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)

	server, ok := h.store.GetServer(id, userID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "服务器不存在"})
		return
	}

	var req struct {
		Name        string   `json:"name"`
		Host        string   `json:"host"`
		Port        int      `json:"port"`
		Username    string   `json:"username"`
		Password    string   `json:"password"`
		PrivateKey  string   `json:"private_key"`
		Group       string   `json:"group"`
		Tags        []string `json:"tags"`
		Description string   `json:"description"`
		Color       string   `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}

	if req.Name != "" { server.Name = req.Name }
	if req.Host != "" { server.Host = req.Host }
	if req.Port != 0 { server.Port = req.Port }
	if req.Username != "" { server.Username = req.Username }
	if req.Password != "" { server.Password = req.Password }
	if req.PrivateKey != "" { server.PrivateKey = req.PrivateKey }
	server.Group = req.Group
	if req.Tags != nil { server.Tags = req.Tags }
	server.Description = req.Description
	if req.Color != "" { server.Color = req.Color }

	if err := h.store.UpdateServer(server); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新服务器失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "服务器更新成功"})
}

func (h *ServerHandler) DeleteServer(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)

	if err := h.store.DeleteServer(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除服务器失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "服务器删除成功"})
}

func (h *ServerHandler) TestConnection(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)

	server, ok := h.store.GetServer(id, userID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "服务器不存在"})
		return
	}

	password, err := h.store.DecryptServerPassword(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "解密密码失败"})
		return
	}

	sshClient, err := createSSHClient(server.Host, server.Port, server.Username, password, server.PrivateKey)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "连接失败: " + err.Error()})
		return
	}
	defer sshClient.Close()

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "连接成功"})
}

func (h *ServerHandler) GetGroups(c *gin.Context) {
	userID := getUserID(c)
	groups := h.store.GetUserGroups(userID)

	type GroupInfo struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	result := make([]GroupInfo, 0, len(groups))
	for name, count := range groups {
		result = append(result, GroupInfo{Name: name, Count: count})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "groups": result})
}
