# OpenClerk Module Install

OpenClerk modules are optional building blocks. Install them only through
`openclerk module`; do not edit SQLite directly.

## Available Modules

| Module | Provider | Skill |
| --- | --- | --- |
| `modules/ollama-embeddings/module.json` | `ollama` | `modules/ollama-embeddings/skill/ollama-embeddings/SKILL.md` |
| `modules/gemini-embeddings/module.json` | `gemini` | `modules/gemini-embeddings/skill/gemini-embeddings/SKILL.md` |

Both modules use:

```text
semantic-retrieval-adapter search
```

The adapter lives at `modules/semantic-retrieval-adapter`. Build or install it
separately from OpenClerk core and make `semantic-retrieval-adapter` available
on `PATH` before registering a provider module.

```bash
mise exec -- go build -o "$HOME/.local/bin/semantic-retrieval-adapter" ./modules/semantic-retrieval-adapter
command -v semantic-retrieval-adapter
```

## Register Modules

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

## Use Semantic Search

Run explicit semantic search after a module is installed:

```bash
printf '%s\n' '{"action":"semantic_search","semantic_search":{"query":"semantic recall citation quality","path_prefix":"docs/","limit":10,"provider":"ollama"}}' |
  openclerk retrieval
```

Gemini requires `runtime_config:GEMINI_API_KEY` in the target OpenClerk
database. OpenClerk reports only the credential reference, request count, retry
count, and backoff seconds; it does not print the key.

There is no hidden provider fallback and no default semantic ranking promotion.
Removing a module removes OpenClerk's registration for that provider; it does
not delete unrelated credentials, external tools, or user-cache artifacts.
