# openclaw-lastpass

`openclaw-lastpass` is a small, read-only LastPass secret resolver for OpenClaw and direct CLI use.

It is a thin adapter around the locally installed LastPass CLI, `lpass`. It does not manage vaults, browser autofill, sharing, editing, migration, or any broader secret-management workflow. It only resolves explicitly mapped secret IDs to specific LastPass entries and fields.

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

- Go 1.24+ to build from source
- `lpass` installed and available in `PATH`
- A working local LastPass CLI session on the machine where the command runs
- macOS or Linux for v1

Windows-specific behavior is not a v1 target beyond anything that works incidentally through Go and `lpass`.

## Install And Build

Build with Go directly:

```bash
go build -o ./dist/openclaw-lastpass ./cmd/openclaw-lastpass
```

Or use the included `Makefile`:

```bash
make build
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
- No secret persistence. Resolved values are not written to disk by this tool.
- No secret logging. The CLI does not print secret values except where the command contract requires them on `stdout`.
- `doctor` validates mappings by doing read-only lookups and discarding the returned values in memory.
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
