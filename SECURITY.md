# Security Policy

## Supported versions

This project is pre-`1.0` and does not yet ship a runnable application or package. Until `0.1.0`, the only supported code line is the current default branch and the most recent `0.y.z` tag, if one exists.

Older pre-`1.0` tags are not guaranteed to receive fixes or backports.

## Reporting a vulnerability

Do not report vulnerabilities in public issues, pull requests, or discussions.

Use GitHub private vulnerability reporting from the repository Security tab. Include:

- a clear description of the issue
- affected files or workflow surfaces
- reproduction steps or proof-of-concept details
- expected impact and any known mitigations

If GitHub private reporting is temporarily unavailable, contact the repository owner through an existing private channel and share only enough detail to establish a private handoff. Do not disclose the vulnerability publicly while that handoff is being arranged.

## Response expectations

These are targets, not contractual guarantees:

| Severity | Initial acknowledgment | Status update target | Patch or mitigation target |
| --- | --- | --- | --- |
| Critical | within 2 business days | within 5 calendar days | within 14 calendar days |
| High | within 3 business days | within 7 calendar days | within 30 calendar days |
| Medium | within 5 business days | within 14 calendar days | next planned release or documented mitigation |
| Low | within 5 business days | as needed | next routine release if accepted |

## Severity handling

Maintainers will triage reports using practical impact on repository users and maintainers:

- Critical: repository compromise, credential exposure, arbitrary code execution in trusted automation, or release-integrity failure.
- High: meaningful integrity or privilege risk without a full repo compromise.
- Medium: exploitable weakness with limited blast radius or clear prerequisites.
- Low: hard-to-exploit issue, defense-in-depth gap, or low-impact misconfiguration.

## Patch and advisory process

- Fixes land privately first when needed to avoid widening exposure.
- Public release notes should avoid exploit-enabling detail until a fix or mitigation is available.
- If the repository later adopts GitHub Security Advisories, maintainers should publish advisories for material fixes.

## Emergency releases and hotfixes

If a vulnerability affects the latest supported code line, maintainers may cut an out-of-band patch tag and GitHub Release outside the normal release cadence.

Because the repository does not yet ship build artifacts or packages, emergency fixes currently mean source changes and release notes only.
