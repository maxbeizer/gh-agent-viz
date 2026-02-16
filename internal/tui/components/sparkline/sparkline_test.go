package sparkline

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

var validBlocks = map[rune]bool{
	'▁': true, '▂': true, '▃': true, '▄': true,
	'▅': true, '▆': true, '▇': true, '█': true,
}

func TestGenerateRunningSession(t *testing.T) {
	now := time.Now()
	created := now.Add(-30 * time.Minute)
	updated := now.Add(-1 * time.Minute) // recently active

	result := Generate("running", created, updated, 8)

	if utf8.RuneCountInString(result) != 8 {
		t.Errorf("expected 8 runes, got %d: %q", utf8.RuneCountInString(result), result)
	}

	runes := []rune(result)
	// Running should ramp up: first char <= last char
	if runes[0] > runes[len(runes)-1] {
		t.Errorf("running sparkline should ramp up, got %q", result)
	}
}

func TestGenerateCompletedSession(t *testing.T) {
	now := time.Now()
	created := now.Add(-24 * time.Hour)
	updated := now.Add(-23 * time.Hour) // finished long ago

	result := Generate("completed", created, updated, 8)

	if utf8.RuneCountInString(result) != 8 {
		t.Errorf("expected 8 runes, got %d: %q", utf8.RuneCountInString(result), result)
	}

	runes := []rune(result)
	// Completed old session: first chars should be >= last chars
	if runes[0] < runes[len(runes)-1] {
		t.Errorf("old completed sparkline should peak early, got %q", result)
	}
}

func TestGenerateQueuedSession(t *testing.T) {
	now := time.Now()
	created := now.Add(-1 * time.Minute)
	updated := created

	result := Generate("queued", created, updated, 8)

	runes := []rune(result)
	if utf8.RuneCountInString(result) != 8 {
		t.Errorf("expected 8 runes, got %d: %q", utf8.RuneCountInString(result), result)
	}

	// All but last should be ▁, last should be ▂
	for i, r := range runes {
		if i < len(runes)-1 {
			if r != '▁' {
				t.Errorf("queued sparkline position %d should be ▁, got %c", i, r)
			}
		} else {
			if r != '▂' {
				t.Errorf("queued sparkline last position should be ▂, got %c", r)
			}
		}
	}
}

func TestGenerateZeroTimestamps(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected rune
	}{
		{"running fallback", "running", '█'},
		{"completed fallback", "completed", '▄'},
		{"unknown fallback", "unknown-status", '▁'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Generate(tt.status, time.Time{}, time.Time{}, 8)
			if utf8.RuneCountInString(result) != 8 {
				t.Errorf("expected 8 runes, got %d", utf8.RuneCountInString(result))
			}
			expected := strings.Repeat(string(tt.expected), 8)
			if result != expected {
				t.Errorf("expected %q, got %q", expected, result)
			}
		})
	}
}

func TestGenerateCorrectWidth(t *testing.T) {
	now := time.Now()
	created := now.Add(-1 * time.Hour)
	updated := now.Add(-5 * time.Minute)

	for _, w := range []int{1, 4, 8, 12, 20} {
		result := Generate("running", created, updated, w)
		got := utf8.RuneCountInString(result)
		if got != w {
			t.Errorf("width %d: expected %d runes, got %d", w, w, got)
		}
	}
}

func TestGenerateDefaultWidth(t *testing.T) {
	result := Generate("running", time.Time{}, time.Time{}, 0)
	if utf8.RuneCountInString(result) != 8 {
		t.Errorf("default width should be 8, got %d", utf8.RuneCountInString(result))
	}
}

func TestGenerateOnlyValidBlocks(t *testing.T) {
	now := time.Now()
	cases := []struct {
		status  string
		created time.Time
		updated time.Time
	}{
		{"running", now.Add(-1 * time.Hour), now.Add(-1 * time.Minute)},
		{"completed", now.Add(-24 * time.Hour), now.Add(-12 * time.Hour)},
		{"queued", now.Add(-1 * time.Minute), now.Add(-1 * time.Minute)},
		{"running", time.Time{}, time.Time{}},
		{"unknown", time.Time{}, time.Time{}},
	}

	for _, tc := range cases {
		result := Generate(tc.status, tc.created, tc.updated, 8)
		for _, r := range result {
			if !validBlocks[r] {
				t.Errorf("invalid block char %c (U+%04X) in %q for status=%q",
					r, r, result, tc.status)
			}
		}
	}
}
