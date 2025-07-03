# repo-doc

Expansion of [go-repo](https://github.com/w-nityammm/go-repo).

## Prerequisites

- Go 1.16 or higher
- GitHub Personal Access Token (Recommended)
- Google Gemini API Key (Required for sentiment analysis)

## Installation

### Prerequisites
1. Make sure you have Go 1.16 or higher installed
2. Ensure your `GOPATH` is set up correctly (usually `~/go` on Unix or `%USERPROFILE%\go` on Windows)
3. Add `$GOPATH/bin` to your system's `PATH` environment variable

### Install from Source (Recommended)
```bash
# Clone the repository
git clone https://github.com/w-nityammm/repo-doc.git
cd repo-doc

# Install the tool
go install
```
After installation, verify it works by running:
```
repo-doc version
```

### Troubleshooting
If `repo-doc` command is not found:
1. Make sure `$GOPATH/bin` is in your system's `PATH`
2. On Unix/Linux:
   ```bash
   echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
   source ~/.bashrc
   ```
3. On Windows:
   - Open System Properties > Advanced > Environment Variables
   - Add or update the PATH variable to include `%USERPROFILE%\go\bin`

## Authentication

### GitHub Authentication
To avoid GitHub API rate limits (60 requests/hour for unauthenticated requests), use a GitHub personal access token:

1. Create a token at: https://github.com/settings/tokens
2. Use it in two ways:
   - Pass as a flag: `--token YOUR_TOKEN` or `-t YOUR_TOKEN`
   - Set it as an environment variable:
   ```bash
   # Windows
   set GITHUB_TOKEN=your_token_here
   
   # Linux/macOS
   export GITHUB_TOKEN=your_token_here
   ```
   Or create a `.env` file in the project root:
   ```
   GITHUB_TOKEN=your_token_here
   ```

### Gemini API Setup
For sentiment analysis features, you'll need a Google Gemini API key:

1. Get an API key from [Google AI Studio](https://makersuite.google.com/)
2. Set it as an environment variable:
   ```bash
   # Windows
   set GEMINI_API_KEY=your_api_key_here
   
   # Linux/macOS
   export GEMINI_API_KEY=your_api_key_here
   ```
   Or add it to your `.env` file:
   ```
   GEMINI_API_KEY=your_api_key_here
   ```
## Usage

### Basic Usage

```bash
# Using owner/repo
repo-doc info golang/go

# Using full GitHub URL
repo-doc info https://github.com/golang/go
```

### Include Pull Requests

```bash
# Show specific number of pull requests (up to 100)
repo-doc info golang/go --prs 15
repo-doc info golang/go -p 15
```

### Output Formats

```bash
# Table format (default)
repo-doc info golang/go --format table
repo-doc info golang/go -f table

# JSON format
repo-doc info golang/go --format json
repo-doc info golang/go -f json
```

### PR Threads

View discussion threads from pull requests including comments and reviews:

```bash
# Show threads from 5 most recent PRs (default)
repo-doc pr-thread golang/go

# Show threads from specific number of PRs
repo-doc pr-thread golang/go --limit 3

# Using full GitHub URL
repo-doc pr-thread https://github.com/golang/go

```

### PR Health Analysis

Analyze the health of pull requests using sentiment analysis:

```bash
# Analyze health of last 5 PRs (default)
repo-doc health golang/go

# Analyze specific number of PRs
repo-doc health golang/go --limit 10

# Using full GitHub URL
repo-doc health https://github.com/golang/go
```

### Help

```bash
repo-doc --help
```

## Example Usage

### Repository Information
```bash
repo-doc info golang/go --prs 2
```

Example output:
```
================================================================================
📦 golang/go
================================================================================
📝 The Go programming language

⭐ Stars:        128418
🍴 Forks:        18139
🐛 Open Issues:  9359
💻 Language:     Go
📅 Created:      2014-08-19
🔄 Updated:      2025-06-17

================================================================================
📋 Recent Pull Requests (2)
================================================================================
🟢 #74251: net/http: reduce allocs in CrossOriginProtection.Check
   👤 jub0bs

🔴 #74249: Victor001 hash patch 1
   👤 victor001-hash
```

### PR Health Analysis
```bash
repo-doc health golang/go --limit 10
```

Example output:
```
📊 PR Health Report (10 PRs, 42 messages analyzed)
==================================================

🎭 Sentiment Analysis:
✅ Positive: 65.0%
😐 Neutral:  25.0%
❌ Negative: 10.0%
📈 Average Sentiment: 0.72/1.0

💬 Sample Messages:
✅ [0.85] This is an excellent contribution! Very clean implementation.
➖ [0.5] I've left some minor comments for improvement.
❌ [0.2] Bro what is this $hit.

🏥 Health Assessment:
👍 Good health - Generally positive discussions
==================================================
```
