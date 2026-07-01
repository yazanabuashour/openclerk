# OpenClerk

The governance layer for LLM-maintained Markdown wikis. One binary. One
`SKILL.md`.

OpenClerk gives coding agents a local, citation-bearing knowledge runtime over
Markdown. Canonical Markdown stays the authority; generated indexes and
optional modules stay derived.

## Quick Start

```bash
tmp_dir="$(mktemp -d)"
curl -fsSLo "$tmp_dir/install.sh" https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh
gh attestation verify "$tmp_dir/install.sh" --repo yazanabuashour/openclerk
OPENCLERK_INSTALL_DIR="$HOME/.local/bin" sh "$tmp_dir/install.sh"

openclerk demo init
openclerk demo ask "what changed and what is stale?"
```

The demo creates an isolated sample vault, changes a source note, and reports
stale synthesis with citations and a repair request.

## Use It With a Coding Agent

```bash
demo_dir="$(mktemp -d)/openclerk-demo"
openclerk demo init --root "$demo_dir" --template codebase-decisions
db="$demo_dir/openclerk.sqlite"

openclerk inspect --db "$db"
openclerk clerk context_pack --db "$db" \
  --task "change the auth callback behavior" --limit 5
openclerk clerk session_record_report --db "$db" \
  --inbox-path examples/knowledge-packs/agent-session-to-docs/handoffs/session.md \
  --task "summarize completed auth callback work into repo knowledge" \
  --limit 5
```

The loop is inspect posture, get cited task context, do the work outside
OpenClerk, then turn an explicit session note into reviewable candidate
knowledge. Durable writes still require approved `openclerk document` lifecycle
requests.

## Core Model

- Local-first runner, local vault, local storage.
- JSON in / JSON out command surfaces for agents.
- Citation-bearing retrieval and provenance-aware document lifecycle APIs.
- Approval before durable writes.
- Optional modules can accelerate recall or extraction, but they do not become
  truth.
- No daemon, background repair loop, hosted service, cloud sync, remote API, or
  multi-user server contract.

## Install and Upgrade

Full install options live in [`docs/install.md`](docs/install.md).

**Or tell your agent:**

```text
Install OpenClerk into $HOME/.local/bin from https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh only after gh attestation verify passes. Register release-matched skills/openclerk/SKILL.md. Verify command -v openclerk, openclerk --version, skill path, runner and skill.
```

**Upgrade prompt:**

```text
Upgrade OpenClerk from https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh only after gh attestation verify passes. Re-register release-matched skills/openclerk/SKILL.md. Verify command -v openclerk, openclerk --version, skill path, runner and skill.
```

Bind a real vault after install:

```bash
openclerk init --vault-root path/to/vault
openclerk inspect
```

## Modules

### Agent Module Instructions

Install prompt:

```text
Install the OpenClerk module <module-provider> using <module-manifest-path>. Use the resolved <module-command> path, register <module-skill-path>, and verify with `openclerk module` list_modules. Do not pass command_args or edit SQLite directly.
```

Upgrade prompt:

```text
Upgrade the OpenClerk module <module-name> to <module-version-or-latest>. Refresh registration through `openclerk module`, preserve existing provider config, and verify with list_modules. Do not edit SQLite directly.
```

Available installable modules:

| Module name | Adds | Manifest | Skill |
|---|---|---|---|
| `ollama-embeddings` | Semantic search (local-first) | `modules/ollama-embeddings/module.json` | `modules/ollama-embeddings/skill/ollama-embeddings/SKILL.md` |
| `gemini-embeddings` | Semantic search (cloud opt-in) | `modules/gemini-embeddings/module.json` | `modules/gemini-embeddings/skill/gemini-embeddings/SKILL.md` |
| `tesseract-ocr` | OCR review for images and scan-only PDFs | `modules/tesseract-ocr/module.json` | `modules/tesseract-ocr/skill/tesseract-ocr/SKILL.md` |

Core lexical search and citation behavior require no modules. Full module
install guidance lives in [`modules/docs/install.md`](modules/docs/install.md).

## Runner Surface

```bash
openclerk inspect      # read-only posture before agent work
openclerk config       # persisted product/profile config
openclerk document     # approved document lifecycle APIs
openclerk retrieval    # search, graph context/reports, records, provenance
openclerk clerk        # read-only context and session-record planning
openclerk module       # optional module registration
openclerk capabilities # compact capability inventory
```

Start with `openclerk inspect`, then use `--help` on the relevant surface for
compact request shapes. Do not treat read-only planning as write approval.

## Learn More

- [Agent Contract](docs/agent-contract.md)
- [Knowledge Packs](docs/knowledge-packs.md)
- [Comparison](docs/comparison.md)
- [Architecture](docs/architecture/agent-knowledge-plane.md)
- [Chronicler Lite Boundary](docs/architecture/chronicler-boundary.md)
- [Artifact and intake decisions](docs/architecture/generalized-artifact-ingestion-adr.md)
- [Release verification](docs/release-verification.md)

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md), [`SECURITY.md`](SECURITY.md), and
[`docs/maintainers.md`](docs/maintainers.md).
