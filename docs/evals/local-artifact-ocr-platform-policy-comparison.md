# Local Artifact OCR Platform Policy Comparison

## Status

Implemented platform-policy comparison for `oc-hbu5`.

This document does not add runner actions, OCR dependencies, model providers,
parser pipelines, storage behavior, public APIs, product behavior, shipped
skill behavior, or implementation authorization.

Governing evidence:

- [`docs/evals/local-ocr-review-extraction-candidates.md`](local-ocr-review-extraction-candidates.md)
- [`docs/evals/results/ockp-local-ocr-review-extraction-candidates.md`](results/ockp-local-ocr-review-extraction-candidates.md)
- [`docs/architecture/local-ocr-review-extraction-candidate-decision.md`](../architecture/local-ocr-review-extraction-candidate-decision.md)

## Policy Candidates

| Candidate | Safety | Capability | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| Runner-owned model/provider egress for local artifacts | Fails as an OCR prerequisite for this product path. It would require a new provider, egress, credential, audit, retention, and private-local-file approval model beyond `artifact_candidate_plan`. | Plausible for OCR quality, but outside the local-first runner contract. | Best simple OCR UX, but surprising privacy posture. | Kill for this path. |
| Runner-owned local OCR runtime policy | Partial. It could preserve local-first behavior, but it creates installation, language data, versioning, page rendering, timeout, platform support, and dependency-update burden. | Plausible for OCR, but not proven and not available in the current local evidence. | Heavy setup for normal users. | Kill for this path. |
| No OpenClerk-owned OCR extraction | Pass. Keeps OpenClerk as a runner-owned markdown/source system, avoids hidden parser/model truth, and preserves existing approval-before-write. | Pass for the supported policy: users can supply text produced elsewhere, and OpenClerk can plan or write only after review/approval. It intentionally does not recover OCR text. | Acceptable when documented as an explicit product boundary. | Select. |

## Selected Policy

Select no OpenClerk-owned OCR extraction for this path.

OpenClerk should not own OCR-capable local artifact extraction until a future
product direction changes the platform boundary. Users may still use external
OCR or multimodal tools outside OpenClerk and provide reviewed text through
existing supported surfaces. OpenClerk then preserves candidate planning,
duplicate checks, authority limits, and approval-before-write on supplied text.

## Safety, Capability, UX

Safety pass: pass. The selected policy preserves runner-only access,
local-first behavior, no hidden model or parser truth, unsupported-file
rejection, duplicate handling for supported text candidates, and
approval-before-write.

Capability pass: pass for the policy decision and partial for OCR as a
feature. OpenClerk deliberately does not recover OCR text, but it safely
handles supplied reviewed text and current text/PDF artifact candidates.

UX quality: acceptable as a final product boundary. Model-assisted OCR remains
the best convenience reference, but the privacy and provenance policy cost is
too high for this OpenClerk path.

## Outcome

Explicitly kill OpenClerk-owned OCR-capable local artifact extraction for this
path. Do not file implementation or further OCR artifact-candidate work items from
this decision.
