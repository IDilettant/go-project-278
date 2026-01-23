package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BindJSONStrict decodes JSON and rejects unknown fields and extra objects.
func BindJSONStrict(c *gin.Context, dst any) error {
	if c.Request.Body == nil {
		return io.EOF
	}

	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}

	var extra any
	if err := dec.Decode(&extra); err == nil {
		return errors.New("extra data after JSON object")
	} else if !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}

func badJSON(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "invalid request",
	})
}
