# Implementation Plan: Fix Relative URLs to Absolute URLs

**Branch**: `002-fix-relative-urls` | **Date**: 2025-12-31 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-fix-relative-urls/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

This bug fix converts all relative URLs (e.g., `/audio/file.mp3`) to absolute URLs (e.g., `http://example.com/audio/file.mp3`) in the RSS feed. The fix requires adding a mandatory `baseURL` configuration field, validating it at startup, and applying URL encoding per RFC 3986 during RSS feed generation. All URLs are converted on-the-fly from stored relative paths to absolute URLs using the configured base URL, ensuring podcast clients can properly access audio files and artwork.

## Technical Context

**Language/Version**: Go 1.21  
**Primary Dependencies**: `github.com/eduncan911/podcast` (RSS generation), `net/http` (standard library), `net/url` (URL parsing and validation)  
**Storage**: Filesystem-based (XML file at `./data/podcast.xml` stores podcast metadata)  
**Testing**: Go standard testing package + integration tests with RSS validators  
**Target Platform**: Linux/macOS server (HTTP server application)  
**Project Type**: Single project (backend HTTP server with embedded HTML templates)  
**Performance Goals**: RSS feed generation < 100ms for feeds with 1000 episodes, startup validation < 50ms  
**Constraints**: Must not break existing podcast feeds during deployment, zero downtime not required (restart acceptable)  
**Scale/Scope**: Single podcast per instance, 1000+ episodes supported, 3-5 files require modification

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. RSS Standard Compliance ✅ **PASS**

**Check**: Generated RSS feeds must validate against RSS 2.0 spec and podcast validators.

**Alignment**: This fix directly improves RSS compliance by converting relative URLs to absolute URLs, which is required per RSS 2.0 best practices. Relative URLs in enclosure tags cause validation warnings and prevent proper audio file access in podcast clients.

**Action**: Add RSS validator integration tests to verify absolute URLs after fix.

---

### II. File-Centric Architecture ✅ **PASS**

**Check**: Audio files are source of truth, no database drift, file operations drive RSS state.

**Alignment**: This fix maintains file-centric architecture. Relative paths continue to be stored in `podcast.xml` (the file-based metadata store). Conversion to absolute URLs happens on-the-fly during RSS feed generation, not during storage. No new database or state tracking is introduced.

**Action**: Ensure URL conversion logic is in RSS feed generation layer only, not storage layer.

---

### III. HTTP-First Design ✅ **PASS**

**Check**: HTTP endpoints for upload, download, and RSS serving remain primary interface.

**Alignment**: This fix only affects URL format in RSS feed responses and configuration. No changes to HTTP endpoints, request/response patterns, or client-facing API. Base URL configuration is via file (existing pattern), not requiring CLI tools.

**Action**: None required - HTTP interface unchanged.

---

### IV. Testing Discipline ✅ **PASS**

**Check**: RSS validation through automated tests, file upload scenarios covered.

**Alignment**: This fix requires new tests for:
- Base URL configuration validation at startup
- URL conversion with special characters (RFC 3986 encoding)
- Absolute URL generation in RSS feed (validator integration)
- Edge cases (missing base URL, trailing slashes, subdirectory paths)

**Action**: Add integration tests with RSS validators and unit tests for URL conversion logic.

---

### V. Simplicity & Maintainability ✅ **PASS**

**Check**: Simplest solution, avoid premature abstraction, justify added complexity.

**Alignment**: This fix uses the simplest approach:
- Single configuration field (`baseURL`)
- Standard library `net/url` for parsing and validation
- On-the-fly conversion (no caching or background processing)
- Fail-fast at startup (no complex fallback logic)

No new abstractions, frameworks, or background workers introduced. URL encoding uses standard library functions.

**Action**: Keep URL conversion as pure functions, avoid introducing URL builder abstraction.

---

### Constitution Compliance Summary

**Status**: ✅ **ALL GATES PASS**

This bug fix aligns with all constitutional principles. It improves RSS compliance, maintains file-centric architecture, preserves HTTP-first design, requires proper test coverage, and uses the simplest possible solution (configuration + standard library URL functions).

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
# Existing structure (no changes to layout)
cmd/
└── server/
    └── main.go              # [MODIFY] Add base URL config loading and validation

internal/
├── handlers/
│   ├── episodes.go          # [MODIFY] Pass base URL to RSS generation
│   ├── feed.go              # [MODIFY] Pass base URL to RSS generation
│   ├── web.go               # [MODIFY] Use base URL for dashboard feed URL display
│   └── static.go            # [NO CHANGE] Static file serving unchanged
├── models/
│   ├── episode.go           # [NO CHANGE] Continue storing relative paths
│   └── podcast.go           # [NO CHANGE] Continue storing relative paths
├── rss/
│   ├── feed.go              # [MODIFY] Convert relative URLs to absolute with encoding
│   └── parser.go            # [NO CHANGE] XML parsing unchanged
├── storage/
│   ├── filesystem.go        # [NO CHANGE] File operations unchanged
│   └── xml.go               # [NO CHANGE] XML persistence unchanged
└── config/                  # [NEW PACKAGE]
    ├── config.go            # [NEW] Configuration loading and validation
    └── config_test.go       # [NEW] Configuration tests

web/
├── templates/
│   ├── index.html           # [MODIFY] Display absolute feed URL from base URL
│   └── components/          # [NO CHANGE] Component templates unchanged
└── static/                  # [NO CHANGE] Static assets unchanged

tests/
├── integration/             # [NEW DIRECTORY]
│   ├── rss_validator_test.go    # [NEW] RSS feed validation tests
│   └── url_conversion_test.go   # [NEW] End-to-end URL conversion tests
└── unit/                    # [NEW DIRECTORY]
    └── config_test.go       # [NEW] Configuration validation unit tests

config.yaml                  # [MODIFY] Add baseURL field (required)
```

**Structure Decision**: Single project structure maintained. This is a bug fix that adds URL conversion logic to existing RSS generation layer (`internal/rss/`) and introduces configuration validation in a new `internal/config/` package. No architectural changes - simply adding URL transformation at RSS feed generation time.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

**No violations identified** - Constitution Check passes all gates without requiring complexity justification.

---

## Phase 0: Research ✅ COMPLETE

**Status**: All technical unknowns resolved

**Artifacts Generated**:
- [research.md](./research.md) - Complete technical research covering:
  - URL parsing and validation strategy (`net/url` package)
  - RFC 3986 encoding requirements and implementation
  - Configuration validation approach (fail-fast at startup)
  - URL conversion timing (on-the-fly during RSS generation)
  - Error handling for malformed paths (skip and log)
  - Testing strategy with RSS validators

**Key Decisions**:
- Use Go standard library `net/url` for all URL operations
- Apply RFC 3986 encoding via `url.PathEscape()` for path components
- Fail-fast validation at startup (no default base URL)
- On-the-fly URL conversion during RSS feed generation
- Skip episodes with malformed paths, log errors, continue with valid episodes

---

## Phase 1: Design & Contracts ✅ COMPLETE

**Status**: Data model, API contracts, and quickstart guide generated

**Artifacts Generated**:
- [data-model.md](./data-model.md) - Entity definitions and relationships:
  - Configuration entity (new `baseURL` field)
  - Podcast entity (unchanged, stores relative paths)
  - Episode entity (unchanged, stores relative paths)
  - URL Conversion logical entity (conversion rules and algorithm)
  - State machine for configuration validation
  - No data migration required

- [contracts/openapi.yaml](./contracts/openapi.yaml) - API contract changes:
  - `/feed.xml` endpoint now returns absolute URLs
  - `GET /` dashboard displays absolute feed URL
  - Audio and artwork endpoints unchanged (referenced by absolute URLs)
  - Configuration schema documented

- [quickstart.md](./quickstart.md) - Implementation guide:
  - Step-by-step implementation (15-20 minutes)
  - Configuration setup and validation
  - URL conversion logic implementation
  - Web dashboard updates
  - Test creation and validation
  - Deployment checklist and troubleshooting

**Agent Context Updated**:
- ✅ AGENTS.md updated with Go 1.21 + `net/url` + RSS generation libraries
- Recent changes documented for this feature

**Constitution Re-Check After Phase 1**:

### I. RSS Standard Compliance ✅ **PASS**
- Design maintains RSS 2.0 compliance
- Absolute URLs align with RSS best practices
- Integration tests include RSS validator checks

### II. File-Centric Architecture ✅ **PASS**
- Design preserves relative path storage in `podcast.xml`
- No new database or persistent state introduced
- URL conversion is stateless, on-the-fly transformation

### III. HTTP-First Design ✅ **PASS**
- HTTP endpoints remain primary interface
- No CLI tools required for configuration (file-based)
- Configuration via `config.yaml` follows existing pattern

### IV. Testing Discipline ✅ **PASS**
- Comprehensive test plan in quickstart
- Unit tests for configuration validation
- Integration tests with RSS validators
- URL encoding and conversion test coverage

### V. Simplicity & Maintainability ✅ **PASS**
- Uses only standard library (no new external dependencies)
- Single configuration field added
- Straightforward URL conversion algorithm
- No caching, queues, or background processing

**Design Validation**: All constitutional principles maintained after Phase 1 design.

---

## Phase 2: Task Breakdown

**Note**: Task breakdown is generated by the `/speckit.tasks` command (separate from `/speckit.plan`).

Tasks will be created based on the following work streams:

1. **Configuration** (2-3 tasks):
   - Add `baseURL` field to config structure
   - Implement configuration validation at startup
   - Add unit tests for configuration

2. **URL Conversion** (3-4 tasks):
   - Implement URL conversion helper functions
   - Update RSS feed generation to apply absolute URLs
   - Update web dashboard to display absolute feed URL
   - Add unit tests for URL conversion logic

3. **Integration & Testing** (2-3 tasks):
   - Create integration tests with RSS validators
   - Add end-to-end tests for URL encoding
   - Validate with Apple Podcasts Connect and W3C validators

4. **Documentation** (1-2 tasks):
   - Update README with base URL configuration instructions
   - Document deployment checklist

**Estimated Total**: 8-12 tasks, 4-6 hours of implementation time

Run `/speckit.tasks` to generate detailed task breakdown with acceptance criteria.

---

## Implementation Readiness

**Status**: ✅ **READY FOR IMPLEMENTATION**

**Checklist**:
- [x] Constitution Check passed (all gates)
- [x] Research completed (all unknowns resolved)
- [x] Data model defined (entities and relationships)
- [x] API contracts documented (OpenAPI spec)
- [x] Quickstart guide created (step-by-step implementation)
- [x] Agent context updated (AGENTS.md with new technologies)
- [x] Constitution re-checked after design (all principles maintained)

**Next Command**: Run `/speckit.tasks` to generate detailed task breakdown for implementation.
