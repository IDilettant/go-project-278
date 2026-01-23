package links

type SortOrder string

const (
	SortAsc  SortOrder = "ASC"
	SortDesc SortOrder = "DESC"
)

type SortField string

const (
	SortFieldID          SortField = "id"
	SortFieldShortName   SortField = "short_name"
	SortFieldOriginalURL SortField = "original_url"
	SortFieldLinkID      SortField = "link_id"
	SortFieldIP          SortField = "ip"
	SortFieldStatus      SortField = "status"
	SortFieldReferer     SortField = "referer"
	SortFieldCreatedAt   SortField = "created_at"
)

type Sort struct {
	Field SortField
	Order SortOrder
}

var DefaultLinksSort = Sort{Field: SortFieldID, Order: SortAsc}
var DefaultLinkVisitsSort = Sort{Field: SortFieldCreatedAt, Order: SortDesc}
