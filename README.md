# Nyx i18n - Go Internationalization Library

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/nyxstack/i18n)](https://goreportcard.com/report/github.com/nyxstack/i18n)
[![GoDoc](https://godoc.org/github.com/nyxstack/i18n?status.svg)](https://godoc.org/github.com/nyxstack/i18n)

A simple, efficient Go internationalization (i18n) library with 0-based indexing and file-based dictionaries.

## Quick Start

### 1. Write Code with i18n Functions

```go
import "github.com/nyxstack/i18n"

func myFunction() {
    // Use simple single-character i18n functions
    greeting := i18n.F("Hello %s", "World")   // Format with placeholders
    title := i18n.S("Dashboard")              // Static text
    bye := i18n.T("goodbye")                  // Direct key lookup
    
    // Get translations for different locales
    fmt.Println(greeting("en"))  // "Hello World"
    fmt.Println(greeting("fr"))  // "Bonjour World"
    fmt.Println(title("fr"))     // "Tableau de bord"
}
```

### 2. Extract Translations

**Option A: Programmatically**
```go
// Generate translation file from your Go source code
err := i18n.GenerateTranslations("en", "./", "")
// Output: ✅ Extracted 2 i18n entries → locales/default.en.json
```

**Option B: CLI**
```bash
go run github.com/nyxstack/i18n/cmd/extract-i18n@latest <source_dir> <locale>
```

### 3. Create Translation Files

Generation creates `locales/default.en.json`:
```json
{
  "meta": {
    "lang": "en",
    "name": "default"
  },
  "translations": {
    "hello-0": "Hello %s",
    "dashboard": "Dashboard"
  }
}
```

Create `locales/default.fr.json` for French:
```json
{
  "meta": {
    "lang": "fr", 
    "name": "default"
  },
  "translations": {
    "hello-0": "Bonjour %s",
    "dashboard": "Tableau de bord"
  }
}
```

### 4. Load and Use

```go
func main() {
    // Load dictionaries
    i18n.Load()              // Loads locales/default.en.json
    i18n.LoadLanguage("fr")  // Loads locales/default.fr.json
    i18n.SetDefaultLanguage("en")
    
    // Use your translations
    greeting := i18n.F("Hello %s", "World")
    fmt.Println(greeting("fr"))  // "Bonjour World"
}
```

## Translation Functions

All functions return `func(locale string) string` for easy use:

| Function | Purpose | Example |
|----------|---------|---------|
| `F(format, args...)` | Format string with placeholders | `F("Hello %s", "John")` |
| `S(text)` | Static text | `S("Dashboard")` |
| `T(key, args...)` | Direct key with placeholders | `T("welcome", "John")` |
| `P(key, count)` | Pluralization | `P("item_count", 5)` |

## Dictionary Management

Dictionaries are JSON files that contain your translations. Each file represents one language:

```json
{
  "meta": {
    "lang": "en",           // Required: language code
    "name": "default"       // Required: dictionary name
  },
  "translations": {
    "hello-0": "Hello {0}!",           // 0-based placeholders
    "welcome": "Welcome",              // Static text
    "item-count": "{count, plural, one {# item} other {# items}}"  // Plurals
  }
}
```

Load and manage dictionaries:

```go
// Load dictionaries
i18n.Load()                    // Loads locales/default.en.json
i18n.LoadLanguage("fr")        // Loads locales/default.fr.json
i18n.LoadFrom("custom.json")   // Load custom path
i18n.SetDefaultLanguage("en")  // Set fallback language

// Create dictionaries programmatically
dict := i18n.NewDictionary("es")
dict.Add("hello-0", "Hola {0}!")
dict.AddAll(map[string]string{
    "welcome": "Bienvenido",
    "goodbye": "Adiós",
})
i18n.Register(dict)  // Make it available
```

## License

MIT

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

**Requirements:**
- All new code must have tests
- Tests must pass
- Follow existing code style
- Update documentation if needed