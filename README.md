# openclaw-lastpass

`openclaw-lastpass` is a small, read-only bridge between LastPass CLI (`lpass`) and OpenClaw's exec-based secret provider model. It also works directly from the command line for one-off lookups.

This project is for people who already use LastPass locally and want a safer workflow for machine secrets than copying API keys, tokens, URLs, and credentials through docs, spreadsheets, or ad hoc notes. It is not a new password manager. It is not a vault platform. It is a narrow adapter for explicit, reviewable secret mappings.

## Why It Exists

If your team is already on LastPass, the fastest improvement is often not "replace your whole secret stack this week." It is "move machine secrets into clear LastPass entries and resolve only the few values your tools actually need."

`openclaw-lastpass` is that smaller step:

- safer than keeping machine secrets in docs or spreadsheets
- lower-friction than buying or migrating to a different secret manager just to get OpenClaw working
- narrower than a full-vault integration, because only mapped secret IDs are resolvable

## Who It Is For

- developers or teams already using LastPass locally on macOS or Linux
- OpenClaw users who want an exec-based secret provider backed by `lpass`
- people who want explicit, least-privilege mappings for API keys, tokens, DB credentials, service URLs, and similar machine-usable secrets

## Who It Is Not For

- people looking for browser autofill or browser automation
- people looking for a full LastPass client, vault browser, or admin platform
- people looking for auto-login, auto-apply, blind vault export, or bulk migration tooling
- people looking for a universal secret-manager abstraction layer

## Safety Boundaries

- read-only only. The tool never creates, edits, or deletes LastPass items.
- metadata-first discovery. `discover` and `init` do not fetch passwords or note bodies by default.
- human-reviewed apply. Draft plans are suggestions, not truth.
- no blind vault dump. The project does not use `lpass export`.
- no browser automation or browser extension integration.
- no silent OpenClaw mutation. Provider setup is printed as guidance only.
- no secret logging. Secret values are never logged and are not persisted by this tool.

## 3-Minute Quickstart

1. Install the latest release:

```bash
curl -fsSL https://raw.githubusercontent.com/chikuma0/openclaw-lastpass/main/install.sh | bash
```

2. Log in with the local LastPass CLI:

```bash
lpass login you@example.com
```

3. Generate discovery metadata and a reviewable draft plan:

```bash
openclaw-lastpass init
```

4. Review the `mapping.draft.json` path printed by `init`, then edit it:

```bash
$EDITOR /path/to/mapping.draft.json
```

5. Preview the approved mapping changes:

```bash
openclaw-lastpass apply --plan /path/to/mapping.draft.json --dry-run
```

6. Validate approved entries through `lpass` before writing:

```bash
openclaw-lastpass apply --plan /path/to/mapping.draft.json --validate
```

Default config locations:

- Linux: `~/.config/openclaw-lastpass`
- macOS: `~/Library/Application Support/openclaw-lastpass`

If you want a first diagnostic pass at any point, run:

```bash
openclaw-lastpass doctor
```

## Install

Normal users do not need `git clone` or `go build`.

```bash
curl -fsSL https://raw.githubusercontent.com/chikuma0/openclaw-lastpass/main/install.sh | bash
```

The installer:

- detects Linux or macOS and amd64 or arm64
- downloads the matching GitHub Releases asset
- installs to `/usr/local/bin` when writable, otherwise `~/.local/bin`
- verifies the archive layout before installing
- prints the next commands: `lpass login` and `openclaw-lastpass init`

If the install directory is not on `PATH`, the script tells you exactly what to export.

## How The Workflow Fits Together

1. `init` or `discover` scans LastPass metadata through `lpass ls --format`.
2. The tool writes `discovery.json` and `mapping.draft.json`.
3. You review the draft plan and mark entries as approved.
4. `apply` writes the final resolver mapping, preferring LastPass item IDs when available.
5. `get` and `openclaw` resolve only the secret IDs present in that mapping file.

That design is deliberate. Discovery is allowed to suggest. Apply is allowed to write approved mappings. Neither step is allowed to guess and silently wire secrets into your workflow.

## Recommended LastPass Structure

The best results come from treating LastPass entries as machine-secret containers, not as general notes or reused personal login items.

### Pattern 1: One Secret Per Entry

Use this when the entry exists to hold one opaque token.

LastPass entry:

- name: `Machine Secrets/acme/prod/OPENAI_API_KEY`
- password field: the actual OpenAI API key

Resolver mapping:

```json
{
  "providers/openai/apiKey": {
    "entry": "Machine Secrets/acme/prod/OPENAI_API_KEY",
    "field": "password"
  }
}
```

### Pattern 2: One Entry Per Repo x Environment

Use this when several related machine secrets belong to one runtime or deployment boundary.

LastPass entry:

- name: `Machine Secrets/acme/prod/runtime`
- custom fields:
  - `ANTHROPIC_API_KEY`
  - `DATABASE_URL`
  - `SUPABASE_SERVICE_ROLE`

Resolver mapping:

```json
{
  "providers/anthropic/apiKey": {
    "entry": "Machine Secrets/acme/prod/runtime",
    "field": "ANTHROPIC_API_KEY"
  },
  "services/acme/databaseUrl": {
    "entry": "Machine Secrets/acme/prod/runtime",
    "field": "DATABASE_URL"
  },
  "supabase/serviceRole": {
    "entry": "Machine Secrets/acme/prod/runtime",
    "field": "SUPABASE_SERVICE_ROLE"
  }
}
```

Built-in field names are normalized like this:

- `password` -> `lpass show --password`
- `username` -> `lpass show --username`
- `notes` or `note` -> `lpass show --notes`
- anything else -> `lpass show --field=<field>`

Prefer lowercase built-ins in config files: `password`, `username`, and `notes`.

One important limitation today: the final `mapping.json` can point several secret IDs at the same LastPass entry, but `discover` and `apply` stay conservative and only draft one suggestion per LastPass item. Bundled custom-field entries are fully supported in the final resolver mapping, but you may still hand-edit the final mapping for extra fields after apply.

More structuring guidance lives in [`docs/secret-structure.md`](./docs/secret-structure.md).

## OpenClaw Integration

Once you have an approved mapping, wire the binary into OpenClaw as an exec-based secret provider. The included example is in [`examples/openclaw.example.json`](./examples/openclaw.example.json).

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

Use `command -v openclaw-lastpass` to find the installed binary path. If you keep the mapping file outside the default location, either pass `--config` in `args` or expose `OPENCLAW_LASTPASS_CONFIG` to the process.

If OpenClaw is installed locally, these commands print a recommended provider snippet without changing anything automatically:

```bash
openclaw-lastpass init --print-openclaw-config
openclaw-lastpass apply --plan /path/to/mapping.draft.json --dry-run --print-openclaw-config
```

Detailed setup notes are in [`docs/openclaw-setup.md`](./docs/openclaw-setup.md).

## Commands

- `init`: guided first run that checks `lpass`, checks login state, and writes `discovery.json` plus `mapping.draft.json`
- `discover`: metadata-only discovery without the first-run guidance
- `apply`: validates and writes approved draft plan entries into the final resolver mapping
- `get`: resolves one configured secret ID directly
- `list`: lists configured secret IDs without printing values
- `doctor`: checks local install, mapping validity, and read-only resolution health
- `openclaw`: protocol mode for OpenClaw's exec-based provider model

Use `openclaw-lastpass <command> --help` for full flags.

## Examples

- Final resolver mapping with entry names: [`examples/mapping.example.json`](./examples/mapping.example.json)
- Final resolver mapping with LastPass item IDs: [`examples/mapping.final.example.json`](./examples/mapping.final.example.json)
- Editable draft plan: [`examples/mapping.draft.example.json`](./examples/mapping.draft.example.json)
- OpenClaw provider snippet: [`examples/openclaw.example.json`](./examples/openclaw.example.json)

## Troubleshooting

Start with:

```bash
openclaw-lastpass doctor
```

Common problems:

- `lpass` missing from `PATH`
- not logged in to LastPass locally
- ambiguous entry names that match multiple vault items
- using `password` or `notes` as custom field names by mistake
- install directory not present on `PATH`

See [`docs/troubleshooting.md`](./docs/troubleshooting.md) for concrete fixes.

## Docs

- [`docs/secret-structure.md`](./docs/secret-structure.md): how to structure LastPass entries for single-secret and bundled repo/environment use cases
- [`docs/openclaw-setup.md`](./docs/openclaw-setup.md): wiring the approved mapping into OpenClaw
- [`docs/troubleshooting.md`](./docs/troubleshooting.md): install, login, field, and path troubleshooting
- [`docs/release.md`](./docs/release.md): release artifact naming and packaging flow
- [`VISION.md`](./VISION.md): project intent, audience, and scope guardrails

## Scope And Non-Goals

This project intentionally does not try to become:

- a new password manager
- a browser extension or browser automation tool
- a full LastPass vault abstraction
- an auto-login system
- a write/edit/delete client for LastPass
- a bulk import, migration, or audit platform
- a generic secret-management platform

Simplicity and trust matter more than feature count here.

## Contributing, Support, And Security

- Contribution guide: [`CONTRIBUTING.md`](./CONTRIBUTING.md)
- Support and bug-reporting guidance: [`SUPPORT.md`](./SUPPORT.md)
- Security reporting guidance: [`SECURITY.md`](./SECURITY.md)

## Build From Source

If you want to build locally:

```bash
go build -o ./dist/openclaw-lastpass ./cmd/openclaw-lastpass
```

Run tests:

```bash
go test ./...
```
