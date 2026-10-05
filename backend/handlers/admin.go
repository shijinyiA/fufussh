package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"web-ssh/store"
	"web-ssh/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	db        *sql.DB
	mu        sync.RWMutex
	sessions  map[string]*adminSession
	uploadDir string
}

type adminSession struct {
	Username        string
	PasswordChanged bool
	ExpiresAt       time.Time
}

type adminUserRow struct {
	ID              int
	Username        string
	Password        string
	PasswordChanged bool
}

var safeIdentPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

func NewAdminHandler(s *store.MySQLStore) *AdminHandler {
	return &AdminHandler{
		db:        s.DB(),
		sessions:  make(map[string]*adminSession),
		uploadDir: "./uploads",
	}
}

func (h *AdminHandler) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未登录"})
			c.Abort()
			return
		}

		h.mu.RLock()
		session, ok := h.sessions[token]
		h.mu.RUnlock()
		if !ok || time.Now().After(session.ExpiresAt) {
			if ok {
				h.mu.Lock()
				delete(h.sessions, token)
				h.mu.Unlock()
			}
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "登录已过期"})
			c.Abort()
			return
		}

		session.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
		c.Set("admin_username", session.Username)
		c.Set("admin_password_changed", session.PasswordChanged)
		c.Next()
	}
}

func (h *AdminHandler) Status(c *gin.Context) {
	var count int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM admin_users").Scan(&count)
	c.JSON(http.StatusOK, gin.H{"success": true, "initialized": count > 0})
}

func (h *AdminHandler) Init(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请输入管理员用户名和密码"})
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if len(req.Username) < 2 || len(req.Username) > 64 || len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "用户名需 2-64 位，密码至少 6 位"})
		return
	}

	var count int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM admin_users").Scan(&count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "检查管理员状态失败"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": "管理员已初始化"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "密码加密失败"})
		return
	}
	if _, err := h.db.Exec(
		"INSERT INTO admin_users (username, password, password_changed) VALUES (?, ?, 1)",
		req.Username, string(hash),
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建管理员失败"})
		return
	}

	h.writeLog(req.Username, "info", "init", "初始化管理员账户", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "管理员账户创建成功"})
}

func (h *AdminHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请输入用户名和密码"})
		return
	}

	admin, err := h.findAdmin(strings.TrimSpace(req.Username))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户名或密码错误"})
		return
	}

	token := utils.GenerateSessionToken()
	h.mu.Lock()
	h.sessions[token] = &adminSession{
		Username:        admin.Username,
		PasswordChanged: admin.PasswordChanged,
		ExpiresAt:       time.Now().Add(7 * 24 * time.Hour),
	}
	h.mu.Unlock()

	h.writeLog(admin.Username, "info", "login", "管理员登录", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"token":            token,
		"username":         admin.Username,
		"password_changed": admin.PasswordChanged,
	})
}

func (h *AdminHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"username":         currentAdmin(c),
		"password_changed": c.GetBool("admin_password_changed"),
	})
}

func (h *AdminHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请输入原密码和新密码"})
		return
	}

	username := currentAdmin(c)
	admin, err := h.findAdmin(username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.OldPassword)) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "原密码错误"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "密码加密失败"})
		return
	}
	if _, err := h.db.Exec("UPDATE admin_users SET password=?, password_changed=1 WHERE username=?", string(hash), username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "修改密码失败"})
		return
	}

	h.updateSessionPasswordChanged(username, true)
	h.logAction(c, "info", "change_password", "修改管理员密码")
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "密码修改成功"})
}

func (h *AdminHandler) Stats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"user_count":    h.countTable("users"),
		"server_count":  h.countTable("servers"),
		"admin_count":   h.countTable("admin_users"),
		"command_count": h.countTable("quick_commands"),
	})
}

func (h *AdminHandler) SystemInfo(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	totalMB := float64(m.Sys) / 1024 / 1024
	usedMB := float64(m.Alloc) / 1024 / 1024
	usagePercent := int(usedMB / totalMB * 100)

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"version":      "V2.1.0 LTS",
		"memory_total": totalMB,
		"memory_used":  usedMB,
		"memory_usage": usagePercent,
	})
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, pageSize, offset := pageParams(c, 20, 100)
	rows, err := h.db.Query(`
		SELECT id, COALESCE(username, ''), COALESCE(avatar, ''), created_at
		FROM users ORDER BY created_at DESC, id ASC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取用户列表失败"})
		return
	}
	defer rows.Close()

	data := make([]gin.H, 0)
	for rows.Next() {
		var id, username, avatar string
		var createdAt sql.NullTime
		if rows.Scan(&id, &username, &avatar, &createdAt) == nil {
			data = append(data, gin.H{
				"id":         id,
				"username":   username,
				"avatar":     avatar,
				"created_at": formatNullTime(createdAt),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data, "total": h.countTable("users"), "page": page})
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req struct {
		ID       string `json:"id" binding:"required"`
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请输入用户ID和密码"})
		return
	}
	req.ID = strings.TrimSpace(req.ID)
	if len(req.ID) < 2 || len(req.ID) > 64 || len(req.Password) < 4 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "用户ID需 2-64 位，密码至少 4 位"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "密码加密失败"})
		return
	}
	if _, err := h.db.Exec(
		"INSERT INTO users (id, username, avatar, password) VALUES (?, ?, ?, ?)",
		req.ID, req.Username, req.Avatar, string(hash),
	); err != nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": "用户已存在或创建失败"})
		return
	}

	h.logAction(c, "info", "create_user", "创建用户 "+req.ID)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "用户创建成功"})
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}
	if _, err := h.db.Exec("UPDATE users SET username=?, avatar=? WHERE id=?", req.Username, req.Avatar, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新用户失败"})
		return
	}
	h.logAction(c, "info", "update_user", "更新用户 "+userID)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "用户已更新"})
}

func (h *AdminHandler) ResetUserPassword(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Password) < 4 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "新密码至少 4 位"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "密码加密失败"})
		return
	}
	if _, err := h.db.Exec("UPDATE users SET password=? WHERE id=?", string(hash), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "重置密码失败"})
		return
	}
	h.logAction(c, "info", "reset_user_password", "重置用户密码 "+userID)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "密码已重置"})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "开启事务失败"})
		return
	}
	defer tx.Rollback()

	_, _ = tx.Exec("DELETE FROM share_tokens WHERE user_id=?", userID)
	_, _ = tx.Exec("DELETE FROM terminal_settings WHERE user_id=?", userID)
	_, _ = tx.Exec("DELETE FROM servers WHERE user_id=?", userID)
	if _, err := tx.Exec("DELETE FROM users WHERE id=?", userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除用户失败"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "提交事务失败"})
		return
	}

	h.logAction(c, "warning", "delete_user", "删除用户及其服务器 "+userID)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "用户已删除"})
}

func (h *AdminHandler) ListServers(c *gin.Context) {
	page, pageSize, offset := pageParams(c, 20, 100)
	rows, err := h.db.Query(`
		SELECT id, user_id, name, host, port, username, COALESCE(group_name, ''), created_at
		FROM servers ORDER BY created_at DESC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取服务器列表失败"})
		return
	}
	defer rows.Close()

	data := make([]gin.H, 0)
	for rows.Next() {
		var id, userID, name, host, username, group string
		var port int
		var createdAt sql.NullTime
		if rows.Scan(&id, &userID, &name, &host, &port, &username, &group, &createdAt) == nil {
			data = append(data, gin.H{
				"id":         id,
				"user_id":    userID,
				"name":       name,
				"host":       host,
				"port":       port,
				"username":   username,
				"group":      group,
				"created_at": formatNullTime(createdAt),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data, "total": h.countTable("servers"), "page": page})
}

func (h *AdminHandler) DeleteServer(c *gin.Context) {
	serverID := c.Param("id")
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "开启事务失败"})
		return
	}
	defer tx.Rollback()

	_, _ = tx.Exec("DELETE FROM share_tokens WHERE server_id=?", serverID)
	_, _ = tx.Exec("DELETE FROM terminal_settings WHERE server_id=?", serverID)
	if _, err := tx.Exec("DELETE FROM servers WHERE id=?", serverID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除服务器失败"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "提交事务失败"})
		return
	}

	h.logAction(c, "warning", "delete_server", "删除服务器 "+serverID)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "服务器已删除"})
}

func (h *AdminHandler) GetSettings(c *gin.Context) {
	settings, err := h.loadSystemSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "读取设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": settings})
}

func (h *AdminHandler) SaveSettings(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	allowed := map[string]bool{
		"site_title": true, "site_description": true, "site_keywords": true, "site_author": true,
		"logo_url": true, "favicon_url": true, "background_url": true, "theme_color": true,
		"custom_css": true, "custom_js": true, "footer_text": true,
		"allow_register": true, "max_sessions": true, "session_timeout": true, "border_radius": true,
		"captcha_enabled": true, "captcha_id": true, "captcha_key": true, "background_blur": true,
	}
	sets := make([]string, 0)
	args := make([]interface{}, 0)
	for key, value := range req {
		if allowed[key] {
			sets = append(sets, key+"=?")
			switch v := value.(type) {
			case bool:
				if v { args = append(args, 1) } else { args = append(args, 0) }
			case float64:
				args = append(args, int(v))
			case string:
				args = append(args, v)
			default:
				args = append(args, fmt.Sprint(v))
			}
		}
	}
	if len(sets) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "没有需要保存的设置"})
		return
	}
	args = append(args, 1)
	if _, err := h.db.Exec("UPDATE system_settings SET "+strings.Join(sets, ", ")+" WHERE id=?", args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存设置失败"})
		return
	}

	h.logAction(c, "info", "save_settings", "保存系统设置")
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "设置已保存"})
}

func (h *AdminHandler) GetCommands(c *gin.Context) {
	rows, err := h.db.Query("SELECT id, name, command, sort_order FROM quick_commands ORDER BY sort_order, id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "读取快捷命令失败"})
		return
	}
	defer rows.Close()

	commands := make([]gin.H, 0)
	for rows.Next() {
		var id, sortOrder int
		var name, command string
		if rows.Scan(&id, &name, &command, &sortOrder) == nil {
			commands = append(commands, gin.H{"id": id, "name": name, "command": command, "sort_order": sortOrder})
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": commands})
}

func (h *AdminHandler) SaveCommands(c *gin.Context) {
	var req struct {
		Commands []struct {
			Name    string `json:"name"`
			Command string `json:"command"`
		} `json:"commands"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "开启事务失败"})
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM quick_commands"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "清空快捷命令失败"})
		return
	}
	for i, cmd := range req.Commands {
		name := strings.TrimSpace(cmd.Name)
		command := strings.TrimSpace(cmd.Command)
		if name == "" || command == "" {
			continue
		}
		if _, err := tx.Exec("INSERT INTO quick_commands (name, command, sort_order) VALUES (?, ?, ?)", name, command, (i+1)*10); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存快捷命令失败"})
			return
		}
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "提交事务失败"})
		return
	}

	h.logAction(c, "info", "save_commands", fmt.Sprintf("保存 %d 条快捷命令", len(req.Commands)))
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "快捷命令已保存"})
}

func (h *AdminHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请选择上传文件"})
		return
	}
	fileType := c.DefaultPostForm("type", "general")
	if !safeIdentPattern.MatchString(fileType) {
		fileType = "general"
	}

	maxSize := int64(5 << 20)
	allowedExt := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".svg": true, ".ico": true}
	if fileType == "favicon" {
		maxSize = 1 << 20
		allowedExt = map[string]bool{".png": true, ".ico": true, ".svg": true}
	}
	if fileType == "logo" {
		maxSize = 2 << 20
	}
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "文件大小超出限制"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExt[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "不支持的文件格式"})
		return
	}

	dir := filepath.Join(h.uploadDir, fileType)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建上传目录失败"})
		return
	}
	name := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), utils.GenerateID()[:8], ext)
	dst := filepath.Join(dir, name)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存文件失败"})
		return
	}

	url := "/uploads/" + fileType + "/" + name
	h.logAction(c, "info", "upload", "上传资源 "+url)
	c.JSON(http.StatusOK, gin.H{"success": true, "url": url})
}

func (h *AdminHandler) ListLogs(c *gin.Context) {
	_, pageSize, offset := pageParams(c, 50, 200)
	rows, err := h.db.Query(`
		SELECT id, COALESCE(admin_username, ''), level, action, message, COALESCE(ip, ''), created_at
		FROM admin_logs ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "读取日志失败"})
		return
	}
	defer rows.Close()

	data := make([]gin.H, 0)
	for rows.Next() {
		var id int64
		var username, level, action, message, ip string
		var createdAt sql.NullTime
		if rows.Scan(&id, &username, &level, &action, &message, &ip, &createdAt) == nil {
			data = append(data, gin.H{
				"id":             id,
				"admin_username": username,
				"level":          level,
				"action":         action,
				"message":        message,
				"ip":             ip,
				"created_at":     formatNullTime(createdAt),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data, "total": h.countTable("admin_logs")})
}

func (h *AdminHandler) ListDatabaseTables(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT TABLE_NAME, COALESCE(TABLE_ROWS, 0)
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		ORDER BY TABLE_NAME`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "读取数据表失败"})
		return
	}
	defer rows.Close()

	data := make([]gin.H, 0)
	for rows.Next() {
		var name string
		var estimate int64
		if rows.Scan(&name, &estimate) == nil {
			data = append(data, gin.H{"name": name, "rows": estimate})
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "tables": data})
}

func (h *AdminHandler) BrowseDatabaseTable(c *gin.Context) {
	table := c.Param("table")
	if !h.tableExists(table) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "数据表不存在"})
		return
	}

	_, pageSize, offset := pageParams(c, 50, 100)
	columns, primaryKey, err := h.tableColumns(table)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "读取表结构失败"})
		return
	}

	rows, err := h.db.Query("SELECT * FROM "+quoteIdent(table)+" LIMIT ? OFFSET ?", pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "读取表数据失败"})
		return
	}
	defer rows.Close()

	data, err := rowsToMaps(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "解析表数据失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"table":       table,
		"columns":     columns,
		"primary_key": primaryKey,
		"rows":        data,
		"total":       h.countTable(table),
	})
}

func (h *AdminHandler) QueryDatabase(c *gin.Context) {
	var req struct {
		SQL string `json:"sql" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请输入 SQL"})
		return
	}
	query, ok := readonlySQL(req.SQL)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "这里只允许 SELECT / SHOW / DESCRIBE / EXPLAIN 只读查询"})
		return
	}

	rows, err := h.db.Query(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer rows.Close()

	data, err := rowsToMaps(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "解析查询结果失败"})
		return
	}
	cols, _ := rows.Columns()
	h.logAction(c, "info", "database_query", "执行只读数据库查询")
	c.JSON(http.StatusOK, gin.H{"success": true, "columns": cols, "rows": data})
}

func (h *AdminHandler) DeleteDatabaseRow(c *gin.Context) {
	table := c.Param("table")
	var req struct {
		PrimaryKey string      `json:"primary_key"`
		Value      interface{} `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}
	if !h.tableExists(table) || !h.isPrimaryColumn(table, req.PrimaryKey) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "表或主键不合法"})
		return
	}
	result, err := h.db.Exec("DELETE FROM "+quoteIdent(table)+" WHERE "+quoteIdent(req.PrimaryKey)+" = ? LIMIT 1", req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除数据失败"})
		return
	}
	affected, _ := result.RowsAffected()
	h.logAction(c, "warning", "database_delete_row", fmt.Sprintf("删除 %s 表记录，影响 %d 行", table, affected))
	c.JSON(http.StatusOK, gin.H{"success": true, "affected": affected})
}

func (h *AdminHandler) DatabaseStatus(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT TABLE_NAME,
			   COALESCE(TABLE_ROWS, 0),
			   ROUND(COALESCE(DATA_LENGTH + INDEX_LENGTH, 0) / 1024, 1),
			   CREATE_TIME
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		ORDER BY TABLE_NAME`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "读取数据库信息失败: " + err.Error()})
		return
	}
	defer rows.Close()

	tables := make([]gin.H, 0)
	totalRows := int64(0)
	totalSize := float64(0)

	for rows.Next() {
		var name string
		var rowCount int64
		var sizeKb float64
		var createTime sql.NullTime

		if err := rows.Scan(&name, &rowCount, &sizeKb, &createTime); err != nil {
			continue
		}
		tables = append(tables, gin.H{
			"name":       name,
			"rows":       rowCount,
			"size":       sizeKb,
			"created_at": formatNullTime(createTime),
		})
		totalRows += rowCount
		totalSize += sizeKb
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"type":        "MySQL",
		"tables":      tables,
		"table_count": len(tables),
		"total_rows":  totalRows,
		"total_size":  math.Round(totalSize/1024*10) / 10,
	})
}

func (h *AdminHandler) BackupDatabase(c *gin.Context) {
	dsn := os.Getenv("MYSQL_DSN")
	dbName := "webssh"
	var dbUser, dbPass, dbHost string
	dbPort := 3306

	if dsn != "" {
		re := regexp.MustCompile(`^([^:]+):([^@]+)@tcp\(([^:]+):(\d+)\)/([^?]+)`)
		m := re.FindStringSubmatch(dsn)
		if len(m) >= 6 {
			dbUser = m[1]
			dbPass = m[2]
			dbHost = m[3]
			dbPort, _ = strconv.Atoi(m[4])
			dbName = m[5]
		}
	}

	filename := fmt.Sprintf("backup_%s_%s.sql", dbName, time.Now().Format("20060102_150405"))
	tmpFile := filepath.Join(os.TempDir(), filename)

	args := []string{
		fmt.Sprintf("--host=%s", dbHost),
		fmt.Sprintf("--port=%d", dbPort),
		fmt.Sprintf("--user=%s", dbUser),
		fmt.Sprintf("--password=%s", dbPass),
		"--single-transaction",
		"--routines",
		"--triggers",
		"--default-character-set=utf8mb4",
		dbName,
	}
	cmd := exec.Command("mysqldump", args...)

	output, err := cmd.Output()
	if err != nil {
		h.logAction(c, "error", "database_backup", "备份失败: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "备份失败: " + err.Error()})
		return
	}

	if err := os.WriteFile(tmpFile, output, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "写入备份文件失败"})
		return
	}

	fileSize := len(output)
	h.logAction(c, "info", "database_backup", fmt.Sprintf("数据库备份完成: %s (%.2f MB)", filename, float64(fileSize)/1024/1024))

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.File(tmpFile)

	go func() {
		time.Sleep(5 * time.Second)
		os.Remove(tmpFile)
	}()
}

func (h *AdminHandler) findAdmin(username string) (*adminUserRow, error) {
	var admin adminUserRow
	var changed int
	err := h.db.QueryRow("SELECT id, username, password, password_changed FROM admin_users WHERE username=?", username).
		Scan(&admin.ID, &admin.Username, &admin.Password, &changed)
	admin.PasswordChanged = changed == 1
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (h *AdminHandler) updateSessionPasswordChanged(username string, changed bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, session := range h.sessions {
		if session.Username == username {
			session.PasswordChanged = changed
		}
	}
}

func (h *AdminHandler) loadSystemSettings() (map[string]interface{}, error) {
	settings := make(map[string]interface{})
	var siteTitle, siteDescription, siteKeywords, siteAuthor sql.NullString
	var logoURL, faviconURL, backgroundURL, themeColor sql.NullString
	var customCSS, customJS, footerText sql.NullString
	var allowRegister int
	var maxSessions, sessionTimeout, borderRadius, backgroundBlur int
	var captchaEnabled int
	var captchaID, captchaKey sql.NullString
	err := h.db.QueryRow(`
		SELECT site_title, site_description, site_keywords, site_author,
		       logo_url, favicon_url, background_url, theme_color,
		       custom_css, custom_js, footer_text,
		       allow_register, max_sessions, session_timeout, border_radius,
		       background_blur, captcha_enabled, captcha_id, captcha_key
		FROM system_settings WHERE id=1`,
	).Scan(&siteTitle, &siteDescription, &siteKeywords, &siteAuthor, &logoURL, &faviconURL, &backgroundURL, &themeColor, &customCSS, &customJS, &footerText, &allowRegister, &maxSessions, &sessionTimeout, &borderRadius, &backgroundBlur, &captchaEnabled, &captchaID, &captchaKey)
	if err != nil {
		return nil, err
	}

	settings["site_title"] = nullString(siteTitle, "芙芙云 SSH Terminal")
	settings["site_description"] = nullString(siteDescription, "")
	settings["site_keywords"] = nullString(siteKeywords, "")
	settings["site_author"] = nullString(siteAuthor, "")
	settings["logo_url"] = nullString(logoURL, "https://logo.fufuidc.com/logo3.png")
	settings["favicon_url"] = nullString(faviconURL, "/favicon.svg")
	settings["background_url"] = nullString(backgroundURL, "https://www.loliapi.com/acg/")
	settings["theme_color"] = nullString(themeColor, "#3B82F6")
	settings["custom_css"] = nullString(customCSS, "")
	settings["custom_js"] = nullString(customJS, "")
	settings["footer_text"] = nullString(footerText, "")
	settings["allow_register"] = allowRegister == 1
	settings["max_sessions"] = maxSessions
	settings["session_timeout"] = sessionTimeout
	settings["border_radius"] = borderRadius
	settings["background_blur"] = backgroundBlur
	settings["captcha_enabled"] = captchaEnabled == 1
	settings["captcha_id"] = nullString(captchaID, "")
	settings["captcha_key"] = nullString(captchaKey, "")
	return settings, nil
}

func (h *AdminHandler) writeLog(username, level, action, message, ip string) {
	_, _ = h.db.Exec(
		"INSERT INTO admin_logs (admin_username, level, action, message, ip) VALUES (?, ?, ?, ?, ?)",
		username, level, action, message, ip,
	)
}

func (h *AdminHandler) logAction(c *gin.Context, level, action, message string) {
	h.writeLog(currentAdmin(c), level, action, message, c.ClientIP())
}

func (h *AdminHandler) countTable(table string) int64 {
	if !safeIdentPattern.MatchString(table) {
		return 0
	}
	var count int64
	_ = h.db.QueryRow("SELECT COUNT(*) FROM " + quoteIdent(table)).Scan(&count)
	return count
}

func (h *AdminHandler) countAppTables() int64 {
	var count int64
	_ = h.db.QueryRow("SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE()").Scan(&count)
	return count
}

func (h *AdminHandler) tableExists(table string) bool {
	if !safeIdentPattern.MatchString(table) {
		return false
	}
	var count int
	_ = h.db.QueryRow(
		"SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?",
		table,
	).Scan(&count)
	return count > 0
}

func (h *AdminHandler) tableColumns(table string) ([]gin.H, string, error) {
	rows, err := h.db.Query(`
		SELECT COLUMN_NAME, DATA_TYPE, COLUMN_KEY
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`, table)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	columns := make([]gin.H, 0)
	primaryKey := ""
	for rows.Next() {
		var name, dataType, key string
		if rows.Scan(&name, &dataType, &key) == nil {
			if key == "PRI" && primaryKey == "" {
				primaryKey = name
			}
			columns = append(columns, gin.H{"name": name, "type": dataType, "primary": key == "PRI"})
		}
	}
	return columns, primaryKey, nil
}

func (h *AdminHandler) isPrimaryColumn(table, column string) bool {
	if !safeIdentPattern.MatchString(table) || !safeIdentPattern.MatchString(column) {
		return false
	}
	var count int
	_ = h.db.QueryRow(`
		SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ? AND COLUMN_KEY = 'PRI'`,
		table, column,
	).Scan(&count)
	return count > 0
}

func currentAdmin(c *gin.Context) string {
	if username, ok := c.Get("admin_username"); ok {
		if s, ok := username.(string); ok {
			return s
		}
	}
	return ""
}

func pageParams(c *gin.Context, defaultPageSize, maxPageSize int) (int, int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(defaultPageSize)))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize, (page - 1) * pageSize
}

func formatNullTime(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format("2006-01-02 15:04:05")
}

func nullString(s sql.NullString, fallback string) string {
	if s.Valid {
		return s.String
	}
	return fallback
}

func quoteIdent(ident string) string {
	return "`" + strings.ReplaceAll(ident, "`", "``") + "`"
}

func readonlySQL(raw string) (string, bool) {
	query := strings.TrimSpace(raw)
	query = strings.TrimRight(query, " ;\n\r\t")
	if query == "" || strings.Contains(query, ";") {
		return "", false
	}
	lower := strings.ToLower(query)
	allowed := strings.HasPrefix(lower, "select ") ||
		strings.HasPrefix(lower, "show ") ||
		strings.HasPrefix(lower, "describe ") ||
		strings.HasPrefix(lower, "desc ") ||
		strings.HasPrefix(lower, "explain ")
	if !allowed {
		return "", false
	}
	if strings.HasPrefix(lower, "select ") && !regexp.MustCompile(`(?i)\blimit\b`).MatchString(query) {
		query += " LIMIT 200"
	}
	return query, true
}

func rowsToMaps(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		row := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			row[col] = normalizeDBValue(values[i])
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func normalizeDBValue(value interface{}) interface{} {
	switch v := value.(type) {
	case nil:
		return nil
	case []byte:
		if json.Valid(v) {
			var decoded interface{}
			if json.Unmarshal(v, &decoded) == nil {
				return decoded
			}
		}
		return string(v)
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	default:
		return v
	}
}
