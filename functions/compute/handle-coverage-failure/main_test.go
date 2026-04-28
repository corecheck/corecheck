package main

import "testing"

func TestSummarizeFailureReason(t *testing.T) {
	tests := []struct {
		name     string
		failure  *coverageFailure
		expected string
	}{
		{
			name: "lambda error message",
			failure: &coverageFailure{
				Cause: `{"errorMessage":"coverage.json was not uploaded"}`,
			},
			expected: "coverage.json was not uploaded",
		},
		{
			name: "batch status reason",
			failure: &coverageFailure{
				Cause: `{"StatusReason":"Essential container in task exited"}`,
			},
			expected: "Essential container in task exited",
		},
		{
			name: "timeout fallback",
			failure: &coverageFailure{
				Error: "States.Timeout",
			},
			expected: "The coverage workflow timed out.",
		},
		{
			name:     "generic fallback",
			failure:  &coverageFailure{},
			expected: defaultFailureReason,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := summarizeFailureReason(test.failure); got != test.expected {
				t.Fatalf("summarizeFailureReason() = %q, want %q", got, test.expected)
			}
		})
	}
}
