# M6 — Advisor

## Scope

Optional Vercel AI Gateway client, streaming advisor panel, consent notice, exact payload preview, secret redaction, and editable suggested reply.

## Acceptance criteria

- Current gateway behavior is verified from official docs.
- No model ID is hard-coded as a default.
- Advisor auto mode defaults off and requires explicit opt-in.
- The exact outbound context is previewable.
- Redaction rules have positive and negative tests.
- Advisor output is never sent to Jules automatically.
- Cancellation stops streaming promptly.
