package handlers

import (
	"fmt"
	"net/http"
	"time"

	"web-ssh/models"
	"web-ssh/store"
	"web-ssh/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	store    store.Store
	sessions map[string]*models.User
}

func NewAuthHandler(s store.Store) *AuthHandler {
	return &AuthHandler{
		store:    s,
		sessions: make(map[string]*models.User),
	}
}

type LoginRequest struct {
	UserID   string         `json:"user_id" binding:"required"`
	Password string         `json:"password" binding:"required"`
	Captcha  *CaptchaResult `json:"captcha,omitempty"`
}

type RegisterRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	UserID   string `json:"user_id,omitempty"`
	Token    string `json:"token,omitempty"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{Success: false, Message: "请输入用户ID和密码"})
		return
	}

	if len(req.UserID) < 2 || len(req.UserID) > 32 {
		c.JSON(http.StatusBadRequest, LoginResponse{Success: false, Message: "用户ID长度需在2-32个字符之间"})
		return
	}

	if len(req.Password) < 4 || len(req.Password) > 64 {
		c.JSON(http.StatusBadRequest, LoginResponse{Success: false, Message: "密码长度需在4-64个字符之间"})
		return
	}

	if h.store.UserExists(req.UserID) {
		c.JSON(http.StatusConflict, LoginResponse{Success: false, Message: "该用户ID已被注册"})
		return
	}

	if err := h.store.CreateUser(req.UserID, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, LoginResponse{Success: false, Message: "注册失败，请稍后重试"})
		return
	}

	token := utils.GenerateSessionToken()
	h.sessions[token] = &models.User{ID: req.UserID}
	utils.SetSessionCookie(c.Writer, token)

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "web_ssh_user",
		Value:    req.UserID,
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})

	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Message: "注册成功",
		UserID:  req.UserID,
		Token:   token,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{Success: false, Message: "请输入用户ID和密码"})
		return
	}

	if err := h.verifyCaptcha(c, req.Captcha); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{Success: false, Message: err.Error()})
		return
	}

	if !h.store.ValidateUser(req.UserID, req.Password) {
		c.JSON(http.StatusUnauthorized, LoginResponse{Success: false, Message: "用户ID或密码错误"})
		return
	}

	token := utils.GenerateSessionToken()
	user, _ := h.store.GetUser(req.UserID)
	h.sessions[token] = user

	utils.SetSessionCookie(c.Writer, token)

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "web_ssh_user",
		Value:    req.UserID,
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})

	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Message: "登录成功",
		UserID:  req.UserID,
		Token:   token,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token := utils.GetCookieValue(c.Request, "web_ssh_session")
	if token != "" {
		delete(h.sessions, token)
	}

	utils.ClearSessionCookie(c.Writer)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "web_ssh_user",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已退出登录"})
}

func (h *AuthHandler) CheckAuth(c *gin.Context) {
	token := utils.GetCookieValue(c.Request, "web_ssh_session")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未登录"})
		return
	}

	user, ok := h.sessions[token]
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "会话已过期"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"user_id": user.ID,
	})
}

func (h *AuthHandler) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := utils.GetCookieValue(c.Request, "web_ssh_session")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未登录，请先登录"})
			c.Abort()
			return
		}

		user, ok := h.sessions[token]
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "会话已过期，请重新登录"})
			c.Abort()
			return
		}

		utils.SetSessionCookie(c.Writer, token)
		c.Set("user", user)
		c.Set("token", token)
		c.Next()
	}
}

func getUserID(c *gin.Context) string {
	if u, exists := c.Get("user"); exists {
		if user, ok := u.(*models.User); ok {
			return user.ID
		}
	}
	return ""
}

func (h *AuthHandler) verifyCaptcha(c *gin.Context, captcha *CaptchaResult) error {
	settings, err := h.store.GetSystemSettings()
	if err != nil {
		return nil
	}
	enabled := false
	if v, ok := settings["captcha_enabled"]; ok {
		switch val := v.(type) {
		case bool:
			enabled = val
		case int:
			enabled = val == 1
		case float64:
			enabled = val == 1
		}
	}
	if !enabled {
		return nil
	}
	captchaID, _ := settings["captcha_id"].(string)
	captchaKey, _ := settings["captcha_key"].(string)
	if captchaID == "" || captchaKey == "" {
		return nil
	}

	if captcha == nil {
		return fmt.Errorf("请完成验证码验证")
	}

	if captcha.LotNumber == "" || captcha.CaptchaOutput == "" || captcha.PassToken == "" || captcha.GenTime == "" {
		return fmt.Errorf("验证码参数不完整")
	}

	ok, err := captchaVerify(captchaID, captchaKey, captcha.LotNumber, captcha.CaptchaOutput, captcha.PassToken, captcha.GenTime)
	if err != nil {
		return fmt.Errorf("验证码验证失败: %v", err)
	}
	if !ok {
		return fmt.Errorf("验证码验证未通过")
	}
	return nil
}

func (h *AuthHandler) CleanupSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
	}
}
