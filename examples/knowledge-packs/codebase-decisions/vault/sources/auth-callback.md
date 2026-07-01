---
type: source
status: active
---
# Auth Callback Source

## Summary
The auth callback validates the state value, exchanges the short-lived code,
and returns users to the requested repo-relative route.

## Notes
- Preserve the state value check before any redirect.
- Keep callback errors visible to the caller.
