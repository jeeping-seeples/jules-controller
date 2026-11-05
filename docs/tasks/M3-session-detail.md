# M3 — Session detail

## Scope

Scrollable activity transcript, Markdown rendering, plan rendering, artifact summaries, diffs on demand, and navigation keys.

## Acceptance criteria

- Agent messages are displayed in full.
- Unknown activity types remain visible with raw JSON available.
- Large transcripts and diffs do not require rendering everything at once.
- Activity rendering is pure and covered by fixture/golden tests.
