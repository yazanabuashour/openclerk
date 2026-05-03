# Maintainer Notes

This repository uses **Beads** (`bd`) in embedded mode for maintainer task
tracking. Recurring security operations are tracked in
[docs/security-operations.md](security-operations.md).

Keep public docs honest about the supported surface: the installed `openclerk`
runner plus `skills/openclerk/SKILL.md`.

## Initial Setup

Preferred tool install:

```bash
mise install
```

Alternative:

```bash
brew install beads dolt
```

## Clone Bootstrap

For a fresh maintainer clone or a second machine:

```bash
git clone git@github.com:yazanabuashour/openclerk.git
cd openclerk
bd bootstrap
bd hooks install
```

If role detection warns in a maintainer clone, set:

```bash
git config beads.role maintainer
```

## Sync Between Machines

Push local Beads state before switching machines, then pull on the other
machine:

```bash
bd dolt push
bd dolt pull
```

If `bd dolt pull` reports uncommitted Dolt changes, commit them first and retry:

```bash
bd dolt commit
bd dolt pull
```

## Public Repo Expectations

- Outside contributors must be able to contribute without Beads.
- Policy, release, and skill files are part of the public contract and should stay reviewable in Git alone.
- Do not document machine-absolute filesystem paths in committed docs.
- Do not assume private infrastructure, deploy secrets, or internal services exist unless they have been added explicitly.
- Do not document a generic skill install location. Agent-specific paths may appear only as clearly labeled examples.

## Repository Administration

The `0.1.0` repository administration target is:

- `main` is the protected default branch.
- Pull requests run only untrusted-safe validation with read-only token scope.
- `main` requires pull requests, status checks, conversation resolution, and one approving review.
- Code-owner review enforcement and admin enforcement remain off while there is only one maintainer.
- GitHub Actions require pinned action SHAs.
- GitHub Releases are created from `v0.y.z` tags.
- Release publication runs in a protected `release` environment.
- `v*` tags are protected against deletion and non-fast-forward updates.
- Published GitHub Releases are immutable.
- Security reports use GitHub private vulnerability reporting.

Tighten code-owner review enforcement, admin bypass, and review separation once
a second maintainer can satisfy those controls.

Untrusted pull request policy:

- Keep pull request workflows fork-safe and read-only unless a trusted workflow boundary justifies more.
- Do not expose release, package, deployment, private infrastructure, or OpenClerk data secrets to code from untrusted forks.
- Avoid `pull_request_target` for workflows that check out or execute contributor-controlled code.
- Dependency review, policy checks, formatting, linting, tests, skill validation, release-doc validation, and CodeQL are acceptable untrusted PR validation surfaces when they run without secrets.

Maintainer and automation isolation:

- Prefer `GITHUB_TOKEN` with explicit job-scoped permissions over personal access tokens or long-lived bot credentials.
- Use a dedicated low-privilege bot identity only when new automation needs privileges that `GITHUB_TOKEN` cannot safely provide.
- Keep release and deployment writes behind the protected `release` environment.
- Do not use self-hosted runners for untrusted pull requests.

Keep GitHub settings aligned with `SECURITY.md`,
`docs/security-operations.md`, `.github/CODEOWNERS`, and the workflows under
`.github/workflows/`.

## Release Publication

Before tagging, add `docs/release-notes/<tag>.md`, update `CHANGELOG.md`, and
run:

```bash
mise exec -- ./scripts/validate-release-docs.sh <tag>
mise exec -- ./scripts/validate-agent-skill.sh skills/openclerk
mise exec -- ./scripts/validate-committed-artifacts.sh
test -z "$(gofmt -l $(git ls-files '*.go'))"
mise exec -- golangci-lint run
mise exec -- go test ./...
mise exec -- go run ./scripts/agent-eval/ockp run --report-name ockp-agentops-production
mise exec -- go run ./scripts/agent-eval/ockp run --parallel 1 --scenario repo-docs-agentops-retrieval,repo-docs-synthesis-maintenance,repo-docs-decision-records,repo-docs-release-readiness,repo-docs-tag-filter,repo-docs-memory-router-recall-report,repo-docs-release-synthesis-freshness --report-name ockp-repo-docs-dogfood
```

The repo-docs dogfood run is mandatory pre-release evidence. It imports only
committed public markdown into an isolated OpenClerk eval vault, exercises the
installed `openclerk document` and `openclerk retrieval` JSON surfaces, and
must pass before tagging. Keep this targeted lane separate from the full
release-blocking AgentOps production gate, but treat failures as tag blockers
until repaired or explicitly reclassified as fixture/reporting defects.

Tag a version like `v0.1.0`, push the tag, and let the release workflow:

- validate release notes, changelog, skill package, formatting, linting, and tests
- build binaries with `openclerk --version` set from the tag
- create or reuse only a draft GitHub Release before assets are attached
- use `docs/release-notes/<tag>.md` as the GitHub Release body
- attach binary archives, skill archive, source archive, installer, checksums, and SBOM
- verify the expected asset set, generate attestations, publish the draft, and verify latest

The release bundle logic lives in `scripts/build-release-bundle.sh`. The
installer logic lives in `scripts/install.sh`. GitHub release immutability is
enabled; fix bad artifacts with a new patch release instead of replacing
published assets.
