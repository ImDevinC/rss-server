<!--
Sync Impact Report:
Version Change: Initial version → 1.0.0
Principles Added:
  1. RSS Standard Compliance
  2. File-Centric Architecture
  3. HTTP-First Design
  4. Testing Discipline
  5. Simplicity & Maintainability
Sections Added:
  - Core Principles
  - Technical Standards
  - Development Workflow
  - Governance
Templates Requiring Updates:
  ✅ plan-template.md - Constitution Check section will reference these principles
  ✅ spec-template.md - Requirements will align with RSS compliance and file handling
  ✅ tasks-template.md - Tasks will include RSS validation and file upload testing
Follow-up TODOs:
  - RATIFICATION_DATE needs to be set when Devin formally adopts this constitution
-->

# RSS Server Constitution

## Core Principles

### I. RSS Standard Compliance

The system MUST generate podcast RSS 2.0 feeds that conform to the official RSS specification and Apple Podcasts requirements. All generated feeds MUST validate against standard RSS validators (W3C Feed Validator, Cast Feed Validator). This is NON-NEGOTIABLE.

**Rationale**: Invalid RSS feeds break podcast aggregators and degrade user experience. Standards compliance ensures broad compatibility.

### II. File-Centric Architecture

Audio files are the single source of truth. The system MUST NOT maintain a separate database for file metadata that could drift from actual file state. File operations (add, delete, list) drive the RSS feed state directly.

**Rationale**: File-centric design eliminates synchronization issues between database and filesystem, reducing complexity and failure modes.

### III. HTTP-First Design

The primary interface is HTTP endpoints for file upload and RSS feed serving. The server MUST expose:
- RSS feed endpoint (GET)
- Audio file upload endpoint (POST)
- Audio file download endpoints (GET)

CLI tools, admin interfaces, or APIs are secondary and MUST NOT be required for core functionality.

**Rationale**: HTTP-first ensures the server is immediately usable by podcast clients and standard web tools without custom clients.

### IV. Testing Discipline

All RSS feed generation MUST be validated through automated tests using RSS validators. All file upload scenarios (success, duplicate, invalid format, size limits) MUST have test coverage. Tests MUST fail before implementation (TDD encouraged but not strictly enforced for this project type).

**Rationale**: RSS feed correctness is critical for podcast distribution. Automated validation prevents regressions.

### V. Simplicity & Maintainability

Start with the simplest solution that satisfies requirements. Avoid abstractions, frameworks, or features not immediately needed. Database usage MUST be justified (file metadata can be computed on-demand initially). Background processing, queues, caching layers require explicit justification.

**Rationale**: Premature optimization and over-engineering increase maintenance burden without proven value. Build complexity when needed, not preemptively.

## Technical Standards

### Technology Stack

- **Language**: Go (version 1.21+)
- **HTTP Server**: Standard library `net/http` (third-party frameworks require justification)
- **File Storage**: Local filesystem (cloud storage is future extension)
- **RSS Generation**: XML encoding via standard library or well-maintained RSS library
- **Audio Format Support**: Initially MP3, extensible to M4A/AAC
- **Testing**: Go standard testing package + RSS validator integration

### RSS Feed Requirements

- MUST include all required RSS 2.0 elements (title, link, description, language)
- MUST include podcast-specific iTunes tags (itunes:author, itunes:category, itunes:image)
- MUST generate valid enclosure tags for audio files (url, length, type)
- MUST serve feeds with `Content-Type: application/rss+xml`
- MUST include pubDate in RFC-822 format

### File Handling Requirements

- MUST validate audio file formats before accepting uploads
- MUST enforce reasonable file size limits (e.g., 500MB default, configurable)
- MUST generate unique, URL-safe filenames if original names conflict
- MUST serve audio files with correct Content-Type headers
- MUST handle concurrent uploads safely

### Security Requirements

- MUST validate and sanitize all file uploads (magic number verification, not just extension)
- MUST prevent path traversal attacks in file operations
- MUST enforce upload size limits to prevent DoS
- Authentication/authorization is NOT required for v1.0 (future addition)

## Development Workflow

### Feature Development Process

1. **Specification First**: Document user scenarios in spec.md before implementation
2. **Plan Before Code**: Create implementation plan including file structure and endpoints
3. **Test Critical Paths**: RSS validation and file upload scenarios MUST have tests
4. **Incremental Delivery**: Each user story should be independently deployable

### Code Quality Gates

- All code MUST pass `go vet` and `golint` (or equivalent linter)
- All RSS feeds generated MUST validate against RSS validator in tests
- All endpoints MUST have at least one integration test
- Breaking changes to RSS feed format require MAJOR version bump

### Documentation Requirements

- README MUST include quickstart for running server and adding first audio file
- API endpoints MUST be documented (path, method, parameters, response format)
- Configuration options MUST be documented (port, storage directory, size limits)

## Governance

This constitution defines the non-negotiable principles for RSS Server development. All feature specifications, implementation plans, and pull requests MUST align with these principles.

### Amendment Process

1. Proposed changes MUST be documented with rationale
2. Impact on existing code/design MUST be assessed
3. Version MUST be incremented per semantic versioning rules
4. Templates (plan, spec, tasks) MUST be updated to reflect changes
5. Migration plan MUST be provided for breaking changes

### Compliance Verification

- Specification reviews MUST verify alignment with RSS Standard Compliance and File-Centric Architecture
- Implementation plans MUST justify any complexity additions per Simplicity principle
- Code reviews MUST verify HTTP-First Design (no CLI-required functionality)
- All PRs MUST include RSS validation test results for feed-affecting changes

### Versioning Policy

- **MAJOR**: Remove/change RSS feed structure, change core API contracts, remove file-centric approach
- **MINOR**: Add new endpoints, add audio format support, add configuration options
- **PATCH**: Bug fixes, documentation updates, test additions, performance improvements

**Version**: 1.0.0 | **Ratified**: TODO(RATIFICATION_DATE): Set when formally adopted | **Last Amended**: 2025-12-31
