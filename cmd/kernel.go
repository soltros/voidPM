package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/kernel"
	"github.com/voidlinux/voidpm/pkg/ui"
	"github.com/voidlinux/voidpm/pkg/xbps"
)

var kernelCmd = &cobra.Command{
	Use:     "kernel",
	Aliases: []string{"k", "kernels"},
	Short:   "Void Linux Kernel management (status, switch, reconfigure, dracut, purge, hold)",
	Run: func(cmd *cobra.Command, args []string) {
		runKernelStatus(cmd, args)
	},
}

var kStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show active running kernel, installed kernels, and purgeable old versions",
	Run:   runKernelStatus,
}

var kReconfigureCmd = &cobra.Command{
	Use:     "reconfigure [kernel_package]",
	Aliases: []string{"rec", "initramfs"},
	Short:   "Reconfigure kernel initramfs & bootloader hooks (xbps-reconfigure -f linux)",
	Run: func(cmd *cobra.Command, args []string) {
		target := ""
		if len(args) > 0 {
			target = args[0]
		}
		if err := kernel.Reconfigure(target); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess("Kernel reconfigured successfully"))
	},
}

var kDracutCmd = &cobra.Command{
	Use:   "dracut",
	Short: "Force regenerate all initramfs images via dracut (--regenerate-all --force)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := kernel.RegenerateInitramfs(); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess("All initramfs images regenerated successfully"))
	},
}

var kPurgeCmd = &cobra.Command{
	Use:   "purge [all|version]",
	Short: "Remove obsolete old kernel versions via vkpurge",
	Run: func(cmd *cobra.Command, args []string) {
		target := "all"
		if len(args) > 0 {
			target = args[0]
		}
		if err := kernel.Purge(target); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess("Kernel purge completed successfully"))
	},
}

var kSwitchCmd = &cobra.Command{
	Use:   "switch <flavor>",
	Short: "Switch primary kernel series (e.g. linux-lts, linux6.6, linux6.12)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		flavor := args[0]
		if err := kernel.SwitchFlavor(flavor); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Switched kernel flavor to '%s'", flavor)))
	},
}

var kHoldCmd = &cobra.Command{
	Use:   "hold [kernel_package]",
	Short: "Hold kernel package to prevent automated kernel major version updates",
	Run: func(cmd *cobra.Command, args []string) {
		target := "linux"
		if len(args) > 0 {
			target = args[0]
		}
		client := xbps.NewClient()
		if err := client.Hold(target); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Kernel package '%s' put on hold", target)))
	},
}

var kUnholdCmd = &cobra.Command{
	Use:   "unhold [kernel_package]",
	Short: "Remove hold on kernel package",
	Run: func(cmd *cobra.Command, args []string) {
		target := "linux"
		if len(args) > 0 {
			target = args[0]
		}
		client := xbps.NewClient()
		if err := client.Unhold(target); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Kernel package '%s' hold removed", target)))
	},
}

func runKernelStatus(cmd *cobra.Command, args []string) {
	sk, err := kernel.GetSystemKernels()
	if err != nil {
		fmt.Println(ui.RenderError(err.Error()))
		os.Exit(1)
	}

	if jsonOutput {
		data, _ := json.MarshalIndent(sk, "", "  ")
		fmt.Println(string(data))
		return
	}

	ui.PrintBanner()
	fmt.Println(ui.RenderHeader("Void Linux Kernel Management Summary"))
	fmt.Printf("Running Kernel:     %s\n\n", ui.InstalledBadge.Render(sk.RunningKernel))

	cols := []ui.Column{
		{Title: "PACKAGE", Width: 20},
		{Title: "VERSION", Width: 18},
		{Title: "STATUS", Width: 14},
		{Title: "METAPACKAGE", Width: 12},
	}

	var rows [][]string
	for _, k := range sk.Installed {
		st := "[installed]"
		if k.IsRunning {
			st = ui.RunningBadge.Render("RUNNING")
		} else if k.IsPurgeable {
			st = ui.OrphanBadge.Render("PURGEABLE")
		}

		isMetaText := "no"
		if k.IsMeta {
			isMetaText = "yes"
		}

		rows = append(rows, []string{
			k.Name,
			k.Version,
			st,
			isMetaText,
		})
	}

	fmt.Println(ui.RenderTable(cols, rows))

	if len(sk.OldPurgeable) > 0 {
		fmt.Printf("\nFound %d old purgeable kernel version(s): %v\n", len(sk.OldPurgeable), sk.OldPurgeable)
		fmt.Println("Run 'vpm kernel purge' to safely remove old kernel files.")
	}
}

var kAvailableCmd = &cobra.Command{
	Use:     "available",
	Aliases: []string{"search", "list-all"},
	Short:   "List all available Linux kernel series in Void Linux repositories",
	Run: func(cmd *cobra.Command, args []string) {
		kernels, err := kernel.ListAvailableKernels()
		if err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(kernels, "", "  ")
			fmt.Println(string(data))
			return
		}

		ui.PrintBanner()
		fmt.Println(ui.RenderHeader("Available Kernel Series in Void Repositories"))

		cols := []ui.Column{
			{Title: "PACKAGE NAME", Width: 22},
			{Title: "LATEST VERSION", Width: 18},
			{Title: "TYPE", Width: 16},
		}

		var rows [][]string
		for _, k := range kernels {
			typeText := "Kernel Series"
			if k.IsMeta {
				typeText = "Metapackage"
			}
			rows = append(rows, []string{
				k.Name,
				k.Version,
				typeText,
			})
		}

		fmt.Println(ui.RenderTable(cols, rows))
		fmt.Println("\nTo switch to a kernel series: vpm kernel switch <package_name>")
	},
}

var kRemoveCmd = &cobra.Command{
	Use:     "remove <kernel_package>",
	Aliases: []string{"rm", "delete", "uninstall"},
	Short:   "Safely uninstall a specified kernel package/series and update bootloader",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pkg := args[0]
		if err := kernel.RemoveKernel(pkg); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Kernel package '%s' removed successfully", pkg)))
	},
}

func init() {
	kernelCmd.AddCommand(kStatusCmd)
	kernelCmd.AddCommand(kAvailableCmd)
	kernelCmd.AddCommand(kReconfigureCmd)
	kernelCmd.AddCommand(kDracutCmd)
	kernelCmd.AddCommand(kPurgeCmd)
	kernelCmd.AddCommand(kSwitchCmd)
	kernelCmd.AddCommand(kRemoveCmd)
	kernelCmd.AddCommand(kHoldCmd)
	kernelCmd.AddCommand(kUnholdCmd)

	rootCmd.AddCommand(kernelCmd)
}
