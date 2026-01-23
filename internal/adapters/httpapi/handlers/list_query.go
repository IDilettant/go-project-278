package handlers

import (
	"encoding/json"
	"math"
	"strings"

	"github.com/gin-gonic/gin"
)

func parseRangeFromRequest(c *gin.Context) (Range, PageQuery, bool, error) {
	raw := strings.TrimSpace(c.Query("range"))
	if raw != "" {
		if strings.HasPrefix(raw, "[") {
			return parseRangeQuery(raw)
		}

		return parseListRange(raw)
	}

	return parseListRange(strings.TrimSpace(c.GetHeader("Range")))
}

func parseRangeQuery(raw string) (Range, PageQuery, bool, error) {
	var values []int64
	if err := json.Unmarshal([]byte(raw), &values); err != nil || len(values) != 2 {
		return Range{}, PageQuery{}, false, errInvalidRange
	}

	start := values[0]
	end := values[1]
	
	if start < 0 || end < start {
		return Range{}, PageQuery{}, false, errInvalidRange
	}

	count := end - start + 1
	if count <= 0 || count > int64(maxRangeLimit) {
		return Range{}, PageQuery{}, false, errInvalidRange
	}

	if start > math.MaxInt32 || count > math.MaxInt32 {
		return Range{}, PageQuery{}, false, errInvalidRange
	}

	rng := Range{Start: int(start), Count: int(count)}
	
	query, err := pageQueryFromRange(rng)
	if err != nil {
		return Range{}, PageQuery{}, false, err
	}

	return rng, query, true, nil
}
