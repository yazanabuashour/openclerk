---
type: source
status: active
---
# Session Handling Source

## Summary
Session renewal happens after callback validation and before the user-visible
redirect.

## Notes
- Renewal failures should return a retryable error.
- Logging should avoid sensitive values.
