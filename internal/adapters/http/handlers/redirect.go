package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Redirect(c *gin.Context) {
	shortName := c.Param("short_name")

	link, err := h.svc.GetByShortName(c.Request.Context(), shortName)
	if err != nil {
		h.fail(c, err)

		return
	}

	c.Redirect(http.StatusFound, link.OriginalURL) // 302
}
