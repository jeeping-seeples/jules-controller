# M1 — API client and scriptable commands

## Scope

Implement the Jules API client, config foundation, and scriptable commands:
- `jules-tui sessions [--json] [--state ...]`
- `jules-tui show <session> [--json]`
- `jules-tui send <session> --message-file -`
- `jules-tui approve <session>`
- `jules-tui new --repo owner/name --branch ... --prompt-file ... [--require-approval]`

## Acceptance criteria

- Every endpoint and payload is verified against current official docs.
- Unknown JSON fields, activity types, and session states do not crash decoding.
- Errors go to stderr; exit codes are documented and stable.
- JSON output is stable and documented.
- No credentials are printed or logged.
- Fixture tests cover documented response variants, pagination, auth failures, 429/5xx behavior, and unknown fields.
- Live tests are opt-in only.

## Out of scope

TUI screens, polling, notifications, advisor, and local checkout integration.
