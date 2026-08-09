package cmd

import (
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
	Short: "Search templates in void-packages srcpkgs/ directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mgr := xbps.NewSrcManager()
		matches, err := mgr.SearchTemplates(args[0])
		if err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderHeader(fmt.Sprintf("Source templates matching '%s' (%d found):", args[0], len(matches))))
		for _, m := range matches {
			fmt.Println("  - " + m)
		}
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
