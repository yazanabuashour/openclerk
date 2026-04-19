# Maintainer Notes

This repository uses **Beads** (`bd`) in embedded mode for maintainer task
tracking.

This repository is public and includes a production `openclerk` runner binary,
an Agent Skills-compatible OpenClerk skill, and a local SQLite runtime. Keep
maintainer docs honest about the actual supported surface.

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

Current readiness assumptions:

- `main` is the protected default branch.
- Pull requests run only untrusted-safe validation with read-only token scope.
- GitHub Releases are created from version tags in the `v0.y.z` form, starting with `v0.1.0`.
- The `release` environment is protected before enabling public tagged releases.
- `v*` tags are protected so only maintainers or trusted automation can create them.
- Security reports are expected through GitHub private vulnerability reporting.

Current review enforcement nuance:

- The repository currently has a single maintainer account.
- `main` requires pull requests, status checks, conversation resolution, and one approving review, but code-owner review enforcement and admin enforcement remain off so the repository does not become unmergeable.
- Tighten code-owner review enforcement, admin bypass, and maintainer isolation once a second maintainer can satisfy the review requirement.

When changing GitHub settings, keep the repo aligned with:

- `SECURITY.md` for disclosure handling and release integrity expectations.
- `.github/CODEOWNERS` for sensitive file ownership.
- `.github/workflows/pull-request.yml` for fork-safe checks.
- `.github/workflows/release.yml` for runner, skill, source, checksum, SBOM, and attestation publication.

## Release Publication

The first public release tag should be `v0.1.0`. Tag a version like `v0.1.0`,
push the tag, and let the release workflow:

- validate tests before publish
- create or reuse the GitHub Release
- attach platform binary archives, the skill archive, the canonical source archive, release installer, SHA256 checksums, and SBOM
- generate GitHub attestations for the published assets

The release bundle logic lives in `scripts/build-release-bundle.sh`. The
installer logic lives in `scripts/install.sh`.

The release installer installs the `openclerk` binary only. It prints the skill
source URL and instructs users to install `skills/openclerk` with their agent's
native skill installer or skill directory.

## Runner Storage Defaults

The installed runner defaults to:

```text
${XDG_DATA_HOME:-~/.local/share}/openclerk
```

That location contains `openclerk.sqlite` plus the `vault/` tree used for
canonical markdown documents.
