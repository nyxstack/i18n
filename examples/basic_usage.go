package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nyxstack/i18n"
)

func createTestTranslations() {
	// Create locales directory
	if err := os.MkdirAll("locales", 0755); err != nil {
		log.Fatal(err)
	}

	// Create English translations
	enContent := `{
  "meta": {
    "lang": "en",
    "name": "default"
  },
  "translations": {
    "hello-0": "Hello {0}!",
    "welcome": "Welcome",
    "dashboard": "Dashboard",
    "item-count": "{count, plural, zero {no items} one {# item} other {# items}}",
    "message-count": "{count, plural, zero {no messages} one {# message} other {# messages}}"
  }
}`

	// Create French translations
	frContent := `{
  "meta": {
    "lang": "fr",
    "name": "default"
  },
  "translations": {
    "hello-0": "Bonjour {0}!",
    "welcome": "Bienvenue",
    "dashboard": "Tableau de bord",
    "item-count": "{count, plural, zero {aucun élément} one {# élément} other {# éléments}}",
    "message-count": "{count, plural, zero {aucun message} one {# message} other {# messages}}"
  }
}`

	// Write translation files
	if err := os.WriteFile(filepath.Join("locales", "default.en.json"), []byte(enContent), 0644); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("locales", "default.fr.json"), []byte(frContent), 0644); err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Create test translations
	createTestTranslations()

	// Load dictionaries
	if err := i18n.Load(); err != nil {
		log.Fatal(err)
	}
	if err := i18n.LoadLanguage("fr"); err != nil {
		log.Fatal(err)
	}

	// Set default language
	i18n.SetDefaultLanguage("en")

	fmt.Println("=== Basic Translation Functions ===")

	// F() - Format with placeholders (auto-generated keys)
	greeting := i18n.F("Hello %s", "World")
	fmt.Printf("F() English: %s\n", greeting("en"))
	fmt.Printf("F() French:  %s\n", greeting("fr"))

	// S() - Static text (auto-generated keys)
	welcome := i18n.S("Welcome")
	fmt.Printf("S() English: %s\n", welcome("en"))
	fmt.Printf("S() French:  %s\n", welcome("fr"))

	dashboard := i18n.S("Dashboard")
	fmt.Printf("S() English: %s\n", dashboard("en"))
	fmt.Printf("S() French:  %s\n", dashboard("fr"))

	// T() - Translation by key
	hello := i18n.T("hello-0", "John")
	fmt.Printf("T() English: %s\n", hello("en"))
	fmt.Printf("T() French:  %s\n", hello("fr"))

	fmt.Println("\n=== Pluralization Examples ===")

	// P() - Pluralization
	items := []int{0, 1, 2, 5}
	for _, count := range items {
		itemFn := i18n.P("item-count", count)
		fmt.Printf("P() English (%d): %s\n", count, itemFn("en"))
		fmt.Printf("P() French  (%d): %s\n", count, itemFn("fr"))
	}

	fmt.Println("\n=== Direct Translation ===")

	// R() - Direct translation (no function wrapping)
	fmt.Printf("R() English: %s\n", i18n.R("en", "Dashboard"))
	fmt.Printf("R() French:  %s\n", i18n.R("fr", "Dashboard"))

	fmt.Println("\n=== Fallback Behavior ===")

	// Test fallback for missing translations
	missing := i18n.S("Missing Translation")
	fmt.Printf("Missing EN: %s\n", missing("en"))
	fmt.Printf("Missing FR: %s (falls back to original)\n", missing("fr"))

	// Test unknown locale fallback
	fmt.Printf("Unknown locale (de): %s (falls back to default)\n", welcome("de"))

	// Clean up test files
	os.RemoveAll("locales")
}
