package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"code/internal/domain"
)

type createLinkRequest struct {
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
}

type updateLinkRequest struct {
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
}

type linkResponse struct {
	ID          int64  `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
	ShortURL    string `json:"short_url"`
}

func (h *Handler) ListLinks(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		h.fail(c, err)
		return
	}

	resp := make([]linkResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, h.toResponse(it))
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateLink(c *gin.Context) {
	var req createLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})

		return
	}

	link, err := h.svc.Create(c.Request.Context(), req.OriginalURL, req.ShortName)
	if err != nil {
		h.fail(c, err)

		return
	}

	c.JSON(http.StatusCreated, h.toResponse(link))
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

	c.JSON(http.StatusOK, h.toResponse(link))
}

func (h *Handler) UpdateLink(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req updateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})

		return
	}

	link, err := h.svc.Update(c.Request.Context(), id, req.OriginalURL, req.ShortName)
	if err != nil {
		h.fail(c, err)

		return
	}

	c.JSON(http.StatusOK, h.toResponse(link))
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
	raw := c.Param("id")

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})

		return 0, false
	}

	return id, true
}

func (h *Handler) toResponse(l domain.Link) linkResponse {
	return linkResponse{
		ID:          l.ID,
		OriginalURL: l.OriginalURL,
		ShortName:   l.ShortName,
		ShortURL:    h.baseURL + "/r/" + l.ShortName,
	}
}
