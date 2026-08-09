package xbps

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/voidlinux/voidpm/pkg/sys"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

// Search searches remote and local repositories for packages matching query
func (c *Client) Search(query string) ([]Package, error) {
	cmd := exec.Command("xbps-query", "-Rs", query)
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		return nil, fmt.Errorf("no packages found for query '%s'", query)
	}

	var pkgs []Package
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	// Output format: "[*] pkgname-version Short description" or "[-] pkgname-version Short description"
	pkgRegex := regexp.MustCompile(`^(\[\*\]|\[-\])\s+([^\s]+)\s+(.*)$`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := pkgRegex.FindStringSubmatch(line)
		if len(matches) == 4 {
			installed := matches[1] == "[*]"
			pkgFull := matches[2]
			desc := matches[3]

			// Separate name and version (version starts after last hyphen or underscore)
			name, ver := splitPkgVersion(pkgFull)

			pkgs = append(pkgs, Package{
				Name:      name,
				Version:   ver,
				ShortDesc: desc,
				Installed: installed,
			})
		}
	}

	return pkgs, nil
}

// ListInstalled returns all installed packages
func (c *Client) ListInstalled() ([]Package, error) {
	cmd := exec.Command("xbps-query", "-l")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list installed packages: %w", err)
	}

	// Fetch hold packages set
	holds, _ := c.ListHolds()
	holdMap := make(map[string]bool)
	for _, h := range holds {
		holdMap[h] = true
	}

	var pkgs []Package
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	// Format: "ii pkgname-version Short description"
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ii ") {
			parts := strings.SplitN(line[3:], " ", 2)
			if len(parts) >= 1 {
				pkgFull := parts[0]
				desc := ""
				if len(parts) == 2 {
					desc = strings.TrimSpace(parts[1])
				}
				name, ver := splitPkgVersion(pkgFull)
				pkgs = append(pkgs, Package{
					Name:      name,
					Version:   ver,
					ShortDesc: desc,
					Installed: true,
					OnHold:    holdMap[name],
				})
			}
		}
	}

	return pkgs, nil
}

// ListOrphans returns list of orphaned packages
func (c *Client) ListOrphans() ([]Package, error) {
	cmd := exec.Command("xbps-remove", "-o", "-n")
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		return nil, nil
	}

	var pkgs []Package
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Remove ") {
			continue
		}
		name, ver := splitPkgVersion(line)
		pkgs = append(pkgs, Package{
			Name:      name,
			Version:   ver,
			Installed: true,
			Orphan:    true,
		})
	}
	return pkgs, nil
}

// ListHolds lists package holds
func (c *Client) ListHolds() ([]string, error) {
	cmd := exec.Command("xbps-query", "-H")
	out, err := cmd.Output()
	if err != nil {
		return nil, nil
	}
	var holds []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			name, _ := splitPkgVersion(line)
			holds = append(holds, name)
		}
	}
	return holds, nil
}

// GetInfo fetches full package info (xbps-query -RS <pkg> or xbps-query -S <pkg>)
func (c *Client) GetInfo(pkgName string) (*Package, error) {
	cmd := exec.Command("xbps-query", "-RS", pkgName)
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		cmd = exec.Command("xbps-query", "-S", pkgName)
		out, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to query package info for '%s': %w", pkgName, err)
		}
	}

	pkg := &Package{Name: pkgName}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		kv := strings.SplitN(line, ": ", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			switch key {
			case "pkgver":
				_, pkg.Version = splitPkgVersion(val)
			case "short_desc":
				pkg.ShortDesc = val
			case "installed_size":
				pkg.InstalledSize = val
			case "filename_size", "download_size":
				pkg.DownloadSize = val
			case "repository":
				pkg.Repository = val
			case "maintainer":
				pkg.Maintainer = val
			case "homepage":
				pkg.Homepage = val
			case "license":
				pkg.License = val
			case "architecture":
				pkg.Architecture = val
			case "run_depends":
				pkg.Dependencies = strings.Fields(val)
			case "state":
				pkg.Installed = strings.Contains(val, "installed")
			}
		}
	}
	return pkg, nil
}

// GetFiles lists files owned by an installed package
func (c *Client) GetFiles(pkgName string) ([]string, error) {
	cmd := exec.Command("xbps-query", "-f", pkgName)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list files for package '%s': %w", pkgName, err)
	}

	var files []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		f := strings.TrimSpace(scanner.Text())
		if f != "" {
			files = append(files, f)
		}
	}
	return files, nil
}

// WhoOwns finds which package owns a file path
func (c *Client) WhoOwns(filepath string) (string, error) {
	cmd := exec.Command("xbps-query", "-o", filepath)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("no package owns file '%s'", filepath)
	}
	return strings.TrimSpace(string(out)), nil
}

// ListExplicit lists explicitly installed packages (xbps-query -m)
func (c *Client) ListExplicit() ([]string, error) {
	cmd := exec.Command("xbps-query", "-m")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to query explicit packages: %w", err)
	}

	var explicit []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			name, _ := splitPkgVersion(line)
			explicit = append(explicit, name)
		}
	}
	return explicit, nil
}

// ImportPackages installs packages listed in slice
func (c *Client) ImportPackages(pkgs []string, yes bool) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := []string{"xbps-install", "-S"}
	if yes {
		args = append(args, "-y")
	}
	args = append(args, pkgs...)
	return sys.RunElevated(args...)
}

// GetPendingUpdatesList queries xbps-install -un to retrieve list of pending updates
func (c *Client) GetPendingUpdatesList() ([]string, error) {
	cmd := exec.Command("xbps-install", "-un")
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		return nil, nil
	}

	var updates []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "Name") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				name, _ := splitPkgVersion(parts[0])
				updates = append(updates, name)
			}
		}
	}
	return updates, nil
}

// Install installs or updates packages with interactive output
func (c *Client) Install(pkgs []string, sync bool) error {
	return c.InstallWithOptions(pkgs, sync, false)
}

func (c *Client) InstallWithOptions(pkgs []string, sync bool, yes bool) error {
	args := []string{"xbps-install"}
	if sync {
		args = append(args, "-S")
	}
	if yes {
		args = append(args, "-y")
	}
	args = append(args, pkgs...)
	return sys.RunElevated(args...)
}

// Remove removes packages
func (c *Client) Remove(pkgs []string, recursive bool) error {
	return c.RemoveWithOptions(pkgs, recursive, false)
}

func (c *Client) RemoveWithOptions(pkgs []string, recursive bool, yes bool) error {
	args := []string{"xbps-remove"}
	if recursive {
		args = append(args, "-R")
	}
	if yes {
		args = append(args, "-y")
	}
	args = append(args, pkgs...)
	return sys.RunElevated(args...)
}

// UpdateSystem performs a full system update (xbps-install -Su)
func (c *Client) UpdateSystem() error {
	return c.UpdateSystemWithOptions(false)
}

func (c *Client) UpdateSystemWithOptions(yes bool) error {
	argsXBPS := []string{"xbps-install", "-Su", "xbps"}
	argsFull := []string{"xbps-install", "-Su"}
	if yes {
		argsXBPS = append(argsXBPS, "-y")
		argsFull = append(argsFull, "-y")
	}
	_ = sys.RunElevated(argsXBPS...)
	return sys.RunElevated(argsFull...)
}

// SyncRepos updates repository index files (xbps-install -S)
func (c *Client) SyncRepos() error {
	return sys.RunElevated("xbps-install", "-S")
}

// Hold puts a package on hold
func (c *Client) Hold(pkg string) error {
	return sys.RunElevated("xbps-pkgdb", "-m", "hold", pkg)
}

// Unhold removes hold on package
func (c *Client) Unhold(pkg string) error {
	return sys.RunElevated("xbps-pkgdb", "-m", "unhold", pkg)
}

func splitPkgVersion(pkgFull string) (string, string) {
	idx := strings.LastIndex(pkgFull, "-")
	if idx <= 0 {
		return pkgFull, ""
	}
	return pkgFull[:idx], pkgFull[idx+1:]
}
