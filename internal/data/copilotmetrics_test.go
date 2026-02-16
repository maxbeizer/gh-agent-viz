package data

import (
	"testing"
	"time"
)

func TestDeriveSessionTelemetry_Duration(t *testing.T) {
	created := time.Now().Add(-2 * time.Hour)
	updated := time.Now()

	workspace := LocalSessionWorkspace{}
	telemetry := deriveSessionTelemetry(workspace, created, updated)

	if telemetry.Duration < 1*time.Hour {
		t.Fatalf("expected duration > 1h, got %s", telemetry.Duration)
	}
}

func TestDeriveSessionTelemetry_ZeroTimesProduceZeroDuration(t *testing.T) {
	workspace := LocalSessionWorkspace{}
	telemetry := deriveSessionTelemetry(workspace, time.Time{}, time.Time{})

	if telemetry.Duration != 0 {
		t.Fatalf("expected zero duration for zero times, got %s", telemetry.Duration)
	}
}

func TestDeriveSessionTelemetry_ConversationCounts(t *testing.T) {
	workspace := LocalSessionWorkspace{
		ConversationHistory: []map[string]interface{}{
			{"role": "user", "content": "fix the bug"},
			{"role": "assistant", "content": "I'll fix that"},
			{"role": "user", "content": "thanks"},
			{"role": "assistant", "content": "done!"},
		},
	}

	telemetry := deriveSessionTelemetry(workspace, time.Time{}, time.Time{})

	if telemetry.UserMessages != 2 {
		t.Fatalf("expected 2 user messages, got %d", telemetry.UserMessages)
	}
	if telemetry.AssistantMessages != 2 {
		t.Fatalf("expected 2 assistant messages, got %d", telemetry.AssistantMessages)
	}
	if telemetry.ConversationTurns != 4 {
		t.Fatalf("expected 4 conversation turns, got %d", telemetry.ConversationTurns)
	}
}

func TestDeriveSessionTelemetry_EmptyConversation(t *testing.T) {
	workspace := LocalSessionWorkspace{}
	telemetry := deriveSessionTelemetry(workspace, time.Time{}, time.Time{})

	if telemetry.ConversationTurns != 0 {
		t.Fatalf("expected 0 turns for empty conversation, got %d", telemetry.ConversationTurns)
	}
}

func TestDeriveSessionTelemetry_MixedCaseRoles(t *testing.T) {
	workspace := LocalSessionWorkspace{
		ConversationHistory: []map[string]interface{}{
			{"role": "User", "content": "hello"},
			{"role": "ASSISTANT", "content": "hi"},
			{"role": "system", "content": "prompt"},
		},
	}

	telemetry := deriveSessionTelemetry(workspace, time.Time{}, time.Time{})

	if telemetry.UserMessages != 1 {
		t.Fatalf("expected 1 user message (case insensitive), got %d", telemetry.UserMessages)
	}
	if telemetry.AssistantMessages != 1 {
		t.Fatalf("expected 1 assistant message (case insensitive), got %d", telemetry.AssistantMessages)
	}
	// System messages are not counted
	if telemetry.ConversationTurns != 2 {
		t.Fatalf("expected 2 turns (system excluded), got %d", telemetry.ConversationTurns)
	}
}

func TestFetchOrgMetrics_EmptyOrg(t *testing.T) {
	result := FetchOrgMetrics("")
	if result.Available {
		t.Fatal("expected unavailable for empty org")
	}
}

func TestFetchOrgMetrics_WhitespaceOrg(t *testing.T) {
	result := FetchOrgMetrics("   ")
	if result.Available {
		t.Fatal("expected unavailable for whitespace org")
	}
}
