package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

const dismissedFileName = ".gh-agent-viz-dismissed.json"

// DismissedStore manages a persistent set of dismissed session IDs.
type DismissedStore struct {
	mu   sync.Mutex
	ids  map[string]struct{}
	path string
}

// NewDismissedStore loads dismissed IDs from the default file.
// If the file is missing or corrupt, starts with an empty set.
func NewDismissedStore() *DismissedStore {
	path := dismissedFilePath()
	s := &DismissedStore{
		ids:  map[string]struct{}{},
		path: path,
	}
	s.load()
	return s
}

// IDs returns a copy of the dismissed ID set.
func (s *DismissedStore) IDs() map[string]struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]struct{}, len(s.ids))
	for id := range s.ids {
		out[id] = struct{}{}
	}
	return out
}

// Add marks a session ID as dismissed and persists to disk.
func (s *DismissedStore) Add(id string) {
	if id == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ids[id] = struct{}{}
	s.save()
}

func (s *DismissedStore) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	var ids []string
	if err := json.Unmarshal(data, &ids); err != nil {
		return
	}
	for _, id := range ids {
		s.ids[id] = struct{}{}
	}
}

func (s *DismissedStore) save() {
	ids := make([]string, 0, len(s.ids))
	for id := range s.ids {
		ids = append(ids, id)
	}
	data, err := json.Marshal(ids)
	if err != nil {
		return
	}
	_ = os.WriteFile(s.path, data, 0600)
}

func dismissedFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return dismissedFileName
	}
	return filepath.Join(home, dismissedFileName)
}
