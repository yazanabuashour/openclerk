#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "usage: $0 <version-label> [out-dir]" >&2
  exit 2
fi

version="$1"
asset_version="${version#v}"
out_dir="${2:-dist}"
source_archive="openclerk_${asset_version}_source.tar.gz"
skill_archive="openclerk_${asset_version}_skill.tar.gz"
checksum_file="openclerk_${asset_version}_checksums.txt"
sbom_file="openclerk_${asset_version}_sbom.json"

mkdir -p "${out_dir}"

targets=(
  "darwin/arm64"
  "darwin/amd64"
  "linux/amd64"
  "linux/arm64"
)

for target in "${targets[@]}"; do
  IFS=/ read -r os arch <<< "${target}"
  name="openclerk_${asset_version}_${os}_${arch}"
  mkdir -p "${out_dir}/${name}"
  GOOS="${os}" GOARCH="${arch}" go build -trimpath -ldflags="-s -w -X main.version=${version}" -o "${out_dir}/${name}/openclerk" ./cmd/openclerk
  tar -C "${out_dir}" -czf "${out_dir}/${name}.tar.gz" "${name}"
  rm -rf "${out_dir:?}/${name}"
done

mkdir -p "${out_dir}/skill/openclerk"
cp skills/openclerk/SKILL.md "${out_dir}/skill/openclerk/SKILL.md"
tar -C "${out_dir}/skill" -czf "${out_dir}/${skill_archive}" openclerk
rm -rf "${out_dir:?}/skill"

git archive \
  --format=tar.gz \
  --prefix="openclerk_${asset_version}/" \
  HEAD \
  -o "${out_dir}/${source_archive}"

go run github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.9.0 \
  mod \
  -json \
  -type library \
  -output "${out_dir}/${sbom_file}" \
  .

sed "s/__OPENCLERK_VERSION__/${version}/g" scripts/install.sh > "${out_dir}/install.sh"
chmod 755 "${out_dir}/install.sh"

(
  cd "${out_dir}"
  shasum -a 256 *.tar.gz "${sbom_file}" install.sh > "${checksum_file}"
)

printf '%s\n' "${out_dir}"/*
