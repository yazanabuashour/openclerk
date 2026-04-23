# Security Policy

## Supported Versions

This project is pre-`1.0` and ships a production `openclerk` runner plus a
single-file OpenClerk skill. The supported code lines are the current default
branch and the most recent `v0.y.z` tag, if one exists.

Older pre-`1.0` tags are not guaranteed to receive fixes or backports.

## Reporting a Vulnerability

Do not report vulnerabilities in public issues, pull requests, or discussions.

Use GitHub private vulnerability reporting from the repository Security tab.
Include:

- a clear description of the issue
- affected files or workflow surfaces
- reproduction steps or proof-of-concept details
- expected impact and known mitigations

If GitHub private reporting is unavailable, contact the repository owner through
an existing private channel and share only enough detail to arrange private
handoff. Do not disclose the vulnerability publicly while that handoff is being
arranged.

## Response Expectations

These are targets, not contractual guarantees:

| Severity | Initial acknowledgment | Status update target | Patch or mitigation target |
| --- | --- | --- | --- |
| Critical | within 2 business days | within 5 calendar days | within 14 calendar days |
| High | within 3 business days | within 7 calendar days | within 30 calendar days |
| Medium | within 5 business days | within 14 calendar days | next planned release or documented mitigation |
| Low | within 5 business days | as needed | next routine release if accepted |

## Severity Handling

Maintainers triage reports using practical impact on repository users and
maintainers:

- Critical: repository compromise, credential exposure, arbitrary code execution in trusted automation, or release-integrity failure.
- High: meaningful integrity or privilege risk without a full repo compromise.
- Medium: exploitable weakness with limited blast radius or clear prerequisites.
- Low: hard-to-exploit issue, defense-in-depth gap, or low-impact misconfiguration.

## Ongoing Security Operations

Maintainers use [docs/security-operations.md](docs/security-operations.md) for
recurring dependency review, advisory rehearsal, threat-model refreshes, and
deeper testing expectations. The private reporting and response expectations in
this file remain the public source of truth for vulnerability reports.

## Patch and Advisory Process

- Fixes land privately first when needed to avoid widening exposure.
- Public release notes should avoid exploit-enabling detail until a fix or mitigation is available.
- If the repository later adopts GitHub Security Advisories, maintainers should publish advisories for material fixes.

## Emergency Releases and Hotfixes

If a vulnerability affects the latest supported code line, maintainers may cut
an out-of-band patch tag and GitHub Release outside the normal release cadence.

Emergency fixes publish updated runner, skill, and source releases with
checksums, SBOMs, and GitHub attestations. OpenClerk does not publish a hosted
service deployment or remote HTTP API contract.
