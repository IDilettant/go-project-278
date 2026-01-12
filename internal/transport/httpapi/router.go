package httpapi

import (
	"time"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"

	"code/internal/app/links"
	"code/internal/transport/httpapi/handlers"
	"code/internal/transport/httpapi/middleware"
)

type RouterDeps struct {
	Links   *links.Service
	BaseURL string

	SentryMiddlewareTimeout time.Duration
	RequestTimeout          time.Duration
}

const (
	linkByIDPath = "/links/:id"
	linksPath    = "/links"
)

func NewRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
		Timeout: deps.SentryMiddlewareTimeout,
	}))

	if deps.RequestTimeout > 0 {
		r.Use(middleware.RequestTimeout(deps.RequestTimeout))
	}

	h := handlers.New(deps.Links, deps.BaseURL)

	r.NoRoute(h.NotFound)

	api := r.Group("/api")
	{
		api.GET(linksPath, h.ListLinks)
		api.POST(linksPath, h.CreateLink)
		api.GET(linkByIDPath, h.GetLink)
		api.PUT(linkByIDPath, h.UpdateLink)
		api.DELETE(linkByIDPath, h.DeleteLink)
	}

	r.GET("/r/:short_name", h.Redirect)

	return r
}
