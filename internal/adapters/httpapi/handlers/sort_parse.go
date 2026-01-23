package handlers

import (
	"encoding/json"
	"strings"

	"code/internal/app/links"
)

const sortPartsCount = 2

// parseReactAdminSort parses the query param `sort` expected as JSON array: ["field","ASC"].
// This is HTTP contract logic and must stay in the adapter layer.
func parseReactAdminSort(raw string) (links.Sort, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return links.Sort{}, true
	}

	var parts []string
	if err := json.Unmarshal([]byte(raw), &parts); err != nil || len(parts) != sortPartsCount {
		return links.Sort{}, false
	}

	field := strings.ToLower(strings.TrimSpace(parts[0]))
	order := strings.ToUpper(strings.TrimSpace(parts[1]))
	if field == "" || order == "" {
		return links.Sort{}, false
	}

	switch field {
	case "short_url":
		field = string(links.SortFieldShortName)
	case "reffer":
		field = string(links.SortFieldReferer)
	}

	return links.Sort{
		Field: links.SortField(field),
		Order: links.SortOrder(order),
	}, true
}
