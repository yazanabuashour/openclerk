# Maintainer Notes

This repository uses **Beads** (`bd`) in embedded mode for maintainer task
tracking.

This repository is public and includes a production `openclerk` runner binary,
an Agent Skills-compatible OpenClerk skill, and a local SQLite runtime. Keep
maintainer docs honest about the actual supported surface.

Recurring security operations are tracked in
[docs/security-operations.md](security-operations.md). Use that runbook for
dependency review cadence, advisory rehearsal, threat-model refreshes, and
deeper testing expectations.

Agents should start with `openclerk document` or `openclerk retrieval`. There
is no supported public importable Go API, remote HTTP API, or daemon path for
`0.1.0`.

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
- GitHub Releases are created from version tags in the `v0.y.z` form, starting with `v0.1.0`.
- Release publication runs in a protected `release` environment with narrowly scoped write permissions.
- `v*` tags are protected so only maintainers or trusted automation can create them.
- Security reports are expected through GitHub private vulnerability reporting.
- GitHub Actions require pinned action SHAs.

Add the CodeQL status check to required `main` checks after the new CodeQL
workflow has run once and GitHub exposes the exact reported check name.

Review enforcement nuance:

- The repository currently has a single maintainer account.
- For `0.1.0`, `main` requires pull requests, status checks, conversation resolution, and one approving review.
- Code-owner review enforcement and admin enforcement remain off while there is only one maintainer, so the repository does not become unmergeable.
- Tighten code-owner review enforcement, admin bypass, and maintainer isolation once a second maintainer can satisfy the review requirement.

Untrusted pull request policy:

- Pull request workflows must stay fork-safe and use read-only `contents` permission unless a specific trusted workflow boundary justifies more.
- Do not expose release, package, deployment, private infrastructure, or OpenClerk data secrets to code from untrusted forks.
- Avoid `pull_request_target` for workflows that check out or execute contributor-controlled code.
- Dependency review, policy checks, formatting, linting, tests, skill validation, release-doc validation, and CodeQL are acceptable untrusted PR validation surfaces when they run without secrets.

Maintainer and automation isolation:

- Prefer `GITHUB_TOKEN` with explicit job-scoped permissions over personal access tokens or long-lived bot credentials.
- Use a dedicated low-privilege bot identity only when new automation needs privileges that `GITHUB_TOKEN` cannot safely provide.
- Keep release and deployment writes behind the protected `release` environment.
- Enable code-owner review enforcement, stricter admin bypass policy, and stronger review separation only after at least two maintainers can satisfy those controls without blocking routine maintenance.
- Do not use self-hosted runners for untrusted pull requests. Only consider self-hosted runners for trusted branches or tags after documenting isolation, secret exposure, cleanup, and network-access controls.

When changing GitHub settings, keep the repo aligned with:

- [SECURITY.md](../SECURITY.md) for disclosure handling and release integrity expectations.
- [docs/security-operations.md](security-operations.md) for recurring security operations and deeper testing expectations.
- [.github/CODEOWNERS](../.github/CODEOWNERS) for sensitive file ownership.
- [.github/workflows/pull-request.yml](../.github/workflows/pull-request.yml) for fork-safe checks.
- [.github/workflows/release.yml](../.github/workflows/release.yml) for runner, skill, source, checksum, SBOM, and attestation publication.

## Release Publication

The first public release tag should be `v0.1.0`. Tag a version like `v0.1.0`,
push the tag, and let the release workflow:

- validate release notes, changelog, skill package, formatting, linting, and tests before publish
- build binaries with `openclerk --version` set from the tag
- require `docs/release-notes/<tag>.md` and a matching `CHANGELOG.md` entry before publishing
- create or reuse only a draft GitHub Release before assets are attached
- use `docs/release-notes/<tag>.md`, for example `docs/release-notes/v0.1.0.md`, as the GitHub Release body
- keep release-note paragraphs and list items on one source line so GitHub Releases and API clients do not show hard-wrapped prose
- attach platform binary archives, the skill archive, the canonical source archive, release installer, SHA256 checksums, and SBOM
- verify the draft release has the expected asset set before publication
- generate GitHub attestations for the published assets
- publish the draft only after all assets and attestations are ready, then verify the release is latest

The release bundle logic lives in `scripts/build-release-bundle.sh`. The
installer logic lives in `scripts/install.sh`.

The release installer installs the `openclerk` binary only. It prints the skill
source URL and instructs users to install `skills/openclerk` with their agent's
native skill installer or skill directory.

Before tagging, add `docs/release-notes/<tag>.md`, update `CHANGELOG.md`, and
run `./scripts/validate-release-docs.sh <tag>` locally. The release workflow
runs the same check before publishing and does not fall back to generated
GitHub release notes.

After this draft-first workflow is active, enable GitHub release immutability
for future releases when repository settings support it. Published release tags
and assets should then be treated as immutable; fix bad artifacts with a new
patch release instead of replacing assets on an existing release.

## Runner Storage Defaults

The installed runner defaults to:

```text
${XDG_DATA_HOME:-~/.local/share}/openclerk
```

That location contains `openclerk.sqlite` plus the `vault/` tree used for
canonical markdown documents.
