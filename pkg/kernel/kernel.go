package kernel

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/voidlinux/voidpm/pkg/sys"
)

type KernelPkg struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	IsRunning   bool   `json:"is_running"`
	IsMeta      bool   `json:"is_meta"`
	IsPurgeable bool   `json:"is_purgeable"`
}

type SystemKernels struct {
	RunningKernel string      `json:"running_kernel"`
	Installed     []KernelPkg `json:"installed"`
	OldPurgeable  []string    `json:"old_purgeable"`
}

var kernelPkgRegex = regexp.MustCompile(`^(linux[0-9.]*|linux-lts|linux-mainline)$`)

// GetSystemKernels collects detailed kernel state in Void Linux
func GetSystemKernels() (*SystemKernels, error) {
	sk := &SystemKernels{}

	// Currently running kernel
	if out, err := exec.Command("uname", "-r").Output(); err == nil {
		sk.RunningKernel = strings.TrimSpace(string(out))
	}

	// Purgeable kernels via vkpurge
	if out, err := exec.Command("vkpurge", "list").Output(); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			k := strings.TrimSpace(scanner.Text())
			if k != "" {
				sk.OldPurgeable = append(sk.OldPurgeable, k)
			}
		}
	}

	// Installed kernel packages via xbps-query
	cmd := exec.Command("xbps-query", "-l")
	out, err := cmd.Output()
	if err == nil {
		purgeMap := make(map[string]bool)
		for _, p := range sk.OldPurgeable {
			purgeMap[p] = true
		}

		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "ii ") {
				parts := strings.Fields(line[3:])
				if len(parts) >= 1 {
					pkgFull := parts[0]
					idx := strings.LastIndex(pkgFull, "-")
					if idx > 0 {
						name := pkgFull[:idx]
						ver := pkgFull[idx+1:]

						if kernelPkgRegex.MatchString(name) {
							isRun := strings.HasPrefix(sk.RunningKernel, ver) || strings.Contains(sk.RunningKernel, ver)
							isMeta := name == "linux" || name == "linux-lts" || name == "linux-mainline"

							sk.Installed = append(sk.Installed, KernelPkg{
								Name:        name,
								Version:     ver,
								IsRunning:   isRun,
								IsMeta:      isMeta,
								IsPurgeable: purgeMap[ver],
							})
						}
					}
				}
			}
		}
	}

	return sk, nil
}

// Reconfigure re-runs dracut initramfs generation & bootloader hooks (xbps-reconfigure -f linux...)
func Reconfigure(kernelPkg string) error {
	if kernelPkg == "" {
		kernelPkg = "linux"
	}
	fmt.Printf("--> Reconfiguring kernel & initramfs for '%s'...\n", kernelPkg)
	return sys.RunElevated("xbps-reconfigure", "-f", kernelPkg)
}

// ReconfigureAll reconfigures all installed kernel packages
func ReconfigureAll() error {
	sk, err := GetSystemKernels()
	if err != nil {
		return err
	}

	for _, k := range sk.Installed {
		fmt.Printf("--> Reconfiguring kernel %s (%s)...\n", k.Name, k.Version)
		if err := sys.RunElevated("xbps-reconfigure", "-f", k.Name); err != nil {
			fmt.Printf("Warning: failed to reconfigure %s: %v\n", k.Name, err)
		}
	}
	return nil
}

// RegenerateInitramfs runs dracut --regenerate-all --force
func RegenerateInitramfs() error {
	if _, err := exec.LookPath("dracut"); err != nil {
		return fmt.Errorf("dracut utility not found on system")
	}
	fmt.Println("--> Regenerating all initramfs images via dracut...")
	return sys.RunElevated("dracut", "--regenerate-all", "--force")
}

// Purge removes old unused kernel versions using vkpurge rm <target>
func Purge(target string) error {
	if _, err := exec.LookPath("vkpurge"); err != nil {
		return fmt.Errorf("vkpurge utility is not installed")
	}

	if target == "" || target == "all" {
		fmt.Println("--> Purging all unused old kernel versions...")
		return sys.RunElevated("vkpurge", "rm", "all")
	}

	fmt.Printf("--> Purging kernel version %s...\n", target)
	return sys.RunElevated("vkpurge", "rm", target)
}

// SwitchFlavor changes kernel metapackage (e.g. linux-lts, linux6.6)
func SwitchFlavor(flavor string) error {
	if !strings.HasPrefix(flavor, "linux") {
		flavor = "linux" + flavor
	}

	headersPkg := flavor + "-headers"
	fmt.Printf("--> Installing kernel flavor '%s' and headers '%s'...\n", flavor, headersPkg)
	if err := sys.RunElevated("xbps-install", "-Sy", flavor, headersPkg); err != nil {
		// Fallback to just kernel flavor if headers fail
		if err := sys.RunElevated("xbps-install", "-Sy", flavor); err != nil {
			return err
		}
	}

	fmt.Println("--> Reconfiguring bootloader and initramfs...")
	return Reconfigure(flavor)
}

// ListAvailableKernels searches repositories for available Linux kernel branches
func ListAvailableKernels() ([]KernelPkg, error) {
	cmd := exec.Command("xbps-query", "-Rs", "linux")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to query kernel packages: %w", err)
	}

	var kernels []KernelPkg
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	// Strictly match linux, linux-lts, linux-mainline, linux6.x, linux5.x
	re := regexp.MustCompile(`^(\[\*\]|\[-\])\s+(linux[0-9.]*|linux-lts|linux-mainline|linux)-([0-9][^\s]*)\s+(.*)$`)

	seen := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) == 5 {
			name := matches[2]
			ver := matches[3]

			if seen[name] || strings.Contains(name, "firmware") || strings.Contains(name, "driver") || strings.Contains(name, "tools") {
				continue
			}
			seen[name] = true

			kernels = append(kernels, KernelPkg{
				Name:      name,
				Version:   ver,
				IsRunning: false,
				IsMeta:    name == "linux" || name == "linux-lts" || name == "linux-mainline",
			})
		}
	}
	return kernels, nil
}
