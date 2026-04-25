# Security Operations

This runbook defines recurring security work for OpenClerk maintainers. It complements the public reporting and response policy in [SECURITY.md](../SECURITY.md); do not put private vulnerability details in public issues, pull requests, release notes, or this document.

## Cadence

- Weekly: triage Dependabot pull requests, dependency-review failures, and new vulnerability alerts for Go modules and GitHub Actions.
- Monthly: review the GitHub Security tab, private vulnerability reporting state, Dependabot alert backlog, code-scanning alerts, and any deferred security issues in Beads.
- Quarterly: rehearse the advisory workflow, refresh the threat model, and confirm that release, automation, and maintainer-isolation assumptions still match the repository.
- Release-bound: review security impact before tagging any release that changes `.github/workflows/release.yml`, `scripts/install.sh`, `skills/openclerk/SKILL.md`, local storage behavior, runner write/update behavior, or release verification docs.

## High-Risk Surfaces

- Local SQLite and vault data rooted at the configured OpenClerk data paths.
- Runner JSON operations in `cmd/openclerk`, `internal/runner`, `internal/runclient`, and `internal/infra/sqlite`, especially document creation, append, replace-section, projection freshness, and retrieval behavior.
- Agent-facing task policy in `skills/openclerk/SKILL.md`, including direct-reject rules, runner-only routine workflows, synthesis freshness, provenance inspection, and lower-level bypass handling.
- Install and release pipeline files: `scripts/install.sh`, `scripts/build-release-bundle.sh`, `.github/workflows/release.yml`, `docs/release-verification.md`, `CHANGELOG.md`, and `docs/release-notes`.
- GitHub Actions and repository policy files under `.github`, including token permissions, environment protection, CODEOWNERS, dependency review, CodeQL, and branch protection assumptions.
- Contributor pull request paths, especially any workflow that runs code from untrusted forks or exposes repository secrets.

## Review Workflow

1. Open or update a Beads issue for any recurring security review that finds follow-up work.
2. Classify findings using the severity expectations in `SECURITY.md`.
3. Keep exploit details private until a fix or mitigation is available.
4. For dependency updates, prefer the smallest reviewable update that clears the alert and keeps `mise exec -- go test ./...` passing.
5. For workflow or release-pipeline changes, verify token permissions remain job-scoped and no release, deployment, package, or repository write permission is granted to untrusted pull request execution.
6. For skill or runner policy changes, confirm the public docs, skill contract, tests, release notes, and current production eval evidence remain aligned.

## Deeper Testing Expectations

- Runner document create, append, replace-section, layout inspection, retrieval, provenance, projection, and promoted-record behavior should have focused validation and idempotency tests before release.
- Storage and projection changes should include tests for persisted data, source references, freshness, and invalid-input rejection.
- Skill policy changes should run `./scripts/validate-agent-skill.sh skills/openclerk` and the relevant OpenClerk production agent eval gate before release.
- Release-pipeline changes should run `./scripts/validate-release-docs.sh <tag>` and `./scripts/validate-committed-artifacts.sh` for the target tag and verify the expected release asset, checksum, SBOM, version, installer, and attestation behavior.
- Add fuzzing or property-style tests when parsing, normalization, source-reference handling, graph/projection derivation, or import logic becomes complex enough that table tests no longer cover realistic malformed input.
- Abuse-case tests should be added before introducing remote APIs, hosted services, secrets-backed integrations, self-hosted runners, broad automation write privileges, or new routine agent transports.

## Advisory Rehearsal

At least quarterly, maintainers should rehearse the private advisory flow without publishing a real advisory:

- Confirm GitHub private vulnerability reporting is enabled and reachable from the repository Security tab.
- Confirm the private fix path, release notes redaction approach, patch-tag process, and release verification steps are still documented.
- Confirm emergency release expectations still match the current artifact set: binary archives, skill archive, installer, checksums, SBOM, and attestations.
- File Beads issues for any gap found during the rehearsal.
