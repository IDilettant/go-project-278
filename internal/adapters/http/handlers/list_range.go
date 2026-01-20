package handlers

import (
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"code/internal/adapters/http/problems"
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

func rangeValue(c *gin.Context) string {
	raw := strings.TrimSpace(c.GetHeader("Range"))
	if raw != "" {
		return raw
	}

	return c.Query("range")
}

func writeInvalidRange(c *gin.Context) {
	problems.WriteProblem(c, problems.Problem{
		Type:   problems.ProblemTypeValidation,
		Title:  problems.TitleValidation,
		Status: http.StatusBadRequest,
		Detail: problems.DetailInvalidRange, // expected [start,count]
	})
}
