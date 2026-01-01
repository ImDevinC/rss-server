package rss

import (
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/eduncan911/podcast"
	"github.com/example/rss-server/internal/models"
)

// convertToAbsoluteURL converts a relative URL to an absolute URL using the base URL
// T030: Helper function for URL conversion with RFC 3986 encoding
func convertToAbsoluteURL(baseURL, relativePath string) (string, error) {
	// If the path is already absolute, return it as-is
	if strings.HasPrefix(relativePath, "http://") || strings.HasPrefix(relativePath, "https://") {
		return relativePath, nil
	}

	// Parse base URL
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	// Parse relative path
	rel, err := url.Parse(relativePath)
	if err != nil {
		return "", fmt.Errorf("invalid relative path: %w", err)
	}

	// Resolve the relative URL against the base URL
	absolute := base.ResolveReference(rel)

	return absolute.String(), nil
}

// GenerateFeed creates an RSS 2.0 + iTunes feed from the podcast model
// T031: Updated signature to accept baseURL parameter
func GenerateFeed(p *models.Podcast, baseURL string) ([]byte, error) {
	now := time.Now()
	pubDate := p.PubDate
	if pubDate.IsZero() {
		pubDate = now
	}

	// Create podcast feed
	feed := podcast.New(
		p.Title,
		p.Link,
		p.Description,
		&pubDate,
		&now,
	)

	// Set language
	feed.Language = p.Language

	// Add iTunes metadata
	feed.IAuthor = p.Author
	if p.Subtitle != "" {
		feed.ISubtitle = p.Subtitle
	}
	if p.Summary != "" {
		feed.ISummary = &podcast.ISummary{Text: p.Summary}
	}
	// T032: Apply URL conversion to podcast ImageURL
	if p.ImageURL != "" {
		absoluteImageURL, err := convertToAbsoluteURL(baseURL, p.ImageURL)
		if err != nil {
			log.Printf("Warning: Failed to convert podcast image URL '%s': %v", p.ImageURL, err)
		} else {
			feed.IImage = &podcast.IImage{HREF: absoluteImageURL}
		}
	}
	if p.Explicit != "" {
		feed.IExplicit = p.Explicit
	}
	if p.Category != "" {
		feed.AddCategory(p.Category, nil)
	}

	// Sort episodes by PubDate (descending - newest first)
	episodes := make([]models.Episode, len(p.Episodes))
	copy(episodes, p.Episodes)
	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].PubDate.After(episodes[j].PubDate)
	})

	// Add episodes
	// T034: Add error handling to skip malformed episodes
	for _, ep := range episodes {
		item := podcast.Item{
			Title:       ep.Title,
			Description: ep.Description,
			PubDate:     &ep.PubDate,
		}

		// Set GUID
		if ep.GUID != "" {
			item.GUID = ep.GUID
		} else {
			item.GUID = ep.ID
		}

		// T033: Apply URL conversion to episode AudioURL with RFC 3986 encoding
		// Add enclosure (audio file)
		if ep.AudioURL != "" {
			absoluteAudioURL, err := convertToAbsoluteURL(baseURL, ep.AudioURL)
			if err != nil {
				// T034: Skip malformed episodes, log error
				log.Printf("Warning: Skipping episode '%s' due to invalid audio URL '%s': %v", ep.ID, ep.AudioURL, err)
				continue
			}
			item.AddEnclosure(absoluteAudioURL, podcast.MP3, ep.AudioLength)
		}

		// iTunes fields
		if ep.Duration != "" {
			item.IDuration = ep.Duration
		}
		if ep.Explicit != "" {
			item.IExplicit = ep.Explicit
		}

		if _, err := feed.AddItem(item); err != nil {
			log.Printf("Warning: Failed to add episode '%s': %v", ep.ID, err)
			continue
		}
	}

	// Generate XML bytes
	return feed.Bytes(), nil
}
