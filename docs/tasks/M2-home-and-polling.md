# M2 — Home screen and polling

## Scope

Grouped project/session overview, filters, collapse/expand, needs-me view, background refresh, adaptive polling, transition detection, and notifications.

## Acceptance criteria

- Human-action-needed states are prominent and sorted first.
- State text remains clear with `NO_COLOR`.
- Polling is cancellable, non-blocking, configurable, and backs off on errors/429s.
- Notification markers persist and transitions notify only once.
- Refresh errors are visible and actionable.
- Tests cover transition detection, persistence, and polling schedules.
