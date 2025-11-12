package output

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// YAMLFormatter formats output as YAML.
type YAMLFormatter struct{}

// FormatList formats a slice of items as a YAML array.
func (f *YAMLFormatter) FormatList(w io.Writer, headers []string, items interface{}) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	defer encoder.Close()
	return encoder.Encode(items)
}

// FormatSingle formats a single item as a YAML object.
func (f *YAMLFormatter) FormatSingle(w io.Writer, item interface{}) error {
	if item == nil {
		return fmt.Errorf("item is nil")
	}

	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	defer encoder.Close()
	return encoder.Encode(item)
}
