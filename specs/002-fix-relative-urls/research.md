# Research: Fix Relative URLs to Absolute URLs

**Feature**: 002-fix-relative-urls  
**Date**: 2025-12-31  
**Status**: Complete

## Research Questions

### 1. URL Parsing and Validation in Go

**Question**: What is the best approach for parsing, validating, and manipulating URLs in Go, specifically for base URL validation and URL joining?

**Decision**: Use `net/url` package from Go standard library

**Rationale**:
- `net/url.Parse()` handles protocol validation, hostname extraction, port parsing
- `url.URL` struct provides structured access to protocol, host, path components
- `url.JoinPath()` (Go 1.19+) correctly handles path joining with proper slash normalization
- `url.QueryEscape()` and `url.PathEscape()` provide RFC 3986 compliant encoding
- Standard library ensures no external dependencies for critical URL operations
- Well-tested and maintained as part of Go core

**Alternatives Considered**:
- **Custom URL builder**: Rejected - reinventing wheel, prone to edge case bugs
- **Third-party URL library**: Rejected - unnecessary dependency for standard operations
- **String concatenation**: Rejected - error-prone for trailing slashes, encoding, path joining

**Code Pattern**:
```go
// Base URL validation at startup
baseURLStr := config.BaseURL
parsedURL, err := url.Parse(baseURLStr)
if err != nil {
    return fmt.Errorf("invalid base URL: %w", err)
}
if parsedURL.Scheme == "" {
    return fmt.Errorf("base URL missing protocol (http:// or https://)")
}
if parsedURL.Host == "" {
    return fmt.Errorf("base URL missing hostname")
}

// URL joining with path
absoluteURL, err := url.JoinPath(baseURL, relativePath)
if err != nil {
    return "", err
}

// URL encoding for special characters
encodedPath := url.PathEscape(relativePath)
```

---

### 2. RFC 3986 URL Encoding Requirements

**Question**: What characters require URL encoding per RFC 3986, and how should special characters in file paths (spaces, Unicode) be handled?

**Decision**: Use `url.PathEscape()` for path components, which encodes per RFC 3986

**Rationale**:
- RFC 3986 defines unreserved characters: `A-Z a-z 0-9 - . _ ~`
- All other characters (including spaces, Unicode) must be percent-encoded
- `url.PathEscape()` correctly encodes path segments (spaces → `%20`, not `+`)
- `url.QueryEscape()` is for query parameters only (spaces → `+`)
- Podcast clients expect path encoding (e.g., `My%20Episode.mp3`)

**Alternatives Considered**:
- **Replace spaces with underscores**: Rejected - changes original filenames, not reversible
- **Only encode spaces**: Rejected - insufficient for Unicode, special chars like `#`, `?`
- **url.QueryEscape()**: Rejected - incorrect encoding for paths (uses `+` for spaces)

**Encoding Examples**:
```
Original: "My Episode 1.mp3" → Encoded: "My%20Episode%201.mp3"
Original: "Café Talk.mp3" → Encoded: "Caf%C3%A9%20Talk.mp3"
Original: "Q&A Session.mp3" → Encoded: "Q%26A%20Session.mp3"
```

---

### 3. Configuration Validation Strategy

**Question**: Should base URL validation happen at startup or lazily on first RSS request? What validation checks are necessary?

**Decision**: Fail-fast validation at startup with comprehensive checks

**Rationale**:
- Fail-fast prevents serving broken RSS feeds to podcast clients
- Startup validation provides immediate feedback to administrators
- Configuration errors are detected before any requests are served
- Avoids partial availability (web dashboard works but RSS feed broken)
- Aligns with Constitution principle V (Simplicity) - no complex fallback logic

**Validation Checks**:
1. Base URL field present (required, not optional)
2. Valid URL format (parseable by `net/url.Parse()`)
3. Protocol present (`http://` or `https://`)
4. Hostname present (not empty after parsing)
5. Trailing slash normalization (strip trailing `/` to prevent double slashes)

**Alternatives Considered**:
- **Lazy validation on first request**: Rejected - allows server to start in broken state
- **Default to `http://localhost:8080`**: Rejected - incorrect for production, per clarifications
- **Runtime validation on each request**: Rejected - performance overhead, startup is sufficient

**Error Message Example**:
```
FATAL: Invalid base URL configuration
  Provided: "example.com"
  Error: base URL missing protocol (http:// or https://)
  Expected format: http://example.com or https://example.com
  Fix: Update baseURL in config.yaml and restart server
```

---

### 4. RSS Feed URL Conversion Approach

**Question**: Should URL conversion happen during storage (when episodes are uploaded) or during RSS feed generation?

**Decision**: On-the-fly conversion during RSS feed generation

**Rationale**:
- Maintains file-centric architecture (Constitution II) - storage remains portable
- Allows base URL changes without migrating stored data
- Relative paths in storage are deployment-agnostic (works across environments)
- Conversion logic centralized in RSS generation layer (`internal/rss/feed.go`)
- Aligns with spec requirement FR-009: "store only relative paths internally"

**Conversion Flow**:
1. Episode uploaded → Store relative path `/audio/filename.mp3` in `podcast.xml`
2. RSS request received → Load podcast metadata with relative paths
3. RSS generation → Convert each relative path to absolute URL using base URL
4. Feed served → Contains absolute URLs like `http://example.com/audio/filename.mp3`

**Alternatives Considered**:
- **Store absolute URLs**: Rejected - breaks portability, requires data migration on base URL change
- **Dual storage (relative + absolute)**: Rejected - unnecessary duplication, sync issues
- **Lazy conversion with caching**: Rejected - premature optimization, violates Constitution V

---

### 5. Error Handling for Malformed Paths

**Question**: How should the system handle edge cases where URL conversion fails (malformed paths, invalid characters)?

**Decision**: Skip problematic episodes, log error, continue feed generation with valid episodes

**Rationale**:
- Per clarification: "Skip problematic episodes, include only valid episodes in feed"
- Prevents total RSS feed failure due to single bad episode
- Podcast clients receive partial feed (better than complete failure)
- Logged errors enable administrator investigation and fix
- Aligns with graceful degradation pattern

**Error Handling Pattern**:
```go
for _, episode := range episodes {
    absoluteURL, err := convertToAbsoluteURL(baseURL, episode.AudioURL)
    if err != nil {
        log.Printf("ERROR: Failed to convert URL for episode %s: %v (skipping)", episode.Title, err)
        continue // Skip this episode
    }
    // Add episode to RSS feed with absolute URL
}
```

**Alternatives Considered**:
- **Return HTTP 500 error**: Rejected - per clarifications, breaks entire feed
- **Return empty feed**: Rejected - per clarifications, better to show partial content
- **Use relative URL as fallback**: Rejected - defeats purpose of fix, still broken in clients

---

### 6. Testing Strategy with RSS Validators

**Question**: How should automated RSS validation be integrated into the test suite?

**Decision**: Integration tests using HTTP requests to feed endpoint + external validator libraries

**Rationale**:
- Test real RSS output, not mocked responses
- Validates end-to-end URL conversion (config → storage → generation → output)
- RSS validator libraries available for Go (e.g., `github.com/mmcdole/gofeed`)
- Apple Podcasts validator can be tested via their public API endpoint
- Tests prevent regressions in URL format during future changes

**Test Structure**:
```go
func TestRSSFeedAbsoluteURLs(t *testing.T) {
    // Setup: Configure base URL, upload test episode
    config := &Config{BaseURL: "http://test.example.com"}
    
    // Generate RSS feed
    feed := generateRSSFeed(config, podcast)
    
    // Validate: All URLs are absolute
    assert.Contains(t, feed, "http://test.example.com/audio/")
    assert.NotContains(t, feed, `<enclosure url="/"`)
    
    // Validate: RSS 2.0 compliance
    fp := gofeed.NewParser()
    parsed, err := fp.ParseString(feed)
    require.NoError(t, err)
    
    // Validate: Absolute URLs in parsed feed
    for _, item := range parsed.Items {
        assert.True(t, strings.HasPrefix(item.Enclosures[0].URL, "http://"))
    }
}
```

**Test Coverage Required**:
- Base URL with HTTP vs HTTPS
- Base URL with subdirectory path (`http://example.com/podcasts/`)
- File paths with spaces (URL encoding validation)
- File paths with Unicode characters
- Missing base URL (startup failure)
- Invalid base URL format (startup failure)

---

## Research Summary

All technical unknowns have been resolved:

1. **URL Operations**: `net/url` standard library for parsing, validation, joining, encoding
2. **Encoding Standard**: RFC 3986 via `url.PathEscape()` for path components
3. **Validation Strategy**: Fail-fast at startup with comprehensive checks
4. **Conversion Timing**: On-the-fly during RSS generation (not at storage)
5. **Error Handling**: Skip malformed episodes, log errors, continue with valid episodes
6. **Testing Approach**: Integration tests with RSS validators + end-to-end URL verification

**No external dependencies required** - all functionality available in Go standard library.

**Next Phase**: Proceed to Phase 1 (Design & Contracts) with research complete.
