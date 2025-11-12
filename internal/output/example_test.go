package output_test

import (
	"fmt"
	"os"

	"cfstream/internal/output"
)

// Video represents a sample video.
type Video struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration int    `json:"duration"`
}

// ExampleTableFormatter_FormatList demonstrates table output for a list.
func ExampleTableFormatter_FormatList() {
	videos := []Video{
		{ID: "abc123", Name: "My Video", Status: "ready", Duration: 332},
		{ID: "def456", Name: "Another Video", Status: "processing", Duration: 735},
	}

	formatter := &output.TableFormatter{}
	headers := []string{"ID", "Name", "Status"}
	formatter.FormatList(os.Stdout, headers, videos)

	// Output will be a table with headers and two rows
}

// ExampleTableFormatter_FormatSingle demonstrates table output for a single item.
func ExampleTableFormatter_FormatSingle() {
	video := Video{
		ID:       "abc123",
		Name:     "My Video",
		Status:   "ready",
		Duration: 332,
	}

	formatter := &output.TableFormatter{}
	formatter.FormatSingle(os.Stdout, video)

	// Output will be a two-column key-value table
}

// ExampleJSONFormatter_FormatList demonstrates JSON output for a list.
func ExampleJSONFormatter_FormatList() {
	videos := []Video{
		{ID: "abc123", Name: "My Video", Status: "ready", Duration: 332},
	}

	formatter := &output.JSONFormatter{}
	formatter.FormatList(os.Stdout, nil, videos)

	// Output:
	// [
	//   {
	//     "id": "abc123",
	//     "name": "My Video",
	//     "status": "ready",
	//     "duration": 332
	//   }
	// ]
}

// ExampleNewFormatter demonstrates the factory function.
func ExampleNewFormatter() {
	// Create a formatter based on user preference
	format := "json" // Could be "table", "json", or "yaml"
	formatter, err := output.NewFormatter(format)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	video := Video{
		ID:     "abc123",
		Name:   "Test Video",
		Status: "ready",
	}

	formatter.FormatSingle(os.Stdout, video)
}
