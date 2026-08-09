package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/ui"
	"github.com/voidlinux/voidpm/pkg/xbps"
)

var cleanCmd = &cobra.Command{
	Use:     "clean",
	Aliases: []string{"cleanup", "purge"},
	Short:   "System maintenance (cache clearing, orphan removal, kernel purges)",
	Run: func(cmd *cobra.Command, args []string) {
		runCleanAll(cmd, args)
	},
}

var cleanAllCache bool

var cleanCacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Remove obsolete XBPS package tarballs from cache (xbps-remove -O)",
	Run: func(cmd *cobra.Command, args []string) {
		cleaner := xbps.NewCleaner()
		if err := cleaner.CleanCache(cleanAllCache); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess("Cache cleanup completed"))
	},
}

var cleanOrphansCmd = &cobra.Command{
	Use:   "orphans",
	Short: "Remove unneeded orphan packages (xbps-remove -o)",
	Run: func(cmd *cobra.Command, args []string) {
		cleaner := xbps.NewCleaner()
		if err := cleaner.CleanOrphans(); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess("Orphan cleanup completed"))
	},
}

var cleanKernelsCmd = &cobra.Command{
	Use:   "kernels",
	Short: "Purge old Linux kernel versions via vkpurge",
	Run: func(cmd *cobra.Command, args []string) {
		cleaner := xbps.NewCleaner()
		list, err := cleaner.ListOldKernels()
		if err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}

		if list == "" {
			fmt.Println(ui.RenderSuccess("No obsolete kernels to remove"))
			return
		}

		fmt.Println(ui.RenderHeader("Found old kernels to purge:"))
		fmt.Println(list)

		if err := cleaner.PurgeOldKernels(); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess("Old kernels purged successfully"))
	},
}

var cleanAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Run full system cleanup (cache + orphans + kernels)",
	Run:   runCleanAll,
}

func runCleanAll(cmd *cobra.Command, args []string) {
	cleaner := xbps.NewCleaner()
	fmt.Println(ui.RenderHeader("Starting Complete Void Linux System Cleanup..."))
	if err := cleaner.PerformAllCleanup(); err != nil {
		fmt.Println(ui.RenderError(err.Error()))
		os.Exit(1)
	}
	fmt.Println(ui.RenderSuccess("System cleanup finished successfully"))
}

func init() {
	cleanCacheCmd.Flags().BoolVarP(&cleanAllCache, "all", "a", false, "Remove ALL cached packages, not just obsolete ones")

	cleanCmd.AddCommand(cleanCacheCmd)
	cleanCmd.AddCommand(cleanOrphansCmd)
	cleanCmd.AddCommand(cleanKernelsCmd)
	cleanCmd.AddCommand(cleanAllCmd)

	rootCmd.AddCommand(cleanCmd)
}
