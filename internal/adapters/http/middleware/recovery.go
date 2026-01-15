package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"code/internal/adapters/http/problems"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Printf("panic recovered: %v\n%s", recovered, debug.Stack())

		problems.WriteProblem(c, problems.Problem{
			Type:   problems.ProblemTypeInternal,
			Title:  problems.TitleInternalError,
			Status: http.StatusInternalServerError,
			Detail: problems.DetailInternalError,
		})
	})
}
