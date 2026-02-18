package data

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// TokenUsage holds aggregated token usage for a single session.
type TokenUsage struct {
	SessionID    string
	Model        string
	InputTokens  int64
	OutputTokens int64
	CachedTokens int64
	Calls        int
}

// responseUsage mirrors the JSON usage block in log output.
type responseUsage struct {
	CompletionTokens int64 `json:"completion_tokens"`
	PromptTokens     int64 `json:"prompt_tokens"`
	PromptDetails    struct {
		CachedTokens int64 `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
	TotalTokens int64 `json:"total_tokens"`
}

// responseBlock mirrors the JSON response block containing model and usage.
type responseBlock struct {
	Model string        `json:"model"`
	Usage responseUsage `json:"usage"`
}

var sessionFlushRe = regexp.MustCompile(`Flushed \d+ events to session ([0-9a-fA-F-]{36})`)

// FetchTokenUsage parses recent Copilot CLI log files and returns per-session
// token usage. Only files modified in the last 7 days are parsed.
func FetchTokenUsage() (map[string]*TokenUsage, error) {
	return fetchTokenUsageFromDir(defaultLogDir())
}

func defaultLogDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".copilot", "logs")
}

func fetchTokenUsageFromDir(logDir string) (map[string]*TokenUsage, error) {
	if logDir == "" {
		return map[string]*TokenUsage{}, nil
	}

	files, err := filepath.Glob(filepath.Join(logDir, "process-*.log"))
	if err != nil {
		return map[string]*TokenUsage{}, nil
	}

	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	result := map[string]*TokenUsage{}

	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil || info.ModTime().Before(cutoff) {
			continue
		}
		parseLogFile(f, result)
	}

	return result, nil
}

func parseLogFile(path string, result map[string]*TokenUsage) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 512*1024), 2*1024*1024)

	var currentSession string
	var inResponse bool

	for scanner.Scan() {
		line := scanner.Text()

		// Track session ID from flush lines
		if matches := sessionFlushRe.FindStringSubmatch(line); len(matches) == 2 {
			currentSession = matches[1]
			continue
		}

		// Detect start of response block
		if strings.Contains(line, "[DEBUG] response (Request-ID") {
			inResponse = true
			continue
		}

		// Look for JSON data lines inside response blocks
		if inResponse && strings.Contains(line, "[DEBUG] {") {
			jsonStr := extractJSON(line)
			if jsonStr != "" {
				var block responseBlock
				if err := json.Unmarshal([]byte(jsonStr), &block); err == nil && block.Usage.TotalTokens > 0 {
					sid := currentSession
					if sid == "" {
						sid = "_unknown"
					}
					usage, ok := result[sid]
					if !ok {
						usage = &TokenUsage{SessionID: sid}
						result[sid] = usage
					}
					usage.InputTokens += block.Usage.PromptTokens
					usage.OutputTokens += block.Usage.CompletionTokens
					usage.CachedTokens += block.Usage.PromptDetails.CachedTokens
					usage.Calls++
					if block.Model != "" {
						usage.Model = block.Model
					}
				}
			}
			inResponse = false
			continue
		}

		// Reset response tracking on non-data lines after response header
		if inResponse && !strings.Contains(line, "[DEBUG] data:") {
			inResponse = false
		}
	}
}

// extractJSON pulls a JSON object from a log line like `[DEBUG] { ... }`
func extractJSON(line string) string {
	idx := strings.Index(line, "{")
	if idx < 0 {
		return ""
	}
	return line[idx:]
}
