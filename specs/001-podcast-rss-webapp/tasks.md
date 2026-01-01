# Tasks: Podcast RSS Webapp

**Input**: Design documents from `/home/devin/Projects/rss-server/specs/001-podcast-rss-webapp/`
**Prerequisites**: plan.md (✓), spec.md (✓), research.md (✓), data-model.md (✓), contracts/ (✓)

**Tests**: Tests are OPTIONAL and NOT included per feature specification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Project structure follows Go conventions from plan.md:
- `cmd/server/` - Application entry point
- `internal/` - Private application code
- `web/` - Frontend templates and static assets
- `data/` - Runtime data storage
- `tests/` - Test organization (NOT included per spec)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 [P] Initialize Go module with `go mod init` at repository root (go.mod, go.sum)
- [X] T002 [P] Create project directory structure per plan.md: cmd/server/, internal/{handlers,rss,storage,models}/, web/{templates,static}/, data/{audio,artwork}/
- [X] T003 [P] Add dependency `github.com/eduncan911/podcast` to go.mod via `go get`
- [X] T004 [P] Create basic config.yaml in repository root with server port (8080), file size limit (500MB), data paths
- [X] T005 [P] Create web/static/styles.css with minimal CSS styles for dashboard UI

**Checkpoint**: Project structure initialized and dependencies ready

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T006 [P] Create Episode model struct in internal/models/episode.go with RSS 2.0 + iTunes fields per data-model.md
- [X] T007 [P] Create Podcast model struct in internal/models/podcast.go with channel-level metadata per data-model.md
- [X] T008 [P] Create AudioFile metadata struct in internal/models/episode.go for file management
- [X] T009 Implement RSSStore with sync.RWMutex in internal/storage/xml.go for concurrent RSS XML access per research.md
- [X] T010 Implement LoadRSSStore() function in internal/storage/xml.go to load/initialize podcast.xml with default values
- [X] T011 Implement atomic XML write function in internal/storage/xml.go (temp file + rename pattern)
- [X] T012 Implement RSS feed generation in internal/rss/feed.go using github.com/eduncan911/podcast library
- [X] T013 Implement RSS XML parsing in internal/rss/parser.go to read existing podcast.xml into Podcast struct
- [X] T014 [P] Implement filesystem audio file save in internal/storage/filesystem.go with unique filename generation
- [X] T015 [P] Implement filesystem audio file delete in internal/storage/filesystem.go with cleanup logic
- [X] T016 Create basic HTTP server in cmd/server/main.go with net/http stdlib and route registration
- [X] T017 Configure server to serve static files from web/static/ directory in cmd/server/main.go
- [X] T018 Create main HTML template in web/templates/index.html with HTMX 1.9+ CDN and base dashboard layout

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Upload Audio and Generate Basic RSS Feed (Priority: P1) 🎯 MVP

**Goal**: Enable podcast creators to upload audio files and generate a valid RSS feed that can be submitted to podcast directories

**Independent Test**: Upload a single MP3 file through web interface at http://localhost:8080, then access RSS feed at http://localhost:8080/feed.xml in browser/validator and verify feed contains episode with valid podcast metadata

### Implementation for User Story 1

- [X] T019 [P] [US1] Implement POST /api/episodes handler in internal/handlers/episodes.go for multipart file upload per openapi.yaml
- [X] T020 [P] [US1] Implement GET /feed.xml handler in internal/handlers/feed.go to serve RSS XML with Content-Type: application/rss+xml
- [X] T021 [P] [US1] Implement GET /audio/{filename} handler in internal/handlers/static.go to stream audio files with Content-Type: audio/mpeg
- [X] T022 [US1] Add file validation logic in internal/handlers/episodes.go: check .mp3 extension and 500MB size limit (FR-010, FR-011)
- [X] T023 [US1] Implement episode ID generation function in internal/models/episode.go using date + sanitized title per data-model.md
- [X] T024 [US1] Implement unique filename generation in internal/storage/filesystem.go to prevent conflicts (data-model.md line 258-271)
- [X] T025 [US1] Integrate upload handler with RSSStore.AddEpisode() to update podcast.xml atomically
- [X] T026 [US1] Implement episode sorting by PubDate (descending) in internal/rss/feed.go per FR-008
- [X] T027 [US1] Add automatic cleanup for failed uploads in internal/handlers/episodes.go per clarification
- [X] T028 [US1] Create HTMX episode upload form partial in web/templates/components/upload_form.html with file input, title, description fields
- [X] T029 [US1] Create HTMX episode list partial in web/templates/components/episode_list.html to display episodes with HTMX swap
- [X] T030 [US1] Implement GET / dashboard handler in internal/handlers/web.go to render index.html with episode list
- [X] T031 [US1] Implement GET /api/episodes handler in internal/handlers/episodes.go to return episode list (JSON or HTML partial)
- [X] T032 [US1] Add HTMX upload progress indicator in web/templates/components/upload_form.html
- [X] T033 [US1] Wire up routes in cmd/server/main.go: POST /api/episodes, GET /feed.xml, GET /audio/{filename}, GET /, GET /api/episodes

**Checkpoint**: At this point, User Story 1 should be fully functional - can upload episodes and generate valid RSS feed independently

---

## Phase 4: User Story 2 - Customize Podcast Metadata (Priority: P2)

**Goal**: Enable podcast creators to customize podcast-level metadata (title, author, artwork, description, category) for professional branding when submitting to directories

**Independent Test**: Configure podcast metadata (title, author, description, artwork) through web interface settings page, then verify RSS feed at http://localhost:8080/feed.xml includes all iTunes-required tags (itunes:author, itunes:image, itunes:category, itunes:summary) and displays correctly when validated

### Implementation for User Story 2

- [X] T034 [P] [US2] Implement POST /api/podcast/settings handler in internal/handlers/episodes.go for multipart form with artwork upload per openapi.yaml
- [X] T035 [P] [US2] Implement GET /api/podcast/settings handler in internal/handlers/episodes.go to return current podcast metadata (JSON or HTML)
- [X] T036 [US2] Add artwork file upload handling in internal/handlers/episodes.go: validate image format and 5MB size limit
- [X] T037 [US2] Implement artwork file save in internal/storage/filesystem.go to data/artwork/ directory
- [X] T038 [US2] Integrate settings handler with RSSStore.UpdatePodcast() to update channel-level metadata atomically
- [X] T039 [US2] Add podcast metadata validation in internal/handlers/episodes.go: required fields, URL format, language code pattern per data-model.md
- [X] T040 [US2] Create HTMX settings form partial in web/templates/components/settings_form.html with fields for title, author, description, link, language, category, explicit, artwork
- [X] T041 [US2] Add iTunes category dropdown in settings form with valid categories from data-model.md lines 452-473
- [X] T042 [US2] Wire up routes in cmd/server/main.go: POST /api/podcast/settings, GET /api/podcast/settings
- [X] T043 [US2] Add settings link to dashboard navigation in web/templates/index.html
- [X] T044 [US2] Implement default podcast metadata initialization in internal/storage/xml.go per data-model.md lines 97-114

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - can upload episodes AND customize podcast branding

---

## Phase 5: User Story 3 - Delete Episodes (Priority: P3)

**Goal**: Enable podcast creators to delete episodes from feed and storage for content management

**Independent Test**: Upload an episode through web interface, then delete it using the delete button, and verify it is removed from RSS feed at http://localhost:8080/feed.xml and audio file is deleted from data/audio/ directory

### Implementation for User Story 3

- [X] T045 [US3] Implement DELETE /api/episodes/{id} handler in internal/handlers/episodes.go per openapi.yaml
- [X] T046 [US3] Implement RSSStore.DeleteEpisode() in internal/storage/xml.go to remove episode from XML and update atomically
- [X] T047 [US3] Integrate delete handler with filesystem.DeleteAudioFile() to remove audio file after XML update
- [X] T048 [US3] Add error handling for non-existent episode deletion (404 response) in internal/handlers/episodes.go
- [X] T049 [US3] Add delete button to episode row in web/templates/components/episode_list.html with HTMX hx-delete and hx-confirm
- [X] T050 [US3] Wire up route in cmd/server/main.go: DELETE /api/episodes/{episodeId}
- [X] T051 [US3] Implement HTMX episode removal on successful delete (hx-target="closest .episode-row" hx-swap="outerHTML")

**Checkpoint**: All user stories should now be independently functional - full CRUD for episodes + metadata customization

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T052 [P] Add local RSS validation function in internal/rss/validator.go for basic field checks per research.md lines 401-436
- [ ] T053 [P] Add proper error responses (400, 413, 415, 500) with JSON error schema from openapi.yaml to all handlers
- [X] T054 [P] Add proper XML character escaping for episode titles and descriptions in internal/rss/feed.go per FR-014
- [X] T055 [P] Add GUID generation for episodes in internal/models/episode.go (use episode ID as GUID)
- [ ] T056 Implement audio duration calculation in internal/storage/filesystem.go (optional but recommended per data-model.md line 188)
- [X] T057 Add logging for all HTTP requests and errors in cmd/server/main.go using stdlib log package
- [X] T058 [P] Create README.md in repository root with quickstart instructions from quickstart.md
- [X] T059 [P] Add HTMX loading indicators and error messages to all forms in web/templates/
- [X] T060 [P] Add CSS styling for episode list, forms, and buttons in web/static/styles.css
- [X] T061 Test RSS feed with W3C Feed Validator manually at https://validator.w3.org/feed/ per quickstart.md
- [X] T062 Verify all constitutional requirements: RSS compliance, file-centric storage, HTTP endpoints, simplicity

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Independent of US1 but both enhance feed quality
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Independent of US1/US2, adds content management

### Within Each User Story

- Models before services (T006-T008 before T009-T013)
- Storage layer before handlers (T009-T015 before T019-T033)
- Handlers before templates (T019-T021 before T028-T029)
- Core implementation before integration (T022-T027 before T033)
- Story complete before moving to next priority

### Parallel Opportunities

**Phase 1 (Setup)**: All tasks T001-T005 can run in parallel

**Phase 2 (Foundational)**:
- Models (T006, T007, T008) can run in parallel
- Storage (T014, T015) can run in parallel after models complete

**Phase 3 (User Story 1)**:
- Handlers (T019, T020, T021) can run in parallel after foundational
- Templates (T028, T029) can run in parallel

**Phase 4 (User Story 2)**:
- Handlers (T034, T035) can run in parallel
- Can start in parallel with US3 after foundational completes

**Phase 5 (User Story 3)**:
- Can start in parallel with US2 after foundational completes

**Phase 6 (Polish)**:
- Validation, error handling, logging (T052, T053, T054, T055, T059, T060) can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all independent handlers for User Story 1 together:
Task: "Implement POST /api/episodes handler in internal/handlers/episodes.go"
Task: "Implement GET /feed.xml handler in internal/handlers/feed.go"
Task: "Implement GET /audio/{filename} handler in internal/handlers/static.go"

# Launch all templates for User Story 1 together:
Task: "Create HTMX episode upload form partial in web/templates/components/upload_form.html"
Task: "Create HTMX episode list partial in web/templates/components/episode_list.html"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T005)
2. Complete Phase 2: Foundational (T006-T018) - CRITICAL, blocks all stories
3. Complete Phase 3: User Story 1 (T019-T033)
4. **STOP and VALIDATE**: 
   - Start server: `go run cmd/server/main.go`
   - Upload test MP3 via http://localhost:8080
   - View feed at http://localhost:8080/feed.xml
   - Validate with W3C Feed Validator
5. Deploy/demo if ready - this is a functional MVP!

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → **Deploy/Demo (MVP!)**
3. Add User Story 2 → Test independently → **Deploy/Demo** (professional branding)
4. Add User Story 3 → Test independently → **Deploy/Demo** (content management)
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (T001-T018)
2. Once Foundational is done:
   - Developer A: User Story 1 (T019-T033) - Core upload functionality
   - Developer B: User Story 2 (T034-T044) - Metadata customization
   - Developer C: User Story 3 (T045-T051) - Episode deletion
3. Stories complete and integrate independently
4. Team reconvenes for Phase 6 Polish

---

## Task Summary

**Total Tasks**: 62 tasks across 6 phases

**Breakdown by Phase**:
- Phase 1 (Setup): 5 tasks
- Phase 2 (Foundational): 13 tasks
- Phase 3 (US1 - P1 MVP): 15 tasks
- Phase 4 (US2 - P2): 11 tasks
- Phase 5 (US3 - P3): 7 tasks
- Phase 6 (Polish): 11 tasks

**Breakdown by User Story**:
- Setup + Foundational: 18 tasks (prerequisite for all stories)
- User Story 1 (P1): 15 tasks - Upload audio and generate RSS feed (MVP)
- User Story 2 (P2): 11 tasks - Customize podcast metadata
- User Story 3 (P3): 7 tasks - Delete episodes
- Cross-cutting (Polish): 11 tasks

**Parallel Opportunities**: 18 tasks marked [P] for parallel execution

**MVP Scope**: Phase 1 + Phase 2 + Phase 3 (33 tasks) delivers functional podcast RSS feed generator

---

## Notes

- [P] tasks = different files, no dependencies - can run in parallel
- [US1]/[US2]/[US3] labels map tasks to specific user stories for traceability
- Each user story is independently completable and testable per spec requirements
- Constitution compliance verified: RSS 2.0 compliance (FR-003), file-centric (XML source of truth), HTTP-first (all endpoints), simplicity (minimal dependencies)
- Tests NOT included per feature specification (optional per spec line 87)
- All file paths are exact per plan.md project structure
- Commit after each task or logical group (e.g., all Phase 1 tasks, all models, complete user story)
- Stop at any checkpoint to validate story independently
- RSS feed validation with W3C Feed Validator is manual in T061 (not automated per research findings)
