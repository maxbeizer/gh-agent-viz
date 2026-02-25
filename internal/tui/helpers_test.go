package tui

import (
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
