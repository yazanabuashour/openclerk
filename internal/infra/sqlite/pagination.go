package sqlite

import "github.com/yazanabuashour/openclerk/internal/domain"

const sqliteMaxPageLimit = 100

func normalizePageLimit(raw int, defaultLimit int) (int, error) {
	limit := raw
	if limit == 0 {
		limit = defaultLimit
	}
	if limit < 1 || limit > sqliteMaxPageLimit {
		return 0, domain.ValidationError("limit must be between 1 and 100", map[string]any{"limit": limit})
	}
	return limit, nil
}

func paginateSlice[T any](items []T, limit int, offset int) ([]T, domain.PageInfo) {
	pageInfo := domain.PageInfo{}
	if len(items) > limit {
		pageInfo.HasMore = true
		pageInfo.NextCursor = encodeCursor(offset + limit)
		items = items[:limit]
	}
	return items, pageInfo
}
