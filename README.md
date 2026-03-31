# openclaw-lastpass

`openclaw-lastpass` is a small, read-only LastPass secret resolver for OpenClaw and direct CLI use.

It is a thin adapter around the locally installed LastPass CLI, `lpass`. It does not manage vaults, browser autofill, sharing, editing, migration, or any broader secret-management workflow. It resolves explicitly mapped secret IDs to specific LastPass entries and fields, and in v2 it can also generate a metadata-only draft mapping plan for human review.

## Quick Start

Normal users do not need `git clone` or `go build`.

```bash
curl -fsSL https://raw.githubusercontent.com/chikuma0/openclaw-lastpass/main/install.sh | bash
lpass login you@example.com
openclaw-lastpass init
```

## What This Is

- A narrow exec-based secret provider for OpenClaw
- A small CLI for resolving one mapped secret at a time
- A least-privilege bridge between local config and `lpass`
- A read-only tool aimed at API keys, tokens, DB credentials, URLs, and similar machine-usable secrets

## What This Is Not

- Not a general password manager
- Not a browser extension or autofill tool
- Not a LastPass vault abstraction layer
- Not a bulk migration or audit platform
- Not a 1Password replacement
- Not a write/edit/delete client for LastPass entries

## Why It Exists

Some teams already have `lpass` installed, authenticated locally, and want OpenClaw to resolve a small set of machine secrets without exposing a whole vault. This project keeps that integration boring and explicit:

1. You define local secret IDs in a JSON mapping file.
2. Each ID points to one LastPass entry and one field.
3. `openclaw-lastpass` resolves only those mapped IDs through `lpass`.

## Prerequisites

- `lpass` installed and available in `PATH`
- A working local LastPass CLI session on the machine where the command runs
- macOS or Linux for v1

Windows-specific behavior is not a v1 target beyond anything that works incidentally through Go and `lpass`.

## Install

Install via the release installer:

```bash
curl -fsSL https://raw.githubusercontent.com/chikuma0/openclaw-lastpass/main/install.sh | bash
```

The installer:

- detects macOS/Linux and amd64/arm64
- resolves the latest GitHub Release tag unless `VERSION` is set
- downloads the matching GitHub Release asset
- installs to `/usr/local/bin` if writable, otherwise `~/.local/bin`
- verifies the downloaded archive before installing
- prints the next steps:
  - install `lpass` if needed
  - `lpass login you@example.com`
  - `openclaw-lastpass init`

Manual release packaging details are documented in [`docs/release.md`](./docs/release.md).

## Build From Source

Go 1.24+ is only required if you want to build from source.

Build with Go directly:

```bash
go build -o ./dist/openclaw-lastpass ./cmd/openclaw-lastpass
```

Or use the included `Makefile`:

```bash
make build
```

Build local release assets:

```bash
make release-assets VERSION=v0.0.1
```

## Configuration

The v1 config format is a plain JSON object mapping OpenClaw-style secret IDs to LastPass entry/field pairs.

```json
{
  "providers/openai/apiKey": {
    "entry": "OpenClaw/OpenAI",
    "field": "password"
  },
  "providers/anthropic/apiKey": {
    "entry": "OpenClaw/Anthropic",
    "field": "password"
  }
}
```

Examples live in [`examples/mapping.example.json`](./examples/mapping.example.json).

Manual mappings may use LastPass entry names or LastPass unique IDs. The v2 `apply` flow prefers unique IDs in the final resolver config to avoid ambiguous names.

### Supported Field Targets

V1 supports:

- `password`
- `username`
- `notes` or `note`
- Custom field names such as `DATABASE_URL` or `API Token`

Built-in field names are case-insensitive and normalized like this:

- `password` or `Password` -> `lpass show --password <entry>`
- `username` or `Username` -> `lpass show --username <entry>`
- `notes`, `Notes`, or `note` -> `lpass show --notes <entry>`
- any other field value -> `lpass show --field=<field> <entry>`

The preferred config convention is lowercase built-in field names such as `password`, `username`, and `notes`.

## V2 Discovery Workflow

V2 adds a review-first workflow for reducing mapping friction without exposing secret values:

1. `discover` scans LastPass metadata only.
2. It writes:
   - `discovery.json`
   - `mapping.draft.json`
3. You review and edit `mapping.draft.json`.
4. `apply` validates and writes approved entries into the final resolver mapping.

Recommended flow:

```bash
openclaw-lastpass discover
$EDITOR ~/.config/openclaw-lastpass/mapping.draft.json
openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json --dry-run
openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json
```

### Metadata-Only Discovery Warning

`discover` is intentionally conservative:

- it uses `lpass ls --format`
- it does not use `lpass export`
- it does not dump passwords
- it does not dump note bodies by default
- suggestions are heuristics only, not truth

Discovery suggestions are based on entry names and paths, not on secret values.

### Draft Plan Schema

The editable draft plan schema looks like this:

```json
{
  "version": 1,
  "generated_at": "2026-03-31T00:00:00Z",
  "entries": [
    {
      "lastpass_id": "377704248130093254",
      "name": "OPENAI_API_KEY",
      "fullname": "API Keys/dera-next/OPENAI_API_KEY",
      "group": "API Keys/dera-next",
      "suggested_ref_id": "providers/openai/apiKey",
      "suggested_field": "notes",
      "confidence": "high",
      "reason": "matched openai/chatgpt naming with api key or token keywords in entry name/path",
      "approved": false
    }
  ]
}
```

An example draft plan lives in [`examples/mapping.draft.example.json`](./examples/mapping.draft.example.json).

Fields intended for human editing:

- `approved`
- `suggested_ref_id`
- `suggested_field`
- `disabled`

### How Apply Works

`apply` only consumes approved, non-disabled entries from the draft plan.

- It merges approved entries into the existing resolver mapping file.
- It does not delete unrelated mappings.
- It writes LastPass unique IDs into the final `entry` field when they are available in the plan.
- It does not resolve secrets unless `--validate` is explicitly requested.

Example final mapping written by `apply`:

```json
{
  "providers/openai/apiKey": {
    "entry": "377704248130093254",
    "field": "notes"
  }
}
```

## V3 Init Workflow

V3 adds a guided first-run onboarding command:

```bash
openclaw-lastpass init
```

`init` is intentionally narrow:

- checks whether `lpass` is installed
- checks whether the LastPass CLI session looks usable
- creates the config directory if needed
- generates or refreshes:
  - `~/.config/openclaw-lastpass/discovery.json`
  - `~/.config/openclaw-lastpass/mapping.draft.json`
- prints exact next commands for review and apply
- does not fetch secret values by default
- does not auto-apply
- does not silently mutate OpenClaw config

If a draft already exists, `init` keeps it and tells you to rerun with `--refresh` if you want a new one.

Example first run:

```bash
lpass login you@example.com
openclaw-lastpass init
```

Example refresh:

```bash
openclaw-lastpass init --refresh
```

### Config Path Resolution

Resolution order:

1. `--config /path/to/mapping.json`
2. `OPENCLAW_LASTPASS_CONFIG`
3. Default platform path

Default paths:

- macOS: `~/Library/Application Support/openclaw-lastpass/mapping.json`
- Linux: `~/.config/openclaw-lastpass/mapping.json`
- Any platform with `XDG_CONFIG_HOME` set: `$XDG_CONFIG_HOME/openclaw-lastpass/mapping.json`

## Commands

### `init`

Run guided first-run onboarding.

```bash
openclaw-lastpass init
```

Refresh an existing draft:

```bash
openclaw-lastpass init --refresh
```

If OpenClaw is installed, print the recommended provider snippet too:

```bash
openclaw-lastpass init --print-openclaw-config
```

### `discover`

Scan LastPass metadata, classify likely machine-secret candidates, and write review files to disk.

```bash
openclaw-lastpass discover
```

By default this writes:

- `~/.config/openclaw-lastpass/discovery.json`
- `~/.config/openclaw-lastpass/mapping.draft.json`

You can override the output directory or individual output files:

```bash
openclaw-lastpass discover --out-dir /tmp/openclaw-lastpass
openclaw-lastpass discover --discovery-out /tmp/discovery.json --draft-out /tmp/mapping.draft.json
```

### `apply`

Read an approved draft plan and merge approved entries into the final resolver mapping.

```bash
openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json
```

Preview changes without writing:

```bash
openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json --dry-run
```

Optional validation through `lpass`:

```bash
openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json --validate --dry-run
```

Optionally print a recommended OpenClaw provider snippet if `openclaw` is installed:

```bash
openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json --print-openclaw-config --dry-run
```

### `get`

Resolve one configured secret ID.

```bash
openclaw-lastpass get providers/openai/apiKey
```

Print JSON instead of raw value:

```bash
openclaw-lastpass get --json providers/openai/apiKey
```

### `list`

List configured IDs only. No secret values are printed.

```bash
openclaw-lastpass list
```

### `doctor`

Run local diagnostics for:

- `lpass` availability
- config presence and readability
- config structure validity
- LastPass CLI session usability
- per-mapping read-only resolution checks

```bash
openclaw-lastpass doctor
```

### `openclaw`

Read an OpenClaw exec-provider request from `stdin` and emit only JSON on `stdout`.

```bash
printf '%s\n' \
  '{"protocolVersion":1,"provider":"lastpass","ids":["providers/openai/apiKey"]}' \
  | openclaw-lastpass openclaw
```

Request shape:

```json
{
  "protocolVersion": 1,
  "provider": "lastpass",
  "ids": [
    "providers/openai/apiKey"
  ]
}
```

Response shape:

```json
{
  "protocolVersion": 1,
  "values": {
    "providers/openai/apiKey": "..."
  },
  "errors": {
    "some/other/id": {
      "message": "mapping not found"
    }
  }
}
```

Partial success is supported. `stdout` is reserved for the machine-readable response. Diagnostics go to `stderr`.

## OpenClaw Integration

An example config snippet is included at [`examples/openclaw.example.json`](./examples/openclaw.example.json).

Example:

```json
{
  "secrets": {
    "providers": {
      "lastpass": {
        "source": "exec",
        "command": "/absolute/path/to/openclaw-lastpass",
        "args": ["openclaw"],
        "passEnv": ["HOME", "PATH"],
        "jsonOnly": true
      }
    }
  }
}
```

## Security Posture And Limitations

- Read-only only. This tool never creates, edits, or deletes LastPass items.
- Least privilege by mapping. Only configured IDs are resolvable.
- Discovery is metadata-only and does not fetch secret values.
- `init` is metadata-first and does not auto-apply or auto-wire secrets into OpenClaw.
- No secret persistence. Resolved values are not written to disk by this tool.
- No secret logging. The CLI does not print secret values except where the command contract requires them on `stdout`.
- `doctor` validates mappings by doing read-only lookups and discarding the returned values in memory.
- `apply` does not fetch secret values unless `--validate` is explicitly requested.
- Discovery suggestions are guesses based on names and paths and must be reviewed by a human before apply.
- `lpass` controls authentication and local session behavior. This project assumes `lpass` is already installed and authenticated locally.
- Entry names must be unique enough for `lpass` to resolve to a single item. If not, use a unique LastPass entry ID in the mapping instead.
- Multiline values are supported, but `lpass` appends a trailing newline to command output, so exact preservation of an intentional final newline depends on `lpass` behavior.

## Migration Guidance

If you are moving machine secrets out of documents or ad hoc notes into LastPass, keep the migration explicit:

- Create dedicated LastPass entries for machine credentials instead of reusing personal login entries.
- Use stable, unique entry names or map directly by LastPass item ID.
- Prefer custom field names for structured machine data such as `DATABASE_URL`, `API_TOKEN`, or `SERVICE_URL`.
- Map only the specific secret IDs that OpenClaw actually needs.
- Treat this tool as a resolver adapter, not a way to expose your whole vault to an agent.

## Testing

Run the test suite:

```bash
go test ./...
```

Format the code:

```bash
go fmt ./...
```
