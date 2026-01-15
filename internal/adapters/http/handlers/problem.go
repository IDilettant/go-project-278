package handlers

import (
	"context"
	"errors"
	"net/http"

	"code/internal/adapters/http/problems"
	"code/internal/domain"
)

func problemFromError(err error) problems.Problem {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return problems.Problem{
			Type:   problems.ProblemTypeNotFound,
			Title:  "Not Found",
			Status: http.StatusNotFound,
			Detail: "not found",
		}
	case errors.Is(err, domain.ErrInvalidURL):
		return problems.Problem{
			Type:   problems.ProblemTypeValidation,
			Title:  problems.ValidationTitle,
			Status: http.StatusBadRequest,
			Detail: "invalid url",
		}
	case errors.Is(err, domain.ErrInvalidShortName):
		return problems.Problem{
			Type:   problems.ProblemTypeValidation,
			Title:  problems.ValidationTitle,
			Status: http.StatusBadRequest,
			Detail: "invalid short_name",
		}
	case errors.Is(err, domain.ErrShortNameConflict):
		return problems.Problem{
			Type:   problems.ProblemTypeConflict,
			Title:  "Conflict",
			Status: http.StatusConflict,
			Detail: "short_name already exists",
		}
	case errors.Is(err, context.DeadlineExceeded):
		return problems.Problem{
			Type:   problems.ProblemTypeTimeout,
			Title:  "Gateway Timeout",
			Status: http.StatusGatewayTimeout,
			Detail: "timeout",
		}
	case errors.Is(err, context.Canceled):
		return problems.Problem{
			Type:   "client_cancelled",
			Title:  "Request Timeout",
			Status: http.StatusRequestTimeout,
			Detail: "request canceled",
		}
	default:
		return problems.Problem{
			Type:   problems.ProblemTypeInternal,
			Title:  "Internal Server Error",
			Status: http.StatusInternalServerError,
			Detail: "internal error",
		}
	}
}
