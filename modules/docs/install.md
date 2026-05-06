# OpenClerk Module Install and Upgrade

OpenClerk modules are optional building blocks. Install them only through
`openclerk module`; do not edit SQLite directly.

## Available Modules

| Module name | Provider | Command | Manifest | Skill |
| --- | --- | --- | --- | --- |
| `ollama-embeddings` | `ollama` | `semantic-retrieval-adapter` | `modules/ollama-embeddings/module.json` | `modules/ollama-embeddings/skill/ollama-embeddings/SKILL.md` |
| `gemini-embeddings` | `gemini` | `semantic-retrieval-adapter` | `modules/gemini-embeddings/module.json` | `modules/gemini-embeddings/skill/gemini-embeddings/SKILL.md` |
| `tesseract-ocr` | `tesseract` | `tesseract` | `modules/tesseract-ocr/module.json` | `modules/tesseract-ocr/skill/tesseract-ocr/SKILL.md` |

## Install a Module Release

Run a module installer to install the latest module release:

```bash
OPENCLERK_MODULE=ollama-embeddings sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/download/ollama-embeddings%2Fv0.1.0/install-module.sh)"
```

Set `OPENCLERK_MODULE_VERSION` to install a pinned module release:

```bash
OPENCLERK_MODULE=ollama-embeddings OPENCLERK_MODULE_VERSION=v0.1.0 sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/download/ollama-embeddings%2Fv0.1.0/install-module.sh)"
```

The installer downloads the module archive, verifies checksums, installs
bundled commands into `OPENCLERK_INSTALL_DIR` or `$HOME/.local/bin`, installs
module files under `${XDG_DATA_HOME:-$HOME/.local/share}/openclerk/modules`,
and prints the registration command to run with `openclerk module`.

## Upgrade a Module Release

Rerun a module installer for the latest or requested version:

```bash
OPENCLERK_MODULE=ollama-embeddings sh -c "$(curl -fsSL https://github.com/yazanabuashour/openclerk/releases/download/ollama-embeddings%2Fv0.1.0/install-module.sh)"
```

Then refresh registration with `openclerk module` using the installed manifest
path printed by the installer. Preserve any existing provider config from:

```bash
printf '%s\n' '{"action":"list_modules"}' | openclerk module
```

Module-only releases are for module code, manifests, and skills. Use a normal
OpenClerk core release when the `openclerk` runner contract or module
registration semantics change.

Maintainers build module release assets with:

```bash
mise exec -- ./scripts/build-module-release-bundle.sh ollama-embeddings v0.1.0 dist
```

Published module releases include `scripts/install-module.sh` as
`install-module.sh`.

Embedding modules use:

```text
semantic-retrieval-adapter search
```

The adapter lives at `modules/semantic-retrieval-adapter`. Build or install it
separately from OpenClerk core and make `semantic-retrieval-adapter` available
on `PATH` before registering a provider module. Semantic module registration
does not support overriding this executable or passing `command_args`; OpenClerk
always runs `semantic-retrieval-adapter search`.

```bash
mise exec -- go build -o "$HOME/.local/bin/semantic-retrieval-adapter" ./modules/semantic-retrieval-adapter
command -v semantic-retrieval-adapter
```

The OCR module uses `tesseract` for image OCR and `ocrmypdf` for PDF OCR.
Install those tools separately and make both commands available on `PATH`
before registering the module.

## Register or Refresh Module Registration

For semantic modules, `command` may be omitted or set to
`semantic-retrieval-adapter` for compatibility with older instructions.
`command_args` are rejected.

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

`gemini_api_base` is compatibility-only for the production adapter and must
remain `https://generativelanguage.googleapis.com/v1beta`; custom Gemini
endpoints and provider mimics are rejected.

Install local Tesseract OCR:

```bash
printf '%s\n' '{"action":"install_module","module":{"kind":"ocr_provider","provider":"tesseract","manifest_path":"modules/tesseract-ocr/module.json","command":"tesseract","provider_config":{"ocrmypdf_command":"ocrmypdf","language":"eng"}}}' |
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

## Use OCR Review

Text-extractable documents do not need OCR. Use normal
`artifact_candidate_plan` local artifact planning for UTF-8 text, markdown, and
text-bearing PDFs.

Use OCR review only for common image files, scan-only PDFs, or PDFs whose
embedded text is bad or partial:

```bash
printf '%s\n' '{"action":"artifact_candidate_plan","artifact":{"local_path":"<explicit-user-local-file>","artifact_kind":"receipt","text_extraction":"ocr_review","ocr_provider":"tesseract","limit":5}}' |
  openclerk document
```

OCR review is read-only candidate planning. It reports extractor identity,
versions, language, provenance, warnings, duplicate status, and
`planned_no_write`; durable writes still require approval through
`create_document` or `ingest_source_url`.
