#!/usr/bin/env bash
set -Eeuo pipefail

repo_root="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root" || exit 1

if ! command -v git >/dev/null 2>&1; then
  printf 'error: git is required\n' >&2
  exit 127
fi

if ! command -v codex >/dev/null 2>&1; then
  printf 'error: codex CLI is required\n' >&2
  exit 127
fi

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  printf 'error: not inside a git work tree\n' >&2
  exit 2
fi

if [ -z "$(git status --porcelain --untracked-files=all)" ]; then
  printf 'No uncommitted changes to review.\n'
  exit 0
fi

review_dir="$(mktemp -d "${TMPDIR:-/tmp}/codex-review-sequence.XXXXXX")" || exit 1

# Defaults can be overridden per repo/user/session:
#   CODEX_REVIEW_MODEL=gpt-5.5
#   CODEX_REVIEW_EFFORT=xhigh
#   CODEX_REVIEW_EXTRA=security,test-gaps,api-compat,concurrency
CODEX_REVIEW_MODEL="${CODEX_REVIEW_MODEL:-gpt-5.5}"
CODEX_REVIEW_EFFORT="${CODEX_REVIEW_EFFORT:-xhigh}"
CODEX_REVIEW_EXTRA="${CODEX_REVIEW_EXTRA:-}"

CODEX_REVIEW_EXTRA="${CODEX_REVIEW_EXTRA//[[:space:]]/}"

IFS=',' read -r -a requested_extras <<<"$CODEX_REVIEW_EXTRA"
for extra in "${requested_extras[@]}"; do
  case "$extra" in
  "" | security | test-gaps | api-compat | concurrency) ;;
  *)
    printf 'error: unknown CODEX_REVIEW_EXTRA item=%s\n' "$extra" >&2
    printf 'valid values: security, test-gaps, api-compat, concurrency\n' >&2
    exit 2
    ;;
  esac
done

has_extra() {
  needle="$1"
  case ",$CODEX_REVIEW_EXTRA," in
  *",$needle,"*) return 0 ;;
  *) return 1 ;;
  esac
}

run_codex() {
  codex --search \
    -m "$CODEX_REVIEW_MODEL" \
    -c "model_reasoning_effort=\"$CODEX_REVIEW_EFFORT\"" \
    "$@"
}

pids=()
names=()
logs=()
msgs=()

start_builtin_review() {
  name="$1"
  log="$review_dir/${name}.log"

  (
    run_codex review --uncommitted
  ) >"$log" 2>&1 &

  pids+=("$!")
  names+=("$name")
  logs+=("$log")
  msgs+=("")
}

start_focused_review() {
  name="$1"
  prompt="$2"

  log="$review_dir/${name}.log"
  msg="$review_dir/${name}.md"

  (
    run_codex \
      exec \
      --sandbox read-only \
      --output-last-message "$msg" \
      "$prompt"
  ) >"$log" 2>&1 &

  pids+=("$!")
  names+=("$name")
  logs+=("$log")
  msgs+=("$msg")
}

complexity_prompt='Review the current uncommitted changes. Focus on avoidable complexity: Rule of Three, YAGNI, and one-liners. Report only actionable simplifications with file:line references and why the simpler alternative preserves behavior. If there are none, say exactly: No actionable avoidable-complexity findings.'
test_gaps_prompt='Review the current uncommitted changes. Focus on missing, weak, or misleading validation for changed behavior, bug fixes, migrations, and compatibility-sensitive changes. Report only actionable test gaps with file:line references and the exact behavior that should be tested. If there are none, say exactly: No actionable test-gap findings.'
security_prompt='Review the current uncommitted changes. Focus on concrete security regressions introduced or exposed by this diff: authn/authz, unsafe filesystem/shell/network/browser/URL handling, injection, path traversal, secret exposure, unsafe deserialization, privilege boundaries, and dependency/config weakening. Report only actionable findings with file:line references, impact, and the smallest safe fix. If there are none, say exactly: No actionable security findings.'
api_compat_prompt='Review the current uncommitted changes. Focus on API, CLI, config/env, schema, migration, generated-client, docs-contract, rollout, and rollback compatibility regressions. Report only actionable risks with file:line references, the expected failure mode, and the smallest safe fix. If there are none, say exactly: No actionable API/migration compatibility findings.'
concurrency_prompt='Review the current uncommitted changes. Focus on concurrency, lifecycle, and operational correctness: races, async ordering, cancellation/cleanup, leaks, retry idempotency, transactions, stale cache/state, timing assumptions, and unsafe parallelism. Report only actionable findings with file:line references, the runtime scenario, and the smallest safe fix. If there are none, say exactly: No actionable concurrency/lifecycle findings.'

printf 'Review output: %s\n' "$review_dir"
printf '\nChanged files:\n'
git status --short --untracked-files=all

# Standard review sequence: keep this cheap enough to run for every work item.
start_builtin_review "correctness-review"
start_focused_review "avoidable-complexity-review" "$complexity_prompt"

# Optional focused reviewers:
#   CODEX_REVIEW_EXTRA=security,test-gaps scripts/codex-review-sequence.sh
if has_extra "test-gaps"; then
  start_focused_review "test-gap-review" "$test_gaps_prompt"
fi

if has_extra "security"; then
  start_focused_review "security-review" "$security_prompt"
fi

if has_extra "api-compat"; then
  start_focused_review "api-compat-review" "$api_compat_prompt"
fi

if has_extra "concurrency"; then
  start_focused_review "concurrency-review" "$concurrency_prompt"
fi

statuses=()
failed=0

for i in "${!pids[@]}"; do
  if wait "${pids[$i]}"; then
    statuses[$i]=0
  else
    statuses[$i]=$?
    failed=1
  fi
done

for i in "${!names[@]}"; do
  name="${names[$i]}"
  log="${logs[$i]}"
  msg="${msgs[$i]}"

  printf '\n--- %s ---\n' "$name"
  if [ -n "$msg" ] && [ -s "$msg" ]; then
    cat "$msg"
    printf '\n'
  else
    cat "$log"
  fi

  if [ "${statuses[$i]}" -ne 0 ]; then
    printf '[reviewer exited with status %s; full log: %s]\n' \
      "${statuses[$i]}" "$log"
  fi
done

if [ "$failed" -ne 0 ]; then
  printf '\nReview command failed:\n' >&2
  for i in "${!names[@]}"; do
    if [ "${statuses[$i]}" -ne 0 ]; then
      printf '  %s=%s log=%s\n' "${names[$i]}" "${statuses[$i]}" "${logs[$i]}" >&2
    fi
  done
  exit 1
fi

printf '\nReview sequence complete. Address actionable findings before committing.\n'
