package sparkline

import (
	"math"
	"strings"
	"time"
)

// blocks maps normalized values (0.0–1.0) to Unicode block characters.
var blocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Generate returns a sparkline string of the given width visualizing activity
// recency. Running sessions ramp up toward the right, completed sessions peak
// early then decay, and queued sessions stay mostly flat.
func Generate(status string, createdAt, updatedAt time.Time, width int) string {
	if width <= 0 {
		width = 8
	}

	status = normalizeStatus(status)

	// Fallback when timestamps are missing.
	if createdAt.IsZero() && updatedAt.IsZero() {
		return fallback(status, width)
	}

	now := time.Now()

	switch status {
	case "queued":
		return queued(width)
	case "running", "active", "in progress", "open":
		return running(createdAt, updatedAt, now, width)
	default:
		return completed(createdAt, updatedAt, now, width)
	}
}

// running produces a ramp-up sparkline: activity increases toward the right.
func running(createdAt, updatedAt, now time.Time, width int) string {
	var buf strings.Builder
	buf.Grow(width * 4) // up to 4 bytes per rune
	for i := range width {
		t := float64(i) / float64(width-1)
		// Quadratic ease-in gives a pleasing ramp.
		v := t * t
		buf.WriteRune(blockChar(v))
	}
	return buf.String()
}

// completed produces a decay sparkline: peaks early then fades.
// The faster it faded (longer since updatedAt), the steeper the decay.
func completed(createdAt, updatedAt, now time.Time, width int) string {
	// recencyRatio: 0 = just finished, 1 = finished long ago.
	age := now.Sub(createdAt).Seconds()
	if age <= 0 {
		age = 1
	}
	recency := now.Sub(updatedAt).Seconds()
	if recency < 0 {
		recency = 0
	}
	recencyRatio := recency / age
	if recencyRatio > 1 {
		recencyRatio = 1
	}

	// peakPos: where in [0,1) the sparkline peaks.
	// Recent completions peak near the middle; old ones peak near the left.
	peakPos := 0.5 * (1 - recencyRatio)
	if peakPos < 0.05 {
		peakPos = 0.05
	}

	var buf strings.Builder
	buf.Grow(width * 4)
	for i := range width {
		t := float64(i) / float64(width-1)
		dist := math.Abs(t - peakPos)
		v := math.Exp(-3 * dist / (1 - recencyRatio + 0.15))
		if v > 1 {
			v = 1
		}
		buf.WriteRune(blockChar(v))
	}
	return buf.String()
}

// queued produces a mostly-flat sparkline with a slight uptick at the end.
func queued(width int) string {
	var buf strings.Builder
	buf.Grow(width * 4)
	for i := range width {
		if i == width-1 {
			buf.WriteRune(blocks[1]) // ▂
		} else {
			buf.WriteRune(blocks[0]) // ▁
		}
	}
	return buf.String()
}

// fallback returns a uniform sparkline based on status when timestamps are
// missing.
func fallback(status string, width int) string {
	var ch rune
	switch status {
	case "running", "active", "in progress", "open":
		ch = blocks[7] // █
	case "completed", "stopped", "done":
		ch = blocks[3] // ▄
	default:
		ch = blocks[0] // ▁
	}
	return strings.Repeat(string(ch), width)
}

// blockChar maps a value in [0,1] to a block character.
func blockChar(v float64) rune {
	idx := int(v * float64(len(blocks)-1))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(blocks) {
		idx = len(blocks) - 1
	}
	return blocks[idx]
}

func normalizeStatus(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
