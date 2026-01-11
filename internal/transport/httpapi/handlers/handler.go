package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

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
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, domain.ErrInvalidURL) || errors.Is(err, domain.ErrInvalidShortName):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrShortNameConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "short_name already exists"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
