# Instructions for coding agents

## Mandatory workflow

1. Inspect the entire repository and its existing instructions before changing anything.
2. Do not begin implementation immediately. First produce a concise plan covering:
   - architecture and package boundaries;
   - milestone sequence and test strategy;
   - Jules API facts verified from current official documentation;
   - exact dependency versions and import paths verified from authoritative sources;
   - open questions, unknowns, and proposed resolutions.
3. Stop and wait for explicit human approval of the plan.
4. After approval, work on exactly one milestone at a time. Run its tests, summarize changes and verification, then create one commit or PR for that milestone only when explicitly authorized by the repository owner.
5. Never silently change scope. Record deviations and unresolved assumptions in `docs/DECISIONS.md` and `docs/VERIFICATION.md`.

## Source-of-truth rules

- Jules API references:
  - https://jules.google/docs/api/reference/sessions/
  - https://jules.google/docs/api/reference/activities/
  - Follow the official API documentation to the Sources reference.
- Verify current Vercel AI Gateway documentation before implementing the advisor:
  https://vercel.com/docs/ai-gateway
- Verify current stable compatible versions and import paths for Bubble Tea, Bubbles, Lip Gloss, Glamour, and teatest from their official repositories or pkg.go.dev.
- Never invent endpoints, fields, enum values, pagination behavior, or request semantics.
- In particular, do not assume the `createTime` activity filter is reliable unless verified. Implement ID-based deduplication/fallback.

## Safety, privacy, and control

- Never commit credentials, tokens, private keys, or real user data.
- Tests use sanitized fixtures. Live tests are opt-in only with `JULES_LIVE_TEST=1` and a key present.
- Never log API keys, authorization headers, or full sensitive request bodies.
- Redact obvious secrets before sending any context to the advisor.
- Advisor consent, exact payload preview, and explicit user action are required. Advisor auto mode defaults off.
- Never automatically send advisor output to Jules.
- Never automatically send a message, approve a plan, create a session, or delete a session.
- Destructive or irreversible actions require an explicit confirmation step.
- No telemetry.

## Engineering rules

- Go code must be formatted with `gofmt`.
- Keep API decoding tolerant of unknown JSON fields, activity types, and session states.
- Keep rendering pure and testable.
- Use Bubble Tea commands for network and file I/O. Respect context cancellation and timeouts.
- Avoid unnecessary background work; keep idle CPU and memory use low.
- Preserve keyboard accessibility and readable text when terminal colors are unavailable.
- Add tests with each behavior. Do not claim live API compatibility when only fixtures were tested.

## Completion report

At the end of each milestone, report:
- files changed;
- behavior implemented;
- tests and commands run, with results;
- live API behaviors verified vs. unverified;
- assumptions, deviations, and follow-up work.
