# Contributing

Outside contributors do not need Beads to contribute to this repository.

## Current project shape

This repository is still in the scaffold stage. There is no runnable app or published package yet, so contribution expectations currently focus on repository policy, documentation, workflow automation, and project setup.

If a pull request introduces the first runnable component, it should also introduce the tests, local run instructions, and compatibility notes needed to support that component.

## Local setup

Maintainers prefer:

```bash
mise install
```

Outside contributors may use their own local tooling if they can satisfy the repository checks.

Beads and Dolt are maintainer-only tools. They are optional for outside contributors and are not required to open, review, or merge pull requests.

## Pull request expectations

- Keep changes reviewable without access to Beads state.
- Update repository docs when the public contract changes.
- Do not commit credentials, private infrastructure details, or sensitive sample data.
- Route security issues through the private process in [SECURITY.md](SECURITY.md), not through public issues or pull requests.

## Checks and review rules

Current pull request checks validate repository policy and dependency-review safety. They do not imply that a runtime, package, or deployment contract exists yet.

Before the project ships runnable code, maintainers expect pull requests to leave the repository in a public-safe, policy-consistent state. As code lands, this document will expand to cover the actual test and compatibility gates.

## Support and compatibility

Before `0.1.0`, compatibility is best effort and may change between releases. Maintainers do not currently promise support for any operating system, runtime, or deployment target because no such product surface exists yet.

Maintainer workflow notes live in [docs/maintainers.md](docs/maintainers.md).
