package handlers

import (
	"errors"
	"strconv"
	"strings"

	"code/internal/app/links"
)

const (
	maxRangeLimit = 1000
)

type Range = links.Range

var errInvalidRange = errors.New("invalid range")

// ParseRangeParam parses ranges in formats:
//  1. [start,count]           e.g. [0,10]          (query param style)
//  2. start-end               e.g. 0-49            (HTTP Range header style)
//     with optional prefix:   resource=start-end   e.g. links=0-49
func ParseRangeParam(raw string) (Range, error) {
	raw = canonicalizeRange(raw)
	if raw == "" {
		return Range{}, errInvalidRange
	}

	if isStartCountRange(raw) {
		return parseStartCountRange(raw)
	}

	return parseStartEndRange(raw)
}

func canonicalizeRange(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// "resource=start-end" -> "start-end"
	if i := strings.IndexByte(raw, '='); i >= 0 {
		raw = raw[i+1:]
	}

	return strings.TrimSpace(raw)
}

func isStartCountRange(s string) bool {
	return len(s) >= 2 && s[0] == '[' && s[len(s)-1] == ']'
}

// parseStartCountRange parses format: [start,count]
func parseStartCountRange(raw string) (Range, error) {
	body := strings.TrimSpace(raw[1 : len(raw)-1])

	startStr, countStr, ok := splitOnce(body, ',')
	if !ok {
		return Range{}, errInvalidRange
	}

	start, ok := parseNonNegativeInt(startStr)
	if !ok {
		return Range{}, errInvalidRange
	}

	count, ok := parsePositiveInt(countStr)
	if !ok || count > maxRangeLimit {
		return Range{}, errInvalidRange
	}

	return Range{Start: start, Count: count}, nil
}

// parseStartEndRange parses format: start-end
func parseStartEndRange(raw string) (Range, error) {
	startStr, endStr, ok := splitOnce(raw, '-')
	if !ok {
		return Range{}, errInvalidRange
	}

	start, ok := parseNonNegativeInt(startStr)
	if !ok {
		return Range{}, errInvalidRange
	}

	end, ok := parseNonNegativeInt(endStr)
	if !ok || end < start {
		return Range{}, errInvalidRange
	}

	count := end - start + 1
	if count <= 0 || count > maxRangeLimit {
		return Range{}, errInvalidRange
	}

	return Range{Start: start, Count: count}, nil
}

// splitOnce splits s by the first occurrence of sep, trims both sides,
// and requires both parts to be non-empty.
func splitOnce(s string, sep byte) (string, string, bool) {
	i := strings.IndexByte(s, sep)
	if i < 0 {
		return "", "", false
	}

	left := strings.TrimSpace(s[:i])
	right := strings.TrimSpace(s[i+1:])

	if left == "" || right == "" {
		return "", "", false
	}

	return left, right, true
}

func parseNonNegativeInt(s string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	return n, err == nil && n >= 0
}

func parsePositiveInt(s string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	return n, err == nil && n > 0
}
