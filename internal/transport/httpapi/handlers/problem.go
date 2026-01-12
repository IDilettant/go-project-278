package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/domain"
)

type Problem struct {
	Type   string `json:"type,omitempty" example:"validation_error"`
	Title  string `json:"title" example:"Validation error"`
	Status int    `json:"status" example:"400"`
	Detail string `json:"detail,omitempty" example:"invalid short_name"`
}

func writeProblem(c *gin.Context, p Problem) {
	c.Header("Content-Type", contentTypeProblemJSON)
	c.JSON(p.Status, p)
}

func problemFromError(err error) Problem {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return Problem{Type: ProblemTypeNotFound, Title: "Not Found", Status: http.StatusNotFound}
	case errors.Is(err, domain.ErrInvalidURL):
		return Problem{
			Type:   ProblemTypeValidation,
			Title:  validationTitle,
			Status: http.StatusBadRequest,
			Detail: "invalid url",
		}
	case errors.Is(err, domain.ErrInvalidShortName):
		return Problem{
			Type:   ProblemTypeValidation,
			Title:  validationTitle,
			Status: http.StatusBadRequest,
			Detail: "invalid short_name",
		}
	case errors.Is(err, domain.ErrShortNameConflict):
		return Problem{
			Type:   ProblemTypeConflict,
			Title:  "Conflict",
			Status: http.StatusConflict,
			Detail: "short_name already exists",
		}
	case errors.Is(err, domain.ErrShortNameImmutable):
		return Problem{
			Type:   ProblemTypeValidation,
			Title:  validationTitle,
			Status: http.StatusUnprocessableEntity,
			Detail: "short_name is immutable",
		}
	case errors.Is(err, context.DeadlineExceeded):
		return Problem{
			Type:   ProblemTypeTimeout,
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
		return Problem{Type: ProblemTypeInternal, Title: "Internal Server Error", Status: http.StatusInternalServerError}
	}
}
