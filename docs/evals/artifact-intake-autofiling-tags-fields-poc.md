---
eval_id: poc-artifact-intake-autofiling-tags-fields
status: complete
source_refs: docs/architecture/artifact-intake-autofiling-tags-fields-adr.md, docs/architecture/agent-knowledge-plane.md
---

# POC: Artifact Intake, Auto-Filing, Tags, and Fields

## Required References

- docs/architecture/agent-knowledge-plane.md
- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md
- https://mitchellh.com/writing/building-block-economy
- https://developers.openai.com/api/docs/guides/prompt-guidance
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/embeddings
- https://developers.openai.com/api/docs/guides/retrieval
- https://docs.mem0.ai/open-source/overview

## Candidate Surfaces

| Candidate | Safety | Capability | UX Quality | Outcome |
| --- | --- | --- | --- | --- |
| Current `validate` plus skill recipe | Pass | Partial | Weak; too much prompt choreography for path/title/body/tags/fields/duplicates | Keep as primitive |
| `create_document` dry-run mode | Pass if read-only | Partial | Confuses validation with domain planning and does not naturally handle source URL handoff | Reject |
| New read-only `artifact_candidate_plan` | Pass | Pass | Strong; one natural runner action returns the full proposal and handoff | Promote |
| Generalized `ingest_artifact` | Risky | Future-only | Too broad; implies parser/OCR truth and durable acquisition | Reject |

## Selected Contract

Request:

```json
{
  "action": "artifact_candidate_plan",
  "artifact": {
    "content": "# Receipt\n\nTotal paid: 42 USD",
    "artifact_kind": "receipt",
    "tags": ["finance"],
    "fields": {"owner": "ap"},
    "limit": 5
  }
}
```

Result fields:

- `artifact_candidate_plan.candidate_path`
- `artifact_candidate_plan.candidate_title`
- `artifact_candidate_plan.body_preview`
- `artifact_candidate_plan.tags`
- `artifact_candidate_plan.metadata_fields`
- `artifact_candidate_plan.duplicate_search`
- `artifact_candidate_plan.likely_duplicate`
- `artifact_candidate_plan.existing_source`
- `artifact_candidate_plan.confidence`
- `artifact_candidate_plan.next_create_document_request`
- `artifact_candidate_plan.next_ingest_source_request`
- `artifact_candidate_plan.agent_handoff`

## Safety Contract

The action is read-only. It never fetches a URL, reads a local file, parses an
opaque artifact, performs OCR, writes markdown, touches SQLite directly, or uses
non-runner transports. Public URL context may produce an `ingest_source_url`
handoff, but durable fetch/write still requires approval.

## POC Outcome

Promote `artifact_candidate_plan`. The evaluated shape satisfies the real need
while preserving canonical markdown authority and approval-before-write.
