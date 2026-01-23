package links

// AllowedSortFields is a semantic rule set for allowed fields per use case.
type AllowedSortFields map[SortField]struct{}

func (a AllowedSortFields) Has(f SortField) bool {
	_, ok := a[f]
	return ok
}

func AllowedLinksSortFields() AllowedSortFields {
	return AllowedSortFields{
		SortFieldID:          {},
		SortFieldShortName:   {},
		SortFieldOriginalURL: {},
	}
}

func AllowedLinkVisitsSortFields() AllowedSortFields {
	return AllowedSortFields{
		SortFieldID:        {},
		SortFieldLinkID:    {},
		SortFieldIP:        {},
		SortFieldStatus:    {},
		SortFieldReferer:   {},
		SortFieldCreatedAt: {},
	}
}
