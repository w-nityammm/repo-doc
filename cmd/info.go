package cmd

import (
	"log"

	"repo-doc/internal/analyzer"
	"repo-doc/internal/output"

	"github.com/spf13/cobra"
)

var (
	format   string
	prs      int
	download string
)

var infoCmd = &cobra.Command{
	Use:   "info [owner/repo or URL]",
	Short: "Get information about a GitHub repository",
	Long: `Analyze a GitHub repository and display comprehensive information including:
- Repository metadata (name, description, language)
- Statistics (stars, forks, open issues)
- Timestamps (created, last updated)
- Recent pull requests (optional)

The repository can be specified in two formats:
  1. Short format: owner/repo (e.g., golang/go)
  2. Full URL: https://github.com/owner/repo

Results can be displayed in multiple formats.`,
	Args: cobra.ExactArgs(1),
	Run:  runAnalyze,
	Example: `  # Basic repository info (table format, no PRs)
  repo-doc info golang/go
  repo-doc info https://github.com/microsoft/vscode

  # Show pull requests (default 5 when --prs used without number)
  repo-doc info golang/go --prs
  repo-doc info golang/go -p

  # Show specific number of pull requests
  repo-doc info golang/go --prs 15
  repo-doc info golang/go -p 25

  # JSON output format
  repo-doc info golang/go --format json
  repo-doc info golang/go -f json

  # Using authentication for higher rate limits
  repo-doc info golang/go --token ghp_xxxxxxxxxxxx --prs 50
  repo-doc info golang/go -t ghp_xxxxxxxxxxxx -p 30 -f json`,
}

func init() {
	rootCmd.AddCommand(infoCmd)

	infoCmd.Flags().StringVarP(&format, "format", "f", "table",
		`Output format for displaying results.
Available options:
  table - Human-readable table format with emojis (default)
  json  - Machine-readable JSON format

Examples:
  --format table  (default, shows nicely formatted table)
  --format json   (shows structured JSON data)
  -f table
  -f json`)

	infoCmd.Flags().IntVarP(&prs, "prs", "p", -1,
		`Number of recent pull requests to display.
Behavior:
  - Not specified: No pull requests shown (default)
  - --prs (no number): Shows 5 recent pull requests
  - --prs N: Shows N recent pull requests (max 100)

Examples:
  --prs      (shows 5 recent PRs)
  --prs 10   (shows 10 recent PRs)`)

}

func runAnalyze(cmd *cobra.Command, args []string) {
	repoURL := args[0]

	owner, repo, err := analyzer.ParseRepoURL(repoURL)
	if err != nil {
		log.Fatalf("Error parsing repository URL: %v", err)
	}

	prLimit := determinePRLimit(cmd)

	if prLimit > 100 {
		log.Fatalf("PR limit must be 100 or less")
	}

	a := analyzer.New(token)

	repoInfo, err := a.FetchRepoInfo(owner, repo)
	if err != nil {
		log.Fatalf("Error fetching repository info: %v", err)
	}

	var prInfos []*analyzer.PRInfo
	if prLimit > 0 {
		prInfos, err = a.FetchPullRequests(owner, repo, prLimit)
		if err != nil {
			log.Fatalf("Error fetching pull requests: %v", err)
		}
	}

	outputManager := output.New(format, download)

	if err := outputManager.Display(repoInfo, prInfos); err != nil {
		log.Fatalf("Error displaying output: %v", err)
	}
}

func determinePRLimit(cmd *cobra.Command) int {
	prsFlagSet := cmd.Flags().Changed("prs")

	if !prsFlagSet {
		return 0
	}

	if prs == -1 {
		return 5
	}
	return prs
}
