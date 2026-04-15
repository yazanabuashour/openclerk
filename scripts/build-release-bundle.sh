#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "usage: $0 <version-label> [out-dir]" >&2
  exit 2
fi

version="$1"
out_dir="${2:-dist}"
archive="${out_dir}/openclerk-${version}.tar.gz"
sbom="${out_dir}/openclerk-${version}.sbom.json"
checksum="${archive}.sha256"

mkdir -p "${out_dir}"

git archive \
  --format=tar.gz \
  --prefix="openclerk-${version}/" \
  HEAD \
  -o "${archive}"

go run github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.9.0 \
  mod \
  -json \
  -type library \
  -output "${sbom}" \
  .

(
  cd "${out_dir}"
  shasum -a 256 "$(basename "${archive}")" > "$(basename "${checksum}")"
)

printf '%s\n' "${archive}" "${sbom}" "${checksum}"
