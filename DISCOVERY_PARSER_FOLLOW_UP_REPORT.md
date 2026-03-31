# DISCOVERY_PARSER_FOLLOW_UP_REPORT

## Root Cause

`discover` and `init` were formatting `lpass ls` metadata as one line with six fields separated by the same control-character delimiter.

That works only if no entry metadata contains that delimiter byte. In the real Linux smoke test, at least one vault item included the separator inside its metadata, so a single record split into seven segments and discovery aborted before writing `discovery.json` or `mapping.draft.json`.

## Patch Approach

- kept the workflow metadata-only and still based on `lpass ls --format`
- changed the record format to use a different control-character separator for each field boundary
- replaced the `strings.Split` parser with a sequential `strings.Cut` parser
- added a line-preview parse error so failures still point to the exact record more usefully
- added regression tests for:
  - a delimiter collision inside entry metadata
  - malformed metadata lines still producing a helpful error

This keeps discovery narrow and read-only while making it much more resilient to delimiter collisions in real vault data.

## Exact Commands To Retest On Linux

From the repository:

```bash
./.tools/go/bin/go test ./internal/lastpass -run 'TestParseMetadataList|TestParseMetadataListAllowsDelimiterCollisionInsideMetadata|TestParseMetadataListRejectsMalformedLineWithPreview'
./.tools/go/bin/go test ./...
./.tools/go/bin/go build -o ./dist/openclaw-lastpass ./cmd/openclaw-lastpass
```

Then rerun the real smoke path:

```bash
./dist/openclaw-lastpass init --refresh
```

Optional direct discovery retest:

```bash
./dist/openclaw-lastpass discover --out-dir /tmp/openclaw-lastpass-smoke
ls -l /tmp/openclaw-lastpass-smoke
```

If you are retesting the installed binary instead of the repo build:

```bash
openclaw-lastpass init --refresh
```
