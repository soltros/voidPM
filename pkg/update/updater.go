package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/voidlinux/voidpm/pkg/sys"
)

const (
	DefaultGitHubRepo = "soltros/voidPM"
	DefaultTargetPath = "/usr/bin/vpm"
)

type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type ReleaseInfo struct {
	TagName string         `json:"tag_name"`
	Name    string         `json:"name"`
	Body    string         `json:"body"`
	Assets  []ReleaseAsset `json:"assets"`
}

// Updater handles GitHub-api powered self-updating into /usr/bin/vpm
type Updater struct {
	Repo       string
	TargetPath string
	HTTPClient *http.Client
}

func NewUpdater(repo, targetPath string) *Updater {
	if repo == "" {
		repo = DefaultGitHubRepo
	}
	if targetPath == "" {
		targetPath = DefaultTargetPath
	}

	return &Updater{
		Repo:       repo,
		TargetPath: targetPath,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// FetchLatestRelease queries the GitHub API for the latest release metadata
func (u *Updater) FetchLatestRelease() (*ReleaseInfo, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", u.Repo)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub API request: %w", err)
	}

	req.Header.Set("User-Agent", "vpm-self-updater")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned HTTP status %d", resp.StatusCode)
	}

	var rel ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub release JSON: %w", err)
	}

	return &rel, nil
}

// FindBinaryURL picks the best binary download URL from release assets or fallback
func (u *Updater) FindBinaryURL(rel *ReleaseInfo) string {
	arch := runtime.GOARCH

	// 1. Look for asset explicitly matching OS/arch or binary name
	for _, asset := range rel.Assets {
		name := asset.Name
		if name == "vpm" || name == "vpm-linux-"+arch || name == "vpm_linux_"+arch {
			return asset.BrowserDownloadURL
		}
	}

	// 2. Look for any asset named vpm*
	for _, asset := range rel.Assets {
		if asset.Name != "" && asset.Name != "README.md" && asset.Name != "LICENSE" {
			return asset.BrowserDownloadURL
		}
	}

	// 3. Fallback direct release binary download URL
	return fmt.Sprintf("https://github.com/%s/releases/download/%s/vpm", u.Repo, rel.TagName)
}

// PerformUpdate downloads the binary and overwrites the target binary path (/usr/bin/vpm)
func (u *Updater) PerformUpdate() (string, error) {
	rel, err := u.FetchLatestRelease()
	downloadURL := ""
	tagName := "latest"

	if err == nil && rel != nil {
		tagName = rel.TagName
		downloadURL = u.FindBinaryURL(rel)
	} else {
		// Fallback to latest release download or main branch binary if API fails
		downloadURL = fmt.Sprintf("https://github.com/%s/releases/latest/download/vpm", u.Repo)
	}

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
	if err != nil {
		tmpFile.Close()
		// If release download failed, try raw binary fallback from main branch
		fallbackURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/vpm", u.Repo)
		reqFallback, _ := http.NewRequest("GET", fallbackURL, nil)
		reqFallback.Header.Set("User-Agent", "vpm-self-updater")
		resp, err = u.HTTPClient.Do(reqFallback)
		if err != nil {
			return "", fmt.Errorf("failed to download update binary from %s: %w", downloadURL, err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tmpFile.Close()
		return "", fmt.Errorf("download server returned HTTP %d for %s", resp.StatusCode, downloadURL)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		return "", fmt.Errorf("failed to write binary update to temp file: %w", err)
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
		// Replace directly as root
		if err := sys.RunElevated("cp", "-f", tmpPath, u.TargetPath); err != nil {
			return "", fmt.Errorf("failed to copy updated binary to %s: %w", u.TargetPath, err)
		}
		if err := os.Chmod(u.TargetPath, 0755); err != nil {
			return "", fmt.Errorf("failed to set permissions on %s: %w", u.TargetPath, err)
		}
	} else {
		// Use elevated privileges (sudo/doas) to copy and set permissions
		if err := sys.RunElevated("cp", "-f", tmpPath, u.TargetPath); err != nil {
			return "", fmt.Errorf("failed to overwrite %s: %w", u.TargetPath, err)
		}
		if err := sys.RunElevated("chmod", "0755", u.TargetPath); err != nil {
			return "", fmt.Errorf("failed to set executable permissions on %s: %w", u.TargetPath, err)
		}
	}

	return tagName, nil
}
