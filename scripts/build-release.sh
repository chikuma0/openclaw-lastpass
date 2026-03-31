#!/usr/bin/env bash
set -euo pipefail

export LC_ALL=C
export LANG=C

BIN_NAME="openclaw-lastpass"
GO="${GO:-go}"
VERSION="${1:-${GITHUB_REF_NAME:-}}"
OUT_DIR="${2:-dist/release}"

if [ -z "$VERSION" ]; then
  printf 'usage: %s <version> [out-dir]\n' "$0" >&2
  exit 1
fi

mkdir -p "$OUT_DIR"

platforms=(
  "linux amd64"
  "linux arm64"
  "darwin amd64"
  "darwin arm64"
)

cleanup_dirs=()
cleanup() {
  for dir in "${cleanup_dirs[@]}"; do
    rm -rf "$dir"
  done
}
trap cleanup EXIT

for platform in "${platforms[@]}"; do
  read -r goos goarch <<<"$platform"
  stage_dir="$(mktemp -d)"
  cleanup_dirs+=("$stage_dir")

  asset="${BIN_NAME}_${VERSION}_${goos}_${goarch}.tar.gz"
  printf 'building %s\n' "$asset"

  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" "$GO" build -trimpath -ldflags='-s -w' -o "${stage_dir}/${BIN_NAME}" ./cmd/openclaw-lastpass
  tar -C "$stage_dir" -czf "${OUT_DIR}/${asset}" "${BIN_NAME}"
done

if command -v sha256sum >/dev/null 2>&1; then
  (
    cd "$OUT_DIR"
    sha256sum ./*.tar.gz > SHA256SUMS
  )
elif command -v shasum >/dev/null 2>&1; then
  (
    cd "$OUT_DIR"
    shasum -a 256 ./*.tar.gz > SHA256SUMS
  )
fi
