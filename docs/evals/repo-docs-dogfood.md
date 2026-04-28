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
  --scenario repo-docs-agentops-retrieval,repo-docs-synthesis-maintenance,repo-docs-decision-records \
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

## Gate Status

This lane is not release-blocking. It is a recurring public dogfood signal for
retrieval quality, synthesis maintenance, and decision-record explainability on
real project documentation. Failures should be classified in the targeted lane
summary and used to decide whether the issue is fixture hygiene, skill guidance,
eval coverage, or a repeated runner capability gap.
