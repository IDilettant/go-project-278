package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"code/internal/domain"
)

func TestValidateShortName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		ok   bool
	}{
		{"ok/min_len", "abcd", true},
		{"ok/mixed_case_digits", "aBcD12", true},
		{"ok/max_len_32", strings.Repeat("a", 32), true},

		{"bad/empty", "", false},
		{"bad/spaces_only", "   ", false},
		{"bad/too_short_3", "abc", false},
		{"bad/too_long_33", strings.Repeat("a", 33), false},
		{"bad/space_inside", "ab cd", false},
		{"bad/slash", "ab/cd", false},
		{"bad/dash_forbidden", "ab-cd", false},
		{"bad/underscore_forbidden", "ab_cd", false},
		{"bad/unicode", "тест", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateShortName(tc.in)
			if tc.ok {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestValidateOriginalURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		ok   bool
	}{
		{"ok/https", "https://example.com", true},
		{"ok/http", "http://example.com", true},
		{"ok/with_path_query_fragment", "https://example.com/a/b?x=1#y", true},

		{"bad/empty", "", false},
		{"bad/space", " ", false},
		{"bad/not_url", "not-a-url", false},
		{"bad/missing_scheme", "example.com", false},
		{"bad/ftp", "ftp://example.com", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateOriginalURL(tc.in)
			if tc.ok {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
