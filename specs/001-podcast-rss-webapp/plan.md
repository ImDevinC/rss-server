# Implementation Plan: Podcast RSS Webapp

**Branch**: `001-podcast-rss-webapp` | **Date**: 2025-12-31 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-podcast-rss-webapp/spec.md`

## Summary

Build a web application that allows podcast creators to upload audio files and generate a podcast-compatible RSS feed. The system uses Go for the backend with HTMX for dynamic frontend interactions. Audio files and metadata are stored in the filesystem, with the RSS XML file serving as the single source of truth. Users can add and delete episodes via both web UI and REST API. This is a single-podcast system with no authentication (private RSS URL only).

**Scope Adjustments from Spec**:
- **Simplified to single podcast**: No multi-podcast support (FR-014 removed)
- **No authentication**: FR-020 authentication removed per user requirements
- **No episode editing**: FR-012 editing removed, add/delete only per user requirements
- **XML as source of truth**: Aligns with constitution's File-Centric Architecture principle
- **Focus on P1**: User Story 1 (upload/delete) + User Story 2 (customize metadata) for MVP

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: 
- Standard library net/http (HTTP server)
- Standard library encoding/xml (RSS generation)
- HTMX 1.9+ (frontend interactivity via CDN)
- Go template library (HTML rendering)

**Storage**: 
- RSS XML file (podcast.xml) - source of truth for metadata
- Local filesystem - audio files and artwork storage
- No database per user requirements

**Testing**: 
- Go standard testing package
- RSS validator integration (W3C Feed Validator API or Cast Feed Validator)
- HTTP integration tests

**Target Platform**: Linux/macOS server (any Go-compatible platform)

**Project Type**: Web application (Go backend + HTMX frontend)

**Performance Goals**: 
- Handle 100 concurrent RSS feed requests
- Support up to 1000 episodes without degradation
- Upload processing: 30 seconds per 100MB

**Constraints**: 
- File size limit: 500MB per upload (configurable)
- Single RSS feed (no multi-tenancy)
- No authentication (private URL security model)

**Scale/Scope**: 
- Single podcast creator
- Up to 1000 episodes
- Designed for personal/small team use

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. RSS Standard Compliance ✅ PASS

- **Requirement**: Generate RSS 2.0 compliant feeds with iTunes tags
- **Plan**: Use Go encoding/xml with custom structs for RSS 2.0 + iTunes namespace
- **Validation**: Integration tests with W3C Feed Validator and Apple Podcasts validator
- **Status**: Compliant - RSS generation is core feature

### II. File-Centric Architecture ✅ PASS

- **Requirement**: Files are source of truth, no separate database
- **Plan**: RSS XML file (podcast.xml) stores all metadata, audio files stored in filesystem
- **Design Decision**: Read/write RSS XML directly for all operations (add episode, delete episode, update podcast metadata)
- **Status**: Fully compliant - XML file IS the database per user requirements

### III. HTTP-First Design ✅ PASS

- **Requirement**: HTTP endpoints for upload, feed serving, file downloads
- **Plan**: 
  - `GET /feed.xml` - RSS feed endpoint
  - `POST /api/episodes` - Upload episode
  - `DELETE /api/episodes/{id}` - Delete episode
  - `GET /audio/{filename}` - Audio file downloads
  - `GET /` - Web UI (HTMX interface)
- **Status**: Compliant - All core functionality via HTTP

### IV. Testing Discipline ✅ PASS

- **Requirement**: RSS validation tests, file upload scenario coverage
- **Plan**: 
  - Unit tests for RSS XML generation
  - Integration tests for upload/delete flows
  - RSS validator integration in CI
- **Status**: Compliant - Test plan covers critical paths

### V. Simplicity & Maintainability ✅ PASS

- **Requirement**: Start simple, avoid unnecessary abstractions
- **Plan**: 
  - No database (XML file storage)
  - No frameworks (standard library HTTP)
  - No background processing (synchronous operations)
  - No caching layer initially
- **Justification**: Direct XML file read/write is simpler than database + sync logic
- **Status**: Compliant - Minimal dependencies, straightforward architecture

### Constitution Compliance Summary

**Overall Status**: ✅ ALL GATES PASS

No complexity violations detected. The design aligns perfectly with all five constitutional principles:
1. RSS validation is mandatory in tests
2. XML file is the single source of truth (file-centric)
3. HTTP-first interface for all operations
4. Testing strategy covers critical paths
5. Simplest possible architecture (no DB, no frameworks, no abstractions)

## Project Structure

### Documentation (this feature)

```text
specs/001-podcast-rss-webapp/
├── plan.md              # This file
├── research.md          # Phase 0: Go RSS libraries, HTMX patterns, XML handling
├── data-model.md        # Phase 1: RSS XML schema, Episode structure, Podcast metadata
├── quickstart.md        # Phase 1: How to run server, upload first episode
├── contracts/           # Phase 1: API endpoint specifications
│   └── openapi.yaml     # REST API contract
└── checklists/
    └── requirements.md  # Spec quality validation
```

### Source Code (repository root)

```text
cmd/
└── server/
    └── main.go          # Server entry point

internal/
├── handlers/
│   ├── feed.go          # RSS feed handler (GET /feed.xml)
│   ├── episodes.go      # Episode API handlers (POST/DELETE)
│   ├── web.go           # Web UI handlers (GET /, dashboard)
│   └── static.go        # Audio/artwork file serving
├── rss/
│   ├── feed.go          # RSS XML generation (structs + marshaling)
│   ├── parser.go        # RSS XML parsing/reading
│   └── validator.go     # RSS validation utilities
├── storage/
│   ├── filesystem.go    # File operations (save/delete audio)
│   └── xml.go           # XML file read/write with locking
└── models/
    ├── episode.go       # Episode data structure
    └── podcast.go       # Podcast metadata structure

web/
├── templates/
│   ├── index.html       # Dashboard template
│   └── components/      # HTMX partial templates
│       ├── episode_list.html
│       ├── upload_form.html
│       └── settings_form.html
└── static/
    └── styles.css       # Minimal CSS

data/
├── audio/               # Audio file storage
├── artwork/             # Podcast artwork storage
└── podcast.xml          # RSS feed (source of truth)

tests/
├── integration/
│   ├── feed_test.go     # RSS feed generation tests
│   ├── upload_test.go   # Upload flow integration tests
│   └── api_test.go      # API endpoint tests
└── unit/
    ├── rss_test.go      # RSS XML generation unit tests
    └── storage_test.go  # File operations unit tests

go.mod                   # Go module definition
go.sum                   # Dependency checksums
README.md                # Project overview + quickstart
config.yaml              # Server configuration (port, limits, paths)
```

**Structure Decision**: 

This is a web application (Option 2 variant) but simplified to a single Go project structure since both frontend (HTMX templates) and backend (Go handlers) are served from the same binary. The structure follows Go best practices:

- `cmd/server/` - Application entry point
- `internal/` - Private application code (handlers, RSS generation, storage)
- `web/` - Frontend assets (HTMX templates, CSS)
- `data/` - Runtime data (audio files, RSS XML)
- `tests/` - Test organization by type

No separate `frontend/` and `backend/` directories needed because HTMX is served as Go templates, not a separate SPA build.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations detected. All constitutional principles are satisfied:

- ✅ No database complexity (XML file storage)
- ✅ No framework dependencies (standard library)
- ✅ No authentication complexity (deferred per requirements)
- ✅ No multi-tenancy complexity (single podcast)
- ✅ No background processing (synchronous operations)

This table remains empty because the design adheres to constitutional simplicity requirements.

## Phase 0: Research (COMPLETED)

**Status**: ✅ Complete  
**Output**: [research.md](./research.md)

### Key Decisions Made

1. **RSS Generation**: `github.com/eduncan911/podcast` library
   - Production-proven with comprehensive iTunes support
   - Clean API with built-in validation

2. **HTMX Integration**: Standard `html/template` with partial rendering
   - Zero external dependencies
   - Natural fit with Go server-side rendering

3. **XML Concurrency**: Atomic writes + `sync.RWMutex`
   - Read-heavy optimization
   - Safe concurrent access without database

4. **RSS Validation**: W3C Feed Validator API
   - Industry standard validation
   - Programmatic testing integration

### Technology Stack Confirmed

| Component | Technology | Justification |
|-----------|-----------|---------------|
| Language | Go 1.21+ | Constitution requirement |
| HTTP Server | net/http stdlib | Avoid frameworks per constitution |
| RSS Library | github.com/eduncan911/podcast | Production-ready, iTunes compliant |
| Frontend | HTMX 1.9+ | Dynamic UX without SPA complexity |
| Templates | html/template stdlib | Native Go, zero dependencies |
| Storage | Filesystem + XML | File-centric architecture |
| Validation | W3C Validator API | Industry standard |

## Phase 1: Design (COMPLETED)

**Status**: ✅ Complete  
**Outputs**: 
- [data-model.md](./data-model.md)
- [contracts/openapi.yaml](./contracts/openapi.yaml)
- [quickstart.md](./quickstart.md)
- AGENTS.md updated with Go 1.21+

### Data Model Summary

**Source of Truth**: `data/podcast.xml` (RSS XML file)

**Core Entities**:
1. **Podcast**: Channel-level metadata (title, author, artwork, category)
2. **Episode**: Item-level metadata (title, description, audio file, duration)
3. **AudioFile**: Physical file storage metadata

**Concurrency Model**: RWMutex for read-heavy workload, atomic file writes

### API Contract Summary

**Endpoints Defined**:
- `POST /api/episodes` - Upload episode (multipart/form-data)
- `DELETE /api/episodes/{id}` - Delete episode
- `GET /api/episodes` - List all episodes
- `POST /api/podcast/settings` - Update podcast metadata
- `GET /api/podcast/settings` - Get current settings
- `GET /feed.xml` - Serve RSS feed (application/rss+xml)
- `GET /audio/{filename}` - Stream audio file (audio/mpeg)
- `GET /` - Web dashboard (HTMX interface)

**Request/Response Formats**: JSON for API, HTML for HTMX partials, XML for RSS

### Quickstart Guide

Documented complete workflow:
1. Clone and setup (< 1 minute)
2. Start server (`go run cmd/server/main.go`)
3. Access dashboard (http://localhost:8080)
4. Upload first episode (via UI or API)
5. View RSS feed (http://localhost:8080/feed.xml)
6. Validate feed (W3C or Cast Feed Validator)
7. Customize podcast metadata
8. Production deployment steps

## Phase 2: Post-Design Constitution Check

**Status**: ✅ PASS (Re-evaluated after Phase 1 design)

### I. RSS Standard Compliance ✅ PASS

- **Design**: Using `github.com/eduncan911/podcast` library with full iTunes support
- **Validation**: W3C Feed Validator integration in tests
- **OpenAPI**: `/feed.xml` endpoint documented with example RSS XML
- **Status**: Compliant - library handles RSS 2.0 + iTunes spec automatically

### II. File-Centric Architecture ✅ PASS

- **Design**: `data/podcast.xml` is single source of truth
- **Data Model**: No database, all state in XML file + filesystem
- **Concurrency**: RWMutex + atomic writes ensure consistency
- **Status**: Fully compliant - XML file IS the database

### III. HTTP-First Design ✅ PASS

- **Design**: All operations via HTTP endpoints
- **OpenAPI**: 8 endpoints documented (upload, delete, feed, audio, dashboard)
- **HTMX**: Web UI uses standard HTTP semantics (POST, DELETE, GET)
- **Status**: Compliant - no CLI required for any operation

### IV. Testing Discipline ✅ PASS

- **Design**: W3C Feed Validator integration planned
- **Research**: Testing strategy documented with local + external validation
- **Coverage**: RSS generation, file upload, concurrency scenarios
- **Status**: Test plan covers critical paths per constitution

### V. Simplicity & Maintainability ✅ PASS

- **Design**: Minimal dependencies (1 RSS library + HTMX CDN)
- **No Database**: Constitutional requirement satisfied
- **No Frameworks**: Using stdlib net/http per constitution
- **No Background Jobs**: Synchronous operations only
- **Status**: Simplest possible architecture - no violations

### Final Constitution Compliance: ✅ ALL GATES PASS

Design maintains full compliance with all constitutional principles. No complexity violations introduced during Phase 1 design.

## Next Steps

**Phase 2 Complete**: Implementation plan ready for task generation.

**Ready for**:
- `/speckit.tasks` - Generate implementation tasks organized by user story
- Manual implementation following plan, data model, and API contracts

**Deliverables Completed**:
1. ✅ Technical research with decisions (research.md)
2. ✅ Data model with RSS XML schema (data-model.md)
3. ✅ API contracts in OpenAPI 3.0 (contracts/openapi.yaml)
4. ✅ Quickstart guide for users (quickstart.md)
5. ✅ Agent context updated (AGENTS.md)
6. ✅ Constitution compliance verified (post-design check)

**Implementation Complexity**: LOW
- Clear architecture with proven libraries
- Well-defined data model (RSS 2.0 standard)
- Complete API contracts (OpenAPI 3.0)
- No constitutional violations
- Straightforward Go + HTMX implementation
