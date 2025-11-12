package output

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/olekukonko/tablewriter"
)

// TableFormatter formats output as ASCII tables.
type TableFormatter struct{}

// FormatList formats a slice of items as a table with headers.
func (f *TableFormatter) FormatList(w io.Writer, headers []string, items interface{}) error {
	v := reflect.ValueOf(items)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("items must be a slice, got %T", items)
	}

	// Handle empty slice
	if v.Len() == 0 {
		return nil
	}

	// Create simple table
	table := tablewriter.NewWriter(w)

	// Set headers
	headerArgs := make([]interface{}, len(headers))
	for i, h := range headers {
		headerArgs[i] = h
	}
	table.Header(headerArgs...)

	// Extract and add rows
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		row, err := extractRow(item, headers)
		if err != nil {
			return err
		}
		// Convert string slice to interface slice for Append
		rowArgs := make([]interface{}, len(row))
		for j, cell := range row {
			rowArgs[j] = cell
		}
		if err := table.Append(rowArgs...); err != nil {
			return err
		}
	}

	return table.Render()
}

// FormatSingle formats a single item as a two-column key-value table.
func (f *TableFormatter) FormatSingle(w io.Writer, item interface{}) error {
	v := reflect.ValueOf(item)

	// Dereference pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return fmt.Errorf("item is nil")
		}
		v = v.Elem()
	}

	// Handle structs and maps
	var pairs [][]string
	switch v.Kind() {
	case reflect.Struct:
		pairs = extractStructPairs(v)
	case reflect.Map:
		pairs = extractMapPairs(v)
	default:
		return fmt.Errorf("unsupported type for single item: %T", item)
	}

	if len(pairs) == 0 {
		return nil
	}

	// Create simple table
	table := tablewriter.NewWriter(w)

	// Convert pairs to [][]interface{} for Bulk
	rows := make([][]interface{}, len(pairs))
	for i, pair := range pairs {
		rows[i] = []interface{}{pair[0], pair[1]}
	}

	if err := table.Bulk(rows); err != nil {
		return err
	}

	return table.Render()
}

// extractRow extracts field values from an item based on headers.
func extractRow(item reflect.Value, headers []string) ([]string, error) {
	// Dereference pointers
	if item.Kind() == reflect.Ptr {
		if item.IsNil() {
			return nil, fmt.Errorf("item is nil")
		}
		item = item.Elem()
	}

	row := make([]string, len(headers))

	switch item.Kind() {
	case reflect.Struct:
		for i, header := range headers {
			fieldName := headerToFieldName(header)
			field := item.FieldByName(fieldName)
			if !field.IsValid() {
				row[i] = ""
				continue
			}
			row[i] = formatValue(field)
		}
	case reflect.Map:
		for i, header := range headers {
			key := reflect.ValueOf(header)
			value := item.MapIndex(key)
			if !value.IsValid() {
				row[i] = ""
				continue
			}
			row[i] = formatValue(value)
		}
	default:
		return nil, fmt.Errorf("unsupported item type: %v", item.Kind())
	}

	return row, nil
}

// extractStructPairs extracts key-value pairs from a struct.
func extractStructPairs(v reflect.Value) [][]string {
	t := v.Type()
	pairs := make([][]string, 0, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Use struct tag if available, otherwise use field name
		key := fieldType.Name
		if tag := fieldType.Tag.Get("json"); tag != "" {
			jsonName := strings.Split(tag, ",")[0]
			if jsonName != "" && jsonName != "-" {
				key = jsonName
			}
		}

		pairs = append(pairs, []string{key, formatValue(field)})
	}

	return pairs
}

// extractMapPairs extracts key-value pairs from a map.
func extractMapPairs(v reflect.Value) [][]string {
	keys := v.MapKeys()
	pairs := make([][]string, 0, len(keys))

	for _, key := range keys {
		value := v.MapIndex(key)
		pairs = append(pairs, []string{
			fmt.Sprintf("%v", key.Interface()),
			formatValue(value),
		})
	}

	return pairs
}

// formatValue formats a reflect.Value as a string.
func formatValue(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}

	// Dereference pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	return fmt.Sprintf("%v", v.Interface())
}

// headerToFieldName converts a header string to a struct field name.
// Examples: "video_id" -> "VideoID", "name" -> "Name"
func headerToFieldName(header string) string {
	parts := strings.Split(header, "_")
	for i, part := range parts {
		if part != "" {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}
