# Tasks: Fix Relative URLs to Absolute URLs

**Branch**: `002-fix-relative-urls`  
**Input**: Design documents from `/home/devin/Projects/rss-server/specs/002-fix-relative-urls/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/openapi.yaml

**Tests**: Tests are included per TDD approach (tests written first, must fail before implementation)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Single Go project structure at repository root:
- `cmd/server/` - Application entry point
- `internal/` - Internal packages (config, handlers, models, rss, storage)
- `tests/` - Test files (unit/ and integration/)
- `web/` - Web templates and static assets

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure for URL fix implementation

- [X] T001 Create internal/config package directory
- [X] T002 [P] Create tests/integration directory for RSS validator tests
- [X] T003 [P] Create tests/unit directory for configuration unit tests
- [X] T004 Verify Go version 1.21+ installed (`go version`)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Configuration infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete. All user stories depend on base URL configuration.

- [X] T005 [P] Create Config struct with BaseURL field in internal/config/config.go
- [X] T006 [P] Implement Load() function with YAML parsing in internal/config/config.go
- [X] T007 Implement Validate() function with base URL checks in internal/config/config.go
- [X] T008 Add baseURL field to config.yaml at repository root
- [X] T009 Update cmd/server/main.go to load and validate configuration at startup
- [X] T010 Add fail-fast error handling in cmd/server/main.go for missing/invalid base URL
- [X] T011 Add startup logging for successful base URL configuration in cmd/server/main.go

**Checkpoint**: Configuration framework ready - base URL can be validated at startup. All user stories can now begin.

---

## Phase 3: User Story 2 - Configure Base URL for Deployment (Priority: P1) 🎯 MVP INFRASTRUCTURE

**Goal**: Enable system administrators to configure the base URL (hostname and protocol) so the system generates correct absolute URLs for the deployment environment.

**Independent Test**: Edit config.yaml to specify a base URL, restart server, verify server starts successfully and logs the base URL.

### Tests for User Story 2 (TDD - Write FIRST, ensure they FAIL)

- [X] T012 [P] [US2] Unit test: Valid HTTP base URL passes validation in tests/unit/config_test.go
- [X] T013 [P] [US2] Unit test: Valid HTTPS base URL passes validation in tests/unit/config_test.go
- [X] T014 [P] [US2] Unit test: Missing base URL returns error in tests/unit/config_test.go
- [X] T015 [P] [US2] Unit test: Invalid scheme (ftp://) returns error in tests/unit/config_test.go
- [X] T016 [P] [US2] Unit test: Missing hostname returns error in tests/unit/config_test.go
- [X] T017 [P] [US2] Unit test: Trailing slash normalization works in tests/unit/config_test.go
- [X] T018 [US2] Run unit tests - verify all FAIL before implementation (`go test ./tests/unit/...`)

### Implementation for User Story 2

Implementation already completed in Phase 2 (Foundational). This phase validates the implementation.

- [X] T019 [US2] Run unit tests - verify all PASS after foundational implementation (`go test ./tests/unit/... -v`)
- [X] T020 [US2] Manual test: Start server without base URL config, verify fatal error
- [X] T021 [US2] Manual test: Start server with invalid base URL, verify validation error
- [X] T022 [US2] Manual test: Start server with valid base URL, verify success log

**Checkpoint**: Base URL configuration is fully functional, validated, and tested independently.

---

## Phase 4: User Story 1 - RSS Feed Validates with Absolute URLs (Priority: P1) 🎯 MVP CORE

**Goal**: Convert all relative URLs in the RSS feed to absolute URLs so podcast directories and clients can properly fetch audio files and artwork.

**Independent Test**: Upload an episode with artwork, access /feed.xml, verify all URLs contain absolute URLs starting with configured base URL (e.g., `http://example.com/audio/file.mp3` instead of `/audio/file.mp3`).

### Tests for User Story 1 (TDD - Write FIRST, ensure they FAIL)

- [X] T023 [P] [US1] Integration test: RSS feed contains absolute audio URLs in tests/integration/url_conversion_test.go
- [X] T024 [P] [US1] Integration test: RSS feed contains absolute image URLs in tests/integration/url_conversion_test.go
- [X] T025 [P] [US1] Integration test: Special characters URL-encoded (spaces → %20) in tests/integration/url_conversion_test.go
- [X] T026 [P] [US1] Integration test: Malformed paths skipped, valid episodes included in tests/integration/url_conversion_test.go
- [X] T027 [P] [US1] Integration test: RSS XML structure valid in tests/integration/rss_validator_test.go
- [X] T028 [P] [US1] Integration test: Enclosure tags have absolute URLs in tests/integration/rss_validator_test.go
- [X] T029 [US1] Run integration tests - verify all FAIL before implementation (`go test ./tests/integration/...`)

### Implementation for User Story 1

- [X] T030 [US1] Add convertToAbsoluteURL() helper function in internal/rss/feed.go
- [X] T031 [US1] Update GenerateFeed() signature to accept baseURL parameter in internal/rss/feed.go
- [X] T032 [US1] Apply URL conversion to podcast ImageURL in internal/rss/feed.go
- [X] T033 [US1] Apply URL conversion to episode AudioURL with RFC 3986 encoding in internal/rss/feed.go
- [X] T034 [US1] Add error handling to skip malformed episodes in internal/rss/feed.go
- [X] T035 [US1] Add baseURL field to RSSStore struct in internal/storage/xml.go
- [X] T036 [US1] Update LoadRSSStore() to accept baseURL parameter in internal/storage/xml.go
- [X] T037 [US1] Update ServeXML() to pass baseURL to GenerateFeed() in internal/storage/xml.go
- [X] T038 [US1] Update saveToDisk() to pass baseURL to GenerateFeed() in internal/storage/xml.go
- [X] T039 [US1] Update LoadRSSStore() call in cmd/server/main.go to pass baseURL
- [X] T040 [US1] Run integration tests - verify all PASS after implementation (`go test ./tests/integration/... -v`)
- [X] T041 [US1] Manual test: Upload episode, fetch /feed.xml, verify absolute URLs
- [X] T042 [US1] Manual test: Check special characters are URL-encoded in RSS feed

**Checkpoint**: RSS feed generates absolute URLs, passes all validation tests independently.

---

## Phase 5: User Story 3 - Web Dashboard Shows Correct Feed URL (Priority: P2)

**Goal**: Display the RSS feed URL on the web dashboard using the configured base URL so users know the correct URL to submit to podcast directories.

**Independent Test**: Access the web dashboard and verify the displayed RSS feed URL matches the configured base URL (e.g., shows `http://podcast.example.com/feed.xml` when that's configured).

### Tests for User Story 3 (TDD - Write FIRST, ensure they FAIL)

- [X] T043 [P] [US3] Integration test: Dashboard displays absolute feed URL in tests/integration/dashboard_test.go
- [X] T044 [P] [US3] Integration test: Feed URL matches configured base URL in tests/integration/dashboard_test.go
- [X] T045 [US3] Run dashboard tests - verify all FAIL before implementation (`go test ./tests/integration/dashboard_test.go`)

### Implementation for User Story 3

- [X] T046 [US3] Add baseURL field to WebHandler struct in internal/handlers/web.go
- [X] T047 [US3] Update NewWebHandler() to accept baseURL parameter in internal/handlers/web.go
- [X] T048 [US3] Update HandleDashboard() to use baseURL for FeedURL in internal/handlers/web.go
- [X] T049 [US3] Update NewWebHandler() call in cmd/server/main.go to pass baseURL
- [X] T050 [US3] Run dashboard tests - verify all PASS after implementation (`go test ./tests/integration/dashboard_test.go -v`)
- [X] T051 [US3] Manual test: Access dashboard at http://localhost:8080, verify absolute feed URL displayed

**Checkpoint**: Dashboard displays absolute feed URL using configured base URL, independently tested.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, final validation, and deployment preparation

- [X] T052 [P] Add Configuration section to README.md documenting base_url field
- [X] T053 [P] Add troubleshooting guide for common errors to README.md
- [X] T054 [P] Add deployment checklist to README.md
- [X] T055 Run all unit tests (`go test ./tests/unit/... -v`)
- [X] T056 Run all integration tests (`go test ./tests/integration/... -v`)
- [X] T057 Run full test suite with coverage (`go test ./... -cover`)
- [X] T058 Manual validation: Server fails to start without base_url
- [X] T059 Manual validation: Server fails with invalid base_url format
- [X] T060 Manual validation: Server starts successfully with valid base_url
- [X] T061 Manual validation: RSS feed contains only absolute URLs
- [X] T062 Manual validation: Special characters URL-encoded in feed
- [X] T063 Manual validation: Dashboard shows absolute feed URL
- [X] T064 Verify all quickstart.md test scenarios pass
- [X] T065 Mark all tasks complete in this tasks.md file

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup (Phase 1) - BLOCKS all user stories
- **User Story 2 (Phase 3)**: Depends on Foundational (Phase 2) - Configuration validation
- **User Story 1 (Phase 4)**: Depends on Foundational (Phase 2) - RSS feed URL conversion
- **User Story 3 (Phase 5)**: Depends on Foundational (Phase 2) - Dashboard display
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Validates configuration infrastructure
- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - Core bug fix, no dependencies on other stories
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - Usability improvement, no dependencies on other stories

### Within Each User Story

- Tests MUST be written FIRST and FAIL before implementation (TDD)
- Configuration before usage
- Core logic before integration
- Unit tests before integration tests
- Implementation before validation

### Parallel Opportunities

**Phase 1 (Setup)**: All 4 tasks can run in parallel (T001-T004)

**Phase 2 (Foundational)**: Tasks T005-T007 can run in parallel (different concerns within config package)

**Phase 3 (US2 Tests)**: Tasks T012-T017 can run in parallel (independent unit tests)

**Phase 4 (US1 Tests)**: Tasks T023-T028 can run in parallel (independent integration tests)

**Phase 5 (US3 Tests)**: Tasks T043-T044 can run in parallel (independent tests)

**Phase 6 (Polish)**: Tasks T052-T054 can run in parallel (documentation in different sections)

**Cross-Story Parallelization**: After Phase 2 completes, User Stories 1, 2, and 3 can be worked on in parallel by different developers (though US2 validates infrastructure that US1 and US3 use).

---

## Parallel Example: User Story 1 Implementation

```bash
# After tests are written and failing, these can run in parallel:
# (These tasks all modify different functions/sections of the same file)

Task T032: "Apply URL conversion to podcast ImageURL in internal/rss/feed.go"
Task T033: "Apply URL conversion to episode AudioURL in internal/rss/feed.go"

# These tasks modify different files:
Task T035: "Add baseURL field to RSSStore struct in internal/storage/xml.go"
Task T030: "Add convertToAbsoluteURL() helper function in internal/rss/feed.go"
```

---

## Implementation Strategy

### MVP First (US2 + US1 Only)

1. Complete Phase 1: Setup (4 tasks)
2. Complete Phase 2: Foundational (7 tasks) - CRITICAL, blocks all stories
3. Complete Phase 3: User Story 2 (11 tasks) - Configuration validation
4. Complete Phase 4: User Story 1 (20 tasks) - Core bug fix
5. **STOP and VALIDATE**: Test US1 independently with RSS feed
6. Deploy/demo MVP (base URL configuration + absolute URLs in RSS)

**MVP Scope**: Phases 1-4 = 42 tasks total

### Incremental Delivery

1. Complete Setup + Foundational → Configuration framework ready
2. Add User Story 2 → Test config validation → Validate
3. Add User Story 1 → Test RSS feed with absolute URLs → Deploy/Demo (Core fix working!)
4. Add User Story 3 → Test dashboard display → Deploy/Demo (Usability improved)
5. Polish phase → Final documentation and validation

### Parallel Team Strategy

With 2-3 developers:

1. Team completes Setup (Phase 1) together (4 tasks, 5 minutes)
2. Team completes Foundational (Phase 2) together (7 tasks, 10 minutes)
3. Once Foundational is done:
   - Developer A: User Story 2 (Configuration validation) - 11 tasks
   - Developer B: User Story 1 (RSS URL conversion) - 20 tasks
   - Developer C: User Story 3 (Dashboard display) - 9 tasks
4. Reconvene for Polish phase validation

---

## Task Summary

**Total Tasks**: 65 tasks across 6 phases

**By Phase**:
- Phase 1 (Setup): 4 tasks
- Phase 2 (Foundational): 7 tasks
- Phase 3 (US2): 11 tasks
- Phase 4 (US1): 20 tasks
- Phase 5 (US3): 9 tasks
- Phase 6 (Polish): 14 tasks

**By User Story**:
- User Story 2 (Configure Base URL): 11 tasks
- User Story 1 (RSS Feed Absolute URLs): 20 tasks
- User Story 3 (Dashboard Display): 9 tasks

**Parallel Opportunities**: 19 tasks marked [P] (29% parallelizable)

**MVP Scope**: 42 tasks (Phases 1-4)

**Independent Test Criteria**:
- US2: Server starts with valid config, fails with invalid config
- US1: RSS feed contains only absolute URLs, passes validation
- US3: Dashboard displays absolute feed URL from config

---

## Notes

- [P] tasks = different files or different sections, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story can be independently completed and tested
- TDD approach: Write tests first, verify they fail, implement, verify they pass
- Tests are included per feature specification request
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Configuration (US2) validates the infrastructure that US1 and US3 use
