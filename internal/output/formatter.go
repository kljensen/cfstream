// Package output provides formatting capabilities for CLI output.
package output

import (
	"fmt"
	"io"
)

// Formatter defines the interface for formatting output data.
type Formatter interface {
	// FormatList formats a slice of items with optional headers.
	// The items parameter should be a slice of structs or maps.
	// The headers parameter specifies which fields to display and their order.
	FormatList(w io.Writer, headers []string, items interface{}) error

	// FormatSingle formats a single item.
	// The item parameter should be a struct or map.
	FormatSingle(w io.Writer, item interface{}) error
}

// NewFormatter creates a new formatter based on the specified format type.
// Supported formats: "table", "json", "yaml".
func NewFormatter(format string) (Formatter, error) {
	switch format {
	case "table":
		return &TableFormatter{}, nil
	case "json":
		return &JSONFormatter{}, nil
	case "yaml":
		return &YAMLFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s (supported: table, json, yaml)", format)
	}
}
