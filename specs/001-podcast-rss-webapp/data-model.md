# Data Model: Podcast RSS Webapp

**Feature**: Podcast RSS Webapp  
**Phase**: Phase 1 - Design  
**Date**: 2025-12-31

## Overview

This document defines the data structures for the podcast RSS feed system. Per the constitution's File-Centric Architecture principle, the RSS XML file (`podcast.xml`) is the single source of truth. All data structures map directly to RSS 2.0 + iTunes podcast specification elements.

## Storage Architecture

**Source of Truth**: `data/podcast.xml` (RSS XML file)

**Storage Locations**:
- `data/podcast.xml` - RSS feed (all metadata + episode list)
- `data/audio/*.mp3` - Audio files
- `data/artwork/*.jpg` - Podcast artwork

**No Database**: All data persists in XML + filesystem per constitutional requirements.

## Core Data Structures

### 1. Podcast (Channel-Level Metadata)

Represents the podcast show itself. Maps to RSS `<channel>` element with iTunes extensions.

**Go Struct** (using `github.com/eduncan911/podcast` library):

```go
package models

import (
    "time"
    "github.com/eduncan911/podcast"
)

type Podcast struct {
    // Required RSS 2.0 fields
    Title       string    // <title>
    Link        string    // <link> - podcast homepage URL
    Description string    // <description>
    Language    string    // <language> - e.g., "en-us"
    PubDate     time.Time // <pubDate> - RFC-822 format
    
    // iTunes-specific fields
    Author      string   // <itunes:author>
    Subtitle    string   // <itunes:subtitle>
    Summary     string   // <itunes:summary>
    ImageURL    string   // <itunes:image href="">
    Explicit    string   // <itunes:explicit> - "yes", "no", "clean"
    Category    string   // <itunes:category text="">
    
    // Episode list
    Episodes []Episode
}
```

**RSS XML Mapping**:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" 
     xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">
  <channel>
    <title>{{ .Title }}</title>
    <link>{{ .Link }}</link>
    <description>{{ .Description }}</description>
    <language>{{ .Language }}</language>
    <pubDate>{{ .PubDate | rfc822 }}</pubDate>
    
    <itunes:author>{{ .Author }}</itunes:author>
    <itunes:subtitle>{{ .Subtitle }}</itunes:subtitle>
    <itunes:summary>{{ .Summary }}</itunes:summary>
    <itunes:image href="{{ .ImageURL }}" />
    <itunes:explicit>{{ .Explicit }}</itunes:explicit>
    <itunes:category text="{{ .Category }}" />
    
    <!-- Episodes as <item> elements -->
  </channel>
</rss>
```

**Validation Rules**:

- `Title`: REQUIRED, non-empty string, max 255 chars
- `Link`: REQUIRED, valid HTTP(S) URL
- `Description`: REQUIRED, non-empty string, max 4000 chars
- `Language`: REQUIRED, valid language code (e.g., "en-us", "es")
- `Author`: REQUIRED for iTunes, non-empty string, max 255 chars
- `ImageURL`: REQUIRED for iTunes, valid HTTP(S) URL, must be square image 1400x1400 to 3000x3000 pixels
- `Explicit`: One of "yes", "no", "clean" (defaults to "no")
- `Category`: REQUIRED for iTunes, must be valid iTunes category

**Default Values**:

```go
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
```

### 2. Episode (Item-Level Metadata)

Represents a single podcast episode. Maps to RSS `<item>` element with iTunes extensions.

**Go Struct**:

```go
package models

import (
    "time"
)

type Episode struct {
    // Unique identifier (generated from filename or UUID)
    ID string // e.g., "ep-20231215-interview-john-doe"
    
    // Required RSS 2.0 fields
    Title       string    // <title>
    Description string    // <description>
    PubDate     time.Time // <pubDate> - RFC-822 format
    GUID        string    // <guid> - unique episode identifier
    
    // Enclosure (audio file)
    AudioURL    string // <enclosure url="">
    AudioLength int64  // <enclosure length=""> in bytes
    AudioType   string // <enclosure type=""> - "audio/mpeg" for MP3
    
    // iTunes-specific fields
    Duration    string // <itunes:duration> - "HH:MM:SS" or seconds
    Explicit    string // <itunes:explicit> - "yes", "no", "clean"
    EpisodeNum  int    // <itunes:episode> - episode number
    SeasonNum   int    // <itunes:season> - season number
    EpisodeType string // <itunes:episodeType> - "full", "trailer", "bonus"
    
    // Metadata for internal use
    Filename    string    // Audio filename on disk (e.g., "ep-001.mp3")
    UploadDate  time.Time // When episode was added
}
```

**RSS XML Mapping**:

```xml
<item>
  <title>{{ .Title }}</title>
  <description>{{ .Description }}</description>
  <pubDate>{{ .PubDate | rfc822 }}</pubDate>
  <guid isPermaLink="false">{{ .GUID }}</guid>
  
  <enclosure url="{{ .AudioURL }}" 
             length="{{ .AudioLength }}" 
             type="{{ .AudioType }}" />
  
  <itunes:duration>{{ .Duration }}</itunes:duration>
  <itunes:explicit>{{ .Explicit }}</itunes:explicit>
  <itunes:episode>{{ .EpisodeNum }}</itunes:episode>
  <itunes:season>{{ .SeasonNum }}</itunes:season>
  <itunes:episodeType>{{ .EpisodeType }}</itunes:episodeType>
</item>
```

**Validation Rules**:

- `ID`: REQUIRED, unique, URL-safe string
- `Title`: REQUIRED, non-empty string, max 255 chars
- `Description`: REQUIRED, non-empty string, max 4000 chars
- `PubDate`: REQUIRED, valid RFC-822 datetime
- `GUID`: REQUIRED, globally unique identifier (can be same as ID or audio URL)
- `AudioURL`: REQUIRED, valid HTTP(S) URL to audio file
- `AudioLength`: REQUIRED, positive integer (bytes)
- `AudioType`: REQUIRED, valid MIME type ("audio/mpeg" for MP3)
- `Duration`: Optional but recommended, format "HH:MM:SS" or seconds as integer
- `Explicit`: One of "yes", "no", "clean" (defaults to inherit from podcast)
- `EpisodeNum`: Optional, positive integer
- `SeasonNum`: Optional, positive integer
- `EpisodeType`: One of "full", "trailer", "bonus" (defaults to "full")

**ID Generation**:

```go
func GenerateEpisodeID(title string, pubDate time.Time) string {
    // Sanitize title for URL safety
    sanitized := strings.ToLower(title)
    sanitized = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(sanitized, "-")
    sanitized = strings.Trim(sanitized, "-")
    
    // Combine date + title (truncate if too long)
    dateStr := pubDate.Format("20060102")
    id := fmt.Sprintf("ep-%s-%s", dateStr, sanitized)
    
    if len(id) > 100 {
        id = id[:100]
    }
    
    return id
}
```

**Sorting**:

Episodes MUST be sorted by `PubDate` in descending order (newest first) when generating RSS feed per FR-008.

```go
func SortEpisodesByDate(episodes []Episode) {
    sort.Slice(episodes, func(i, j int) bool {
        return episodes[i].PubDate.After(episodes[j].PubDate)
    })
}
```

### 3. Audio File Metadata

Represents physical audio file storage. Not directly in RSS but needed for file management.

**Go Struct**:

```go
package models

import (
    "time"
)

type AudioFile struct {
    // Storage information
    Filename     string    // Stored filename (unique, URL-safe)
    OriginalName string    // Original uploaded filename
    FilePath     string    // Full path on disk (e.g., "data/audio/ep-001.mp3")
    
    // File properties
    Size         int64     // File size in bytes
    MimeType     string    // MIME type (e.g., "audio/mpeg")
    Duration     string    // Calculated duration "HH:MM:SS"
    
    // Metadata
    UploadDate   time.Time // When file was uploaded
}
```

**File Naming Strategy**:

```go
func GenerateUniqueFilename(originalName string) string {
    ext := filepath.Ext(originalName)
    base := strings.TrimSuffix(originalName, ext)
    
    // Sanitize base name
    sanitized := regexp.MustCompile(`[^a-zA-Z0-9-_]`).ReplaceAllString(base, "-")
    
    // Add timestamp for uniqueness
    timestamp := time.Now().Format("20060102-150405")
    
    return fmt.Sprintf("%s-%s%s", sanitized, timestamp, ext)
}
```

## Data Flow Diagrams

### Add Episode Flow

```
User Upload Request
        ↓
[Validate Audio File]
        ↓
[Generate Episode ID]
        ↓
[Save Audio to Filesystem]
        ↓
[Read podcast.xml]
        ↓
[Add Episode to XML]
        ↓
[Sort Episodes by PubDate]
        ↓
[Write podcast.xml (atomic)]
        ↓
[Return Success]
```

### Delete Episode Flow

```
Delete Request (Episode ID)
        ↓
[Read podcast.xml]
        ↓
[Find Episode by ID]
        ↓
[Remove Episode from XML]
        ↓
[Write podcast.xml (atomic)]
        ↓
[Delete Audio File from Filesystem]
        ↓
[Return Success]
```

### Serve RSS Feed Flow

```
GET /feed.xml Request
        ↓
[RWMutex.RLock()]
        ↓
[Read In-Memory Feed]
        ↓
[Encode to XML]
        ↓
[Set Content-Type: application/rss+xml]
        ↓
[RWMutex.RUnlock()]
        ↓
[Stream XML to Client]
```

## State Transitions

### Episode States

Episodes have minimal state as they are either present or absent in the RSS feed:

- **Not Uploaded**: Episode does not exist
- **Uploaded**: Episode exists in RSS feed and audio file on disk
- **Deleted**: Episode removed from RSS feed and audio file deleted

No intermediate states (no "draft", "pending", "archived") per simplified requirements.

### Podcast Metadata States

Podcast metadata can be:

- **Default**: Using system-generated defaults
- **Customized**: User has configured podcast-level settings

Changes to podcast metadata update the RSS XML file immediately (no approval workflow).

## Concurrency Considerations

### Read Operations (Concurrent)

Multiple concurrent reads are safe and do not block each other:

- `GET /feed.xml` - Serve RSS feed
- `GET /` - Dashboard view (reads episode list)
- `GET /audio/{filename}` - Audio file downloads

### Write Operations (Serialized)

Write operations acquire exclusive lock and are serialized:

- `POST /api/episodes` - Add new episode
- `DELETE /api/episodes/{id}` - Delete episode
- `POST /api/podcast/settings` - Update podcast metadata

### Lock Strategy

```go
type RSSStore struct {
    mu       sync.RWMutex
    feed     *podcast.Podcast
    filepath string
}

// Read: Many concurrent readers
func (s *RSSStore) GetFeed() *podcast.Podcast {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.feed
}

// Write: Exclusive access
func (s *RSSStore) AddEpisode(ep Episode) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Modify feed
    s.feed.AddItem(ep)
    
    // Atomic write to disk
    return s.writeToFile()
}
```

## XML Schema Reference

### Complete RSS 2.0 + iTunes Structure

```xml
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" 
     xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">
  <channel>
    <!-- Required RSS 2.0 elements -->
    <title>Podcast Title</title>
    <link>https://example.com</link>
    <description>Podcast description</description>
    <language>en-us</language>
    <pubDate>Mon, 01 Jan 2024 00:00:00 GMT</pubDate>
    
    <!-- iTunes podcast elements -->
    <itunes:author>Author Name</itunes:author>
    <itunes:subtitle>Short subtitle</itunes:subtitle>
    <itunes:summary>Detailed podcast summary</itunes:summary>
    <itunes:image href="https://example.com/artwork.jpg" />
    <itunes:explicit>no</itunes:explicit>
    <itunes:category text="Technology">
      <itunes:category text="Podcasting" />
    </itunes:category>
    
    <!-- Episodes (items) -->
    <item>
      <title>Episode Title</title>
      <description>Episode description</description>
      <pubDate>Mon, 15 Jan 2024 10:00:00 GMT</pubDate>
      <guid isPermaLink="false">ep-20240115-episode-title</guid>
      
      <enclosure url="https://example.com/audio/episode.mp3" 
                 length="12345678" 
                 type="audio/mpeg" />
      
      <itunes:duration>45:30</itunes:duration>
      <itunes:explicit>no</itunes:explicit>
      <itunes:episode>1</itunes:episode>
      <itunes:season>1</itunes:season>
      <itunes:episodeType>full</itunes:episodeType>
    </item>
    
    <!-- More items... -->
  </channel>
</rss>
```

## iTunes Category Reference

Valid top-level categories for `<itunes:category>`:

- Arts
- Business
- Comedy
- Education
- Fiction
- Government
- History
- Health & Fitness
- Kids & Family
- Leisure
- Music
- News
- Religion & Spirituality
- Science
- Society & Culture
- Sports
- Technology
- True Crime
- TV & Film

Each category may have subcategories (e.g., Technology > Podcasting).

## File Size Constraints

| Resource | Constraint | Rationale |
|----------|-----------|-----------|
| Audio file | 500 MB max | Constitutional requirement (FR-010) |
| Artwork image | 5 MB max | iTunes recommendation |
| Episode title | 255 chars | RSS best practices |
| Episode description | 4000 chars | iTunes limit |
| Podcast title | 255 chars | RSS best practices |
| Podcast description | 4000 chars | iTunes limit |
| Total episodes | 1000 episodes | SC-010 performance target |

## Example Data Instances

### Example Podcast XML

```xml
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">
  <channel>
    <title>Tech Talks Daily</title>
    <link>https://techtalks.example.com</link>
    <description>Daily conversations about technology and innovation</description>
    <language>en-us</language>
    <pubDate>Wed, 31 Dec 2025 12:00:00 GMT</pubDate>
    
    <itunes:author>Jane Developer</itunes:author>
    <itunes:subtitle>Technology insights for developers</itunes:subtitle>
    <itunes:summary>Daily conversations about technology and innovation in the software industry</itunes:summary>
    <itunes:image href="https://techtalks.example.com/artwork.jpg" />
    <itunes:explicit>no</itunes:explicit>
    <itunes:category text="Technology">
      <itunes:category text="Podcasting" />
    </itunes:category>
    
    <item>
      <title>Getting Started with Go</title>
      <description>An introduction to the Go programming language</description>
      <pubDate>Wed, 31 Dec 2025 10:00:00 GMT</pubDate>
      <guid isPermaLink="false">ep-20251231-getting-started-with-go</guid>
      
      <enclosure url="https://techtalks.example.com/audio/ep-20251231-getting-started-with-go.mp3" 
                 length="45678901" 
                 type="audio/mpeg" />
      
      <itunes:duration>32:15</itunes:duration>
      <itunes:explicit>no</itunes:explicit>
      <itunes:episode>1</itunes:episode>
      <itunes:season>1</itunes:season>
      <itunes:episodeType>full</itunes:episodeType>
    </item>
  </channel>
</rss>
```

### Example Go Usage

```go
// Load existing podcast
store, err := LoadRSSStore("data/podcast.xml")

// Add new episode
episode := Episode{
    ID:          "ep-20251231-new-episode",
    Title:       "My New Episode",
    Description: "This is a great episode",
    PubDate:     time.Now(),
    GUID:        "ep-20251231-new-episode",
    AudioURL:    "https://example.com/audio/ep-001.mp3",
    AudioLength: 12345678,
    AudioType:   "audio/mpeg",
    Duration:    "45:30",
    Explicit:    "no",
    EpisodeNum:  1,
    SeasonNum:   1,
    EpisodeType: "full",
    Filename:    "ep-001.mp3",
    UploadDate:  time.Now(),
}

err = store.AddEpisode(episode)

// Serve RSS feed
http.HandleFunc("/feed.xml", func(w http.ResponseWriter, r *http.Request) {
    store.ServeXML(w)
})
```

## Summary

This data model provides:

- **File-centric architecture**: RSS XML as source of truth
- **RSS 2.0 compliance**: All required elements present
- **iTunes compatibility**: Full podcast namespace support
- **Concurrent safety**: RWMutex pattern for read-heavy workload
- **Atomic updates**: Temp file + rename for write safety
- **Simple state management**: Episodes are either present or absent
- **Extensibility**: Easy to add fields per iTunes spec updates
