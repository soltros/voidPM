package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/ui"
	"github.com/voidlinux/voidpm/pkg/xbps"
)

var syncYes bool

var syncCmd = &cobra.Command{
	Use:     "sync",
	Aliases: []string{"upgrade", "up"},
	Short:   "Synchronize repositories and perform non-interactive system update",
	Run: func(cmd *cobra.Command, args []string) {
		client := xbps.NewClient()
		fmt.Println(ui.RenderInfo("Synchronizing repositories & running system upgrade..."))

		if err := client.UpdateSystemWithOptions(syncYes); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}

		fmt.Println(ui.RenderSuccess("System update finished successfully"))
	},
}

func init() {
	syncCmd.Flags().BoolVarP(&syncYes, "yes", "y", false, "Skip confirmation prompts")

	rootCmd.AddCommand(syncCmd)
}
