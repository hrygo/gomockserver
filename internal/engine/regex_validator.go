// Package engine provides the rule matching engine for the mock server.
package engine

import (
	"fmt"
	"regexp"
	"strings"
)

// RegexValidationError represents a regex validation error
type RegexValidationError struct {
	Pattern string
	Message string
}

func (e *RegexValidationError) Error() string {
	return fmt.Sprintf("regex validation error for pattern '%s': %s", e.Pattern, e.Message)
}

// validateRegexPattern validates a regex pattern for complexity and security
func validateRegexPattern(pattern string) error {
	// Check pattern length
	if len(pattern) > 1000 {
		return &RegexValidationError{
			Pattern: pattern,
			Message: "pattern too long (max 1000 characters)",
		}
	}

	// Check nesting depth
	nestingDepth := 0
	maxNesting := 0
	for _, char := range pattern {
		switch char {
		case '(', '[', '{':
			nestingDepth++
			if nestingDepth > maxNesting {
				maxNesting = nestingDepth
			}
		case ')', ']', '}':
			if nestingDepth > 0 {
				nestingDepth--
			}
		}
	}

	if maxNesting > 10 {
		return &RegexValidationError{
			Pattern: pattern,
			Message: fmt.Sprintf("nesting too deep (max 10 levels, got %d)", maxNesting),
		}
	}

	// Check for potentially dangerous patterns
	dangerousPatterns := []string{
		".*.*",     // Nested wildcards
		".*+",      // Wildcard followed by +
		".*{",      // Wildcard followed by repetition
		"a?a?a?a?", // Catastrophic backtracking pattern example
	}

	for _, dangerous := range dangerousPatterns {
		if strings.Contains(pattern, dangerous) {
			return &RegexValidationError{
				Pattern: pattern,
				Message: fmt.Sprintf("contains potentially dangerous pattern: %s", dangerous),
			}
		}
	}

	// Check for excessive repetition
	repetitionRegex := regexp.MustCompile(`\{(\d+),(\d+)\}`)
	matches := repetitionRegex.FindAllStringSubmatch(pattern, -1)
	for _, match := range matches {
		if len(match) == 3 {
			min, max := 0, 0
			fmt.Sscanf(match[1], "%d", &min)
			fmt.Sscanf(match[2], "%d", &max)
			if max-min > 1000 {
				return &RegexValidationError{
					Pattern: pattern,
					Message: fmt.Sprintf("excessive repetition range: {%d,%d}", min, max),
				}
			}
		}
	}

	// Check for excessive consecutive quantifiers
	consecutiveQuantifiers := 0
	maxConsecutive := 0
	for _, char := range pattern {
		if char == '*' || char == '+' || char == '?' || char == '{' {
			consecutiveQuantifiers++
			if consecutiveQuantifiers > maxConsecutive {
				maxConsecutive = consecutiveQuantifiers
			}
		} else {
			consecutiveQuantifiers = 0
		}
	}

	if maxConsecutive > 5 {
		return &RegexValidationError{
			Pattern: pattern,
			Message: fmt.Sprintf("too many consecutive quantifiers (max 5, got %d)", maxConsecutive),
		}
	}

	return nil
}