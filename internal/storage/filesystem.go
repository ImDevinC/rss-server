package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/example/rss-server/internal/models"
)

// SaveAudioFile saves an uploaded audio file to the filesystem with a unique filename
func SaveAudioFile(originalName string, data []byte, audioDir string) (*models.AudioFile, error) {
	// Generate unique filename
	filename := GenerateUniqueFilename(originalName)
	filePath := filepath.Join(audioDir, filename)

	// Ensure directory exists
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audio directory: %w", err)
	}

	// Write file to disk
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write audio file: %w", err)
	}

	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat audio file: %w", err)
	}

	return &models.AudioFile{
		Filename:     filename,
		OriginalName: originalName,
		FilePath:     filePath,
		Size:         info.Size(),
		MimeType:     "audio/mpeg", // Assuming MP3
		UploadDate:   time.Now(),
	}, nil
}

// DeleteAudioFile removes an audio file from the filesystem
func DeleteAudioFile(filename string, audioDir string) error {
	filePath := filepath.Join(audioDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("audio file not found: %s", filename)
	}

	// Delete file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete audio file: %w", err)
	}

	return nil
}

// GenerateUniqueFilename creates a unique filename from the original name
func GenerateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	base := strings.TrimSuffix(originalName, ext)

	// Sanitize base name (remove non-alphanumeric characters)
	sanitized := regexp.MustCompile(`[^a-zA-Z0-9-_]`).ReplaceAllString(base, "-")
	sanitized = strings.Trim(sanitized, "-")

	// Limit length
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}

	// Add timestamp for uniqueness
	timestamp := time.Now().Format("20060102-150405")

	return fmt.Sprintf("%s-%s%s", sanitized, timestamp, ext)
}

// SaveArtworkFile saves an uploaded artwork file to the filesystem
func SaveArtworkFile(originalName string, data []byte, artworkDir string) (string, error) {
	// Generate unique filename
	filename := GenerateUniqueFilename(originalName)
	filePath := filepath.Join(artworkDir, filename)

	// Ensure directory exists
	if err := os.MkdirAll(artworkDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create artwork directory: %w", err)
	}

	// Write file to disk
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write artwork file: %w", err)
	}

	return filename, nil
}
