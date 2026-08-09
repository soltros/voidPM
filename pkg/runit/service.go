package runit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type ServiceStatus string

const (
	StatusRunning  ServiceStatus = "RUNNING"
	StatusStopped  ServiceStatus = "STOPPED"
	StatusFailed   ServiceStatus = "FAILED"
	StatusDisabled ServiceStatus = "DISABLED"
	StatusUnknown  ServiceStatus = "UNKNOWN"
)

type Service struct {
	Name          string        `json:"name"`
	Enabled       bool          `json:"enabled"`
	Status        ServiceStatus `json:"status"`
	PID           int           `json:"pid,omitempty"`
	UptimeSeconds int64         `json:"uptime_seconds,omitempty"`
	StatusRaw     string        `json:"status_raw"`
	ServiceDir    string        `json:"service_dir"`
	ActiveLink    string        `json:"active_link"`
	HasLog        bool          `json:"has_log"`
	IsUser        bool          `json:"is_user"`
}

// Regex to parse 'sv status' output:
// e.g. "run: sshd: (pid 12345) 678s; run: log: (pid 12346) 678s"
// e.g. "down: sshd: 45s, normally up"
var (
	runPattern  = regexp.MustCompile(`^run:\s+([a-zA-Z0-9_-]+):\s+\(pid\s+(\d+)\)\s+(\d+)s`)
	downPattern = regexp.MustCompile(`^down:\s+([a-zA-Z0-9_-]+):\s+(\d+)s`)
)

func (s *Service) FormattedUptime() string {
	if s.UptimeSeconds <= 0 {
		return "-"
	}
	sec := s.UptimeSeconds
	if sec < 60 {
		return fmt.Sprintf("%ds", sec)
	}
	min := sec / 60
	if min < 60 {
		return fmt.Sprintf("%dm %ds", min, sec%60)
	}
	hr := min / 60
	if hr < 24 {
		return fmt.Sprintf("%dh %dm", hr, min%60)
	}
	days := hr / 24
	return fmt.Sprintf("%dd %dh", days, hr%24)
}

// InspectService checks service status in /etc/sv and /var/service
func InspectService(name, svDir, activeDir string) (*Service, error) {
	sDir := filepath.Join(svDir, name)
	aLink := filepath.Join(activeDir, name)

	if fi, err := os.Stat(sDir); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("service '%s' not found in %s", name, svDir)
	}

	srv := &Service{
		Name:       name,
		ServiceDir: sDir,
		ActiveLink: aLink,
		Status:     StatusDisabled,
	}

	// Check if enabled (symlink in activeDir exists)
	if _, err := os.Lstat(aLink); err == nil {
		srv.Enabled = true
	}

	// Check if log service exists
	if fi, err := os.Stat(filepath.Join(sDir, "log")); err == nil && fi.IsDir() {
		srv.HasLog = true
	}

	if !srv.Enabled {
		srv.StatusRaw = "disabled"
		return srv, nil
	}

	// Run 'sv status <name>' or 'sv status <aLink>'
	out, err := exec.Command("sv", "status", aLink).CombinedOutput()
	raw := strings.TrimSpace(string(out))
	srv.StatusRaw = raw

	if err != nil && !strings.Contains(raw, "down:") && !strings.Contains(raw, "run:") {
		if strings.Contains(raw, "access denied") || strings.Contains(raw, "unable to open supervise") {
			srv.Status = StatusRunning
			srv.StatusRaw = "enabled (root permissions needed for PID/uptime)"
			return srv, nil
		}
		if strings.Contains(raw, "fail:") {
			srv.Status = StatusFailed
		} else {
			srv.Status = StatusUnknown
		}
		return srv, nil
	}

	if matches := runPattern.FindStringSubmatch(raw); len(matches) >= 4 {
		srv.Status = StatusRunning
		srv.PID, _ = strconv.Atoi(matches[2])
		srv.UptimeSeconds, _ = strconv.ParseInt(matches[3], 10, 64)
	} else if matches := downPattern.FindStringSubmatch(raw); len(matches) >= 3 {
		srv.Status = StatusStopped
		srv.UptimeSeconds, _ = strconv.ParseInt(matches[2], 10, 64)
	} else if strings.HasPrefix(raw, "run:") {
		srv.Status = StatusRunning
	} else if strings.HasPrefix(raw, "down:") {
		srv.Status = StatusStopped
	}

	return srv, nil
}
