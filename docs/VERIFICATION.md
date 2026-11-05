# Verification ledger

Update this file as documentation and behavior are checked. Distinguish documented facts, tested behavior, and assumptions.

| Area | Verification required | Status |
|---|---|---|
| Jules Sessions API | Endpoints (`GET/POST /v1alpha/sessions`, `GET /v1alpha/sessions/{id}`, `POST /v1alpha/sessions/{id}:sendMessage`, `POST /v1alpha/sessions/{id}:approvePlan`), fields, states, pagination | Verified against official Jules API documentation & fixture unit tests |
| Jules Activities API | Activity variants, ordering, pagination, incremental retrieval | Verified against official Jules API documentation (implementation in M3) |
| Jules Sources API | Source listing (`GET /v1alpha/sources`), pagination, filtering, and repository selection | Verified against official Jules API documentation & unit tests |
| Activity `createTime` filter | Confirm semantics; otherwise use ID diff fallback | Checked docs (TIMESTAMP filter documented; fallback retained) |
| Vercel AI Gateway | OpenAI-compatible URL, auth, streaming, models list | Not checked (deferred to M6) |
| Charm dependencies | Stable compatible versions and import paths | Deferred to M2 TUI milestone |
| teatest | Current package path and supported test APIs | Deferred to M2 TUI milestone |
| Live Jules service | Opt-in verification only (`JULES_LIVE_TEST=1`) | Unverified against live service (no live key in environment; fixture tests pass) |

## Live test policy

Live tests may run only when `JULES_LIVE_TEST=1` and a Jules API key is present. Never print the key or authorization header. Fixture-based tests remain the default.
