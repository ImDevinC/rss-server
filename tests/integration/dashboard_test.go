package integration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/rss-server/internal/handlers"
	"github.com/example/rss-server/internal/models"
	"github.com/example/rss-server/internal/storage"
)

// T043: Dashboard displays absolute feed URL
func TestDashboardAbsoluteFeedURL(t *testing.T) {
	// Create temporary RSS file for testing
	tmpFile := t.TempDir() + "/podcast.xml"
	baseURL := "http://example.com:8080"

	store, err := storage.LoadRSSStore(tmpFile, baseURL)
	if err != nil {
		t.Fatalf("Failed to create RSS store: %v", err)
	}

	// Create web handler with base URL
	handler, err := handlers.NewWebHandler(store, "./../../web/templates", baseURL)
	if err != nil {
		t.Fatalf("Failed to create web handler: %v", err)
	}

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Call handler
	handler.HandleDashboard(rec, req)

	// Check response
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()

	// Check that the absolute feed URL is displayed
	expectedFeedURL := "http://example.com:8080/feed.xml"
	if !strings.Contains(body, expectedFeedURL) {
		t.Errorf("Expected dashboard to contain absolute feed URL '%s', but it was not found", expectedFeedURL)
	}
}

// T044: Feed URL matches configured base URL
func TestDashboardFeedURLMatchesConfig(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		expectedURL string
	}{
		{
			name:        "HTTP with port",
			baseURL:     "http://podcast.example.com:8080",
			expectedURL: "http://podcast.example.com:8080/feed.xml",
		},
		{
			name:        "HTTPS without port",
			baseURL:     "https://mypodcast.com",
			expectedURL: "https://mypodcast.com/feed.xml",
		},
		{
			name:        "HTTP localhost",
			baseURL:     "http://localhost:3000",
			expectedURL: "http://localhost:3000/feed.xml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary RSS file for testing
			tmpFile := t.TempDir() + "/podcast.xml"

			store, err := storage.LoadRSSStore(tmpFile, tt.baseURL)
			if err != nil {
				t.Fatalf("Failed to create RSS store: %v", err)
			}

			// Update podcast with some data
			podcast := models.NewDefaultPodcast()
			if err := store.UpdatePodcast(podcast); err != nil {
				t.Fatalf("Failed to update podcast: %v", err)
			}

			// Create web handler with base URL
			handler, err := handlers.NewWebHandler(store, "./../../web/templates", tt.baseURL)
			if err != nil {
				t.Fatalf("Failed to create web handler: %v", err)
			}

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			// Call handler
			handler.HandleDashboard(rec, req)

			// Check response
			if rec.Code != http.StatusOK {
				t.Fatalf("Expected status 200, got %d", rec.Code)
			}

			body := rec.Body.String()

			// Check that the feed URL matches the configured base URL
			if !strings.Contains(body, tt.expectedURL) {
				t.Errorf("Expected dashboard to contain feed URL '%s', but it was not found", tt.expectedURL)
			}
		})
	}
}
