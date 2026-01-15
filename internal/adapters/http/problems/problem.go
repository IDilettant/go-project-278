package problems

import (
	"github.com/gin-gonic/gin"
)

type Problem struct {
	Type   string `json:"type" example:"validation_error"`
	Title  string `json:"title" example:"Validation error"`
	Status int    `json:"status" example:"400"`
	Detail string `json:"detail,omitempty" example:"invalid short_name"`
}

func WriteProblem(c *gin.Context, p Problem) {
	c.Header("Content-Type", ContentTypeProblemJSON)
	c.JSON(p.Status, p)
}
