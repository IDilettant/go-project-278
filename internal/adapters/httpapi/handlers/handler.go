package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/adapters/httpapi/problems"
	"code/internal/app/links"
	"code/internal/domain"
)

type Handler struct {
	svc     links.UseCase
	baseURL string
}

func New(svc links.UseCase, baseURL string) *Handler {
	return &Handler{svc: svc, baseURL: baseURL}
}

func (h *Handler) fail(c *gin.Context, err error) {
	if errs, ok := validationErrorsFromDomain(err); ok {
		writeValidationErrors(c, errs)

		return
	}

	if errors.Is(err, links.ErrInvalidSort) {
		writeInvalidSort(c)

		return
	}

	problems.WriteProblem(c, problemFromError(err))
}

func (h *Handler) NotFound(c *gin.Context) {
	problems.WriteProblem(c, problems.Problem{
		Type:   problems.ProblemTypeNotFound,
		Title:  problems.TitleNotFound,
		Status: http.StatusNotFound,
		Detail: problems.DetailNotFound,
	})
}

func validationErrorsFromDomain(err error) (map[string]string, bool) {
	switch {
	case errors.Is(err, domain.ErrInvalidURL):
		return map[string]string{"original_url": "invalid url"}, true
	case errors.Is(err, domain.ErrInvalidShortName):
		return map[string]string{"short_name": "invalid short_name"}, true
	case errors.Is(err, domain.ErrShortNameConflict):
		return map[string]string{"short_name": "short name already in use"}, true
	default:
		return nil, false
	}
}
