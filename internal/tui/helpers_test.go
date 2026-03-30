package tui

import (
	"fmt"
	"testing"
	"time"

	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

func TestSessionFingerprint_Deterministic(t *testing.T) {
	sessions := []data.Session{
		{ID: "a", Status: "running", UpdatedAt: time.Unix(1000, 0)},
		{ID: "b", Status: "completed", UpdatedAt: time.Unix(2000, 0)},
	}
	fp1 := sessionFingerprint(sessions)
	fp2 := sessionFingerprint(sessions)
	if fp1 != fp2 {
		t.Fatalf("same input should produce same fingerprint: %s != %s", fp1, fp2)
	}
}

func TestSessionFingerprint_ChangesOnStatusUpdate(t *testing.T) {
	sessions := []data.Session{
		{ID: "a", Status: "running", UpdatedAt: time.Unix(1000, 0)},
	}
	fp1 := sessionFingerprint(sessions)

	sessions[0].Status = "completed"
	fp2 := sessionFingerprint(sessions)
	if fp1 == fp2 {
		t.Fatal("fingerprint should change when status changes")
	}
}

func TestSessionFingerprint_ChangesOnTimeUpdate(t *testing.T) {
	sessions := []data.Session{
		{ID: "a", Status: "running", UpdatedAt: time.Unix(1000, 0)},
	}
	fp1 := sessionFingerprint(sessions)

	sessions[0].UpdatedAt = time.Unix(2000, 0)
	fp2 := sessionFingerprint(sessions)
	if fp1 == fp2 {
		t.Fatal("fingerprint should change when UpdatedAt changes")
	}
}

func TestSessionFingerprint_EmptySessions(t *testing.T) {
	fp := sessionFingerprint(nil)
	if fp == "" {
		t.Fatal("fingerprint of empty sessions should not be empty")
	}
}

func TestSessionFingerprint_SameDataDifferentOrder(t *testing.T) {
	a := []data.Session{
		{ID: "a", Status: "running", UpdatedAt: time.Unix(1000, 0)},
		{ID: "b", Status: "completed", UpdatedAt: time.Unix(2000, 0)},
	}
	b := []data.Session{
		{ID: "b", Status: "completed", UpdatedAt: time.Unix(2000, 0)},
		{ID: "a", Status: "running", UpdatedAt: time.Unix(1000, 0)},
	}
	fp1 := sessionFingerprint(a)
	fp2 := sessionFingerprint(b)
	// Order matters — different order produces different fingerprint.
	// This is fine because session order is deterministic.
	if fp1 == fp2 {
		t.Fatal("different ordering should produce different fingerprints")
	}
}

func TestMergeSessions_CapsAtMaxSessions(t *testing.T) {
	m := &Model{ctx: NewProgramContext()}
	// Generate more sessions than the cap
	sessions := make([]data.Session, maxSessions+100)
	for i := range sessions {
		sessions[i] = data.Session{
			ID:        fmt.Sprintf("session-%d", i),
			Status:    "completed",
			UpdatedAt: time.Unix(int64(i), 0),
		}
	}
	m.mergeSessions(sessions)

	if len(m.allSessions) != maxSessions {
		t.Fatalf("expected %d sessions after cap, got %d", maxSessions, len(m.allSessions))
	}

	// Verify the newest sessions were kept (highest UpdatedAt)
	for _, s := range m.allSessions {
		if s.UpdatedAt.Unix() < int64(100) {
			t.Fatalf("oldest sessions should have been trimmed, found UpdatedAt=%d", s.UpdatedAt.Unix())
		}
	}
}

func TestMergeSessions_BelowCapUnchanged(t *testing.T) {
	m := &Model{ctx: NewProgramContext()}
	sessions := []data.Session{
		{ID: "a", Status: "running", UpdatedAt: time.Unix(1000, 0)},
		{ID: "b", Status: "completed", UpdatedAt: time.Unix(2000, 0)},
	}
	m.mergeSessions(sessions)

	if len(m.allSessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(m.allSessions))
	}
}

func TestMergeSessions_AutoUndismissUrgent(t *testing.T) {
	store := data.NewDismissedStoreFromPath(t.TempDir() + "/dismissed.json")
	store.Add("urgent-session")
	store.Add("normal-session")

	m := &Model{
		ctx:            NewProgramContext(),
		dismissedStore: store,
		prevSessions:   map[string]string{},
	}

	sessions := []data.Session{
		{ID: "urgent-session", Status: "failed", UpdatedAt: time.Unix(3000, 0)},
		{ID: "normal-session", Status: "completed", UpdatedAt: time.Unix(2000, 0)},
		{ID: "visible-session", Status: "running", UpdatedAt: time.Unix(1000, 0)},
	}
	m.mergeSessions(sessions)

	// The urgent (failed) session should have been un-dismissed and be visible
	ids := store.IDs()
	if _, ok := ids["urgent-session"]; ok {
		t.Error("expected urgent-session to be auto-undismissed")
	}
	// The normal completed session should remain dismissed
	if _, ok := ids["normal-session"]; !ok {
		t.Error("expected normal-session to remain dismissed")
	}

	// Check that the urgent session is in the visible set (ctx.Counts should include it)
	if m.ctx.Counts.Attention != 1 {
		t.Errorf("expected 1 attention count, got %d", m.ctx.Counts.Attention)
	}
	if m.ctx.Counts.All != 2 { // urgent + visible (normal is still dismissed)
		t.Errorf("expected 2 visible sessions, got %d", m.ctx.Counts.All)
	}
}
