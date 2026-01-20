package httpapi

import (
	"github.com/gin-gonic/gin"

	"code/internal/adapters/http/handlers"
	"code/internal/app/links"
)

const (
	linkByIDPath   = "/links/:id"
	linksPath      = "/links"
	linkVisitsPath = "/link_visits"
)

type RouterDeps struct {
	Links   links.UseCase
	BaseURL string
}

type EnginePlugin func(*gin.Engine)

// NewEngine creates a bare gin.Engine and applies plugins in order.
func NewEngine(plugins ...EnginePlugin) *gin.Engine {
	r := gin.New()
	r.TrustedPlatform = gin.PlatformCloudflare

	for _, p := range plugins {
		p(r)
	}

	return r
}

// RegisterRoutes attaches routes/handlers to an existing engine.
func RegisterRoutes(r *gin.Engine, deps RouterDeps) {
	h := handlers.New(deps.Links, deps.BaseURL)

	r.NoRoute(h.NotFound)
	r.GET("/ping", h.Ping)

	api := r.Group("/api")
	{
		api.GET(linksPath, h.ListLinks)
		api.POST(linksPath, h.CreateLink)
		api.GET(linkByIDPath, h.GetLink)
		api.PUT(linkByIDPath, h.UpdateLink)
		api.DELETE(linkByIDPath, h.DeleteLink)
		api.GET(linkVisitsPath, h.ListLinkVisits)
	}

	r.GET("/r/:code", h.Redirect)
}
