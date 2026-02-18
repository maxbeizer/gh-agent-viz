package data

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleLog = `2025-01-15 10:00:00 [INFO] Starting session
2025-01-15 10:00:01 [INFO] Flushed 5 events to session abc12345-1234-1234-1234-abcdef123456
2025-01-15 10:00:02 [DEBUG] response (Request-ID: req-123)
2025-01-15 10:00:02 [DEBUG] data: {"id":"chatcmpl-1"}
2025-01-15 10:00:02 [DEBUG] {"model":"claude-opus-4.5","usage":{"completion_tokens":100,"prompt_tokens":5000,"prompt_tokens_details":{"cached_tokens":3000},"total_tokens":5100}}
2025-01-15 10:00:10 [DEBUG] response (Request-ID: req-456)
2025-01-15 10:00:10 [DEBUG] data: {"id":"chatcmpl-2"}
2025-01-15 10:00:10 [DEBUG] {"model":"claude-opus-4.5","usage":{"completion_tokens":200,"prompt_tokens":8000,"prompt_tokens_details":{"cached_tokens":6000},"total_tokens":8200}}
2025-01-15 10:01:00 [INFO] Flushed 3 events to session def67890-5678-5678-5678-abcdef567890
2025-01-15 10:01:02 [DEBUG] response (Request-ID: req-789)
2025-01-15 10:01:02 [DEBUG] data: {"id":"chatcmpl-3"}
2025-01-15 10:01:02 [DEBUG] {"model":"gpt-4.1","usage":{"completion_tokens":50,"prompt_tokens":2000,"prompt_tokens_details":{"cached_tokens":1000},"total_tokens":2050}}
`

func TestParseLogFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "process-test.log")
	if err := os.WriteFile(path, []byte(sampleLog), 0o644); err != nil {
		t.Fatal(err)
	}

	result := map[string]*TokenUsage{}
	parseLogFile(path, result)

	if len(result) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(result))
	}

	s1 := result["abc12345-1234-1234-1234-abcdef123456"]
	if s1 == nil {
		t.Fatal("session abc not found")
	}
	if s1.InputTokens != 13000 {
		t.Errorf("expected input 13000, got %d", s1.InputTokens)
	}
	if s1.OutputTokens != 300 {
		t.Errorf("expected output 300, got %d", s1.OutputTokens)
	}
	if s1.CachedTokens != 9000 {
		t.Errorf("expected cached 9000, got %d", s1.CachedTokens)
	}
	if s1.Calls != 2 {
		t.Errorf("expected 2 calls, got %d", s1.Calls)
	}
	if s1.Model != "claude-opus-4.5" {
		t.Errorf("expected model claude-opus-4.5, got %s", s1.Model)
	}

	s2 := result["def67890-5678-5678-5678-abcdef567890"]
	if s2 == nil {
		t.Fatal("session def not found")
	}
	if s2.InputTokens != 2000 {
		t.Errorf("expected input 2000, got %d", s2.InputTokens)
	}
	if s2.Calls != 1 {
		t.Errorf("expected 1 call, got %d", s2.Calls)
	}
	if s2.Model != "gpt-4.1" {
		t.Errorf("expected model gpt-4.1, got %s", s2.Model)
	}
}

func TestFetchTokenUsageEmptyDir(t *testing.T) {
	dir := t.TempDir()
	result, err := fetchTokenUsageFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

func TestFetchTokenUsageMissingDir(t *testing.T) {
	result, err := fetchTokenUsageFromDir("")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

func TestMalformedLinesSkipped(t *testing.T) {
	dir := t.TempDir()
	content := `2025-01-15 10:00:01 [INFO] Flushed 5 events to session abc12345-1234-1234-1234-abcdef123456
2025-01-15 10:00:02 [DEBUG] response (Request-ID: req-123)
2025-01-15 10:00:02 [DEBUG] data: {"id":"chatcmpl-1"}
2025-01-15 10:00:02 [DEBUG] {not valid json at all!!!
2025-01-15 10:00:03 [DEBUG] response (Request-ID: req-456)
2025-01-15 10:00:03 [DEBUG] data: {"id":"chatcmpl-2"}
2025-01-15 10:00:03 [DEBUG] {"model":"claude-opus-4.5","usage":{"completion_tokens":50,"prompt_tokens":1000,"prompt_tokens_details":{"cached_tokens":500},"total_tokens":1050}}
`
	path := filepath.Join(dir, "process-test.log")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	result := map[string]*TokenUsage{}
	parseLogFile(path, result)

	s := result["abc12345-1234-1234-1234-abcdef123456"]
	if s == nil {
		t.Fatal("session not found")
	}
	if s.Calls != 1 {
		t.Errorf("expected 1 valid call, got %d", s.Calls)
	}
	if s.InputTokens != 1000 {
		t.Errorf("expected input 1000, got %d", s.InputTokens)
	}
}
