# Quickstart: Fix Relative URLs to Absolute URLs

**Feature**: 002-fix-relative-urls  
**Date**: 2025-12-31  
**Estimated Time**: 15-20 minutes

## Overview

This guide will help you implement the bug fix that converts relative URLs to absolute URLs in the RSS feed. You'll add base URL configuration, implement URL conversion logic, and verify the fix with tests.

---

## Prerequisites

- Go 1.21+ installed
- RSS server repository cloned and dependencies installed (`go mod download`)
- Familiarity with Go standard library (`net/url` package)
- Existing RSS feed serving relative URLs (the bug being fixed)

---

## Implementation Steps

### Step 1: Add Base URL Configuration (5 minutes)

**Goal**: Add `baseURL` field to configuration structure and validate at startup.

**Files to modify**:
- `config.yaml` - Add baseURL field
- `internal/config/config.go` - Create new package for configuration loading
- `cmd/server/main.go` - Load and validate configuration at startup

**Actions**:

1. **Create configuration package**:
   ```bash
   mkdir -p internal/config
   touch internal/config/config.go
   touch internal/config/config_test.go
   ```

2. **Add baseURL to config.yaml**:
   ```yaml
   # Add this field at the top of config.yaml
   baseURL: "http://localhost:8080"  # Change to your production URL
   ```

3. **Implement configuration loading** in `internal/config/config.go`:
   ```go
   package config
   
   import (
       "fmt"
       "net/url"
       "strings"
   )
   
   type Config struct {
       BaseURL string `yaml:"baseURL"`
       // ... existing fields
   }
   
   func (c *Config) Validate() error {
       // Check base URL is present
       if c.BaseURL == "" {
           return fmt.Errorf("baseURL is required in configuration")
       }
       
       // Parse and validate base URL
       parsedURL, err := url.Parse(c.BaseURL)
       if err != nil {
           return fmt.Errorf("invalid baseURL format: %w", err)
       }
       
       // Check protocol
       if parsedURL.Scheme == "" {
           return fmt.Errorf("baseURL missing protocol (http:// or https://)")
       }
       
       // Check hostname
       if parsedURL.Host == "" {
           return fmt.Errorf("baseURL missing hostname")
       }
       
       // Normalize: strip trailing slash
       c.BaseURL = strings.TrimRight(c.BaseURL, "/")
       
       return nil
   }
   ```

4. **Update main.go to validate config**:
   ```go
   // In cmd/server/main.go
   func main() {
       // Load configuration
       cfg, err := config.Load("config.yaml")
       if err != nil {
           log.Fatalf("FATAL: Failed to load configuration: %v", err)
       }
       
       // Validate configuration (including base URL)
       if err := cfg.Validate(); err != nil {
           log.Fatalf("FATAL: Invalid configuration: %v", err)
       }
       
       log.Printf("Configuration loaded successfully. Base URL: %s", cfg.BaseURL)
       
       // Continue with server startup...
   }
   ```

**Verification**:
```bash
# Test without baseURL - should fail
go run cmd/server/main.go
# Expected output: FATAL: Invalid configuration: baseURL is required

# Test with invalid baseURL - should fail
echo "baseURL: example.com" >> config.yaml  # Missing protocol
go run cmd/server/main.go
# Expected output: FATAL: Invalid configuration: baseURL missing protocol

# Test with valid baseURL - should start
echo "baseURL: http://localhost:8080" > config.yaml
go run cmd/server/main.go
# Expected output: Configuration loaded successfully. Base URL: http://localhost:8080
```

---

### Step 2: Implement URL Conversion Logic (10 minutes)

**Goal**: Create URL conversion functions that transform relative paths to absolute URLs with RFC 3986 encoding.

**Files to modify**:
- `internal/rss/feed.go` - Add URL conversion to RSS generation

**Actions**:

1. **Add URL conversion helper** in `internal/rss/feed.go`:
   ```go
   package rss
   
   import (
       "fmt"
       "log"
       "net/url"
   )
   
   // convertToAbsoluteURL converts a relative path to an absolute URL using the base URL
   func convertToAbsoluteURL(baseURL, relativePath string) (string, error) {
       if relativePath == "" {
           return "", fmt.Errorf("relative path is empty")
       }
       
       // URL-encode the path per RFC 3986
       // Note: We encode the path segments but preserve slashes
       segments := strings.Split(relativePath, "/")
       for i, segment := range segments {
           segments[i] = url.PathEscape(segment)
       }
       encodedPath := strings.Join(segments, "/")
       
       // Join base URL with encoded path
       absoluteURL, err := url.JoinPath(baseURL, encodedPath)
       if err != nil {
           return "", fmt.Errorf("failed to join URL: %w", err)
       }
       
       return absoluteURL, nil
   }
   ```

2. **Update RSS generation** to use absolute URLs:
   ```go
   // In GenerateFeed function in internal/rss/feed.go
   func GenerateFeed(p *models.Podcast, baseURL string) (string, error) {
       feed := &podcast.Podcast{
           Title:       p.Title,
           Link:        baseURL,  // Use base URL for channel link
           Description: p.Description,
           Language:    p.Language,
       }
       
       // Convert podcast artwork URL to absolute
       if p.ImageURL != "" {
           absoluteImageURL, err := convertToAbsoluteURL(baseURL, p.ImageURL)
           if err != nil {
               log.Printf("ERROR: Failed to convert podcast image URL: %v", err)
           } else {
               feed.IImage = &podcast.IImage{HREF: absoluteImageURL}
           }
       }
       
       // Convert episode URLs to absolute
       for _, ep := range p.Episodes {
           absoluteAudioURL, err := convertToAbsoluteURL(baseURL, ep.AudioURL)
           if err != nil {
               log.Printf("ERROR: Failed to convert audio URL for episode %s: %v (skipping)", ep.Title, err)
               continue  // Skip episodes with malformed URLs
           }
           
           item := podcast.Item{
               Title:       ep.Title,
               Description: ep.Description,
               PubDate:     &ep.PubDate,
           }
           item.AddEnclosure(absoluteAudioURL, podcast.MP3, ep.AudioLength)
           
           if _, err := feed.AddItem(item); err != nil {
               log.Printf("ERROR: Failed to add item to feed: %v", err)
           }
       }
       
       return feed.String()
   }
   ```

3. **Update RSS handler** to pass base URL:
   ```go
   // In internal/handlers/feed.go
   func FeedHandler(cfg *config.Config) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           // Load podcast metadata
           podcast, err := storage.LoadPodcast(cfg.Paths.RSSFile)
           if err != nil {
               http.Error(w, "Failed to load podcast", http.StatusInternalServerError)
               return
           }
           
           // Generate RSS feed with absolute URLs
           feedXML, err := rss.GenerateFeed(podcast, cfg.BaseURL)
           if err != nil {
               http.Error(w, "Failed to generate feed", http.StatusInternalServerError)
               return
           }
           
           w.Header().Set("Content-Type", "application/rss+xml")
           w.Write([]byte(feedXML))
       }
   }
   ```

**Verification**:
```bash
# Start server
go run cmd/server/main.go

# Check RSS feed output
curl http://localhost:8080/feed.xml | grep -o 'url="[^"]*"'
# Expected: url="http://localhost:8080/audio/episode.mp3" (absolute URL)
# Not: url="/audio/episode.mp3" (relative URL)
```

---

### Step 3: Update Web Dashboard (5 minutes)

**Goal**: Display absolute feed URL on the dashboard using the configured base URL.

**Files to modify**:
- `internal/handlers/web.go` - Pass base URL to template
- `web/templates/index.html` - Display feed URL from config

**Actions**:

1. **Update web handler** in `internal/handlers/web.go`:
   ```go
   func WebHandler(cfg *config.Config) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           // Load podcast metadata
           podcast, err := storage.LoadPodcast(cfg.Paths.RSSFile)
           if err != nil {
               // Handle error...
           }
           
           data := struct {
               Podcast *models.Podcast
               FeedURL string  // Use configured base URL
           }{
               Podcast: podcast,
               FeedURL: cfg.BaseURL + "/feed.xml",  // Absolute URL from config
           }
           
           tmpl.Execute(w, data)
       }
   }
   ```

2. **Verify template** in `web/templates/index.html`:
   ```html
   <!-- Should already display {{.FeedURL}}, just verify it uses config -->
   <p>RSS Feed: <a href="{{.FeedURL}}">{{.FeedURL}}</a></p>
   ```

**Verification**:
```bash
# Access dashboard
open http://localhost:8080

# Check displayed feed URL
# Expected: "RSS Feed: http://localhost:8080/feed.xml"
# Not: "RSS Feed: http://<request-hostname>/feed.xml"
```

---

### Step 4: Add Tests (10 minutes)

**Goal**: Verify URL conversion, encoding, and RSS validation with automated tests.

**Files to create**:
- `tests/integration/rss_validator_test.go`
- `tests/unit/config_test.go`

**Actions**:

1. **Create unit tests for configuration**:
   ```go
   // tests/unit/config_test.go
   package config_test
   
   import (
       "testing"
       "your-module/internal/config"
   )
   
   func TestConfigValidation_MissingBaseURL(t *testing.T) {
       cfg := &config.Config{BaseURL: ""}
       err := cfg.Validate()
       if err == nil {
           t.Fatal("expected error for missing baseURL")
       }
   }
   
   func TestConfigValidation_MissingProtocol(t *testing.T) {
       cfg := &config.Config{BaseURL: "example.com"}
       err := cfg.Validate()
       if err == nil {
           t.Fatal("expected error for missing protocol")
       }
   }
   
   func TestConfigValidation_TrailingSlash(t *testing.T) {
       cfg := &config.Config{BaseURL: "http://example.com/"}
       err := cfg.Validate()
       if err != nil {
           t.Fatalf("unexpected error: %v", err)
       }
       // Verify trailing slash stripped
       if cfg.BaseURL != "http://example.com" {
           t.Errorf("expected normalized URL, got %s", cfg.BaseURL)
       }
   }
   ```

2. **Create integration tests for RSS feed**:
   ```go
   // tests/integration/rss_validator_test.go
   package integration_test
   
   import (
       "strings"
       "testing"
       "your-module/internal/rss"
       "your-module/internal/models"
   )
   
   func TestRSSFeed_AbsoluteURLs(t *testing.T) {
       podcast := &models.Podcast{
           Title: "Test Podcast",
           ImageURL: "/static/artwork/test.jpg",
           Episodes: []models.Episode{
               {
                   Title: "Episode 1",
                   AudioURL: "/audio/episode1.mp3",
               },
           },
       }
       
       baseURL := "http://test.example.com"
       feedXML, err := rss.GenerateFeed(podcast, baseURL)
       if err != nil {
           t.Fatalf("failed to generate feed: %v", err)
       }
       
       // Verify absolute URLs present
       if !strings.Contains(feedXML, "http://test.example.com/audio/episode1.mp3") {
           t.Error("expected absolute audio URL in feed")
       }
       
       // Verify no relative URLs
       if strings.Contains(feedXML, `url="/"`) {
           t.Error("feed contains relative URLs starting with /")
       }
   }
   
   func TestRSSFeed_URLEncoding(t *testing.T) {
       podcast := &models.Podcast{
           Episodes: []models.Episode{
               {Title: "Test", AudioURL: "/audio/My Episode.mp3"},
           },
       }
       
       feedXML, _ := rss.GenerateFeed(podcast, "http://test.com")
       
       // Verify spaces encoded as %20
       if !strings.Contains(feedXML, "My%20Episode.mp3") {
           t.Error("expected URL-encoded spaces in feed")
       }
   }
   ```

3. **Run tests**:
   ```bash
   go test ./tests/unit/...
   go test ./tests/integration/...
   ```

**Verification**:
```bash
# All tests should pass
go test ./... -v
# Expected: PASS for all URL conversion and validation tests
```

---

## Testing the Fix

### Manual Validation

1. **Start server with configuration**:
   ```bash
   # Edit config.yaml with your base URL
   echo "baseURL: http://podcast.example.com" > config.yaml
   go run cmd/server/main.go
   ```

2. **Upload a test episode** (if needed):
   ```bash
   curl -F "file=@test.mp3" -F "title=Test Episode" http://localhost:8080/api/episodes
   ```

3. **Fetch RSS feed and verify absolute URLs**:
   ```bash
   curl http://localhost:8080/feed.xml > feed.xml
   cat feed.xml | grep 'enclosure url='
   # Expected output: enclosure url="http://podcast.example.com/audio/..."
   ```

4. **Validate with RSS validators**:
   - Apple Podcasts Connect: https://podcastsconnect.apple.com/
   - W3C Feed Validator: https://validator.w3.org/feed/
   - Cast Feed Validator: https://castfeedvalidator.com/

### Automated Testing

```bash
# Run all tests
go test ./... -v

# Run only URL-related tests
go test ./tests/integration -run TestRSSFeed

# Check test coverage
go test ./internal/rss -cover
```

---

## Deployment Checklist

- [ ] Add `baseURL` to production `config.yaml`
- [ ] Verify base URL uses HTTPS for production
- [ ] Run automated tests (`go test ./...`)
- [ ] Validate RSS feed with Apple Podcasts Connect validator
- [ ] Validate RSS feed with W3C Feed Validator
- [ ] Test with at least one podcast client (Apple Podcasts, Spotify)
- [ ] Verify artwork displays in podcast client
- [ ] Verify audio playback works in podcast client
- [ ] Check logs for any URL conversion errors
- [ ] Document base URL configuration in README

---

## Troubleshooting

### Server won't start

**Error**: `FATAL: Invalid configuration: baseURL is required`

**Solution**: Add `baseURL` field to `config.yaml`:
```yaml
baseURL: "http://your-domain.com"
```

---

### RSS feed still shows relative URLs

**Problem**: Feed shows `/audio/file.mp3` instead of absolute URL

**Possible causes**:
1. Base URL not being passed to RSS generation function
2. URL conversion logic not applied to all URL fields
3. Caching (old feed cached by browser/client)

**Solution**:
```bash
# Clear cache and regenerate
curl -H "Cache-Control: no-cache" http://localhost:8080/feed.xml
# Verify all <enclosure> and <itunes:image> tags have absolute URLs
```

---

### Special characters not encoded

**Problem**: File with spaces shows `My Episode.mp3` instead of `My%20Episode.mp3`

**Solution**: Verify `url.PathEscape()` is used in `convertToAbsoluteURL()`:
```go
segments[i] = url.PathEscape(segment)  // Must be PathEscape, not QueryEscape
```

---

### Episodes missing from feed

**Problem**: Some episodes don't appear in RSS feed

**Possible cause**: URL conversion failing for certain episodes (malformed paths)

**Solution**: Check logs for conversion errors:
```bash
# Look for "Failed to convert audio URL" messages
grep "Failed to convert" server.log
```

Fix the malformed paths or filenames causing conversion failures.

---

## Next Steps

After implementing this fix:

1. **Update documentation**:
   - Add `baseURL` configuration to README
   - Document URL encoding behavior for special characters
   - Note server restart required for base URL changes

2. **Monitor in production**:
   - Check logs for URL conversion errors
   - Verify podcast clients successfully fetch audio files
   - Monitor RSS validator results periodically

3. **Consider future enhancements** (out of scope for this fix):
   - CDN URL support (different base URL for media files)
   - Automatic base URL detection from request headers
   - Hot-reload of base URL without restart

---

## Summary

You've successfully implemented the URL fix by:
1. ✅ Adding required `baseURL` configuration with validation
2. ✅ Implementing URL conversion with RFC 3986 encoding
3. ✅ Updating RSS feed generation to use absolute URLs
4. ✅ Updating web dashboard to display absolute feed URL
5. ✅ Adding automated tests for validation

The RSS feed now generates absolute URLs that work correctly in all podcast clients and pass RSS validator checks.
