# Release Verification

Tagged OpenClerk releases publish:

- `openclerk_<version>_<os>_<arch>.tar.gz`
- `openclerk_<version>_skill.tar.gz`
- `openclerk_<version>_source.tar.gz`
- `openclerk_<version>_checksums.txt`
- `openclerk_<version>_sbom.json`
- `install.sh`

The platform archives contain the `openclerk` runner. The skill archive
contains `skills/openclerk/SKILL.md`. Checksums and GitHub attestations verify
that release assets were produced by this repository's workflow.

## Verify a Release

Download the assets from the GitHub Release page for the tag you want to verify,
then run:

```bash
shasum -a 256 -c openclerk_<version>_checksums.txt
gh attestation verify openclerk_<version>_<os>_<arch>.tar.gz --repo yazanabuashour/openclerk
gh attestation verify openclerk_<version>_skill.tar.gz --repo yazanabuashour/openclerk
gh attestation verify openclerk_<version>_source.tar.gz --repo yazanabuashour/openclerk
gh attestation verify install.sh --repo yazanabuashour/openclerk
```

For the latest release, verify GitHub's latest pointer resolves to the expected
tag:

```bash
gh release view --repo yazanabuashour/openclerk --json tagName --jq .tagName
```

Published release assets are intended to be immutable going forward. If an
artifact is wrong, ship a new patch release instead of mutating the existing
release.

## Smoke-Test an Install

Install into a temporary directory, then verify the runner version and commands:

```bash
install_dir="$(mktemp -d)"
OPENCLERK_INSTALL_DIR="$install_dir" \
  OPENCLERK_VERSION=v0.2.0 \
  sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/download/v0.2.0/install.sh)"

export PATH="$install_dir:$PATH"
command -v openclerk
openclerk --version
openclerk --help
printf '%s\n' '{"action":"resolve_paths"}' | openclerk document
```

The valid runner domains are `document` and `retrieval`. A complete install
also registers the matching `skills/openclerk/SKILL.md` with the user's agent.

## SBOM

The JSON SBOM asset is intended for audit tooling and manual inspection:

```bash
jq '.components | length' openclerk_<version>_sbom.json
```

The SBOM is generated from the tagged source contents during the release
workflow and attached to the same GitHub Release as the binary, skill, and
source archives.
