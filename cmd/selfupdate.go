package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/ui"
	"github.com/voidlinux/voidpm/pkg/update"
)

var (
	selfUpdateTargetRepo   string
	selfUpdateTargetBranch string
	selfUpdateTargetPath   string
)

var selfUpdateCmd = &cobra.Command{
	Use:     "self-update",
	Aliases: []string{"selfupdate", "upgrade-vpm"},
	Short:   "Self-update vpm directly from GitHub repo main branch into /usr/bin/vpm",
	Long: `Downloads the latest pre-compiled binary directly from the main branch of soltros/voidPM
on GitHub, and safely overwrites the target binary at /usr/bin/vpm with elevated privileges if necessary.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.RenderInfo("Downloading latest vpm binary from GitHub repository..."))

		updater := update.NewUpdater(selfUpdateTargetRepo, selfUpdateTargetBranch, selfUpdateTargetPath)
		branchInfo, err := updater.PerformUpdate()
		if err != nil {
			fmt.Println(ui.RenderError(fmt.Sprintf("Self-update failed: %v", err)))
			os.Exit(1)
		}

		fmt.Println(ui.RenderSuccess(fmt.Sprintf("vpm successfully updated from %s and installed at %s", branchInfo, updater.TargetPath)))
	},
}

func init() {
	selfUpdateCmd.Flags().StringVarP(&selfUpdateTargetRepo, "repo", "r", update.DefaultGitHubRepo, "GitHub repository (owner/repo)")
	selfUpdateCmd.Flags().StringVarP(&selfUpdateTargetBranch, "branch", "b", update.DefaultGitHubBranch, "Repository branch")
	selfUpdateCmd.Flags().StringVarP(&selfUpdateTargetPath, "path", "p", update.DefaultTargetPath, "Target destination path for binary overwrite")

	rootCmd.AddCommand(selfUpdateCmd)
}
