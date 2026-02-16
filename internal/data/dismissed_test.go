package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDismissedStore_AddAndIDs(t *testing.T) {
	dir := t.TempDir()
	s := &DismissedStore{
		ids:  map[string]struct{}{},
		path: filepath.Join(dir, "dismissed.json"),
	}

	s.Add("abc")
	s.Add("def")

	ids := s.IDs()
	if _, ok := ids["abc"]; !ok {
		t.Error("expected abc in dismissed IDs")
	}
	if _, ok := ids["def"]; !ok {
		t.Error("expected def in dismissed IDs")
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 IDs, got %d", len(ids))
	}
}

func TestDismissedStore_AddEmpty(t *testing.T) {
	dir := t.TempDir()
	s := &DismissedStore{
		ids:  map[string]struct{}{},
		path: filepath.Join(dir, "dismissed.json"),
	}

	s.Add("")
	if len(s.IDs()) != 0 {
		t.Error("empty ID should not be added")
	}
}

func TestDismissedStore_PersistsAcrossReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dismissed.json")

	s1 := &DismissedStore{
		ids:  map[string]struct{}{},
		path: path,
	}
	s1.Add("session-1")
	s1.Add("session-2")

	// Load fresh from same file
	s2 := &DismissedStore{
		ids:  map[string]struct{}{},
		path: path,
	}
	s2.load()

	ids := s2.IDs()
	if len(ids) != 2 {
		t.Fatalf("expected 2 IDs after reload, got %d", len(ids))
	}
	if _, ok := ids["session-1"]; !ok {
		t.Error("expected session-1 after reload")
	}
	if _, ok := ids["session-2"]; !ok {
		t.Error("expected session-2 after reload")
	}
}

func TestDismissedStore_MissingFile(t *testing.T) {
	s := &DismissedStore{
		ids:  map[string]struct{}{},
		path: "/nonexistent/path/dismissed.json",
	}
	s.load()
	if len(s.IDs()) != 0 {
		t.Error("expected empty set when file is missing")
	}
}

func TestDismissedStore_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dismissed.json")
	_ = os.WriteFile(path, []byte("not json!!!"), 0600)

	s := &DismissedStore{
		ids:  map[string]struct{}{},
		path: path,
	}
	s.load()
	if len(s.IDs()) != 0 {
		t.Error("expected empty set when file is corrupt")
	}
}

func TestDismissedStore_Deduplication(t *testing.T) {
	dir := t.TempDir()
	s := &DismissedStore{
		ids:  map[string]struct{}{},
		path: filepath.Join(dir, "dismissed.json"),
	}

	s.Add("same-id")
	s.Add("same-id")
	s.Add("same-id")

	if len(s.IDs()) != 1 {
		t.Errorf("expected 1 ID after dedup, got %d", len(s.IDs()))
	}
}

func TestDismissedStore_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dismissed.json")

	s := &DismissedStore{
		ids:  map[string]struct{}{},
		path: path,
	}
	s.Add("test")

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600 permissions, got %o", info.Mode().Perm())
	}
}

func TestDismissedStore_FileFormatIsJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dismissed.json")

	s := &DismissedStore{
		ids:  map[string]struct{}{},
		path: path,
	}
	s.Add("id-1")

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	var ids []string
	if err := json.Unmarshal(raw, &ids); err != nil {
		t.Fatalf("file is not valid JSON: %v", err)
	}
	if len(ids) != 1 || ids[0] != "id-1" {
		t.Errorf("unexpected file content: %v", ids)
	}
}
