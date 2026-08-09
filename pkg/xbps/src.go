package xbps

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/voidlinux/voidpm/pkg/sys"
)

type SrcPackage struct {
	Name         string `json:"name"`
	ShortDesc    string `json:"short_desc"`
	IsRestricted bool   `json:"is_restricted"`
	IsInstalled  bool   `json:"is_installed"`
	License      string `json:"license,omitempty"`
}

type SrcManager struct {
	RepoDir string
}

func NewSrcManager() *SrcManager {
	dir := os.Getenv("VOIDPM_SRC_DIR")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dotDir := filepath.Join(home, ".void-packages")
		legacyDir := filepath.Join(home, "void-packages")
		if _, err := os.Stat(dotDir); err == nil {
			dir = dotDir
		} else if _, err := os.Stat(legacyDir); err == nil {
			dir = legacyDir
		} else {
			dir = dotDir
		}
	}
	return &SrcManager{RepoDir: dir}
}

// IsSetup checks if void-packages git repo exists and xbps-src script is executable
func (s *SrcManager) IsSetup() bool {
	script := filepath.Join(s.RepoDir, "xbps-src")
	fi, err := os.Stat(script)
	return err == nil && !fi.IsDir()
}

// EnsureSetup auto-initializes void-packages if missing
func (s *SrcManager) EnsureSetup() error {
	if !s.IsSetup() {
		fmt.Printf("[INFO] void-packages repository not found at %s. Initializing now...\n", s.RepoDir)
		return s.Setup()
	}
	return nil
}

// Setup clones void-packages repo into hidden ~/.void-packages and runs binary-bootstrap
func (s *SrcManager) Setup() error {
	if s.IsSetup() {
		return fmt.Errorf("void-packages repository already initialized at %s", s.RepoDir)
	}

	fmt.Printf("--> Cloning void-packages repository into %s...\n", s.RepoDir)
	os.MkdirAll(filepath.Dir(s.RepoDir), 0755)
	cmd := exec.Command("git", "clone", "--depth=1", "https://github.com/void-linux/void-packages.git", s.RepoDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone void-packages repo: %w", err)
	}

	// Auto-enable restricted packages in etc/conf
	_ = s.EnableRestrictedInConfig()

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
	if err := s.EnsureSetup(); err != nil {
		return err
	}

	fmt.Printf("--> Updating void-packages repository in %s...\n", s.RepoDir)
	cmd := exec.Command("git", "pull", "--rebase")
	cmd.Dir = s.RepoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// EnableRestrictedInConfig ensures XBPS_ALLOW_RESTRICTED=yes is in void-packages/etc/conf
func (s *SrcManager) EnableRestrictedInConfig() error {
	confPath := filepath.Join(s.RepoDir, "etc", "conf")
	os.MkdirAll(filepath.Dir(confPath), 0755)

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

// CheckIfTemplateRestricted inspects srcpkgs/<pkg>/template for restricted=yes
func (s *SrcManager) CheckIfTemplateRestricted(pkgName string) bool {
	tmplPath := filepath.Join(s.RepoDir, "srcpkgs", pkgName, "template")
	file, err := os.Open(tmplPath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "restricted=") {
			val := strings.TrimPrefix(line, "restricted=")
			val = strings.Trim(val, `"`)
			val = strings.Trim(val, `'`)
			if val == "yes" || val == "1" || val == "true" {
				return true
			}
		}
	}
	return false
}

// Build compiles a package from source using ./xbps-src pkg <name> with auto-restricted handling
func (s *SrcManager) Build(pkgName string, allowRestricted bool) error {
	if err := s.EnsureSetup(); err != nil {
		return err
	}

	// Auto-detect restricted template flag
	if !allowRestricted {
		allowRestricted = s.CheckIfTemplateRestricted(pkgName)
	}

	if allowRestricted {
		_ = s.EnableRestrictedInConfig()
	}

	args := []string{"./xbps-src"}
	if allowRestricted {
		args = append(args, "-m", "restricted")
	}
	args = append(args, "pkg", pkgName)

	fmt.Printf("--> Building %s from source (repo: %s)...\n", pkgName, s.RepoDir)
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

// SearchTemplates searches for source package templates in srcpkgs/ with descriptions & restricted status
func (s *SrcManager) SearchTemplates(query string) ([]SrcPackage, error) {
	if err := s.EnsureSetup(); err != nil {
		return nil, err
	}

	srcpkgsDir := filepath.Join(s.RepoDir, "srcpkgs")
	entries, err := os.ReadDir(srcpkgsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read srcpkgs in %s: %w", s.RepoDir, err)
	}

	// Fetch installed packages map for quick status check
	client := NewClient()
	installedPkgs, _ := client.ListInstalled()
	installedMap := make(map[string]bool)
	for _, p := range installedPkgs {
		installedMap[p.Name] = true
	}

	queryLower := strings.ToLower(query)
	var matches []SrcPackage

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pkgName := entry.Name()

		if strings.Contains(strings.ToLower(pkgName), queryLower) {
			sp := SrcPackage{
				Name:        pkgName,
				IsInstalled: installedMap[pkgName],
			}

			// Read template metadata
			tmplPath := filepath.Join(srcpkgsDir, pkgName, "template")
			if file, err := os.Open(tmplPath); err == nil {
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					line := strings.TrimSpace(scanner.Text())
					if strings.HasPrefix(line, "short_desc=") {
						sp.ShortDesc = strings.Trim(strings.TrimPrefix(line, "short_desc="), `"`)
					} else if strings.HasPrefix(line, "restricted=") {
						val := strings.Trim(strings.TrimPrefix(line, "restricted="), `"`)
						sp.IsRestricted = val == "yes" || val == "1"
					} else if strings.HasPrefix(line, "license=") {
						sp.License = strings.Trim(strings.TrimPrefix(line, "license="), `"`)
					}
				}
				file.Close()
			}

			matches = append(matches, sp)
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Name < matches[j].Name
	})

	return matches, nil
}
