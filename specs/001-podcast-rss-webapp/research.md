# Research: Podcast RSS Webapp

**Feature**: Podcast RSS Webapp  
**Phase**: Phase 0 - Research & Technology Decisions  
**Date**: 2025-12-31

## Overview

This document captures research findings and technology decisions for implementing a Go-based podcast RSS feed generator with HTMX frontend. The research focused on four critical areas: RSS feed generation, HTMX integration patterns, concurrent XML file handling, and RSS validation strategies.

## 1. RSS 2.0 Feed Generation

### Decision

Use **`github.com/eduncan911/podcast`** library for RSS feed generation.

### Rationale

- Most mature and actively maintained podcast RSS library in Go ecosystem
- Comprehensive iTunes namespace support with all required podcast tags
- Clean, idiomatic Go API with built-in validation
- Production-proven (author reports 6 iTunes-accepted podcasts using this library)
- Excellent documentation with detailed examples
- Handles RSS 2.0 spec + iTunes podcast extensions automatically
- Proper XML namespace handling without manual intervention

### Alternatives Considered

1. **`github.com/gorilla/feeds`**
   - More generic RSS/Atom generator
   - Lacks comprehensive iTunes podcast tag support
   - Would require custom extensions for podcast-specific features
   - Better suited for blog RSS feeds

2. **Custom `encoding/xml` structs**
   - Maximum control over XML output
   - Requires manual implementation of iTunes spec
   - More code to maintain and validate
   - Higher risk of RSS validation failures
   - Only justified if library doesn't meet needs

3. **`github.com/rssblue/types`**
   - Good Podcast 2.0 namespace support
   - Less mature than eduncan911/podcast
   - Podcast 2.0 features are out of scope for MVP

### Implementation Notes

```go
import "github.com/eduncan911/podcast"

// Create podcast feed
p := podcast.New(
    "My Podcast",                    // title
    "https://example.com",           // link
    "Podcast description",           // description
    &pubDate,                        // pub date
    &lastBuildDate,                  // last build date
)

// Add iTunes-specific metadata
p.IAuthor = "Author Name"
p.ISubtitle = "Podcast subtitle"
p.ISummary = "Detailed summary"
p.IImage = &podcast.IImage{HREF: "https://example.com/artwork.jpg"}
p.IExplicit = "no"
p.AddCategory("Technology", []string{"Podcasting"})

// Add episode
item := podcast.Item{
    Title:       "Episode 1",
    Description: "Episode description",
    PubDate:     &episodePubDate,
}
item.AddEnclosure("https://example.com/episode1.mp3", podcast.MP3, 12345678)
p.AddItem(item)

// Generate XML
xmlBytes, err := p.Bytes()
// Or stream to writer: p.Encode(w)
```

**Key Features Used**:
- Automatic RSS 2.0 + iTunes namespace generation
- `AddEnclosure()` for audio file metadata (URL, type, length)
- `AddCategory()` for iTunes categories
- Built-in validation of required fields

## 2. HTMX Integration with Go Templates

### Decision

Use Go's standard **`html/template`** package with partial rendering pattern for HTMX integration.

### Rationale

- Native Go solution with zero external dependencies
- HTMX works naturally with any backend that returns HTML fragments
- Simpler than JavaScript frameworks while providing dynamic UX
- Excellent community examples and patterns available
- Natural fit with Go's server-side rendering philosophy
- No build step or transpilation required

### Alternatives Considered

1. **Template component libraries** (htmgo, go-htmx helpers)
   - Add abstraction layer with marginal benefit
   - Extra dependencies for functionality achievable with standard library
   - Not worth complexity for this project size

2. **Server-side rendering frameworks** (Templ, Gomponents)
   - Overkill for simple podcast feed management UI
   - Adds build complexity
   - Standard templates sufficient for this use case

### Implementation Notes

**Template Organization**:
```go
//go:embed web/templates
var templateFS embed.FS

var templates = template.Must(
    template.ParseFS(templateFS, "web/templates/*.html"),
)
```

**Partial Rendering Pattern**:
```
web/templates/
├── index.html           # Full page layout
└── components/
    ├── episode_row.html    # Single episode in list
    ├── upload_form.html    # Upload progress indicator
    └── settings_form.html  # Podcast settings
```

**HTMX Request Handling**:
```go
func (h *Handler) AddEpisode(w http.ResponseWriter, r *http.Request) {
    // Parse multipart form for file upload
    r.ParseMultipartForm(500 << 20) // 500MB limit
    
    // Validate and save episode
    episode := h.saveEpisode(r)
    
    // Return partial HTML for HTMX swap
    templates.ExecuteTemplate(w, "episode-row", episode)
}
```

**Common HTMX Patterns**:
```html
<!-- Add episode: POST and prepend to list -->
<form hx-post="/api/episodes" 
      hx-target="#episode-list" 
      hx-swap="afterbegin"
      hx-indicator="#upload-spinner">
    <input type="file" name="audio" accept=".mp3">
    <input type="text" name="title" placeholder="Episode Title">
    <button type="submit">Upload</button>
</form>

<!-- Delete episode: DELETE and remove from DOM -->
<button hx-delete="/api/episodes/ep123" 
        hx-target="closest .episode-row" 
        hx-swap="outerHTML"
        hx-confirm="Delete this episode?">
    Delete
</button>

<!-- Loading indicator -->
<div id="upload-spinner" class="htmx-indicator">
    Uploading...
</div>
```

**File Upload Handling**:
- Use standard `r.ParseMultipartForm()` for file uploads
- Return HTMX-friendly partial HTML after processing
- Use `HX-Redirect` response header for full page navigation after success
- Implement progress tracking via JavaScript/HTMX polling if needed

## 3. XML File Handling with Concurrency

### Decision

Use **atomic file writes** combined with **`sync.RWMutex`** for in-memory feed protection.

### Rationale

- RSS feeds are read-heavy, write-infrequent workloads (perfect for RWMutex)
- Atomic writes via temp file + rename prevent corruption from crashes
- RWMutex allows many concurrent readers while serializing writers
- Simple, reliable, and idiomatic Go pattern
- No external dependencies or OS-specific file locking

### Alternatives Considered

1. **File locking (flock)**
   - OS-specific implementation (different on Windows/Linux)
   - Adds complexity with syscalls
   - Not needed when in-memory caching is sufficient

2. **Database storage**
   - Overkill for single RSS file
   - Violates constitution's File-Centric Architecture principle
   - Unnecessary complexity

3. **No locking**
   - Risks file corruption with concurrent writes
   - Could lose episodes or corrupt XML
   - Unacceptable for production use

### Implementation Notes

**RSS Store with Safe Concurrency**:
```go
type RSSStore struct {
    mu       sync.RWMutex
    feed     *podcast.Podcast
    filepath string
}

// Read operations (many concurrent readers allowed)
func (s *RSSStore) GetFeed() *podcast.Podcast {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.feed // Return copy or immutable view
}

func (s *RSSStore) ServeXML(w http.ResponseWriter) error {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
    return s.feed.Encode(w)
}

// Write operations (exclusive lock, atomic file update)
func (s *RSSStore) UpdateFeed(p *podcast.Podcast) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Write to temp file first
    dir := filepath.Dir(s.filepath)
    tmpFile := filepath.Join(dir, ".podcast.xml.tmp")
    
    f, err := os.Create(tmpFile)
    if err != nil {
        return err
    }
    
    if err := p.Encode(f); err != nil {
        f.Close()
        os.Remove(tmpFile)
        return err
    }
    f.Close()
    
    // Atomic rename (replaces old file)
    if err := os.Rename(tmpFile, s.filepath); err != nil {
        return err
    }
    
    // Update in-memory copy only after successful write
    s.feed = p
    return nil
}

// Initialize on startup
func LoadRSSStore(filepath string) (*RSSStore, error) {
    data, err := os.ReadFile(filepath)
    if err != nil {
        // Create new empty feed if not exists
        return &RSSStore{
            feed:     createDefaultPodcast(),
            filepath: filepath,
        }, nil
    }
    
    // Parse existing feed
    p, err := podcast.Load(data)
    if err != nil {
        return nil, err
    }
    
    return &RSSStore{
        feed:     p,
        filepath: filepath,
    }, nil
}
```

**Key Patterns**:
- Keep feed in memory for fast reads (no file I/O per request)
- Use RWMutex for concurrent read access
- Write to temp file first, then atomic rename
- Reload from disk on server startup
- Update in-memory copy only after successful write

**Benefits**:
- Concurrent readers don't block each other
- Writes are serialized (prevents corruption)
- Atomic rename ensures file is never partially written
- Fast RSS feed serving (no disk I/O)

## 4. RSS Feed Validation

### Decision

Use **W3C Feed Validator API** programmatically for automated validation.

### Rationale

- Industry-standard validator trusted by RSS ecosystem
- Free HTTP API available at `https://validator.w3.org/feed/check.cgi`
- Validates both RSS 2.0 spec and common extensions
- Reliable and well-maintained by W3C
- Can validate by URL or raw XML content
- Returns structured error/warning data

### Alternatives Considered

1. **Cast Feed Validator (Podbase)**
   - Podcast-specific validation (iTunes tags)
   - Less accessible for automated testing
   - No official API for programmatic use
   - Better as manual tool for final checks

2. **Custom validation logic**
   - Incomplete coverage of RSS edge cases
   - Difficult to maintain spec compliance
   - Reinventing the wheel
   - Only justified for basic field validation

3. **Go validation libraries**
   - None found with comprehensive RSS 2.0 + iTunes validation
   - Would need to implement spec ourselves

### Implementation Notes

**Validation Client**:
```go
package rss

import (
    "bytes"
    "encoding/xml"
    "fmt"
    "net/http"
    "net/url"
)

type ValidationResult struct {
    Valid    bool
    Errors   []string
    Warnings []string
}

// Validate by posting feed URL
func ValidateFeedURL(feedURL string) (*ValidationResult, error) {
    resp, err := http.PostForm(
        "https://validator.w3.org/feed/check.cgi",
        url.Values{
            "url":    {feedURL},
            "output": {"soap12"}, // Machine-readable format
        },
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // Parse SOAP response for errors/warnings
    return parseValidationResponse(resp.Body)
}

// Validate raw XML content
func ValidateRawFeed(xmlContent []byte) (*ValidationResult, error) {
    req, err := http.NewRequest(
        "POST",
        "https://validator.w3.org/feed/check.cgi?rawdata=on&output=soap12",
        bytes.NewReader(xmlContent),
    )
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/xml")
    
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    return parseValidationResponse(resp.Body)
}

// Local validation (before calling external API)
func ValidateLocally(p *podcast.Podcast) []error {
    var errors []error
    
    // Check required fields
    if p.Title == "" {
        errors = append(errors, fmt.Errorf("title is required"))
    }
    if p.Link == "" {
        errors = append(errors, fmt.Errorf("link is required"))
    }
    if p.Description == "" {
        errors = append(errors, fmt.Errorf("description is required"))
    }
    
    // Validate URLs
    if _, err := url.Parse(p.Link); err != nil {
        errors = append(errors, fmt.Errorf("invalid link URL: %w", err))
    }
    
    // Validate episode enclosures
    for i, item := range p.Items {
        if item.Enclosure == nil {
            errors = append(errors, fmt.Errorf("episode %d missing enclosure", i))
        } else {
            if item.Enclosure.URL == "" {
                errors = append(errors, fmt.Errorf("episode %d enclosure missing URL", i))
            }
            if item.Enclosure.Length == 0 {
                errors = append(errors, fmt.Errorf("episode %d enclosure missing length", i))
            }
        }
    }
    
    return errors
}
```

**Testing Strategy**:
```go
func TestRSSFeedValidation(t *testing.T) {
    // Generate test feed
    feed := generateTestFeed()
    
    // Local validation first
    if errs := ValidateLocally(feed); len(errs) > 0 {
        t.Fatalf("Local validation failed: %v", errs)
    }
    
    // W3C validation
    xmlBytes, _ := feed.Bytes()
    result, err := ValidateRawFeed(xmlBytes)
    if err != nil {
        t.Fatalf("W3C validation request failed: %v", err)
    }
    
    if !result.Valid {
        t.Errorf("RSS feed validation failed:\nErrors: %v\nWarnings: %v", 
            result.Errors, result.Warnings)
    }
}
```

**Best Practices**:
1. Implement client-side validation first (required fields, valid URLs, etc.)
2. Use W3C validator as final check before publishing
3. Cache validation results to avoid API rate limits
4. Run validation in integration tests, not on every request
5. Consider adding podcast-specific validator (Cast Feed Validator) for manual iTunes compliance checks

## Technology Stack Summary

**Final Technology Decisions**:

| Category | Technology | Version | Justification |
|----------|-----------|---------|---------------|
| Language | Go | 1.21+ | Constitution requirement, excellent HTTP/XML support |
| HTTP Server | net/http (stdlib) | - | Constitution: avoid frameworks unless justified |
| RSS Generation | github.com/eduncan911/podcast | Latest | Production-proven, comprehensive iTunes support |
| Frontend | HTMX | 1.9+ | Dynamic UX without JavaScript framework complexity |
| Templating | html/template (stdlib) | - | Native Go, zero dependencies |
| Concurrency | sync.RWMutex | - | Read-heavy workload optimization |
| Storage | Filesystem + XML | - | Constitution: file-centric architecture |
| Validation | W3C Feed Validator API | - | Industry standard, free, reliable |

**Dependencies (go.mod)**:
```
module github.com/example/podcast-rss-server

go 1.21

require (
    github.com/eduncan911/podcast v1.4.2
)
```

**External Services**:
- W3C Feed Validator API (testing only)
- HTMX CDN (or self-hosted for production)

## Implementation Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| RSS validation failures | Feed rejected by podcast directories | Comprehensive test suite with W3C validator integration |
| Concurrent write corruption | Lost episodes or corrupted XML | Atomic writes + RWMutex pattern |
| Large file uploads (500MB) | Memory pressure, slow uploads | Stream file uploads directly to disk, multipart form handling |
| W3C API downtime | Test failures | Implement retry logic + local validation fallback |
| HTMX browser compatibility | Older browsers fail | Graceful degradation with standard forms |

## Next Steps (Phase 1)

1. **data-model.md**: Define RSS XML schema, Episode struct, Podcast metadata struct
2. **contracts/openapi.yaml**: Document REST API endpoints (POST /api/episodes, DELETE /api/episodes/{id}, GET /feed.xml)
3. **quickstart.md**: Document how to run server, upload first episode, access RSS feed
4. **Update agent context**: Add Go + HTMX to development environment context
