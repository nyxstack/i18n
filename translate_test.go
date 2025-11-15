package i18n

import (
	"fmt"
	"testing"
)

func setupTestDictionaries() {
	// Clean up existing dictionaries
	muDicts.Lock()
	dictionaries = make(map[string]*Dictionary)
	muDicts.Unlock()

	// Set up English dictionary
	enDict := NewDictionary("en")
	enDict.AddAll(map[string]string{
		"hello-0":       "Hello {0}",
		"welcome":       "Welcome",
		"dashboard":     "Dashboard",
		"goodbye":       "Goodbye",
		"hello-0-world": "Hello {0} World",
		"item-count":    "{count, plural, one {# item} other {# items}}",
		"message-count": "{count, plural, zero {no messages} one {# message} other {# messages}}",
	})
	Register(enDict)

	// Set up French dictionary
	frDict := NewDictionary("fr")
	frDict.AddAll(map[string]string{
		"hello-0":       "Bonjour {0}",
		"welcome":       "Bienvenue",
		"dashboard":     "Tableau de bord",
		"hello-0-world": "Bonjour {0} Monde",
		"item-count":    "{count, plural, one {# élément} other {# éléments}}",
	})
	Register(frDict)

	SetDefaultLanguage("en")
}

func TestT_BasicTranslation(t *testing.T) {
	setupTestDictionaries()
	defer func() {
		muDicts.Lock()
		dictionaries = make(map[string]*Dictionary)
		muDicts.Unlock()
	}()

	fn := T("hello-0", "John")

	result := fn("en")
	if result != "Hello John" {
		t.Errorf("Expected 'Hello John', got '%s'", result)
	}

	result = fn("fr")
	if result != "Bonjour John" {
		t.Errorf("Expected 'Bonjour John', got '%s'", result)
	}
}

func TestT_FallbackToDefault(t *testing.T) {
	setupTestDictionaries()
	defer func() {
		muDicts.Lock()
		dictionaries = make(map[string]*Dictionary)
		muDicts.Unlock()
	}()

	fn := T("goodbye")

	// French doesn't have "goodbye", should fallback to English
	result := fn("fr")
	if result != "Goodbye" {
		t.Errorf("Expected 'Goodbye' (fallback), got '%s'", result)
	}
}

func TestT_NoTranslation(t *testing.T) {
	setupTestDictionaries()
	defer func() {
		muDicts.Lock()
		dictionaries = make(map[string]*Dictionary)
		muDicts.Unlock()
	}()

	fn := T("nonexistent-key", "arg")

	result := fn("en")
	if result != "nonexistent-key" {
		t.Errorf("Expected 'nonexistent-key', got '%s'", result)
	}
}

func TestF_BasicFormat(t *testing.T) {
	setupTestDictionaries()
	defer func() {
		muDicts.Lock()
		dictionaries = make(map[string]*Dictionary)
		muDicts.Unlock()
	}()

	fn := F("Hello %s World", "Beautiful")

	result := fn("en")
	if result != "Hello Beautiful World" {
		t.Errorf("Expected 'Hello Beautiful World', got '%s'", result)
	}

	result = fn("fr")
	if result != "Bonjour Beautiful Monde" {
		t.Errorf("Expected 'Bonjour Beautiful Monde', got '%s'", result)
	}
}

func TestF_NoTranslation(t *testing.T) {
	setupTestDictionaries()
	defer func() {
		muDicts.Lock()
		dictionaries = make(map[string]*Dictionary)
		muDicts.Unlock()
	}()

	fn := F("Unknown %s format", "test")

	// Should use normalized template since no translation exists
	result := fn("en")
	if result != "Unknown test format" {
		t.Errorf("Expected 'Unknown test format', got '%s'", result)
	}
}

func TestS_StaticText(t *testing.T) {
	setupTestDictionaries()
	defer func() {
		muDicts.Lock()
		dictionaries = make(map[string]*Dictionary)
		muDicts.Unlock()
	}()

	fn := S("Dashboard")

	result := fn("en")
	if result != "Dashboard" {
		t.Errorf("Expected 'Dashboard', got '%s'", result)
	}

	result = fn("fr")
	if result != "Tableau de bord" {
		t.Errorf("Expected 'Tableau de bord', got '%s'", result)
	}
}

func TestS_FallbackToOriginal(t *testing.T) {
	setupTestDictionaries()
	defer func() {
		muDicts.Lock()
		dictionaries = make(map[string]*Dictionary)
		muDicts.Unlock()
	}()

	fn := S("Unknown Text")

	result := fn("en")
	if result != "Unknown Text" {
		t.Errorf("Expected 'Unknown Text', got '%s'", result)
	}

	result = fn("fr")
	if result != "Unknown Text" {
		t.Errorf("Expected 'Unknown Text', got '%s'", result)
	}
}

func TestP_Pluralization(t *testing.T) {
	setupTestDictionaries()
	defer func() {
		muDicts.Lock()
		dictionaries = make(map[string]*Dictionary)
		muDicts.Unlock()
	}()

	// Add more comprehensive plural templates
	enDict := GetDictionary("en")
	enDict.Add("advanced-count", "{count, plural, zero {no items} one {# item} other {# items}}")

	ruDict := NewDictionary("ru")
	ruDict.Add("advanced-count", "{count, plural, zero {нет элементов} one {# элемент} few {# элемента} many {# элементов}}")
	Register(ruDict)

	arDict := NewDictionary("ar")
	arDict.Add("advanced-count", "{count, plural, zero {لا عناصر} one {عنصر واحد} two {عنصران} few {# عناصر} many {# عنصر} other {# عنصر}}")
	Register(arDict)

	tests := []struct {
		locale   string
		count    int
		expected string
	}{
		// English tests
		{"en", 0, "no items"},
		{"en", 1, "1 item"},
		{"en", 2, "2 items"},
		{"en", 5, "5 items"},

		// Russian tests (complex Slavic rules)
		{"ru", 0, "нет элементов"},
		{"ru", 1, "1 элемент"},
		{"ru", 2, "2 элемента"},
		{"ru", 3, "3 элемента"},
		{"ru", 4, "4 элемента"},
		{"ru", 5, "5 элементов"},
		{"ru", 10, "10 элементов"},

		// Arabic tests (even more complex)
		{"ar", 0, "لا عناصر"},
		{"ar", 1, "عنصر واحد"},
		{"ar", 2, "عنصران"},
		{"ar", 3, "3 عناصر"},
		{"ar", 5, "5 عناصر"},
		{"ar", 10, "10 عناصر"},
		{"ar", 11, "11 عنصر"},
		{"ar", 50, "50 عنصر"},
		{"ar", 100, "100 عنصر"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%d", tt.locale, tt.count), func(t *testing.T) {
			fn := P("advanced-count", tt.count)
			result := fn(tt.locale)
			if result != tt.expected {
				t.Errorf("P('advanced-count', %d)(%q) = %q, expected %q",
					tt.count, tt.locale, result, tt.expected)
			}
		})
	}

	// Test original simple cases for backwards compatibility
	fn := P("item-count", 1)
	result := fn("en")
	if result != "1 item" {
		t.Errorf("Expected '1 item', got '%s'", result)
	}

	result = fn("fr")
	if result != "1 élément" {
		t.Errorf("Expected '1 élément', got '%s'", result)
	}

	fn = P("item-count", 5)
	result = fn("en")
	if result != "5 items" {
		t.Errorf("Expected '5 items', got '%s'", result)
	}

	result = fn("fr")
	if result != "5 éléments" {
		t.Errorf("Expected '5 éléments', got '%s'", result)
	}
}

func TestP_FallbackToSimpleSubstitution(t *testing.T) {
	setupTestDictionaries()
	defer func() {
		muDicts.Lock()
		dictionaries = make(map[string]*Dictionary)
		muDicts.Unlock()
	}()

	// Add a simple template without ICU plural syntax
	enDict := GetDictionary("en")
	enDict.Add("simple-count", "{count} things")

	fn := P("simple-count", 3)
	result := fn("en")
	if result != "3 things" {
		t.Errorf("Expected '3 things', got '%s'", result)
	}
}

func TestR_DirectTranslation(t *testing.T) {
	setupTestDictionaries()
	defer func() {
		muDicts.Lock()
		dictionaries = make(map[string]*Dictionary)
		muDicts.Unlock()
	}()

	result := R("en", "Dashboard")
	if result != "Dashboard" {
		t.Errorf("Expected 'Dashboard', got '%s'", result)
	}

	result = R("fr", "Dashboard")
	if result != "Tableau de bord" {
		t.Errorf("Expected 'Tableau de bord', got '%s'", result)
	}

	result = R("fr", "Unknown Text")
	if result != "Unknown Text" {
		t.Errorf("Expected 'Unknown Text', got '%s'", result)
	}
}

func TestMultipleArgs(t *testing.T) {
	setupTestDictionaries()
	defer func() {
		muDicts.Lock()
		dictionaries = make(map[string]*Dictionary)
		muDicts.Unlock()
	}()

	// Add templates with multiple placeholders
	enDict := GetDictionary("en")
	enDict.Add("multi-args", "Hello {0}, you have {1} messages and {2} notifications")

	frDict := GetDictionary("fr")
	frDict.Add("multi-args", "Bonjour {0}, vous avez {1} messages et {2} notifications")

	fn := T("multi-args", "John", 5, 3)

	result := fn("en")
	expected := "Hello John, you have 5 messages and 3 notifications"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	result = fn("fr")
	expected = "Bonjour John, vous avez 5 messages et 3 notifications"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestEmptyArgs(t *testing.T) {
	setupTestDictionaries()
	defer func() {
		muDicts.Lock()
		dictionaries = make(map[string]*Dictionary)
		muDicts.Unlock()
	}()

	fn := T("welcome")

	result := fn("en")
	if result != "Welcome" {
		t.Errorf("Expected 'Welcome', got '%s'", result)
	}

	result = fn("fr")
	if result != "Bienvenue" {
		t.Errorf("Expected 'Bienvenue', got '%s'", result)
	}
}

func TestNonExistentLocale(t *testing.T) {
	setupTestDictionaries()
	defer func() {
		muDicts.Lock()
		dictionaries = make(map[string]*Dictionary)
		muDicts.Unlock()
	}()

	fn := T("welcome")

	// Should fallback to default language (en) for unknown locale
	result := fn("de")
	if result != "Welcome" {
		t.Errorf("Expected 'Welcome' (fallback), got '%s'", result)
	}
}
