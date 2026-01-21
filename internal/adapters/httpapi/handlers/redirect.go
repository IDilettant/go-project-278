package handlers

import (
	"github.com/gin-gonic/gin"

	"code/internal/app/links"
)

func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("code")

	meta := links.VisitMeta{
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Referer:   c.GetHeader("Referer"),
	}

	url, status, err := h.svc.Redirect(c.Request.Context(), code, meta)
	if err != nil {
		h.fail(c, err)

		return
	}

	c.Redirect(status, url)
}
