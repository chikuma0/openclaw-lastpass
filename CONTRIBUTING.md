# Contributing

Thanks for considering a contribution.

This project is intentionally small. The main maintenance goal is not feature count. It is trust, clarity, and a reliable read-only workflow for explicitly mapped machine secrets.

## Scope Guardrails

Contributions that are usually a good fit:

- bug fixes
- install or release fixes
- test coverage improvements
- docs improvements
- error-message clarity
- small portability fixes within the current Linux and macOS scope

Contributions that might be a fit if they stay tightly scoped:

- small workflow improvements that preserve the read-only, review-first model
- small OpenClaw integration polish that does not turn this into a platform

Contributions that are out of scope:

- browser automation
- browser extension work
- auto-login
- LastPass write, edit, create, or delete support
- full-vault browsing or export tooling
- migration/import platforms
- generic multi-provider secret-management abstractions
- telemetry, cloud services, or background agents

If a change expands the project's purpose, open an issue before writing code.

## Development Basics

Build the binary:

```bash
go build -o ./dist/openclaw-lastpass ./cmd/openclaw-lastpass
```

Run tests:

```bash
go test ./...
```

Format Go code:

```bash
go fmt ./...
```

If you change installer or release scripts, also run:

```bash
bash -n install.sh
bash -n scripts/build-release.sh
```

## PR Expectations

- Keep changes focused.
- Preserve the read-only model.
- Do not log or persist secret values.
- Keep stdout and stderr behavior stable for command contracts that depend on machine-readable output.
- Update docs and tests when behavior changes.
- Call out any rough edges or tradeoffs directly.

## Reporting Bugs

Please include:

- the exact command
- the exact stderr output
- OS and architecture
- install method
- `openclaw-lastpass` version
- `lpass` version if relevant

Please redact secret values before posting.

Support and reporting guidance is in [`SUPPORT.md`](./SUPPORT.md). Security reporting guidance is in [`SECURITY.md`](./SECURITY.md).
