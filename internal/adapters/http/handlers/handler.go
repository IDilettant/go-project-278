package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/adapters/http/problems"
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
