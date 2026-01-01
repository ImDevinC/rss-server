package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/example/rss-server/internal/storage"
)

// WebHandler handles web UI requests
type WebHandler struct {
	store     *storage.RSSStore
	templates *template.Template
	baseURL   string // T046: Add baseURL field
}

// NewWebHandler creates a new web handler
// T047: Updated to accept baseURL parameter
func NewWebHandler(store *storage.RSSStore, templatesDir string, baseURL string) (*WebHandler, error) {
	// Parse all templates including components
	tmpl, err := template.ParseGlob(templatesDir + "/*.html")
	if err != nil {
		return nil, err
	}

	// Parse component templates
	tmpl, err = tmpl.ParseGlob(templatesDir + "/components/*.html")
	if err != nil {
		return nil, err
	}

	return &WebHandler{
		store:     store,
		templates: tmpl,
		baseURL:   baseURL,
	}, nil
}

// HandleDashboard handles GET /
// T048: Updated to use baseURL for FeedURL
func (h *WebHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get podcast data
	podcast := h.store.GetPodcast()

	// Prepare template data
	data := map[string]interface{}{
		"Podcast": podcast,
		"FeedURL": fmt.Sprintf("%s/feed.xml", h.baseURL),
	}

	// Render template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}
}
