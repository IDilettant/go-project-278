package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestTimeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := c.Request.Context().Deadline(); ok {
			c.Next()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
