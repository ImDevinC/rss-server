package rss

import (
	"fmt"
	"sort"
	"time"

	"github.com/eduncan911/podcast"
	"github.com/example/rss-server/internal/models"
)

// GenerateFeed creates an RSS 2.0 + iTunes feed from the podcast model
func GenerateFeed(p *models.Podcast) ([]byte, error) {
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
	if p.ImageURL != "" {
		feed.IImage = &podcast.IImage{HREF: p.ImageURL}
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

		// Add enclosure (audio file)
		if ep.AudioURL != "" {
			item.AddEnclosure(ep.AudioURL, podcast.MP3, ep.AudioLength)
		}

		// iTunes fields
		if ep.Duration != "" {
			item.IDuration = ep.Duration
		}
		if ep.Explicit != "" {
			item.IExplicit = ep.Explicit
		}

		if _, err := feed.AddItem(item); err != nil {
			return nil, fmt.Errorf("failed to add episode %s: %w", ep.ID, err)
		}
	}

	// Generate XML bytes
	return feed.Bytes(), nil
}
