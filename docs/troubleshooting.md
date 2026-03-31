# Troubleshooting

Start with:

```bash
openclaw-lastpass doctor
```

That is the fastest way to check `lpass`, the mapping file, and read-only resolution health.

## `lpass` Missing

Symptom:

- `openclaw-lastpass init` says `lpass` is not installed or not found in `PATH`

Fix:

1. install the LastPass CLI from your package manager or from the official `lastpass-cli` project
2. confirm it is on `PATH`

```bash
command -v lpass
```

3. rerun:

```bash
lpass login you@example.com
openclaw-lastpass init
```

## Not Logged In To LastPass

Symptom:

- `init`, `discover`, `doctor`, or `get` says login is required

Fix:

```bash
lpass login you@example.com
lpass status
```

Then rerun the failing command.

## Ambiguous Entry

Symptom:

- resolution fails because one entry name matches multiple LastPass items

Fix:

- rename the LastPass entries so they are unique
- or use the LastPass item ID in `mapping.json`

Example:

```json
{
  "providers/openai/apiKey": {
    "entry": "377704248130093254",
    "field": "password"
  }
}
```

## Wrong Field Type

Symptom:

- `lpass show --password` works but `--field=Password` does not
- a mapping points at the wrong built-in or custom field

Fix:

- use `password`, `username`, or `notes` for built-in fields
- use the exact custom field name for everything else

Examples:

- good built-in field: `password`
- also valid built-in spelling: `Password`
- good custom field: `DATABASE_URL`
- wrong field choice for a custom-field value: `password` when the value actually lives in `DATABASE_URL`

## `openclaw-lastpass` Installed But Not On `PATH`

Symptom:

- the installer says the binary was installed, but the shell cannot find it

Fix:

1. find the install directory the script printed
2. add it to `PATH`

Example:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Then reopen the shell or re-source your shell profile.

## Release Install Problems

If the installer cannot find an asset:

- confirm the GitHub release exists
- confirm the asset name matches your OS and architecture
- confirm `curl`, `tar`, `mktemp`, and `uname` are available

Useful checks:

```bash
uname -s
uname -m
command -v curl
command -v tar
```

To pin a specific published version:

```bash
curl -fsSL https://raw.githubusercontent.com/chikuma0/openclaw-lastpass/main/install.sh | VERSION=v0.0.1 bash
```

To force a user-local install directory:

```bash
curl -fsSL https://raw.githubusercontent.com/chikuma0/openclaw-lastpass/main/install.sh | INSTALL_DIR="$HOME/.local/bin" bash
```

## OpenClaw Wiring Problems

If OpenClaw is not seeing the provider:

- confirm the `command` path is absolute and correct
- confirm the mapping file is in the default location or passed with `--config`
- confirm `HOME` and `PATH` are passed through
- if you rely on `OPENCLAW_LASTPASS_CONFIG`, include it in `passEnv`

Use these checks before blaming OpenClaw:

```bash
openclaw-lastpass doctor
printf '%s\n' '{"protocolVersion":1,"provider":"lastpass","ids":["providers/openai/apiKey"]}' | openclaw-lastpass openclaw
```

## Still Stuck

Open an issue with:

- the exact command you ran
- the exact stderr output
- your OS and architecture
- whether you used the release installer or built from source
- a redacted mapping snippet if relevant

Do not paste secret values into issues.
