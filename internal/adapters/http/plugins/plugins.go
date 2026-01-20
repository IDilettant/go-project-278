package plugins

import (
	"time"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"

	"code/internal/adapters/http/middleware"
)

func Logger() func(*gin.Engine) {
	return func(r *gin.Engine) {
		r.Use(gin.Logger())
	}
}

func Recovery() func(*gin.Engine) {
	return func(r *gin.Engine) {
		r.Use(middleware.Recovery())
	}
}

func Sentry(timeout time.Duration) func(*gin.Engine) {
	return func(r *gin.Engine) {
		r.Use(sentrygin.New(sentrygin.Options{
			Repanic: true,
			Timeout: timeout,
		}))
	}
}

func RequestTimeout(d time.Duration) func(*gin.Engine) {
	return func(r *gin.Engine) {
		if d > 0 {
			r.Use(middleware.RequestTimeout(d))
		}
	}
}

func RequestID() func(*gin.Engine) {
	return func(r *gin.Engine) {
		r.Use(middleware.RequestID())
	}
}

func CORS(origins []string) func(*gin.Engine) {
	return func(r *gin.Engine) {
		if len(origins) > 0 {
			r.Use(middleware.CORS(origins))
		}
	}
}
