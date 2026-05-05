package runner

import "testing"

func TestRunnerLimitHelpers(t *testing.T) {
	t.Parallel()

	if got := defaultRunnerLimit(0, 10); got != 10 {
		t.Fatalf("defaultRunnerLimit default = %d, want 10", got)
	}
	if got := defaultRunnerLimit(7, 10); got != 7 {
		t.Fatalf("defaultRunnerLimit explicit = %d, want 7", got)
	}
	if got := cappedRunnerLimit(99, 10, 20); got != 20 {
		t.Fatalf("cappedRunnerLimit = %d, want 20", got)
	}
	if got, err := boundedRunnerLimit(0, 10, 100, "report"); err != nil || got != 10 {
		t.Fatalf("boundedRunnerLimit default = %d, %v; want 10, nil", got, err)
	}
	if _, err := boundedRunnerLimit(101, 10, 100, "report"); err == nil {
		t.Fatal("boundedRunnerLimit over max succeeded, want error")
	}
	if rejection := rejectNegativeRunnerLimits(0, 1, -1); rejection != negativeRunnerLimitRejection {
		t.Fatalf("rejectNegativeRunnerLimits = %q, want %q", rejection, negativeRunnerLimitRejection)
	}
	if rejection := rejectNegativeRunnerLimits(0, 1, 2); rejection != "" {
		t.Fatalf("rejectNegativeRunnerLimits accepted positive limits = %q, want empty", rejection)
	}
}
