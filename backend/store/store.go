package store

import "web-ssh/models"

type Store interface {
	GetUser(userID string) (*models.User, bool)
	CreateUser(id, password string) error
	UpdateUserProfile(userID, username, avatar string) error
	ChangeUserPassword(userID, oldPassword, newPassword string) error
	ValidateUser(userID, password string) bool
	UserExists(userID string) bool

	GetServers(userID string) []*models.Server
	GetServer(id, userID string) (*models.Server, bool)
	GetServerByID(id string) (*models.Server, bool)
	CreateServer(server *models.Server) error
	UpdateServer(server *models.Server) error
	DeleteServer(id, userID string) error
	DecryptServerPassword(serverID string) (string, error)
	GetUserGroups(userID string) map[string]int

	GetTerminalSettings(userID, serverID string) (*models.TerminalSettings, error)
	SaveTerminalSettings(settings *models.TerminalSettings) error

	GetSystemSettings() (map[string]interface{}, error)
	GetQuickCommands() ([]map[string]string, error)

	CreateShareToken(serverID, userID string) (*models.ShareToken, error)
	CreateShareTokenWithCredentials(serverID, userID, password, privateKey string) (*models.ShareToken, error)
	GetShareToken(token string) (*models.ShareToken, error)
	UseShareToken(token string) error
}
