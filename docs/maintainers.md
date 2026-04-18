# Maintainer Notes

This repository uses **Beads** (`bd`) in embedded mode for maintainer task tracking.

The production agent surface is the [`cmd/openclerk-agentops`](../cmd/openclerk-agentops) JSON runner backed by [`agentops`](../agentops). The developer product surface is the embedded Go module exposed through the code-first [`client/local`](../client/local) SDK facade. The generated [`client/openclerk`](../client/openclerk) package remains available for raw OpenAPI fallback work. Backend-specific generated clients are not part of the public surface. There is no hosted deployment target, and the default user path does not require a daemon or bound port.

Until the first release tag is published, the install command for consumers is:

```bash
go get github.com/yazanabuashour/openclerk/client/local@main
```

Agents should start with `go run ./cmd/openclerk-agentops document` or `go run ./cmd/openclerk-agentops retrieval`. Go consumers should start with `local.OpenClient(local.Config{})`. They can import [`client/openclerk`](../client/openclerk) from the same module only when generated request and response types are needed. [`cmd/openclerkd`](../cmd/openclerkd) remains an intentional HTTP debug and compatibility surface, not the primary runtime path.

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
- Policy, release, and skill files are part of the public contract and should stay reviewable in Git alone.
- Do not document machine-absolute filesystem paths in committed docs.
- Do not assume private infrastructure, deploy secrets, or internal services exist unless they have been added explicitly.

## Repository administration

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

- [SECURITY.md](../SECURITY.md) for disclosure handling and release integrity expectations.
- [.github/CODEOWNERS](../.github/CODEOWNERS) for sensitive file ownership.
- [.github/workflows/pull-request.yml](../.github/workflows/pull-request.yml) for fork-safe checks.
- [.github/workflows/release.yml](../.github/workflows/release.yml) for source bundle, checksum, SBOM, and attestation publication.

## Release workflow

Before cutting the first public tag (`v0.1.0`) or any later public tag:

```bash
gh workflow run release.yml -f ref=main
```

Tagged releases are the first distributable artifact. A `v0.y.z` tag triggers:

- source archive generation
- SHA-256 checksum generation
- CycloneDX SBOM generation
- Sigstore-backed provenance attestation
- Sigstore-backed SBOM attestation
- GitHub Release publication with the generated assets

The release bundle logic lives in [`scripts/build-release-bundle.sh`](../scripts/build-release-bundle.sh).

Do not add downloadable binaries or package-manager artifacts in this pass. The public release contract remains source-only.

## Runtime storage defaults

The embedded runtime defaults to:

```text
${XDG_DATA_HOME:-~/.local/share}/openclerk
```

That location contains `openclerk.sqlite` plus the `vault/` tree used for canonical markdown documents.
