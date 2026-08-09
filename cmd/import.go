package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/voidlinux/voidpm/pkg/ui"
	"github.com/voidlinux/voidpm/pkg/xbps"
)

var importYes bool

var importCmd = &cobra.Command{
	Use:   "import <package_list_file>",
	Short: "Bulk install packages from plain text file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println(ui.RenderError(fmt.Sprintf("Failed to open package list file '%s': %v", filePath, err)))
			os.Exit(1)
		}
		defer file.Close()

		var pkgs []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				pkgs = append(pkgs, line)
			}
		}

		if len(pkgs) == 0 {
			fmt.Println(ui.RenderWarning("No valid package entries found in file."))
			return
		}

		fmt.Println(ui.RenderInfo(fmt.Sprintf("Importing and installing %d package(s) from %s...", len(pkgs), filePath)))
		client := xbps.NewClient()
		if err := client.ImportPackages(pkgs, importYes); err != nil {
			fmt.Println(ui.RenderError(err.Error()))
			os.Exit(1)
		}

		fmt.Println(ui.RenderSuccess("Package import finished successfully"))
	},
}

func init() {
	importCmd.Flags().BoolVarP(&importYes, "yes", "y", false, "Skip confirmation prompts")

	rootCmd.AddCommand(importCmd)
}
