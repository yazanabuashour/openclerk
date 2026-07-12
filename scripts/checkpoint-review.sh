#!/usr/bin/env bash
set -Eeuo pipefail

repo_root="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root" || exit 1

if [ "${CHECKPOINT_REVIEW_ACTIVE:-}" = "1" ]; then
  printf 'Checkpoint review already active; recursive invocation skipped.\n'
  exit 0
fi
export CHECKPOINT_REVIEW_ACTIVE=1

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

REVIEW_MODEL="${REVIEW_MODEL:-gpt-5.6-sol}"

correctness_effort=""
complexity_effort=""
security_effort=""
test_gaps_effort=""
api_compat_effort=""
concurrency_effort=""
policy_effort=""
custom_effort=""
custom_prompt=""
requested_reviews=()

validate_effort() {
  case "$1" in
  none | minimal | low | medium | high | xhigh | max) ;;
  ultra)
    printf 'error: reasoning effort ultra is not allowed\n' >&2
    exit 2
    ;;
  *)
    printf 'error: unknown reasoning effort: %s\n' "$1" >&2
    printf 'valid values: none, minimal, low, medium, high, xhigh, max\n' >&2
    exit 2
    ;;
  esac
}

while [ "$#" -gt 0 ]; do
  case "$1" in
  --review)
    if [ "$#" -lt 3 ]; then
      printf 'error: --review requires TYPE and EFFORT\n' >&2
      exit 2
    fi
    review_type="$2"
    effort="$3"
    validate_effort "$effort"
    case "$review_type" in
    correctness)
      effort_var=correctness_effort
      ;;
    complexity)
      effort_var=complexity_effort
      ;;
    security)
      effort_var=security_effort
      ;;
    test-gaps)
      effort_var=test_gaps_effort
      ;;
    api-compat)
      effort_var=api_compat_effort
      ;;
    concurrency)
      effort_var=concurrency_effort
      ;;
    policy)
      effort_var=policy_effort
      ;;
    *)
      printf 'error: unknown review type: %s\n' "$review_type" >&2
      printf 'valid values: correctness, complexity, security, test-gaps, api-compat, concurrency, policy\n' >&2
      exit 2
      ;;
    esac
    if [ -n "${!effort_var}" ]; then
      printf 'error: duplicate review type: %s\n' "$review_type" >&2
      exit 2
    fi
    printf -v "$effort_var" '%s' "$effort"
    requested_reviews+=("$review_type=$effort")
    shift 3
    ;;
  --custom-review)
    if [ "$#" -lt 3 ]; then
      printf 'error: --custom-review requires EFFORT and PROMPT\n' >&2
      exit 2
    fi
    if [ -n "$custom_effort" ]; then
      printf 'error: --custom-review may be specified only once\n' >&2
      exit 2
    fi
    validate_effort "$2"
    if [ -z "${3//[[:space:]]/}" ]; then
      printf 'error: custom review prompt is empty\n' >&2
      exit 2
    fi
    custom_effort="$2"
    custom_prompt="$3"
    requested_reviews+=("custom-review=$custom_effort")
    shift 3
    ;;
  *)
    printf 'error: unknown argument: %s\n' "$1" >&2
    exit 2
    ;;
  esac
done

if [ -z "$correctness_effort" ] || [ -z "$complexity_effort" ]; then
  printf 'error: --review correctness EFFORT and --review complexity EFFORT are required\n' >&2
  exit 2
fi

run_codex() {
  effort="$1"
  shift
  codex --search \
    -m "$REVIEW_MODEL" \
    -c "model_reasoning_effort=\"$effort\"" \
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
  effort="$2"
  log="$review_dir/${name}.log"

  run_codex "$effort" --sandbox read-only review --uncommitted >"$log" 2>&1 &

  pids+=("$!")
  names+=("$name")
  logs+=("$log")
  msgs+=("")
}

start_focused_review() {
  name="$1"
  effort="$2"
  prompt="$3"

  log="$review_dir/${name}.log"
  msg="$review_dir/${name}.md"

  run_codex "$effort" \
    exec \
    --sandbox read-only \
    --output-last-message "$msg" \
    "$prompt" >"$log" 2>&1 &

  pids+=("$!")
  names+=("$name")
  logs+=("$log")
  msgs+=("$msg")
}

review_prefix='Review the current uncommitted changes. Do not edit files.'
complexity_prompt="$review_prefix Focus on avoidable complexity: Rule of Three, YAGNI, and one-liners. Report only actionable simplifications with file:line references and why the simpler alternative preserves behavior. If there are none, say exactly: No actionable avoidable-complexity findings."
test_gaps_prompt="$review_prefix Focus on missing, weak, or misleading validation for changed behavior, bug fixes, migrations, and compatibility-sensitive changes. Report only actionable test gaps with file:line references and the exact behavior that should be tested. If there are none, say exactly: No actionable test-gap findings."
security_prompt="$review_prefix Focus on concrete security regressions introduced or exposed by this diff: authn/authz, unsafe filesystem/shell/network/browser/URL handling, injection, path traversal, secret exposure, unsafe deserialization, privilege boundaries, and dependency/config weakening. Report only actionable findings with file:line references, impact, and the smallest safe fix. If there are none, say exactly: No actionable security findings."
api_compat_prompt="$review_prefix Focus on API, CLI, config/env, schema, migration, generated-client, docs-contract, rollout, and rollback compatibility regressions. Report only actionable risks with file:line references, the expected failure mode, and the smallest safe fix. If there are none, say exactly: No actionable API/migration compatibility findings."
concurrency_prompt="$review_prefix Focus on concurrency, lifecycle, and operational correctness: races, async ordering, cancellation/cleanup, leaks, retry idempotency, transactions, stale cache/state, timing assumptions, and unsafe parallelism. Report only actionable findings with file:line references, the runtime scenario, and the smallest safe fix. If there are none, say exactly: No actionable concurrency/lifecycle findings."
policy_prompt="$review_prefix Focus on orchestration-policy quality: ambiguous delegation rules, over-orchestration risk, under-orchestration risk, thread/worktree/subagent/goal sequencing contradictions, review/commit contract contradictions, and coding-agent portability. Report only actionable findings with file:line references and the smallest wording or script change that resolves the issue. If there are none, say exactly: No actionable orchestration-policy findings."

printf 'Requested reviews: %s\n' "${requested_reviews[*]}"
printf 'Review output: %s\n' "$review_dir"
printf '\nChanged files:\n'
git_status_short
git_status_short >"$review_dir/changed-files.txt"
git diff HEAD --stat >"$review_dir/diff-stat.txt"
review_state_hash="$(snapshot_state)"

# Standard checkpoint review: keep this cheap enough to run for every work item.
# These are independent Codex CLI review processes, not interactive subagent
# threads, so they work in non-interactive checkpoint scripts.
start_builtin_review "correctness-review" "$correctness_effort"
start_focused_review "avoidable-complexity-review" "$complexity_effort" "$complexity_prompt"

# Optional focused reviewers:
if [ -n "$test_gaps_effort" ]; then
  start_focused_review "test-gap-review" "$test_gaps_effort" "$test_gaps_prompt"
fi

if [ -n "$security_effort" ]; then
  start_focused_review "security-review" "$security_effort" "$security_prompt"
fi

if [ -n "$api_compat_effort" ]; then
  start_focused_review "api-compat-review" "$api_compat_effort" "$api_compat_prompt"
fi

if [ -n "$concurrency_effort" ]; then
  start_focused_review "concurrency-review" "$concurrency_effort" "$concurrency_prompt"
fi

if [ -n "$policy_effort" ]; then
  start_focused_review "policy-review" "$policy_effort" "$policy_prompt"
fi

if [ -n "$custom_effort" ]; then
  start_focused_review "custom-review" "$custom_effort" "$review_prefix $custom_prompt"
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
