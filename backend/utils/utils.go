package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
)

func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GenerateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func FormatFileSize(size int64) string {
	if size < 0 {
		return "-"
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	floatSize := float64(size)
	unitIndex := 0

	for floatSize >= 1024 && unitIndex < len(units)-1 {
		floatSize /= 1024
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%d %s", size, units[unitIndex])
	}
	return fmt.Sprintf("%.1f %s", floatSize, units[unitIndex])
}

func GetCookieValue(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "web_ssh_session",
		Value:    token,
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "web_ssh_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func ParseFileMode(mode string) string {
	if len(mode) < 9 {
		return mode
	}

	var result strings.Builder
	characters := []struct {
		read  byte
		write byte
		exec  byte
	}{
		{mode[0], mode[1], mode[2]},
		{mode[3], mode[4], mode[5]},
		{mode[6], mode[7], mode[8]},
	}

	if characters[0].read == 'r' {
		result.WriteString("r")
	} else {
		result.WriteString("-")
	}
	if characters[0].write == 'w' {
		result.WriteString("w")
	} else {
		result.WriteString("-")
	}
	if characters[0].exec == 'x' {
		result.WriteString("x")
	} else {
		result.WriteString("-")
	}

	if characters[1].read == 'r' {
		result.WriteString("r")
	} else {
		result.WriteString("-")
	}
	if characters[1].write == 'w' {
		result.WriteString("w")
	} else {
		result.WriteString("-")
	}
	if characters[1].exec == 'x' {
		result.WriteString("x")
	} else {
		result.WriteString("-")
	}

	if characters[2].read == 'r' {
		result.WriteString("r")
	} else {
		result.WriteString("-")
	}
	if characters[2].write == 'w' {
		result.WriteString("w")
	} else {
		result.WriteString("-")
	}
	if characters[2].exec == 'x' {
		result.WriteString("x")
	} else {
		result.WriteString("-")
	}

	return result.String()
}

func RandomColor() string {
	colors := []string{
		"#3B82F6", "#EF4444", "#10B981", "#F59E0B",
		"#8B5CF6", "#EC4899", "#06B6D4", "#F97316",
		"#6366F1", "#14B8A6", "#E11D48", "#84CC16",
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(colors))))
	return colors[n.Int64()]
}

func FormatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
