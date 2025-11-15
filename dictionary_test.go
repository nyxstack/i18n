package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestNewDictionary(t *testing.T) {
	dict := NewDictionary("en")
	if dict.Lang != "en" {
		t.Errorf("Expected lang 'en', got '%s'", dict.Lang)
	}
	if dict.Translations == nil {
		t.Error("Expected translations map to be initialized")
	}
	if len(dict.Translations) != 0 {
		t.Errorf("Expected empty translations, got %d entries", len(dict.Translations))
	}
}

func TestDictionaryAdd(t *testing.T) {
	dict := NewDictionary("en")
	dict.Add("test_key", "test_value")

	if dict.Count() != 1 {
		t.Errorf("Expected 1 translation, got %d", dict.Count())
	}

	value := dict.Get("test_key")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}
}

func TestDictionaryAddAll(t *testing.T) {
	dict := NewDictionary("en")
	translations := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	dict.AddAll(translations)

	if dict.Count() != 3 {
		t.Errorf("Expected 3 translations, got %d", dict.Count())
	}

	for key, expectedValue := range translations {
		if value := dict.Get(key); value != expectedValue {
			t.Errorf("Expected '%s' for key '%s', got '%s'", expectedValue, key, value)
		}
	}
}

func TestDictionaryGet_Fallback(t *testing.T) {
	// Set up default language
	SetDefaultLanguage("en")

	// Create and register default dictionary
	defaultDict := NewDictionary("en")
	defaultDict.Add("test_key", "default_value")
	Register(defaultDict)

	// Create secondary dictionary without the key
	frDict := NewDictionary("fr")
	Register(frDict)

	// Should fallback to default language
	value := frDict.Get("test_key")
	if value != "default_value" {
		t.Errorf("Expected fallback 'default_value', got '%s'", value)
	}

	// Cleanup
	muDicts.Lock()
	dictionaries = make(map[string]*Dictionary)
	muDicts.Unlock()
}

func TestDictionaryGet_ReturnKeyIfNotFound(t *testing.T) {
	dict := NewDictionary("en")
	SetDefaultLanguage("en")
	Register(dict)

	value := dict.Get("nonexistent_key")
	if value != "nonexistent_key" {
		t.Errorf("Expected key as fallback 'nonexistent_key', got '%s'", value)
	}

	// Cleanup
	muDicts.Lock()
	dictionaries = make(map[string]*Dictionary)
	muDicts.Unlock()
}

func TestDictionaryHas(t *testing.T) {
	dict := NewDictionary("en")
	dict.Add("existing_key", "value")

	if !dict.Has("existing_key") {
		t.Error("Expected Has() to return true for existing key")
	}

	if dict.Has("nonexistent_key") {
		t.Error("Expected Has() to return false for nonexistent key")
	}
}

func TestDictionaryKeys(t *testing.T) {
	dict := NewDictionary("en")
	expectedKeys := []string{"key1", "key2", "key3"}

	for _, key := range expectedKeys {
		dict.Add(key, "value")
	}

	keys := dict.Keys()
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	// Check that all expected keys are present
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	for _, expectedKey := range expectedKeys {
		if !keyMap[expectedKey] {
			t.Errorf("Expected key '%s' not found in keys list", expectedKey)
		}
	}
}

func TestDictionaryCount(t *testing.T) {
	dict := NewDictionary("en")

	if dict.Count() != 0 {
		t.Errorf("Expected count 0 for empty dictionary, got %d", dict.Count())
	}

	dict.Add("key1", "value1")
	if dict.Count() != 1 {
		t.Errorf("Expected count 1, got %d", dict.Count())
	}

	dict.Add("key2", "value2")
	if dict.Count() != 2 {
		t.Errorf("Expected count 2, got %d", dict.Count())
	}
}

func TestSetDefaultLanguage(t *testing.T) {
	original := DefaultLanguage()
	defer SetDefaultLanguage(original) // Restore original

	SetDefaultLanguage("fr")
	if DefaultLanguage() != "fr" {
		t.Errorf("Expected default language 'fr', got '%s'", DefaultLanguage())
	}
}

func TestRegisterAndGetDictionary(t *testing.T) {
	dict := NewDictionary("test")
	dict.Add("test_key", "test_value")

	Register(dict)

	retrieved := GetDictionary("test")
	if retrieved == nil {
		t.Error("Expected dictionary to be retrieved, got nil")
	}

	if retrieved.Lang != "test" {
		t.Errorf("Expected lang 'test', got '%s'", retrieved.Lang)
	}

	if retrieved.Get("test_key") != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", retrieved.Get("test_key"))
	}

	// Cleanup
	muDicts.Lock()
	delete(dictionaries, "test")
	muDicts.Unlock()
}

func TestLoadDictionaryFile(t *testing.T) {
	// Create temporary test file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")

	testData := TranslationFile{
		Meta: struct {
			Lang      string `json:"lang"`
			Name      string `json:"name"`
			Version   string `json:"version,omitempty"`
			Author    string `json:"author,omitempty"`
			Updated   string `json:"updated,omitempty"`
			Direction string `json:"direction,omitempty"`
		}{
			Lang: "en",
			Name: "test",
		},
		Translations: map[string]string{
			"hello": "Hello",
			"world": "World",
		},
	}

	data, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	dict, err := LoadDictionaryFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load dictionary: %v", err)
	}

	if dict.Lang != "en" {
		t.Errorf("Expected lang 'en', got '%s'", dict.Lang)
	}

	if dict.Get("hello") != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", dict.Get("hello"))
	}

	if dict.Get("world") != "World" {
		t.Errorf("Expected 'World', got '%s'", dict.Get("world"))
	}
}

func TestLoadDictionaryFile_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "invalid.json")

	if err := os.WriteFile(filePath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadDictionaryFile(filePath)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestLoadDictionaryFile_MissingLang(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "nolang.json")

	testData := TranslationFile{
		Translations: map[string]string{"test": "value"},
	}

	data, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = LoadDictionaryFile(filePath)
	if err == nil {
		t.Error("Expected error for missing lang field, got nil")
	}
}

func TestValidateTranslationFile(t *testing.T) {
	tests := []struct {
		name    string
		tf      TranslationFile
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid file",
			tf: TranslationFile{
				Meta: struct {
					Lang      string `json:"lang"`
					Name      string `json:"name"`
					Version   string `json:"version,omitempty"`
					Author    string `json:"author,omitempty"`
					Updated   string `json:"updated,omitempty"`
					Direction string `json:"direction,omitempty"`
				}{
					Lang: "en",
					Name: "test",
				},
				Translations: map[string]string{
					"hello": "Hello",
					"world": "World",
				},
			},
			wantErr: false,
		},
		{
			name: "missing lang",
			tf: TranslationFile{
				Meta: struct {
					Lang      string `json:"lang"`
					Name      string `json:"name"`
					Version   string `json:"version,omitempty"`
					Author    string `json:"author,omitempty"`
					Updated   string `json:"updated,omitempty"`
					Direction string `json:"direction,omitempty"`
				}{
					Name: "test",
				},
				Translations: map[string]string{"hello": "Hello"},
			},
			wantErr: true,
			errMsg:  "missing required 'meta.lang' field",
		},
		{
			name: "missing name",
			tf: TranslationFile{
				Meta: struct {
					Lang      string `json:"lang"`
					Name      string `json:"name"`
					Version   string `json:"version,omitempty"`
					Author    string `json:"author,omitempty"`
					Updated   string `json:"updated,omitempty"`
					Direction string `json:"direction,omitempty"`
				}{
					Lang: "en",
				},
				Translations: map[string]string{"hello": "Hello"},
			},
			wantErr: true,
			errMsg:  "missing required 'meta.name' field",
		},
		{
			name: "invalid lang code - too short",
			tf: TranslationFile{
				Meta: struct {
					Lang      string `json:"lang"`
					Name      string `json:"name"`
					Version   string `json:"version,omitempty"`
					Author    string `json:"author,omitempty"`
					Updated   string `json:"updated,omitempty"`
					Direction string `json:"direction,omitempty"`
				}{
					Lang: "e",
					Name: "test",
				},
				Translations: map[string]string{"hello": "Hello"},
			},
			wantErr: true,
			errMsg:  "invalid language code 'e': must be 2-5 characters",
		},
		{
			name: "invalid lang code - too long",
			tf: TranslationFile{
				Meta: struct {
					Lang      string `json:"lang"`
					Name      string `json:"name"`
					Version   string `json:"version,omitempty"`
					Author    string `json:"author,omitempty"`
					Updated   string `json:"updated,omitempty"`
					Direction string `json:"direction,omitempty"`
				}{
					Lang: "toolong",
					Name: "test",
				},
				Translations: map[string]string{"hello": "Hello"},
			},
			wantErr: true,
			errMsg:  "invalid language code 'toolong': must be 2-5 characters",
		},
		{
			name: "invalid lang code - invalid characters",
			tf: TranslationFile{
				Meta: struct {
					Lang      string `json:"lang"`
					Name      string `json:"name"`
					Version   string `json:"version,omitempty"`
					Author    string `json:"author,omitempty"`
					Updated   string `json:"updated,omitempty"`
					Direction string `json:"direction,omitempty"`
				}{
					Lang: "en@US",
					Name: "test",
				},
				Translations: map[string]string{"hello": "Hello"},
			},
			wantErr: true,
			errMsg:  "invalid language code 'en@US': contains invalid character '@'",
		},
		{
			name: "empty key",
			tf: TranslationFile{
				Meta: struct {
					Lang      string `json:"lang"`
					Name      string `json:"name"`
					Version   string `json:"version,omitempty"`
					Author    string `json:"author,omitempty"`
					Updated   string `json:"updated,omitempty"`
					Direction string `json:"direction,omitempty"`
				}{
					Lang: "en",
					Name: "test",
				},
				Translations: map[string]string{
					"":      "Hello",
					"world": "World",
				},
			},
			wantErr: true,
			errMsg:  "translation has empty key",
		},
		{
			name: "empty value",
			tf: TranslationFile{
				Meta: struct {
					Lang      string `json:"lang"`
					Name      string `json:"name"`
					Version   string `json:"version,omitempty"`
					Author    string `json:"author,omitempty"`
					Updated   string `json:"updated,omitempty"`
					Direction string `json:"direction,omitempty"`
				}{
					Lang: "en",
					Name: "test",
				},
				Translations: map[string]string{
					"hello": "",
					"world": "World",
				},
			},
			wantErr: true,
			errMsg:  "translation key 'hello' has empty value",
		},
		{
			name: "valid plural template",
			tf: TranslationFile{
				Meta: struct {
					Lang      string `json:"lang"`
					Name      string `json:"name"`
					Version   string `json:"version,omitempty"`
					Author    string `json:"author,omitempty"`
					Updated   string `json:"updated,omitempty"`
					Direction string `json:"direction,omitempty"`
				}{
					Lang: "en",
					Name: "test",
				},
				Translations: map[string]string{
					"items": "{count, plural, one {# item} other {# items}}",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid plural template - unbalanced braces",
			tf: TranslationFile{
				Meta: struct {
					Lang      string `json:"lang"`
					Name      string `json:"name"`
					Version   string `json:"version,omitempty"`
					Author    string `json:"author,omitempty"`
					Updated   string `json:"updated,omitempty"`
					Direction string `json:"direction,omitempty"`
				}{
					Lang: "en",
					Name: "test",
				},
				Translations: map[string]string{
					"items": "{count, plural, one {# item other {# items}}",
				},
			},
			wantErr: true,
			errMsg:  "unbalanced braces",
		},
		{
			name: "invalid plural template - no valid forms",
			tf: TranslationFile{
				Meta: struct {
					Lang      string `json:"lang"`
					Name      string `json:"name"`
					Version   string `json:"version,omitempty"`
					Author    string `json:"author,omitempty"`
					Updated   string `json:"updated,omitempty"`
					Direction string `json:"direction,omitempty"`
				}{
					Lang: "en",
					Name: "test",
				},
				Translations: map[string]string{
					"items": "{count, plural, invalid {# item}}",
				},
			},
			wantErr: true,
			errMsg:  "no valid plural forms found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTranslationFile(&tt.tf)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateTranslationFile() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateTranslationFile() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateTranslationFile() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidatePluralTemplate(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		template string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "not a plural template",
			key:      "hello",
			template: "Hello {0}",
			wantErr:  false,
		},
		{
			name:     "valid plural template",
			key:      "items",
			template: "{count, plural, one {# item} other {# items}}",
			wantErr:  false,
		},
		{
			name:     "unbalanced braces - missing closing",
			key:      "items",
			template: "{count, plural, one {# item other {# items}}",
			wantErr:  true,
			errMsg:   "unbalanced braces",
		},
		{
			name:     "unbalanced braces - extra closing",
			key:      "items",
			template: "{count, plural, one {# item} other {# items}}}",
			wantErr:  true,
			errMsg:   "unbalanced braces",
		},
		{
			name:     "no valid plural forms",
			key:      "items",
			template: "{count, plural, invalid {# item}}",
			wantErr:  true,
			errMsg:   "no valid plural forms found",
		},
		{
			name:     "complex valid template",
			key:      "messages",
			template: "{count, plural, zero {no messages} one {# message} few {# messages} other {# messages}}",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePluralTemplate(tt.key, tt.template)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validatePluralTemplate() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validatePluralTemplate() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validatePluralTemplate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestDictionaryConcurrency(t *testing.T) {
	dict := NewDictionary("en")

	// Test concurrent access
	var wg sync.WaitGroup
	numGoroutines := 10
	numOpsPerGoroutine := 100

	// Concurrent writes
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOpsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)
				dict.Add(key, value)
			}
		}(i)
	}

	// Concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOpsPerGoroutine; j++ {
				dict.Get("nonexistent_key")
				dict.Has("nonexistent_key")
				dict.Count()
				dict.Keys()
			}
		}(i)
	}

	wg.Wait()

	expectedCount := numGoroutines * numOpsPerGoroutine
	if dict.Count() != expectedCount {
		t.Errorf("Expected %d translations after concurrent operations, got %d",
			expectedCount, dict.Count())
	}
}
