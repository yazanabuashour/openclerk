# OpenClerk Local Artifact OCR Platform Policy Comparison

## Summary

`oc-hbu5` compared the platform prerequisites left by the local
OCR/scanned-PDF artifact-candidate path: runner-owned model/provider egress,
runner-owned local OCR runtime policy, and no OpenClerk-owned OCR extraction.

The result is explicit kill for OpenClerk-owned OCR extraction on this path.
No product implementation work item is filed.

## Targeted Policy Evidence

| Policy | Safety pass | Capability pass | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| Model/provider egress over local artifacts | fail | partial | best convenience but poor privacy taste | Kill for this path. |
| Local OCR runtime/dependency policy | partial | partial | setup burden | Kill for this path. |
| No OpenClerk-owned OCR extraction | pass | pass for policy, partial for OCR feature | acceptable explicit boundary | Select. |

## Decision

OpenClerk should not own OCR-capable local artifact extraction in the current
artifact-candidate path. The safe supported workflow is:

- external OCR or multimodal reading may happen outside OpenClerk
- the user supplies reviewed text or explicit content to OpenClerk
- OpenClerk uses existing candidate planning, duplicate checks, and approved
  durable write actions

This keeps OCR/model/parser output out of OpenClerk authority until a separate
future product direction deliberately changes the platform boundary.

## Safety, Capability, UX

Safety pass: pass. The selected policy preserves runner-only access,
local-first behavior, no hidden local-file egress, extractor/model non-use,
unsupported-file rejection, current duplicate handling, and
approval-before-write.

Capability pass: pass for the policy decision. It intentionally does not add
OCR recovery; current supported capability remains supplied text, UTF-8 text,
markdown, and text-bearing PDF artifact planning.

UX quality: acceptable as a final boundary. It is less convenient than OCR, but
less surprising and safer for local-first OpenClerk users.

## Outcome

Explicitly kill OpenClerk-owned OCR extraction for this path. No additional
OCR artifact-candidate follow-up is filed.
