# jules-tui

A lightweight terminal client for Google Jules, written in Go with the Charm ecosystem.

> **Status:** project scaffold only. No Jules API behavior has been implemented or verified.

## Start here

1. Read [`AGENTS.md`](AGENTS.md).
2. Review [`docs/PLAN.md`](docs/PLAN.md) and approve or amend it before implementation.
3. Jules must verify the current Google Jules API documentation and current compatible Charm package versions before writing code.
4. Implement one milestone at a time, with tests and a separate commit/PR for each milestone.

## Project decisions

- The human remains in control: no message, plan approval, or other Jules action is sent without an explicit user action.
- The advisor is optional and advisory only. It never sends a Jules reply.
- No telemetry and no real credentials in the repository, logs, or tests.
- Unknown API fields, activity types, and states must be handled safely.
- Network and file I/O must not block the Bubble Tea update or view path.

## Documentation

- [`docs/PLAN.md`](docs/PLAN.md) — proposed architecture, milestones, and approval gate.
- [`docs/DECISIONS.md`](docs/DECISIONS.md) — decision log and unresolved questions.
- [`docs/VERIFICATION.md`](docs/VERIFICATION.md) — API and dependency verification checklist.
- [`docs/tasks/`](docs/tasks/) — milestone task briefs.

## Local development

The module path in `go.mod` is a placeholder. Replace `github.com/OWNER/jules-tui` with the actual repository path before publishing or importing this module.

No implementation dependencies are pinned yet. Resolve and record mutually compatible stable versions during the approved planning milestone.
