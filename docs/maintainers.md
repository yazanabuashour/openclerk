# Maintainer Notes

This repository uses **Beads** (`bd`) in embedded mode for maintainer task tracking.

This repository is public, but it is still only a scaffold. There is no deploy pipeline, package publication target, or runnable application yet. Keep maintainer docs honest about that status.

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

Push local Beads state before switching machines, then pull on the other machine:

```bash
bd dolt push
bd dolt pull
```

If `bd dolt pull` reports uncommitted Dolt changes, commit them first and retry:

```bash
bd dolt commit
bd dolt pull
```

## Public repo expectations

- Outside contributors must be able to contribute without Beads.
- Policy and workflow files are part of the public contract and should stay reviewable in Git alone.
- Do not document machine-absolute filesystem paths in committed docs.
- Do not assume private infrastructure, deploy secrets, or internal services exist unless they have been added explicitly.

## Repository administration

Current readiness assumptions:

- `main` is the protected default branch.
- Pull requests run only untrusted-safe validation with read-only token scope.
- GitHub Releases are created from version tags in the `v0.y.z` form.
- Security reports are expected through GitHub private vulnerability reporting.

Current review enforcement nuance:

- The repository currently has a single maintainer account.
- `main` requires pull requests, status checks, conversation resolution, and one approving review, but code-owner review enforcement and admin enforcement remain off so the repository does not become unmergeable.
- Tighten code-owner review enforcement, admin bypass, and maintainer isolation once a second maintainer can satisfy the review requirement.

When changing GitHub settings, keep the repo aligned with:

- [SECURITY.md](../SECURITY.md) for disclosure handling and patch timing.
- [.github/CODEOWNERS](../.github/CODEOWNERS) for sensitive file ownership.
- [.github/workflows/pull-request.yml](../.github/workflows/pull-request.yml) for fork-safe checks.
- [.github/workflows/release.yml](../.github/workflows/release.yml) for release-note generation.

## Release notes

The current release contract is GitHub Releases only. Tag a version like `v0.1.0`, push the tag, and let the release workflow generate notes from the tag. Do not attach build artifacts, checksums, provenance, or SBOMs until the project actually ships distributable outputs and the stronger release process is implemented.
