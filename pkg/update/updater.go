package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/voidlinux/voidpm/pkg/sys"
)

const (
	DefaultGitHubRepo   = "soltros/voidPM"
	DefaultGitHubBranch = "main"
	DefaultTargetPath   = "/usr/bin/vpm"
)

// Updater handles direct GitHub repository binary self-updating into /usr/bin/vpm
type Updater struct {
	Repo       string
	Branch     string
	TargetPath string
	HTTPClient *http.Client
}

func NewUpdater(repo, branch, targetPath string) *Updater {
	if repo == "" {
		repo = DefaultGitHubRepo
	}
	if branch == "" {
		branch = DefaultGitHubBranch
	}
	if targetPath == "" {
		targetPath = DefaultTargetPath
	}

	return &Updater{
		Repo:       repo,
		Branch:     branch,
		TargetPath: targetPath,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// GetRawBinaryURL returns the direct raw GitHub repository download URL for the binary
func (u *Updater) GetRawBinaryURL() string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/vpm", u.Repo, u.Branch)
}

// GetFallbackBinaryURL returns an alternate raw GitHub repository download URL
func (u *Updater) GetFallbackBinaryURL() string {
	return fmt.Sprintf("https://github.com/%s/raw/%s/vpm", u.Repo, u.Branch)
}

// PerformUpdate downloads the latest vpm binary directly from the repo main branch and overwrites /usr/bin/vpm
func (u *Updater) PerformUpdate() (string, error) {
	downloadURL := u.GetRawBinaryURL()

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "vpm-self-update-*.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file for update: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("User-Agent", "vpm-self-updater")

	resp, err := u.HTTPClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		// Try secondary raw fallback URL
		downloadURL = u.GetFallbackBinaryURL()
		reqFallback, errFallback := http.NewRequest("GET", downloadURL, nil)
		if errFallback != nil {
			tmpFile.Close()
			return "", fmt.Errorf("failed to create fallback download request: %w", errFallback)
		}
		reqFallback.Header.Set("User-Agent", "vpm-self-updater")

		resp, err = u.HTTPClient.Do(reqFallback)
		if err != nil {
			tmpFile.Close()
			return "", fmt.Errorf("failed to download update binary from repository: %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tmpFile.Close()
		return "", fmt.Errorf("repository returned HTTP %d when downloading %s", resp.StatusCode, downloadURL)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		return "", fmt.Errorf("failed to write downloaded binary to temp file: %w", err)
	}

	// Set executable permissions on temp file
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return "", fmt.Errorf("failed to set executable mode on updated binary: %w", err)
	}

	// Ensure target directory exists
	targetDir := filepath.Dir(u.TargetPath)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := sys.RunElevated("mkdir", "-p", targetDir); err != nil {
			return "", fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
		}
	}

	// Overwrite target binary (/usr/bin/vpm)
	if sys.IsRoot() {
		if err := sys.RunElevated("cp", "-f", tmpPath, u.TargetPath); err != nil {
			return "", fmt.Errorf("failed to copy updated binary to %s: %w", u.TargetPath, err)
		}
		if err := os.Chmod(u.TargetPath, 0755); err != nil {
			return "", fmt.Errorf("failed to set permissions on %s: %w", u.TargetPath, err)
		}
	} else {
		if err := sys.RunElevated("cp", "-f", tmpPath, u.TargetPath); err != nil {
			return "", fmt.Errorf("failed to overwrite %s: %w", u.TargetPath, err)
		}
		if err := sys.RunElevated("chmod", "0755", u.TargetPath); err != nil {
			return "", fmt.Errorf("failed to set executable permissions on %s: %w", u.TargetPath, err)
		}
	}

	return fmt.Sprintf("%s (%s)", u.Repo, u.Branch), nil
}
