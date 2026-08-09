package xbps

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/voidlinux/voidpm/pkg/sys"
)

type CleanResult struct {
	CacheCleaned   bool     `json:"cache_cleaned"`
	OrphansCleaned bool     `json:"orphans_cleaned"`
	KernelsPurged  bool     `json:"kernels_purged"`
	DryRun         bool     `json:"dry_run"`
	Details        []string `json:"details"`
}

type Cleaner struct{}

func NewCleaner() *Cleaner {
	return &Cleaner{}
}

// CleanCache removes obsolete package files from XBPS cache (xbps-remove -O / xbps-remove -c)
func (c *Cleaner) CleanCache(all bool) error {
	return c.CleanCacheWithOptions(all, false, false)
}

func (c *Cleaner) CleanCacheWithOptions(all bool, dryRun bool, yes bool) error {
	args := []string{"xbps-remove"}
	if all {
		args = append(args, "-c")
	} else {
		args = append(args, "-O")
	}
	if dryRun {
		args = append(args, "-n")
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	if yes {
		args = append(args, "-y")
	}
	return sys.RunElevated(args...)
}

// CleanOrphans removes orphaned packages (xbps-remove -o)
func (c *Cleaner) CleanOrphans() error {
	return c.CleanOrphansWithOptions(false, false)
}

func (c *Cleaner) CleanOrphansWithOptions(dryRun bool, yes bool) error {
	args := []string{"xbps-remove", "-o"}
	if dryRun {
		args = append(args, "-n")
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	if yes {
		args = append(args, "-y")
	}
	return sys.RunElevated(args...)
}

// ListOldKernels returns list of old kernels removable by vkpurge
func (c *Cleaner) ListOldKernels() (string, error) {
	if _, err := exec.LookPath("vkpurge"); err != nil {
		return "", fmt.Errorf("vkpurge utility is not installed")
	}

	cmd := exec.Command("vkpurge", "list")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// PurgeOldKernels removes old Linux kernels using vkpurge rm all
func (c *Cleaner) PurgeOldKernels() error {
	return c.PurgeOldKernelsWithOptions("all", false, false)
}

func (c *Cleaner) PurgeOldKernelsWithOptions(target string, dryRun bool, yes bool) error {
	if _, err := exec.LookPath("vkpurge"); err != nil {
		return fmt.Errorf("vkpurge utility is not installed")
	}

	if target == "" {
		target = "all"
	}

	if dryRun {
		list, err := c.ListOldKernels()
		if err != nil {
			return err
		}
		if list == "" {
			fmt.Println("[dry-run] No obsolete kernels to purge")
		} else {
			fmt.Printf("[dry-run] Obsolete kernels that would be purged:\n%s\n", list)
		}
		return nil
	}

	return sys.RunElevated("vkpurge", "rm", target)
}

// PerformAllCleanup cleans cache, removes orphans, and purges old kernels
func (c *Cleaner) PerformAllCleanup() error {
	_, err := c.PerformCleanup(true, true, true, false, false, false)
	return err
}

func (c *Cleaner) PerformCleanup(orphans, cache, kernels, allCache, dryRun, yes bool) (*CleanResult, error) {
	res := &CleanResult{
		DryRun: dryRun,
	}

	if cache {
		msg := "Cleaning XBPS package cache..."
		if dryRun {
			msg = "[dry-run] Checking XBPS package cache..."
		}
		fmt.Printf("--> %s\n", msg)
		if err := c.CleanCacheWithOptions(allCache, dryRun, yes); err != nil {
			res.Details = append(res.Details, fmt.Sprintf("Cache cleanup failed: %v", err))
		} else {
			res.CacheCleaned = true
			res.Details = append(res.Details, "Cache cleanup completed successfully")
		}
	}

	if orphans {
		msg := "Removing orphaned packages..."
		if dryRun {
			msg = "[dry-run] Checking orphaned packages..."
		}
		fmt.Printf("--> %s\n", msg)
		if err := c.CleanOrphansWithOptions(dryRun, yes); err != nil {
			res.Details = append(res.Details, fmt.Sprintf("Orphan cleanup failed: %v", err))
		} else {
			res.OrphansCleaned = true
			res.Details = append(res.Details, "Orphan cleanup completed successfully")
		}
	}

	if kernels {
		msg := "Purging old kernel versions..."
		if dryRun {
			msg = "[dry-run] Checking old kernel versions..."
		}
		fmt.Printf("--> %s\n", msg)
		if err := c.PurgeOldKernelsWithOptions("all", dryRun, yes); err != nil {
			res.Details = append(res.Details, fmt.Sprintf("Kernel purge skipped/failed: %v", err))
		} else {
			res.KernelsPurged = true
			res.Details = append(res.Details, "Kernel purge completed successfully")
		}
	}

	return res, nil
}
