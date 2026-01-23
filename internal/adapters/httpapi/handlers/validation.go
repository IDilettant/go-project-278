package handlers

import (
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var initValidationOnce sync.Once

// InitValidation configures validator tag names to use JSON field names.
func InitValidation() {
	initValidationOnce.Do(func() {
		v, ok := binding.Validator.Engine().(*validator.Validate)
		if !ok {
			return
		}

		v.RegisterTagNameFunc(func(field reflect.StructField) string {
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if name == "-" {
				return ""
			}

			return name
		})
	})
}

func validateStruct(v any) (map[string]string, bool) {
	if err := binding.Validator.ValidateStruct(v); err != nil {
		verrs, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, false
		}

		out := make(map[string]string, len(verrs))
		for _, verr := range verrs {
			field := verr.Field()
			if field == "" {
				continue
			}

			if _, exists := out[field]; exists {
				continue
			}

			out[field] = verr.Error()
		}

		if len(out) == 0 {
			return nil, false
		}

		return out, true
	}

	return nil, false
}

func writeValidationErrors(c *gin.Context, errs map[string]string) {
	c.JSON(http.StatusUnprocessableEntity, gin.H{
		"errors": errs,
	})
}
