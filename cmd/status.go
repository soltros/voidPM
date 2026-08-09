package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/kernel"
	"github.com/voidlinux/voidpm/pkg/runit"
	"github.com/voidlinux/voidpm/pkg/ui"
	"github.com/voidlinux/voidpm/pkg/xbps"
)

type StatusServices struct {
	Active int `json:"active"`
	Failed int `json:"failed"`
}

type SystemStatus struct {
	UpdatesPending   int            `json:"updates_pending"`
	OrphansCount     int            `json:"orphans_count"`
	RunningKernel    string         `json:"running_kernel"`
	InstalledKernels []string       `json:"installed_kernels"`
	Services         StatusServices `json:"services"`
	HoldPackages     []string       `json:"hold_packages"`
}

var updatesOnly bool

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"stat", "health"},
	Short:   "Provide system health, pending update counts, and status bar metrics",
	Run: func(cmd *cobra.Command, args []string) {
		runStatus(cmd, args)
	},
}

func runStatus(cmd *cobra.Command, args []string) {
	client := xbps.NewClient()
	svMgr := getSvManager()

	updates, _ := client.GetPendingUpdatesList()
	if updatesOnly {
		if jsonOutput {
			data, _ := json.MarshalIndent(map[string]interface{}{
				"updates_pending": len(updates),
				"packages":        updates,
			}, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("%d pending updates\n", len(updates))
		}
		return
	}

	orphans, _ := client.ListOrphans()
	holds, _ := client.ListHolds()
	if holds == nil {
		holds = []string{}
	}

	sk, _ := kernel.GetSystemKernels()
	runningKernel := ""
	var installedKernels []string
	if sk != nil {
		runningKernel = sk.RunningKernel
		for _, k := range sk.Installed {
			installedKernels = append(installedKernels, k.Version)
		}
	}
	if installedKernels == nil {
		installedKernels = []string{}
	}

	services, _ := svMgr.ListServices()
	activeServices := 0
	failedServices := 0
	for _, s := range services {
		if s.Enabled || s.Status == runit.StatusRunning {
			activeServices++
		}
		if s.Status == runit.StatusFailed {
			failedServices++
		}
	}

	st := SystemStatus{
		UpdatesPending:   len(updates),
		OrphansCount:     len(orphans),
		RunningKernel:    runningKernel,
		InstalledKernels: installedKernels,
		Services: StatusServices{
			Active: activeServices,
			Failed: failedServices,
		},
		HoldPackages: holds,
	}

	if jsonOutput {
		data, _ := json.MarshalIndent(st, "", "  ")
		fmt.Println(string(data))
		return
	}

	ui.PrintBanner()
	fmt.Println(ui.RenderHeader("VoidPM System Status & Health Summary"))
	fmt.Printf("Pending Updates:    %d\n", st.UpdatesPending)
	fmt.Printf("Orphaned Packages:  %d\n", st.OrphansCount)
	fmt.Printf("Running Kernel:     %s\n", st.RunningKernel)
	fmt.Printf("Installed Kernels:  %v\n", st.InstalledKernels)
	fmt.Printf("Active Services:    %d\n", st.Services.Active)
	fmt.Printf("Failed Services:    %d\n", st.Services.Failed)
	fmt.Printf("Hold Packages:      %v\n", st.HoldPackages)
}

func init() {
	statusCmd.Flags().BoolVarP(&updatesOnly, "updates-only", "U", false, "Show only pending update statistics")

	rootCmd.AddCommand(statusCmd)
}
