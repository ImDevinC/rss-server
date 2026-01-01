package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/example/rss-server/internal/models"
	"github.com/example/rss-server/internal/rss"
)

// RSSStore manages the RSS feed with thread-safe access
type RSSStore struct {
	mu       sync.RWMutex
	podcast  *models.Podcast
	filepath string
}

// LoadRSSStore loads or creates a new RSS store from the given file path
func LoadRSSStore(path string) (*RSSStore, error) {
	store := &RSSStore{
		filepath: path,
	}

	// Try to load existing feed
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new default podcast if file doesn't exist
			store.podcast = models.NewDefaultPodcast()

			// Save initial default podcast to disk
			if err := store.saveToDisk(); err != nil {
				return nil, fmt.Errorf("failed to save default podcast: %w", err)
			}

			return store, nil
		}
		return nil, fmt.Errorf("failed to read RSS file: %w", err)
	}

	// Parse existing feed
	p, err := parsePodcastXML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS XML: %w", err)
	}

	store.podcast = p
	return store, nil
}

// GetPodcast returns a copy of the current podcast (thread-safe read)
func (s *RSSStore) GetPodcast() *models.Podcast {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modifications
	p := *s.podcast
	return &p
}

// AddEpisode adds a new episode to the feed and saves atomically
func (s *RSSStore) AddEpisode(ep models.Episode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Add episode to list
	s.podcast.Episodes = append(s.podcast.Episodes, ep)

	// Update podcast pub date to latest episode date
	if ep.PubDate.After(s.podcast.PubDate) {
		s.podcast.PubDate = ep.PubDate
	}

	// Save to disk atomically
	return s.saveToDisk()
}

// DeleteEpisode removes an episode from the feed by ID
func (s *RSSStore) DeleteEpisode(episodeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and remove episode
	found := false
	newEpisodes := make([]models.Episode, 0, len(s.podcast.Episodes))
	for _, ep := range s.podcast.Episodes {
		if ep.ID != episodeID {
			newEpisodes = append(newEpisodes, ep)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("episode not found: %s", episodeID)
	}

	s.podcast.Episodes = newEpisodes

	// Save to disk atomically
	return s.saveToDisk()
}

// UpdatePodcast updates the podcast-level metadata
func (s *RSSStore) UpdatePodcast(p *models.Podcast) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Preserve episodes
	p.Episodes = s.podcast.Episodes

	s.podcast = p

	// Save to disk atomically
	return s.saveToDisk()
}

// ServeXML writes the RSS feed XML to the provided writer
func (s *RSSStore) ServeXML() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return rss.GenerateFeed(s.podcast)
}

// saveToDisk writes the podcast to disk using atomic write (temp file + rename)
func (s *RSSStore) saveToDisk() error {
	// Generate RSS XML
	xmlData, err := rss.GenerateFeed(s.podcast)
	if err != nil {
		return fmt.Errorf("failed to generate RSS XML: %w", err)
	}

	// Write to temp file first
	dir := filepath.Dir(s.filepath)
	tmpFile := filepath.Join(dir, ".podcast.xml.tmp")

	if err := os.WriteFile(tmpFile, xmlData, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename (replaces old file)
	if err := os.Rename(tmpFile, s.filepath); err != nil {
		os.Remove(tmpFile) // Clean up temp file on error
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// parsePodcastXML parses RSS XML into podcast model
func parsePodcastXML(data []byte) (*models.Podcast, error) {
	return rss.ParseFeed(data)
}
