package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Ping(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("pong"))
}
