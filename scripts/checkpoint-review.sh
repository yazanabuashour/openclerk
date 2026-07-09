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

# Defaults can be overridden per repo/user/session.
#   REVIEW_MODEL=gpt-5.5
#   REVIEW_EFFORT=xhigh
#   REVIEW_EXTRA=security,test-gaps,api-compat,concurrency,policy
#   REVIEW_CUSTOM_PROMPT_FILE=path/to/review-prompt.md
REVIEW_MODEL="${REVIEW_MODEL:-gpt-5.5}"
REVIEW_EFFORT="${REVIEW_EFFORT:-xhigh}"
REVIEW_EXTRA="${REVIEW_EXTRA:-}"
REVIEW_CUSTOM_PROMPT="${REVIEW_CUSTOM_PROMPT:-}"
REVIEW_CUSTOM_PROMPT_FILE="${REVIEW_CUSTOM_PROMPT_FILE:-}"
REVIEW_CUSTOM_PROMPT_FILES="${REVIEW_CUSTOM_PROMPT_FILES:-}"

REVIEW_EXTRA="${REVIEW_EXTRA//[[:space:]]/}"

IFS=',' read -r -a requested_extras <<<"$REVIEW_EXTRA"
for extra in "${requested_extras[@]}"; do
  case "$extra" in
  "" | security | test-gaps | api-compat | concurrency | policy) ;;
  *)
    printf 'error: unknown review extra item=%s\n' "$extra" >&2
    printf 'valid values: security, test-gaps, api-compat, concurrency, policy\n' >&2
    exit 2
    ;;
  esac
done

has_extra() {
  needle="$1"
  case ",$REVIEW_EXTRA," in
  *",$needle,"*) return 0 ;;
  *) return 1 ;;
  esac
}

run_codex() {
  codex --search \
    -m "$REVIEW_MODEL" \
    -c "model_reasoning_effort=\"$REVIEW_EFFORT\"" \
    "$@"
}

git_status() {
  git status "$@" --untracked-files=all
}

git_status_short() {
  git_status --short
}

git_status_porcelain() {
  git_status --porcelain=v1
}

if [ -z "$(git_status_porcelain)" ]; then
  printf 'No uncommitted changes; no review checkpoint was run.\n'
  exit 0
fi

review_dir="$(mktemp -d "${TMPDIR:-/tmp}/checkpoint-review.XXXXXX")" || exit 1
summary_file="$review_dir/summary.txt"

snapshot_state() {
  {
    git_status_porcelain
    git diff --cached --no-ext-diff --
    git diff --no-ext-diff --
    git ls-files --others --exclude-standard -z |
      while IFS= read -r -d '' path; do
        printf 'untracked:%s\n' "$path"
        if [ -f "$path" ]; then
          git hash-object -- "$path"
        fi
      done
  } | git hash-object --stdin
}

pids=()
names=()
logs=()
msgs=()

cleanup_children() {
  for pid in "${pids[@]}"; do
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
    fi
  done
  wait "${pids[@]}" 2>/dev/null || true
}

cleanup_and_mark() {
  status=$?
  if [ "$#" -gt 0 ]; then
    status="$1"
  fi
  trap - INT TERM HUP EXIT
  cleanup_children
  exit "$status"
}

trap cleanup_and_mark EXIT
trap 'cleanup_and_mark 130' INT
trap 'cleanup_and_mark 143' TERM
trap 'cleanup_and_mark 129' HUP

start_builtin_review() {
  name="$1"
  log="$review_dir/${name}.log"

  run_codex --sandbox read-only review --uncommitted >"$log" 2>&1 &

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

  run_codex \
    exec \
    --sandbox read-only \
    --output-last-message "$msg" \
    "$prompt" >"$log" 2>&1 &

  pids+=("$!")
  names+=("$name")
  logs+=("$log")
  msgs+=("$msg")
}

start_custom_prompt_review() {
  name="$1"
  prompt="$2"

  if [ -z "${prompt//[[:space:]]/}" ]; then
    printf 'error: custom review prompt is empty for %s\n' "$name" >&2
    exit 2
  fi

  start_focused_review "$name" "$prompt"
}

review_prefix='Review the current uncommitted changes. Do not edit files.'
complexity_prompt="$review_prefix Focus on avoidable complexity: Rule of Three, YAGNI, and one-liners. Report only actionable simplifications with file:line references and why the simpler alternative preserves behavior. If there are none, say exactly: No actionable avoidable-complexity findings."
test_gaps_prompt="$review_prefix Focus on missing, weak, or misleading validation for changed behavior, bug fixes, migrations, and compatibility-sensitive changes. Report only actionable test gaps with file:line references and the exact behavior that should be tested. If there are none, say exactly: No actionable test-gap findings."
security_prompt="$review_prefix Focus on concrete security regressions introduced or exposed by this diff: authn/authz, unsafe filesystem/shell/network/browser/URL handling, injection, path traversal, secret exposure, unsafe deserialization, privilege boundaries, and dependency/config weakening. Report only actionable findings with file:line references, impact, and the smallest safe fix. If there are none, say exactly: No actionable security findings."
api_compat_prompt="$review_prefix Focus on API, CLI, config/env, schema, migration, generated-client, docs-contract, rollout, and rollback compatibility regressions. Report only actionable risks with file:line references, the expected failure mode, and the smallest safe fix. If there are none, say exactly: No actionable API/migration compatibility findings."
concurrency_prompt="$review_prefix Focus on concurrency, lifecycle, and operational correctness: races, async ordering, cancellation/cleanup, leaks, retry idempotency, transactions, stale cache/state, timing assumptions, and unsafe parallelism. Report only actionable findings with file:line references, the runtime scenario, and the smallest safe fix. If there are none, say exactly: No actionable concurrency/lifecycle findings."
policy_prompt="$review_prefix Focus on orchestration-policy quality: ambiguous delegation rules, over-orchestration risk, under-orchestration risk, thread/worktree/subagent/goal sequencing contradictions, review/commit contract contradictions, and coding-agent portability. Report only actionable findings with file:line references and the smallest wording or script change that resolves the issue. If there are none, say exactly: No actionable orchestration-policy findings."

if [ -n "$REVIEW_CUSTOM_PROMPT" ] && [ -z "${REVIEW_CUSTOM_PROMPT//[[:space:]]/}" ]; then
  printf 'error: custom review prompt is empty\n' >&2
  exit 2
fi

custom_prompt_files="$REVIEW_CUSTOM_PROMPT_FILES"
if [ -n "$REVIEW_CUSTOM_PROMPT_FILE" ]; then
  if [ -n "$custom_prompt_files" ]; then
    custom_prompt_files="$REVIEW_CUSTOM_PROMPT_FILE,$custom_prompt_files"
  else
    custom_prompt_files="$REVIEW_CUSTOM_PROMPT_FILE"
  fi
fi

custom_prompt_file_list=()
if [ -n "$custom_prompt_files" ]; then
  IFS=',' read -r -a custom_prompt_file_list <<<"$custom_prompt_files"
  for prompt_file in "${custom_prompt_file_list[@]}"; do
    if [ -z "$prompt_file" ]; then
      continue
    fi
    if [ ! -f "$prompt_file" ]; then
      printf 'error: custom review prompt file not found: %s\n' "$prompt_file" >&2
      exit 2
    fi
    if [ ! -r "$prompt_file" ]; then
      printf 'error: custom review prompt file is not readable: %s\n' "$prompt_file" >&2
      exit 2
    fi
    if ! grep -q '[^[:space:]]' "$prompt_file"; then
      printf 'error: custom review prompt file is empty: %s\n' "$prompt_file" >&2
      exit 2
    fi
  done
fi

printf 'Review output: %s\n' "$review_dir"
printf '\nChanged files:\n'
git_status_short
git_status_short >"$review_dir/changed-files.txt"
git diff HEAD --stat >"$review_dir/diff-stat.txt"
review_state_hash="$(snapshot_state)"

# Standard checkpoint review: keep this cheap enough to run for every work item.
# These are independent Codex CLI review processes, not interactive subagent
# threads, so they work in non-interactive checkpoint scripts.
start_builtin_review "correctness-review"
start_focused_review "avoidable-complexity-review" "$complexity_prompt"

# Optional focused reviewers:
#   REVIEW_EXTRA=security,test-gaps scripts/checkpoint-review.sh
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

if has_extra "policy"; then
  start_focused_review "policy-review" "$policy_prompt"
fi

if [ -n "$REVIEW_CUSTOM_PROMPT" ]; then
  start_custom_prompt_review "custom-review" "$REVIEW_CUSTOM_PROMPT"
fi

if [ "${#custom_prompt_file_list[@]}" -gt 0 ]; then
  custom_index=1
  for prompt_file in "${custom_prompt_file_list[@]}"; do
    if [ -z "$prompt_file" ]; then
      continue
    fi
    start_custom_prompt_review "custom-review-$custom_index" "$(cat "$prompt_file")"
    custom_index=$((custom_index + 1))
  done
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
trap - INT TERM HUP EXIT

write_summary() {
  {
    printf 'Review output: %s\n' "$review_dir"
    printf 'Changed files:\n'
    cat "$review_dir/changed-files.txt"
    printf '\nReviewers:\n'
    for i in "${!names[@]}"; do
      printf '%s status=%s log=%s' "${names[$i]}" "${statuses[$i]}" "${logs[$i]}"
      if [ -n "${msgs[$i]}" ]; then
        printf ' message=%s' "${msgs[$i]}"
      fi
      printf '\n'
    done
  } >"$summary_file"
}

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
  write_summary
  printf '\nReview command failed:\n' >&2
  for i in "${!names[@]}"; do
    if [ "${statuses[$i]}" -ne 0 ]; then
      printf '  %s=%s log=%s\n' "${names[$i]}" "${statuses[$i]}" "${logs[$i]}" >&2
    fi
  done
  exit 1
fi

if [ "$(snapshot_state)" != "$review_state_hash" ]; then
  write_summary
  printf '\nReview command failed: worktree changed while reviewers were running.\n' >&2
  printf 'Rerun the checkpoint review for the current diff before committing.\n' >&2
  exit 1
fi

write_summary

printf '\nCheckpoint review complete. Address actionable findings before committing.\n'
printf 'Review summary: %s\n' "$summary_file"
