package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/domain"
)

func (h *Handler) Redirect(c *gin.Context) {
	shortName := c.Param("short_name")

	link, err := h.svc.GetByShortName(c.Request.Context(), shortName)
	if err != nil {
		if err == domain.ErrNotFound {
			c.String(http.StatusNotFound, "not found")

			return
		}
		if err == domain.ErrInvalidShortName {
			c.String(http.StatusBadRequest, "invalid short_name")

			return
		}

		c.String(http.StatusInternalServerError, "internal error")

		return
	}

	c.Redirect(http.StatusFound, link.OriginalURL) // 302
}
