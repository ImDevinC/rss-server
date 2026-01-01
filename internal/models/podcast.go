package models

import "time"

// Podcast represents the podcast show with channel-level metadata
type Podcast struct {
	// Required RSS 2.0 fields
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description"`
	Language    string    `json:"language"`
	PubDate     time.Time `json:"pubDate"`

	// iTunes-specific fields
	Author   string `json:"author"`
	Subtitle string `json:"subtitle,omitempty"`
	Summary  string `json:"summary,omitempty"`
	ImageURL string `json:"imageURL,omitempty"`
	Explicit string `json:"explicit,omitempty"` // "yes", "no", "clean"
	Category string `json:"category,omitempty"`

	// Episode list
	Episodes []Episode `json:"episodes"`
}

// NewDefaultPodcast creates a new podcast with default values
func NewDefaultPodcast() *Podcast {
	return &Podcast{
		Title:       "My Podcast",
		Link:        "https://example.com",
		Description: "A podcast about interesting topics",
		Language:    "en-us",
		PubDate:     time.Now(),
		Author:      "Podcast Creator",
		Subtitle:    "Interesting conversations",
		Summary:     "A podcast about interesting topics",
		ImageURL:    "/static/default-podcast-artwork.jpg",
		Explicit:    "no",
		Category:    "Technology",
		Episodes:    []Episode{},
	}
}
