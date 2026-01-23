package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"code/internal/adapters/httpapi/dto"
	"code/internal/adapters/httpapi/problems"
	"code/internal/app/links"
	"code/internal/domain"
)

type CreateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required" example:"https://example.com"`
	ShortName   string `json:"short_name" binding:"omitempty,min=3,max=32" example:"abc123"`
}

type UpdateLinkRequest struct {
	ID          int64  `json:"id"`
	OriginalURL string `json:"original_url" binding:"required" example:"https://example.com/updated"`
	ShortName   string `json:"short_name" binding:"omitempty,min=3,max=32" example:"abc123"`
	ShortURL    string `json:"short_url" example:"https://example.com/r/abc123"`
}

func (h *Handler) ListLinks(c *gin.Context) {
	rng, _, hasRange, err := parseRangeFromRequest(c)
	if err != nil {
		writeInvalidRange(c)

		return
	}

	rawSort, ok := parseReactAdminSort(c.Query("sort"))
	if !ok {
		h.fail(c, links.ErrInvalidSort)

		return
	}

	sort, err := links.NormalizeAndValidateSort(rawSort, links.DefaultLinksSort, links.AllowedLinksSortFields())
	if err != nil {
		h.fail(c, err)

		return
	}

	var (
		items []domain.Link
		total int64
	)

	query := links.LinksQuery{Sort: sort}
	if hasRange {
		query.Range = &rng
	}

	items, total, err = h.svc.ListLinks(c.Request.Context(), query)

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

	if err := BindJSONStrict(c, &req); err != nil {
		badJSON(c)

		return
	}

	req.OriginalURL = strings.TrimSpace(req.OriginalURL)
	req.ShortName = strings.TrimSpace(req.ShortName)

	if errs, ok := validateStruct(req); ok {
		writeValidationErrors(c, errs)

		return
	}

	link, err := h.svc.Create(c.Request.Context(), req.OriginalURL, req.ShortName)
	if err != nil {
		h.fail(c, err)

		return
	}

	c.Header("Location", fmt.Sprintf("/api/links/%d", link.ID))
	c.JSON(http.StatusCreated, dto.FromDomain(link, h.baseURL))
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

	if err := BindJSONStrict(c, &req); err != nil {
		badJSON(c)

		return
	}

	req.OriginalURL = strings.TrimSpace(req.OriginalURL)
	req.ShortName = strings.TrimSpace(req.ShortName)

	if errs, ok := validateStruct(req); ok {
		writeValidationErrors(c, errs)

		return
	}

	link, err := h.svc.Update(c.Request.Context(), id, req.OriginalURL, req.ShortName)
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
			Title:  problems.TitleValidation,
			Status: http.StatusBadRequest,
			Detail: problems.DetailInvalidID,
		})

		return 0, false
	}

	return id, true
}
