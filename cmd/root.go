package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const asciiArt = `
██████╗ ███████╗██████╗  ██████╗       ██████╗  ██████╗  ██████╗
██╔══██╗██╔════╝██╔══██╗██╔═══██╗      ██╔══██╗██╔═══██╗██╔════╝
██████╔╝█████╗  ██████╔╝██║   ██║█████╗██║  ██║██║   ██║██║     
██╔══██╗██╔══╝  ██╔═══╝ ██║   ██║╚════╝██║  ██║██║   ██║██║     
██║  ██║███████╗██║     ╚██████╔╝      ██████╔╝╚██████╔╝╚██████╗
╚═╝  ╚═╝╚══════╝╚═╝      ╚═════╝       ╚═════╝  ╚═════╝  ╚═════╝
`

var rootCmd = &cobra.Command{
	Use:   "repo-doc",
	Short: "A comprehensive CLI tool for GitHub repository analysis and insights",
	Long: asciiArt + `
Advanced GitHub repo insights, right from your terminal.

Analyze repos, surface rich metadata, and check PR health with AI. Get stats, thread insights, and flexible output—fast.


Authentication:
  Use --token or set GITHUB_TOKEN for higher rate limits.
  Get your token at: https://github.com/settings/tokens`,

	Example: `  # Repository information
  repo-doc info golang/go
  repo-doc info microsoft/vscode --prs 10 --format json
 
  # PR analysis
  repo-doc pr-thread golang/go --limit 3
  repo-doc health golang/go --limit 5`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&token, "token", "t", "",
		`GitHub personal access token for authenticated API requests.
Can also be set via GITHUB_TOKEN environment variable.
Without a token, you're limited to 60 requests per hour.
With a token, you get 5000 requests per hour.
Get your token at: https://github.com/settings/tokens`)
}

var token string
