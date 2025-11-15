package i18n

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

// GenerateTranslations scans a Go codebase for i18n function calls (F, S, T, P)
// and generates translation keys + source strings into a dictionary file in the locales/ folder.
func GenerateTranslations(locale, root, outputPath string) error {
	results := make(map[string]string)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		fs := token.NewFileSet()
		node, err := parser.ParseFile(fs, path, nil, parser.AllErrors)
		if err != nil {
			return nil
		}

		ast.Inspect(node, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "i18n" {
				return true
			}

			funcName := sel.Sel.Name
			if funcName != "F" && funcName != "S" && funcName != "T" && funcName != "P" {
				return true
			}

			if len(call.Args) == 0 {
				return true
			}

			firstArg, ok := call.Args[0].(*ast.BasicLit)
			if !ok || firstArg.Kind != token.STRING {
				return true
			}

			// Clean up the string literal quotes
			raw := firstArg.Value
			if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
				raw = raw[1 : len(raw)-1]
			}

			key := slugify(raw)
			results[key] = raw

			pos := fs.Position(firstArg.Pos())
			fmt.Printf("[%s] %s.%s → %s → key: %s\n",
				pos, pkg.Name, funcName, raw, key)

			return true
		})
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking files: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("no i18n calls found")
		return nil
	}

	// Use default output path if empty
	if outputPath == "" {
		outputPath = filepath.Join(DefaultFolder, fmt.Sprintf("%s.%s.json", DefaultDictionary, locale))
	}

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create TranslationFile structure for saving
	tf := TranslationFile{
		Meta: struct {
			Lang      string `json:"lang"`
			Name      string `json:"name"`
			Version   string `json:"version,omitempty"`
			Author    string `json:"author,omitempty"`
			Updated   string `json:"updated,omitempty"`
			Direction string `json:"direction,omitempty"`
		}{
			Lang: locale,
			Name: DefaultDictionary,
		},
		Translations: results,
	}

	// Save to JSON file
	data, err := json.MarshalIndent(tf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal dictionary: %w", err)
	}

	if err := os.WriteFile(filepath.Clean(outputPath), data, 0644); err != nil {
		return fmt.Errorf("failed to save dictionary: %w", err)
	}

	fmt.Printf("✅ Extracted %d i18n entries → %s\n", len(results), outputPath)
	return nil
}

// Generate is a convenience function that generates translations to the default location
func Generate(locale, root string) error {
	return GenerateTranslations(locale, root, "")
}
