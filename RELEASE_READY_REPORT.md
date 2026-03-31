# RELEASE_READY_REPORT

## Status

Release is ready. `v0.0.1` is published and the release assets exist.

Verified:

- GitHub Actions release workflow completed successfully
- GitHub Release `v0.0.1` exists
- published assets match the installer and workflow naming:
  - `openclaw-lastpass_v0.0.1_linux_amd64.tar.gz`
  - `openclaw-lastpass_v0.0.1_linux_arm64.tar.gz`
  - `openclaw-lastpass_v0.0.1_darwin_amd64.tar.gz`
  - `openclaw-lastpass_v0.0.1_darwin_arm64.tar.gz`
  - `SHA256SUMS`
- `install.sh` downloads GitHub Release assets, not source archives
- the published install command is:
  - `curl -fsSL https://raw.githubusercontent.com/chikuma0/openclaw-lastpass/main/install.sh | bash`
- a real installer smoke test against the published `v0.0.1` asset succeeded

## Exact Release/Tag Commands

Already done:

```bash
git push origin main
git tag v0.0.1
git push origin v0.0.1
```

If you ever need to confirm the release and assets manually:

```bash
curl -fsSL https://api.github.com/repos/chikuma0/openclaw-lastpass/releases/tags/v0.0.1
```

## Exact Linux Smoke-Test Commands

```bash
curl -fsSL https://raw.githubusercontent.com/chikuma0/openclaw-lastpass/main/install.sh | bash
openclaw-lastpass --help
command -v openclaw-lastpass
command -v lpass || echo "lpass missing"
```

If `lpass` is installed:

```bash
lpass login you@example.com
openclaw-lastpass init
cat ~/.config/openclaw-lastpass/mapping.draft.json
openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json --dry-run
openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json --validate
```

## Remaining Rough Edges

- the installer relies on the GitHub Releases API when `VERSION` is not set
- the installer verifies archive shape but not signatures
- `init` still depends on `lpass` being installed and logged in, by design
