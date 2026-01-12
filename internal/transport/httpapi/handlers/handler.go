package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/app/links"
)

type Handler struct {
	svc     links.UseCase
	baseURL string
}

func New(svc links.UseCase, baseURL string) *Handler {
	return &Handler{svc: svc, baseURL: baseURL}
}

func (h *Handler) fail(c *gin.Context, err error) {
	writeProblem(c, problemFromError(err))
}

func (h *Handler) NotFound(c *gin.Context) {
	writeProblem(c, Problem{Type: problemTypeNotFound, Title: "Not Found", Status: http.StatusNotFound})
}
