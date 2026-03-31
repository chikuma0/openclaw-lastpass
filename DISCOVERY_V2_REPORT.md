# DISCOVERY_V2_REPORT

## 1. What was built

V2 adds two new commands without changing the existing v1 resolver flow:

- `discover`
  - scans LastPass in a metadata-first way using `lpass ls --format`
  - writes a raw metadata snapshot to `discovery.json`
  - writes an editable suggestion plan to `mapping.draft.json`
  - prints a short human-readable summary
- `apply`
  - reads an approved draft plan
  - validates approved entries
  - writes or updates the local resolver mapping file
  - prefers LastPass unique IDs in the final mapping entries
  - supports `--dry-run`
  - supports optional `--validate` read-only checks through `lpass`
  - can optionally print a recommended OpenClaw exec-provider snippet when OpenClaw is installed

The implementation also adds:

- a draft plan schema in `internal/plan`
- isolated suggestion heuristics in `internal/heuristic`
- metadata parsing for `lpass ls --format` in `internal/lastpass`
- discovery plan generation in `internal/discovery`
- apply logic in `internal/apply`

## 2. What was intentionally left out

- no browser automation
- no LastPass write/edit/delete support
- no `lpass export`
- no password or note-body dumping during discovery
- no automatic approval or auto-apply flow
- no TUI or GUI
- no OpenClaw config mutation command guessing
- no deletion/pruning of existing resolver mappings during apply

## 3. Safety decisions made

- Discovery uses metadata only and never requests passwords or note bodies
- Discovery does not use `lpass export`
- Suggested mappings are written as a draft for human review, not applied automatically
- Apply only writes approved, non-disabled plan entries
- Apply merges approved entries into the existing resolver mapping instead of deleting unrelated mappings
- Apply does not fetch secret values unless `--validate` is explicitly requested
- When validation is requested, resolved values are discarded immediately and never persisted
- Output files are written with `0600` permissions

## 4. Exact commands to test locally

These are the exact commands used here for code-level verification:

```bash
./.tools/go/bin/gofmt -w $(rg --files -g '*.go')
./.tools/go/bin/go test ./...
./.tools/go/bin/go build -o ./dist/openclaw-lastpass ./cmd/openclaw-lastpass
```

Recommended local integration checks on a machine with `lpass` installed and logged in:

```bash
./dist/openclaw-lastpass discover
cat ~/.config/openclaw-lastpass/mapping.draft.json
./dist/openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json --dry-run
./dist/openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json
./dist/openclaw-lastpass list
```

Optional validation and OpenClaw snippet output:

```bash
./dist/openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json --validate --dry-run
./dist/openclaw-lastpass apply --plan ~/.config/openclaw-lastpass/mapping.draft.json --print-openclaw-config --dry-run
```

## 5. Whether any assumptions were made about `lpass`

Yes:

- `lpass ls --format` is available and returns safe metadata fields such as ID, name, fullname, group, URL, and username
- LastPass unique IDs returned by `lpass ls` are valid resolver targets for later `lpass show` calls
- Username is metadata and may be included when it passes the local safety filter, but discovery heuristics do not depend on it
- `lpass` remains the source of truth for authentication and local session state

## 6. Possible next steps

- add more heuristic patterns for common providers and infrastructure tools
- add optional JSON output for `discover` and `apply`
- add an explicit prune mode for removing mappings that were previously applied but are no longer approved
- add richer metadata filtering controls for discovery
- add integration tests against a controlled fake `lpass` fixture script
