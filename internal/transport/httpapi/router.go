package httpapi

import (
	"net/http"
	"time"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"

	"code/internal/app/links"
	"code/internal/transport/httpapi/handlers"
)

type RouterDeps struct {
	Links   *links.Service
	BaseURL string

	SentryMiddlewareTimeout time.Duration
}

func NewRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
		Timeout: deps.SentryMiddlewareTimeout,
	}))

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	h := handlers.New(deps.Links, deps.BaseURL)

	api := r.Group("/api")
	{
		api.GET("/links", h.ListLinks)
		api.POST("/links", h.CreateLink)
		api.GET("/links/:id", h.GetLink)
		api.PUT("/links/:id", h.UpdateLink)
		api.DELETE("/links/:id", h.DeleteLink)
	}

	r.GET("/r/:short_name", h.Redirect)

	return r
}
