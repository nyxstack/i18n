// Package i18n provides internationalization and localization functionality.
//
// Translation API Functions:
// - T(key, args...) - Translate by key with placeholder substitution
// - F(format, args...) - Translate by format string (auto-generates key from format)
// - S(text) - Translate static text (auto-generates key from text)
// - P(key, count) - Pluralization support
// - R(locale, format) - Direct translation (no function wrapping)
//
// Example usage:
//
//	greeting := i18n.T("hello_world")
//	msg := i18n.F("Welcome %s!", "John")
//	title := i18n.S("Dashboard")
package i18n

import (
	"fmt"
	"strings"
)

// TranslatedFunc returns a localized string when called with a locale.
// This allows you to prepare a translation function and call it later with different locales.
type TranslatedFunc func(locale string) string

// T translates by exact key with placeholder substitution.
// Use this when you have predefined translation keys in your dictionary files.
// Placeholders are numbered: {0}, {1}, {2}, etc.
//
// Example:
//
//	fn := i18n.T("welcome_user", "John")
//	fmt.Println(fn("en")) // "Welcome John!"
//	fmt.Println(fn("fr")) // "Bienvenue John!"
//
// Dictionary should contain:
//
//	"welcome_user": "Welcome {0}!"
func T(key string, args ...any) TranslatedFunc {
	return func(locale string) string {
		dict := GetDictionary(locale)
		template := key

		if dict != nil {
			if tr := dict.Get(key); tr != "" && tr != key {
				template = tr
			}
		} else if defaultDict := GetDictionary(DefaultLanguage()); defaultDict != nil {
			if tr := defaultDict.Get(key); tr != "" && tr != key {
				template = tr
			}
		}

		// Replace placeholders {0}, {1}, {2}, etc.
		for i, arg := range args {
			placeholder := fmt.Sprintf("{%d}", i)
			template = strings.ReplaceAll(template, placeholder, fmt.Sprint(arg))
		}

		return template
	}
}

// F translates by format string with auto-generated key.
// This automatically generates a translation key from the format string and normalizes placeholders.
// Use this when you want to use the English text as the source and auto-generate keys.
//
// Example:
//
//	fn := i18n.F("Hello %s, you have %d messages", "John", 5)
//	fmt.Println(fn("en")) // "Hello John, you have 5 messages"
//	fmt.Println(fn("fr")) // "Bonjour John, vous avez 5 messages"
//
// Auto-generated key: "hello-1-you-have-2-messages"
// Dictionary should contain:
//
//	"hello-1-you-have-2-messages": "Bonjour {0}, vous avez {1} messages"
func F(format string, args ...any) TranslatedFunc {
	key := slugify(format)
	normalizedTemplate, _ := normalize(format)

	return func(locale string) string {
		dict := GetDictionary(locale)
		template := normalizedTemplate

		if dict != nil {
			if tr := dict.Get(key); tr != "" && tr != key {
				template = tr
			}
		} else if defaultDict := GetDictionary(DefaultLanguage()); defaultDict != nil {
			if tr := defaultDict.Get(key); tr != "" && tr != key {
				template = tr
			}
		}

		// Replace placeholders {0}, {1}, {2}, etc.
		for i, arg := range args {
			placeholder := fmt.Sprintf("{%d}", i)
			template = strings.ReplaceAll(template, placeholder, fmt.Sprint(arg))
		}

		return template
	}
}

// S translates static text with auto-generated key.
// Use this for simple static strings without placeholders.
//
// Example:
//
//	fn := i18n.S("Dashboard")
//	fmt.Println(fn("en")) // "Dashboard"
//	fmt.Println(fn("fr")) // "Tableau de bord"
//
// Auto-generated key: "dashboard"
// Dictionary should contain:
//
//	"dashboard": "Tableau de bord"
func S(text string) TranslatedFunc {
	key := slugify(text)

	return func(locale string) string {
		dict := GetDictionary(locale)

		if dict != nil {
			if tr := dict.Get(key); tr != "" && tr != key {
				return tr
			}
		}

		if defaultDict := GetDictionary(DefaultLanguage()); defaultDict != nil {
			if tr := defaultDict.Get(key); tr != "" && tr != key {
				return tr
			}
		}

		return text
	}
}

// P handles pluralization for a given key and count.
// Supports ICU-style plural forms: zero, one, two, few, many, other.
//
// Example:
//
//	fn := i18n.P("item_count", 5)
//	fmt.Println(fn("en")) // "5 items"
//
// Dictionary should contain:
//
//	"item_count": "{count, plural, zero {no items} one {# item} other {# items}}"
func P(key string, count int) TranslatedFunc {
	return func(locale string) string {
		dict := GetDictionary(locale)
		template := key

		if dict != nil {
			template = dict.Get(key)
		} else if defaultDict := GetDictionary(DefaultLanguage()); defaultDict != nil {
			template = defaultDict.Get(key)
		}

		// Handle ICU-style plural syntax
		if strings.Contains(template, "{count, plural") {
			// Determine the appropriate plural form for the locale
			form := determinePluralForm(locale, count)

			// Extract the appropriate plural form from template
			if result := extractPluralForm(template, form, count); result != "" {
				return result
			}

			// Fallback to "other" if specific form not found
			if form != "other" {
				if result := extractPluralForm(template, "other", count); result != "" {
					return result
				}
			}
		}

		// Fallback: simple string substitution
		return strings.ReplaceAll(template, "{count}", fmt.Sprint(count))
	}
}

// R performs direct translation without function wrapping.
// Use this when you want immediate translation without creating a TranslatedFunc.
//
// Example:
//
//	text := i18n.R("en", "Dashboard")
//	fmt.Println(text) // "Dashboard"
func R(locale, text string) string {
	key := slugify(text)
	dict := GetDictionary(locale)

	if dict != nil {
		if tr := dict.Get(key); tr != "" && tr != key {
			return tr
		}
	}

	if defaultDict := GetDictionary(DefaultLanguage()); defaultDict != nil {
		if tr := defaultDict.Get(key); tr != "" && tr != key {
			return tr
		}
	}

	return text
}
