# Release Verification

Tagged OpenClerk releases publish:

- `openclerk_<version>_<os>_<arch>.tar.gz`
- `openclerk_<version>_skill.tar.gz`
- `openclerk_<version>_source.tar.gz`
- `openclerk_<version>_checksums.txt`
- `openclerk_<version>_sbom.json`
- `install.sh`

The platform archives contain the production `openclerk` binary. The skill archive contains the shipped `skills/openclerk/SKILL.md`. The source archive is the canonical Go module and local runtime source artifact.

The installer verifies the matching platform archive, installs the same-tag runner, prints `openclerk --version`, and tells users to register the same-tag skill source or archive with their agent. Checksums and GitHub attestations verify that release assets were produced by this repository's workflow.

The release workflow publishes through a draft-first path and verifies the draft asset set before publication, so future GitHub immutable releases can lock tags and assets only after every release asset and attestation is ready.

## Verify a Release

Download the assets from the GitHub Release page for the tag you want to verify, then run:

```bash
shasum -a 256 -c openclerk_<version>_checksums.txt
gh attestation verify openclerk_<version>_<os>_<arch>.tar.gz --repo yazanabuashour/openclerk
gh attestation verify openclerk_<version>_skill.tar.gz --repo yazanabuashour/openclerk
gh attestation verify openclerk_<version>_source.tar.gz --repo yazanabuashour/openclerk
gh attestation verify install.sh --repo yazanabuashour/openclerk
```

If these commands succeed, the assets match the published checksums and have valid GitHub attestations for this repository.

For the latest release, verify GitHub's latest pointer resolves to the expected tag:

```bash
gh release view --repo yazanabuashour/openclerk --json tagName --jq .tagName
```

When repository-level release immutability is enabled, published release tags and assets cannot be replaced after publication. If an artifact is wrong, ship a new patch release instead of mutating the existing release.

## Smoke-Test an Install

Install into a temporary directory, then verify the runner version and commands:

```bash
install_dir="$(mktemp -d)"
OPENCLERK_INSTALL_DIR="$install_dir" \
  OPENCLERK_VERSION=v0.1.0 \
  sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/download/v0.1.0/install.sh)"

export PATH="$install_dir:$PATH"
command -v openclerk
openclerk --version
openclerk --help
printf '%s\n' '{"action":"resolve_paths"}' | openclerk document
```

The valid runner domains are `document` and `retrieval`.

## SBOM

The JSON SBOM asset is intended for audit tooling and manual inspection:

```bash
jq '.components | length' openclerk_<version>_sbom.json
```

The SBOM is generated from the tagged source contents during the release workflow and attached to the same GitHub Release as the binary, skill, and source archives.
