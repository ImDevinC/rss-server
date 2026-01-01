package models

import "time"

// Episode represents a single podcast episode with RSS 2.0 + iTunes metadata
type Episode struct {
	// Unique identifier (generated from filename or UUID)
	ID string `json:"id"`

	// Required RSS 2.0 fields
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PubDate     time.Time `json:"pubDate"`
	GUID        string    `json:"guid"`

	// Enclosure (audio file)
	AudioURL    string `json:"audioURL"`
	AudioLength int64  `json:"audioLength"` // bytes
	AudioType   string `json:"audioType"`   // "audio/mpeg"

	// iTunes-specific fields
	Duration    string `json:"duration,omitempty"`    // "HH:MM:SS" or seconds
	Explicit    string `json:"explicit,omitempty"`    // "yes", "no", "clean"
	EpisodeNum  int    `json:"episodeNum,omitempty"`  // episode number
	SeasonNum   int    `json:"seasonNum,omitempty"`   // season number
	EpisodeType string `json:"episodeType,omitempty"` // "full", "trailer", "bonus"

	// Metadata for internal use
	Filename   string    `json:"filename"`   // Audio filename on disk
	UploadDate time.Time `json:"uploadDate"` // When episode was added
}

// AudioFile represents the actual audio file stored by the system
type AudioFile struct {
	// Storage information
	Filename     string // Stored filename (unique, URL-safe)
	OriginalName string // Original uploaded filename
	FilePath     string // Full path on disk

	// File properties
	Size     int64  // File size in bytes
	MimeType string // MIME type (e.g., "audio/mpeg")
	Duration string // Calculated duration "HH:MM:SS"

	// Metadata
	UploadDate time.Time // When file was uploaded
}
