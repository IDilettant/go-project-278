package handlers

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "-" {
			return ""
		}

		return name
	})

	return v
}

func validateStruct(v any) (map[string]string, bool) {
	if err := validate.Struct(v); err != nil {
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
