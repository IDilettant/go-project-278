package domain

import (
	"net/url"
	"regexp"
	"strings"
)

var shortNameRe = regexp.MustCompile(`^[a-zA-Z0-9]{3,32}$`)

func ValidateOriginalURL(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return ErrInvalidURL
	}

	u, err := url.ParseRequestURI(s)
	if err != nil {
		return ErrInvalidURL
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrInvalidURL
	}

	return nil
}

func ValidateShortName(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return ErrInvalidShortName
	}

	if !shortNameRe.MatchString(s) {
		return ErrInvalidShortName
	}

	return nil
}
