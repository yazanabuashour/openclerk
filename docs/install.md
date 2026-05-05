# OpenClerk Install and Upgrade

Use these commands when an agent needs exact install or upgrade steps from a
shell.

## Install

Install the latest runner:

```bash
OPENCLERK_INSTALL_DIR="$HOME/.local/bin" sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh)"
```

Install a pinned release:

```bash
OPENCLERK_INSTALL_DIR="$HOME/.local/bin" OPENCLERK_VERSION=v0.2.3 sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/download/v0.2.3/install.sh)"
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
OPENCLERK_INSTALL_DIR="$HOME/.local/bin" sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh)"
```

Then re-register the matching `skills/openclerk/SKILL.md` skill and verify:

```bash
command -v openclerk
openclerk --version
```
