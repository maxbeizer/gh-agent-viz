package taskdetail

import (
	"fmt"
	"strings"
	"time"

	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

const (
	blockIdle   = '░'
	blockInact  = '▒'
	blockActive = '▓'
	blockFull   = '█'
)

// nowFunc is overridable for testing.
var nowFunc = time.Now

// RenderTimeline returns a Unicode timeline bar showing the session lifecycle.
// The bar uses ░ (idle/pre-creation), ▒ (created but inactive), ▓ (active), █ (current activity).
// width is the character width of the bar (typically 24-32 chars).
func RenderTimeline(session *data.Session, width int) string {
	if session == nil || session.CreatedAt.IsZero() || width < 1 {
		return ""
	}

	now := nowFunc()
	created := session.CreatedAt
	updated := session.UpdatedAt

	if created.After(now) {
		created = now
	}

	totalSpan := now.Sub(created)
	if totalSpan <= 0 {
		// Session just created — show a single block for current state
		bar := strings.Repeat(string(blockChar(session.Status)), width)
		return fmt.Sprintf("%s  %s → now", bar, formatRelative(totalSpan))
	}

	// Determine the active window: from created to updated (or created if no update)
	activeEnd := created
	if !updated.IsZero() && updated.After(created) {
		activeEnd = updated
		if activeEnd.After(now) {
			activeEnd = now
		}
	}

	bar := make([]rune, width)
	for i := range bar {
		// Map character position to a point in time
		posTime := created.Add(totalSpan * time.Duration(i) / time.Duration(width))

		if posTime.Before(activeEnd) || posTime.Equal(activeEnd) {
			if data.StatusIsActive(session.Status) {
				bar[i] = blockFull
			} else {
				bar[i] = blockActive
			}
		} else {
			bar[i] = blockIdle
		}
	}

	// Mark the right edge based on current status
	if data.StatusIsActive(session.Status) {
		bar[width-1] = blockFull
	}

	return fmt.Sprintf("%s  %s → now", string(bar), formatRelative(totalSpan))
}

func blockChar(status string) rune {
	if data.StatusIsActive(status) {
		return blockFull
	}
	return blockActive
}

func formatRelative(d time.Duration) string {
	if d < time.Minute {
		return "<1m ago"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	hours := int(d.Hours())
	if hours < 24 {
		return fmt.Sprintf("%dh ago", hours)
	}
	days := hours / 24
	return fmt.Sprintf("%dd ago", days)
}
