package i18n

import (
	"fmt"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"Hello %s", "hello-0"},
		{"Hello %s World", "hello-0-world"},
		{"Welcome %s, you have %d messages", "welcome-0-you-have-1-messages"},
		{"Simple text", "simple-text"},
		{"Text with   multiple   spaces", "text-with-multiple-spaces"},
		{"UPPERCASE TEXT", "uppercase-text"},
		{"Mixed Case Text", "mixed-case-text"},
		{"Text-with-dashes", "text-with-dashes"},
		{"", ""},
		{"%s", "0"},
		{"%s%s%s", "0-1-2"},
		{"Start %s middle %d end", "start-0-middle-1-end"},
		{"No placeholders here", "no-placeholders-here"},
		{"Hello %v world", "hello-0-world"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := slugify(tt.input)
			if result != tt.expected {
				t.Errorf("slugify(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		input           string
		expectedOutput  string
		expectedMatches []string
	}{
		{
			"Hello %s",
			"Hello {0}",
			[]string{"%s"},
		},
		{
			"Hello %s World",
			"Hello {0} World",
			[]string{"%s"},
		},
		{
			"Welcome %s, you have %d messages",
			"Welcome {0}, you have {1} messages",
			[]string{"%s", "%d"},
		},
		{
			"No placeholders",
			"No placeholders",
			[]string{},
		},
		{
			"",
			"",
			[]string{},
		},
		{
			"%s%s%s",
			"{0}{1}{2}",
			[]string{"%s", "%s", "%s"},
		},
		{
			"Start %s middle %d end %v",
			"Start {0} middle {1} end {2}",
			[]string{"%s", "%d", "%v"},
		},
		{
			"Mixed %v and %s types",
			"Mixed {0} and {1} types",
			[]string{"%v", "%s"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			output, matches := normalize(tt.input)

			if output != tt.expectedOutput {
				t.Errorf("normalize(%q) output = %q, expected %q", tt.input, output, tt.expectedOutput)
			}

			if len(matches) != len(tt.expectedMatches) {
				t.Errorf("normalize(%q) matches length = %d, expected %d",
					tt.input, len(matches), len(tt.expectedMatches))
			} else {
				for i, match := range matches {
					if match != tt.expectedMatches[i] {
						t.Errorf("normalize(%q) match[%d] = %q, expected %q",
							tt.input, i, match, tt.expectedMatches[i])
					}
				}
			}
		})
	}
}

func TestSlugifyNormalizeConsistency(t *testing.T) {
	// Test that slugify and normalize work together correctly
	testCases := []string{
		"Hello %s",
		"Hello %s World",
		"Welcome %s, you have %d messages",
		"No placeholders",
		"Multiple %s %d %v placeholders",
	}

	for _, input := range testCases {
		t.Run(input, func(t *testing.T) {
			// Slugify should create a key that corresponds to the normalized template
			key := slugify(input)
			normalized, matches := normalize(input)

			// The key should reflect the number of placeholders
			placeholderCount := len(matches)

			// Count the number of numbered placeholders in the key
			keyPlaceholderCount := 0
			for i := 0; i < placeholderCount; i++ {
				if contains(key, string(rune('0'+i))) {
					keyPlaceholderCount++
				}
			}

			// For non-zero placeholder counts, verify they match
			if placeholderCount > 0 {
				if keyPlaceholderCount != placeholderCount {
					t.Errorf("Inconsistency for %q: key has %d placeholder indicators, normalized has %d placeholders",
						input, keyPlaceholderCount, placeholderCount)
				}
			}

			// Ensure normalized template has correct placeholder format
			for i := 0; i < placeholderCount; i++ {
				expectedPlaceholder := "{" + string(rune('0'+i)) + "}"
				if !contains(normalized, expectedPlaceholder) {
					t.Errorf("Normalized template %q missing expected placeholder %q",
						normalized, expectedPlaceholder)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func BenchmarkSlugify(b *testing.B) {
	testCases := []string{
		"Hello World",
		"Hello %s World",
		"Welcome %s, you have %d messages and %v notifications",
		"Simple text without placeholders",
		"UPPERCASE TEXT WITH MULTIPLE WORDS",
	}

	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			slugify(tc)
		}
	}
}

func BenchmarkNormalize(b *testing.B) {
	testCases := []string{
		"Hello World",
		"Hello %s World",
		"Welcome %s, you have %d messages and %v notifications",
		"Simple text without placeholders",
		"%s%s%s%s%s",
	}

	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			normalize(tc)
		}
	}
}

func TestDeterminePluralForm(t *testing.T) {
	tests := []struct {
		locale   string
		count    int
		expected string
	}{
		// English
		{"en", 0, "zero"},
		{"en", 1, "one"},
		{"en", 2, "other"},
		{"en", 5, "other"},

		// French (0 and 1 are singular)
		{"fr", 0, "zero"},
		{"fr", 1, "one"},
		{"fr", 2, "other"},
		{"fr", 5, "other"},

		// Russian (more complex rules)
		{"ru", 0, "zero"},
		{"ru", 1, "one"},
		{"ru", 2, "few"},
		{"ru", 3, "few"},
		{"ru", 4, "few"},
		{"ru", 5, "many"},
		{"ru", 10, "many"},

		// Arabic (even more complex)
		{"ar", 0, "zero"},
		{"ar", 1, "one"},
		{"ar", 2, "two"},
		{"ar", 3, "few"},
		{"ar", 5, "few"},
		{"ar", 10, "few"},
		{"ar", 11, "many"},
		{"ar", 50, "many"},
		{"ar", 99, "many"},
		{"ar", 100, "other"},

		// Default rules for unknown locale
		{"unknown", 0, "zero"},
		{"unknown", 1, "one"},
		{"unknown", 2, "other"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%d", tt.locale, tt.count), func(t *testing.T) {
			result := determinePluralForm(tt.locale, tt.count)
			if result != tt.expected {
				t.Errorf("determinePluralForm(%q, %d) = %q, expected %q",
					tt.locale, tt.count, result, tt.expected)
			}
		})
	}
}

func TestExtractPluralForm(t *testing.T) {
	tests := []struct {
		template string
		form     string
		count    int
		expected string
	}{
		{
			"{count, plural, zero {no items} one {# item} other {# items}}",
			"zero",
			0,
			"no items",
		},
		{
			"{count, plural, zero {no items} one {# item} other {# items}}",
			"one",
			1,
			"1 item",
		},
		{
			"{count, plural, zero {no items} one {# item} other {# items}}",
			"other",
			5,
			"5 items",
		},
		{
			"{count, plural, one {# message} other {# messages}}",
			"one",
			1,
			"1 message",
		},
		{
			"{count, plural, one {# message} other {# messages}}",
			"other",
			3,
			"3 messages",
		},
		{
			"{count, plural, few {# items} many {# items} other {# items}}",
			"few",
			3,
			"3 items",
		},
		{
			"{count, plural, few {# items} many {# items} other {# items}}",
			"many",
			10,
			"10 items",
		},
		{
			// Nested braces
			"{count, plural, one {You have {#} item} other {You have {#} items}}",
			"one",
			1,
			"You have {1} item",
		},
		{
			// Form not found
			"{count, plural, one {# item} other {# items}}",
			"few",
			3,
			"",
		},
		{
			// No plural syntax
			"Simple template with {count}",
			"other",
			5,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s_%d", tt.template[:min(20, len(tt.template))], tt.form, tt.count), func(t *testing.T) {
			result := extractPluralForm(tt.template, tt.form, tt.count)
			if result != tt.expected {
				t.Errorf("extractPluralForm(%q, %q, %d) = %q, expected %q",
					tt.template, tt.form, tt.count, result, tt.expected)
			}
		})
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
