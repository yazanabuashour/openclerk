package main

func classifyTargetedProfileConfigResult(result jobResult) (string, string) {
	if result.Status != "completed" {
		return "agent_run_incomplete", "profile configuration smoke did not complete"
	}
	if !result.Verification.Passed {
		return "profile_config_gap", "profile config smoke did not prove configure, inspect, write gate, retrieval gate, override, and clear"
	}
	return "none", "profile config smoke verified persisted defaults, request overrides, and restored built-in defaults"
}
