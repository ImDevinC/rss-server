package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/rss-server/internal/config"
)

// T012: Valid HTTP base URL passes validation
func TestValidHTTPBaseURL(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://example.com:8080",
	}
	cfg.Server.Port = "8080"
	cfg.Server.Host = "0.0.0.0"
	cfg.Paths.DataDir = "./data"
	cfg.Paths.AudioDir = "./data/audio"
	cfg.Paths.ArtworkDir = "./data/artwork"
	cfg.Paths.RSSFile = "./data/podcast.xml"
	cfg.Upload.MaxFileSizeMB = 500

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Expected valid HTTP URL to pass validation, got error: %v", err)
	}
}

// T013: Valid HTTPS base URL passes validation
func TestValidHTTPSBaseURL(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "https://example.com",
	}
	cfg.Server.Port = "8080"
	cfg.Server.Host = "0.0.0.0"
	cfg.Paths.DataDir = "./data"
	cfg.Paths.AudioDir = "./data/audio"
	cfg.Paths.ArtworkDir = "./data/artwork"
	cfg.Paths.RSSFile = "./data/podcast.xml"
	cfg.Upload.MaxFileSizeMB = 500

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Expected valid HTTPS URL to pass validation, got error: %v", err)
	}
}

// T014: Missing base URL returns error
func TestMissingBaseURL(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "",
	}
	cfg.Server.Port = "8080"
	cfg.Server.Host = "0.0.0.0"
	cfg.Paths.DataDir = "./data"
	cfg.Paths.AudioDir = "./data/audio"
	cfg.Paths.ArtworkDir = "./data/artwork"
	cfg.Paths.RSSFile = "./data/podcast.xml"
	cfg.Upload.MaxFileSizeMB = 500

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for missing base URL, got nil")
	}
}

// T015: Invalid scheme (ftp://) returns error
func TestInvalidScheme(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "ftp://example.com",
	}
	cfg.Server.Port = "8080"
	cfg.Server.Host = "0.0.0.0"
	cfg.Paths.DataDir = "./data"
	cfg.Paths.AudioDir = "./data/audio"
	cfg.Paths.ArtworkDir = "./data/artwork"
	cfg.Paths.RSSFile = "./data/podcast.xml"
	cfg.Upload.MaxFileSizeMB = 500

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for invalid scheme (ftp://), got nil")
	}
}

// T016: Missing hostname returns error
func TestMissingHostname(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://",
	}
	cfg.Server.Port = "8080"
	cfg.Server.Host = "0.0.0.0"
	cfg.Paths.DataDir = "./data"
	cfg.Paths.AudioDir = "./data/audio"
	cfg.Paths.ArtworkDir = "./data/artwork"
	cfg.Paths.RSSFile = "./data/podcast.xml"
	cfg.Upload.MaxFileSizeMB = 500

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for missing hostname, got nil")
	}
}

// T017: Trailing slash normalization works
func TestTrailingSlashNormalization(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://example.com:8080/",
	}
	cfg.Server.Port = "8080"
	cfg.Server.Host = "0.0.0.0"
	cfg.Paths.DataDir = "./data"
	cfg.Paths.AudioDir = "./data/audio"
	cfg.Paths.ArtworkDir = "./data/artwork"
	cfg.Paths.RSSFile = "./data/podcast.xml"
	cfg.Upload.MaxFileSizeMB = 500

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Expected URL with trailing slash to be valid, got error: %v", err)
	}

	// Verify trailing slash is removed
	normalized := cfg.GetBaseURL()
	if normalized == "http://example.com:8080/" {
		t.Errorf("Expected trailing slash to be removed, got: %s", normalized)
	}
	if normalized != "http://example.com:8080" {
		t.Errorf("Expected normalized URL 'http://example.com:8080', got: %s", normalized)
	}
}

// T018: Integration test - Load config from file
func TestLoadConfigFromFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `base_url: "http://localhost:8080"
server:
  port: 8080
  host: "0.0.0.0"
upload:
  max_file_size_mb: 500
paths:
  data_dir: "./data"
  audio_dir: "./data/audio"
  artwork_dir: "./data/artwork"
  rss_file: "./data/podcast.xml"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	// Load and validate
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	// Verify values
	if cfg.GetBaseURL() != "http://localhost:8080" {
		t.Errorf("Expected base URL 'http://localhost:8080', got: %s", cfg.GetBaseURL())
	}
}
