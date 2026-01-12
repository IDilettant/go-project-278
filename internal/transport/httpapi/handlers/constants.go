package handlers

const (
	contentTypeProblemJSON = "application/problem+json"

	problemTypeValidation  = "validation_error"
	problemTypeInvalidJSON = "invalid_json"
	problemTypeNotFound    = "about:blank"
	problemTypeConflict    = "conflict"
	problemTypeTimeout     = "timeout"
	problemTypeInternal    = "internal_error"

	validationTitle = "Validation error"
)
