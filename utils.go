package i18n

import (
	"fmt"
	"regexp"
	"strings"
)

// Pre-compiled regex pattern for better performance
var argPattern = regexp.MustCompile(`%[sdvqxXo]`)

// slugify creates a dash-separated key like "hello-%s world" → "hello-0-world".
// This function is optimized for performance with pre-compiled regex.
func slugify(format string) string {
	parts := argPattern.Split(format, -1)
	matches := argPattern.FindAllString(format, -1)
	out := make([]string, 0, len(parts)+len(matches))

	for i, p := range parts {
		if p != "" {
			// Clean the part: remove punctuation, normalize spaces, convert to lowercase
			cleaned := strings.ToLower(p)
			// Replace any non-alphanumeric characters with spaces
			var builder strings.Builder
			for _, r := range cleaned {
				if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
					builder.WriteRune(r)
				} else {
					builder.WriteRune(' ')
				}
			}
			cleaned = builder.String()
			// Replace multiple spaces with single space and trim
			cleaned = strings.TrimSpace(cleaned)
			for strings.Contains(cleaned, "  ") {
				cleaned = strings.ReplaceAll(cleaned, "  ", " ")
			}
			// Replace spaces with dashes
			if cleaned != "" {
				cleaned = strings.ReplaceAll(cleaned, " ", "-")
				out = append(out, cleaned)
			}
		}
		if i < len(matches) {
			out = append(out, fmt.Sprintf("%d", i)) // 0-based indexing
		}
	}

	key := strings.Join(out, "-")
	// Clean up any double dashes that might have been created
	for strings.Contains(key, "--") {
		key = strings.ReplaceAll(key, "--", "-")
	}
	return strings.Trim(key, "-")
}

// normalize replaces printf-style tokens with numbered placeholders {0}, {1}, …
func normalize(format string) (string, []string) {
	matches := argPattern.FindAllString(format, -1)
	counter := 0
	out := argPattern.ReplaceAllStringFunc(format, func(_ string) string {
		placeholder := fmt.Sprintf("{%d}", counter)
		counter++
		return placeholder
	})
	return out, matches
}

// determinePluralForm determines the appropriate plural form based on locale and count
func determinePluralForm(locale string, count int) string {
	// Simplified plural rules for common languages
	// In a production system, you'd want to use a proper CLDR implementation
	switch locale {
	case "en", "de", "it", "es", "pt":
		// Germanic and Romance languages (simplified)
		if count == 0 {
			return "zero"
		} else if count == 1 {
			return "one"
		} else {
			return "other"
		}
	case "fr":
		// French: 0 is zero, 1 is singular, others are plural
		if count == 0 {
			return "zero"
		} else if count == 1 {
			return "one"
		} else {
			return "other"
		}
	case "ru", "uk", "be":
		// Slavic languages (simplified)
		if count == 0 {
			return "zero"
		} else if count == 1 {
			return "one"
		} else if count >= 2 && count <= 4 {
			return "few"
		} else {
			return "many"
		}
	case "pl":
		// Polish (simplified)
		if count == 0 {
			return "zero"
		} else if count == 1 {
			return "one"
		} else if count >= 2 && count <= 4 {
			return "few"
		} else {
			return "many"
		}
	case "ar":
		// Arabic (simplified)
		if count == 0 {
			return "zero"
		} else if count == 1 {
			return "one"
		} else if count == 2 {
			return "two"
		} else if count >= 3 && count <= 10 {
			return "few"
		} else if count >= 11 && count <= 99 {
			return "many"
		} else {
			return "other"
		}
	default:
		// Default English-like rules
		if count == 0 {
			return "zero"
		} else if count == 1 {
			return "one"
		} else {
			return "other"
		}
	}
}

// extractPluralForm extracts the appropriate plural form from an ICU-style template
func extractPluralForm(template, form string, count int) string {
	// Look for the pattern: "form {content}"
	start := fmt.Sprintf("%s {", form)
	idx := strings.Index(template, start)
	if idx == -1 {
		return ""
	}

	// Find the matching closing brace
	content := template[idx+len(start):]
	braceCount := 1
	end := 0

	for i, r := range content {
		if r == '{' {
			braceCount++
		} else if r == '}' {
			braceCount--
			if braceCount == 0 {
				end = i
				break
			}
		}
	}

	if end == 0 {
		return ""
	}

	result := content[:end]
	// Replace # with the actual count
	result = strings.ReplaceAll(result, "#", fmt.Sprint(count))
	return strings.TrimSpace(result)
}
