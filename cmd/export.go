package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/ui"
	"github.com/voidlinux/voidpm/pkg/xbps"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export explicitly installed packages to plain text list",
	Run: func(cmd *cobra.Command, args []string) {
		client := xbps.NewClient()
		pkgs, err := client.ListExplicit()
		if err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}

		for _, p := range pkgs {
			fmt.Println(p)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
