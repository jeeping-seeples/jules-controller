# Proposed implementation plan — approval required

**Status:** Draft. No implementation should begin until the owner explicitly approves this plan.

## Goal

Build `jules-tui`, a low-resource keyboard-first terminal client for Google Jules, with scriptable commands and an optional privacy-controlled Vercel AI Gateway advisor.

## Proposed architecture

- `cmd/jules-tui`: entry point and non-interactive command parsing.
- `internal/julesapi`: Jules HTTP client, tolerant API types, pagination, retries, and request/response handling.
- `internal/config`: OS config/state paths, TOML configuration, environment overrides, secure file permissions.
- `internal/poller`: background refresh scheduling, state transition detection, and persisted notification markers.
- `internal/notify`: terminal bell, in-app notifications, and best-effort desktop notifications.
- `internal/render`: pure rendering helpers for Markdown, plans, activities, and diffs.
- `internal/tui`: Bubble Tea screens and typed messages.
- `internal/advisor`: Vercel AI Gateway client, streaming parser, context construction, consent/preview, and redaction.
- `testdata`: sanitized Jules API fixtures.

Keep API and UI layers separate. Share the same API client between the TUI and scriptable commands. Use commands for I/O and keep `Update`/`View` non-blocking.

## Milestones

1. API client, config, and scriptable commands (`sessions`, `show`, `send`, `approve`, `new`).
2. Home screen, grouped projects, polling, transition detection, and notifications.
3. Session detail transcript, Markdown, plans, artifacts, and diffs.
4. Reply composition and confirmed plan approval.
5. New-session workflow and task-file launcher.
6. Optional advisor, streaming, consent, preview, redaction, and editable suggested replies.
7. UX polish, docs, CI, release configuration, and cross-platform verification.

Each milestone must have focused tests and a reviewable change set. Do not start the next milestone until the current one is reviewed.

## Verification gates

Before implementation:
- Read current official Jules Sessions and Activities references and the linked Sources reference.
- Record documented endpoints, request/response shapes, enum values, pagination, and known ambiguities.
- Verify current Vercel AI Gateway OpenAI-compatible API details and whether a models-list endpoint is documented.
- Verify current stable compatible Charm library versions and import paths, including teatest.
- Confirm the Go version and dependency compatibility.

For every milestone:
- Run relevant unit and `httptest` tests.
- Run `go test ./...` and `go vet ./...` when the repository is buildable.
- Use fixtures by default; live tests only when explicitly enabled.
- Record what was and was not verified against the live Jules service.

## Open questions to resolve during planning

1. What exact Jules API operations and fields are documented for creating, replying to, approving, listing, and deleting sessions?
2. What are the complete documented session states and activity variants?
3. What pagination and incremental activity retrieval behavior is guaranteed?
4. Does the current Vercel AI Gateway documentation expose a model-list endpoint suitable for a picker?
5. Which stable Charm package major versions are mutually compatible?
6. Which desktop-notification approach is sufficiently small and reliable across supported operating systems?

## Approval

Owner decision: **PENDING**

Do not write implementation code until the owner approves this plan or supplies a revised plan.
