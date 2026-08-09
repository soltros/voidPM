package xbps

import (
	"fmt"
	"os/exec"

	"github.com/voidlinux/voidpm/pkg/sys"
)

type Cleaner struct{}

func NewCleaner() *Cleaner {
	return &Cleaner{}
}

// CleanCache removes obsolete package files from XBPS cache (xbps-remove -O / xbps-remove -c)
func (c *Cleaner) CleanCache(all bool) error {
	if all {
		return sys.RunElevated("xbps-remove", "-c")
	}
	return sys.RunElevated("xbps-remove", "-O")
}

// CleanOrphans removes orphaned packages (xbps-remove -o)
func (c *Cleaner) CleanOrphans() error {
	return sys.RunElevated("xbps-remove", "-o")
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
	return string(out), nil
}

// PurgeOldKernels removes old Linux kernels using vkpurge rm all
func (c *Cleaner) PurgeOldKernels() error {
	if _, err := exec.LookPath("vkpurge"); err != nil {
		return fmt.Errorf("vkpurge utility is not installed")
	}

	return sys.RunElevated("vkpurge", "rm", "all")
}

// PerformAllCleanup cleans cache, removes orphans, and purges old kernels
func (c *Cleaner) PerformAllCleanup() error {
	fmt.Println("--> Cleaning XBPS package cache...")
	if err := c.CleanCache(false); err != nil {
		fmt.Printf("Warning: cache cleanup failed: %v\n", err)
	}

	fmt.Println("--> Removing orphaned packages...")
	if err := c.CleanOrphans(); err != nil {
		fmt.Printf("Warning: orphan cleanup failed: %v\n", err)
	}

	fmt.Println("--> Purging old kernel versions...")
	if err := c.PurgeOldKernels(); err != nil {
		fmt.Printf("Note: kernel purge skipped or not available (%v)\n", err)
	}

	return nil
}
