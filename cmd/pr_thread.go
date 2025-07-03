package cmd

import (
	"fmt"
	"log"
	"strings"

	"repo-doc/internal/analyzer"

	"github.com/spf13/cobra"
)

var (
	discussionsLimit int
)

var prThreadCmd = &cobra.Command{
	Use:   "pr-thread [owner/repo or URL]",
	Short: "Display discussion threads from pull requests",
	Long: `Fetch and display discussion threads from the most recent pull requests in a repository.

This command shows the conversation history including:
- PR description (first message)
- General comments on the PR
- Review comments on the code

Results are shown in chronological order for each PR.`,
	Args: cobra.ExactArgs(1),
	Run:  runPRDiscussions,
	Example: `  # Show threads from the 5 most recent PRs
  repo-doc pr-thread golang/go

  # Show threads from 3 most recent PRs
  repo-doc pr-thread golang/go --limit 3

  # Show threads using full GitHub URL
  repo-doc pr-thread https://github.com/golang/go

  # Using authentication for private repositories
  repo-doc pr-thread myorg/private-repo --token ghp_xxxxxxxxxxxx`,
}

func init() {
	rootCmd.AddCommand(prThreadCmd)

	prThreadCmd.Flags().IntVarP(&discussionsLimit, "limit", "l", 5,
		`Number of most recent PRs to fetch threads from (max 20).
Use a higher limit with caution as it may hit rate limits.`)
}

func runPRDiscussions(cmd *cobra.Command, args []string) {
	repoURL := args[0]

	owner, repo, err := analyzer.ParseRepoURL(repoURL)
	if err != nil {
		log.Fatalf("Error parsing repository URL: %v", err)
	}

	if discussionsLimit < 1 || discussionsLimit > 20 {
		discussionsLimit = 5
	}

	a := analyzer.New(token)

	discussions, err := a.FetchPRDiscussions(owner, repo, discussionsLimit)
	if err != nil {
		log.Fatalf("Error fetching PR discussions: %v", err)
	}

	for _, discussion := range discussions {
		statusEmoji := "🟢" // Open PR
		if discussion.Merged {
			statusEmoji = "🟣" // Merged PR
		} else if strings.EqualFold(discussion.State, "closed") {
			statusEmoji = "🔴" // Closed PR
		}

		header := fmt.Sprintf("%s #%d: %s (👤 %s)", statusEmoji, discussion.PRNumber, discussion.Title, discussion.Author)
		fmt.Println("\n" + strings.Repeat("=", len(header)))
		fmt.Println(header)
		fmt.Println(strings.Repeat("=", len(header)))

		for i, msg := range discussion.Messages {
			if i > 0 {
				fmt.Println("\n" + strings.Repeat("─", 60))
			}
			authorEmoji := "💬"
			if msg.IsPRBody {
				authorEmoji = "📝"
			}

			header := fmt.Sprintf("%s %s (%s)", authorEmoji, msg.Author, msg.CreatedAt)
			if msg.IsPRBody {
				header = "📌 " + header
			}

			fmt.Printf("\n%s\n%s\n", header, strings.Repeat("-", len(header)))
			fmt.Println(msg.Body)
		}
		fmt.Println("\n" + strings.Repeat("=", 50))
	}
}
