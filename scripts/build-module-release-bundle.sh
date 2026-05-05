#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 || $# -gt 3 ]]; then
  echo "usage: $0 <module-name> <version-label> [out-dir]" >&2
  exit 2
fi

module="$1"
version="$2"
asset_version="${version#v}"
out_dir="${3:-dist}"
repo_prefix="openclerk-module-${module}_${asset_version}"
checksum_file="${repo_prefix}_checksums.txt"
source_archive="${repo_prefix}_source.tar.gz"
skill_archive="${repo_prefix}_skill.tar.gz"
sbom_file="${repo_prefix}_sbom.json"

case "${module}" in
  ollama-embeddings)
    provider="ollama"
    command="semantic-retrieval-adapter"
    module_dir="modules/ollama-embeddings"
    skill_dir="modules/ollama-embeddings/skill/ollama-embeddings"
    targets=("darwin/arm64" "darwin/amd64" "linux/amd64" "linux/arm64")
    ;;
  gemini-embeddings)
    provider="gemini"
    command="semantic-retrieval-adapter"
    module_dir="modules/gemini-embeddings"
    skill_dir="modules/gemini-embeddings/skill/gemini-embeddings"
    targets=("darwin/arm64" "darwin/amd64" "linux/amd64" "linux/arm64")
    ;;
  tesseract-ocr)
    provider="tesseract"
    command="tesseract"
    module_dir="modules/tesseract-ocr"
    skill_dir="modules/tesseract-ocr/skill/tesseract-ocr"
    targets=("darwin/arm64" "darwin/amd64" "linux/amd64" "linux/arm64")
    ;;
  *)
    echo "unsupported module: ${module}" >&2
    exit 2
    ;;
esac

mkdir -p "${out_dir}"

copy_module_files() {
  local dest="$1"
  mkdir -p "${dest}/${module_dir}" "${dest}/${skill_dir}"
  cp "${module_dir}/module.json" "${dest}/${module_dir}/module.json"
  cp "${skill_dir}/SKILL.md" "${dest}/${skill_dir}/SKILL.md"
}

for target in "${targets[@]}"; do
  IFS=/ read -r os arch <<< "${target}"
  name="${repo_prefix}_${os}_${arch}"
  root="${out_dir}/${name}"
  copy_module_files "${root}"
  if [[ "${command}" == "semantic-retrieval-adapter" ]]; then
    mkdir -p "${root}/bin"
    GOOS="${os}" GOARCH="${arch}" go build -trimpath -ldflags="-s -w -X main.version=${version}" -o "${root}/bin/semantic-retrieval-adapter" ./modules/semantic-retrieval-adapter
  fi
  tar -C "${out_dir}" -czf "${out_dir}/${name}.tar.gz" "${name}"
  rm -rf "${root}"
done

mkdir -p "${out_dir}/skill/${module}"
cp "${skill_dir}/SKILL.md" "${out_dir}/skill/${module}/SKILL.md"
tar -C "${out_dir}/skill" -czf "${out_dir}/${skill_archive}" "${module}"
rm -rf "${out_dir:?}/skill"

git archive \
  --format=tar.gz \
  --prefix="${repo_prefix}/" \
  HEAD \
  -o "${out_dir}/${source_archive}"

go run github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.9.0 \
  mod \
  -json \
  -type library \
  -output "${out_dir}/${sbom_file}" \
  .

sed \
  -e "s/__OPENCLERK_MODULE__/${module}/g" \
  -e "s/__OPENCLERK_MODULE_VERSION__/${version}/g" \
  scripts/install-module.sh > "${out_dir}/install-module.sh"
chmod 755 "${out_dir}/install-module.sh"

(
  cd "${out_dir}"
  shasum -a 256 *.tar.gz "${sbom_file}" install-module.sh > "${checksum_file}"
)

printf '%s\n' "module=${module}" "provider=${provider}" "command=${command}" "${out_dir}"/*
