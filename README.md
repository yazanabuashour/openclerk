# OpenClerk

OpenClerk is a local-first knowledge-plane runtime for agents. The public
surface is the `openclerk` JSON runner plus the OpenClerk skill.

## Install

Install the runner:

```bash
OPENCLERK_INSTALL_DIR="$HOME/.local/bin" sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/latest/download/install.sh)"
```

Install a pinned release:

```bash
OPENCLERK_INSTALL_DIR="$HOME/.local/bin" OPENCLERK_VERSION=v0.2.3 sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/download/v0.2.3/install.sh)"
```

Register the skill from `skills/openclerk/SKILL.md` with the agent's native
skill system. A complete install has both:

```bash
command -v openclerk
openclerk --version
```

## Upgrade

Rerun the installer, then re-register the matching `skills/openclerk/SKILL.md`
skill.

## Runner

OpenClerk reads one strict JSON object from stdin and writes one JSON result to
stdout:

```bash
printf '%s\n' '{"action":"search","search":{"text":"architecture","limit":10}}' |
  openclerk retrieval
```

Runner domains:

```bash
openclerk document
openclerk retrieval
```

Use runner help for current request shapes:

```bash
openclerk document --help
openclerk retrieval --help
```

`openclerk retrieval search` remains lexical and citation-bearing. Use the
installed runner's help output for other supported actions.

## Modules

Module commands are available from the current source checkout and require a
runner build that includes the `openclerk module` domain. Released runners
through `v0.2.3` do not include this domain.

```bash
mise exec -- go build -o "$HOME/.local/bin/openclerk" ./cmd/openclerk
openclerk module --help
```

Modules are optional building blocks. Installing a module stores only its
enabled state, manifest digest, command, args, and redacted provider config in
SQLite `runtime_config`. OpenClerk verifies the manifest before routing
`semantic_search` through the module. Removing a module removes the OpenClerk
registration; it does not delete unrelated credentials or external tools.

Available installable modules:

| Module | Provider | Purpose | Skill |
| --- | --- | --- | --- |
| `modules/ollama-embeddings/module.json` | `ollama` | Local-first semantic retrieval | `modules/ollama-embeddings/skill/ollama-embeddings/SKILL.md` |
| `modules/gemini-embeddings/module.json` | `gemini` | Explicit opt-in provider semantic retrieval with retry/backoff | `modules/gemini-embeddings/skill/gemini-embeddings/SKILL.md` |

Shared command used by both modules:

```text
semantic-retrieval-adapter search
```

The adapter lives at `modules/semantic-retrieval-adapter`. Install or build it
separately from OpenClerk core and make `semantic-retrieval-adapter` available
on `PATH` before registering a provider module.

From a repo checkout:

```bash
mise exec -- go build -o "$HOME/.local/bin/semantic-retrieval-adapter" ./modules/semantic-retrieval-adapter
command -v semantic-retrieval-adapter
```

### Agent Module Instructions

Tell an agent:

```text
Install an OpenClerk module only through `openclerk module`.
Do not edit SQLite directly.
Use repo-relative manifest and skill paths in docs or reports.
Register the module skill only when the host opts into that module.
After install, verify with `openclerk module` list_modules and run explicit `semantic_search`.
```

Install Ollama embeddings:

```bash
printf '%s\n' '{"action":"install_module","module":{"provider":"ollama","manifest_path":"modules/ollama-embeddings/module.json","command":"semantic-retrieval-adapter","provider_config":{"embedding_model":"embeddinggemma","ollama_url":"http://localhost:11434"}}}' |
  openclerk module
```

Install Gemini embeddings:

```bash
printf '%s\n' '{"action":"install_module","module":{"provider":"gemini","manifest_path":"modules/gemini-embeddings/module.json","command":"semantic-retrieval-adapter","provider_config":{"embedding_model":"gemini-embedding-001","gemini_api_base":"https://generativelanguage.googleapis.com/v1beta","embedding_output_dimensions":"3072"}}}' |
  openclerk module
```

Configure a module:

```bash
printf '%s\n' '{"action":"configure_module","config":{"provider":"ollama","enabled":true,"provider_config":{"embedding_model":"embeddinggemma"}}}' |
  openclerk module
```

List modules:

```bash
printf '%s\n' '{"action":"list_modules"}' | openclerk module
```

Remove a module:

```bash
printf '%s\n' '{"action":"remove_module","provider":"ollama"}' | openclerk module
```

Run explicit semantic search after a module is installed:

```bash
printf '%s\n' '{"action":"semantic_search","semantic_search":{"query":"semantic recall citation quality","path_prefix":"docs/","limit":10,"provider":"ollama"}}' |
  openclerk retrieval
```

Gemini requires `runtime_config:GEMINI_API_KEY` in the target OpenClerk
database. OpenClerk reports only the credential reference, request count, retry
count, and backoff seconds; it does not print the key. There is no hidden
provider fallback and no default semantic ranking promotion.

## Local Storage

The default database is:

```text
${XDG_DATA_HOME:-~/.local/share}/openclerk/openclerk.sqlite
```

Override it with `OPENCLERK_DATABASE_PATH` or `--db`.

Inspect configured paths:

```bash
printf '%s\n' '{"action":"resolve_paths"}' | openclerk document
printf '%s\n' '{"action":"inspect_layout"}' | openclerk document
```

Bind an existing vault once:

```bash
openclerk init --vault-root <vault-root>
```

## Development

Use repo-pinned tools through `mise exec -- ...`:

```bash
mise install
test -z "$(gofmt -l $(git ls-files '*.go'))"
mise exec -- golangci-lint run ./...
mise exec -- go test ./...
mise exec -- ./scripts/validate-committed-artifacts.sh
mise exec -- ./scripts/validate-agent-skill.sh skills/openclerk
mise exec -- ./scripts/validate-agent-skill.sh modules/ollama-embeddings/skill/ollama-embeddings
mise exec -- ./scripts/validate-agent-skill.sh modules/gemini-embeddings/skill/gemini-embeddings
mise exec -- ./scripts/validate-release-docs.sh v0.2.3
```

## Releases

Tagged releases publish platform archives, the skill archive, installer,
source archive, checksums, SBOM, and GitHub attestations. See
`docs/release-verification.md`.

## Contributing

See `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, and
`docs/maintainers.md`.
