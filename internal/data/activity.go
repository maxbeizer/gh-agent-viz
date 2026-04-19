package data

import "time"

// HourlyActivity buckets sessions by hour-of-day (0-23) based on CreatedAt.
// Returns a [24]int array where index 0 = midnight, 23 = 11pm.
func HourlyActivity(sessions []Session) [24]int {
	var buckets [24]int
	for _, s := range sessions {
		if !s.CreatedAt.IsZero() {
			buckets[s.CreatedAt.Hour()]++
		}
	}
	return buckets
}

// DailySessionCounts returns session counts per day for the last N days.
// Index 0 = N days ago, index N-1 = today. Uses CreatedAt.
func DailySessionCounts(sessions []Session, days int) []int {
	counts := make([]int, days)
	now := time.Now()
	for _, s := range sessions {
		if s.CreatedAt.IsZero() {
			continue
		}
		daysAgo := int(now.Sub(s.CreatedAt).Hours() / 24)
		idx := days - 1 - daysAgo
		if idx >= 0 && idx < days {
			counts[idx]++
		}
	}
	return counts
}
