package report

import (
	"fmt"
	"io"
	"strings"
)

// NewWriter creates a result writer for the requested format.
func NewWriter(format string, w io.Writer) (ResultWriter, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "json":
		return NewJSONWriter(w), nil
	case "csv":
		return NewCSVWriter(w), nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}
