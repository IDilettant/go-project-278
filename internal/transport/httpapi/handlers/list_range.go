package handlers

import (
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type PageQuery struct {
	Offset int32
	Limit  int32
}

func parseListRange(raw string) (Range, PageQuery, bool, error) {
	if strings.TrimSpace(raw) == "" {
		return Range{}, PageQuery{}, false, nil
	}

	rng, err := ParseRangeParam(raw)
	if err != nil {
		return Range{}, PageQuery{}, false, err
	}

	if rng.Start > math.MaxInt32 || rng.Count > math.MaxInt32 {
		return Range{}, PageQuery{}, false, errInvalidRange
	}

	return rng, PageQuery{
		Offset: int32(rng.Start),
		Limit:  int32(rng.Count),
	}, true, nil
}

func writeInvalidRange(c *gin.Context) {
	writeProblem(c, Problem{
		Type:   ProblemTypeValidation,
		Title:  validationTitle,
		Status: http.StatusBadRequest,
		Detail: "invalid range", // expected [start,count]
	})
}
