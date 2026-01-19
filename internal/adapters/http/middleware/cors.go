package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	allowedMethods = "GET,POST,PUT,DELETE,OPTIONS"
	allowedHeaders = "Content-Type, Authorization"
	exposeHeaders  = "Content-Range, Location"
)

func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowAll := false
	allowed := make(map[string]struct{})

	for _, origin := range allowedOrigins {
		origin = normalizeOrigin(origin)
		if origin == "" {
			continue
		}

		if origin == "*" {
			allowAll = true
			allowed = nil

			break
		}

		allowed[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := normalizeOrigin(c.GetHeader("Origin"))
		allow := false

		if origin != "" {
			if allowAll {
				allow = true
				c.Header("Access-Control-Allow-Origin", "*")
			} else if _, ok := allowed[origin]; ok {
				allow = true
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
		}

		if c.Request.Method == http.MethodOptions {
			if !allow {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			c.Header("Access-Control-Allow-Methods", allowedMethods)
			c.Header("Access-Control-Allow-Headers", allowedHeaders)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		if allow {
			c.Header("Access-Control-Allow-Methods", allowedMethods)
			c.Header("Access-Control-Allow-Headers", allowedHeaders)
			c.Header("Access-Control-Expose-Headers", exposeHeaders)
		}

		c.Next()
	}
}

func normalizeOrigin(origin string) string {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return ""
	}

	return strings.TrimRight(origin, "/")
}
