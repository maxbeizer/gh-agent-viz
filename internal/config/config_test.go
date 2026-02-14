package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.RefreshInterval != 30 {
		t.Errorf("expected RefreshInterval to be 30, got %d", cfg.RefreshInterval)
	}

	if cfg.DefaultFilter != "all" {
		t.Errorf("expected DefaultFilter to be 'all', got '%s'", cfg.DefaultFilter)
	}

	if len(cfg.Repos) != 0 {
		t.Errorf("expected empty Repos list, got %d items", len(cfg.Repos))
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yml")

	configContent := `repos:
  - owner/repo1
  - owner/repo2
refreshInterval: 60
defaultFilter: running
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if len(cfg.Repos) != 2 {
		t.Errorf("expected 2 repos, got %d", len(cfg.Repos))
	}
	if cfg.Repos[0] != "owner/repo1" {
		t.Errorf("expected first repo to be 'owner/repo1', got '%s'", cfg.Repos[0])
	}
	if cfg.Repos[1] != "owner/repo2" {
		t.Errorf("expected second repo to be 'owner/repo2', got '%s'", cfg.Repos[1])
	}

	if cfg.RefreshInterval != 60 {
		t.Errorf("expected RefreshInterval to be 60, got %d", cfg.RefreshInterval)
	}

	if cfg.DefaultFilter != "running" {
		t.Errorf("expected DefaultFilter to be 'running', got '%s'", cfg.DefaultFilter)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "does-not-exist.yml")

	cfg, err := Load(nonExistentPath)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}

	// Should return default config
	if cfg.RefreshInterval != 30 {
		t.Errorf("expected default RefreshInterval of 30, got %d", cfg.RefreshInterval)
	}
	if cfg.DefaultFilter != "all" {
		t.Errorf("expected default DefaultFilter of 'all', got '%s'", cfg.DefaultFilter)
	}
	if len(cfg.Repos) != 0 {
		t.Errorf("expected empty Repos list, got %d items", len(cfg.Repos))
	}
}

func TestLoad_PartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "partial-config.yml")

	// Only specify repos, other fields should use defaults
	configContent := `repos:
  - owner/repo1
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	// Specified field should be loaded
	if len(cfg.Repos) != 1 {
		t.Errorf("expected 1 repo, got %d", len(cfg.Repos))
	}
	if cfg.Repos[0] != "owner/repo1" {
		t.Errorf("expected repo to be 'owner/repo1', got '%s'", cfg.Repos[0])
	}

	// Unspecified fields should have default values
	if cfg.RefreshInterval != 30 {
		t.Errorf("expected default RefreshInterval of 30, got %d", cfg.RefreshInterval)
	}
	if cfg.DefaultFilter != "all" {
		t.Errorf("expected default DefaultFilter of 'all', got '%s'", cfg.DefaultFilter)
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	// When path is empty, it should try default location and return defaults if not found
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error with empty path: %v", err)
	}

	// Should return default values
	if cfg.RefreshInterval != 30 {
		t.Errorf("expected default RefreshInterval of 30, got %d", cfg.RefreshInterval)
	}
	if cfg.DefaultFilter != "all" {
		t.Errorf("expected default DefaultFilter of 'all', got '%s'", cfg.DefaultFilter)
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "save-test.yml")

	cfg := &Config{
		Repos:           []string{"owner/repo1", "owner/repo2"},
		RefreshInterval: 45,
		DefaultFilter:   "completed",
	}

	err := Save(cfg, configPath)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Load it back and verify
	loadedCfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if len(loadedCfg.Repos) != 2 {
		t.Errorf("expected 2 repos after save/load, got %d", len(loadedCfg.Repos))
	}
	if loadedCfg.RefreshInterval != 45 {
		t.Errorf("expected RefreshInterval of 45 after save/load, got %d", loadedCfg.RefreshInterval)
	}
	if loadedCfg.DefaultFilter != "completed" {
		t.Errorf("expected DefaultFilter of 'completed' after save/load, got '%s'", loadedCfg.DefaultFilter)
	}
}
