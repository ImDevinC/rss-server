package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	BaseURL string `yaml:"base_url"`
	Server  struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	Upload struct {
		MaxFileSizeMB     int      `yaml:"max_file_size_mb"`
		AllowedExtensions []string `yaml:"allowed_extensions"`
	} `yaml:"upload"`
	Paths struct {
		DataDir    string `yaml:"data_dir"`
		AudioDir   string `yaml:"audio_dir"`
		ArtworkDir string `yaml:"artwork_dir"`
		RSSFile    string `yaml:"rss_file"`
	} `yaml:"paths"`
	Podcast struct {
		DefaultTitle       string `yaml:"default_title"`
		DefaultAuthor      string `yaml:"default_author"`
		DefaultDescription string `yaml:"default_description"`
		DefaultLanguage    string `yaml:"default_language"`
		DefaultExplicit    string `yaml:"default_explicit"`
		DefaultCategory    string `yaml:"default_category"`
	} `yaml:"podcast"`
}

// Load reads and parses the configuration file
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// Validate checks that the configuration is valid
func (c *Config) Validate() error {
	// Validate base_url is present
	if c.BaseURL == "" {
		return fmt.Errorf("base_url is required in configuration")
	}

	// Validate base_url format
	parsedURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("base_url is invalid: %w", err)
	}

	// Ensure scheme is http or https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("base_url must use http or https scheme, got: %s", parsedURL.Scheme)
	}

	// Ensure host is present
	if parsedURL.Host == "" {
		return fmt.Errorf("base_url must include a host (e.g., http://example.com)")
	}

	// Normalize base_url by removing trailing slash
	c.BaseURL = strings.TrimRight(c.BaseURL, "/")

	return nil
}

// GetBaseURL returns the validated base URL
func (c *Config) GetBaseURL() string {
	return c.BaseURL
}
