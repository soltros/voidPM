package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/ui"
	"github.com/voidlinux/voidpm/pkg/xbps"
)

var srcCmd = &cobra.Command{
	Use:     "src",
	Aliases: []string{"source", "xbps-src"},
	Short:   "Void source packages (xbps-src) build overlay",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var srcSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Clone void-packages repo into ~/void-packages and run binary-bootstrap",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := xbps.NewSrcManager()
		if err := mgr.Setup(); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess("void-packages repository set up successfully!"))
	},
}

var srcSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Update local void-packages git repository (git pull --rebase)",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := xbps.NewSrcManager()
		if err := mgr.Sync(); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess("void-packages repository updated successfully!"))
	},
}

var allowRestricted bool

var srcBuildCmd = &cobra.Command{
	Use:   "build <package>",
	Short: "Build package from source using ./xbps-src pkg <package>",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mgr := xbps.NewSrcManager()
		if err := mgr.Build(args[0], allowRestricted); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Package '%s' built successfully!", args[0])))
	},
}

var srcInstallCmd = &cobra.Command{
	Use:   "install <package>",
	Short: "Install built package from hostdir/binpkgs",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mgr := xbps.NewSrcManager()
		if err := mgr.InstallBuilt(args[0]); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Built package '%s' installed successfully!", args[0])))
	},
}

var srcSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search templates in void-packages repository with metadata & restricted status",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mgr := xbps.NewSrcManager()
		matches, err := mgr.SearchTemplates(args[0])
		if err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(matches, "", "  ")
			fmt.Println(string(data))
			return
		}

		ui.PrintBanner()
		fmt.Println(ui.RenderHeader(fmt.Sprintf("Source templates matching '%s' (%d found in %s):", args[0], len(matches), mgr.RepoDir)))

		cols := []ui.Column{
			{Title: "STATUS", Width: 8},
			{Title: "RESTRICTED", Width: 12},
			{Title: "PACKAGE NAME", Width: 24},
			{Title: "DESCRIPTION", Width: 45},
		}

		var rows [][]string
		for _, m := range matches {
			st := "[ ]"
			if m.IsInstalled {
				st = ui.InstalledBadge.Render("[*]")
			}

			restr := "-"
			if m.IsRestricted {
				restr = ui.OrphanBadge.Render("[RESTRICTED]")
			}

			rows = append(rows, []string{
				st,
				restr,
				m.Name,
				m.ShortDesc,
			})
		}

		fmt.Println(ui.RenderTable(cols, rows))
	},
}

var srcAllowRestrictedCmd = &cobra.Command{
	Use:     "allow-restricted",
	Aliases: []string{"enable-restricted", "restricted"},
	Short:   "Enable XBPS_ALLOW_RESTRICTED=yes in void-packages/etc/conf for proprietary software",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := xbps.NewSrcManager()
		if err := mgr.EnableRestrictedInConfig(); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess("XBPS_ALLOW_RESTRICTED=yes enabled in void-packages/etc/conf!"))
	},
}

func init() {
	srcBuildCmd.Flags().BoolVarP(&allowRestricted, "restricted", "m", false, "Allow restricted packages (XBPS_ALLOW_RESTRICTED=yes)")

	srcCmd.AddCommand(srcSetupCmd)
	srcCmd.AddCommand(srcSyncCmd)
	srcCmd.AddCommand(srcBuildCmd)
	srcCmd.AddCommand(srcInstallCmd)
	srcCmd.AddCommand(srcSearchCmd)
	srcCmd.AddCommand(srcAllowRestrictedCmd)

	rootCmd.AddCommand(srcCmd)
}
