package generators

import (
	"os"
	"path/filepath"
	"text/template"
)

// fileExists checks if a file exists
func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// writeFile writes a file using a template
func writeFile(filePath string, tmpl *template.Template, data interface{}) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

