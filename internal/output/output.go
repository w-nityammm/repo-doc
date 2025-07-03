package output

import (
	"encoding/json"
	"fmt"
	"repo-doc/internal/analyzer"
	"strings"
)

type Manager struct {
	format string
}

func New(format, download string) *Manager {
	return &Manager{
		format: format,
	}
}

func (m *Manager) Display(info *analyzer.RepoInfo, prs []*analyzer.PRInfo) error {

	switch m.format {
	case "json":
		return m.handleJSON(info, prs)
	case "table":
		return m.handleTable(info, prs)
	default:
		return fmt.Errorf("unknown format: %s. Use 'table' or 'json'", m.format)
	}
}

func (m *Manager) handleJSON(info *analyzer.RepoInfo, prs []*analyzer.PRInfo) error {
	data := struct {
		Repository   *analyzer.RepoInfo `json:"repository"`
		PullRequests []*analyzer.PRInfo `json:"pull_requests"`
	}{
		Repository:   info,
		PullRequests: prs,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}
	if m.format == "json" {
		fmt.Println(string(jsonData))
	}

	return nil
}

func (m *Manager) handleTable(info *analyzer.RepoInfo, prs []*analyzer.PRInfo) error {

	output := m.formatTable(info, prs)
	fmt.Print(output)

	return nil
}

func (m *Manager) formatTable(info *analyzer.RepoInfo, prs []*analyzer.PRInfo) string {
	output := ""
	lineSeparator := strings.Repeat("=", 80) + "\n"

	output += lineSeparator
	output += fmt.Sprintf("📦 %s\n", info.FullName)
	output += lineSeparator

	if info.Description != "" {
		output += fmt.Sprintf("📝 %s\n\n", info.Description)
	}

	output += fmt.Sprintf("⭐ Stars:        %d\n", info.Stars)
	output += fmt.Sprintf("🍴 Forks:        %d\n", info.Forks)
	output += fmt.Sprintf("🐛 Open Issues:  %d\n", info.OpenIssues)
	output += fmt.Sprintf("💻 Language:     %s\n", info.Language)
	output += fmt.Sprintf("📅 Created:      %s\n", info.CreatedAt)
	output += fmt.Sprintf("🔄 Updated:      %s\n", info.UpdatedAt)

	if len(prs) > 0 {
		output += "\n" + lineSeparator
		output += fmt.Sprintf("📋 Recent Pull Requests (%d)\n", len(prs))
		output += lineSeparator

		for _, pr := range prs {
			status := "🟢" // Open PR
			if pr.Merged {
				status = "🟣" // Merged PR
			} else if pr.State == "closed" {
				status = "🔴" // Closed PR
			}

			output += fmt.Sprintf("%s #%d: %s\n", status, pr.Number, pr.Title)
			output += fmt.Sprintf("   👤 %s\n\n", pr.Author)
		}
	}

	return output
}
