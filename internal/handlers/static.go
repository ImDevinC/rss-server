package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// StaticHandler handles serving audio files
type StaticHandler struct {
	audioDir string
}

// NewStaticHandler creates a new static file handler
func NewStaticHandler(audioDir string) *StaticHandler {
	return &StaticHandler{audioDir: audioDir}
}

// HandleAudio handles GET /audio/{filename}
func (h *StaticHandler) HandleAudio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract filename from URL path
	filename := strings.TrimPrefix(r.URL.Path, "/audio/")
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	// Prevent directory traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// Build file path
	filePath := filepath.Join(h.audioDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Audio file not found", http.StatusNotFound)
		return
	}

	// Serve file with correct content type
	w.Header().Set("Content-Type", "audio/mpeg")
	http.ServeFile(w, r, filePath)
}
