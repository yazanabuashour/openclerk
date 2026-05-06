#!/usr/bin/env sh
set -eu

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$repo_root"

count=0
for skill_dir in skills/openclerk modules/*/skill/*; do
  if [ -f "${skill_dir}/SKILL.md" ]; then
    ./scripts/validate-agent-skill.sh "$skill_dir"
    count=$((count + 1))
  fi
done

if [ "$count" -lt 2 ]; then
  echo "expected core skill plus at least one module skill" >&2
  exit 1
fi

printf 'validated %s agent skills\n' "$count"
