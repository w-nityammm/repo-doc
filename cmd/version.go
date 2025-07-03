package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long: `Display the current version of repo-doc CLI tool.

This command shows the version number and build information.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("repo-doc v1.0.0")
		fmt.Println("A GitHub repository analyzer CLI tool")
	},
	Example: `  repo-doc version`,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
