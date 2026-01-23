package postgres

// Order matches Scan in listLinks.
var sqlLinksSelectCols = []string{
	qualify(sqlAliasLinks, sqlColID),
	qualify(sqlAliasLinks, sqlColOriginalURL),
	qualify(sqlAliasLinks, sqlColShortName),
	qualify(sqlAliasLinks, sqlColCreatedAt),
}

// Order matches Scan in listLinkVisits.
var sqlVisitsSelectCols = []string{
	qualify(sqlAliasVisits, sqlColID),
	qualify(sqlAliasVisits, sqlColLinkID),
	qualify(sqlAliasVisits, sqlColCreatedAt),
	qualify(sqlAliasVisits, sqlColIP),
	qualify(sqlAliasVisits, sqlColUserAgent),
	qualify(sqlAliasVisits, sqlColReferer),
	qualify(sqlAliasVisits, sqlColStatus),
}
