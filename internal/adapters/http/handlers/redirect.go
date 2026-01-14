package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Redirect godoc
// @Summary Redirect by short name
// @Description Redirects to the original URL by short name.
// @Tags redirect
// @Produce application/problem+json
// @Param short_name path string true "Short name" minlength(4) maxlength(32)
// @Success 302 {string} string "Found"
// @Header 302 {string} Location "Redirect target"
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Failure 500 {object} Problem
// @Router /r/{short_name} [get]
func (h *Handler) Redirect(c *gin.Context) {
	shortName := c.Param("short_name")

	link, err := h.svc.GetByShortName(c.Request.Context(), shortName)
	if err != nil {
		h.fail(c, err)

		return
	}

	c.Redirect(http.StatusFound, link.OriginalURL) // 302
}
