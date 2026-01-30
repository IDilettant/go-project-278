package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	allowedMethods = "GET,POST,PUT,DELETE,OPTIONS"
	allowedHeaders = "Content-Type, Authorization, Range"
	exposeHeaders  = "Content-Range, Location"
)

func CORS(allowedOrigins []string) gin.HandlerFunc {
	policy := newCORSPolicy(allowedOrigins)

	return func(c *gin.Context) {
		origin := normalizeOrigin(c.GetHeader("Origin"))
		allow := policy.applyOrigin(c, origin)

		if c.Request.Method == http.MethodOptions {
			handlePreflight(c, allow)
			
			return
		}

		if allow {
			applyCORSHeaders(c)
		}

		c.Next()
	}
}

type corsPolicy struct {
	allowAll bool
	allowed  map[string]struct{}
}

func newCORSPolicy(allowedOrigins []string) corsPolicy {
	policy := corsPolicy{
		allowed: make(map[string]struct{}),
	}

	for _, origin := range allowedOrigins {
		origin = normalizeOrigin(origin)
		if origin == "" {
			continue
		}

		if origin == "*" {
			policy.allowAll = true
			policy.allowed = nil

			break
		}

		policy.allowed[origin] = struct{}{}
	}

	return policy
}

func (p corsPolicy) applyOrigin(c *gin.Context, origin string) bool {
	if origin == "" {
		return false
	}

	if p.allowAll {
		c.Header("Access-Control-Allow-Origin", "*")
		
		return true
	}

	if _, ok := p.allowed[origin]; !ok {
		return false
	}

	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Vary", "Origin")

	return true
}

func handlePreflight(c *gin.Context, allow bool) {
	if !allow {
		c.AbortWithStatus(http.StatusForbidden)
		
		return
	}

	c.Header("Access-Control-Allow-Methods", allowedMethods)
	c.Header("Access-Control-Allow-Headers", allowedHeaders)
	c.AbortWithStatus(http.StatusNoContent)
}

func applyCORSHeaders(c *gin.Context) {
	c.Header("Access-Control-Allow-Methods", allowedMethods)
	c.Header("Access-Control-Allow-Headers", allowedHeaders)
	c.Header("Access-Control-Expose-Headers", exposeHeaders)
}

func normalizeOrigin(origin string) string {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return ""
	}

	return strings.TrimRight(origin, "/")
}
