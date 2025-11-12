package output

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// testVideo represents a sample video for testing.
type testVideo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration int    `json:"duration"`
}

// testVideoPtr is a version with pointer fields to test pointer handling.
type testVideoPtr struct {
	ID     *string `json:"id"`
	Name   *string `json:"name"`
	Status *string `json:"status"`
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
		wantTyp interface{}
	}{
		{
			name:    "table formatter",
			format:  "table",
			wantErr: false,
			wantTyp: &TableFormatter{},
		},
		{
			name:    "json formatter",
			format:  "json",
			wantErr: false,
			wantTyp: &JSONFormatter{},
		},
		{
			name:    "yaml formatter",
			format:  "yaml",
			wantErr: false,
			wantTyp: &YAMLFormatter{},
		},
		{
			name:    "invalid formatter",
			format:  "xml",
			wantErr: true,
		},
		{
			name:    "empty format",
			format:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := NewFormatter(tt.format)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, formatter)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, formatter)
			assert.IsType(t, tt.wantTyp, formatter)
		})
	}
}

func TestJSONFormatter_FormatList(t *testing.T) {
	formatter := &JSONFormatter{}

	tests := []struct {
		name    string
		items   interface{}
		headers []string
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name: "format list of videos",
			items: []testVideo{
				{ID: "vid1", Name: "Video 1", Status: "ready", Duration: 120},
				{ID: "vid2", Name: "Video 2", Status: "processing", Duration: 300},
			},
			headers: []string{"ID", "Name", "Status"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				var videos []testVideo
				err := json.Unmarshal([]byte(output), &videos)
				require.NoError(t, err)
				assert.Len(t, videos, 2)
				assert.Equal(t, "vid1", videos[0].ID)
				assert.Equal(t, "Video 2", videos[1].Name)
			},
		},
		{
			name:    "format empty list",
			items:   []testVideo{},
			headers: []string{"ID", "Name"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				var videos []testVideo
				err := json.Unmarshal([]byte(output), &videos)
				require.NoError(t, err)
				assert.Len(t, videos, 0)
			},
		},
		{
			name: "format list of maps",
			items: []map[string]interface{}{
				{"id": "vid1", "name": "Video 1"},
				{"id": "vid2", "name": "Video 2"},
			},
			headers: []string{"id", "name"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				var items []map[string]interface{}
				err := json.Unmarshal([]byte(output), &items)
				require.NoError(t, err)
				assert.Len(t, items, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := formatter.FormatList(&buf, tt.headers, tt.items)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, buf.String())
			}
		})
	}
}

func TestJSONFormatter_FormatSingle(t *testing.T) {
	formatter := &JSONFormatter{}

	tests := []struct {
		name    string
		item    interface{}
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name: "format single video",
			item: testVideo{
				ID:       "vid1",
				Name:     "Test Video",
				Status:   "ready",
				Duration: 120,
			},
			wantErr: false,
			check: func(t *testing.T, output string) {
				var video testVideo
				err := json.Unmarshal([]byte(output), &video)
				require.NoError(t, err)
				assert.Equal(t, "vid1", video.ID)
				assert.Equal(t, "Test Video", video.Name)
				assert.Equal(t, 120, video.Duration)
			},
		},
		{
			name: "format single map",
			item: map[string]interface{}{
				"id":   "vid1",
				"name": "Test Video",
			},
			wantErr: false,
			check: func(t *testing.T, output string) {
				var item map[string]interface{}
				err := json.Unmarshal([]byte(output), &item)
				require.NoError(t, err)
				assert.Equal(t, "vid1", item["id"])
			},
		},
		{
			name:    "format nil item",
			item:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := formatter.FormatSingle(&buf, tt.item)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, buf.String())
			}
		})
	}
}

func TestYAMLFormatter_FormatList(t *testing.T) {
	formatter := &YAMLFormatter{}

	tests := []struct {
		name    string
		items   interface{}
		headers []string
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name: "format list of videos",
			items: []testVideo{
				{ID: "vid1", Name: "Video 1", Status: "ready", Duration: 120},
				{ID: "vid2", Name: "Video 2", Status: "processing", Duration: 300},
			},
			headers: []string{"ID", "Name", "Status"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				var videos []testVideo
				err := yaml.Unmarshal([]byte(output), &videos)
				require.NoError(t, err)
				assert.Len(t, videos, 2)
				assert.Equal(t, "vid1", videos[0].ID)
				assert.Equal(t, "Video 2", videos[1].Name)
			},
		},
		{
			name:    "format empty list",
			items:   []testVideo{},
			headers: []string{"ID", "Name"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				var videos []testVideo
				err := yaml.Unmarshal([]byte(output), &videos)
				require.NoError(t, err)
				assert.Len(t, videos, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := formatter.FormatList(&buf, tt.headers, tt.items)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, buf.String())
			}
		})
	}
}

func TestYAMLFormatter_FormatSingle(t *testing.T) {
	formatter := &YAMLFormatter{}

	tests := []struct {
		name    string
		item    interface{}
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name: "format single video",
			item: testVideo{
				ID:       "vid1",
				Name:     "Test Video",
				Status:   "ready",
				Duration: 120,
			},
			wantErr: false,
			check: func(t *testing.T, output string) {
				var video testVideo
				err := yaml.Unmarshal([]byte(output), &video)
				require.NoError(t, err)
				assert.Equal(t, "vid1", video.ID)
				assert.Equal(t, "Test Video", video.Name)
			},
		},
		{
			name:    "format nil item",
			item:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := formatter.FormatSingle(&buf, tt.item)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, buf.String())
			}
		})
	}
}

func TestTableFormatter_FormatList(t *testing.T) {
	formatter := &TableFormatter{}

	tests := []struct {
		name    string
		items   interface{}
		headers []string
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name: "format list of videos",
			items: []testVideo{
				{ID: "vid1", Name: "Video 1", Status: "ready", Duration: 120},
				{ID: "vid2", Name: "Video 2", Status: "processing", Duration: 300},
			},
			headers: []string{"ID", "Name", "Status"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				// Headers may be uppercased by tablewriter
				upperOutput := strings.ToUpper(output)
				assert.Contains(t, upperOutput, "ID")
				assert.Contains(t, upperOutput, "NAME")
				assert.Contains(t, upperOutput, "STATUS")
				// Data should be present
				assert.Contains(t, output, "vid1")
				assert.Contains(t, output, "Video 1")
				assert.Contains(t, output, "ready")
				assert.Contains(t, output, "vid2")
				assert.Contains(t, output, "Video 2")
			},
		},
		{
			name:    "format empty list",
			items:   []testVideo{},
			headers: []string{"ID", "Name"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				// Empty list should produce no output
				assert.Empty(t, output)
			},
		},
		{
			name: "format list of maps",
			items: []map[string]string{
				{"ID": "vid1", "Name": "Video 1"},
				{"ID": "vid2", "Name": "Video 2"},
			},
			headers: []string{"ID", "Name"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "vid1")
				assert.Contains(t, output, "Video 1")
			},
		},
		{
			name:    "format non-slice",
			items:   "not a slice",
			headers: []string{"ID"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := formatter.FormatList(&buf, tt.headers, tt.items)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, buf.String())
			}
		})
	}
}

func TestTableFormatter_FormatSingle(t *testing.T) {
	formatter := &TableFormatter{}

	tests := []struct {
		name    string
		item    interface{}
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name: "format single video struct",
			item: testVideo{
				ID:       "vid1",
				Name:     "Test Video",
				Status:   "ready",
				Duration: 120,
			},
			wantErr: false,
			check: func(t *testing.T, output string) {
				// Check that key-value pairs are present
				assert.Contains(t, output, "id")
				assert.Contains(t, output, "vid1")
				assert.Contains(t, output, "name")
				assert.Contains(t, output, "Test Video")
				assert.Contains(t, output, "status")
				assert.Contains(t, output, "ready")
			},
		},
		{
			name: "format single video pointer",
			item: &testVideo{
				ID:       "vid1",
				Name:     "Test Video",
				Status:   "ready",
				Duration: 120,
			},
			wantErr: false,
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "vid1")
				assert.Contains(t, output, "Test Video")
			},
		},
		{
			name: "format single map",
			item: map[string]string{
				"id":   "vid1",
				"name": "Test Video",
			},
			wantErr: false,
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "id")
				assert.Contains(t, output, "vid1")
				assert.Contains(t, output, "name")
				assert.Contains(t, output, "Test Video")
			},
		},
		{
			name:    "format nil pointer",
			item:    (*testVideo)(nil),
			wantErr: true,
		},
		{
			name:    "format unsupported type",
			item:    []string{"not", "a", "struct"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := formatter.FormatSingle(&buf, tt.item)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, buf.String())
			}
		})
	}
}

func TestTableFormatter_HeaderToFieldName(t *testing.T) {
	tests := []struct {
		header string
		want   string
	}{
		{"id", "Id"},
		{"name", "Name"},
		{"video_id", "VideoId"},
		{"created_at", "CreatedAt"},
		{"ID", "ID"},
		{"api_token", "ApiToken"},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := headerToFieldName(tt.header)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTableFormatter_FormatValue(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{
			name:  "string value",
			value: "test",
			want:  "test",
		},
		{
			name:  "int value",
			value: 123,
			want:  "123",
		},
		{
			name:  "bool value",
			value: true,
			want:  "true",
		},
		{
			name:  "nil pointer",
			value: (*string)(nil),
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a reflect.Value from the interface
			var v interface{} = tt.value
			if tt.value == (*string)(nil) {
				// Special case for nil pointer
				var nilPtr *string
				v = nilPtr
			}

			rv := formatValue(reflect.ValueOf(v))
			if tt.value == (*string)(nil) {
				assert.Empty(t, rv)
			} else {
				assert.Equal(t, tt.want, rv)
			}
		})
	}
}

func TestTableFormatter_WithPointerFields(t *testing.T) {
	formatter := &TableFormatter{}

	id := "vid1"
	name := "Test Video"
	status := "ready"

	video := testVideoPtr{
		ID:     &id,
		Name:   &name,
		Status: &status,
	}

	var buf bytes.Buffer
	err := formatter.FormatSingle(&buf, video)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "vid1")
	assert.Contains(t, output, "Test Video")
	assert.Contains(t, output, "ready")
}

func TestTableFormatter_WithNilPointerFields(t *testing.T) {
	formatter := &TableFormatter{}

	id := "vid1"
	video := testVideoPtr{
		ID:     &id,
		Name:   nil,
		Status: nil,
	}

	var buf bytes.Buffer
	err := formatter.FormatSingle(&buf, video)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "vid1")
	// Nil fields should still be present but empty
	lines := strings.Split(output, "\n")
	// Should have at least 3 lines (id, name, status fields)
	assert.GreaterOrEqual(t, len(lines), 3)
}
