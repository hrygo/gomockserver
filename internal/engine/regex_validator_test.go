package engine

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRegexPattern(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		expectError bool
	}{
		{
			name:        "Valid simple pattern",
			pattern:     "test.*pattern",
			expectError: false,
		},
		{
			name:        "Valid complex pattern",
			pattern:     "(test|example)-[0-9]+",
			expectError: false,
		},
		{
			name:        "Pattern too long",
			pattern:     strings.Repeat("a", 1001),
			expectError: true,
		},
		{
			name:        "Nesting too deep",
			pattern:     "(((((((((((nested)))))))))))",
			expectError: true,
		},
		{
			name:        "Dangerous nested wildcards",
			pattern:     "a.*.*b",
			expectError: true,
		},
		{
			name:        "Dangerous wildcard plus",
			pattern:     "a.*+b",
			expectError: true,
		},
		{
			name:        "Dangerous wildcard repetition",
			pattern:     "a.*{1,10}b",
			expectError: true,
		},
		{
			name:        "Catastrophic backtracking",
			pattern:     "(a?a?a?a?a?a?aaaa)",
			expectError: true,
		},
		{
			name:        "Excessive repetition range",
			pattern:     "a{1,2000}b",
			expectError: true,
		},
		{
			name:        "High quantifier count",
			pattern:     "******", // 6 consecutive quantifiers
			expectError: true,
		},
		{
			name:        "Valid repetition range",
			pattern:     "a{1,100}b",
			expectError: false,
		},
		{
			name:        "Valid quantifier count",
			pattern:     "a*a*a*a*",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegexPattern(tt.pattern)
			if tt.expectError {
				assert.Error(t, err)
				// Check that it's the right type of error
				_, ok := err.(*RegexValidationError)
				assert.True(t, ok, "Error should be of type RegexValidationError")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegexValidationError(t *testing.T) {
	err := &RegexValidationError{
		Pattern: "test.*pattern",
		Message: "test error",
	}
	
	expected := "regex validation error for pattern 'test.*pattern': test error"
	assert.Equal(t, expected, err.Error())
}