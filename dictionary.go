// Package i18n provides a comprehensive internationalization system with:
// - Consistent 0-based indexing for placeholders
// - File-based translation dictionaries
// - Advanced pluralization support for multiple languages
// - Thread-safe operations
// - Automatic fallback to default languages
// - Code generation for extracting translation keys
//
// Basic usage:
//
//	// Load translations
//	i18n.Load()
//	i18n.LoadLanguage("fr")
//
//	// Use translations
//	greeting := i18n.F("Hello %s", "World")
//	fmt.Println(greeting("en")) // "Hello World"
//	fmt.Println(greeting("fr")) // "Bonjour World"
//
//	title := i18n.S("Dashboard")
//	fmt.Println(title("fr")) // "Tableau de bord"
//
//	count := i18n.P("item-count", 5)
//	fmt.Println(count("en")) // "5 items"
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// -----------------------------------------------------------------------------
// Constants and defaults
// -----------------------------------------------------------------------------

const (
	DefaultLang       = "en"
	DefaultDictionary = "default"
	DefaultFolder     = "locales"
	DefaultFilePath   = "locales/default.en.json"
)

// -----------------------------------------------------------------------------
// Data structures
// -----------------------------------------------------------------------------

// TranslationFile represents a single dictionary file
type TranslationFile struct {
	Meta struct {
		Lang      string `json:"lang"`
		Name      string `json:"name"`
		Version   string `json:"version,omitempty"`
		Author    string `json:"author,omitempty"`
		Updated   string `json:"updated,omitempty"`
		Direction string `json:"direction,omitempty"`
	} `json:"meta"`
	Translations map[string]string `json:"translations"`
}

// Dictionary represents one language's translations
type Dictionary struct {
	Lang         string
	Translations map[string]string
	mu           sync.RWMutex
}

// -----------------------------------------------------------------------------
// Registry management
// -----------------------------------------------------------------------------

var (
	dictionaries  = map[string]*Dictionary{}
	muDicts       sync.RWMutex
	currentLang   = DefaultLang
	muDefaultLang sync.RWMutex
)

// SetDefaultLanguage sets the fallback language code
func SetDefaultLanguage(lang string) {
	muDefaultLang.Lock()
	defer muDefaultLang.Unlock()
	currentLang = lang
}

// DefaultLanguage returns the current fallback language
func DefaultLanguage() string {
	muDefaultLang.RLock()
	defer muDefaultLang.RUnlock()
	return currentLang
}

// Register adds a dictionary to the global registry
func Register(dict *Dictionary) {
	muDicts.Lock()
	defer muDicts.Unlock()
	dictionaries[dict.Lang] = dict
}

// GetDictionary returns a dictionary by language code
func GetDictionary(lang string) *Dictionary {
	muDicts.RLock()
	defer muDicts.RUnlock()
	return dictionaries[lang]
}

// -----------------------------------------------------------------------------
// Dictionary creation and loading
// -----------------------------------------------------------------------------

// NewDictionary creates an empty dictionary for a language
func NewDictionary(lang string) *Dictionary {
	return &Dictionary{
		Lang:         lang,
		Translations: make(map[string]string),
	}
}

// LoadDictionaryFile loads a single dictionary file
func LoadDictionaryFile(path string) (*Dictionary, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var tf TranslationFile
	if err := json.Unmarshal(data, &tf); err != nil {
		return nil, fmt.Errorf("invalid translation file %s: %w", path, err)
	}

	// Validate translation file structure
	if err := validateTranslationFile(&tf); err != nil {
		return nil, fmt.Errorf("validation failed for %s: %w", path, err)
	}

	dict := NewDictionary(tf.Meta.Lang)
	dict.AddAll(tf.Translations)
	return dict, nil
}

// validateTranslationFile validates the structure and content of a translation file
func validateTranslationFile(tf *TranslationFile) error {
	// Check required meta fields
	if tf.Meta.Lang == "" {
		return fmt.Errorf("missing required 'meta.lang' field")
	}

	if tf.Meta.Name == "" {
		return fmt.Errorf("missing required 'meta.name' field")
	}

	// Validate language code format (basic validation)
	if len(tf.Meta.Lang) < 2 || len(tf.Meta.Lang) > 5 {
		return fmt.Errorf("invalid language code '%s': must be 2-5 characters", tf.Meta.Lang)
	}

	// Check for valid characters in language code (letters, numbers, hyphens)
	for _, r := range tf.Meta.Lang {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-') {
			return fmt.Errorf("invalid language code '%s': contains invalid character '%c'", tf.Meta.Lang, r)
		}
	}

	// Validate translations
	if tf.Translations == nil {
		return fmt.Errorf("missing 'translations' field")
	}

	// Check for empty keys or values
	for key, value := range tf.Translations {
		if key == "" {
			return fmt.Errorf("translation has empty key")
		}
		if value == "" {
			return fmt.Errorf("translation key '%s' has empty value", key)
		}

		// Validate placeholder consistency in ICU plural forms
		if err := validatePluralTemplate(key, value); err != nil {
			return fmt.Errorf("invalid plural template for key '%s': %w", key, err)
		}
	}

	return nil
}

// validatePluralTemplate validates ICU-style plural templates
func validatePluralTemplate(key, template string) error {
	if !strings.Contains(template, "{count, plural") {
		return nil // Not a plural template, skip validation
	}

	// Check for balanced braces
	braceCount := 0
	for _, r := range template {
		if r == '{' {
			braceCount++
		} else if r == '}' {
			braceCount--
			if braceCount < 0 {
				return fmt.Errorf("unbalanced braces: too many closing braces")
			}
		}
	}

	if braceCount != 0 {
		return fmt.Errorf("unbalanced braces: missing %d closing brace(s)", braceCount)
	}

	// Validate that it contains at least one valid plural form
	validForms := []string{"zero", "one", "two", "few", "many", "other"}
	foundValidForm := false

	for _, form := range validForms {
		if strings.Contains(template, form+" {") {
			foundValidForm = true
			break
		}
	}

	if !foundValidForm {
		return fmt.Errorf("no valid plural forms found (valid forms: %s)", strings.Join(validForms, ", "))
	}

	return nil
}

// Load auto-loads the default dictionary from locales/default.en.json
func Load() error {
	return LoadFrom(DefaultFilePath)
}

// LoadFrom loads and registers a dictionary from a specific path
func LoadFrom(path string) error {
	dict, err := LoadDictionaryFile(path)
	if err != nil {
		return err
	}
	Register(dict)
	return nil
}

// LoadLanguage loads a dictionary for a specific language from locales/default.{lang}.json
func LoadLanguage(lang string) error {
	path := filepath.Join(DefaultFolder, fmt.Sprintf("%s.%s.json", DefaultDictionary, lang))
	return LoadFrom(path)
}

// -----------------------------------------------------------------------------
// Dictionary operations
// -----------------------------------------------------------------------------

// Add inserts or updates a translation
func (d *Dictionary) Add(key, value string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.Translations == nil {
		d.Translations = make(map[string]string)
	}
	d.Translations[key] = value
}

// AddAll merges translations from a map
func (d *Dictionary) AddAll(translations map[string]string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.Translations == nil {
		d.Translations = make(map[string]string)
	}
	for k, v := range translations {
		d.Translations[k] = v
	}
}

// Get retrieves a translation with fallback to default language
func (d *Dictionary) Get(key string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Try to get from this dictionary first
	if value, ok := d.Translations[key]; ok {
		return value
	}

	// Fallback to default language dictionary if this isn't the default
	if d.Lang != DefaultLanguage() {
		if defaultDict := GetDictionary(DefaultLanguage()); defaultDict != nil && defaultDict != d {
			return defaultDict.Get(key)
		}
	}

	// Return key if not found
	return key
}

// Has checks if a translation key exists
func (d *Dictionary) Has(key string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, ok := d.Translations[key]
	return ok
}

// Keys returns all translation keys
func (d *Dictionary) Keys() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	keys := make([]string, 0, len(d.Translations))
	for k := range d.Translations {
		keys = append(keys, k)
	}
	return keys
}

// Count returns the number of translations
func (d *Dictionary) Count() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.Translations)
}
