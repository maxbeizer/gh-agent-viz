package sparkline

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		width  int
		check  func(string) bool
		desc   string
	}{
		{
			name:   "empty values",
			values: []float64{},
			width:  10,
			check:  func(s string) bool { return s == "" },
			desc:   "should return empty string",
		},
		{
			name:   "uniform values",
			values: []float64{5, 5, 5, 5, 5},
			width:  5,
			check: func(s string) bool {
				// All chars should be the same
				runes := []rune(s)
				for _, r := range runes {
					if r != runes[0] {
						return false
					}
				}
				return len(runes) == 5
			},
			desc: "all chars should be identical for uniform values",
		},
		{
			name:   "ascending values",
			values: []float64{1, 2, 3, 4, 5, 6, 7, 8},
			width:  8,
			check: func(s string) bool {
				runes := []rune(s)
				if len(runes) != 8 {
					return false
				}
				// First char should be lowest, last should be highest
				return runes[0] == []rune(SparkChars)[0] && runes[7] == []rune(SparkChars)[7]
			},
			desc: "should go from lowest to highest block",
		},
		{
			name:   "downsampling more values than width",
			values: []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			width:  4,
			check: func(s string) bool {
				runes := []rune(s)
				return len(runes) == 4
			},
			desc: "output should have exactly width characters",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.values, tt.width)
			if !tt.check(got) {
				t.Errorf("Render() = %q, %s", got, tt.desc)
			}
		})
	}
}

func TestTrendArrow(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   string
	}{
		{
			name:   "upward trend",
			values: []float64{1, 2, 3, 10, 20, 30},
			want:   "↑",
		},
		{
			name:   "downward trend",
			values: []float64{30, 20, 10, 3, 2, 1},
			want:   "↓",
		},
		{
			name:   "flat trend",
			values: []float64{10, 10, 10, 10, 10, 10},
			want:   "→",
		},
		{
			name:   "too few values",
			values: []float64{1, 2, 3},
			want:   "→",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrendArrow(tt.values)
			if got != tt.want {
				t.Errorf("TrendArrow() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderHeatmap(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		width  int
		check  func(string) bool
		desc   string
	}{
		{
			name:   "empty values",
			values: []float64{},
			width:  5,
			check:  func(s string) bool { return s == "" },
			desc:   "should return empty string",
		},
		{
			name:   "mixed values with zeros",
			values: []float64{0, 5, 0, 10, 0},
			width:  5,
			check: func(s string) bool {
				runes := []rune(s)
				if len(runes) != 5 {
					return false
				}
				// Zero values should be space
				if runes[0] != ' ' || runes[2] != ' ' || runes[4] != ' ' {
					return false
				}
				// Non-zero values should not be space
				if runes[1] == ' ' || runes[3] == ' ' {
					return false
				}
				return true
			},
			desc: "zeros should map to space, non-zeros to block chars",
		},
		{
			name:   "all zeros",
			values: []float64{0, 0, 0},
			width:  3,
			check: func(s string) bool {
				return s == strings.Repeat(" ", 3)
			},
			desc: "all zeros should be spaces",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderHeatmap(tt.values, tt.width)
			if !tt.check(got) {
				t.Errorf("RenderHeatmap() = %q, %s", got, tt.desc)
			}
		})
	}
}
