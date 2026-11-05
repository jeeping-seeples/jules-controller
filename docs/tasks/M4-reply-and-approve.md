# M4 — Reply and approve

## Scope

Multiline reply composer, send/cancel behavior, plan approval with explicit confirmation, and reconciliation after refresh.

## Acceptance criteria

- No message is sent without an explicit send action.
- Plan approval is available only in the documented awaiting-approval state.
- Approval requires explicit confirmation.
- Errors are actionable and optimistic UI is reconciled with API state.
- Key flows are tested.
