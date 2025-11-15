package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateTranslations(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a test Go file with i18n function calls
	testGoFile := filepath.Join(tempDir, "test.go")
	testGoContent := `package main

import "github.com/nyxstack/i18n"

func main() {
	greeting := i18n.F("Hello %s", "World")
	welcome := i18n.S("Welcome")
	goodbye := i18n.T("goodbye_message", "John")
	count := i18n.P("item_count", 5)
}
`

	if err := os.WriteFile(testGoFile, []byte(testGoContent), 0644); err != nil {
		t.Fatalf("Failed to create test Go file: %v", err)
	}

	// Set up output path
	outputPath := filepath.Join(tempDir, "locales", "default.en.json")

	// Generate translations
	err := GenerateTranslations("en", tempDir, outputPath)
	if err != nil {
		t.Fatalf("GenerateTranslations failed: %v", err)
	}

	// Verify the output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file was not created: %s", outputPath)
	}

	// Read and parse the generated file
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	var tf TranslationFile
	if err := json.Unmarshal(data, &tf); err != nil {
		t.Fatalf("Failed to parse generated JSON: %v", err)
	}

	// Verify metadata
	if tf.Meta.Lang != "en" {
		t.Errorf("Expected lang 'en', got '%s'", tf.Meta.Lang)
	}

	if tf.Meta.Name != "default" {
		t.Errorf("Expected name 'default', got '%s'", tf.Meta.Name)
	}

	// Verify translations were extracted
	expectedTranslations := map[string]string{
		"hello-0":         "Hello %s",
		"welcome":         "Welcome",
		"goodbye-message": "goodbye_message", // T() uses the key as-is
		"item-count":      "item_count",      // P() uses the key as-is
	}

	for expectedKey, expectedValue := range expectedTranslations {
		if actualValue, exists := tf.Translations[expectedKey]; !exists {
			t.Errorf("Expected key '%s' not found in translations", expectedKey)
		} else if actualValue != expectedValue {
			t.Errorf("For key '%s', expected '%s', got '%s'",
				expectedKey, expectedValue, actualValue)
		}
	}

	// Verify we have the expected number of translations
	if len(tf.Translations) != len(expectedTranslations) {
		t.Errorf("Expected %d translations, got %d",
			len(expectedTranslations), len(tf.Translations))
	}
}

func TestGenerate(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Change to temp directory to test default output path
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create a test Go file
	testGoFile := "test.go"
	testGoContent := `package main

import "github.com/nyxstack/i18n"

func main() {
	title := i18n.S("Dashboard")
}
`

	if err := os.WriteFile(testGoFile, []byte(testGoContent), 0644); err != nil {
		t.Fatalf("Failed to create test Go file: %v", err)
	}

	// Generate translations using the convenience function
	err = Generate("fr", ".")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify the default output file was created
	expectedPath := filepath.Join("locales", "default.fr.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Default output file was not created: %s", expectedPath)
	}

	// Read and verify content
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	var tf TranslationFile
	if err := json.Unmarshal(data, &tf); err != nil {
		t.Fatalf("Failed to parse generated JSON: %v", err)
	}

	if tf.Meta.Lang != "fr" {
		t.Errorf("Expected lang 'fr', got '%s'", tf.Meta.Lang)
	}

	if _, exists := tf.Translations["dashboard"]; !exists {
		t.Error("Expected 'dashboard' key not found in translations")
	}
}

func TestGenerateTranslations_NoI18nCalls(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a Go file without any i18n calls
	testGoFile := filepath.Join(tempDir, "test.go")
	testGoContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello World")
}
`

	if err := os.WriteFile(testGoFile, []byte(testGoContent), 0644); err != nil {
		t.Fatalf("Failed to create test Go file: %v", err)
	}

	outputPath := filepath.Join(tempDir, "locales", "default.en.json")

	// Generate translations (should succeed but create no file)
	err := GenerateTranslations("en", tempDir, outputPath)
	if err != nil {
		t.Fatalf("GenerateTranslations failed: %v", err)
	}

	// Verify no output file was created since there were no i18n calls
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Error("Output file should not have been created when no i18n calls found")
	}
}

func TestGenerateTranslations_InvalidGoFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create an invalid Go file
	testGoFile := filepath.Join(tempDir, "test.go")
	testGoContent := `this is not valid go code`

	if err := os.WriteFile(testGoFile, []byte(testGoContent), 0644); err != nil {
		t.Fatalf("Failed to create test Go file: %v", err)
	}

	outputPath := filepath.Join(tempDir, "locales", "default.en.json")

	// Generate translations (should succeed but skip invalid files)
	err := GenerateTranslations("en", tempDir, outputPath)
	if err != nil {
		t.Fatalf("GenerateTranslations failed: %v", err)
	}

	// Verify no output file was created since the Go file was invalid
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Error("Output file should not have been created when Go file is invalid")
	}
}

func TestGenerateTranslations_ComplexExpressions(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a test Go file with complex i18n expressions
	testGoFile := filepath.Join(tempDir, "test.go")
	testGoContent := `package main

import "github.com/nyxstack/i18n"

func main() {
	// These should be extracted
	simple := i18n.S("Simple")
	format := i18n.F("Hello %s", "world")
	
	// These should be ignored (not string literals)
	variable := "dynamic"
	dynamic := i18n.S(variable)
	
	// Non-i18n calls should be ignored
	fmt.Printf("Not i18n")
	other.F("Not our package")
}
`

	if err := os.WriteFile(testGoFile, []byte(testGoContent), 0644); err != nil {
		t.Fatalf("Failed to create test Go file: %v", err)
	}

	outputPath := filepath.Join(tempDir, "locales", "default.en.json")

	// Generate translations
	err := GenerateTranslations("en", tempDir, outputPath)
	if err != nil {
		t.Fatalf("GenerateTranslations failed: %v", err)
	}

	// Read and parse the generated file
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	var tf TranslationFile
	if err := json.Unmarshal(data, &tf); err != nil {
		t.Fatalf("Failed to parse generated JSON: %v", err)
	}

	// Should only have the string literal calls
	expectedKeys := []string{"simple", "hello-0"}
	if len(tf.Translations) != len(expectedKeys) {
		t.Errorf("Expected %d translations, got %d", len(expectedKeys), len(tf.Translations))
	}

	for _, key := range expectedKeys {
		if _, exists := tf.Translations[key]; !exists {
			t.Errorf("Expected key '%s' not found in translations", key)
		}
	}
}

func TestGenerateTranslations_DirectoryCreation(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a test Go file
	testGoFile := filepath.Join(tempDir, "test.go")
	testGoContent := `package main

import "github.com/nyxstack/i18n"

func main() {
	msg := i18n.S("Test")
}
`

	if err := os.WriteFile(testGoFile, []byte(testGoContent), 0644); err != nil {
		t.Fatalf("Failed to create test Go file: %v", err)
	}

	// Use a nested output path that doesn't exist
	outputPath := filepath.Join(tempDir, "nested", "deep", "locales", "test.en.json")

	// Generate translations
	err := GenerateTranslations("en", tempDir, outputPath)
	if err != nil {
		t.Fatalf("GenerateTranslations failed: %v", err)
	}

	// Verify the output file was created and directories were created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file was not created: %s", outputPath)
	}

	// Verify directory structure was created
	if _, err := os.Stat(filepath.Dir(outputPath)); os.IsNotExist(err) {
		t.Fatalf("Output directory was not created: %s", filepath.Dir(outputPath))
	}
}
