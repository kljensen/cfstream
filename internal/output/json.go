package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONFormatter formats output as pretty-printed JSON.
type JSONFormatter struct{}

// FormatList formats a slice of items as a JSON array.
func (f *JSONFormatter) FormatList(w io.Writer, headers []string, items interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(items)
}

// FormatSingle formats a single item as a JSON object.
func (f *JSONFormatter) FormatSingle(w io.Writer, item interface{}) error {
	if item == nil {
		return fmt.Errorf("item is nil")
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(item)
}
