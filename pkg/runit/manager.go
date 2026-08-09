package runit

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/voidlinux/voidpm/pkg/sys"
)

const (
	DefaultSvDir     = "/etc/sv"
	DefaultActiveDir = "/var/service"
)

type Manager struct {
	SvDir     string
	ActiveDir string
	IsUser    bool
}

func NewManager() *Manager {
	return &Manager{
		SvDir:     DefaultSvDir,
		ActiveDir: DefaultActiveDir,
		IsUser:    false,
	}
}

func NewUserManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	userSvDir := filepath.Join(homeDir, ".config", "sv")
	userActiveDir := filepath.Join(homeDir, "service")
	if _, err := os.Stat(userActiveDir); os.IsNotExist(err) {
		userActiveDir = filepath.Join(homeDir, ".config", "service")
	}

	return &Manager{
		SvDir:     userSvDir,
		ActiveDir: userActiveDir,
		IsUser:    true,
	}, nil
}

// ListServices fetches all available services from SvDir and checks their status
func (m *Manager) ListServices() ([]*Service, error) {
	entries, err := os.ReadDir(m.SvDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read service directory %s: %w", m.SvDir, err)
	}

	var services []*Service
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		srv, err := InspectService(entry.Name(), m.SvDir, m.ActiveDir)
		if err != nil {
			continue
		}
		srv.IsUser = m.IsUser
		services = append(services, srv)
	}

	if activeEntries, err := os.ReadDir(m.ActiveDir); err == nil {
		existing := make(map[string]bool)
		for _, s := range services {
			existing[s.Name] = true
		}
		for _, entry := range activeEntries {
			name := entry.Name()
			if !existing[name] && !strings.HasPrefix(name, ".") {
				srv, err := InspectService(name, m.SvDir, m.ActiveDir)
				if err == nil {
					srv.IsUser = m.IsUser
					services = append(services, srv)
				}
			}
		}
	}

	sort.Slice(services, func(i, j int) bool {
		if services[i].Enabled != services[j].Enabled {
			return services[i].Enabled
		}
		return services[i].Name < services[j].Name
	})

	return services, nil
}

// EnableService links /etc/sv/<name> to /var/service/<name>
func (m *Manager) EnableService(name string) error {
	svPath := filepath.Join(m.SvDir, name)
	activePath := filepath.Join(m.ActiveDir, name)

	if _, err := os.Stat(svPath); os.IsNotExist(err) {
		return fmt.Errorf("service '%s' does not exist in %s", name, m.SvDir)
	}

	if _, err := os.Lstat(activePath); err == nil {
		return fmt.Errorf("service '%s' is already enabled in %s", name, m.ActiveDir)
	}

	if m.IsUser {
		os.MkdirAll(m.ActiveDir, 0755)
		return os.Symlink(svPath, activePath)
	}

	_, err := sys.RunElevatedCombined("ln", "-s", svPath, activePath)
	return err
}

// DisableService removes symlink /var/service/<name>
func (m *Manager) DisableService(name string) error {
	activePath := filepath.Join(m.ActiveDir, name)

	if _, err := os.Lstat(activePath); os.IsNotExist(err) {
		return fmt.Errorf("service '%s' is not enabled in %s", name, m.ActiveDir)
	}

	_ = m.Stop(name)

	if m.IsUser {
		return os.Remove(activePath)
	}

	_, err := sys.RunElevatedCombined("rm", activePath)
	return err
}

// ExecuteAction runs 'sv <action> <service>'
func (m *Manager) ExecuteAction(action, name string) error {
	activePath := filepath.Join(m.ActiveDir, name)
	target := activePath
	if _, err := os.Lstat(activePath); os.IsNotExist(err) {
		target = filepath.Join(m.SvDir, name)
	}

	if m.IsUser {
		cmd := exec.Command("sv", action, target)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("sv %s %s failed: %s (%w)", action, name, strings.TrimSpace(string(out)), err)
		}
		return nil
	}

	out, err := sys.RunElevatedCombined("sv", action, target)
	if err != nil {
		return fmt.Errorf("sv %s %s failed: %s (%w)", action, name, out, err)
	}
	return nil
}

func (m *Manager) Start(name string) error   { return m.ExecuteAction("up", name) }
func (m *Manager) Stop(name string) error    { return m.ExecuteAction("down", name) }
func (m *Manager) Restart(name string) error { return m.ExecuteAction("restart", name) }
func (m *Manager) Reload(name string) error  { return m.ExecuteAction("reload", name) }
func (m *Manager) Check(name string) error   { return m.ExecuteAction("check", name) }

// ResolveLogFile finds the log file compliant with Void Linux logging documentation
func (m *Manager) ResolveLogFile(name string) (string, error) {
	if m.IsUser {
		home, _ := os.UserHomeDir()
		candidates := []string{
			filepath.Join(m.ActiveDir, name, "log", "main", "current"),
			filepath.Join(m.SvDir, name, "log", "main", "current"),
			filepath.Join(home, ".local", "state", name, "log"),
		}
		for _, cand := range candidates {
			if _, err := os.Stat(cand); err == nil {
				return cand, nil
			}
		}
		return "", fmt.Errorf("no log file found for user service '%s'", name)
	}

	candidates := []string{
		filepath.Join(m.ActiveDir, name, "log", "main", "current"),
		filepath.Join(m.SvDir, name, "log", "main", "current"),
		filepath.Join("/var/log", name, "current"),
		fmt.Sprintf("/var/log/%s.log", name),
		filepath.Join("/var/log/socklog", name, "current"),
		filepath.Join("/var/log/socklog", name, "everything"),
		"/var/log/socklog/everything/current",
		"/var/log/messages",
		"/var/log/dmesg.log",
	}

	for _, cand := range candidates {
		if _, err := os.Stat(cand); err == nil {
			return cand, nil
		}
	}

	return "", fmt.Errorf("no log file found for service or log target '%s'", name)
}

var tai64nRegex = regexp.MustCompile(`^@([0-9a-fA-F]{24})\s+(.*)$`)

// FormatLogLine parses TAI64N timestamps (svlogd) into human-readable local time
func FormatLogLine(line string) string {
	matches := tai64nRegex.FindStringSubmatch(line)
	if len(matches) == 3 {
		taiHex := matches[1]
		rest := matches[2]
		if len(taiHex) >= 16 {
			// Extract TAI64 seconds (first 16 hex chars = 8 bytes)
			secHex := taiHex[:16]
			secVal, err := strconv.ParseUint(secHex, 16, 64)
			if err == nil && secVal >= 0x4000000000000000 {
				unixSec := int64(secVal - 0x4000000000000000 - 10)
				t := time.Unix(unixSec, 0).Local()
				return fmt.Sprintf("%s  %s", t.Format("2006-01-02 15:04:05"), rest)
			}
		}
	}
	return line
}

// TailLogs reads and formats log entries for a service
func (m *Manager) TailLogs(name string, lines int) ([]string, error) {
	logPath, err := m.ResolveLogFile(name)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", logPath, err)
	}
	defer file.Close()

	var allLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allLines = append(allLines, FormatLogLine(scanner.Text()))
	}

	if len(allLines) <= lines {
		return allLines, nil
	}
	return allLines[len(allLines)-lines:], nil
}

// StreamLogs continuously streams live log output to writer
func (m *Manager) StreamLogs(name string, out io.Writer) error {
	logPath, err := m.ResolveLogFile(name)
	if err != nil {
		return err
	}

	// Use tail -f or native loop for continuous streaming
	cmd := exec.Command("tail", "-n", "50", "-f", logPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Fprintln(out, FormatLogLine(scanner.Text()))
	}
	return cmd.Wait()
}
