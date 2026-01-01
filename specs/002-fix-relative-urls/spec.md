# Feature Specification: Fix Relative URLs to Absolute URLs

**Feature Branch**: `002-fix-relative-urls`  
**Created**: 2025-12-31  
**Status**: Draft  
**Input**: User description: "I have identified a bug in the application. All URL's should use the full URL, not relative URL paths. Please update them accordingly"

## Clarifications

### Session 2025-12-31

- Q: What level of logging should the system provide for URL conversion operations? → A: Minimal: Log only critical errors (startup validation failures)
- Q: What should happen when RSS feed generation fails due to URL construction errors (e.g., malformed relative path)? → A: Skip problematic episodes, include only valid episodes in feed
- Q: When the base URL is missing or invalid at startup, should the server start with a default or fail to start? → A: Fail to start with clear error message requiring base URL configuration
- Q: Should URL path normalization handle special characters (spaces, Unicode) in file paths? → A: Yes: URL-encode special characters (spaces become %20, etc.) per RFC 3986
- Q: Should the system validate that generated absolute URLs are reachable before including them in the RSS feed? → A: No: Assume URLs are reachable if format is valid (trust configured base URL)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - RSS Feed Validates with Absolute URLs (Priority: P1)

As a podcast creator, I need my RSS feed to contain absolute URLs (with full domain) instead of relative paths, so that podcast directories and clients can properly fetch my audio files and artwork.

**Why this priority**: This is a critical bug fix. Without absolute URLs, the RSS feed is non-functional in podcast clients like Apple Podcasts, Spotify, and other aggregators. Relative URLs like `/audio/file.mp3` cannot be resolved without a hostname, making the podcast unplayable. This blocks the entire core value proposition of the application.

**Independent Test**: Can be fully tested by uploading an episode with artwork, accessing the RSS feed at `/feed.xml`, and verifying that all URLs (audio enclosures, podcast artwork, episode images) contain absolute URLs starting with the configured base URL (e.g., `http://example.com/audio/file.mp3` instead of `/audio/file.mp3`).

**Acceptance Scenarios**:

1. **Given** I have configured a base URL of `http://podcast.example.com` and uploaded an episode with audio, **When** I access the RSS feed XML, **Then** the audio enclosure URL is `http://podcast.example.com/audio/filename.mp3` (absolute URL with full domain)
2. **Given** I have uploaded podcast artwork, **When** I access the RSS feed XML, **Then** the podcast image URL in `<itunes:image>` tag contains the full absolute URL like `http://podcast.example.com/static/artwork/image.jpg`
3. **Given** I paste my RSS feed into Apple Podcasts Connect validator, **When** the validator processes the feed, **Then** it successfully fetches and displays the podcast artwork and episode audio files without broken links
4. **Given** I subscribe to my podcast feed in a podcast client app, **When** the app attempts to play an episode, **Then** it successfully streams the audio file using the absolute URL from the feed

---

### User Story 2 - Configure Base URL for Deployment (Priority: P1)

As a system administrator deploying the podcast server, I need to configure the base URL (hostname and protocol) in the application configuration, so the system knows what absolute URLs to generate for my specific deployment environment.

**Why this priority**: This is essential infrastructure for the bug fix. Different deployments will have different hostnames (localhost for development, podcast.example.com for production, etc.). Without configurable base URL, the system cannot generate correct absolute URLs for the deployment environment.

**Independent Test**: Can be tested by editing the configuration file to specify a base URL, restarting the server, uploading an episode, and verifying the RSS feed contains URLs starting with the configured base URL.

**Acceptance Scenarios**:

1. **Given** I set `baseURL: "http://localhost:8080"` in the configuration file, **When** I start the server and upload an episode, **Then** all generated URLs start with `http://localhost:8080`
2. **Given** I set `baseURL: "https://podcast.example.com"` in production configuration, **When** I upload an episode, **Then** all generated URLs use HTTPS protocol and the production domain
3. **Given** I do not configure a base URL in the configuration file, **When** I start the server, **Then** the server fails to start with a clear error message indicating that base URL is required
4. **Given** I configure an invalid base URL format (missing protocol or trailing slash), **When** I start the server, **Then** it validates the format and fails to start with a clear error message explaining the expected format

---

### User Story 3 - Web Dashboard Shows Correct Feed URL (Priority: P2)

As a podcast creator viewing the web dashboard, I want to see the RSS feed URL displayed using the configured base URL, so I know the correct URL to submit to podcast directories.

**Why this priority**: This improves usability by showing users the correct feed URL to copy and submit to directories. While the RSS feed functionality (P1) is the critical fix, the dashboard display helps prevent user confusion about which URL to use.

**Independent Test**: Can be tested by accessing the web dashboard and verifying the displayed RSS feed URL matches the configured base URL (e.g., shows `http://podcast.example.com/feed.xml` when that's the configured base).

**Acceptance Scenarios**:

1. **Given** I have configured base URL as `http://podcast.example.com`, **When** I view the web dashboard, **Then** the RSS feed URL displayed is `http://podcast.example.com/feed.xml`
2. **Given** I am running the server locally with base URL `http://localhost:8080`, **When** I view the dashboard, **Then** the feed URL shows the local address
3. **Given** I copy the RSS feed URL from the dashboard, **When** I paste it into a podcast directory submission form, **Then** the URL is correctly formatted and accessible

---

### Edge Cases

- What happens when base URL is configured without a protocol (e.g., just "example.com")? The system validates on startup and requires a full URL with protocol (http:// or https://).
- What happens when base URL has a trailing slash (e.g., "http://example.com/")? The system normalizes by removing trailing slashes to prevent double-slash URLs like `http://example.com//audio/file.mp3`.
- What happens when an episode is uploaded before base URL is configured? The relative URL is stored initially, but should be converted to absolute URL on-the-fly during RSS feed generation using current base URL configuration.
- What happens when base URL is changed after episodes are already uploaded? All existing relative URLs are converted using the new base URL during RSS feed generation (URLs are not stored in database, only paths).
- What happens when default podcast artwork URL is used? The system converts the default relative path to an absolute URL using the configured base URL.
- What happens when the RSS feed is accessed directly by IP address instead of domain? The feed still contains absolute URLs based on the configured base URL, not the request hostname.
- What happens if someone configures base URL with a path component (e.g., "http://example.com/podcasts")? The system supports this and correctly generates URLs like `http://example.com/podcasts/audio/file.mp3`.
- What happens when RSS feed generation encounters a malformed relative path that cannot be converted to an absolute URL? The system skips that specific episode in the RSS feed, includes only episodes with valid URLs, and logs the error for administrator investigation.
- What happens when file paths contain special characters (spaces, Unicode characters)? The system URL-encodes special characters per RFC 3986 (e.g., spaces become %20) to ensure URLs are valid and accessible in podcast clients.
- What happens if a generated absolute URL points to a non-existent or unreachable resource? The system includes the URL in the RSS feed without reachability validation, trusting that the configured base URL and stored file paths are correct. Validation of resource availability is out of scope.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support configuration of a base URL (protocol, hostname, and optional port) in the application configuration file
- **FR-002**: System MUST validate base URL format on startup and fail to start with clear error message if invalid (missing protocol, malformed URL) or missing (URL reachability validation is not required)
- **FR-003**: System MUST normalize base URL by removing trailing slashes to prevent malformed URLs
- **FR-004**: System MUST URL-encode special characters in file paths (spaces, Unicode) per RFC 3986 when converting to absolute URLs
- **FR-005**: System MUST generate absolute URLs for all audio file enclosures in the RSS feed using the configured base URL
- **FR-006**: System MUST generate absolute URLs for podcast artwork (channel-level image) in the RSS feed using the configured base URL
- **FR-007**: System MUST generate absolute URLs for episode images in the RSS feed using the configured base URL
- **FR-008**: System MUST display the absolute RSS feed URL on the web dashboard using the configured base URL
- **FR-009**: System MUST store only relative paths internally (in XML or database) and convert to absolute URLs during RSS feed generation
- **FR-010**: System MUST apply base URL to default podcast artwork path when no custom artwork is uploaded
- **FR-011**: System MUST generate absolute URLs regardless of how the web interface or RSS feed is accessed (direct IP, localhost, or domain name)
- **FR-012**: System MUST support base URLs with path components (e.g., `http://example.com/podcasts`) and correctly append resource paths
- **FR-013**: System MUST fail to start and log critical error if base URL is not configured
- **FR-014**: Generated RSS feed MUST pass validation in Apple Podcasts Connect validator with absolute URLs
- **FR-015**: Generated RSS feed MUST pass validation in W3C Feed Validator with absolute URLs
- **FR-016**: System MUST update absolute URLs immediately when base URL configuration is changed (requires server restart)
- **FR-017**: System MUST skip episodes with malformed paths that cannot be converted to valid absolute URLs and continue generating RSS feed with remaining valid episodes
- **FR-018**: System MUST log critical errors only (startup validation failures, missing or invalid base URL configuration, malformed episode paths during feed generation)

### Key Entities

- **Base URL Configuration**: Represents the configured base URL for the application deployment. Attributes include protocol (http/https), hostname, optional port number, and optional path prefix. Used by all URL generation logic to convert relative paths to absolute URLs.

- **URL Path**: Represents the relative path component of a resource URL. Attributes include path string (e.g., `/audio/filename.mp3`), resource type (audio, artwork, static asset). Stored internally and converted to absolute URLs using base URL configuration during output generation.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: RSS feeds generated by the system pass Apple Podcasts Connect validator with zero errors related to URL accessibility after fix
- **SC-002**: RSS feeds generated by the system pass W3C Feed Validator with zero warnings about relative URLs
- **SC-003**: Podcast clients (Apple Podcasts, Spotify, Overcast) successfully fetch and play audio files from RSS feed URLs within 10 seconds of feed refresh
- **SC-004**: Podcast artwork displays correctly in at least 3 major podcast client applications when subscribed to the feed
- **SC-005**: 100% of audio enclosure URLs in generated RSS feeds are absolute URLs (contain protocol and hostname)
- **SC-006**: 100% of image URLs (podcast and episode artwork) in generated RSS feeds are absolute URLs
- **SC-007**: System startup validation catches 100% of invalid base URL configurations and provides actionable error messages
- **SC-008**: RSS feed URL displayed on web dashboard matches the actual accessible feed URL in 100% of cases
- **SC-009**: Changing base URL configuration and restarting server results in all URLs updating to new base within 5 seconds
- **SC-010**: System supports deployment in subdirectory paths (e.g., `http://example.com/podcasts/`) without broken URL generation

## Assumptions

- Base URL will be configured once during initial deployment and rarely changed (not a runtime configuration)
- System administrators have access to edit the configuration file and restart the server
- The base URL configured represents the publicly accessible URL for the podcast server
- Internal storage will continue using relative paths for portability (only RSS feed output uses absolute URLs)
- All podcast clients and directories expect absolute URLs in RSS feeds per RSS 2.0 specification best practices
- The server will validate base URL format on startup rather than runtime to avoid performance overhead
- Base URL changes require server restart to take effect (hot-reload is not required)
- The default podcast artwork path will be resolved to absolute URL using the same base URL logic
- RSS feed generation will convert URLs on-the-fly rather than storing absolute URLs in the data store
- The application will use the configured base URL regardless of request headers or proxy configurations

## Out of Scope

The following items are explicitly excluded from this bug fix:

- Automatic detection of base URL from HTTP request headers (e.g., X-Forwarded-Host)
- Hot-reloading of base URL configuration without server restart
- Per-episode custom base URLs or multiple base URL configurations
- URL shortening or custom domain mapping for individual episodes
- Migration of existing relative URLs stored in data files (on-the-fly conversion handles this)
- CDN or proxy URL rewriting support
- Automatic SSL/TLS certificate provisioning for HTTPS base URLs
- Load balancer or reverse proxy configuration guidance
- URL validation beyond basic format checking (e.g., DNS resolution, accessibility checks)
- Resource reachability validation for generated absolute URLs before including in RSS feed
- Backward compatibility with podcast clients that might expect relative URLs (all standard clients require absolute URLs)
- Web dashboard URL configuration separate from RSS feed URL configuration
- Dynamic base URL selection based on request origin or user session
