package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"

	"repo-doc/internal/analyzer"
)

var (
	healthLimit int
)

type SentimentResponse struct {
	Sentiment string  `json:"sentiment"`
	Score     float64 `json:"score"`
}

type HealthReport struct {
	PRCount          int
	MessageCount     int
	PositiveScore    float64
	NegativeScore    float64
	NeutralScore     float64
	AverageSentiment float64
	Messages         []MessageAnalysis
}

type MessageAnalysis struct {
	Content   string
	Sentiment string
	Score     float64
}

var healthCmd = &cobra.Command{
	Use:   "health [owner/repo or URL]",
	Short: "Analyze PR health using sentiment analysis",
	Long: `Analyze the health of pull requests using sentiment analysis.

This command analyzes the sentiment of PR discussions to provide
insights into the overall health and tone of the project's PRs.`,
	Args: cobra.ExactArgs(1),
	Run:  runHealthAnalysis,
	Example: `  # Analyze health of last 5 PRs
  repo-doc health golang/go

  # Analyze specific number of PRs
  repo-doc health golang/go --limit 10

  # Using full GitHub URL
  repo-doc health https://github.com/golang/go`,
}

func init() {
	rootCmd.AddCommand(healthCmd)

	healthCmd.Flags().IntVarP(&healthLimit, "limit", "l", 5,
		`Number of most recent PRs to analyze (max 20).`)
}

func runHealthAnalysis(cmd *cobra.Command, args []string) {
	if os.Getenv("GEMINI_API_KEY") == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required for health analysis. Please set it in .env file or environment variables")
	}

	repoURL := args[0]

	owner, repo, err := analyzer.ParseRepoURL(repoURL)
	if err != nil {
		log.Fatalf("Error parsing repository URL: %v", err)
	}

	if healthLimit < 1 || healthLimit > 20 {
		healthLimit = 5
	}

	a := analyzer.New(token)

	discussions, err := a.FetchPRDiscussions(owner, repo, healthLimit)
	if err != nil {
		log.Fatalf("Error fetching PR discussions: %v", err)
	}

	report := analyzePRHealth(discussions)
	displayHealthReport(report)
}

func cleanTextForAnalysis(text string) string {
	// Remove code blocks
	re := regexp.MustCompile("(?s)```.*?```")
	text = re.ReplaceAllString(text, " ")

	// Remove inline code
	re = regexp.MustCompile("`[^`]+`")
	text = re.ReplaceAllString(text, " ")

	// Remove URLs
	re = regexp.MustCompile(`https?://\S+`)
	text = re.ReplaceAllString(text, " ")

	// Remove markdown headers, lists, etc.
	re = regexp.MustCompile(`[#*\-_=~]+`)
	text = re.ReplaceAllString(text, " ")

	// Remove extra whitespace
	text = strings.Join(strings.Fields(text), " ")

	return strings.TrimSpace(text)
}

func analyzeWithGemini(ctx context.Context, text string) (string, float64, error) {
	cleanText := cleanTextForAnalysis(text)
	if cleanText == "" {
		return "neutral", 0.5, nil
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", 0, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	clientCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	client, err := genai.NewClient(clientCtx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create Gemini client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro-latest")

	temp := float32(0.2)
	topP := float32(0.9)
	topK := int32(40)
	maxTokens := int32(1024)

	model.Temperature = &temp
	model.TopP = &topP
	model.TopK = &topK
	model.MaxOutputTokens = &maxTokens

	prompt := fmt.Sprintf(`Analyze the sentiment of this GitHub PR discussion text and respond with a JSON object containing "sentiment" (one of: "positive", "neutral", "negative") and "score" (0.0 to 1.0, where 0 is most negative and 1 is most positive).

Text to analyze:
%s

Respond with only the JSON object, nothing else.`, cleanText)

	log.Printf("Sending request to model with prompt length: %d", len(prompt))
	resp, err := model.GenerateContent(clientCtx, genai.Text(prompt))
	if err != nil {
		log.Printf("Error details: %v", err)
		return "", 0, fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", 0, fmt.Errorf("no content in response")
	}
	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			responseText += string(textPart)
		}
	}

	log.Printf("Raw response: %s", responseText)

	var result struct {
		Sentiment string  `json:"sentiment"`
		Score     float64 `json:"score"`
	}

	jsonStart := strings.Index(responseText, "{")
	jsonEnd := strings.LastIndex(responseText, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return "", 0, fmt.Errorf("invalid JSON response: %s", responseText)
	}

	jsonStr := responseText[jsonStart : jsonEnd+1]
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("Failed to parse JSON response: %v\nResponse: %s", err, responseText)
		return "", 0, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	switch result.Sentiment {
	case "positive", "neutral", "negative":
	default:
		return "", 0, fmt.Errorf("invalid sentiment value: %s", result.Sentiment)
	}
	if result.Score < 0 || result.Score > 1 {
		return "", 0, fmt.Errorf("score out of range: %f", result.Score)
	}

	log.Printf("Analysis result - Sentiment: %s, Score: %.2f", result.Sentiment, result.Score)
	return result.Sentiment, result.Score, nil
}

func analyzeSentiment(text string) (string, float64) {
	if text == "" {
		return "neutral", 0.5
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sentiment, score, err := analyzeWithGemini(ctx, text)
	if err != nil {
		log.Printf("Error analyzing message with Gemini: %v", err)
		return "neutral", 0.5
	}

	return sentiment, score
}

func analyzePRHealth(discussions []*analyzer.PRDiscussion) *HealthReport {
	report := &HealthReport{
		PRCount:  len(discussions),
		Messages: make([]MessageAnalysis, 0),
	}

	totalScore := 0.0
	messageCount := 0

	for _, d := range discussions {
		for _, msg := range d.Messages {
			if msg.Body == "" || isBotComment(msg.Author) {
				continue
			}

			sentimentLabel, score := analyzeSentiment(msg.Body)

			if sentimentLabel == "" {
				switch {
				case score > 0.7:
					sentimentLabel = "positive"
				case score < 0.4:
					sentimentLabel = "negative"
				default:
					sentimentLabel = "neutral"
				}
			}

			msgAnalysis := MessageAnalysis{
				Content:   msg.Body,
				Sentiment: sentimentLabel,
				Score:     score,
			}

			report.Messages = append(report.Messages, msgAnalysis)

			switch sentimentLabel {
			case "positive":
				report.PositiveScore++
			case "negative":
				report.NegativeScore++
			default:
				report.NeutralScore++
			}

			totalScore += score
			messageCount++
		}
	}

	report.MessageCount = len(report.Messages)
	if report.MessageCount > 0 {
		report.AverageSentiment = totalScore / float64(report.MessageCount)
	}

	return report
}

func isBotComment(author string) bool {
	botNames := []string{
		// GitHub bots
		"dependabot", "github-actions", "github[bot]", "actions-user", "actions\\[bot\\]",
		// CI/CD services
		"travis", "circleci", "jenkins", "gitlab-ci", "azure-pipelines", "circleci[bot]",
		// Code quality bots
		"codecov", "codeclimate", "sonarcloud", "snyk-bot", "dependabot-preview",
		// Common bot patterns
		"bot", "ci", "cd", "deploy", "test", "automation", "bors-", "tldr-",
		// Cloud providers
		"aws-", "gcp-", "azure-", "google-cloud", "aws-sdk",
		// Other common bots
		"renovate", "greenkeeper", "hound", "stale", "mergify", "allcontributors", "code-rabbit", "app/", "app\\/",
	}

	author = strings.ToLower(author)
	for _, bot := range botNames {
		if strings.Contains(author, bot) {
			return true
		}
	}

	if strings.HasSuffix(author, "[bot]") ||
		strings.HasSuffix(author, "-bot") ||
		strings.HasSuffix(author, "-ci") ||
		strings.HasSuffix(author, "-deploy") {
		return true
	}

	return false
}

func displayHealthReport(report *HealthReport) {
	if report.MessageCount == 0 {
		fmt.Println("\n🔍 No messages found to analyze.")
		return
	}

	fmt.Printf("\n📊 PR Health Report (%d PRs, %d messages analyzed)\n", report.PRCount, report.MessageCount)
	fmt.Println(strings.Repeat("=", 50))
	positivePct := 0.0
	neutralPct := 0.0
	negativePct := 0.0

	if report.MessageCount > 0 {
		total := float64(report.MessageCount)
		positivePct = (float64(report.PositiveScore) / total) * 100
		neutralPct = (float64(report.NeutralScore) / total) * 100
		negativePct = (float64(report.NegativeScore) / total) * 100
	}

	fmt.Printf("\n🎭 Sentiment Analysis:")
	fmt.Printf("\n✅ Positive: %.1f%%\n", positivePct)
	fmt.Printf("😐 Neutral:  %.1f%%\n", neutralPct)
	fmt.Printf("❌ Negative: %.1f%%\n", negativePct)
	fmt.Printf("📈 Average Sentiment: %.1f/1.0\n", report.AverageSentiment)

	fmt.Println("\n💬 Sample Messages:")
	printed := 0
	for _, msg := range report.Messages {
		if printed >= 3 {
			break
		}
		var emoji string
		switch msg.Sentiment {
		case "positive":
			emoji = "✅"
		case "negative":
			emoji = "❌"
		default:
			emoji = "➖"
		}
		content := msg.Content
		if len(content) > 100 {
			content = content[:97] + "..."
		}
		fmt.Printf("%s [%.1f] %s\n", emoji, msg.Score, content)
		printed++
	}

	fmt.Println("\n🏥 Health Assessment:")
	switch {
	case report.MessageCount == 0:
		fmt.Println("ℹ️  No messages to analyze")
	case report.NegativeScore/float64(report.MessageCount) > 0.5:
		fmt.Println("⚠️  Needs attention - High level of negative sentiment")
	case report.PositiveScore/float64(report.MessageCount) > 0.7:
		fmt.Println("🌟 Excellent health - Very positive discussions")
	case report.AverageSentiment > 0.6:
		fmt.Println("👍 Good health - Generally positive discussions")
	case report.NeutralScore/float64(report.MessageCount) > 0.7:
		fmt.Println("➖ Neutral - Mostly technical discussions")
	default:
		fmt.Println("⚠️  Mixed sentiment - Review recommended")
	}

	fmt.Println(strings.Repeat("=", 50))
}
