package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/domain"
)

type Problem struct {
	Type   string `json:"type,omitempty"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

const validationTitle = "Validation error"

func writeProblem(c *gin.Context, p Problem) {
	c.Header("Content-Type", "application/problem+json")
	c.JSON(p.Status, p)
}

func problemFromError(err error) Problem {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return Problem{Type: "about:blank", Title: "Not Found", Status: http.StatusNotFound}
	case errors.Is(err, domain.ErrInvalidURL):
		return Problem{
			Type:   "validation_error",
			Title:  validationTitle,
			Status: http.StatusBadRequest,
			Detail: "invalid url",
		}
	case errors.Is(err, domain.ErrInvalidShortName):
		return Problem{
			Type:   "validation_error",
			Title:  validationTitle,
			Status: http.StatusBadRequest,
			Detail: "invalid short_name",
		}
	case errors.Is(err, domain.ErrShortNameConflict):
		return Problem{
			Type:   "conflict",
			Title:  "Conflict",
			Status: http.StatusConflict,
			Detail: "short_name already exists",
		}
	case errors.Is(err, domain.ErrShortNameImmutable):
		return Problem{
			Type:   "validation_error",
			Title:  validationTitle,
			Status: http.StatusUnprocessableEntity,
			Detail: "short_name is immutable",
		}
	case errors.Is(err, context.DeadlineExceeded):
		return Problem{
			Type:   "timeout",
			Title:  "Gateway Timeout",
			Status: http.StatusGatewayTimeout,
		}
	case errors.Is(err, context.Canceled):
		return Problem{
			Type:   "client_cancelled",
			Title:  "Request Timeout",
			Status: http.StatusRequestTimeout,
			Detail: "request canceled",
		}
	default:
		return Problem{Type: "internal_error", Title: "Internal Server Error", Status: http.StatusInternalServerError}
	}
}

func badJSON(c *gin.Context) {
	writeProblem(c, Problem{
		Type:   "invalid_json",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: "invalid json",
	})
}
