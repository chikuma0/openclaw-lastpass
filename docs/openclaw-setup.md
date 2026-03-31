# OpenClaw Setup

This guide assumes:

- `openclaw-lastpass` is installed
- `lpass` is installed and logged in locally
- you have reviewed a draft plan and produced an approved resolver mapping

## 1. Confirm The Binary Path

Find the installed binary:

```bash
command -v openclaw-lastpass
```

Use that absolute path in your OpenClaw config.

## 2. Confirm The Mapping

If you used the default config location, `openclaw-lastpass openclaw` will find it automatically.

Default mapping paths:

- Linux: `~/.config/openclaw-lastpass/mapping.json`
- macOS: `~/Library/Application Support/openclaw-lastpass/mapping.json`

If you keep the mapping elsewhere, either:

- add `--config /absolute/path/to/mapping.json` to the provider `args`
- or expose `OPENCLAW_LASTPASS_CONFIG` and include it in `passEnv`

## 3. Add The Exec Provider

Example provider snippet:

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

If you need a non-default mapping path, this is also valid:

```json
{
  "secrets": {
    "providers": {
      "lastpass": {
        "source": "exec",
        "command": "/absolute/path/to/openclaw-lastpass",
        "args": ["openclaw", "--config", "/absolute/path/to/mapping.json"],
        "passEnv": ["HOME", "PATH"],
        "jsonOnly": true
      }
    }
  }
}
```

The checked-in example lives at [`examples/openclaw.example.json`](../examples/openclaw.example.json).

## 4. Verify Locally Before Wiring It In

Test the resolver directly first:

```bash
openclaw-lastpass doctor
openclaw-lastpass get providers/openai/apiKey
```

Then test protocol mode:

```bash
printf '%s\n' '{"protocolVersion":1,"provider":"lastpass","ids":["providers/openai/apiKey"]}' | openclaw-lastpass openclaw
```

`stdout` is reserved for the JSON response. Errors and diagnostics go to `stderr`.

## 5. Recommended Audit Flow

When you change mappings:

1. update the draft plan or final mapping
2. run `openclaw-lastpass apply --dry-run`
3. run `openclaw-lastpass apply --validate`
4. run `openclaw-lastpass doctor`
5. reload or restart OpenClaw if your environment requires it

## Snippet Helpers

These commands print a recommended provider snippet if `openclaw` is available locally:

```bash
openclaw-lastpass init --print-openclaw-config
openclaw-lastpass apply --plan /path/to/mapping.draft.json --dry-run --print-openclaw-config
```

They do not modify OpenClaw config automatically.
