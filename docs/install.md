# OpenClerk Install and Upgrade

Use these commands when an agent needs exact install or upgrade steps from a
shell.

## Install

Install the latest runner:

```bash
tmp_dir="$(mktemp -d)"
curl -fsSLo "$tmp_dir/install.sh" https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh
gh attestation verify "$tmp_dir/install.sh" --repo yazanabuashour/openclerk
OPENCLERK_INSTALL_DIR="$HOME/.local/bin" sh "$tmp_dir/install.sh"
```

Install a pinned release:

```bash
tmp_dir="$(mktemp -d)"
curl -fsSLo "$tmp_dir/install.sh" https://github.com/yazanabuashour/openclerk/releases/download/v0.2.4/install.sh
gh attestation verify "$tmp_dir/install.sh" --repo yazanabuashour/openclerk
OPENCLERK_INSTALL_DIR="$HOME/.local/bin" OPENCLERK_VERSION=v0.2.4 sh "$tmp_dir/install.sh"
```

Register the matching `skills/openclerk/SKILL.md` with the agent's native skill
system.

Verify the install:

```bash
command -v openclerk
openclerk --version
```

A complete install has both the runner and the skill.

## Upgrade

Rerun the installer for the latest or requested version:

```bash
tmp_dir="$(mktemp -d)"
curl -fsSLo "$tmp_dir/install.sh" https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh
gh attestation verify "$tmp_dir/install.sh" --repo yazanabuashour/openclerk
OPENCLERK_INSTALL_DIR="$HOME/.local/bin" sh "$tmp_dir/install.sh"
```

Then re-register the matching `skills/openclerk/SKILL.md` skill and verify:

```bash
command -v openclerk
openclerk --version
```
