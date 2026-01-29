# Quality Initiative: Testing, Semantics, Documentation

**Date:** 2026-01-28
**Status:** Brainstorm complete, ready for planning

## What We're Building

A unified quality initiative addressing three areas of the Yutani terminal display server:

1. **Semantic consistency** -- Fix proto-level naming inconsistencies and Go-side gaps
2. **Acceptance testing** -- Automated end-to-end workflow tests and API contract tests
3. **Documentation cleanup** -- Audit stale files, expand AGENTS.md, fill contributor gaps

## Why This Approach

**Bottom-Up ordering:** Fix semantics first, then write tests against the corrected API, then document the final state. This avoids rewriting tests and docs after naming changes. Since there are no external consumers of the gRPC API yet, breaking proto changes are acceptable now.

## Key Decisions

### Semantic Consistency

- **Proto naming changes are in scope.** Breaking changes accepted since there are no external consumers.
- **Specific issues to fix:**
  - Standardize Selection/Selected naming across List, Table, Tree services
  - Standardize Clear message naming (prefix consistency)
  - Disambiguate AddChild/AddItem/AddPage naming
  - Add missing widget types (Image, ProgressBar) to ServerCapabilities in `session.go`
  - Register TestService in the integration test harness (`pkg/client/testing/integration.go`)
  - Register TestService in E2E test setup (`test/e2e/e2e_test.go`)
  - Resolve `NewServer` vs `New` constructor ambiguity

### Acceptance Testing

- **Two test categories:**
  - **User workflow tests:** Full client-to-server scenarios (create session, build UI, interact, verify screen)
  - **API contract tests:** Every gRPC method returns correct responses, errors, and edge cases per proto spec
- **Use the headless TestService** (InjectKey, InjectText, InjectMouse, WaitForIdle) for interaction testing once it's registered in the test harness
- **Replace `time.Sleep` synchronization** in test setup with proper readiness checks where possible
- **Consolidate duplicated test helpers** (`boolPtr`, `strPtr`, `int32Ptr`) into a shared package

### Documentation Cleanup

- **Audit each stale root-level markdown file** individually (17+ files: CHANGES_SUMMARY.md, BUILD_FIXES_SUMMARY.md, DISPLAY_FIX.md, etc.) to decide keep/delete/merge
- **Expand AGENTS.md** with architecture overview, coding conventions, tview thread safety rules, service implementation patterns, testing requirements, proto workflow
- **Fill gaps:** CONTRIBUTING.md, LICENSE (both currently TBD in README)
- **Update PRD.md** to reflect current state (Phase 6 checkboxes, directory structure)

## Open Questions

1. What should the standardized naming pattern be for Selection operations? Options: `GetSelected`/`SetSelected` (List's current pattern) vs `GetSelection`/`SetSelection` (Table's current pattern)
2. Should `NewServer` be removed in favor of `New` with config, or should both exist?
3. What license should the project use?
4. Should acceptance tests live in `test/e2e/` alongside existing E2E tests, or in a new `test/acceptance/` directory?

## Execution Order

1. Semantic fixes (proto + Go)
2. Acceptance test infrastructure + tests
3. Documentation audit and expansion

## Scope Boundaries

- **In scope:** Proto naming fixes, Go-side consistency fixes, new acceptance tests, documentation audit and expansion
- **Out of scope:** New features, performance optimization, CI/CD pipeline changes, dependency upgrades
