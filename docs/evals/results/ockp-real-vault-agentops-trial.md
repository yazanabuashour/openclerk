# OpenClerk Real-Vault AgentOps Trial

This is a sanitized real-vault trial report for `oc-rvt`.

## Status

Result: completed with sanitized evidence.

The trial used a harness-managed current-branch runner install as production-like
setup infrastructure. The runner was built into `<run-root>/bin/openclerk`,
`<run-root>/bin` was prepended to `PATH`, and private OpenClerk storage was
provided through `OPENCLERK_DATA_DIR`, `OPENCLERK_DATABASE_PATH`, and
`OPENCLERK_VAULT_ROOT`.

The workflow under test used only installed runner JSON through
`openclerk document` and `openclerk retrieval`. The trial did not use direct
SQLite, broad repo search, HTTP/MCP transports, source-built runner paths during
workflow execution, direct vault inspection, copied vault files, screenshots, or
raw logs as evidence.

## Environment Preflight

| Check | Result | Classification |
| --- | --- | --- |
| Harness-managed current-branch runner available at `<run-root>/bin/openclerk` | pass | setup |
| Private vault root available to the runner through `OPENCLERK_VAULT_ROOT` | pass | setup |
| Isolated data directory and database available through `OPENCLERK_DATA_DIR` and `OPENCLERK_DATABASE_PATH` | pass | setup |
| `inspect_layout` returned runner-visible layout state | pass | setup |

The setup mirrors the production-like AgentOps eval pattern: harness setup may
prepare the installed runner, but the workflow itself only invokes
`openclerk document` and `openclerk retrieval`.

## Workflow Coverage

| Workflow label | Runner actions used | Result | Existing runner actions sufficient? | Failure classification | Follow-up |
| --- | --- | --- | --- | --- | --- |
| Source discovery from canonical docs | `list_documents`, `get_document`, `search` | pass | yes | none | none |
| Synthesis create or update | `list_documents`, `create_document`, `replace_section` | pass | yes | none | none |
| Freshness/provenance inspection | `projection_states`, `provenance_events` | pass | yes | none | none |
| Decision-record lookup | `decisions_lookup`, `decision_record` | pass | yes | none | none |
| Stale or duplicate synthesis detection | `create_document` duplicate-path rejection | pass | yes | none | none |

## Sanitized Evidence

- Source discovery found registered private-vault documents, retrieved one
  document body through `get_document`, and returned cited search hits. The
  report intentionally omits private paths, titles, snippets, citations, and
  document identifiers.
- Synthesis create/update used neutral validation documents in `<private-vault>`
  and completed through documented document actions. The report intentionally
  omits validation document paths and raw JSON.
- Freshness/provenance inspection returned a synthesis projection with fresh
  state and at least one provenance event for the validation workflow.
- Decision-record lookup returned a validation decision record and a successful
  `decision_record` read through runner JSON.
- Duplicate synthesis detection rejected a duplicate path through the runner.

## Conclusion

For the tested real-vault workflows, the existing v1 AgentOps surfaces remain
sufficient:

- `openclerk document`
- `openclerk retrieval`

No product implementation follow-up is justified by this trial. No deferred
capability should be promoted from this evidence: Mem0 or memory API,
autonomous router, semantic graph layer, broad contradiction engine, and new
public runner actions remain deferred behind `docs/architecture/deferred-capability-promotion-gates.md`.

Released-binary validation remains useful as future release-parity evidence, but
`oc-rvt` is not blocked on a globally installed or released binary because the
harness-managed private install preserves the same agent-facing runner contract.
