#!/usr/bin/env bash
set -uo pipefail

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$repo_root" || exit 1

review_dir="$(mktemp -d "${TMPDIR:-/tmp}/codex-review-sequence.XXXXXX")" || exit 1

correctness_log="$review_dir/correctness-review.log"
complexity_log="$review_dir/avoidable-complexity-review.log"
complexity_msg="$review_dir/avoidable-complexity-review.md"

complexity_prompt='Review the current uncommitted changes for avoidable complexity only. Do not modify files. Inspect staged, unstaged, and untracked changes. Report only actionable deletion or simplification findings with file/line references. Use one line per finding: <file>:L<line>: <tag> <what to cut>. <replacement>. Tags: delete, stdlib, native, yagni, shrink. End with net: -<N> lines possible. If there are no findings, say "Lean already. Ship." and stop. Do not report correctness, security, performance, provenance, auditability, or necessary-test issues; leave those to the normal review. Do not suggest removing safety, correctness, provenance, auditability, or necessary tests.'

printf 'Review output: %s\n' "$review_dir"

codex --search -m gpt-5.5 -c 'model_reasoning_effort="xhigh"' review --uncommitted \
  >"$correctness_log" 2>&1 &
correctness_pid=$!

codex --search -m gpt-5.5 -c 'model_reasoning_effort="xhigh"' exec \
  --sandbox read-only \
  --output-last-message "$complexity_msg" \
  "$complexity_prompt" \
  >"$complexity_log" 2>&1 &
complexity_pid=$!

wait "$correctness_pid"
correctness_status=$?
wait "$complexity_pid"
complexity_status=$?

printf '\n--- Correctness review ---\n'
sed -n '1,240p' "$correctness_log"

printf '\n--- Avoidable-complexity review ---\n'
if [ -s "$complexity_msg" ]; then
  cat "$complexity_msg"
else
  sed -n '1,240p' "$complexity_log"
fi

if [ "$correctness_status" -ne 0 ] || [ "$complexity_status" -ne 0 ]; then
  printf '\nReview command failed: correctness=%s avoidable_complexity=%s\n' \
    "$correctness_status" "$complexity_status" >&2
  exit 1
fi
