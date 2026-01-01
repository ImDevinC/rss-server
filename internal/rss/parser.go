package rss

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/example/rss-server/internal/models"
)

// RSS represents the RSS 2.0 XML structure
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	XMLNS   string   `xml:"xmlns:itunes,attr"`
	Channel Channel  `xml:"channel"`
}

// Channel represents the RSS channel element
type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Language    string `xml:"language"`
	PubDate     string `xml:"pubDate,omitempty"`

	// iTunes fields
	Author   string `xml:"itunes:author,omitempty"`
	Subtitle string `xml:"itunes:subtitle,omitempty"`
	Summary  string `xml:"itunes:summary,omitempty"`
	ImageURL string `xml:"itunes:image,omitempty"`
	Explicit string `xml:"itunes:explicit,omitempty"`
	Category string `xml:"itunes:category,omitempty"`

	Items []Item `xml:"item"`
}

// Item represents an RSS item (episode)
type Item struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	PubDate     string    `xml:"pubDate"`
	GUID        string    `xml:"guid"`
	Enclosure   Enclosure `xml:"enclosure"`

	// iTunes fields
	Duration    string `xml:"itunes:duration,omitempty"`
	Explicit    string `xml:"itunes:explicit,omitempty"`
	EpisodeNum  int    `xml:"itunes:episode,omitempty"`
	SeasonNum   int    `xml:"itunes:season,omitempty"`
	EpisodeType string `xml:"itunes:episodeType,omitempty"`
}

// Enclosure represents the audio file enclosure
type Enclosure struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

// ParseFeed parses RSS XML into a podcast model
func ParseFeed(data []byte) (*models.Podcast, error) {
	var rss RSS
	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil, fmt.Errorf("failed to parse RSS XML: %w", err)
	}

	// Parse pub date
	pubDate := time.Now()
	if rss.Channel.PubDate != "" {
		parsed, err := time.Parse(time.RFC1123Z, rss.Channel.PubDate)
		if err == nil {
			pubDate = parsed
		}
	}

	podcast := &models.Podcast{
		Title:       rss.Channel.Title,
		Link:        rss.Channel.Link,
		Description: rss.Channel.Description,
		Language:    rss.Channel.Language,
		PubDate:     pubDate,
		Author:      rss.Channel.Author,
		Subtitle:    rss.Channel.Subtitle,
		Summary:     rss.Channel.Summary,
		ImageURL:    rss.Channel.ImageURL,
		Explicit:    rss.Channel.Explicit,
		Category:    rss.Channel.Category,
		Episodes:    make([]models.Episode, 0, len(rss.Channel.Items)),
	}

	// Parse episodes
	for _, item := range rss.Channel.Items {
		epPubDate := time.Now()
		if item.PubDate != "" {
			parsed, err := time.Parse(time.RFC1123Z, item.PubDate)
			if err == nil {
				epPubDate = parsed
			}
		}

		episode := models.Episode{
			ID:          item.GUID,
			Title:       item.Title,
			Description: item.Description,
			PubDate:     epPubDate,
			GUID:        item.GUID,
			AudioURL:    item.Enclosure.URL,
			AudioLength: item.Enclosure.Length,
			AudioType:   item.Enclosure.Type,
			Duration:    item.Duration,
			Explicit:    item.Explicit,
			EpisodeNum:  item.EpisodeNum,
			SeasonNum:   item.SeasonNum,
			EpisodeType: item.EpisodeType,
		}

		podcast.Episodes = append(podcast.Episodes, episode)
	}

	return podcast, nil
}
