package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/ui"
	"github.com/voidlinux/voidpm/pkg/xbps"
)

var (
	syncYes  bool
	syncSelf bool
)

var syncCmd = &cobra.Command{
	Use:     "sync",
	Aliases: []string{"upgrade", "up"},
	Short:   "Synchronize repositories and perform non-interactive system update",
	Run: func(cmd *cobra.Command, args []string) {
		if syncSelf {
			fmt.Println(ui.RenderInfo("Self-updating vpm binary..."))
			selfUpdateCmd.Run(cmd, args)
		}

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
	syncCmd.Flags().BoolVarP(&syncSelf, "self", "s", false, "Self-update vpm binary from GitHub releases before system update")

	rootCmd.AddCommand(syncCmd)
}
