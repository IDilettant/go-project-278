package handlers

const (
	contentTypeProblemJSON = "application/problem+json"

	ProblemTypeValidation  = "validation_error"
	ProblemTypeInvalidJSON = "invalid_json"
	ProblemTypeNotFound    = "about:blank"
	ProblemTypeConflict    = "conflict"
	ProblemTypeTimeout     = "timeout"
	ProblemTypeInternal    = "internal_error"

	validationTitle = "Validation error"
)
