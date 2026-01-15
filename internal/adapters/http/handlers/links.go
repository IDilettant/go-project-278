package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"code/internal/adapters/http/dto"
	"code/internal/adapters/http/problems"
	"code/internal/domain"
)

type CreateLinkRequest struct {
	OriginalURL string `json:"original_url" example:"https://example.com"`
	ShortName   string `json:"short_name" example:"abc123"`
}

type UpdateLinkRequest struct {
	OriginalURL string  `json:"original_url" example:"https://example.com/updated"`
	ShortName   *string `json:"short_name" example:"abc123"`
}

func (h *Handler) ListLinks(c *gin.Context) {
	rng, query, hasRange, err := parseListRange(c.Query("range"))
	if err != nil {
		writeInvalidRange(c)
		return
	}

	var (
		items []domain.Link
		total int64
	)

	if hasRange {
		items, total, err = h.svc.ListPage(c.Request.Context(), query.Offset, query.Limit, true)
	} else {
		items, err = h.svc.ListAll(c.Request.Context())
	}

	if err != nil {
		h.fail(c, err)
		return
	}

	resp := make([]dto.LinkResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, dto.FromDomain(it, h.baseURL))
	}

	if hasRange {
		if len(items) == 0 {
			c.Header("Content-Range", fmt.Sprintf("links */%d", total))
		} else {
			end := rng.Start + len(items) - 1
			c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", rng.Start, end, total))
		}
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateLink(c *gin.Context) {
	var req CreateLinkRequest

	err := bindJSONStrict(c, &req)
	if err != nil {
		badJSON(c)

		return
	}

	link, err := h.svc.Create(c.Request.Context(), req.OriginalURL, req.ShortName)
	if err != nil {
		h.fail(c, err)

		return
	}

	c.Header("Location", fmt.Sprintf("/api/links/%d", link.ID))
	c.Status(http.StatusCreated)
}

func (h *Handler) GetLink(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	link, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		h.fail(c, err)

		return
	}

	c.JSON(http.StatusOK, dto.FromDomain(link, h.baseURL))
}

func (h *Handler) UpdateLink(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req UpdateLinkRequest

	err := bindJSONStrict(c, &req)
	if err != nil {
		badJSON(c)

		return
	}

	if req.ShortName == nil || strings.TrimSpace(*req.ShortName) == "" {
		problems.WriteProblem(c, problems.Problem{
			Type:   problems.ProblemTypeValidation,
			Title:  problems.ValidationTitle,
			Status: http.StatusBadRequest,
			Detail: "invalid short_name",
		})

		return
	}

	link, err := h.svc.Update(c.Request.Context(), id, req.OriginalURL, *req.ShortName)
	if err != nil {
		h.fail(c, err)

		return
	}

	c.JSON(http.StatusOK, dto.FromDomain(link, h.baseURL))
}

func (h *Handler) DeleteLink(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	err := h.svc.Delete(c.Request.Context(), id)
	if err != nil {
		h.fail(c, err)

		return
	}

	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		problems.WriteProblem(c, problems.Problem{
			Type:   problems.ProblemTypeValidation,
			Title:  problems.ValidationTitle,
			Status: http.StatusBadRequest,
			Detail: "invalid id",
		})

		return 0, false
	}

	return id, true
}
