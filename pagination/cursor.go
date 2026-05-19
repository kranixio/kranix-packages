package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"sort"
	"strings"
)

const (
	DefaultLimit = 50
	MaxLimit     = 500
)

// CursorPayload is the decoded cursor token (stable sort by resource id).
type CursorPayload struct {
	AfterID string `json:"after_id"`
	V       int    `json:"v"`
}

// PageInfo describes cursor-based pagination metadata.
type PageInfo struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	TotalCount int    `json:"total_count"`
}

// Params from query string.
type Params struct {
	Limit  int
	Cursor string
}

// ParseParams reads limit and cursor from query values.
func ParseParams(limitStr, cursor string) Params {
	limit := DefaultLimit
	if limitStr != "" {
		if n, err := parsePositiveInt(limitStr); err == nil {
			limit = n
		}
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return Params{Limit: limit, Cursor: strings.TrimSpace(cursor)}
}

func parsePositiveInt(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("invalid")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// EncodeCursor returns a URL-safe opaque cursor.
func EncodeCursor(afterID string) string {
	if afterID == "" {
		return ""
	}
	b, _ := json.Marshal(CursorPayload{AfterID: afterID, V: 1})
	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodeCursor parses an opaque cursor token.
func DecodeCursor(cursor string) (CursorPayload, error) {
	cursor = strings.TrimSpace(cursor)
	if cursor == "" {
		return CursorPayload{}, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return CursorPayload{}, errors.New("invalid cursor")
	}
	var p CursorPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return CursorPayload{}, errors.New("invalid cursor")
	}
	return p, nil
}

// IDProvider returns a stable string id for sorting and cursors.
type IDProvider interface {
	GetID() string
}

// SlicePage paginates a sorted slice of items by ID ascending.
func SlicePage[T IDProvider](all []T, p Params) (page []T, info PageInfo) {
	info.TotalCount = len(all)
	info.Limit = p.Limit

	sorted := make([]T, len(all))
	copy(sorted, all)
	sort.Slice(sorted, func(i, j int) bool {
		return strings.Compare(sorted[i].GetID(), sorted[j].GetID()) < 0
	})

	start := 0
	if p.Cursor != "" {
		cur, err := DecodeCursor(p.Cursor)
		if err == nil && cur.AfterID != "" {
			for i, item := range sorted {
				if item.GetID() > cur.AfterID {
					start = i
					break
				}
				if i == len(sorted)-1 {
					start = len(sorted)
				}
			}
		}
	}

	end := start + p.Limit
	if end > len(sorted) {
		end = len(sorted)
	}
	page = sorted[start:end]
	info.HasMore = end < len(sorted)
	if info.HasMore && len(page) > 0 {
		info.NextCursor = EncodeCursor(page[len(page)-1].GetID())
	}
	return page, info
}
