package sys

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type SystemInfo struct {
	OS            string
	Arch          string
	Kernel        string
	VoidVersion   string
	TotalServices int
	ActiveServices int
	InstalledPkgs int
	OrphanPkgs    int
}

func GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{
		OS:   "Void Linux",
		Arch: runtime.GOARCH,
	}

	// Kernel version
	if out, err := exec.Command("uname", "-r").Output(); err == nil {
		info.Kernel = strings.TrimSpace(string(out))
	}

	// Void OS release if available
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				info.VoidVersion = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
			}
		}
	}
	if info.VoidVersion == "" {
		info.VoidVersion = "Void Linux"
	}

	// Installed package count
	if out, err := exec.Command("xbps-query", "-l").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		count := 0
		for _, line := range lines {
			if strings.HasPrefix(line, "ii ") {
				count++
			}
		}
		info.InstalledPkgs = count
	}

	// Orphan package count
	if out, err := exec.Command("xbps-remove", "-o", "-n").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		count := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				count++
			}
		}
		info.OrphanPkgs = count
	}

	// Services count
	if entries, err := os.ReadDir("/etc/sv"); err == nil {
		info.TotalServices = len(entries)
	}
	if entries, err := os.ReadDir("/var/service"); err == nil {
		info.ActiveServices = len(entries)
	}

	return info, nil
}
