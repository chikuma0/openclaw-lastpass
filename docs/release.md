# Release Packaging

Release artifacts are built for:

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`

Asset naming format:

```text
openclaw-lastpass_<tag>_<os>_<arch>.tar.gz
```

Example:

```text
openclaw-lastpass_v0.0.1_linux_amd64.tar.gz
```

Local packaging command:

```bash
./scripts/build-release.sh v0.0.1
```

GitHub Actions publishes release assets automatically when a `v*` tag is pushed.
