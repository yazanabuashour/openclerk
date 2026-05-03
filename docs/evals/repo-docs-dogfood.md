# Repo Docs Dogfood Eval

This targeted lane dogfoods OpenClerk against this repository's committed
public markdown docs. The harness copies the repository into an isolated eval
run directory, imports eligible `.md` files into the eval vault through the
OpenClerk document runner, and then asks the production agent to work only
through installed `openclerk document` and `openclerk retrieval` JSON results.

The lane is intentionally public-data-only. It does not read a private notes
vault, does not require personal data, and does not use direct vault
inspection, broad repo search, direct SQLite, source-built runner paths,
HTTP/MCP bypasses, backend variants, module cache inspection, or generated
server files.

Run the lane with:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario repo-docs-agentops-retrieval,repo-docs-synthesis-maintenance,repo-docs-decision-records,repo-docs-release-readiness,repo-docs-tag-filter,repo-docs-memory-router-recall-report,repo-docs-release-synthesis-freshness \
  --report-name ockp-repo-docs-dogfood
```

## Coverage

- `repo-docs-agentops-retrieval` checks filtered retrieval over architecture
  docs and requires citation evidence for the AgentOps-only production surface.
- `repo-docs-synthesis-maintenance` creates a source-linked synthesis note from
  committed eval docs and verifies sources, freshness, and duplicate avoidance.
- `repo-docs-decision-records` checks that canonical ADR markdown remains
  authoritative while decision records, projections, and provenance are
  runner-visible derived evidence.
- `repo-docs-release-readiness` checks whether imported release docs support
  mandatory dogfood before tagging and requires tag-filtered release-doc
  listing.
- `repo-docs-tag-filter` checks the read-side `tag` filter over imported
  public release docs while preserving canonical markdown authority.
- `repo-docs-memory-router-recall-report` checks the read-only
  `memory_router_recall_report` surface and its no-memory-transport boundary
  inside the dogfood lane.
- `repo-docs-release-synthesis-freshness` checks runner-visible synthesis
  projection freshness and provenance over release procedure docs without
  repairing the synthesis.

## Gate Status

This lane is mandatory pre-release evidence but remains separate from the full
AgentOps production gate. Before tagging a release, refresh the reduced
`docs/evals/results/ockp-repo-docs-dogfood.md` and `.json` reports and require
all selected dogfood scenarios to pass. Failures block tagging unless repaired
or explicitly classified as fixture/reporting-only defects.
