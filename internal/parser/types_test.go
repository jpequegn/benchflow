package parser

import (
	"strings"
	"testing"
)

func TestParseError_Error(t *testing.T) {
	tests := []struct {
		name         string
		err          *ParseError
		wantContains string
	}{
		{
			name: "error with line number",
			err: &ParseError{
				Line:    42,
				Message: "invalid format",
				Input:   "test input",
			},
			wantContains: "at line",
		},
		{
			name: "error without line number",
			err: &ParseError{
				Line:    0,
				Message: "no benchmarks found",
				Input:   "",
			},
			wantContains: "no benchmarks found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("Error() = %v, want to contain %v", got, tt.wantContains)
			}
		})
	}
}
