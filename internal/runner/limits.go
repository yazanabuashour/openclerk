package runner

import (
	"strconv"

	"github.com/yazanabuashour/openclerk/internal/domain"
)

const negativeRunnerLimitRejection = "limit must be greater than or equal to 0"

func defaultRunnerLimit(raw int, defaultLimit int) int {
	if raw == 0 {
		return defaultLimit
	}
	return raw
}

func cappedRunnerLimit(raw int, defaultLimit int, maxLimit int) int {
	limit := defaultRunnerLimit(raw, defaultLimit)
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

func boundedRunnerLimit(raw int, defaultLimit int, maxLimit int, field string) (int, error) {
	limit := defaultRunnerLimit(raw, defaultLimit)
	if limit < 1 || limit > maxLimit {
		return 0, domain.ValidationError(field+".limit must be between 1 and "+strconv.Itoa(maxLimit), map[string]any{"limit": limit})
	}
	return limit, nil
}

func rejectNegativeRunnerLimits(limits ...int) string {
	for _, limit := range limits {
		if limit < 0 {
			return negativeRunnerLimitRejection
		}
	}
	return ""
}
