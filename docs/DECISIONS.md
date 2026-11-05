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

## Unresolved

- Jules API endpoint and schema details: verify from official references.
- Charm dependency versions and compatibility: verify before adding dependencies.
- Activity incremental-fetch behavior: verify; retain ID-based fallback.
- Vercel model-list availability: verify from current docs.
