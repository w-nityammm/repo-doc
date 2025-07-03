package analyzer

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

type RepoInfo struct {
	Name        string
	FullName    string
	Description string
	Stars       int
	Forks       int
	OpenIssues  int
	Language    string
	CreatedAt   string
	UpdatedAt   string
}

type PRInfo struct {
	Number int
	Title  string
	State  string
	Author string
	Merged bool
}

type PRDiscussion struct {
	PRNumber int
	Title    string
	Author   string
	State    string
	Merged   bool
	Messages []DiscussionMessage
}

type DiscussionMessage struct {
	Author    string
	Body      string
	CreatedAt string
	IsPRBody  bool
}

type Analyzer struct {
	client *github.Client
}

func New(token string) *Analyzer {
	client := createGitHubClient(token)
	return &Analyzer{client: client}
}

func ParseRepoURL(url string) (string, string, error) {
	if !strings.Contains(url, "github.com") && strings.Contains(url, "/") {
		parts := strings.Split(url, "/")
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
	}

	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "github.com/")

	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repository format. Use 'owner/repo' or full GitHub URL")
	}

	return parts[0], parts[1], nil
}

func (a *Analyzer) FetchRepoInfo(owner, repo string) (*RepoInfo, error) {
	ctx := context.Background()

	repository, _, err := a.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	info := &RepoInfo{
		Name:        safeString(repository.Name),
		FullName:    safeString(repository.FullName),
		Description: safeString(repository.Description),
		Stars:       safeInt(repository.StargazersCount),
		Forks:       safeInt(repository.ForksCount),
		OpenIssues:  safeInt(repository.OpenIssuesCount),
		Language:    safeString(repository.Language),
	}

	if repository.CreatedAt != nil {
		info.CreatedAt = repository.CreatedAt.Format("2006-01-02")
	}
	if repository.UpdatedAt != nil {
		info.UpdatedAt = repository.UpdatedAt.Format("2006-01-02")
	}

	return info, nil
}

func (a *Analyzer) IsMerged(owner, repo string, prNumber int) (bool, error) {
	ctx := context.Background()
	isMerged, _, err := a.client.PullRequests.IsMerged(ctx, owner, repo, prNumber)
	if err != nil {
		return false, nil // Assume not merged if there's an error
	}
	return isMerged, nil
}

func (a *Analyzer) FetchPullRequests(owner, repo string, limit int) ([]*PRInfo, error) {
	ctx := context.Background()

	opts := &github.PullRequestListOptions{
		State: "all",
		ListOptions: github.ListOptions{
			PerPage: limit,
		},
	}

	prs, _, err := a.client.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}

	var prInfos []*PRInfo
	for _, pr := range prs {
		if len(prInfos) >= limit {
			break
		}

		var author string
		if pr.User != nil && pr.User.Login != nil {
			author = *pr.User.Login
		}

		isMerged := pr.GetState() == "closed" && !pr.GetMergedAt().IsZero()

		prInfos = append(prInfos, &PRInfo{
			Number: pr.GetNumber(),
			Title:  pr.GetTitle(),
			State:  pr.GetState(),
			Author: author,
			Merged: isMerged,
		})
	}

	return prInfos, nil
}

func (a *Analyzer) FetchPRDiscussions(owner, repo string, limit int) ([]*PRDiscussion, error) {
	prs, err := a.FetchPullRequests(owner, repo, limit)
	if err != nil {
		return nil, fmt.Errorf("error fetching pull requests: %v", err)
	}

	var discussions []*PRDiscussion
	for _, pr := range prs {
		isMerged := pr.Merged

		discussion := &PRDiscussion{
			PRNumber: pr.Number,
			Title:    pr.Title,
			Author:   pr.Author,
			State:    pr.State,
			Merged:   isMerged,
		}

		ctx := context.Background()
		prDetail, _, _ := a.client.PullRequests.Get(ctx, owner, repo, pr.Number)
		if prDetail != nil && prDetail.Body != nil && *prDetail.Body != "" {
			discussion.Messages = append(discussion.Messages, DiscussionMessage{
				Author:    pr.Author,
				Body:      *prDetail.Body,
				CreatedAt: prDetail.CreatedAt.Format("2006-01-02 15:04:05"),
				IsPRBody:  true,
			})
		}

		comments, _, _ := a.client.Issues.ListComments(ctx, owner, repo, pr.Number, nil)
		for _, comment := range comments {
			if comment.Body != nil && *comment.Body != "" {
				author := ""
				if comment.User != nil && comment.User.Login != nil {
					author = *comment.User.Login
				}
				discussion.Messages = append(discussion.Messages, DiscussionMessage{
					Author:    author,
					Body:      *comment.Body,
					CreatedAt: comment.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}
		}

		reviewComments, _, _ := a.client.PullRequests.ListComments(ctx, owner, repo, pr.Number, nil)
		for _, comment := range reviewComments {
			if comment.Body != nil && *comment.Body != "" {
				author := ""
				if comment.User != nil && comment.User.Login != nil {
					author = *comment.User.Login
				}
				discussion.Messages = append(discussion.Messages, DiscussionMessage{
					Author:    author,
					Body:      *comment.Body,
					CreatedAt: comment.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}
		}

		discussions = append(discussions, discussion)
	}

	return discussions, nil
}

func createGitHubClient(token string) *github.Client {
	ctx := context.Background()

	// Check token from parameter first, then environment
	githubToken := token
	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_TOKEN")
	}

	if githubToken == "" {
		fmt.Println("Warning: No GitHub token provided. Using unauthenticated client (rate limited)")
		fmt.Println("Set GITHUB_TOKEN environment variable or use --token flag")
		return github.NewClient(nil)
	}

	// Create authenticated client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// Helper functions
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func safeInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}
