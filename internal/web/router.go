package web

import (
	"net/http"
	"time"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	app := gin.Default()

	app.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
		Timeout: 2 * time.Second,
	}))

	app.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	app.GET("/debug-sentry", func(c *gin.Context) {
		panic("sentry debug panic")
	})

	app.GET("/debug-sentry-message", func(c *gin.Context) {
		if hub := sentrygin.GetHubFromContext(c); hub != nil {
			hub.CaptureMessage("sentry debug message")
		}
		c.String(http.StatusOK, "sent")
	})

	return app
}
