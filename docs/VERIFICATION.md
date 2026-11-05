# Verification ledger

Update this file as documentation and behavior are checked. Distinguish documented facts, tested behavior, and assumptions.

| Area | Verification required | Status |
|---|---|---|
| Jules Sessions API | Endpoints, fields, states, pagination, create/reply/approve/delete semantics | Not checked |
| Jules Activities API | Activity variants, ordering, pagination, incremental retrieval | Not checked |
| Jules Sources API | Source listing and repository selection | Not checked |
| Activity `createTime` filter | Confirm semantics; otherwise use ID diff fallback | Not checked |
| Vercel AI Gateway | OpenAI-compatible URL, auth, streaming, models list | Not checked |
| Charm dependencies | Stable compatible versions and import paths | Not checked |
| teatest | Current package path and supported test APIs | Not checked |
| Live Jules service | Opt-in verification only | Not checked |

## Live test policy

Live tests may run only when `JULES_LIVE_TEST=1` and a Jules API key is present. Never print the key or authorization header. Fixture-based tests remain the default.
