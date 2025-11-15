// CLI tool for extracting i18n translation keys from Go source code
package main

import (
	"fmt"
	"os"

	"github.com/nyxstack/i18n"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: extract-i18n <source_dir> <locale> [output_path]")
		fmt.Println("  source_dir: Directory to scan for Go files")
		fmt.Println("  locale:     Language code (e.g., 'en', 'fr', 'es')")
		fmt.Println("  output_path: Optional custom output path")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  extract-i18n . en")
		fmt.Println("  extract-i18n ./src fr")
		fmt.Println("  extract-i18n . en ./translations/en.json")
		os.Exit(1)
	}

	sourceDir := os.Args[1]
	locale := os.Args[2]

	var outputPath string
	if len(os.Args) > 3 {
		outputPath = os.Args[3]
	}

	err := i18n.GenerateTranslations(locale, sourceDir, outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
