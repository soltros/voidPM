package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/ui"
	"github.com/voidlinux/voidpm/pkg/xbps"
)

var checkUpdatesOnly bool

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Quick system health check and pending update lookup",
	Run: func(cmd *cobra.Command, args []string) {
		client := xbps.NewClient()
		updates, err := client.GetPendingUpdatesList()
		if err != nil {
			fmt.Println(ui.RenderError(fmt.Sprintf("Check failed: %v", err)))
			return
		}

		if jsonOutput {
			res := map[string]interface{}{
				"updates_pending": len(updates),
				"packages":        updates,
			}
			data, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(data))
			return
		}

		if checkUpdatesOnly {
			if len(updates) == 0 {
				fmt.Println(ui.RenderSuccess("System is up to date!"))
			} else {
				fmt.Println(ui.RenderInfo(fmt.Sprintf("%d pending update(s) available", len(updates))))
			}
			return
		}

		if len(updates) == 0 {
			fmt.Println(ui.RenderSuccess("System is fully up to date. No pending updates."))
		} else {
			fmt.Println(ui.RenderHeader(fmt.Sprintf("Pending System Updates (%d total):", len(updates))))
			for _, u := range updates {
				fmt.Printf("  - %s\n", u)
			}
			fmt.Println("\nRun 'vpm sync -y' or 'vpm update' to perform upgrade.")
		}
	},
}

func init() {
	checkCmd.Flags().BoolVarP(&checkUpdatesOnly, "updates-only", "U", false, "Quick check returning update count only")

	rootCmd.AddCommand(checkCmd)
}
