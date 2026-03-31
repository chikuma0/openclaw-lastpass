# INSTALL_INIT_V3_REPORT

## 1. What was built

V3 adds productized installation and first-run onboarding:

- `install.sh`
  - downloads the correct GitHub Release asset for Linux/macOS amd64/arm64
  - installs the binary to `/usr/local/bin` when writable, otherwise `~/.local/bin`
  - verifies the archive shape before installing
  - prints the next steps for `lpass login` and `openclaw-lastpass init`
- `init`
  - checks for `lpass`
  - checks whether the local LastPass CLI session looks usable
  - creates the config directory if needed
  - generates or refreshes `discovery.json` and `mapping.draft.json`
  - prints the exact human-review next steps
  - optionally prints the recommended OpenClaw provider snippet
- release packaging
  - `scripts/build-release.sh`
  - `.github/workflows/release.yml`
  - release assets for:
    - `linux/amd64`
    - `linux/arm64`
    - `darwin/amd64`
    - `darwin/arm64`

## 2. What was intentionally left out

- no browser automation
- no auto-login
- no shell-completion installer
- no automatic OpenClaw config mutation
- no automatic apply step after init
- no password or note-body reads during install/init
- no GUI/TUI
- no telemetry

## 3. How installation works

The installer script:

1. detects OS and architecture
2. resolves the latest GitHub Release tag unless `VERSION` is explicitly set
3. chooses the matching release asset name
4. downloads the binary archive from GitHub Releases
5. verifies that the archive is a valid tarball and contains exactly one `openclaw-lastpass` binary
6. installs it into `/usr/local/bin` if writable, otherwise `~/.local/bin`
7. prints the next steps for `lpass login` and `openclaw-lastpass init`

The release workflow publishes assets with the naming format:

```text
openclaw-lastpass_<tag>_<os>_<arch>.tar.gz
```

Example:

```text
openclaw-lastpass_v0.0.1_linux_amd64.tar.gz
```

## 4. Exact first-run commands on Linux/macOS

Install:

```bash
curl -fsSL https://raw.githubusercontent.com/chikuma0/openclaw-lastpass/main/install.sh | bash
```

Log in to LastPass:

```bash
lpass login you@example.com
```

Run guided setup:

```bash
openclaw-lastpass init
```

Review and validate before applying:

```bash
cat ~/.config/openclaw-lastpass/mapping.draft.json
openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json --dry-run
openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json --validate
```

## 5. Any assumptions about GitHub Releases assets

- release tags use the `v*` pattern, for example `v0.0.1`
- when `VERSION` is not set, `install.sh` can read the latest tag from the GitHub Releases API
- release asset names follow:
  - `openclaw-lastpass_<tag>_<os>_<arch>.tar.gz`
- each asset tarball contains exactly one `openclaw-lastpass` binary
- GitHub Releases is the installation source for normal users rather than source tarballs

## 6. Any assumptions about `lpass`

- `lpass` must be installed separately
- `lpass login` is a human step and remains outside this tool
- `lpass status` is enough for init preflight to decide whether setup can continue
- init/discovery continue to rely on metadata-only listing for safe first-run behavior

## 7. Suggested next steps

- add signed checksum publishing if you want stronger release verification later
- add a dedicated `completion` command only if it stays small and maintenance-free
- add a small fixture-based test for the release build script if the packaging logic grows
- consider a future `apply --prune` mode only if users explicitly ask for mapping cleanup
