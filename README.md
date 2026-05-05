# OpenClerk

OpenClerk is a local-first knowledge-plane runtime for agents. The public
surface is the `openclerk` JSON runner plus the OpenClerk skill.

## Install

Tell your agent:

```text
Install OpenClerk into $HOME/.local/bin from the latest release unless I
specify a version. Register skills/openclerk/SKILL.md with your native skill
system. Verify command -v openclerk, openclerk --version, and the installed
skill path. Do not report OpenClerk installed until both the runner and skill
are installed.
```

Detailed install commands live in `docs/install.md`.

## Upgrade

Tell your agent:

```text
Upgrade OpenClerk by rerunning the installer for the latest or requested
version. Keep the durable runner location, re-register the matching
skills/openclerk/SKILL.md skill, and verify command -v openclerk,
openclerk --version, and the installed skill path.
```

Detailed upgrade commands live in `docs/install.md`.

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

### Agent Module Instructions

Tell your agent:

```text
Install an OpenClerk module only through `openclerk module`.
Do not edit SQLite directly.
Use repo-relative manifest and skill paths in docs or reports.
Register the module skill only when the host opts into that module.
After install, verify with `openclerk module` list_modules and explicit `semantic_search`.
```

Modules are optional building blocks. OpenClerk verifies the manifest before
routing `semantic_search` through an installed provider module.

Available installable modules:

| Module | Provider | Purpose | Skill |
| --- | --- | --- | --- |
| `modules/ollama-embeddings/module.json` | `ollama` | Local-first semantic retrieval | `modules/ollama-embeddings/skill/ollama-embeddings/SKILL.md` |
| `modules/gemini-embeddings/module.json` | `gemini` | Explicit opt-in provider semantic retrieval with retry/backoff | `modules/gemini-embeddings/skill/gemini-embeddings/SKILL.md` |

Exact module commands and provider setup live in `modules/docs/install.md`.

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
