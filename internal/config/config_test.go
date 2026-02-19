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

	if cfg.DefaultFilter != "" {
		t.Errorf("expected DefaultFilter to be empty (defers to TUI default), got '%s'", cfg.DefaultFilter)
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
	if cfg.DefaultFilter != "" {
		t.Errorf("expected empty DefaultFilter for missing file, got '%s'", cfg.DefaultFilter)
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
	if cfg.DefaultFilter != "" {
		t.Errorf("expected empty DefaultFilter for partial config, got '%s'", cfg.DefaultFilter)
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
	if cfg.DefaultFilter != "" {
		t.Errorf("expected empty DefaultFilter for empty path, got '%s'", cfg.DefaultFilter)
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

func TestAnimationsEnabled_DefaultTrue(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.AnimationsEnabled() {
		t.Error("expected animations enabled by default")
	}
}

func TestAnimationsEnabled_ExplicitTrue(t *testing.T) {
	b := true
	cfg := &Config{Animations: &b}
	if !cfg.AnimationsEnabled() {
		t.Error("expected animations enabled when explicitly true")
	}
}

func TestAnimationsEnabled_ExplicitFalse(t *testing.T) {
	b := false
	cfg := &Config{Animations: &b}
	if cfg.AnimationsEnabled() {
		t.Error("expected animations disabled when explicitly false")
	}
}

func TestLoad_AnimationsFalse(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "anim-config.yml")

	configContent := "animations: false\n"
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.AnimationsEnabled() {
		t.Error("expected animations disabled from config file")
	}
}

func TestAsciiHeaderEnabled_DefaultTrue(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.AsciiHeaderEnabled() {
		t.Error("expected ascii header enabled by default")
	}
}

func TestAsciiHeaderEnabled_ExplicitTrue(t *testing.T) {
	b := true
	cfg := &Config{AsciiHeader: &b}
	if !cfg.AsciiHeaderEnabled() {
		t.Error("expected ascii header enabled when explicitly true")
	}
}

func TestAsciiHeaderEnabled_ExplicitFalse(t *testing.T) {
	b := false
	cfg := &Config{AsciiHeader: &b}
	if cfg.AsciiHeaderEnabled() {
		t.Error("expected ascii header disabled when explicitly false")
	}
}

func TestLoad_AsciiHeaderFalse(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "ascii-config.yml")

	configContent := "asciiHeader: false\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.AsciiHeaderEnabled() {
		t.Error("expected ascii header disabled from config file")
	}
}

func TestLoad_ThemeField(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "theme-config.yml")

	configContent := "theme: dark\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Theme != "dark" {
		t.Errorf("expected theme 'dark', got %q", cfg.Theme)
	}
}

func TestDefaultConfig_FieldValues(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Animations != nil {
		t.Error("expected Animations to be nil by default")
	}
	if cfg.AsciiHeader != nil {
		t.Error("expected AsciiHeader to be nil by default")
	}
	if cfg.Theme != "" {
		t.Errorf("expected empty Theme by default, got %q", cfg.Theme)
	}
}
