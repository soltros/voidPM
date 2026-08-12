package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/ui"
	"github.com/voidlinux/voidpm/pkg/update"
)

var (
	selfUpdateTargetRepo string
	selfUpdateTargetPath string
)

var selfUpdateCmd = &cobra.Command{
	Use:     "self-update",
	Aliases: []string{"selfupdate", "upgrade-vpm"},
	Short:   "Self-update vpm from latest GitHub release assets directly into /usr/bin/vpm",
	Long: `Queries the GitHub API for the latest soltros/voidPM release, downloads the latest binary,
and safely overwrites the target binary at /usr/bin/vpm with elevated privileges if necessary.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.RenderInfo("Checking GitHub API for latest vpm release..."))

		updater := update.NewUpdater(selfUpdateTargetRepo, selfUpdateTargetPath)
		tag, err := updater.PerformUpdate()
		if err != nil {
			fmt.Println(ui.RenderError(fmt.Sprintf("Self-update failed: %v", err)))
			os.Exit(1)
		}

		fmt.Println(ui.RenderSuccess(fmt.Sprintf("vpm successfully updated to %s and installed at %s", tag, updater.TargetPath)))
	},
}

func init() {
	selfUpdateCmd.Flags().StringVarP(&selfUpdateTargetRepo, "repo", "r", update.DefaultGitHubRepo, "GitHub repository (owner/repo)")
	selfUpdateCmd.Flags().StringVarP(&selfUpdateTargetPath, "path", "p", update.DefaultTargetPath, "Target destination path for binary overwrite")

	rootCmd.AddCommand(selfUpdateCmd)
}
