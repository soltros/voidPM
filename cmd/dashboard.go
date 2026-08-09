package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/tui"
)

var dashboardCmd = &cobra.Command{
	Use:     "dashboard",
	Aliases: []string{"tui", "ui"},
	Short:   "Launch interactive TUI dashboard for visual service & package management",
	RunE:    runDashboardCmd,
}

func runDashboardCmd(cmd *cobra.Command, args []string) error {
	m, err := tui.NewModel()
	if err != nil {
		return fmt.Errorf("failed to initialize TUI: %w", err)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running dashboard: %v\n", err)
		os.Exit(1)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
