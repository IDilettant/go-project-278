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
			Title:  problems.TitleNotFound,
			Status: http.StatusNotFound,
			Detail: problems.DetailNotFound,
		}
	case errors.Is(err, domain.ErrInvalidURL):
		return problems.Problem{
			Type:   problems.ProblemTypeValidation,
			Title:  problems.TitleValidation,
			Status: http.StatusBadRequest,
			Detail: problems.DetailInvalidURL,
		}
	case errors.Is(err, domain.ErrInvalidShortName):
		return problems.Problem{
			Type:   problems.ProblemTypeValidation,
			Title:  problems.TitleValidation,
			Status: http.StatusBadRequest,
			Detail: problems.DetailInvalidShortName,
		}
	case errors.Is(err, domain.ErrShortNameConflict):
		return problems.Problem{
			Type:   problems.ProblemTypeConflict,
			Title:  problems.TitleConflict,
			Status: http.StatusConflict,
			Detail: problems.DetailShortNameConflict,
		}
	case errors.Is(err, context.DeadlineExceeded):
		return problems.Problem{
			Type:   problems.ProblemTypeTimeout,
			Title:  problems.TitleGatewayTimeout,
			Status: http.StatusGatewayTimeout,
			Detail: problems.DetailTimeout,
		}
	case errors.Is(err, context.Canceled):
		return problems.Problem{
			Type:   problems.ProblemTypeCanceled,
			Title:  problems.TitleRequestTimeout,
			Status: http.StatusRequestTimeout,
			Detail: problems.DetailRequestCanceled,
		}
	default:
		return problems.Problem{
			Type:   problems.ProblemTypeInternal,
			Title:  problems.TitleInternalError,
			Status: http.StatusInternalServerError,
			Detail: problems.DetailInternalError,
		}
	}
}
