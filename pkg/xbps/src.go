package xbps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/voidlinux/voidpm/pkg/sys"
)

type SrcManager struct {
	RepoDir string
}

func NewSrcManager() *SrcManager {
	dir := os.Getenv("VOIDPM_SRC_DIR")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, "void-packages")
	}
	return &SrcManager{RepoDir: dir}
}

// IsSetup checks if void-packages git repo exists and xbps-src script is executable
func (s *SrcManager) IsSetup() bool {
	script := filepath.Join(s.RepoDir, "xbps-src")
	fi, err := os.Stat(script)
	return err == nil && !fi.IsDir()
}

// Setup clones void-packages repo and runs binary-bootstrap
func (s *SrcManager) Setup() error {
	if s.IsSetup() {
		return fmt.Errorf("void-packages repository already initialized at %s", s.RepoDir)
	}

	fmt.Printf("--> Cloning void-packages repository into %s...\n", s.RepoDir)
	cmd := exec.Command("git", "clone", "--depth=1", "https://github.com/void-linux/void-packages.git", s.RepoDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone void-packages repo: %w", err)
	}

	fmt.Println("--> Running binary-bootstrap...")
	bsCmd := exec.Command("./xbps-src", "binary-bootstrap")
	bsCmd.Dir = s.RepoDir
	bsCmd.Stdout = os.Stdout
	bsCmd.Stderr = os.Stderr
	if err := bsCmd.Run(); err != nil {
		return fmt.Errorf("failed to run binary-bootstrap: %w", err)
	}

	return nil
}

// Sync updates void-packages repository via git pull
func (s *SrcManager) Sync() error {
	if !s.IsSetup() {
		return fmt.Errorf("void-packages repository not found at %s. Run 'vpm src setup' first", s.RepoDir)
	}

	fmt.Println("--> Updating void-packages repository...")
	cmd := exec.Command("git", "pull", "--rebase")
	cmd.Dir = s.RepoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// EnableRestrictedInConfig ensures XBPS_ALLOW_RESTRICTED=yes is in void-packages/etc/conf
func (s *SrcManager) EnableRestrictedInConfig() error {
	if !s.IsSetup() {
		return fmt.Errorf("void-packages repository not found at %s. Run 'vpm src setup' first", s.RepoDir)
	}

	confPath := filepath.Join(s.RepoDir, "etc", "conf")
	data, _ := os.ReadFile(confPath)
	content := string(data)

	if strings.Contains(content, "XBPS_ALLOW_RESTRICTED=yes") {
		return nil
	}

	f, err := os.OpenFile(confPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %w", confPath, err)
	}
	defer f.Close()

	_, err = f.WriteString("\nXBPS_ALLOW_RESTRICTED=yes\n")
	return err
}

// Build compiles a package from source using ./xbps-src pkg <name>
func (s *SrcManager) Build(pkgName string, allowRestricted bool) error {
	if !s.IsSetup() {
		return fmt.Errorf("void-packages repository not found at %s. Run 'vpm src setup' first", s.RepoDir)
	}

	if allowRestricted {
		_ = s.EnableRestrictedInConfig()
	}

	args := []string{"./xbps-src"}
	if allowRestricted {
		args = append(args, "-m", "restricted")
	}
	args = append(args, "pkg", pkgName)

	fmt.Printf("--> Building %s from source...\n", pkgName)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = s.RepoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "XBPS_ALLOW_RESTRICTED=yes")

	return cmd.Run()
}

// InstallBuilt installs a package built in hostdir/binpkgs
func (s *SrcManager) InstallBuilt(pkgName string) error {
	if !s.IsSetup() {
		return fmt.Errorf("void-packages repository not found at %s", s.RepoDir)
	}

	binpkgsDir := filepath.Join(s.RepoDir, "hostdir", "binpkgs")
	repoArgs := []string{"--repository", binpkgsDir}

	// Add subdirectories like nonfree, multilib, multilib/nonfree if they exist
	subdirs := []string{"nonfree", "multilib", "multilib/nonfree"}
	for _, sub := range subdirs {
		cand := filepath.Join(binpkgsDir, sub)
		if _, err := os.Stat(cand); err == nil {
			repoArgs = append(repoArgs, "--repository", cand)
		}
	}

	args := append([]string{"xbps-install", "-y"}, repoArgs...)
	args = append(args, pkgName)

	fmt.Printf("--> Installing built package %s...\n", pkgName)
	return sys.RunElevated(args...)
}

// SearchTemplates searches for source package templates in srcpkgs/
func (s *SrcManager) SearchTemplates(query string) ([]string, error) {
	if !s.IsSetup() {
		return nil, fmt.Errorf("void-packages repository not found at %s", s.RepoDir)
	}

	srcpkgsDir := filepath.Join(s.RepoDir, "srcpkgs")
	entries, err := os.ReadDir(srcpkgsDir)
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	var matches []string
	for _, entry := range entries {
		if entry.IsDir() && strings.Contains(strings.ToLower(entry.Name()), queryLower) {
			matches = append(matches, entry.Name())
		}
	}
	return matches, nil
}
