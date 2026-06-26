#!/bin/sh
set -eu

repo="yazanabuashour/openclerk"
default_module="__OPENCLERK_MODULE__"
default_version="__OPENCLERK_MODULE_VERSION__"

log() {
  printf '%s\n' "$*"
}

fail() {
  printf 'openclerk module install: %s\n' "$*" >&2
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

select_module() {
  module="${OPENCLERK_MODULE:-$default_module}"
  case "$module" in
    "" | "__OPENCLERK_MODULE__") fail "OPENCLERK_MODULE is required" ;;
    ollama-embeddings | gemini-embeddings | tesseract-ocr) printf '%s' "$module" ;;
    *) fail "unsupported OPENCLERK_MODULE: $module" ;;
  esac
}

url_tag() {
  printf '%s%%2F%s' "$1" "$2"
}

resolve_latest_module_version() {
  module="$1"
  releases_json="$(curl -fsSL "https://api.github.com/repos/${repo}/releases?per_page=100")" ||
    fail "could not resolve module releases"
  tag="$(printf '%s\n' "$releases_json" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | grep "^${module}/v" | head -n 1 || true)"
  [ -n "$tag" ] || fail "could not find latest release for ${module}"
  printf '%s' "${tag#${module}/}"
}

select_version() {
  module="$1"
  requested="${OPENCLERK_MODULE_VERSION:-$default_version}"
  if [ "$requested" = "" ] || [ "$requested" = "__OPENCLERK_MODULE_VERSION__" ] || [ "$requested" = "latest" ]; then
    resolve_latest_module_version "$module"
    return
  fi
  case "$requested" in
    ${module}/v*) printf '%s' "${requested#${module}/}" ;;
    v*) printf '%s' "$requested" ;;
    *) printf 'v%s' "$requested" ;;
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

verify_attestation() {
	asset="$1"
	gh attestation verify "$asset" --repo "$repo" >/dev/null ||
		fail "attestation verification failed for ${asset}"
}

validate_archive_members() {
	archive="$1"
	root="$2"
	member_list="members-${archive}.txt"
	tar -tzf "$archive" > "$member_list" ||
		fail "archive listing failed for ${archive}"
	while IFS= read -r member; do
		[ -n "$member" ] || fail "archive contains an empty member name"
		case "$member" in
			/* | ../* | */../* | */.. | *\\*)
				fail "archive member escapes extraction root: ${member}"
				;;
		esac
		case "$member" in
			"$root" | "$root/" | "$root"/*)
				;;
			*)
				fail "archive member outside expected root ${root}: ${member}"
				;;
		esac
	done < "$member_list"
}

reject_extracted_symlinks() {
	root="$1"
	[ -d "$root" ] || fail "archive root missing after extraction: ${root}"
	symlink="$(find "$root" -type l -print | sed -n '1p')"
	[ -z "$symlink" ] || fail "archive contains symlink member: ${symlink}"
}

select_install_dir() {
  if [ -n "${OPENCLERK_INSTALL_DIR:-}" ]; then
    printf '%s' "$OPENCLERK_INSTALL_DIR"
    return
  fi
  [ -n "${HOME:-}" ] || fail "HOME is not set and OPENCLERK_INSTALL_DIR was not provided"
  printf '%s/.local/bin' "$HOME"
}

select_module_root() {
  if [ -n "${OPENCLERK_MODULE_DIR:-}" ]; then
    printf '%s' "$OPENCLERK_MODULE_DIR"
    return
  fi
  if [ -n "${XDG_DATA_HOME:-}" ]; then
    printf '%s/openclerk/modules' "$XDG_DATA_HOME"
    return
  fi
  [ -n "${HOME:-}" ] || fail "HOME is not set and OPENCLERK_MODULE_DIR was not provided"
  printf '%s/.local/share/openclerk/modules' "$HOME"
}

provider_for_module() {
  case "$1" in
    ollama-embeddings) printf 'ollama' ;;
    gemini-embeddings) printf 'gemini' ;;
    tesseract-ocr) printf 'tesseract' ;;
  esac
}

command_for_module() {
  case "$1" in
    ollama-embeddings | gemini-embeddings) printf 'semantic-retrieval-adapter' ;;
    tesseract-ocr) printf 'tesseract' ;;
  esac
}

registration_json() {
  module="$1"
  manifest_path="$2"
  command_name="$3"
  case "$module" in
    ollama-embeddings)
      printf '{"action":"install_module","module":{"provider":"ollama","manifest_path":"%s","command":"%s","provider_config":{"embedding_model":"embeddinggemma","ollama_url":"http://localhost:11434"}}}' "$manifest_path" "$command_name"
      ;;
    gemini-embeddings)
      printf '{"action":"install_module","module":{"provider":"gemini","manifest_path":"%s","command":"%s","provider_config":{"embedding_model":"gemini-embedding-001","gemini_api_base":"https://generativelanguage.googleapis.com/v1beta","embedding_output_dimensions":"3072"}}}' "$manifest_path" "$command_name"
      ;;
    tesseract-ocr)
      printf '{"action":"install_module","module":{"kind":"ocr_provider","provider":"tesseract","manifest_path":"%s","command":"%s","provider_config":{"ocrmypdf_command":"ocrmypdf","language":"eng"}}}' "$manifest_path" "$command_name"
      ;;
  esac
}

need_cmd curl
need_cmd gh
need_cmd find
need_cmd tar

module="$(select_module)"
version="$(select_version "$module")"
asset_version="${version#v}"
os="$(detect_os)"
arch="$(detect_arch)"
archive="openclerk-module-${module}_${asset_version}_${os}_${arch}.tar.gz"
archive_root="openclerk-module-${module}_${asset_version}_${os}_${arch}"
checksum="openclerk-module-${module}_${asset_version}_checksums.txt"
release_url="https://github.com/${repo}/releases/download/$(url_tag "$module" "$version")"
tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/openclerk-module-install.XXXXXX")"
install_dir="$(select_install_dir)"
module_root="$(select_module_root)"
provider="$(provider_for_module "$module")"
command_name="$(command_for_module "$module")"
registration_command="$command_name"

cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

log "Installing OpenClerk module ${module} ${version} for ${os}/${arch}"

cd "$tmp_dir"
download "${release_url}/${archive}" "$archive"
download "${release_url}/${checksum}" "$checksum"
verify_attestation "$archive"
verify_attestation "$checksum"
verify_archive "$checksum" "$archive"
validate_archive_members "$archive" "$archive_root"

tar -xzf "$archive"
reject_extracted_symlinks "$archive_root"
[ -d "${archive_root}/modules/${module}" ] && [ ! -L "${archive_root}/modules/${module}" ] ||
  fail "archive missing regular module directory"
[ -f "${archive_root}/modules/${module}/module.json" ] && [ ! -L "${archive_root}/modules/${module}/module.json" ] ||
  fail "archive missing regular module manifest"
mkdir -p "$module_root" "$install_dir"
rm -rf "$module_root/$module"
cp -R "${archive_root}/modules/${module}" "$module_root/${module}"

if [ -d "${archive_root}/bin" ]; then
  for file in "${archive_root}/bin"/*; do
    [ -e "$file" ] || continue
    [ -f "$file" ] && [ ! -L "$file" ] ||
      fail "archive bundled command is not a regular file: ${file}"
    cp "$file" "$install_dir/$(basename "$file")"
    chmod 755 "$install_dir/$(basename "$file")"
    if [ "$(basename "$file")" = "$command_name" ]; then
      registration_command="${install_dir}/${command_name}"
    fi
  done
fi

manifest_path="${module_root}/${module}/module.json"
skill_path="${module_root}/${module}/skill/${module}/SKILL.md"

log "Installed module files to ${module_root}/${module}"
log "Provider: ${provider}"
log "Manifest: ${manifest_path}"
log "Skill: ${skill_path}"
if [ -d "${archive_root}/bin" ]; then
  log "Installed bundled commands to ${install_dir}"
fi

log ""
log "Refresh module registration:"
printf "printf '%%s\\n' '%s' | openclerk module\n" "$(registration_json "$module" "$manifest_path" "$registration_command")"
log ""
log "Then verify:"
log "printf '%s\n' '{\"action\":\"list_modules\"}' | openclerk module"
