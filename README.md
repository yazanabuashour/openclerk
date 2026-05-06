# OpenClerk

**A local-first knowledge-plane runtime for agents.**  
One binary. One `SKILL.md`. Canonical markdown stays yours.

---

## What is this?

OpenClerk gives agents a citation-bearing, provenance-tracked knowledge base
over your local markdown vault. Agents read and write through a strict JSON
runner — not by grepping files or calling an LLM to remember things. Knowledge
compounds: useful synthesis becomes durable, inspectable markdown instead of
being rediscovered on every query.

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

RAG pipelines make the index the truth. OpenClerk keeps markdown the truth and
makes the index a recall layer. Obsidian has no agent-write contract.
NotebookLM is not local.

## Try it in 5 minutes

Tell your agent:

```text
Install OpenClerk into $HOME/.local/bin from the latest release.
Register skills/openclerk/SKILL.md with your native skill system.
Verify command -v openclerk, openclerk --version, and the installed skill path.
Do not report OpenClerk installed until both the runner and skill are installed.
```

Then bind your vault and run a search:

```bash
openclerk init --vault-root ~/notes
printf '%s\n' '{"action":"search","search":{"text":"architecture","limit":5}}' \
  | openclerk retrieval
```

Each result carries a `doc_id`, `chunk_id`, and citation path. That's the
contract.

Full install details: [`docs/install.md`](docs/install.md)

## What to test first

These are the eval-worthy surfaces in priority order:

1. **Lexical search** — does retrieval return correct citations, not hallucinated paths?
2. **Document write → re-search** — does a new note surface in subsequent queries?
3. **Synthesis page lifecycle** — create a `synthesis/` page, mark sources, update it, inspect provenance.
4. **Duplicate candidate detection** — ingest a near-duplicate and watch `duplicate_candidate_report` surface it rather than silently creating a second document.
5. **Stale projection detection** — update a source doc, confirm downstream synthesis shows as stale before repair.

Report correctness, tool call count, and wall time. That's how the maintainers
gate new features.

## Building blocks, not a platform

OpenClerk follows the [building block economy](https://mitchellh.com/writing/building-block-economy)
model deliberately. The mainline runner stays narrow. Optional behavior ships
as separately installed, manifest-verified modules:

| Module | Provider | Adds |
|---|---|---|
| `ollama-embeddings` | Local Ollama | Semantic search (local-first) |
| `gemini-embeddings` | Gemini API | Semantic search (cloud opt-in) |
| `tesseract-ocr` | Tesseract | OCR review for images and scan-only PDFs |

Core lexical search and citation behavior require no modules. Semantic search
is explicit opt-in — it accelerates recall, it does not become the authority
layer. No hidden provider fallback. No committed embedding cache.

Inspect the current block inventory:

```bash
openclerk capabilities
```

Install a module (tell your agent):

```text
Install the OpenClerk module ollama using modules/ollama-embeddings/module.json.
Use command semantic-retrieval-adapter, register the module skill, and verify
with `openclerk module` list_modules. Do not edit SQLite directly.
```

## What is explicitly not supported yet

- **Browsing / URL ingestion as a default path** — `ingest_source_url` exists
  but is placement-plan-first; it does not browse the open web autonomously.
- **Automatic video transcript acquisition** — not supported.
- **Hosted service or cloud sync** — fully local. No OpenClerk server, no SaaS.
- **Multi-user / team server** — single-user, single-machine runtime.
- **Broad vector DB memory** — no Pinecone, Weaviate, or default durable vector
  index. Semantic modules are optional and local-only by default.
- **Autonomous memory and routing** — memory and routing are deferred until the
  docs, synthesis, and truth-sync layers are reliable. See
  [`docs/architecture/memory-routing-reference-decision.md`](docs/architecture/memory-routing-reference-decision.md).

---

## Runner

```bash
openclerk document    # doc writes, registry, paths
openclerk retrieval   # search, provenance, synthesis inspection
openclerk capabilities
```

JSON in, JSON out. See runner help:

```bash
openclerk document --help
openclerk retrieval --help
```

Storage: `${XDG_DATA_HOME:-~/.local/share}/openclerk/openclerk.sqlite`  
Override: `OPENCLERK_DATABASE_PATH` or `--db`

## Architecture

[Agent knowledge plane →](docs/architecture/agent-knowledge-plane.md)

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md), [`SECURITY.md`](SECURITY.md), and [`docs/maintainers.md`](docs/maintainers.md).
