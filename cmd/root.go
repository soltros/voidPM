package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/ui"
)

var (
	jsonOutput bool
	userMode   bool
)

var rootCmd = &cobra.Command{
	Use:   "vpm",
	Short: "voidPM - Elegant Void Linux System & Package Manager Overlay",
	Long: `voidPM (vpm) is a powerful, elegant helper tool for Void Linux.
It unifies Runit service management, XBPS package management, void-packages (xbps-src) 
building, and system cleanup into a fast, intuitive interface.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Launch interactive dashboard if run without subcommands
		if err := runDashboardCmd(cmd, args); err != nil {
			fmt.Printf("Dashboard error: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(ui.RenderError(err.Error()))
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Output results in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&userMode, "user", "u", false, "Operate on user-level services/packages")
}
