package postgres

import "code/internal/app/links"

func qualify(alias, col string) string {
	return alias + "." + col
}

func normalizeOrder(o links.SortOrder) (links.SortOrder, bool) {
	switch o {
	case links.SortAsc, links.SortDesc:
		return o, true
	default:
		return "", false
	}
}

func orderExpr(alias, col string, ord links.SortOrder) string {
	return qualify(alias, col) + " " + string(ord)
}

func orderExprWithTie(alias, col string, ord links.SortOrder) string {
	return orderExpr(alias, col, ord) + ", " + orderExpr(alias, sqlColID, ord)
}

func orderByLinks(sort links.Sort) (string, error) {
	ord, ok := normalizeOrder(sort.Order)
	if !ok {
		return "", links.ErrInvalidSort
	}

	switch sort.Field {
	case links.SortFieldID:
		return orderExpr(sqlAliasLinks, sqlColID, ord), nil
	case links.SortFieldShortName:
		return orderExprWithTie(sqlAliasLinks, sqlColShortName, ord), nil
	case links.SortFieldOriginalURL:
		return orderExprWithTie(sqlAliasLinks, sqlColOriginalURL, ord), nil
	default:
		return "", links.ErrInvalidSort
	}
}

func orderByLinkVisits(sort links.Sort) (string, error) {
	ord, ok := normalizeOrder(sort.Order)
	if !ok {
		return "", links.ErrInvalidSort
	}

	switch sort.Field {
	case links.SortFieldID:
		return orderExpr(sqlAliasVisits, sqlColID, ord), nil
	case links.SortFieldLinkID:
		return orderExprWithTie(sqlAliasVisits, sqlColLinkID, ord), nil
	case links.SortFieldIP:
		return orderExprWithTie(sqlAliasVisits, sqlColIP, ord), nil
	case links.SortFieldStatus:
		return orderExprWithTie(sqlAliasVisits, sqlColStatus, ord), nil
	case links.SortFieldReferer:
		return orderExprWithTie(sqlAliasVisits, sqlColReferer, ord), nil
	case links.SortFieldCreatedAt:
		return orderExprWithTie(sqlAliasVisits, sqlColCreatedAt, ord), nil
	default:
		return "", links.ErrInvalidSort
	}
}
