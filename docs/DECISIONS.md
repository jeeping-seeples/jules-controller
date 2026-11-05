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
