package kernel

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"sort"

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

	if out, err := exec.Command("uname", "-r").Output(); err == nil {
		sk.RunningKernel = strings.TrimSpace(string(out))
	}

	if out, err := exec.Command("vkpurge", "list").Output(); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			k := strings.TrimSpace(scanner.Text())
			if k != "" {
				sk.OldPurgeable = append(sk.OldPurgeable, k)
			}
		}
	}

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

// ResolveActualKernelPkg resolves metapackages (linux, linux-lts, linux-mainline) to concrete package (e.g. linux7.1)
func ResolveActualKernelPkg(flavor string) string {
	cmd := exec.Command("xbps-query", "-RS", flavor)
	out, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("xbps-query", "-S", flavor)
		out, err = cmd.Output()
	}

	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "run_depends:") {
				for scanner.Scan() {
					depLine := strings.TrimSpace(scanner.Text())
					if !strings.HasPrefix(depLine, "linux") {
						break
					}
					depName := strings.Fields(depLine)[0]
					depName = strings.Split(depName, ">=")[0]
					depName = strings.Split(depName, "<=")[0]
					depName = strings.Split(depName, "=")[0]
					if strings.HasPrefix(depName, "linux") && depName != "linux-base" && depName != flavor {
						return depName
					}
				}
			}
		}
	}
	return flavor
}

// Reconfigure re-runs dracut initramfs generation & bootloader hooks (xbps-reconfigure -f <pkg>)
func Reconfigure(kernelPkg string) error {
	if kernelPkg == "" {
		kernelPkg = "linux"
	}
	actual := ResolveActualKernelPkg(kernelPkg)

	fmt.Printf("--> Reconfiguring kernel package '%s' (%s)...\n", kernelPkg, actual)
	if err := sys.RunElevated("xbps-reconfigure", "-f", actual); err != nil {
		_ = sys.RunElevated("xbps-reconfigure", "-f", kernelPkg)
	}

	// Reconfigure DKMS packages if any installed
	ReconfigureDKMS()

	// Reconfigure NetworkManager & audio if installed
	_ = sys.RunElevated("xbps-reconfigure", "-f", "NetworkManager")

	return RegenerateInitramfs()
}

// ReconfigureDKMS reconfigures all installed DKMS module packages for the active kernel
func ReconfigureDKMS() {
	cmd := exec.Command("xbps-query", "-l")
	out, err := cmd.Output()
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ii ") && strings.Contains(line, "-dkms") {
			parts := strings.Fields(line[3:])
			if len(parts) >= 1 {
				pkgFull := parts[0]
				idx := strings.LastIndex(pkgFull, "-")
				if idx > 0 {
					name := pkgFull[:idx]
					fmt.Printf("--> Reconfiguring DKMS package '%s'...\n", name)
					_ = sys.RunElevated("xbps-reconfigure", "-f", name)
				}
			}
		}
	}
}

// ReconfigureAll reconfigures all installed kernel packages
func ReconfigureAll() error {
	sk, err := GetSystemKernels()
	if err != nil {
		return err
	}

	for _, k := range sk.Installed {
		if !k.IsMeta {
			fmt.Printf("--> Reconfiguring kernel %s (%s)...\n", k.Name, k.Version)
			if err := sys.RunElevated("xbps-reconfigure", "-f", k.Name); err != nil {
				fmt.Printf("Warning: failed to reconfigure %s: %v\n", k.Name, err)
			}
		}
	}

	ReconfigureDKMS()
	return RegenerateInitramfs()
}

// RegenerateInitramfs runs dracut --regenerate-all --force
func RegenerateInitramfs() error {
	if _, err := exec.LookPath("dracut"); err != nil {
		return fmt.Errorf("dracut utility not found on system")
	}
	fmt.Println("--> Regenerating all initramfs images via dracut...")
	return sys.RunElevated("dracut", "--regenerate-all", "--force")
}

// RemoveKernel safely uninstalls a specified kernel series or package
func RemoveKernel(kernelPkg string) error {
	if !strings.HasPrefix(kernelPkg, "linux") {
		kernelPkg = "linux" + kernelPkg
	}

	sk, err := GetSystemKernels()
	if err == nil {
		if strings.HasPrefix(sk.RunningKernel, strings.TrimPrefix(kernelPkg, "linux")) {
			return fmt.Errorf("cannot remove currently running kernel (%s). Switch/boot into another kernel first", sk.RunningKernel)
		}
	}

	actualPkg := ResolveActualKernelPkg(kernelPkg)
	headersPkg := kernelPkg + "-headers"
	actualHeadersPkg := actualPkg + "-headers"

	pkgsToRemove := []string{
		kernelPkg,
		headersPkg,
		actualPkg,
		actualHeadersPkg,
		"linux-mainline",
		"linux-mainline-headers",
		"linux-lts",
		"linux-lts-headers",
	}

	// Filter installed packages
	var validRemovals []string
	seen := make(map[string]bool)
	for _, p := range pkgsToRemove {
		if !seen[p] {
			seen[p] = true
			if out, err := exec.Command("xbps-query", "-S", p).Output(); err == nil && len(out) > 0 {
				// Verify if p is related to target kernel or metapackage pointing to target
				if strings.Contains(p, strings.TrimPrefix(kernelPkg, "linux")) || strings.Contains(p, strings.TrimPrefix(actualPkg, "linux")) || strings.Contains(p, "mainline") || strings.Contains(p, "lts") {
					validRemovals = append(validRemovals, p)
				}
			}
		}
	}

	if len(validRemovals) == 0 {
		return fmt.Errorf("kernel package '%s' is not installed", kernelPkg)
	}

	fmt.Printf("--> Removing kernel packages: %v...\n", validRemovals)
	args := append([]string{"xbps-remove", "-R", "-y"}, validRemovals...)
	if err := sys.RunElevated(args...); err != nil {
		return fmt.Errorf("failed to remove kernel packages: %w", err)
	}

	fmt.Println("--> Cleaning leftover boot files via vkpurge...")
	_ = sys.RunElevated("vkpurge", "rm", "all")

	fmt.Println("--> Reconfiguring bootloader & remaining initramfs images...")
	return ReconfigureAll()
}

// Purge removes old unused kernel versions using vkpurge rm <target>
func Purge(target string) error {
	return PurgeWithOptions(target, 0, false, false)
}

func PurgeWithOptions(target string, keep int, dryRun bool, yes bool) error {
	if _, err := exec.LookPath("vkpurge"); err != nil {
		return fmt.Errorf("vkpurge utility is not installed")
	}

	sk, err := GetSystemKernels()
	if err != nil {
		return err
	}

	oldKernels := sk.OldPurgeable
	if len(oldKernels) == 0 {
		fmt.Println("No obsolete kernels to purge.")
		return nil
	}

	var toPurge []string
	if keep > 0 {
		if len(oldKernels) <= keep {
			fmt.Printf("Found %d old kernel(s); keeping %d. Nothing to purge.\n", len(oldKernels), keep)
			return nil
		}
		toPurge = oldKernels[:len(oldKernels)-keep]
	} else if target != "" && target != "all" {
		toPurge = []string{target}
	} else {
		toPurge = oldKernels
	}

	if dryRun {
		fmt.Printf("[dry-run] The following kernel version(s) would be purged:\n%v\n", toPurge)
		return nil
	}

	if len(toPurge) == len(oldKernels) && (target == "" || target == "all") && keep == 0 {
		fmt.Println("--> Purging all unused old kernel versions...")
		return sys.RunElevated("vkpurge", "rm", "all")
	}

	for _, kver := range toPurge {
		fmt.Printf("--> Purging kernel version %s...\n", kver)
		if err := sys.RunElevated("vkpurge", "rm", kver); err != nil {
			return err
		}
	}

	return nil
}

// FindInstalledKernelModules discovers installed kernel module and driver packages (DKMS, nvidia, zfs, wifi)
func FindInstalledKernelModules() []string {
	cmd := exec.Command("xbps-query", "-l")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var modules []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	modRegex := regexp.MustCompile(`^(nvidia|zfs|v4l2loopback|wireguard|broadcom-wl|realtek|virtualbox-ose|tp_smapi|acpi_call|ddcci)(-.*)?$`)

	seen := make(map[string]bool)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ii ") {
			parts := strings.Fields(line[3:])
			if len(parts) >= 1 {
				pkgFull := parts[0]
				idx := strings.LastIndex(pkgFull, "-")
				if idx > 0 {
					name := pkgFull[:idx]
					if (modRegex.MatchString(name) || strings.HasSuffix(name, "-dkms")) && !seen[name] {
						seen[name] = true
						modules = append(modules, name)
					}
				}
			}
		}
	}
	return modules
}

// SwitchFlavor changes kernel series, resolving metapackages, preserving drivers, DKMS, and dracut
func SwitchFlavor(flavor string) error {
	if !strings.HasPrefix(flavor, "linux") {
		flavor = "linux" + flavor
	}

	actualPkg := ResolveActualKernelPkg(flavor)
	headersPkg := flavor + "-headers"
	actualHeadersPkg := actualPkg + "-headers"

	fmt.Printf("--> Switching kernel to series '%s' (concrete: %s)...\n", flavor, actualPkg)

	// Core kernel + firmware + audio topology bundle
	pkgsToInstall := []string{
		flavor,
		headersPkg,
		actualPkg,
		actualHeadersPkg,
		"linux-firmware",
		"wifi-firmware",
		"sof-firmware",
		"alsa-firmware",
		"alsa-ucm-conf",
	}

	// Preserve any existing hardware driver / DKMS packages
	driverModules := FindInstalledKernelModules()
	if len(driverModules) > 0 {
		fmt.Printf("--> Preserving existing driver/DKMS packages: %v...\n", driverModules)
		pkgsToInstall = append(pkgsToInstall, driverModules...)
	}

	var validPkgs []string
	seen := make(map[string]bool)
	for _, p := range pkgsToInstall {
		if !seen[p] {
			seen[p] = true
			validPkgs = append(validPkgs, p)
		}
	}

	args := append([]string{"xbps-install", "-Sy"}, validPkgs...)
	fmt.Printf("--> Installing kernel, headers, firmware, and drivers: %v...\n", validPkgs)
	if err := sys.RunElevated(args...); err != nil {
		_ = sys.RunElevated("xbps-install", "-Sy", flavor, actualPkg)
	}

	fmt.Println("--> Reconfiguring kernel, DKMS, udevd, NetworkManager, and dracut...")
	return Reconfigure(actualPkg)
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

	sort.Slice(kernels, func(i, j int) bool {
		return kernels[i].Name < kernels[j].Name
	})

	return kernels, nil
}
