package problems

const (
	ContentTypeProblemJSON = "application/problem+json"

	ProblemTypeValidation  = "validation_error"
	ProblemTypeInvalidJSON = "invalid_json"
	ProblemTypeNotFound    = "about:blank"
	ProblemTypeConflict    = "conflict"
	ProblemTypeTimeout     = "timeout"
	ProblemTypeInternal    = "internal_error"
	ProblemTypeCanceled    = "client_cancelled"

	TitleBadRequest     = "Bad Request"
	TitleValidation     = "Validation error"
	TitleConflict       = "Conflict"
	TitleNotFound       = "Not Found"
	TitleGatewayTimeout = "Gateway Timeout"
	TitleRequestTimeout = "Request Timeout"
	TitleInternalError  = "Internal Server Error"

	DetailInvalidURL        = "invalid url"
	DetailInvalidShortName  = "invalid short_name"
	DetailInvalidJSON       = "invalid json"
	DetailInvalidRange      = "invalid range"
	DetailInvalidID         = "invalid id"
	DetailShortNameConflict = "short_name already exists"
	DetailNotFound          = "not found"
	DetailTimeout           = "timeout"
	DetailRequestCanceled   = "request canceled"
	DetailInternalError     = "internal error"
)