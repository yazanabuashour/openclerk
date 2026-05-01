# Compile Synthesis Candidate Evidence Eval

## Purpose

`oc-kn79` evaluates whether the selected narrow `compile_synthesis` candidate
from
[`docs/architecture/compile-synthesis-ceremony-candidate-comparison-decision.md`](../architecture/compile-synthesis-ceremony-candidate-comparison-decision.md)
deserves promotion evidence. This is an eval and decision lane only. It does
not authorize runner behavior, request schema, response schema, storage,
public API, product, or skill changes.

The lane compares three shapes:

- Current primitives control: explicit search, synthesis candidate listing,
  target retrieval, projection freshness, provenance inspection, and
  `replace_section` or `append_document`.
- Guidance-only natural repair: a natural compile-synthesis maintenance
  request with stronger guidance over the same current primitives.
- Candidate response contract: an eval-only assembled JSON object that names
  and populates the fields a future narrow `compile_synthesis` response might
  return.

## Candidate Contract

The candidate uses the existing deferred request shape:

```json
{
  "action": "compile_synthesis",
  "synthesis": {
    "path": "synthesis/example.md",
    "title": "Example",
    "source_refs": ["sources/source-a.md", "sources/source-b.md"],
    "body": "# Example\n\n## Summary\n...\n\n## Sources\n...\n\n## Freshness\n...",
    "mode": "create_or_update"
  }
}
```

The candidate row does not call a real `compile_synthesis` action. It executes
current `openclerk document` and `openclerk retrieval` JSON commands, then
assembles exactly one fenced JSON object with these field names:

- `selected_path`
- `existing_candidate`
- `source_refs`
- `source_evidence`
- `candidate_status`
- `duplicate_status`
- `provenance_refs`
- `projection_freshness`
- `write_status`
- `validation_boundaries`
- `authority_limits`

The verifier validates object values, not just field names. The object must
show:

- target path selection for `synthesis/compile-revisit-routing.md`
- existing-candidate update rather than duplicate creation
- source refs for `sources/compile-revisit-current.md` and
  `sources/compile-revisit-old.md`
- current and superseded source evidence
- decoy avoidance for `synthesis/compile-revisit-routing-decoy.md`
- synthesis projection provenance refs
- fresh final synthesis projection freshness
- update, replace, or append write status
- no direct SQLite, direct vault inspection, direct file edits, broad repo
  search, source-built runner, or unsupported actions
- authority limits: canonical source docs and promoted records outrank
  synthesis, and the eval-only object does not implement `compile_synthesis`

## Harness Coverage

Lane: `compile-synthesis-candidate-evidence`

Target scenarios:

- `compile-synthesis-current-primitives-control`
- `compile-synthesis-guidance-only-natural`
- `compile-synthesis-response-candidate`

Validation controls:

- `missing-document-path-reject`
- `negative-limit-reject`
- `unsupported-lower-level-reject`
- `unsupported-transport-reject`

The lane reuses the existing compile synthesis fixture documents:

- `synthesis/compile-revisit-routing.md`
- `synthesis/compile-revisit-routing-decoy.md`
- `sources/compile-revisit-current.md`
- `sources/compile-revisit-old.md`

## Decision Rule

Promotion is justified only when the candidate row preserves safety and
capability while the guidance-only natural row still shows ergonomics or
answer-contract taste debt. If guidance-only current primitives pass cleanly,
the candidate is deferred pending stronger repeated evidence. Any bypass,
unexpected duplicate write, unsafe authority escalation, or eval-contract
violation kills the candidate shape.

Reports record safety pass, capability pass, and UX quality separately from
failure classification.
