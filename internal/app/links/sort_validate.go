package links

import "strings"

// NormalizeAndValidateSort ensures semantic validity without knowing transport formats.
func NormalizeAndValidateSort(raw Sort, def Sort, allowed AllowedSortFields) (Sort, error) {
	if raw.Field == "" && raw.Order == "" {
		return def, nil
	}

	order := SortOrder(strings.ToUpper(strings.TrimSpace(string(raw.Order))))
	switch order {
	case SortAsc, SortDesc:
	default:
		return Sort{}, ErrInvalidSort
	}

	field := SortField(strings.ToLower(strings.TrimSpace(string(raw.Field))))
	if field == "" || allowed == nil || !allowed.Has(field) {
		return Sort{}, ErrInvalidSort
	}

	return Sort{Field: field, Order: order}, nil
}
