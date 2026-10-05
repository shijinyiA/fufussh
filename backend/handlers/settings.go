package handlers

import (
	"net/http"

	"web-ssh/models"
	"web-ssh/store"
	"web-ssh/utils"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	store store.Store
}

func NewSettingsHandler(s store.Store) *SettingsHandler {
	return &SettingsHandler{store: s}
}

func (h *SettingsHandler) GetSettings(c *gin.Context) {
	userID := getUserID(c)
	serverID := c.Param("id")

	settings, err := h.store.GetTerminalSettings(userID, serverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取设置失败"})
		return
	}

	if settings == nil {
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"settings": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"settings": settings,
	})
}

func (h *SettingsHandler) SaveSettings(c *gin.Context) {
	userID := getUserID(c)
	serverID := c.Param("id")

	var req models.TerminalSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}

	existing, _ := h.store.GetTerminalSettings(userID, serverID)

	settings := &models.TerminalSettings{
		UserID:            userID,
		ServerID:          serverID,
		FontSize:          req.FontSize,
		FontFamily:        req.FontFamily,
		FontWeight:        req.FontWeight,
		LineHeight:        req.LineHeight,
		CursorStyle:       req.CursorStyle,
		CursorBlink:       req.CursorBlink,
		CursorColor:       req.CursorColor,
		Scrollback:        req.Scrollback,
		Background:        req.Background,
		Foreground:        req.Foreground,
		ThemeBlack:        req.ThemeBlack,
		ThemeRed:          req.ThemeRed,
		ThemeGreen:        req.ThemeGreen,
		ThemeYellow:       req.ThemeYellow,
		ThemeBlue:         req.ThemeBlue,
		ThemeMagenta:      req.ThemeMagenta,
		ThemeCyan:         req.ThemeCyan,
		ThemeWhite:        req.ThemeWhite,
		ThemeBrightBlack:  req.ThemeBrightBlack,
		ThemeBrightRed:    req.ThemeBrightRed,
		ThemeBrightGreen:  req.ThemeBrightGreen,
		ThemeBrightYellow: req.ThemeBrightYellow,
		ThemeBrightBlue:   req.ThemeBrightBlue,
		ThemeBrightMagenta: req.ThemeBrightMagenta,
		ThemeBrightCyan:   req.ThemeBrightCyan,
		ThemeBrightWhite:  req.ThemeBrightWhite,
		SelectionBg:       req.SelectionBg,
		LetterSpacing:     req.LetterSpacing,
		BellStyle:         req.BellStyle,
		AllowTransparency: req.AllowTransparency,
		BgImageUrl:        req.BgImageUrl,
		BgOpacity:         req.BgOpacity,
		BgBlur:            req.BgBlur,
	}

	if existing != nil {
		settings.ID = existing.ID
	} else {
		settings.ID = utils.GenerateID()
	}

	if settings.FontSize <= 0 {
		settings.FontSize = 14
	}
	if settings.FontFamily == "" {
		settings.FontFamily = `"Cascadia Code", "Fira Code", "JetBrains Mono", Menlo, Monaco, monospace`
	}
	if settings.CursorStyle == "" {
		settings.CursorStyle = "bar"
	}
	if settings.Scrollback <= 0 {
		settings.Scrollback = 10000
	}
	if settings.Background == "" {
		settings.Background = "#0f0f1a"
	}
	if settings.Foreground == "" {
		settings.Foreground = "#e0e0e0"
	}

	if err := h.store.SaveTerminalSettings(settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "保存设置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "设置已保存"})
}
