package sparkline

import (
	"math"
	"strings"
)

// SparkChars contains the 8 block characters used for sparklines.
const SparkChars = "▁▂▃▄▅▆▇█"

// HeatChars contains 5 intensity levels for heatmaps.
const HeatChars = " ░▒▓█"

var sparkRunes = []rune(SparkChars)
var heatRunes = []rune(HeatChars)

// Render renders a sparkline string of the given width.
// If values is shorter than width, pad with the last value.
// If longer, downsample by averaging.
// Returns empty string for empty values.
func Render(values []float64, width int) string {
	if len(values) == 0 || width <= 0 {
		return ""
	}

	data := resample(values, width)

	minVal, maxVal := data[0], data[0]
	for _, v := range data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	var sb strings.Builder
	numChars := len(sparkRunes)
	for _, v := range data {
		idx := mapIndex(v, minVal, maxVal, numChars)
		sb.WriteRune(sparkRunes[idx])
	}
	return sb.String()
}

// RenderHeatmap renders a heatmap string using HeatChars (5 levels).
// Values of 0 map to space " ".
func RenderHeatmap(values []float64, width int) string {
	if len(values) == 0 || width <= 0 {
		return ""
	}

	data := resample(values, width)

	// Find max of non-zero values for scaling
	maxVal := 0.0
	for _, v := range data {
		if v > maxVal {
			maxVal = v
		}
	}

	var sb strings.Builder
	for _, v := range data {
		if v == 0 {
			sb.WriteRune(heatRunes[0]) // space
			continue
		}
		if maxVal == 0 {
			sb.WriteRune(heatRunes[0])
			continue
		}
		// Map non-zero values to indices 1..4 (skip the space at index 0)
		idx := int(math.Round(v / maxVal * 4))
		if idx < 1 {
			idx = 1
		}
		if idx > 4 {
			idx = 4
		}
		sb.WriteRune(heatRunes[idx])
	}
	return sb.String()
}

// TrendArrow returns "↑" if upward trend, "↓" if downward, "→" if flat.
// Compares average of last 3 values to average of previous 3.
// Uses a 10% threshold for flat.
func TrendArrow(values []float64) string {
	if len(values) < 6 {
		return "→"
	}

	n := len(values)
	recent := avg(values[n-3:])
	previous := avg(values[n-6 : n-3])

	if previous == 0 {
		if recent > 0 {
			return "↑"
		}
		return "→"
	}

	change := (recent - previous) / math.Abs(previous)
	if change > 0.10 {
		return "↑"
	}
	if change < -0.10 {
		return "↓"
	}
	return "→"
}

func avg(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func resample(values []float64, width int) []float64 {
	n := len(values)
	if n == width {
		out := make([]float64, n)
		copy(out, values)
		return out
	}
	if n < width {
		// Pad with last value
		out := make([]float64, width)
		copy(out, values)
		last := values[n-1]
		for i := n; i < width; i++ {
			out[i] = last
		}
		return out
	}
	// Downsample by averaging
	out := make([]float64, width)
	for i := 0; i < width; i++ {
		start := i * n / width
		end := (i + 1) * n / width
		if end > n {
			end = n
		}
		sum := 0.0
		for j := start; j < end; j++ {
			sum += values[j]
		}
		out[i] = sum / float64(end-start)
	}
	return out
}

func mapIndex(val, minVal, maxVal float64, numChars int) int {
	if maxVal == minVal {
		return numChars / 2
	}
	idx := int((val - minVal) / (maxVal - minVal) * float64(numChars-1))
	if idx < 0 {
		idx = 0
	}
	if idx >= numChars {
		idx = numChars - 1
	}
	return idx
}
