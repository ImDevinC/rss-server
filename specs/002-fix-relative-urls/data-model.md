# Data Model: Fix Relative URLs to Absolute URLs

**Feature**: 002-fix-relative-urls  
**Date**: 2025-12-31

## Overview

This bug fix introduces minimal data model changes - primarily adding base URL configuration and defining URL conversion behavior. Existing entities (Podcast, Episode) maintain their current structure with relative path storage.

---

## Entities

### Configuration (New)

**Description**: Application configuration including the newly required base URL field.

**Attributes**:
- `BaseURL` (string, required): The publicly accessible base URL for the podcast server
  - Format: `{protocol}://{hostname}[:{port}][/{path}]`
  - Examples: `http://localhost:8080`, `https://podcast.example.com`, `https://example.com/podcasts`
  - Validation: Must include protocol (http/https), must include hostname, trailing slashes stripped
- `Server.Port` (int): HTTP server port (existing field)
- `Server.Host` (string): HTTP server bind address (existing field)
- `Paths.AudioDir` (string): Audio file storage directory (existing field)
- `Paths.ArtworkDir` (string): Artwork storage directory (existing field)
- `Upload.MaxFileSizeMB` (int): Maximum upload size (existing field)

**Validation Rules**:
- `BaseURL` is **mandatory** - server will not start if missing
- `BaseURL` must parse successfully via `url.Parse()`
- `BaseURL` must have non-empty `Scheme` (http or https)
- `BaseURL` must have non-empty `Host`
- Trailing slashes are automatically stripped during normalization

**Storage**: `config.yaml` file at repository root

**Example**:
```yaml
# config.yaml
baseURL: "https://podcast.example.com"  # NEW FIELD (required)

server:
  port: 8080
  host: "0.0.0.0"

paths:
  data_dir: "./data"
  audio_dir: "./data/audio"
  artwork_dir: "./data/artwork"
```

---

### Podcast (Modified)

**Description**: Represents podcast-level metadata. **No structural changes** - continues to store relative paths.

**Attributes** (unchanged):
- `Title` (string): Podcast title
- `Author` (string): Podcast author name
- `Description` (string): Podcast description
- `ImageURL` (string): **Relative path** to podcast artwork (e.g., `/static/artwork/image.jpg`)
  - ⚠️ **Stored as relative path, converted to absolute during RSS generation**
- `Language` (string): Podcast language code
- `Category` (string): iTunes category
- `Episodes` ([]Episode): List of episodes

**URL Conversion Behavior**:
- **Storage**: `ImageURL` remains relative path in `podcast.xml`
- **RSS Output**: Converted to absolute URL like `https://podcast.example.com/static/artwork/image.jpg`
- **Conversion Logic**: Applied in `internal/rss/feed.go` during RSS generation

**No migration required** - existing relative paths work without changes.

---

### Episode (Modified)

**Description**: Represents a single podcast episode. **No structural changes** - continues to store relative paths.

**Attributes** (unchanged):
- `Title` (string): Episode title
- `Description` (string): Episode description
- `AudioURL` (string): **Relative path** to audio file (e.g., `/audio/episode.mp3`)
  - ⚠️ **Stored as relative path, converted to absolute during RSS generation**
- `AudioLength` (int64): Audio file size in bytes
- `PubDate` (time.Time): Publication date
- `Duration` (string): Episode duration

**URL Conversion Behavior**:
- **Storage**: `AudioURL` remains relative path in `podcast.xml`
- **RSS Output**: Converted to absolute URL like `https://podcast.example.com/audio/episode.mp3`
- **Conversion Logic**: Applied in `internal/rss/feed.go` during RSS generation
- **Encoding**: Special characters URL-encoded per RFC 3986 (spaces → `%20`)

**No migration required** - existing relative paths work without changes.

---

### URL Conversion (New Concept)

**Description**: Logical entity representing the conversion process from relative path to absolute URL. Not stored, exists only during RSS feed generation.

**Attributes**:
- `BaseURL` (string): Configured base URL from application config
- `RelativePath` (string): Input relative path from Podcast/Episode entity
- `AbsoluteURL` (string): Output absolute URL for RSS feed

**Conversion Rules**:
1. Parse base URL and validate (already done at startup)
2. URL-encode relative path per RFC 3986 using `url.PathEscape()`
3. Join base URL and encoded path using `url.JoinPath()`
4. Return absolute URL or error if conversion fails

**Conversion Algorithm**:
```
Input: baseURL="https://podcast.example.com", relativePath="/audio/My Episode.mp3"

Step 1: Encode path
  "/audio/My Episode.mp3" → "/audio/My%20Episode.mp3"

Step 2: Join with base URL
  url.JoinPath("https://podcast.example.com", "/audio/My%20Episode.mp3")
  → "https://podcast.example.com/audio/My%20Episode.mp3"

Output: "https://podcast.example.com/audio/My%20Episode.mp3"
```

**Error Cases**:
- Malformed relative path → Skip episode, log error, continue feed generation
- Empty base URL → Caught at startup, server fails to start
- Path with invalid characters → URL encoding handles most cases, skip if encoding fails

---

## State Transitions

### Configuration Loading State Machine

```
[Server Startup]
    |
    v
[Load config.yaml]
    |
    +-- baseURL missing? --> [FATAL ERROR: Exit with message]
    |
    +-- baseURL invalid format? --> [FATAL ERROR: Exit with message]
    |
    +-- baseURL valid? --> [Normalize (strip trailing slash)]
                              |
                              v
                         [Store in memory]
                              |
                              v
                         [Start HTTP server]
```

**States**:
1. **Uninitialized**: Server starting, config not loaded
2. **Loading**: Reading and parsing `config.yaml`
3. **Validating**: Checking base URL format and requirements
4. **Normalized**: Trailing slashes stripped, URL ready for use
5. **Active**: Config in memory, server running
6. **Error**: Validation failed, server exited with error message

**Transitions**:
- Uninitialized → Loading: Startup begins
- Loading → Validating: Config file parsed successfully
- Validating → Normalized: Base URL passes all validation checks
- Validating → Error: Base URL missing or invalid
- Normalized → Active: Server starts accepting requests
- Error → (terminal state): Server exits, admin must fix config

---

## Relationships

```
┌─────────────────┐
│  Configuration  │  1:1 relationship with server instance
└────────┬────────┘
         │ provides baseURL to
         │
         v
┌─────────────────┐
│  URL Converter  │  Logical conversion layer
└────────┬────────┘
         │ converts paths from
         │
         v
┌─────────────────┐       1:N       ┌─────────────┐
│    Podcast      │─────────────────│   Episode   │
│  (ImageURL)     │                 │ (AudioURL)  │
└─────────────────┘                 └─────────────┘
         │                                 │
         │ relative paths stored           │
         v                                 v
   podcast.xml (filesystem storage)
         │
         │ loaded during RSS generation
         v
┌──────────────────────┐
│   RSS Feed Output    │  Contains absolute URLs
│  (HTTP response)     │
└──────────────────────┘
```

**Key Relationships**:
1. Configuration → URL Converter: 1:1, base URL provided for all conversions
2. Podcast → Episodes: 1:N, unchanged relationship
3. URL Converter → Podcast/Episode: Uses relative paths from entities
4. URL Converter → RSS Feed: Produces absolute URLs for output
5. Configuration → Storage: Configuration loaded from `config.yaml` file

---

## Data Migration

**Migration Required**: ❌ **NO**

**Rationale**:
- Existing `podcast.xml` files continue to work unchanged
- Relative paths remain in storage (no data transformation needed)
- URL conversion happens on-the-fly during RSS generation
- Only requirement: Administrator must add `baseURL` to `config.yaml`

**Deployment Steps**:
1. Deploy updated code
2. Add `baseURL: "https://your-domain.com"` to `config.yaml`
3. Restart server
4. Server validates base URL at startup
5. RSS feed now contains absolute URLs (no data migration needed)

**Rollback**:
- Relative paths still stored, so rollback is safe
- Remove `baseURL` from config and deploy previous version
- RSS feed returns to relative URLs (though this is the buggy behavior)

---

## Validation Rules Summary

| Field | Required | Validation | Error Handling |
|-------|----------|------------|----------------|
| `Configuration.BaseURL` | ✅ Yes | Must parse, have protocol and hostname | Fatal error at startup |
| `Podcast.ImageURL` | ❌ No | Relative path | Convert to absolute or use default |
| `Episode.AudioURL` | ✅ Yes | Relative path | Convert to absolute or skip episode |

---

## Storage Format

### config.yaml (Modified)

```yaml
# NEW FIELD (required)
baseURL: "https://podcast.example.com"

# Existing fields unchanged
server:
  port: 8080
  host: "0.0.0.0"

upload:
  max_file_size_mb: 500
  allowed_extensions:
    - ".mp3"

paths:
  data_dir: "./data"
  audio_dir: "./data/audio"
  artwork_dir: "./data/artwork"
  rss_file: "./data/podcast.xml"

podcast:
  default_title: "My Podcast"
  default_author: "Podcast Creator"
```

### podcast.xml (Unchanged)

```xml
<!-- Storage format remains unchanged - continues to store relative paths -->
<podcast>
  <title>My Podcast</title>
  <imageURL>/static/artwork/podcast.jpg</imageURL>  <!-- Relative path -->
  <episodes>
    <episode>
      <title>Episode 1</title>
      <audioURL>/audio/episode1.mp3</audioURL>  <!-- Relative path -->
      <audioLength>5242880</audioLength>
    </episode>
  </episodes>
</podcast>
```

**Note**: Relative paths in storage are converted to absolute URLs only during RSS feed generation, keeping storage portable and deployment-agnostic.
