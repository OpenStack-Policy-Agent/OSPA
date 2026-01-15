package generators

import (
	"os"
	"path/filepath"
	"strings"
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
	if err := tmpl.Execute(file, data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return nil
}

// ToPascal converts a snake_case or kebab-case identifier into PascalCase.
// Examples:
//   - "security_group_rule" -> "SecurityGroupRule"
//   - "floating_ip" -> "FloatingIP"
func ToPascal(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		p := parts[i]
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

