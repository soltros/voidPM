package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/runit"
	"github.com/voidlinux/voidpm/pkg/ui"
)

var serviceCmd = &cobra.Command{
	Use:     "service",
	Aliases: []string{"sv", "services"},
	Short:   "Manage runit system services (enable, disable, start, stop, restart, status, logs)",
	Run: func(cmd *cobra.Command, args []string) {
		runServiceStatus(cmd, args)
	},
}

var svStatusCmd = &cobra.Command{
	Use:   "status [service_name]",
	Short: "Show status of all services or specific service",
	Run:   runServiceStatus,
}

var svEnableCmd = &cobra.Command{
	Use:   "enable <service...>",
	Short: "Enable Runit service(s) (symlink /etc/sv/<name> -> /var/service/)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mgr := getSvManager()
		for _, name := range args {
			if err := mgr.EnableService(name); err != nil {
				fmt.Println(ui.RenderError(fmt.Sprintf("Failed to enable '%s': %v", name, err)))
			} else {
				fmt.Println(ui.RenderSuccess(fmt.Sprintf("Enabled and started service '%s'", name)))
			}
		}
	},
}

var svDisableCmd = &cobra.Command{
	Use:   "disable <service...>",
	Short: "Disable Runit service(s) (stop and remove /var/service/ link)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mgr := getSvManager()
		for _, name := range args {
			if err := mgr.DisableService(name); err != nil {
				fmt.Println(ui.RenderError(fmt.Sprintf("Failed to disable '%s': %v", name, err)))
			} else {
				fmt.Println(ui.RenderSuccess(fmt.Sprintf("Disabled service '%s'", name)))
			}
		}
	},
}

var svStartCmd = &cobra.Command{
	Use:   "start <service...>",
	Short: "Start runit service(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mgr := getSvManager()
		for _, name := range args {
			if err := mgr.Start(name); err != nil {
				fmt.Println(ui.RenderError(fmt.Sprintf("Failed to start '%s': %v", name, err)))
			} else {
				fmt.Println(ui.RenderSuccess(fmt.Sprintf("Started service '%s'", name)))
			}
		}
	},
}

var svStopCmd = &cobra.Command{
	Use:   "stop <service...>",
	Short: "Stop runit service(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mgr := getSvManager()
		for _, name := range args {
			if err := mgr.Stop(name); err != nil {
				fmt.Println(ui.RenderError(fmt.Sprintf("Failed to stop '%s': %v", name, err)))
			} else {
				fmt.Println(ui.RenderSuccess(fmt.Sprintf("Stopped service '%s'", name)))
			}
		}
	},
}

var svRestartCmd = &cobra.Command{
	Use:   "restart <service...>",
	Short: "Restart runit service(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mgr := getSvManager()
		for _, name := range args {
			if err := mgr.Restart(name); err != nil {
				fmt.Println(ui.RenderError(fmt.Sprintf("Failed to restart '%s': %v", name, err)))
			} else {
				fmt.Println(ui.RenderSuccess(fmt.Sprintf("Restarted service '%s'", name)))
			}
		}
	},
}

var svReloadCmd = &cobra.Command{
	Use:   "reload <service...>",
	Short: "Reload runit service configuration (HUP signal)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mgr := getSvManager()
		for _, name := range args {
			if err := mgr.Reload(name); err != nil {
				fmt.Println(ui.RenderError(fmt.Sprintf("Failed to reload '%s': %v", name, err)))
			} else {
				fmt.Println(ui.RenderSuccess(fmt.Sprintf("Reloaded service '%s'", name)))
			}
		}
	},
}

var (
	svLogLines  int
	svLogFollow bool
)

var svLogCmd = &cobra.Command{
	Use:     "log <service>",
	Aliases: []string{"logs"},
	Short:   "Tail or stream logs for a runit service or socklog facility (auth, daemon, syslog, etc.)",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mgr := getSvManager()
		serviceName := args[0]

		if svLogFollow {
			fmt.Println(ui.RenderHeader(fmt.Sprintf("Streaming logs for %s (Ctrl+C to exit):", serviceName)))
			if err := mgr.StreamLogs(serviceName, os.Stdout); err != nil {
				fmt.Println(ui.RenderError(err.Error()))
				os.Exit(1)
			}
			return
		}

		lines, err := mgr.TailLogs(serviceName, svLogLines)
		if err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}

		fmt.Println(ui.RenderHeader(fmt.Sprintf("Logs for %s (last %d lines):", serviceName, len(lines))))
		for _, line := range lines {
			fmt.Println(line)
		}
	},
}

func getSvManager() *runit.Manager {
	if userMode {
		uMgr, err := runit.NewUserManager()
		if err == nil {
			return uMgr
		}
	}
	return runit.NewManager()
}

func runServiceStatus(cmd *cobra.Command, args []string) {
	mgr := getSvManager()
	services, err := mgr.ListServices()
	if err != nil {
		fmt.Println(ui.RenderError(err.Error()))
		os.Exit(1)
	}

	// Filter by specific service if requested
	if len(args) > 0 {
		target := args[0]
		var filtered []*runit.Service
		for _, s := range services {
			if strings.EqualFold(s.Name, target) {
				filtered = append(filtered, s)
			}
		}
		services = filtered
	}

	if jsonOutput {
		data, _ := json.MarshalIndent(services, "", "  ")
		fmt.Println(string(data))
		return
	}

	ui.PrintBanner()
	fmt.Println(ui.RenderHeader(fmt.Sprintf("Runit Services (%d total)", len(services))))

	cols := []ui.Column{
		{Title: "SERVICE", Width: 22},
		{Title: "STATE", Width: 12},
		{Title: "ENABLED", Width: 10},
		{Title: "PID", Width: 8},
		{Title: "UPTIME", Width: 12},
	}

	var rows [][]string
	for _, s := range services {
		stateBadge := ""
		switch s.Status {
		case runit.StatusRunning:
			stateBadge = ui.RunningBadge.Render("RUNNING")
		case runit.StatusStopped:
			stateBadge = ui.StoppedBadge.Render("STOPPED")
		case runit.StatusDisabled:
			stateBadge = ui.DisabledBadge.Render("DISABLED")
		default:
			stateBadge = ui.ErrorBadge.Render(string(s.Status))
		}

		enabledText := "no"
		if s.Enabled {
			enabledText = "yes"
		}

		pidText := "-"
		if s.PID > 0 {
			pidText = strconv.Itoa(s.PID)
		}

		rows = append(rows, []string{
			s.Name,
			stateBadge,
			enabledText,
			pidText,
			s.FormattedUptime(),
		})
	}

	fmt.Println(ui.RenderTable(cols, rows))
}

func init() {
	svLogCmd.Flags().IntVarP(&svLogLines, "lines", "n", 50, "Number of log lines to show")
	svLogCmd.Flags().BoolVarP(&svLogFollow, "follow", "f", false, "Stream live log output continuously")

	serviceCmd.AddCommand(svStatusCmd)
	serviceCmd.AddCommand(svEnableCmd)
	serviceCmd.AddCommand(svDisableCmd)
	serviceCmd.AddCommand(svStartCmd)
	serviceCmd.AddCommand(svStopCmd)
	serviceCmd.AddCommand(svRestartCmd)
	serviceCmd.AddCommand(svReloadCmd)
	serviceCmd.AddCommand(svLogCmd)

	rootCmd.AddCommand(serviceCmd)
}
