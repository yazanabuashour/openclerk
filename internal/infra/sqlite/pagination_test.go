package sqlite

import "testing"

func TestNormalizePageLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		raw       int
		defaultTo int
		want      int
		wantErr   bool
	}{
		{name: "default", raw: 0, defaultTo: 20, want: 20},
		{name: "explicit", raw: 5, defaultTo: 20, want: 5},
		{name: "negative", raw: -1, defaultTo: 20, wantErr: true},
		{name: "too high", raw: 101, defaultTo: 20, wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizePageLimit(tt.raw, tt.defaultTo)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("normalizePageLimit(%d, %d) succeeded, want error", tt.raw, tt.defaultTo)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizePageLimit(%d, %d): %v", tt.raw, tt.defaultTo, err)
			}
			if got != tt.want {
				t.Fatalf("limit = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPaginateSlice(t *testing.T) {
	t.Parallel()

	items, pageInfo := paginateSlice([]int{1, 2, 3}, 2, 4)
	if len(items) != 2 || items[0] != 1 || items[1] != 2 {
		t.Fatalf("items = %+v, want first two", items)
	}
	if !pageInfo.HasMore || pageInfo.NextCursor != encodeCursor(6) {
		t.Fatalf("pageInfo = %+v, want next cursor for offset+limit", pageInfo)
	}
}
