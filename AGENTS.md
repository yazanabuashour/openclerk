- For all committed docs, reports, and artifact references, use repo-relative paths or neutral repo-relative placeholders. Never use machine-absolute filesystem paths.
- For repo-pinned developer tools declared in `mise.toml`, run commands through `mise exec -- ...` so agents use the same tool versions as local docs and CI.

# ADR/POC/Eval Decision Taste Review

When doing OpenClerk ADR, POC, eval, promotion, or deferred-capability decision work, keep the existing evidence discipline but add a taste check before accepting a defer/reference outcome:

- Ask whether a normal user would expect a simpler OpenClerk surface than the one being preserved.
- Distinguish read/fetch/inspect permission from durable-write approval. A public user-provided URL can be enough to fetch through the runner; approval belongs at durable writes, privileged access, purchases/actions, or other irreversible changes.
- Prefer extending the natural existing runner action when the input clearly belongs there, instead of declaring the adjacent UX unsupported.
- Treat "completed but ceremonial" eval passes as possible taste debt when they require high step count, long latency, exact prompt choreography, or surprising clarification turns.
- Record safety pass, capability pass, and UX quality separately when a report or decision needs to justify defer/reference.
- When taste debt, defer, keep-as-reference, or another non-promotion outcome still leaves a real capability, ergonomics, safety, auditability, or workflow need, identify whether the evaluated shape failed while the need remains valid. If it does, continue iterating on the candidate-surface comparison before handoff when feasible, normally with 2-3 plausible shapes unless the decision documents why only one is viable. The follow-up must compare candidates, choose the best, combine useful behaviors if appropriate, defer or kill the track, or record `none viable yet`.
- Before closing any ADR, POC, eval, promotion, or deferred-capability decision epic with outcome `keep-as-reference`, `defer`, `more evidence`, `candidate selected`, `none viable yet`, or another non-promotion result, check existing follow-up work in public docs. If none exists and the work cannot be resolved in the current checkpoint, record the remaining comparison need in the decision or handoff instead of opening a new issue unless the maintainer explicitly asks.
- Do not use taste review to bypass safety or evidence discipline: authority, citations, provenance, freshness, local-first behavior, duplicate handling, runner-only access, approval-before-write, and ADR/POC/eval/promotion decisions still apply.
