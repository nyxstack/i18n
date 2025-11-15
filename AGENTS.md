# AGENTS.md - Quick Reference for AI Agents

## Package Overview

`github.com/nyxstack/i18n` - A Go internationalization library with 0-based indexing and file-based dictionaries.

## Core API Functions

All functions return `func(locale string) string` for deferred execution:

```go
// Format with placeholders (auto-generates key from format)
greeting := i18n.F("Hello %s", "World")  
fmt.Println(greeting("en"))  // "Hello World"
fmt.Println(greeting("fr"))  // "Bonjour World"

// Static text (auto-generates key from text)
title := i18n.S("Dashboard")
fmt.Println(title("fr"))     // "Tableau de bord"

// Direct key lookup with placeholders  
msg := i18n.T("welcome_user", "John")
fmt.Println(msg("en"))       // "Welcome John!" (from dictionary)

// Pluralization with ICU-style rules
count := i18n.P("item_count", 5)
fmt.Println(count("en"))     // "5 items"

// Direct translation (no function wrapping)
text := i18n.R("fr", "Dashboard")  // "Tableau de bord"
```

## Setup Pattern

```go
// 1. Load dictionaries
i18n.Load()                    // locales/default.en.json
i18n.LoadLanguage("fr")        // locales/default.fr.json  
i18n.SetDefaultLanguage("en")  // fallback language

// 2. Use translations
greeting := i18n.F("Hello %s", name)
return greeting(userLocale)
```

## Translation File Format

```json
{
  "meta": {
    "lang": "en",
    "name": "default"
  },
  "translations": {
    "hello-0": "Hello {0}!",
    "dashboard": "Dashboard",
    "item-count": "{count, plural, one {# item} other {# items}}"
  }
}
```

## Key Generation Rules

- `F("Hello %s", x)` → key: `"hello-0"`, template: `"Hello {0}"`
- `S("Dashboard")` → key: `"dashboard"`  
- `T("custom_key", x)` → uses exact key: `"custom_key"`
- `P("item_count", n)` → uses exact key: `"item_count"`

## Code Generation

```go
// Extract translation keys from Go source
err := i18n.GenerateTranslations("en", "./src", "")
// Scans for i18n.F(), i18n.S(), i18n.T(), i18n.P() calls
```

## Pluralization Support

Supports ICU-style forms: `zero`, `one`, `two`, `few`, `many`, `other`

Built-in rules for: English, French, Russian, Polish, Arabic, German, Italian, Spanish

## Thread Safety

All operations are thread-safe with internal mutex protection.

## Error Handling

- Missing translations return the original key/text
- Invalid files return descriptive errors
- Automatic fallback to default language
- File validation includes JSON structure and ICU plural syntax

## Performance Notes

- Regex patterns pre-compiled for better performance
- Dictionary lookups are O(1) hash map operations  
- Thread-safe but optimized for read-heavy workloads