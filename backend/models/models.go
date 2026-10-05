package models

import "time"

type Server struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	PrivateKey  string    `json:"private_key,omitempty"`
	Group       string    `json:"group,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Password string `json:"-"`
}

type ShareToken struct {
	ID         string    `json:"id"`
	ServerID   string    `json:"server_id"`
	UserID     string    `json:"user_id"`
	Token      string    `json:"token"`
	Password   string    `json:"-"`
	PrivateKey string    `json:"-"`
	Used       bool      `json:"used"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type Session struct {
	ID          string    `json:"id"`
	ServerID    string    `json:"server_id"`
	UserID      string    `json:"user_id"`
	Connected   bool      `json:"connected"`
	ConnectedAt time.Time `json:"connected_at"`
}

type FileInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	Mode      string    `json:"mode"`
	ModTime   time.Time `json:"mod_time"`
	IsDir     bool      `json:"is_dir"`
	IsSymlink bool      `json:"is_symlink"`
}

type SFTPRequest struct {
	Action  string   `json:"action"`
	Path    string   `json:"path"`
	Dst     string   `json:"dst,omitempty"`
	Name    string   `json:"name,omitempty"`
	Content string   `json:"content,omitempty"`
	Paths   []string `json:"paths,omitempty"`
}

type SFTPResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Files   []FileInfo `json:"files,omitempty"`
	Data    string     `json:"data,omitempty"`
}

type WSMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

type TerminalSettings struct {
	ID                string `json:"id"`
	UserID            string `json:"user_id"`
	ServerID          string `json:"server_id"`
	FontSize          int    `json:"font_size"`
	FontFamily        string `json:"font_family"`
	FontWeight        string `json:"font_weight"`
	LineHeight        int    `json:"line_height"`
	CursorStyle       string `json:"cursor_style"`
	CursorBlink       bool   `json:"cursor_blink"`
	CursorColor       string `json:"cursor_color"`
	Scrollback        int    `json:"scrollback"`
	Background        string `json:"background"`
	Foreground        string `json:"foreground"`
	ThemeBlack        string `json:"theme_black"`
	ThemeRed          string `json:"theme_red"`
	ThemeGreen        string `json:"theme_green"`
	ThemeYellow       string `json:"theme_yellow"`
	ThemeBlue         string `json:"theme_blue"`
	ThemeMagenta      string `json:"theme_magenta"`
	ThemeCyan         string `json:"theme_cyan"`
	ThemeWhite        string `json:"theme_white"`
	ThemeBrightBlack  string `json:"theme_bright_black"`
	ThemeBrightRed    string `json:"theme_bright_red"`
	ThemeBrightGreen  string `json:"theme_bright_green"`
	ThemeBrightYellow string `json:"theme_bright_yellow"`
	ThemeBrightBlue   string `json:"theme_bright_blue"`
	ThemeBrightMagenta string `json:"theme_bright_magenta"`
	ThemeBrightCyan   string `json:"theme_bright_cyan"`
	ThemeBrightWhite  string `json:"theme_bright_white"`
	SelectionBg       string `json:"selection_bg"`
	LetterSpacing     int    `json:"letter_spacing"`
	BellStyle         string `json:"bell_style"`
	AllowTransparency bool   `json:"allow_transparency"`
	BgImageUrl        string `json:"bg_image_url"`
	BgOpacity         int    `json:"bg_opacity"`
	BgBlur            bool   `json:"bg_blur"`
}
