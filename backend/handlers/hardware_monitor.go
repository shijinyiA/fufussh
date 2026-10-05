package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"web-ssh/store"
)

type HardwareMonitorHandler struct {
	store store.Store
}

func NewHardwareMonitorHandler(s store.Store) *HardwareMonitorHandler {
	return &HardwareMonitorHandler{store: s}
}

func (h *HardwareMonitorHandler) GetRemoteStats(c *gin.Context) {
	serverID := c.Query("server_id")
	token := c.Query("token")

	var host, username, password, privateKey string
	var port int

	if serverID != "" {
		server, ok := h.store.GetServerByID(serverID)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "服务器不存在"})
			return
		}
		var err error
		password, err = h.store.DecryptServerPassword(serverID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "解密密码失败"})
			return
		}
		host = server.Host
		port = server.Port
		username = server.Username
		privateKey = server.PrivateKey
	} else if token != "" {
		session := GetDirectSession(token)
		if session == nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "会话不存在或已过期"})
			return
		}
		host = session.Host
		username = session.Username
		password = session.Password
		privateKey = session.PrivateKey
		if p, err := strconv.Atoi(session.Port); err == nil {
			port = p
		} else {
			port = 22
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "缺少服务器ID或令牌"})
		return
	}

	client, err := createSSHClient(host, port, username, password, privateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": fmt.Sprintf("SSH连接失败: %s", err.Error())})
		return
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建会话失败"})
		return
	}
	defer session.Close()

	cmd := "echo '===CPU==='; grep 'cpu ' /proc/stat | head -1; sleep 0.1; grep 'cpu ' /proc/stat | head -1; echo '===MEM==='; grep -E '^MemTotal|^MemAvailable' /proc/meminfo; echo '===DISK==='; df -h / | tail -1; echo '===UPTIME==='; cat /proc/uptime"

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		log.Printf("Hardware monitor command failed: %v, stderr: %s", err, stderr.String())
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "执行命令失败"})
		return
	}

	output := stdout.String()
	stats := parseRemoteStats(output)
	stats["timestamp"] = time.Now().UnixMilli()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

func parseRemoteStats(output string) map[string]interface{} {
	stats := map[string]interface{}{
		"cpu_usage":    0,
		"memory_usage": 0,
		"memory_used":  "0 MB",
		"memory_total": "0 MB",
		"disk_usage":   0,
		"disk_used":    "0 GB",
		"disk_total":   "0 GB",
		"uptime":       "-",
	}

	output = strings.ReplaceAll(output, "\r\n", "\n")
	output = strings.ReplaceAll(output, "\r", "\n")

	sections := strings.Split(output, "===CPU===")
	if len(sections) < 2 {
		return stats
	}
	afterCPU := sections[1]

	cpuSections := strings.SplitN(afterCPU, "===MEM===", 2)
	cpuBlock := ""
	memBlock := ""
	diskBlock := ""
	uptimeBlock := ""

	if len(cpuSections) >= 2 {
		cpuBlock = cpuSections[0]
		memAndRest := cpuSections[1]

		memSections := strings.SplitN(memAndRest, "===DISK===", 2)
		if len(memSections) >= 2 {
			memBlock = memSections[0]
			diskAndRest := memSections[1]

			diskSections := strings.SplitN(diskAndRest, "===UPTIME===", 2)
			if len(diskSections) >= 2 {
				diskBlock = diskSections[0]
				uptimeBlock = diskSections[1]
			} else {
				diskBlock = diskAndRest
			}
		} else {
			memBlock = memAndRest
		}
	} else {
		cpuBlock = afterCPU
	}

	cpuLines := strings.Split(strings.TrimSpace(cpuBlock), "\n")
	var cpuLinesFiltered []string
	for _, line := range cpuLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "cpu ") {
			cpuLinesFiltered = append(cpuLinesFiltered, line)
		}
	}

	if len(cpuLinesFiltered) >= 2 {
		stat1 := parseCPUStatLine(cpuLinesFiltered[0])
		stat2 := parseCPUStatLine(cpuLinesFiltered[1])
		if stat1 != nil && stat2 != nil {
			totalDiff := stat2.Total - stat1.Total
			idleDiff := stat2.Idle - stat1.Idle
			if totalDiff > 0 {
				usage := (totalDiff - idleDiff) / totalDiff * 100
				if usage < 0 {
					usage = 0
				}
				if usage > 100 {
					usage = 100
				}
				stats["cpu_usage"] = int(usage + 0.5)
			}
		}
	}

	memLines := strings.Split(strings.TrimSpace(memBlock), "\n")
	var memTotalKB uint64
	var memAvailKB uint64
	for _, line := range memLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "MemTotal:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "MemTotal:"))
			val = strings.TrimSuffix(val, " kB")
			if v, err := strconv.ParseUint(val, 10, 64); err == nil {
				memTotalKB = v
			}
		}
		if strings.HasPrefix(line, "MemAvailable:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "MemAvailable:"))
			val = strings.TrimSuffix(val, " kB")
			if v, err := strconv.ParseUint(val, 10, 64); err == nil {
				memAvailKB = v
			}
		}
	}

	if memTotalKB > 0 {
		usedKB := memTotalKB - memAvailKB
		usagePercent := int(float64(usedKB) / float64(memTotalKB) * 100 + 0.5)
		if usagePercent < 0 {
			usagePercent = 0
		}
		if usagePercent > 100 {
			usagePercent = 100
		}
		stats["memory_usage"] = usagePercent
		stats["memory_used"] = formatKB(usedKB)
		stats["memory_total"] = formatKB(memTotalKB)
	}

	diskLines := strings.Split(strings.TrimSpace(diskBlock), "\n")
	for _, line := range diskLines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Filesystem") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 6 {
			totalStr := parts[1]
			usedStr := parts[2]
			pctStr := parts[4]
			pctStr = strings.TrimSuffix(pctStr, "%")

			if pct, err := strconv.Atoi(pctStr); err == nil {
				stats["disk_usage"] = pct
				stats["disk_used"] = usedStr
				stats["disk_total"] = totalStr
			}
		}
		break
	}

	uptimeStr := strings.TrimSpace(uptimeBlock)
	uptimeStr = strings.Split(uptimeStr, "\n")[0]
	uptimeStr = strings.TrimSpace(uptimeStr)

	if parts := strings.Fields(uptimeStr); len(parts) >= 1 {
		if secondsFloat, err := strconv.ParseFloat(parts[0], 64); err == nil {
			seconds := int(secondsFloat)
			days := seconds / 86400
			hours := (seconds % 86400) / 3600
			minutes := (seconds % 3600) / 60

			result := ""
			if days > 0 {
				result += fmt.Sprintf("%d天", days)
			}
			if hours > 0 || days > 0 {
				result += fmt.Sprintf("%d时", hours)
			}
			result += fmt.Sprintf("%d分", minutes)
			stats["uptime"] = result
		}
	}

	return stats
}

type cpuStat struct {
	Idle  float64
	Total float64
}

func parseCPUStatLine(line string) *cpuStat {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return nil
	}

	user, _ := strconv.ParseFloat(fields[1], 64)
	nice, _ := strconv.ParseFloat(fields[2], 64)
	system, _ := strconv.ParseFloat(fields[3], 64)
	idle, _ := strconv.ParseFloat(fields[4], 64)

	var iowait float64
	if len(fields) > 5 {
		iowait, _ = strconv.ParseFloat(fields[5], 64)
	}

	total := user + nice + system + idle + iowait

	return &cpuStat{Idle: idle + iowait, Total: total}
}

func formatKB(kb uint64) string {
	const (
		MB = 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case kb >= TB:
		return fmt.Sprintf("%.2f TB", float64(kb)/float64(TB))
	case kb >= GB:
		return fmt.Sprintf("%.2f GB", float64(kb)/float64(GB))
	case kb >= MB:
		return fmt.Sprintf("%.1f MB", float64(kb)/float64(MB))
	default:
		return fmt.Sprintf("%d KB", kb)
	}
}
