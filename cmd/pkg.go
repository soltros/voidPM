package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/ui"
	"github.com/voidlinux/voidpm/pkg/xbps"
)

var pkgCmd = &cobra.Command{
	Use:     "pkg",
	Aliases: []string{"package", "packages"},
	Short:   "Unified XBPS package management overlay (install, remove, search, update, info, orphans)",
	Run: func(cmd *cobra.Command, args []string) {
		runPkgSearch(cmd, args)
	},
}

var pkgSearchCmd = &cobra.Command{
	Use:     "search <query>",
	Aliases: []string{"find"},
	Short:   "Search remote & local XBPS repositories for packages",
	Args:    cobra.MinimumNArgs(1),
	Run:     runPkgSearch,
}

var syncRepo bool

var pkgInstallCmd = &cobra.Command{
	Use:     "install <package...>",
	Aliases: []string{"add", "i"},
	Short:   "Install or update package(s)",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := xbps.NewClient()
		fmt.Println(ui.RenderInfo(fmt.Sprintf("Installing package(s): %v", args)))
		if err := client.Install(args, syncRepo); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess("Installation finished successfully"))
	},
}

var recursiveRemove bool

var pkgRemoveCmd = &cobra.Command{
	Use:     "remove <package...>",
	Aliases: []string{"rm", "uninstall"},
	Short:   "Remove installed package(s)",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := xbps.NewClient()
		fmt.Println(ui.RenderInfo(fmt.Sprintf("Removing package(s): %v", args)))
		if err := client.Remove(args, recursiveRemove); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess("Removal finished successfully"))
	},
}

var pkgUpdateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"upgrade", "up"},
	Short:   "Synchronize repositories and perform full system update (xbps-install -Su)",
	Run: func(cmd *cobra.Command, args []string) {
		client := xbps.NewClient()
		fmt.Println(ui.RenderInfo("Synchronizing repos & running system update..."))
		if err := client.UpdateSystem(); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess("System update complete"))
	},
}

var pkgInfoCmd = &cobra.Command{
	Use:     "info <package>",
	Aliases: []string{"show"},
	Short:   "Show detailed metadata for a package",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := xbps.NewClient()
		pkg, err := client.GetInfo(args[0])
		if err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(pkg, "", "  ")
			fmt.Println(string(data))
			return
		}

		fmt.Println(ui.RenderHeader(fmt.Sprintf("Package Info: %s (%s)", pkg.Name, pkg.Version)))
		fmt.Printf("Summary:       %s\n", pkg.ShortDesc)
		fmt.Printf("State:         %s\n", func() string {
			if pkg.Installed {
				return ui.InstalledBadge.Render("Installed")
			}
			return "Available"
		}())
		fmt.Printf("Installed Size: %s\n", pkg.InstalledSize)
		fmt.Printf("Download Size:  %s\n", pkg.DownloadSize)
		fmt.Printf("Repository:    %s\n", pkg.Repository)
		fmt.Printf("Architecture:  %s\n", pkg.Architecture)
		fmt.Printf("Maintainer:    %s\n", pkg.Maintainer)
		fmt.Printf("Homepage:      %s\n", pkg.Homepage)
		fmt.Printf("License:       %s\n", pkg.License)
		if len(pkg.Dependencies) > 0 {
			fmt.Printf("Dependencies:  %v\n", pkg.Dependencies)
		}
	},
}

var pkgOrphansCmd = &cobra.Command{
	Use:   "orphans",
	Short: "List orphaned packages (installed as dependencies but no longer needed)",
	Run: func(cmd *cobra.Command, args []string) {
		client := xbps.NewClient()
		orphans, err := client.ListOrphans()
		if err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}

		if len(orphans) == 0 {
			fmt.Println(ui.RenderSuccess("No orphaned packages found on system!"))
			return
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(orphans, "", "  ")
			fmt.Println(string(data))
			return
		}

		fmt.Println(ui.RenderHeader(fmt.Sprintf("Orphaned Packages (%d total)", len(orphans))))
		for _, p := range orphans {
			fmt.Printf("  - %s (%s)\n", ui.OrphanBadge.Render(p.Name), p.Version)
		}
		fmt.Println("\nRun 'vpm clean orphans' to safely remove them.")
	},
}

var pkgWhoownsCmd = &cobra.Command{
	Use:   "whoowns <file_path>",
	Short: "Find which package owns a specified file path",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := xbps.NewClient()
		owner, err := client.WhoOwns(args[0])
		if err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("'%s' is owned by: %s", args[0], owner)))
	},
}

var pkgFilesCmd = &cobra.Command{
	Use:   "files <package>",
	Short: "List all files provided by an installed package",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := xbps.NewClient()
		files, err := client.GetFiles(args[0])
		if err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderHeader(fmt.Sprintf("Files installed by %s (%d files):", args[0], len(files))))
		for _, f := range files {
			fmt.Println("  " + f)
		}
	},
}

var pkgHoldCmd = &cobra.Command{
	Use:   "hold <package>",
	Short: "Put package on hold to prevent automatic updates",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := xbps.NewClient()
		if err := client.Hold(args[0]); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Package '%s' put on hold", args[0])))
	},
}

var pkgUnholdCmd = &cobra.Command{
	Use:   "unhold <package>",
	Short: "Remove hold on package",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := xbps.NewClient()
		if err := client.Unhold(args[0]); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Package '%s' hold removed", args[0])))
	},
}

func runPkgSearch(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Help()
		return
	}
	query := args[0]
	client := xbps.NewClient()
	pkgs, err := client.Search(query)
	if err != nil {
		fmt.Println(ui.RenderError(err.Error()))
		os.Exit(1)
	}

	if jsonOutput {
		data, _ := json.MarshalIndent(pkgs, "", "  ")
		fmt.Println(string(data))
		return
	}

	ui.PrintBanner()
	fmt.Println(ui.RenderHeader(fmt.Sprintf("Search results for '%s' (%d packages found):", query, len(pkgs))))

	cols := []ui.Column{
		{Title: "STATUS", Width: 8},
		{Title: "PACKAGE", Width: 26},
		{Title: "VERSION", Width: 16},
		{Title: "DESCRIPTION", Width: 45},
	}

	var rows [][]string
	for _, p := range pkgs {
		st := "[ ]"
		if p.Installed {
			st = ui.InstalledBadge.Render("[*]")
		}
		rows = append(rows, []string{
			st,
			p.Name,
			p.Version,
			p.ShortDesc,
		})
	}

	fmt.Println(ui.RenderTable(cols, rows))
}

func init() {
	pkgInstallCmd.Flags().BoolVarP(&syncRepo, "sync", "S", true, "Synchronize repositories before installing")
	pkgRemoveCmd.Flags().BoolVarP(&recursiveRemove, "recursive", "R", false, "Recursively remove unneeded dependencies")

	pkgCmd.AddCommand(pkgSearchCmd)
	pkgCmd.AddCommand(pkgInstallCmd)
	pkgCmd.AddCommand(pkgRemoveCmd)
	pkgCmd.AddCommand(pkgUpdateCmd)
	pkgCmd.AddCommand(pkgInfoCmd)
	pkgCmd.AddCommand(pkgOrphansCmd)
	pkgCmd.AddCommand(pkgWhoownsCmd)
	pkgCmd.AddCommand(pkgFilesCmd)
	pkgCmd.AddCommand(pkgHoldCmd)
	pkgCmd.AddCommand(pkgUnholdCmd)

	rootCmd.AddCommand(pkgCmd)

	// Top-level aliases for ultra convenience:
	// 'vpm install', 'vpm remove', 'vpm search', 'vpm update'
	rootCmd.AddCommand(&cobra.Command{
		Use:    "install <package...>",
		Hidden: false,
		Short:  "Alias for 'vpm pkg install'",
		Run:    pkgInstallCmd.Run,
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:    "remove <package...>",
		Hidden: false,
		Short:  "Alias for 'vpm pkg remove'",
		Run:    pkgRemoveCmd.Run,
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:    "search <query>",
		Hidden: false,
		Short:  "Alias for 'vpm pkg search'",
		Run:    pkgSearchCmd.Run,
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:    "update",
		Hidden: false,
		Short:  "Alias for 'vpm pkg update'",
		Run:    pkgUpdateCmd.Run,
	})
}
