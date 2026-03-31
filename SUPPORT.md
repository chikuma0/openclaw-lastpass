# Support

## Before Opening An Issue

Run:

```bash
openclaw-lastpass doctor
```

That catches most local setup problems quickly.

## Where To Ask For Help

Use GitHub Issues for:

- bug reports
- installation problems
- focused usage questions
- docs problems

Use [`SECURITY.md`](./SECURITY.md) for security-sensitive reports.

## What To Include

Please include:

- `openclaw-lastpass` version
- OS and architecture
- whether you used the release installer or built from source
- the exact command you ran
- the exact stderr output
- whether `lpass` is installed and logged in
- a redacted mapping snippet if the issue is config-related

Helpful commands:

```bash
command -v openclaw-lastpass
command -v lpass
openclaw-lastpass doctor
```

## What Not To Include

Do not paste:

- secret values
- full vault exports
- screenshots that reveal tokens
- internal service credentials

If an issue only reproduces with real values, redact them before posting.
