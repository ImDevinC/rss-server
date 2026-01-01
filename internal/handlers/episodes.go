package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/example/rss-server/internal/models"
	"github.com/example/rss-server/internal/storage"
)

// EpisodesHandler handles episode-related requests
type EpisodesHandler struct {
	store        *storage.RSSStore
	audioDir     string
	artworkDir   string
	maxSizeMB    int64
	maxArtworkMB int64
	templates    *template.Template
}

// NewEpisodesHandler creates a new episodes handler
func NewEpisodesHandler(store *storage.RSSStore, audioDir string, artworkDir string, maxSizeMB int64, templates *template.Template) *EpisodesHandler {
	return &EpisodesHandler{
		store:        store,
		audioDir:     audioDir,
		artworkDir:   artworkDir,
		maxSizeMB:    maxSizeMB,
		maxArtworkMB: 5, // 5MB limit for artwork
		templates:    templates,
	}
}

// HandleUpload handles POST /api/episodes
func (h *EpisodesHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (limit: maxSizeMB)
	maxSize := h.maxSizeMB * 1024 * 1024
	if err := r.ParseMultipartForm(maxSize); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	// Get audio file
	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "Audio file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension (.mp3 only)
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".mp3") {
		http.Error(w, "Only MP3 files are supported", http.StatusUnsupportedMediaType)
		return
	}

	// Validate file size
	if header.Size > maxSize {
		http.Error(w, fmt.Sprintf("File too large (max %d MB)", h.maxSizeMB), http.StatusRequestEntityTooLarge)
		return
	}

	// Get form fields
	title := r.FormValue("title")
	description := r.FormValue("description")

	if title == "" || description == "" {
		http.Error(w, "Title and description required", http.StatusBadRequest)
		return
	}

	// Read audio file data
	audioData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read audio file: %v", err), http.StatusInternalServerError)
		return
	}

	// Save audio file
	audioFile, err := storage.SaveAudioFile(header.Filename, audioData, h.audioDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save audio file: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse publication date (optional)
	pubDate := time.Now()
	if pubDateStr := r.FormValue("pubDate"); pubDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, pubDateStr); err == nil {
			pubDate = parsed
		}
	}

	// Generate episode ID
	episodeID := GenerateEpisodeID(title, pubDate)

	// Build audio URL (relative to server)
	audioURL := fmt.Sprintf("/audio/%s", audioFile.Filename)

	// Create episode
	episode := models.Episode{
		ID:          episodeID,
		Title:       title,
		Description: description,
		PubDate:     pubDate,
		GUID:        episodeID,
		AudioURL:    audioURL,
		AudioLength: audioFile.Size,
		AudioType:   "audio/mpeg",
		Duration:    audioFile.Duration,
		Explicit:    r.FormValue("explicit"),
		Filename:    audioFile.Filename,
		UploadDate:  audioFile.UploadDate,
	}

	// Parse optional fields
	if epNumStr := r.FormValue("episodeNumber"); epNumStr != "" {
		fmt.Sscanf(epNumStr, "%d", &episode.EpisodeNum)
	}
	if seasonNumStr := r.FormValue("seasonNumber"); seasonNumStr != "" {
		fmt.Sscanf(seasonNumStr, "%d", &episode.SeasonNum)
	}
	if epType := r.FormValue("episodeType"); epType != "" {
		episode.EpisodeType = epType
	}

	// Add episode to store
	if err := h.store.AddEpisode(episode); err != nil {
		// Cleanup: delete audio file if episode creation fails
		os.Remove(filepath.Join(h.audioDir, audioFile.Filename))
		http.Error(w, fmt.Sprintf("Failed to add episode: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(episode)
}

// HandleList handles GET /api/episodes
func (h *EpisodesHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	podcast := h.store.GetPodcast()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(podcast.Episodes)
}

// GenerateEpisodeID generates a unique episode ID from title and date
func GenerateEpisodeID(title string, pubDate time.Time) string {
	// Sanitize title for URL safety
	sanitized := strings.ToLower(title)
	sanitized = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(sanitized, "-")
	sanitized = strings.Trim(sanitized, "-")

	// Limit length
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}

	// Combine date + title
	dateStr := pubDate.Format("20060102")
	id := fmt.Sprintf("ep-%s-%s", dateStr, sanitized)

	return id
}

// HandleDelete handles DELETE /api/episodes/{id}
func (h *EpisodesHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract episode ID from URL path
	// Expected format: /api/episodes/{episodeId}
	path := r.URL.Path
	episodeID := strings.TrimPrefix(path, "/api/episodes/")

	if episodeID == "" || episodeID == path {
		http.Error(w, "Episode ID required", http.StatusBadRequest)
		return
	}

	// Get episode to find the audio filename before deleting
	podcast := h.store.GetPodcast()
	var audioFilename string
	found := false

	for _, ep := range podcast.Episodes {
		if ep.ID == episodeID {
			audioFilename = ep.Filename
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Episode not found", http.StatusNotFound)
		return
	}

	// Delete episode from RSS store first
	if err := h.store.DeleteEpisode(episodeID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete episode: %v", err), http.StatusInternalServerError)
		return
	}

	// Delete audio file from filesystem
	if audioFilename != "" {
		if err := storage.DeleteAudioFile(audioFilename, h.audioDir); err != nil {
			// Log error but don't fail the request since episode is already removed from RSS
			fmt.Printf("Warning: Failed to delete audio file %s: %v\n", audioFilename, err)
		}
	}

	// Return success (for HTMX - empty response to remove element)
	w.WriteHeader(http.StatusOK)
}

// HandleGetSettings handles GET /api/podcast/settings
func (h *EpisodesHandler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	podcast := h.store.GetPodcast()

	// Render settings form template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "settings_form.html", podcast); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}
}

// HandleUpdateSettings handles POST /api/podcast/settings
func (h *EpisodesHandler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (for artwork upload)
	maxSize := h.maxArtworkMB * 1024 * 1024
	if err := r.ParseMultipartForm(maxSize); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	// Get current podcast to preserve episodes
	currentPodcast := h.store.GetPodcast()

	// Build updated podcast from form data
	podcast := &models.Podcast{
		Title:       r.FormValue("title"),
		Link:        r.FormValue("link"),
		Description: r.FormValue("description"),
		Language:    r.FormValue("language"),
		Author:      r.FormValue("author"),
		Subtitle:    r.FormValue("subtitle"),
		Summary:     r.FormValue("summary"),
		Explicit:    r.FormValue("explicit"),
		Category:    r.FormValue("category"),
		PubDate:     currentPodcast.PubDate,
		ImageURL:    currentPodcast.ImageURL, // Keep existing unless new artwork uploaded
		Episodes:    currentPodcast.Episodes, // Preserve episodes
	}

	// Validate required fields
	if podcast.Title == "" || podcast.Link == "" || podcast.Description == "" || podcast.Language == "" {
		http.Error(w, "Title, link, description, and language are required", http.StatusBadRequest)
		return
	}

	// Validate URL format
	if !strings.HasPrefix(podcast.Link, "http://") && !strings.HasPrefix(podcast.Link, "https://") {
		http.Error(w, "Link must be a valid HTTP(S) URL", http.StatusBadRequest)
		return
	}

	// Validate language code pattern (e.g., "en-us", "es")
	if matched, _ := regexp.MatchString(`^[a-z]{2}(-[a-z]{2})?$`, podcast.Language); !matched {
		http.Error(w, "Language must be a valid language code (e.g., 'en-us', 'es')", http.StatusBadRequest)
		return
	}

	// Handle artwork upload if provided
	if file, header, err := r.FormFile("artwork"); err == nil {
		defer file.Close()

		// Validate file size
		if header.Size > maxSize {
			http.Error(w, fmt.Sprintf("Artwork too large (max %d MB)", h.maxArtworkMB), http.StatusRequestEntityTooLarge)
			return
		}

		// Validate image format (basic extension check)
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			http.Error(w, "Artwork must be JPG or PNG", http.StatusUnsupportedMediaType)
			return
		}

		// Read artwork data
		artworkData, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read artwork: %v", err), http.StatusInternalServerError)
			return
		}

		// Save artwork file
		artworkFilename, err := storage.SaveArtworkFile(header.Filename, artworkData, h.artworkDir)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to save artwork: %v", err), http.StatusInternalServerError)
			return
		}

		// Update image URL
		podcast.ImageURL = fmt.Sprintf("/static/artwork/%s", artworkFilename)
	}

	// Update podcast settings
	if err := h.store.UpdatePodcast(podcast); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update settings: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success message (for HTMX)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<div class="success-message">Settings saved successfully! <a href="/">Back to Dashboard</a></div>`))
}
