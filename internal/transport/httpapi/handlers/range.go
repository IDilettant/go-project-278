package handlers

import (
	"errors"
	"strconv"
	"strings"
)

const (
	maxRangeLimit = 1000
)

type Range struct {
	Start int
	Count int
}

var errInvalidRange = errors.New("invalid range")

// ParseRangeParam parses ranges in the format [start,count].
func ParseRangeParam(raw string) (Range, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) < 5 || raw[0] != '[' || raw[len(raw)-1] != ']' {
		return Range{}, errInvalidRange
	}

	body := strings.TrimSpace(raw[1 : len(raw)-1])
	parts := strings.Split(body, ",")

	if len(parts) != 2 {
		return Range{}, errInvalidRange
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return Range{}, errInvalidRange
	}

	count, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return Range{}, errInvalidRange
	}

	if start < 0 || count <= 0 {
		return Range{}, errInvalidRange
	}

	if count > maxRangeLimit {
		return Range{}, errInvalidRange
	}

	return Range{Start: start, Count: count}, nil
}
