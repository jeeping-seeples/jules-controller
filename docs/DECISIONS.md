# Decision log

Record decisions here with date, context, decision, and consequences.

## Initial constraints

- The application is a standalone Go repository.
- Charm/Bubble Tea is the intended terminal UI stack, subject to current-version verification.
- Jules API behavior must be grounded in current official documentation.
- The advisor is optional, consent-gated, previewable, and advisory only.
- No telemetry.
- No automatic Jules actions.
- The owner must approve the implementation plan before coding.

## Milestone 1 Decisions

- **Source resolution (`new`)**:
  - Passthrough raw `sources/...` identifiers.
  - Query `GET /v1alpha/sources` with pagination and match `owner/repo`.
  - Fall back to derived `sources/github-<owner>-<repo>` with a warning if the list call fails.
- **`send` message flags**:
  - Accept `--message-file <path|->` and `--message "text"` (mutually exclusive, exactly one required).
- **`show` metadata scope**:
  - Show metadata and output pull requests.
  - Activity details deferred to future `--activities` / milestone 3.
- **CLI & config dependencies**:
  - Go standard library `flag` package with subcommands and custom arg splitting for positional arguments.
  - `github.com/pelletier/go-toml/v2` for config parsing.
  - Standard library `text/tabwriter` for `sessions` table output.
- **Exit codes**:
  - `0`: Success
  - `1`: General error
  - `2`: Usage error
  - `3`: Auth error (401/403 or missing API key)
  - `4`: Resource not found (404)
  - `5`: API/Network error after retries
- **Retries**:
  - Backoff retry on 429 and 5xx responses, respecting `Retry-After` header when provided.
- **Credentials**:
  - API key precedence: `--api-key` flag > `JULES_API_KEY` env var > `config.toml`.
  - Transmitted exclusively via `x-goog-api-key` header, never in query string, never logged.

## Milestone 2 Decisions

- **Charm ecosystem**:
  - Added `github.com/charmbracelet/bubbletea` v1.3.4, `github.com/charmbracelet/lipgloss` v1.0.0, and `github.com/charmbracelet/bubbles` v0.20.0.
- **TUI & CLI integration**:
  - Launching `jules-tui` without arguments (or with `jules-tui tui`) launches the interactive Home screen.
  - Scriptable subcommands (`sessions`, `show`, `send`, `approve`, `new`) continue to operate non-interactively.
- **Grouping & Urgency Triage**:
  - Grouped by repository/source name (`owner/repo`), with fallback to `Repoless / Unknown`.
  - Triage levels: UrgencyHigh (`AWAITING_PLAN_APPROVAL`, `AWAITING_USER_FEEDBACK`), UrgencyMedium (`IN_PROGRESS`, `PLANNING`, `QUEUED`), UrgencyLow (`PAUSED`, `COMPLETED`, `FAILED`).
  - High urgency sessions are prioritized and sorted at top within each group.
  - "Needs Me" filter toggle (`tab` or `n`) isolates sessions requiring human action.
- **Collapse/Expand**:
  - Groups can be collapsed or expanded with `space` or `enter`.
- **Notifications**:
  - Persistent state saved to XDG state path `~/.local/state/jules-tui/notifications.json`.
  - Transitions notify only once.
  - Notifier emits terminal bell `\a` and best-effort OS notification (`osascript` on macOS, `notify-send` on Linux).
- **Adaptive Polling**:
  - Configurable 30s base interval with exponential backoff on 429/5xx and network errors up to 5m cap.
  - Respects cancellation and runs asynchronously outside the Bubble Tea View/Update thread.
- **NO_COLOR**:
  - Clean text and styling fallback when `NO_COLOR` is detected in environment.
