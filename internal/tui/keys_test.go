package tui

import "testing"

func TestKeybindings_OpenRepoExists(t *testing.T) {
	kb := NewKeybindings()
	keys := kb.OpenRepo.Keys()
	if len(keys) == 0 {
		t.Fatal("expected OpenRepo to have keys")
	}
	if keys[0] != "!" {
		t.Fatalf("expected OpenRepo key to be '!', got %q", keys[0])
	}
}

func TestKeybindings_FileIssueExists(t *testing.T) {
	kb := NewKeybindings()
	keys := kb.FileIssue.Keys()
	if len(keys) == 0 {
		t.Fatal("expected FileIssue to have keys")
	}
	if keys[0] != "@" {
		t.Fatalf("expected FileIssue key to be '@', got %q", keys[0])
	}
}
