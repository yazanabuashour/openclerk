---
type: decision
decision_id: adr-auth-callback
decision_title: Keep callback validation before redirects
decision_status: accepted
decision_scope: auth
decision_owner: platform
decision_date: 2026-06-15
source_refs: sources/auth-callback.md, sources/session-handling.md
---
# Keep Callback Validation Before Redirects

## Decision
Validate callback state and renew the session before redirecting to a caller
route.

## Rationale
This keeps redirect behavior predictable and preserves a single validation
boundary for auth callback changes.
