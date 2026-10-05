package store

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"web-ssh/models"

	"golang.org/x/crypto/bcrypt"
)

type MySQLStore struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewMySQLStore(dsn string) (*MySQLStore, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &MySQLStore{db: db}

	if err := store.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return store, nil
}

func (s *MySQLStore) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(64) PRIMARY KEY,
			username VARCHAR(128),
			avatar VARCHAR(512),
			password VARCHAR(255) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS servers (
			id VARCHAR(64) PRIMARY KEY,
			user_id VARCHAR(64) NOT NULL,
			name VARCHAR(255) NOT NULL,
			host VARCHAR(255) NOT NULL,
			port INT NOT NULL DEFAULT 22,
			username VARCHAR(128) NOT NULL,
			password TEXT,
			private_key TEXT,
			group_name VARCHAR(128),
			tags JSON,
			description TEXT,
			color VARCHAR(32),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS share_tokens (
			id VARCHAR(64) PRIMARY KEY,
			server_id VARCHAR(64) NOT NULL,
			user_id VARCHAR(64) NOT NULL,
			token VARCHAR(128) NOT NULL UNIQUE,
			password TEXT,
			private_key TEXT,
			used TINYINT(1) DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			INDEX idx_token (token),
			INDEX idx_server (server_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS terminal_settings (
			id VARCHAR(64) PRIMARY KEY,
			user_id VARCHAR(64) NOT NULL,
			server_id VARCHAR(64) NOT NULL,
			font_size INT NOT NULL DEFAULT 14,
			font_family VARCHAR(256) NOT NULL DEFAULT '"Cascadia Code", "Fira Code", "JetBrains Mono", Menlo, Monaco, monospace',
			font_weight VARCHAR(32) NOT NULL DEFAULT 'normal',
			line_height INT NOT NULL DEFAULT 1,
			cursor_style VARCHAR(16) NOT NULL DEFAULT 'bar',
			cursor_blink TINYINT(1) NOT NULL DEFAULT 1,
			cursor_color VARCHAR(32) NOT NULL DEFAULT '#00d4ff',
			scrollback INT NOT NULL DEFAULT 10000,
			background VARCHAR(32) NOT NULL DEFAULT '#0f0f1a',
			foreground VARCHAR(32) NOT NULL DEFAULT '#e0e0e0',
			theme_black VARCHAR(32) NOT NULL DEFAULT '#0f0f1a',
			theme_red VARCHAR(32) NOT NULL DEFAULT '#ff4757',
			theme_green VARCHAR(32) NOT NULL DEFAULT '#2ed573',
			theme_yellow VARCHAR(32) NOT NULL DEFAULT '#ffa502',
			theme_blue VARCHAR(32) NOT NULL DEFAULT '#3B82F6',
			theme_magenta VARCHAR(32) NOT NULL DEFAULT '#a855f7',
			theme_cyan VARCHAR(32) NOT NULL DEFAULT '#00d4ff',
			theme_white VARCHAR(32) NOT NULL DEFAULT '#e0e0e0',
			theme_bright_black VARCHAR(32) NOT NULL DEFAULT '#5a6480',
			theme_bright_red VARCHAR(32) NOT NULL DEFAULT '#ff6b81',
			theme_bright_green VARCHAR(32) NOT NULL DEFAULT '#7bed9f',
			theme_bright_yellow VARCHAR(32) NOT NULL DEFAULT '#ffc048',
			theme_bright_blue VARCHAR(32) NOT NULL DEFAULT '#60a5fa',
			theme_bright_magenta VARCHAR(32) NOT NULL DEFAULT '#c084fc',
			theme_bright_cyan VARCHAR(32) NOT NULL DEFAULT '#22d3ee',
			theme_bright_white VARCHAR(32) NOT NULL DEFAULT '#ffffff',
			selection_bg VARCHAR(64) NOT NULL DEFAULT 'rgba(0,212,255,0.3)',
			letter_spacing INT NOT NULL DEFAULT 0,
			bell_style VARCHAR(16) NOT NULL DEFAULT 'none',
			allow_transparency TINYINT(1) NOT NULL DEFAULT 1,
			bg_image_url VARCHAR(512) NOT NULL DEFAULT '',
			bg_opacity INT NOT NULL DEFAULT 30,
			bg_blur TINYINT(1) NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_user_server (user_id, server_id),
			INDEX idx_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS system_settings (
			id INT PRIMARY KEY DEFAULT 1,
			site_title VARCHAR(255) DEFAULT '芙芙云 SSH Terminal',
			site_description TEXT,
			site_keywords VARCHAR(500) DEFAULT '',
			site_author VARCHAR(100) DEFAULT '',
			logo_url VARCHAR(500) DEFAULT 'https://logo.fufuidc.com/logo3.png',
			favicon_url VARCHAR(500) DEFAULT '/favicon.svg',
			background_url VARCHAR(500) DEFAULT 'https://www.loliapi.com/acg/',
			theme_color VARCHAR(20) DEFAULT '#3B82F6',
			custom_css TEXT,
			custom_js TEXT,
			footer_text VARCHAR(500) DEFAULT ''
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`INSERT IGNORE INTO system_settings (id) VALUES (1)`,
		`CREATE TABLE IF NOT EXISTS quick_commands (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			command TEXT NOT NULL,
			sort_order INT NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS admin_users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(64) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			password_changed TINYINT(1) NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS admin_logs (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			admin_username VARCHAR(64),
			level VARCHAR(16) NOT NULL DEFAULT 'info',
			action VARCHAR(64) NOT NULL DEFAULT '',
			message TEXT NOT NULL,
			ip VARCHAR(64) DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_created_at (created_at),
			INDEX idx_action (action)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`INSERT INTO quick_commands (name, command, sort_order)
		 SELECT '查看内存', 'free -h', 10 FROM DUAL
		 WHERE NOT EXISTS (SELECT 1 FROM quick_commands)`,
		`INSERT INTO quick_commands (name, command, sort_order)
		 SELECT '磁盘使用', 'df -hT', 20 FROM DUAL
		 WHERE NOT EXISTS (SELECT 1 FROM quick_commands WHERE name = '磁盘使用')`,
		`INSERT INTO quick_commands (name, command, sort_order)
		 SELECT '系统信息', 'uname -a && cat /etc/os-release', 30 FROM DUAL
		 WHERE NOT EXISTS (SELECT 1 FROM quick_commands WHERE name = '系统信息')`,
		`INSERT INTO quick_commands (name, command, sort_order)
		 SELECT '监听端口', 'ss -tlnp', 40 FROM DUAL
		 WHERE NOT EXISTS (SELECT 1 FROM quick_commands WHERE name = '监听端口')`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	alterQueries := []string{
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(128)`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar VARCHAR(512)`,
		`ALTER TABLE terminal_settings ADD COLUMN IF NOT EXISTS bg_image_url VARCHAR(512) NOT NULL DEFAULT ''`,
		`ALTER TABLE terminal_settings ADD COLUMN IF NOT EXISTS bg_opacity INT NOT NULL DEFAULT 30`,
		`ALTER TABLE terminal_settings ADD COLUMN IF NOT EXISTS bg_blur TINYINT(1) NOT NULL DEFAULT 0`,
	}
	for _, q := range alterQueries {
		s.db.Exec(q)
	}

	s.ensureColumn("share_tokens", "password", "TEXT")
	s.ensureColumn("share_tokens", "private_key", "TEXT")
	s.ensureColumn("system_settings", "site_description", "TEXT")
	s.ensureColumn("system_settings", "site_keywords", "VARCHAR(500) DEFAULT ''")
	s.ensureColumn("system_settings", "site_author", "VARCHAR(100) DEFAULT ''")
	s.ensureColumn("system_settings", "logo_url", "VARCHAR(500) DEFAULT 'https://logo.fufuidc.com/logo3.png'")
	s.ensureColumn("system_settings", "favicon_url", "VARCHAR(500) DEFAULT '/favicon.svg'")
	s.ensureColumn("system_settings", "background_url", "VARCHAR(500) DEFAULT 'https://www.loliapi.com/acg/'")
	s.ensureColumn("system_settings", "theme_color", "VARCHAR(20) DEFAULT '#3B82F6'")
	s.ensureColumn("system_settings", "custom_css", "TEXT")
	s.ensureColumn("system_settings", "custom_js", "TEXT")
	s.ensureColumn("system_settings", "footer_text", "TEXT")
	s.ensureColumn("system_settings", "allow_register", "TINYINT(1) DEFAULT 1")
	s.ensureColumn("system_settings", "max_sessions", "INT DEFAULT 10")
	s.ensureColumn("system_settings", "session_timeout", "INT DEFAULT 30")
	s.ensureColumn("system_settings", "border_radius", "INT DEFAULT 8")
	s.ensureColumn("system_settings", "captcha_enabled", "TINYINT(1) DEFAULT 0")
	s.ensureColumn("system_settings", "captcha_id", "VARCHAR(100) DEFAULT ''")
	s.ensureColumn("system_settings", "captcha_key", "VARCHAR(100) DEFAULT ''")
	s.ensureColumn("system_settings", "background_blur", "INT DEFAULT 0")

	return nil
}

func (s *MySQLStore) ensureColumn(table, column, definition string) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
		table, column,
	).Scan(&count)
	if err == nil && count == 0 {
		sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)
		if _, err := s.db.Exec(sql); err != nil {
			fmt.Printf("Failed to add column %s.%s: %v\n", table, column, err)
		} else {
			fmt.Printf("Added column %s.%s\n", table, column)
		}
	}
}

func (s *MySQLStore) Close() error {
	return s.db.Close()
}

func (s *MySQLStore) DB() *sql.DB {
	return s.db
}

func (s *MySQLStore) GetUser(userID string) (*models.User, bool) {
	var user models.User
	err := s.db.QueryRow("SELECT id, COALESCE(username, ''), COALESCE(avatar, ''), password FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Username, &user.Avatar, &user.Password)
	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		return nil, false
	}
	return &user, true
}

func (s *MySQLStore) CreateUser(id, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	_, err = s.db.Exec(
		"INSERT INTO users (id, password) VALUES (?, ?)",
		id, string(hashedPassword),
	)
	return err
}

func (s *MySQLStore) ValidateUser(userID, password string) bool {
	var hashedPassword string
	err := s.db.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&hashedPassword)
	if err != nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *MySQLStore) UserExists(userID string) bool {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", userID).Scan(&count)
	return count > 0
}

func (s *MySQLStore) GetServers(userID string) []*models.Server {
	rows, err := s.db.Query(`
		SELECT id, name, host, port, username, password, private_key, group_name, tags, description, color, created_at, updated_at
		FROM servers WHERE user_id = ? ORDER BY name`, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var servers []*models.Server
	for rows.Next() {
		server := parseServerRow(rows)
		if server != nil {
			servers = append(servers, server)
		}
	}
	return servers
}

func (s *MySQLStore) GetServer(id, userID string) (*models.Server, bool) {
	row := s.db.QueryRow(`
		SELECT id, name, host, port, username, password, private_key, group_name, tags, description, color, created_at, updated_at
		FROM servers WHERE id = ? AND user_id = ?`, id, userID)

	server := scanServer(row)
	if server == nil {
		return nil, false
	}
	return server, true
}

func (s *MySQLStore) GetServerByID(id string) (*models.Server, bool) {
	row := s.db.QueryRow(`
		SELECT id, name, host, port, username, password, private_key, group_name, tags, description, color, created_at, updated_at
		FROM servers WHERE id = ?`, id)

	server := scanServer(row)
	if server == nil {
		return nil, false
	}
	return server, true
}

func (s *MySQLStore) CreateServer(server *models.Server) error {
	tagsJSON := "[]"
	if len(server.Tags) > 0 {
		tagsJSON = strings.Join(server.Tags, ",")
	}

	_, err := s.db.Exec(`
		INSERT INTO servers (id, user_id, name, host, port, username, password, private_key, group_name, tags, description, color)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		server.ID, server.UserID, server.Name, server.Host, server.Port, server.Username,
		server.Password, server.PrivateKey, server.Group, tagsJSON, server.Description, server.Color,
	)
	return err
}

func (s *MySQLStore) UpdateServer(server *models.Server) error {
	tagsJSON := "[]"
	if len(server.Tags) > 0 {
		tagsJSON = strings.Join(server.Tags, ",")
	}

	_, err := s.db.Exec(`
		UPDATE servers SET name=?, host=?, port=?, username=?, password=?, private_key=?,
			group_name=?, tags=?, description=?, color=?
		WHERE id = ? AND user_id = ?`,
		server.Name, server.Host, server.Port, server.Username, server.Password, server.PrivateKey,
		server.Group, tagsJSON, server.Description, server.Color, server.ID, server.UserID,
	)
	return err
}

func (s *MySQLStore) DeleteServer(id, userID string) error {
	_, err := s.db.Exec("DELETE FROM servers WHERE id = ? AND user_id = ?", id, userID)
	return err
}

func (s *MySQLStore) DecryptServerPassword(serverID string) (string, error) {
	var password string
	err := s.db.QueryRow("SELECT password FROM servers WHERE id = ?", serverID).Scan(&password)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("server not found")
	}
	return password, err
}

func (s *MySQLStore) GetUserGroups(userID string) map[string]int {
	rows, err := s.db.Query(`
		SELECT COALESCE(group_name, '默认') AS g, COUNT(*) AS c
		FROM servers WHERE user_id = ? GROUP BY g`, userID)
	if err != nil {
		return map[string]int{}
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var g string
		var c int
		if rows.Scan(&g, &c) == nil {
			result[g] = c
		}
	}
	return result
}

func parseServerRow(rows *sql.Rows) *models.Server {
	var server models.Server
	var tagsStr sql.NullString
	var group sql.NullString

	err := rows.Scan(
		&server.ID, &server.Name, &server.Host, &server.Port, &server.Username,
		&server.Password, &server.PrivateKey, &group, &tagsStr,
		&server.Description, &server.Color, &server.CreatedAt, &server.UpdatedAt,
	)
	if err != nil {
		return nil
	}

	server.UserID = ""
	server.Group = group.String
	if tagsStr.Valid && tagsStr.String != "" && tagsStr.String != "[]" {
		for _, t := range strings.Split(tagsStr.String, ",") {
			server.Tags = append(server.Tags, strings.TrimSpace(t))
		}
	}
	return &server
}

func scanServer(row *sql.Row) *models.Server {
	var server models.Server
	var tagsStr sql.NullString
	var group sql.NullString

	err := row.Scan(
		&server.ID, &server.Name, &server.Host, &server.Port, &server.Username,
		&server.Password, &server.PrivateKey, &group, &tagsStr,
		&server.Description, &server.Color, &server.CreatedAt, &server.UpdatedAt,
	)
	if err != nil {
		return nil
	}

	server.UserID = ""
	server.Group = group.String
	if tagsStr.Valid && tagsStr.String != "" && tagsStr.String != "[]" {
		for _, t := range strings.Split(tagsStr.String, ",") {
			server.Tags = append(server.Tags, strings.TrimSpace(t))
		}
	}
	return &server
}

func (s *MySQLStore) GetTerminalSettings(userID, serverID string) (*models.TerminalSettings, error) {
	var ts models.TerminalSettings
	err := s.db.QueryRow(`
		SELECT id, user_id, server_id, font_size, font_family, font_weight, line_height,
			cursor_style, cursor_blink, cursor_color, scrollback,
			background, foreground,
			theme_black, theme_red, theme_green, theme_yellow, theme_blue, theme_magenta, theme_cyan, theme_white,
			theme_bright_black, theme_bright_red, theme_bright_green, theme_bright_yellow, theme_bright_blue, theme_bright_magenta, theme_bright_cyan, theme_bright_white,
			selection_bg, letter_spacing, bell_style, allow_transparency,
			bg_image_url, bg_opacity, bg_blur
		FROM terminal_settings WHERE user_id = ? AND server_id = ?`, userID, serverID).Scan(
		&ts.ID, &ts.UserID, &ts.ServerID, &ts.FontSize, &ts.FontFamily, &ts.FontWeight, &ts.LineHeight,
		&ts.CursorStyle, &ts.CursorBlink, &ts.CursorColor, &ts.Scrollback,
		&ts.Background, &ts.Foreground,
		&ts.ThemeBlack, &ts.ThemeRed, &ts.ThemeGreen, &ts.ThemeYellow, &ts.ThemeBlue, &ts.ThemeMagenta, &ts.ThemeCyan, &ts.ThemeWhite,
		&ts.ThemeBrightBlack, &ts.ThemeBrightRed, &ts.ThemeBrightGreen, &ts.ThemeBrightYellow, &ts.ThemeBrightBlue, &ts.ThemeBrightMagenta, &ts.ThemeBrightCyan, &ts.ThemeBrightWhite,
		&ts.SelectionBg, &ts.LetterSpacing, &ts.BellStyle, &ts.AllowTransparency,
		&ts.BgImageUrl, &ts.BgOpacity, &ts.BgBlur,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ts, nil
}

func (s *MySQLStore) SaveTerminalSettings(settings *models.TerminalSettings) error {
	existing, err := s.GetTerminalSettings(settings.UserID, settings.ServerID)
	if err != nil {
		return err
	}

	if existing != nil {
		_, err = s.db.Exec(`
			UPDATE terminal_settings SET
				font_size=?, font_family=?, font_weight=?, line_height=?,
				cursor_style=?, cursor_blink=?, cursor_color=?, scrollback=?,
				background=?, foreground=?,
				theme_black=?, theme_red=?, theme_green=?, theme_yellow=?, theme_blue=?, theme_magenta=?, theme_cyan=?, theme_white=?,
				theme_bright_black=?, theme_bright_red=?, theme_bright_green=?, theme_bright_yellow=?, theme_bright_blue=?, theme_bright_magenta=?, theme_bright_cyan=?, theme_bright_white=?,
				selection_bg=?, letter_spacing=?, bell_style=?, allow_transparency=?,
				bg_image_url=?, bg_opacity=?, bg_blur=?
			WHERE user_id = ? AND server_id = ?`,
			settings.FontSize, settings.FontFamily, settings.FontWeight, settings.LineHeight,
			settings.CursorStyle, settings.CursorBlink, settings.CursorColor, settings.Scrollback,
			settings.Background, settings.Foreground,
			settings.ThemeBlack, settings.ThemeRed, settings.ThemeGreen, settings.ThemeYellow, settings.ThemeBlue, settings.ThemeMagenta, settings.ThemeCyan, settings.ThemeWhite,
			settings.ThemeBrightBlack, settings.ThemeBrightRed, settings.ThemeBrightGreen, settings.ThemeBrightYellow, settings.ThemeBrightBlue, settings.ThemeBrightMagenta, settings.ThemeBrightCyan, settings.ThemeBrightWhite,
			settings.SelectionBg, settings.LetterSpacing, settings.BellStyle, settings.AllowTransparency,
			settings.BgImageUrl, settings.BgOpacity, settings.BgBlur,
			settings.UserID, settings.ServerID,
		)
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO terminal_settings (
			id, user_id, server_id, font_size, font_family, font_weight, line_height,
			cursor_style, cursor_blink, cursor_color, scrollback,
			background, foreground,
			theme_black, theme_red, theme_green, theme_yellow, theme_blue, theme_magenta, theme_cyan, theme_white,
			theme_bright_black, theme_bright_red, theme_bright_green, theme_bright_yellow, theme_bright_blue, theme_bright_magenta, theme_bright_cyan, theme_bright_white,
			selection_bg, letter_spacing, bell_style, allow_transparency,
			bg_image_url, bg_opacity, bg_blur
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		settings.ID, settings.UserID, settings.ServerID, settings.FontSize, settings.FontFamily, settings.FontWeight, settings.LineHeight,
		settings.CursorStyle, settings.CursorBlink, settings.CursorColor, settings.Scrollback,
		settings.Background, settings.Foreground,
		settings.ThemeBlack, settings.ThemeRed, settings.ThemeGreen, settings.ThemeYellow, settings.ThemeBlue, settings.ThemeMagenta, settings.ThemeCyan, settings.ThemeWhite,
		settings.ThemeBrightBlack, settings.ThemeBrightRed, settings.ThemeBrightGreen, settings.ThemeBrightYellow, settings.ThemeBrightBlue, settings.ThemeBrightMagenta, settings.ThemeBrightCyan, settings.ThemeBrightWhite,
		settings.SelectionBg, settings.LetterSpacing, settings.BellStyle, settings.AllowTransparency,
		settings.BgImageUrl, settings.BgOpacity, settings.BgBlur,
	)
	return err
}

func (s *MySQLStore) UpdateUserProfile(userID, username, avatar string) error {
	_, err := s.db.Exec("UPDATE users SET username = ?, avatar = ? WHERE id = ?", username, avatar, userID)
	return err
}

func (s *MySQLStore) ChangeUserPassword(userID, oldPassword, newPassword string) error {
	var hashedPassword string
	err := s.db.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&hashedPassword)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(oldPassword))
	if err != nil {
		return fmt.Errorf("incorrect old password")
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password")
	}

	_, err = s.db.Exec("UPDATE users SET password = ? WHERE id = ?", string(newHashedPassword), userID)
	return err
}

func (s *MySQLStore) CreateShareToken(serverID, userID string) (*models.ShareToken, error) {
	token := fmt.Sprintf("share_%s_%d", serverID[:8], time.Now().UnixNano())
	id := fmt.Sprintf("st_%d", time.Now().UnixNano())
	expiresAt := time.Now().Add(24 * time.Hour)

	shareToken := &models.ShareToken{
		ID:        id,
		ServerID:  serverID,
		UserID:    userID,
		Token:     token,
		Used:      false,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	_, err := s.db.Exec(`
		INSERT INTO share_tokens (id, server_id, user_id, token, used, expires_at)
		VALUES (?, ?, ?, ?, 0, ?)`,
		shareToken.ID, shareToken.ServerID, shareToken.UserID, shareToken.Token, shareToken.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	return shareToken, nil
}

func (s *MySQLStore) CreateShareTokenWithCredentials(serverID, userID, password, privateKey string) (*models.ShareToken, error) {
	token := fmt.Sprintf("share_%s_%d", serverID[:8], time.Now().UnixNano())
	id := fmt.Sprintf("st_%d", time.Now().UnixNano())
	expiresAt := time.Now().Add(24 * time.Hour)

	shareToken := &models.ShareToken{
		ID:         id,
		ServerID:   serverID,
		UserID:     userID,
		Token:      token,
		Password:   password,
		PrivateKey: privateKey,
		Used:       false,
		CreatedAt:  time.Now(),
		ExpiresAt:  expiresAt,
	}

	_, err := s.db.Exec(`
		INSERT INTO share_tokens (id, server_id, user_id, token, password, private_key, used, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, 0, ?)`,
		shareToken.ID, shareToken.ServerID, shareToken.UserID, shareToken.Token, shareToken.Password, shareToken.PrivateKey, shareToken.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	return shareToken, nil
}

func (s *MySQLStore) GetShareToken(token string) (*models.ShareToken, error) {
	var st models.ShareToken
	var password, privateKey sql.NullString
	err := s.db.QueryRow(`
		SELECT id, server_id, user_id, token, password, private_key, used, created_at, expires_at
		FROM share_tokens WHERE token = ?`, token,
	).Scan(&st.ID, &st.ServerID, &st.UserID, &st.Token, &password, &privateKey, &st.Used, &st.CreatedAt, &st.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if password.Valid {
		st.Password = password.String
	}
	if privateKey.Valid {
		st.PrivateKey = privateKey.String
	}
	return &st, nil
}

func (s *MySQLStore) UseShareToken(token string) error {
	_, err := s.db.Exec("UPDATE share_tokens SET used = 1 WHERE token = ?", token)
	return err
}

func (s *MySQLStore) GetSystemSettings() (map[string]interface{}, error) {
	settings := make(map[string]interface{})

	var siteTitle, siteDescription, siteKeywords, siteAuthor sql.NullString
	var logoURL, faviconURL, backgroundURL, themeColor sql.NullString
	var customCSS, customJS, footerText sql.NullString
	var allowRegister, maxSessions, sessionTimeout, borderRadius, backgroundBlur int
	var captchaID, captchaKey sql.NullString
	var captchaEnabled int

	err := s.db.QueryRow(`
		SELECT site_title, site_description, site_keywords, site_author,
		       logo_url, favicon_url, background_url, theme_color,
		       custom_css, custom_js, footer_text,
		       allow_register, max_sessions, session_timeout, border_radius,
		       background_blur, captcha_enabled, captcha_id, captcha_key
		FROM system_settings WHERE id = 1
	`).Scan(
		&siteTitle, &siteDescription, &siteKeywords, &siteAuthor,
		&logoURL, &faviconURL, &backgroundURL, &themeColor,
		&customCSS, &customJS, &footerText,
		&allowRegister, &maxSessions, &sessionTimeout, &borderRadius,
		&backgroundBlur, &captchaEnabled, &captchaID, &captchaKey,
	)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if siteTitle.Valid {
		settings["site_title"] = siteTitle.String
	} else {
		settings["site_title"] = "芙芙云 SSH Terminal"
	}
	if siteDescription.Valid {
		settings["site_description"] = siteDescription.String
	}
	if siteKeywords.Valid {
		settings["site_keywords"] = siteKeywords.String
	}
	if siteAuthor.Valid {
		settings["site_author"] = siteAuthor.String
	}
	if logoURL.Valid {
		settings["logo_url"] = logoURL.String
	}
	if faviconURL.Valid {
		settings["favicon_url"] = faviconURL.String
	}
	if backgroundURL.Valid {
		settings["background_url"] = backgroundURL.String
	}
	if themeColor.Valid {
		settings["theme_color"] = themeColor.String
	} else {
		settings["theme_color"] = "#3B82F6"
	}
	if customCSS.Valid {
		settings["custom_css"] = customCSS.String
	}
	if customJS.Valid {
		settings["custom_js"] = customJS.String
	}
	if footerText.Valid {
		settings["footer_text"] = footerText.String
	} else {
		settings["footer_text"] = ""
	}
	settings["allow_register"] = allowRegister == 1
	settings["max_sessions"] = maxSessions
	settings["session_timeout"] = sessionTimeout
	settings["border_radius"] = borderRadius
	settings["background_blur"] = backgroundBlur
	settings["captcha_enabled"] = captchaEnabled == 1
	if captchaID.Valid {
		settings["captcha_id"] = captchaID.String
	} else {
		settings["captcha_id"] = ""
	}
	if captchaKey.Valid {
		settings["captcha_key"] = captchaKey.String
	} else {
		settings["captcha_key"] = ""
	}

	return settings, nil
}

func (s *MySQLStore) GetQuickCommands() ([]map[string]string, error) {
	rows, err := s.db.Query("SELECT name, command FROM quick_commands ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []map[string]string
	for rows.Next() {
		var name, command string
		if err := rows.Scan(&name, &command); err != nil {
			continue
		}
		commands = append(commands, map[string]string{
			"name":    name,
			"command": command,
		})
	}

	return commands, nil
}
