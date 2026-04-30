# Next Eval Candidate Pipeline

## Status

Roadmap for 14 candidate eval-to-decision epics:

- path/title capture: `oc-hhap`, `oc-xh72`, `oc-yjuz`, `oc-xtbl`, `oc-3zd9`
- URL/artifact intake: `oc-0cme`, `oc-wqlb`, `oc-69h3`, `oc-ijdk`
- high-touch workflows: `oc-7feg`, `oc-k8ba`, `oc-oowv`, `oc-nu12`,
  `oc-qnwd`

This note records Beads planning only. It does not authorize runner behavior,
eval harness scenarios, eval execution, promotion decisions, schemas, storage
changes, public APIs, skill behavior changes, product behavior changes, or
implementation work.

## Path And Title Capture Epics

| Epic | Candidate | Source design | Lane |
| --- | --- | --- | --- |
| `oc-hhap` | Low-risk capture ceremony | [`docs/evals/path-title-capture-ceremony-eval-design.md`](../evals/path-title-capture-ceremony-eval-design.md) | `capture-low-risk-ceremony` |
| `oc-xh72` | Explicit overrides in smoother capture | [`docs/evals/path-title-capture-ceremony-eval-design.md`](../evals/path-title-capture-ceremony-eval-design.md) | `capture-explicit-overrides` |
| `oc-yjuz` | Duplicate candidate update versus new path | [`docs/evals/path-title-capture-ceremony-eval-design.md`](../evals/path-title-capture-ceremony-eval-design.md) | `capture-duplicate-candidate-update` |
| `oc-xtbl` | Save-this-note capture ceremony | [`docs/evals/path-title-capture-ceremony-eval-design.md`](../evals/path-title-capture-ceremony-eval-design.md) | `capture-save-this-note-candidate` |
| `oc-3zd9` | Document-these-links placement | [`docs/evals/path-title-capture-ceremony-eval-design.md`](../evals/path-title-capture-ceremony-eval-design.md) | `capture-document-these-links-placement` |

## URL And Artifact Intake Epics

| Epic | Candidate | Source design | Lane |
| --- | --- | --- | --- |
| `oc-0cme` | Unsupported artifact kind intake | [`docs/evals/url-artifact-intake-future-eval-design.md`](../evals/url-artifact-intake-future-eval-design.md) | `artifact-unsupported-kind-intake` |
| `oc-wqlb` | Richer public product-page intake | [`docs/evals/url-artifact-intake-future-eval-design.md`](../evals/url-artifact-intake-future-eval-design.md) | `web-product-page-rich-public-intake` |
| `oc-69h3` | Native media transcript acquisition | [`docs/evals/url-artifact-intake-future-eval-design.md`](../evals/url-artifact-intake-future-eval-design.md) | `artifact-native-media-transcript-acquisition` |
| `oc-ijdk` | Local file artifact intake ladder | [`docs/evals/url-artifact-intake-future-eval-design.md`](../evals/url-artifact-intake-future-eval-design.md) | `artifact-local-file-intake-ladder` |

## High-Touch Workflow Epics

| Epic | Candidate | Source design | Lane |
| --- | --- | --- | --- |
| `oc-7feg` | Compile synthesis ceremony | [`docs/evals/high-touch-successful-workflows-ceremony-eval-design.md`](../evals/high-touch-successful-workflows-ceremony-eval-design.md) | `high-touch-compile-synthesis-ceremony` |
| `oc-k8ba` | Document lifecycle ceremony | [`docs/evals/high-touch-successful-workflows-ceremony-eval-design.md`](../evals/high-touch-successful-workflows-ceremony-eval-design.md) | `high-touch-document-lifecycle-ceremony` |
| `oc-oowv` | Relationship and record lookup ceremony | [`docs/evals/high-touch-successful-workflows-ceremony-eval-design.md`](../evals/high-touch-successful-workflows-ceremony-eval-design.md) | `high-touch-relationship-record-ceremony` |
| `oc-nu12` | Memory router recall ceremony | [`docs/evals/high-touch-successful-workflows-ceremony-eval-design.md`](../evals/high-touch-successful-workflows-ceremony-eval-design.md) | `high-touch-memory-router-recall-ceremony` |
| `oc-qnwd` | Web URL stale repair ceremony | [`docs/evals/high-touch-successful-workflows-ceremony-eval-design.md`](../evals/high-touch-successful-workflows-ceremony-eval-design.md) | `high-touch-web-url-stale-repair-ceremony` |

Each epic follows the same sequence:

1. Add executable eval harness coverage for the candidate lane.
2. Run the targeted eval and publish a reduced report under `docs/evals/results/`.
3. Write a promote, defer, kill, or reference decision.
4. File an implementation bead only if the decision promotes a surface.

The final step is conditional tracking, not implementation. If the decision
does not promote, the conditional implementation-bead task should close with a
no-op reason and no implementation bead should be created.

## Evidence Requirements

Targeted eval reports and decisions should record these conclusions separately:

- safety pass: authority, citations or source refs, provenance, freshness,
  local-first behavior, duplicate handling, runner-only access, and
  approval-before-write were preserved
- capability pass: current `openclerk document` and `openclerk retrieval`
  primitives can or cannot express the workflow
- UX quality: the workflow is acceptable for routine use, or remains taste
  debt because it is too ceremonial, slow, brittle, high-step, retry-prone, or
  guidance-dependent

Promotion can be justified by either a capability gap where the current
primitives cannot safely express the workflow, or by a serious ergonomics and
taste gap where the primitives technically pass but preserve a surface that a
normal user would reasonably find too ceremonial, slow, brittle, high-step,
retry-prone, guidance-dependent, or surprising. Safety remains the hard gate for
both paths: do not promote if authority, citations or source refs, provenance,
freshness, local-first behavior, duplicate handling, runner-only access, or
approval-before-write are weakened.

Reports should also include tool or command count, assistant calls, wall time,
prompt specificity, retries, latency, brittleness, guidance dependence, and
safety risks.

## Non-Authorization Boundary

These epics are the next evidence pipeline, not product approval. They do not
authorize direct create, autonomous autofiling, source or synthesis placement,
special product-page handling, browser automation, parser pipelines, schema
changes, storage changes, public APIs, skill behavior changes, or release
gates.

Any future implementation still requires targeted eval evidence and an
explicit promotion decision naming the exact public surface, request and
response shape, compatibility expectations, failure modes, and safety gates.
