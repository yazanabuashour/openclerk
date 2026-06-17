# OpenClerk

The governance layer for LLM-maintained Markdown wikis. One binary. One `SKILL.md`.

Local-first, citation-bearing memory for coding agents that should not reread,
duplicate, or silently rewrite truth.

## Watch OpenClerk catch stale memory in 60 seconds

```bash
curl -fsSL https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh | sh
openclerk demo init
openclerk demo ask "what changed and what is stale?"
```

The demo creates an isolated sample vault, writes a source note and synthesis
page, changes the source, and reports the synthesis as stale with citations and
a `compile_synthesis` repair request.

## What is this?

OpenClerk gives agents a citation-bearing, provenance-tracked knowledge base
over your local markdown vault. Agents read and write through a strict JSON
runner — a stable, citable contract that keeps canonical markdown as the
human-readable authority. Knowledge compounds: useful synthesis becomes durable,
inspectable markdown instead of being rediscovered on every query.

## Who is it for?

Technical users building agent workflows over local markdown, source notes, or
research vaults. If you want your agent to *know* your notes — not just search
them — and you want to audit every write, this is the runtime for that.

## Why not just RAG / Obsidian / NotebookLM?

| | OpenClerk | RAG pipeline | Obsidian | NotebookLM |
|---|---|---|---|---|
| Canonical authority | Markdown in vault | Embedding index | Markdown in vault | Google's servers |
| Agent-native writes | Yes, with provenance | No | No | No |
| Citations on retrieval | Yes, always | Varies | No | Yes |
| Local-first | Yes | Depends | Yes | No |
| Composable modules | Yes | DIY | Plugin ecosystem | No |

Many RAG setups make the index the primary agent interface. OpenClerk keeps
markdown as the human-readable authority and treats indexes and projections as
derived recall layers. Obsidian has no agent-write contract. NotebookLM is not
local.

For a more detailed buyer matrix, see [`docs/comparison.md`](docs/comparison.md).

## Try it in 5 minutes

**Direct install:**

```bash
curl -fsSL https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh | sh
```

Full install options: [`docs/install.md`](docs/install.md)

**Or tell your agent:**

```text
Install OpenClerk into $HOME/.local/bin using https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh or the requested release. Register release-matched skills/openclerk/SKILL.md from installer output. Verify command -v openclerk, openclerk --version, and skill path. Report only after runner and skill verify.
```

**Upgrade prompt:**

```text
Upgrade OpenClerk using https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh or the requested release. Re-register release-matched skills/openclerk/SKILL.md from installer output. Verify command -v openclerk, openclerk --version, and skill path. Report only after runner and skill verify.
```

**Then bind your vault and verify retrieval:**

```bash
openclerk init --vault-root path/to/vault
printf '%s\n' '{"action":"search","search":{"text":"architecture","limit":5}}' \
  | openclerk retrieval
```

Each result carries a `doc_id`, `chunk_id`, and citation path. That's the
contract.

## What to test first

These are the eval-worthy surfaces in priority order:

1. **Lexical search** — search for a topic you know is present and verify five cited results.
2. **Document write → re-search** — create a short note, then search for it and confirm the cited result appears.
3. **Synthesis page lifecycle** — create a `synthesis/` page from source paths, update it, inspect provenance.
4. **Duplicate candidate detection** — ingest a near-duplicate, then confirm `duplicate_candidate_report` surfaces it rather than silently creating a second document.
5. **Relationship graph context** — run `graph_context_report` for a known markdown page and confirm canonical relationship text, links/backlinks, graph freshness, and provenance refs are returned without creating graph truth.
6. **Relationship graph reports** — run `graph_relationship_report` for the same page and confirm relationship paths, direct-vs-derived evidence, typed candidates, and limited graph audit findings stay cited and read-only.
7. **Relationship maintenance plans** — run `graph_relationship_maintenance_plan` for the same page and confirm candidate section content, next approved write requests, duplicate handling, rollback/audit path, and failure modes are returned with `planned_no_write`.
8. **Retrieval replay loop** — explicitly run `retrieval_eval_capture` for a dogfood search, then run `retrieval_eval_replay` after changes and inspect Jaccard, top-1, and latency metrics. Capture is local-only and stores result ids/paths, not raw vault content.
9. **Search diagnostics** — run `search_diagnostics_report` for a query and confirm it recommends default `search` versus explicit `semantic_search`, shows module readiness/cost/latency posture, and reports no default ranking change.
10. **Maintenance report** — run `maintenance_report` for a relevant query/path prefix and confirm layout, projection freshness, relationship context, duplicate risk, module posture, and git lifecycle posture are packaged read-only with no repair.
11. **Stale projection detection** — update a source doc, then confirm downstream synthesis shows as stale before repair.

Report correctness, tool call count, and wall time. That's how the maintainers
gate new features.

## Artifact And File-Type Support

OpenClerk treats artifact content as candidate evidence until a durable write is
approved. The current support matrix is:

| Input | Supported path | Boundary |
|---|---|---|
| Public PDF, HTML, Markdown, or GitHub README/blob URL | `ingest_source_url` | Runner-owned inspect/plan, fetch, provenance, duplicate checks, and approved create/update only |
| Supplied video transcript | `ingest_video_url` | Transcript text and provenance must be supplied; no native media acquisition |
| Pasted or explicit content | `artifact_candidate_plan` | Read-only path/title/body/tags/fields/duplicate proposal before approval |
| Explicit local text, markdown, or text-bearing PDF | `artifact_candidate_plan` with `local_path` | Reads only the supplied file; no durable write |
| Common images or scan-only PDFs | `artifact_candidate_plan` with `text_extraction: "ocr_review"` and verified `tesseract-ocr` module | OCR text is review-required candidate evidence |
| Opaque binaries, slide decks, emails, chats, forms, native media without transcript | Unsupported by default | Paste reviewed text or use a supported runner path |

No parser output, OCR result, local file metadata, or fetched source becomes
canonical until the user approves the existing write action.

## Modules

OpenClerk follows the [building block economy](https://mitchellh.com/writing/building-block-economy)
model deliberately. The mainline runner stays narrow. Optional behavior ships
as separately installed, manifest-verified modules:

### Agent Module Instructions

Install prompt:

```text
Install the OpenClerk module <module-provider> using <module-manifest-path>.
Use <module-command> on PATH, register <module-skill-path>, and verify with `openclerk module` list_modules. Do not pass command_args or edit SQLite directly.
```

Upgrade prompt:

```text
Upgrade the OpenClerk module <module-name> to <module-version-or-latest>.
Refresh registration through `openclerk module`, preserve existing provider config, and verify with list_modules. Do not edit SQLite directly.
```

Available installable modules:

| Module name | Provider | Adds | Manifest | Skill |
|---|---|---|---|---|
| `ollama-embeddings` | `ollama` | Semantic search (local-first) | `modules/ollama-embeddings/module.json` | `modules/ollama-embeddings/skill/ollama-embeddings/SKILL.md` |
| `gemini-embeddings` | `gemini` | Semantic search (cloud opt-in) | `modules/gemini-embeddings/module.json` | `modules/gemini-embeddings/skill/gemini-embeddings/SKILL.md` |
| `tesseract-ocr` | `tesseract` | OCR review for images and scan-only PDFs | `modules/tesseract-ocr/module.json` | `modules/tesseract-ocr/skill/tesseract-ocr/SKILL.md` |

Core lexical search and citation behavior require no modules. Semantic search
is explicit opt-in — it accelerates recall, it does not become the authority
layer. No hidden provider fallback. No committed embedding cache. Full module
install guidance lives in `modules/docs/install.md`.

Inspect the current block inventory:

```bash
openclerk capabilities
```

## What is explicitly not supported yet

- **Autonomous browsing / recursive URL crawling** — `ingest_source_url` can inspect, plan, or ingest supplied public URLs through the runner, but it does not browse the open web autonomously or recursively crawl discovered links.
- **Automatic video transcript acquisition** — not supported.
- **Hosted service or cloud sync** — fully local. No OpenClerk server, no SaaS.
- **Multi-user / team server** — single-user, single-machine runtime.
- **Broad vector DB memory** — no Pinecone, Weaviate, or default durable vector index. Semantic modules are optional and local-only by default.
- **Autonomous memory and routing** — deferred until docs, synthesis, and truth-sync layers are reliable. See [`docs/architecture/memory-routing-reference-decision.md`](docs/architecture/memory-routing-reference-decision.md).

## Runner

```bash
openclerk config      # persisted product/profile config
openclerk document    # doc writes, registry, paths
openclerk retrieval   # search, graph context/reports, records, provenance
openclerk clerk       # optional read-only Chronicler orchestration
openclerk capabilities
```

JSON in, JSON out. See runner help:

```bash
openclerk config --help
openclerk document --help
openclerk retrieval --help
openclerk clerk --help
```

`openclerk config inspect_config` is the read-only effective config summary for
storage, profile defaults, module summaries, and git lifecycle gate posture.
`openclerk config` owns persisted product/profile preferences such as the
default autonomy profile. `openclerk module` owns optional provider writes.
Vault-root binding is storage bootstrap config: initialize or intentionally
rebind it with `openclerk init --vault-root`, not `openclerk config`.
Request-level `document` and `retrieval` `autonomy` fields override persisted
profile defaults field-by-field. Git checkpoint enablement remains
invocation-scoped through `--git-checkpoints` or `OPENCLERK_GIT_CHECKPOINTS`.
`openclerk clerk run --once` is the combined Chronicler MVP report. The same
read-only primitives are also exposed as `openclerk clerk inbox_scan` for
explicit local inbox candidate planning and `openclerk clerk context_pack` for
task context, must-read documents, decisions, and citations. All three emit
`openclerk-clerk.v1`, report `planned_no_write: true`, and perform no durable
vault writes. Planning that inspects Core evidence requires existing
OpenClerk storage; Chronicler returns a blocker rather than initializing
SQLite from a read-only command.

Storage: `${XDG_DATA_HOME:-~/.local/share}/openclerk/openclerk.sqlite`  
Override: `OPENCLERK_DATABASE_PATH` or `--db`

## Architecture

- [Agent knowledge plane →](docs/architecture/agent-knowledge-plane.md)
- [Chronicler boundary →](docs/architecture/chronicler-boundary.md)
- [Consumer infrastructure and purchase ledger roadmap →](docs/architecture/consumer-infrastructure-and-purchase-ledger-roadmap.md)

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md), [`SECURITY.md`](SECURITY.md), and [`docs/maintainers.md`](docs/maintainers.md).
