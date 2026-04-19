#!/bin/sh
set -eu

repo="yazanabuashour/openclerk"
default_version="__OPENCLERK_VERSION__"

log() {
  printf '%s\n' "$*"
}

fail() {
  printf 'openclerk install: %s\n' "$*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

detect_os() {
  case "$(uname -s)" in
    Darwin) printf 'darwin' ;;
    Linux) printf 'linux' ;;
    *) fail "unsupported operating system: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64 | amd64) printf 'amd64' ;;
    arm64 | aarch64) printf 'arm64' ;;
    *) fail "unsupported CPU architecture: $(uname -m)" ;;
  esac
}

resolve_latest_version() {
  latest_json="$(curl -fsSL "https://api.github.com/repos/${repo}/releases/latest")" ||
    fail "could not resolve latest GitHub Release"
  latest_tag="$(printf '%s\n' "$latest_json" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
  [ -n "$latest_tag" ] || fail "could not read latest release tag"
  printf '%s' "$latest_tag"
}

select_version() {
  requested="${OPENCLERK_VERSION:-$default_version}"
  case "$requested" in
    "" | "__OPENCLERK_VERSION__" | latest)
      resolve_latest_version
      ;;
    v*)
      printf '%s' "$requested"
      ;;
    *)
      printf 'v%s' "$requested"
      ;;
  esac
}

download() {
  url="$1"
  output="$2"
  curl -fsSL "$url" -o "$output" || fail "download failed: $url"
}

verify_archive() {
  checksum_file="$1"
  archive="$2"
  expected_line="expected-${archive}.sha256"

  awk -v file="$archive" '$2 == file { print; found = 1 } END { exit found ? 0 : 1 }' "$checksum_file" > "$expected_line" ||
    fail "checksum entry not found for ${archive}"

  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 -c "$expected_line" >/dev/null ||
      fail "checksum verification failed for ${archive}"
    return
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum -c "$expected_line" >/dev/null ||
      fail "checksum verification failed for ${archive}"
    return
  fi

  fail "missing required command: shasum or sha256sum"
}

first_writable_path_dir() {
  old_ifs="$IFS"
  IFS=:
  for dir in ${PATH:-}; do
    IFS="$old_ifs"
    [ -n "$dir" ] || dir="."
    [ "$dir" = "." ] && continue
    if [ -d "$dir" ] && [ -w "$dir" ]; then
      printf '%s' "$dir"
      return 0
    fi
    IFS=:
  done
  IFS="$old_ifs"
  return 1
}

select_install_dir() {
  if [ -n "${OPENCLERK_INSTALL_DIR:-}" ]; then
    printf '%s' "$OPENCLERK_INSTALL_DIR"
    return
  fi

  if dir="$(first_writable_path_dir)"; then
    printf '%s' "$dir"
    return
  fi

  [ -n "${HOME:-}" ] || fail "HOME is not set and no writable PATH directory was found"
  printf '%s/.local/bin' "$HOME"
}

path_contains_dir() {
  needle="$1"
  old_ifs="$IFS"
  IFS=:
  for dir in ${PATH:-}; do
    IFS="$old_ifs"
    [ "$dir" = "$needle" ] && return 0
    IFS=:
  done
  IFS="$old_ifs"
  return 1
}

need_cmd curl
need_cmd tar

os="$(detect_os)"
arch="$(detect_arch)"
tag="$(select_version)"
asset_version="${tag#v}"
archive="openclerk_${asset_version}_${os}_${arch}.tar.gz"
checksum="openclerk_${asset_version}_checksums.txt"
release_url="https://github.com/${repo}/releases/download/${tag}"
tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/openclerk-install.XXXXXX")"
install_dir="$(select_install_dir)"

cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

log "Installing OpenClerk ${tag} for ${os}/${arch}"

cd "$tmp_dir"
download "${release_url}/${archive}" "$archive"
download "${release_url}/${checksum}" "$checksum"
verify_archive "$checksum" "$archive"

tar -xzf "$archive"
mkdir -p "$install_dir"
cp "openclerk_${asset_version}_${os}_${arch}/openclerk" "${install_dir}/openclerk"
chmod 755 "${install_dir}/openclerk"

log "Installed openclerk runner to ${install_dir}/openclerk"

if path_contains_dir "$install_dir"; then
  openclerk --help
else
  "${install_dir}/openclerk" --help
  log ""
  log "Add this directory to PATH before using the skill:"
  log "  export PATH=\"${install_dir}:\$PATH\""
fi

log ""
log "Install the Agent Skills-compatible skill with your agent's native skill installer or skill directory:"
log "  https://github.com/${repo}/tree/${tag}/skills/openclerk"
log ""
log "This installer does not choose an agent-specific skill location."
