package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/sys"
	"github.com/voidlinux/voidpm/pkg/ui"
)

var infoCmd = &cobra.Command{
	Use:     "info",
	Aliases: []string{"sysinfo", "version"},
	Short:   "Display Void Linux system overview and voidPM status",
	Run: func(cmd *cobra.Command, args []string) {
		info, err := sys.GetSystemInfo()
		if err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			return
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(info, "", "  ")
			fmt.Println(string(data))
			return
		}

		ui.PrintBanner()
		fmt.Println(ui.RenderHeader("Void Linux System Summary"))
		fmt.Printf("OS:               %s\n", info.VoidVersion)
		fmt.Printf("Kernel:           %s\n", info.Kernel)
		fmt.Printf("Architecture:     %s\n", info.Arch)
		fmt.Printf("Runit Services:   %d total (%d active)\n", info.TotalServices, info.ActiveServices)
		fmt.Printf("XBPS Packages:    %d installed (%d orphans)\n", info.InstalledPkgs, info.OrphanPkgs)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
