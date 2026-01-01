package handlers

import (
	"net/http"

	"github.com/example/rss-server/internal/storage"
)

// FeedHandler handles RSS feed requests
type FeedHandler struct {
	store *storage.RSSStore
}

// NewFeedHandler creates a new feed handler
func NewFeedHandler(store *storage.RSSStore) *FeedHandler {
	return &FeedHandler{store: store}
}

// HandleFeed handles GET /feed.xml
func (h *FeedHandler) HandleFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	xmlData, err := h.store.ServeXML()
	if err != nil {
		http.Error(w, "Failed to generate RSS feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Write(xmlData)
}
