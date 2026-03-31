#!/usr/bin/env bash
set -euo pipefail

BIN_NAME="openclaw-lastpass"
REPO="${REPO:-chikuma0/openclaw-lastpass}"
VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-}"

info() {
  printf '%s\n' "$*"
}

warn() {
  printf 'warning: %s\n' "$*" >&2
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

normalize_platform() {
  local uname_os uname_arch os arch

  uname_os="$(uname -s)"
  uname_arch="$(uname -m)"

  case "$uname_os" in
    Linux) os="linux" ;;
    Darwin) os="darwin" ;;
    *) die "unsupported OS: $uname_os. Supported platforms are Linux and macOS." ;;
  esac

  case "$uname_arch" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) die "unsupported architecture: $uname_arch. Supported architectures are amd64 and arm64." ;;
  esac

  printf '%s %s\n' "$os" "$arch"
}

choose_install_dir() {
  if [ -n "$INSTALL_DIR" ]; then
    printf '%s\n' "$INSTALL_DIR"
    return
  fi

  if [ -d /usr/local/bin ] && [ -w /usr/local/bin ]; then
    printf '/usr/local/bin\n'
    return
  fi

  if [ -d /usr/local ] && [ -w /usr/local ]; then
    printf '/usr/local/bin\n'
    return
  fi

  printf '%s/.local/bin\n' "$HOME"
}

release_url() {
  local version="$1"
  local asset="$2"
  printf 'https://github.com/%s/releases/download/%s/%s\n' "$REPO" "$version" "$asset"
}

resolve_version() {
  if [ "$VERSION" != "latest" ]; then
    printf '%s\n' "$VERSION"
    return
  fi

  local api_url="https://api.github.com/repos/${REPO}/releases/latest"
  local tag

  tag="$(curl -fsSL "$api_url" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)"
  [ -n "$tag" ] || die "unable to determine the latest release tag from ${api_url}. Try again later or rerun with VERSION=v0.0.1."
  printf '%s\n' "$tag"
}

require_cmd curl
require_cmd tar
require_cmd mktemp
require_cmd uname

read -r OS ARCH <<EOF
$(normalize_platform)
EOF

TARGET_DIR="$(choose_install_dir)"
RESOLVED_VERSION="$(resolve_version)"
ASSET="${BIN_NAME}_${RESOLVED_VERSION}_${OS}_${ARCH}.tar.gz"
URL="$(release_url "$RESOLVED_VERSION" "$ASSET")"
TMP_DIR="$(mktemp -d)"
ARCHIVE_PATH="${TMP_DIR}/${ASSET}"

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

info "Downloading ${ASSET} from ${URL}"
if ! curl -fsSL "$URL" -o "$ARCHIVE_PATH"; then
  die "download failed for ${ASSET}. Verify that the GitHub Release exists and includes that asset."
fi

if ! tar -tzf "$ARCHIVE_PATH" >/dev/null 2>&1; then
  die "downloaded file is not a valid tar.gz archive"
fi

BINARY_RELATIVE_PATH="$(tar -tzf "$ARCHIVE_PATH" | awk '$0 ~ /(^|\/)openclaw-lastpass$/ {print $0}' | head -n1)"
[ -n "$BINARY_RELATIVE_PATH" ] || die "release archive does not contain ${BIN_NAME}"

MATCH_COUNT="$(tar -tzf "$ARCHIVE_PATH" | awk '$0 ~ /(^|\/)openclaw-lastpass$/ {count++} END {print count+0}')"
[ "$MATCH_COUNT" -eq 1 ] || die "release archive contains an unexpected binary layout for ${BIN_NAME}"

mkdir -p "$TARGET_DIR"
[ -w "$TARGET_DIR" ] || die "install directory is not writable: $TARGET_DIR"

tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR" "$BINARY_RELATIVE_PATH"
cp "$TMP_DIR/$BINARY_RELATIVE_PATH" "$TARGET_DIR/$BIN_NAME"
chmod 0755 "$TARGET_DIR/$BIN_NAME"

info
info "Installed ${BIN_NAME} to ${TARGET_DIR}/${BIN_NAME}"

case ":$PATH:" in
  *":$TARGET_DIR:"*) ;;
  *)
    warn "${TARGET_DIR} is not currently on PATH."
    warn "Add it with: export PATH=\"${TARGET_DIR}:\$PATH\""
    ;;
esac

if command -v lpass >/dev/null 2>&1; then
  info "LastPass CLI detected at $(command -v lpass)"
else
  warn "LastPass CLI is not installed."
  warn "Install lpass from your package manager or from https://github.com/lastpass/lastpass-cli"
fi

info
info "Next steps:"
info "  1. lpass login you@example.com"
info "  2. ${BIN_NAME} init"
