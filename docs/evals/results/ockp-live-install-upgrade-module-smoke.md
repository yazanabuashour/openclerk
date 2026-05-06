# OpenClerk Live Install/Upgrade/Module Smoke

This report records a local temp HOME/CODEX_HOME smoke validation for install, upgrade, skill registration, and module agent install commands.

## Result

- Version: `v0.0.0-smoke`
- Install: `true`
- Upgrade: `true`
- Skill: `true`
- Module install/config/upgrade/list/remove: `true`

## Evidence

- Installer invocation: `sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh)"`
- Binary path: `$HOME/.local/bin/openclerk`
- Command path: `$HOME/.local/bin/openclerk`
- Version output: `openclerk v0.0.0-smoke`
- Skill path: `$CODEX_HOME/skills/openclerk/SKILL.md`
- Module manifest: `modules/ollama-embeddings/module.json`
- Module skill: `modules/ollama-embeddings/skill/ollama-embeddings/SKILL.md`
- Module provider config: `embedding_model=nomic-embed-text`, `ollama_url=http://localhost:11434`
- Module upgrade preserved config: `true`
- Module verification/redaction: `verified` / `redacted`

## Boundaries

local temp HOME/CODEX_HOME only; release transport is a deterministic local fixture for the GitHub release URLs; no durable host install, no network release fetch, no external provider call, no direct SQLite edit, and no source-built runner invocation after install.
