package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexReviewSequenceRoutesRequestedExtras(t *testing.T) {
	repo, toolsDir := setupReviewSequenceRepo(t)
	logPath := filepath.Join(t.TempDir(), "codex.log")

	output, err := runReviewSequenceScript(t, repo, toolsDir, logPath, " security, test-gaps ", "")
	if err != nil {
		t.Fatalf("review sequence failed: %v\n%s", err, output)
	}

	assertContains(t, output, "--- correctness-review ---")
	assertContains(t, output, "--- avoidable-complexity-review ---")
	assertContains(t, output, "--- test-gap-review ---")
	assertContains(t, output, "--- security-review ---")
	assertContains(t, output, "Review sequence complete.")
	assertNotContains(t, output, "--- api-compat-review ---")
	assertNotContains(t, output, "--- concurrency-review ---")

	calls := readTestFile(t, logPath)
	lines := nonEmptyLines(calls)
	if len(lines) != 4 {
		t.Fatalf("expected 4 codex calls, got %d:\n%s", len(lines), calls)
	}
	assertContains(t, calls, "<--search> <-m> <gpt-5.5> <-c> <model_reasoning_effort=\"xhigh\"> <review> <--uncommitted>")
	assertContains(t, calls, "<exec> <--sandbox> <read-only> <--output-last-message>")
	assertContains(t, calls, "Review the current uncommitted changes.")
	assertContains(t, calls, "security regressions")
	assertContains(t, calls, "missing, weak, or misleading validation")
	assertNotContains(t, calls, "API, CLI, config/env")
	assertNotContains(t, calls, "concurrency, lifecycle")

	rejectLogPath := filepath.Join(t.TempDir(), "codex.log")
	output, err = runReviewSequenceScript(t, repo, toolsDir, rejectLogPath, "security,bad", "")
	if err == nil {
		t.Fatalf("review sequence accepted unknown extra:\n%s", output)
	}
	assertContains(t, output, "unknown CODEX_REVIEW_EXTRA item=bad")
	if data, err := os.ReadFile(rejectLogPath); err == nil && len(data) > 0 {
		t.Fatalf("codex should not be invoked for unknown extras:\n%s", data)
	}
}

func TestCodexReviewSequenceReportsFocusedReviewerFailureAfterWaiting(t *testing.T) {
	repo, toolsDir := setupReviewSequenceRepo(t)
	logPath := filepath.Join(t.TempDir(), "codex.log")

	output, err := runReviewSequenceScript(t, repo, toolsDir, logPath, "security,test-gaps", "security regressions")
	if err == nil {
		t.Fatalf("review sequence succeeded despite focused reviewer failure:\n%s", output)
	}

	assertContains(t, output, "--- correctness-review ---")
	assertContains(t, output, "--- avoidable-complexity-review ---")
	assertContains(t, output, "--- test-gap-review ---")
	assertContains(t, output, "--- security-review ---")
	assertContains(t, output, "security-review=7")
	assertContains(t, output, "Review command failed:")

	calls := readTestFile(t, logPath)
	lines := nonEmptyLines(calls)
	if len(lines) != 4 {
		t.Fatalf("expected all 4 reviewers to run before failure was reported, got %d:\n%s", len(lines), calls)
	}
}

func setupReviewSequenceRepo(t *testing.T) (string, string) {
	t.Helper()

	repo := t.TempDir()
	scriptsDir := filepath.Join(repo, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatalf("mkdir scripts dir: %v", err)
	}

	script, err := os.ReadFile("codex-review-sequence.sh")
	if err != nil {
		t.Fatalf("read review script: %v", err)
	}
	if err := os.WriteFile(filepath.Join(scriptsDir, "codex-review-sequence.sh"), script, 0o755); err != nil {
		t.Fatalf("write review script: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("before\n"), 0o644); err != nil {
		t.Fatalf("write tracked file: %v", err)
	}

	runGit(t, repo, "init", "-q")
	runGit(t, repo, "config", "user.email", "test@example.com")
	runGit(t, repo, "config", "user.name", "Test User")
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-q", "-m", "init")
	if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("after\n"), 0o644); err != nil {
		t.Fatalf("modify tracked file: %v", err)
	}

	toolsDir := filepath.Join(repo, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools dir: %v", err)
	}
	writeExecutable(t, filepath.Join(toolsDir, "codex"), `#!/usr/bin/env bash
log="${OPENCLERK_REVIEW_TEST_LOG:?}"
line="args:"
for arg in "$@"; do
  line="$line <$arg>"
done
printf '%s\n' "$line" >> "$log"

fail_match="${OPENCLERK_REVIEW_TEST_FAIL_MATCH:-__openclerk_no_match__}"
case " $* " in
  *"$fail_match"*) exit 7 ;;
esac

prev=""
for arg in "$@"; do
  if [ "$prev" = "--output-last-message" ]; then
    printf 'last message\n' > "$arg"
  fi
  prev="$arg"
done
`)

	return repo, toolsDir
}

func runReviewSequenceScript(t *testing.T, repo string, toolsDir string, logPath string, extra string, failMatch string) (string, error) {
	t.Helper()

	cmd := exec.Command("bash", "scripts/codex-review-sequence.sh")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(),
		"HOME="+t.TempDir(),
		"PATH="+joinPath(toolsDir, systemPath()),
		"CODEX_REVIEW_EXTRA="+extra,
		"OPENCLERK_REVIEW_TEST_LOG="+logPath,
		"OPENCLERK_REVIEW_TEST_FAIL_MATCH="+failMatch,
	)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
}

func nonEmptyLines(value string) []string {
	var lines []string
	for _, line := range strings.Split(value, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func assertNotContains(t *testing.T, haystack string, needle string) {
	t.Helper()

	if strings.Contains(haystack, needle) {
		t.Fatalf("expected output not to contain %q:\n%s", needle, haystack)
	}
}
