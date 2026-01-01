package integration

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/example/rss-server/internal/models"
	"github.com/example/rss-server/internal/rss"
)

// RSS structure for validation
type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Items []Item `xml:"item"`
}

type Item struct {
	Enclosure Enclosure `xml:"enclosure"`
}

type Enclosure struct {
	URL string `xml:"url,attr"`
}

// T027: RSS XML structure valid
func TestRSSXMLStructureValid(t *testing.T) {
	baseURL := "http://example.com"
	podcast := &models.Podcast{
		Title:       "Test Podcast",
		Link:        "http://example.com",
		Description: "Test Description",
		Language:    "en-us",
		Author:      "Test Author",
		PubDate:     time.Now(),
		Episodes: []models.Episode{
			{
				ID:          "ep1",
				Title:       "Episode 1",
				Description: "First episode",
				PubDate:     time.Now(),
				AudioURL:    "/audio/episode1.mp3",
				AudioLength: 12345,
			},
		},
	}

	xmlBytes, err := rss.GenerateFeed(podcast, baseURL)
	if err != nil {
		t.Fatalf("Failed to generate feed: %v", err)
	}

	// Try to parse the generated XML
	var feed RSSFeed
	if err := xml.Unmarshal(xmlBytes, &feed); err != nil {
		t.Fatalf("Generated RSS XML is invalid: %v", err)
	}

	// Verify basic structure
	if len(feed.Channel.Items) != 1 {
		t.Errorf("Expected 1 item in RSS feed, got %d", len(feed.Channel.Items))
	}
}

// T028: Enclosure tags have absolute URLs
func TestEnclosureAbsoluteURLs(t *testing.T) {
	baseURL := "http://podcast.example.com:8080"
	podcast := &models.Podcast{
		Title:       "Test Podcast",
		Link:        "http://example.com",
		Description: "Test Description",
		Language:    "en-us",
		Author:      "Test Author",
		PubDate:     time.Now(),
		Episodes: []models.Episode{
			{
				ID:          "ep1",
				Title:       "Episode 1",
				Description: "First episode",
				PubDate:     time.Now(),
				AudioURL:    "/audio/episode1.mp3",
				AudioLength: 12345,
			},
			{
				ID:          "ep2",
				Title:       "Episode 2",
				Description: "Second episode",
				PubDate:     time.Now(),
				AudioURL:    "/audio/episode2.mp3",
				AudioLength: 54321,
			},
		},
	}

	xmlBytes, err := rss.GenerateFeed(podcast, baseURL)
	if err != nil {
		t.Fatalf("Failed to generate feed: %v", err)
	}

	// Parse the generated XML
	var feed RSSFeed
	if err := xml.Unmarshal(xmlBytes, &feed); err != nil {
		t.Fatalf("Generated RSS XML is invalid: %v", err)
	}

	// Check that all enclosures have absolute URLs
	for i, item := range feed.Channel.Items {
		if item.Enclosure.URL == "" {
			t.Errorf("Item %d has empty enclosure URL", i)
			continue
		}

		// Check that URL starts with http:// or https://
		if !strings.HasPrefix(item.Enclosure.URL, "http://") && !strings.HasPrefix(item.Enclosure.URL, "https://") {
			t.Errorf("Item %d enclosure URL is not absolute: %s", i, item.Enclosure.URL)
		}

		// Check that URL contains the base URL
		if !strings.HasPrefix(item.Enclosure.URL, baseURL) {
			t.Errorf("Item %d enclosure URL does not start with base URL '%s': got '%s'", i, baseURL, item.Enclosure.URL)
		}
	}
}
