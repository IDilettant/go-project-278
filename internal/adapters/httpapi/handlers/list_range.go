package handlers

import (
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"code/internal/adapters/httpapi/problems"
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

	query, err := pageQueryFromRange(rng)
	if err != nil {
		return Range{}, PageQuery{}, false, err
	}

	return rng, query, true, nil
}

func pageQueryFromRange(rng Range) (PageQuery, error) {
	if rng.Start > math.MaxInt32 || rng.Count > math.MaxInt32 {
		return PageQuery{}, errInvalidRange
	}

	return PageQuery{
		Offset: int32(rng.Start),
		Limit:  int32(rng.Count),
	}, nil
}

func writeInvalidRange(c *gin.Context) {
	problems.WriteProblem(c, problems.Problem{
		Type:   problems.ProblemTypeValidation,
		Title:  problems.TitleValidation,
		Status: http.StatusBadRequest,
		Detail: problems.DetailInvalidRange, // expected [start,count], [start,end], or start-end
	})
}

func writeInvalidSort(c *gin.Context) {
	problems.WriteProblem(c, problems.Problem{
		Type:   problems.ProblemTypeValidation,
		Title:  problems.TitleValidation,
		Status: http.StatusBadRequest,
		Detail: problems.DetailInvalidSort,
	})
}
