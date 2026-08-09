package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/ui"
	"github.com/voidlinux/voidpm/pkg/xbps"
)

var (
	cleanOrphans bool
	cleanCache   bool
	cleanKernels bool
	cleanAll     bool
	cleanDryRun  bool
	cleanYes     bool
)

var cleanCmd = &cobra.Command{
	Use:     "clean",
	Aliases: []string{"cleanup", "purge"},
	Short:   "System maintenance (cache clearing, orphan removal, kernel purges)",
	Run: func(cmd *cobra.Command, args []string) {
		runClean(cmd, args)
	},
}

var cleanAllCache bool

var cleanCacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Remove obsolete XBPS package tarballs from cache (xbps-remove -O)",
	Run: func(cmd *cobra.Command, args []string) {
		cleaner := xbps.NewCleaner()
		res, err := cleaner.PerformCleanup(false, true, false, cleanAllCache, cleanDryRun, cleanYes)
		if err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		if jsonOutput {
			data, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(data))
			return
		}
		fmt.Println(ui.RenderSuccess("Cache cleanup completed"))
	},
}

var cleanOrphansCmd = &cobra.Command{
	Use:   "orphans",
	Short: "Remove unneeded orphan packages (xbps-remove -o)",
	Run: func(cmd *cobra.Command, args []string) {
		cleaner := xbps.NewCleaner()
		res, err := cleaner.PerformCleanup(true, false, false, false, cleanDryRun, cleanYes)
		if err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		if jsonOutput {
			data, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(data))
			return
		}
		fmt.Println(ui.RenderSuccess("Orphan cleanup completed"))
	},
}

var cleanKernelsCmd = &cobra.Command{
	Use:   "kernels",
	Short: "Purge old Linux kernel versions via vkpurge",
	Run: func(cmd *cobra.Command, args []string) {
		cleaner := xbps.NewCleaner()
		res, err := cleaner.PerformCleanup(false, false, true, false, cleanDryRun, cleanYes)
		if err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		if jsonOutput {
			data, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(data))
			return
		}
		fmt.Println(ui.RenderSuccess("Kernel purge completed"))
	},
}

var cleanAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Run full system cleanup (cache + orphans + kernels)",
	Run: func(cmd *cobra.Command, args []string) {
		cleanAll = true
		runClean(cmd, args)
	},
}

func runClean(cmd *cobra.Command, args []string) {
	cleaner := xbps.NewCleaner()

	doOrphans := cleanOrphans
	doCache := cleanCache
	doKernels := cleanKernels

	if cleanAll || (!doOrphans && !doCache && !doKernels) {
		doOrphans = true
		doCache = true
		doKernels = true
	}

	if !jsonOutput {
		if cleanDryRun {
			fmt.Println(ui.RenderHeader("Starting System Cleanup (DRY-RUN)..."))
		} else {
			fmt.Println(ui.RenderHeader("Starting Void Linux System Cleanup..."))
		}
	}

	res, err := cleaner.PerformCleanup(doOrphans, doCache, doKernels, cleanAllCache, cleanDryRun, cleanYes)
	if err != nil {
		fmt.Println(ui.RenderError(err.Error()))
		os.Exit(1)
	}

	if jsonOutput {
		data, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(data))
		return
	}

	for _, d := range res.Details {
		fmt.Println(ui.RenderInfo(d))
	}
	fmt.Println(ui.RenderSuccess("System cleanup finished successfully"))
}

func init() {
	cleanCmd.Flags().BoolVarP(&cleanOrphans, "orphans", "o", false, "Remove orphaned packages (xbps-remove -o)")
	cleanCmd.Flags().BoolVarP(&cleanCache, "cache", "c", false, "Remove outdated packages from XBPS cache (xbps-remove -O)")
	cleanCmd.Flags().BoolVarP(&cleanKernels, "kernels", "k", false, "Remove old/unused kernels via vkpurge")
	cleanCmd.Flags().BoolVarP(&cleanAll, "all", "a", false, "Enable all cleanup modes")
	cleanCmd.Flags().BoolVarP(&cleanDryRun, "dry-run", "n", false, "Preview actions without modifying system files")
	cleanCmd.Flags().BoolVarP(&cleanYes, "yes", "y", false, "Skip confirmation prompts")

	cleanCacheCmd.Flags().BoolVarP(&cleanAllCache, "all", "a", false, "Remove ALL cached packages, not just obsolete ones")
	cleanCacheCmd.Flags().BoolVarP(&cleanDryRun, "dry-run", "n", false, "Preview cache cleanup")
	cleanCacheCmd.Flags().BoolVarP(&cleanYes, "yes", "y", false, "Skip confirmation prompts")

	cleanOrphansCmd.Flags().BoolVarP(&cleanDryRun, "dry-run", "n", false, "Preview orphan removal")
	cleanOrphansCmd.Flags().BoolVarP(&cleanYes, "yes", "y", false, "Skip confirmation prompts")

	cleanKernelsCmd.Flags().BoolVarP(&cleanDryRun, "dry-run", "n", false, "Preview kernel purge")
	cleanKernelsCmd.Flags().BoolVarP(&cleanYes, "yes", "y", false, "Skip confirmation prompts")

	cleanAllCmd.Flags().BoolVarP(&cleanDryRun, "dry-run", "n", false, "Preview full cleanup")
	cleanAllCmd.Flags().BoolVarP(&cleanYes, "yes", "y", false, "Skip confirmation prompts")

	cleanCmd.AddCommand(cleanCacheCmd)
	cleanCmd.AddCommand(cleanOrphansCmd)
	cleanCmd.AddCommand(cleanKernelsCmd)
	cleanCmd.AddCommand(cleanAllCmd)

	rootCmd.AddCommand(cleanCmd)
}
