# RELEASE_SMOKE_TEST_REPORT

## What I verified

- Release asset naming is consistent across:
  - `scripts/build-release.sh`
  - `.github/workflows/release.yml`
  - `install.sh`
- Expected asset format is:
  - `openclaw-lastpass_<tag>_<os>_<arch>.tar.gz`
- Intended release targets match the requested scope:
  - `linux/amd64`
  - `linux/arm64`
  - `darwin/amd64`
  - `darwin/arm64`
- `install.sh` downloads GitHub Release assets rather than source archives
- `install.sh` detects OS/arch correctly for Linux/macOS and amd64/arm64
- `install.sh` installs to `/usr/local/bin` when writable, otherwise `~/.local/bin`
- `install.sh` warns clearly when:
  - the release asset is missing
  - the install directory is not on `PATH`
  - `lpass` is missing
- `openclaw-lastpass init` does not auto-apply and does not silently mutate OpenClaw config
- `openclaw-lastpass init` fails safely and concisely when `lpass` is missing
- The Go test suite passes
- Local release packaging succeeds for all four target platforms

## What I changed

- Fixed stale doc examples that still referenced `v0.3.0`; they now use `v0.0.1`
- Tightened the installer error message when latest-release tag lookup fails so it suggests rerunning with `VERSION=v0.0.1`
- Added this release/smoke-test handoff report

## Release readiness

`v0.0.1` is ready to release.

Current remaining step is publication: commit the prepared tree, push `main`, and push the `v0.0.1` tag so the GitHub Actions release workflow can publish the release assets.

## Exact commands if manual publishing is needed

From the repository root:

```bash
git add .
git commit -m "Prepare v0.0.1 release"
git push origin main
git tag v0.0.1
git push origin v0.0.1
```

If you want to build the release assets locally before tagging:

```bash
./scripts/build-release.sh v0.0.1
```

## Exact commands for the first Linux smoke test

On a Linux machine with network access:

```bash
curl -fsSL https://raw.githubusercontent.com/chikuma0/openclaw-lastpass/main/install.sh | bash
openclaw-lastpass --help
command -v openclaw-lastpass
command -v lpass || echo "lpass missing"
```

If `lpass` is not installed yet, install it using that machine’s package manager, then continue:

```bash
lpass login you@example.com
openclaw-lastpass init
cat ~/.config/openclaw-lastpass/mapping.draft.json
openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json --dry-run
openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json --validate
```

Optional OpenClaw guidance:

```bash
openclaw-lastpass init --print-openclaw-config
```

## Exact install command users should run after release

```bash
curl -fsSL https://raw.githubusercontent.com/chikuma0/openclaw-lastpass/main/install.sh | bash
```

## Known rough edges

- `install.sh` resolves the latest release tag through the GitHub Releases API when `VERSION` is not set
- the installer verifies archive shape, not signatures
- `init` can only print the OpenClaw provider snippet when `openclaw` is installed locally
- the current smoke test here could not complete a real `lpass login` flow because `lpass` is not installed in this environment
