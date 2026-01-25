package handlers_test

import (
	"testing"

	"code/internal/adapters/httpapi/handlers"
)

func TestParseRangeParam(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    handlers.Range
		wantErr bool
	}{
		{name: "empty", raw: "", wantErr: true},
		{name: "spaces", raw: "   ", wantErr: true},
		{name: "valid", raw: "[0,10]", want: handlers.Range{Start: 0, Count: 10}},
		{name: "valid_with_spaces", raw: "[ 5 , 10 ]", want: handlers.Range{Start: 5, Count: 10}},
		{name: "valid_with_resource_prefix", raw: "links=[0,10]", want: handlers.Range{Start: 0, Count: 10}},
		{name: "header_style", raw: "0-49", want: handlers.Range{Start: 0, Count: 50}},
		{name: "header_style_single", raw: "0-0", want: handlers.Range{Start: 0, Count: 1}},
		{name: "header_style_with_resource", raw: "links=0-49", want: handlers.Range{Start: 0, Count: 50}},
		{name: "header_style_with_resource_spaces", raw: "links = 1-2", want: handlers.Range{Start: 1, Count: 2}},
		{name: "invalid_text", raw: "bad", wantErr: true},
		{name: "invalid_empty_brackets", raw: "[]", wantErr: true},
		{name: "invalid_single_value", raw: "[0]", wantErr: true},
		{name: "invalid_extra_values", raw: "[0,10,20]", wantErr: true},
		{name: "invalid_header_empty_start", raw: "-10", wantErr: true},
		{name: "invalid_header_empty_end", raw: "10-", wantErr: true},
		{name: "invalid_header_extra_dash", raw: "0-1-2", wantErr: true},
		{name: "invalid_header_negative_start", raw: "-1-10", wantErr: true},
		{name: "invalid_header_end_before_start", raw: "10-9", wantErr: true},
		{name: "invalid_header_non_numeric_end", raw: "0-foo", wantErr: true},
		{name: "invalid_header_limit_too_large", raw: "0-10000", wantErr: true},
		{name: "non_numeric_start", raw: "[a,10]", wantErr: true},
		{name: "non_numeric_end", raw: "[0,b]", wantErr: true},
		{name: "negative_start", raw: "[-1,5]", wantErr: true},
		{name: "zero_count", raw: "[0,0]", wantErr: true},
		{name: "negative_count", raw: "[5,-1]", wantErr: true},
		{name: "limit_max_ok", raw: "[0,10000]", want: handlers.Range{Start: 0, Count: 10000}},
		{name: "limit_too_large", raw: "[0,10001]", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := handlers.ParseRangeParam(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.want {
				t.Fatalf("unexpected range: got=%+v want=%+v", got, tc.want)
			}
		})
	}
}
