package handlers

import (
	"net/http"

	"web-ssh/store"

	"github.com/gin-gonic/gin"
)

type SystemSettingsHandler struct {
	store store.Store
}

func NewSystemSettingsHandler(s store.Store) *SystemSettingsHandler {
	return &SystemSettingsHandler{store: s}
}

func (h *SystemSettingsHandler) GetPublicSettings(c *gin.Context) {
	settings, err := h.store.GetSystemSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取设置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"settings": settings,
	})
}

func (h *SystemSettingsHandler) GetPublicCommands(c *gin.Context) {
	commands, err := h.store.GetQuickCommands()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "获取快捷命令失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"commands": commands,
	})
}
