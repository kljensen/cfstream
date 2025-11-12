# cfstream

A command-line interface for managing Cloudflare Stream videos.

## Features

- **Upload videos** using TUS protocol (resumable uploads for large files)
- **Manage videos** (list, get details, update metadata, delete)
- **Generate links** (preview, signed URLs, thumbnails, HLS, DASH)
- **Get embed codes** with customization options
- **Multiple output formats** (table, JSON, YAML)
- **Interactive configuration** setup

## Installation

### From Source

```bash
git clone <repository-url>
cd kyletube
go build -o cfstream
```

### Install to PATH

```bash
go install
```

## Quick Start

### 1. Configure Your Credentials

Run the interactive configuration setup:

```bash
cfstream config init
```

This will prompt you for:
- Account ID (found in your Cloudflare dashboard)
- API Token (create one at https://dash.cloudflare.com/profile/api-tokens)
- Default output format (table/json/yaml)
- Default signed URL duration

Or use environment variables:

```bash
export CFSTREAM_ACCOUNT_ID=your_account_id
export CFSTREAM_API_TOKEN=your_api_token
```

### 2. Upload Your First Video

```bash
cfstream upload file video.mp4 --name "My First Video"
```

This will:
- Upload the video using TUS protocol (resumable)
- Show a progress bar
- Wait for processing to complete
- Return the video ID and preview URL

### 3. Get Links to Your Video

```bash
# Get preview URL
cfstream link preview VIDEO_ID

# Generate signed URL (expires in 2 hours)
cfstream link signed VIDEO_ID --duration 2h

# Get thumbnail
cfstream link thumbnail VIDEO_ID --time 30s
```

### 4. Get Embed Code

```bash
cfstream embed code VIDEO_ID --responsive --autoplay --muted
```

## Commands

### Configuration Management

```bash
# Interactive setup
cfstream config init

# Show current configuration
cfstream config show
```

### Upload Videos

```bash
# Upload local file with TUS protocol
cfstream upload file video.mp4
cfstream upload file video.mp4 --name "My Video" --metadata '{"category":"tutorial"}'

# Upload from URL
cfstream upload url https://example.com/video.mp4

# Generate direct upload URL (for end users)
cfstream upload direct --expires 2h --max-duration 3600
```

### Manage Videos

```bash
# List all videos
cfstream video list

# List with filtering
cfstream video list --search "tutorial" --limit 100 --status ready

# Get video details
cfstream video get VIDEO_ID

# Update video metadata
cfstream video update VIDEO_ID --name "New Name" --metadata '{"key":"value"}'

# Delete video
cfstream video delete VIDEO_ID
cfstream video delete VIDEO_ID --yes  # Skip confirmation
```

### Get Video Links

```bash
# Preview/HLS manifest URL
cfstream link preview VIDEO_ID

# Signed URL (short-lived)
cfstream link signed VIDEO_ID --duration 2h

# Thumbnail URL
cfstream link thumbnail VIDEO_ID --time 30s

# HLS manifest
cfstream link hls VIDEO_ID

# DASH manifest
cfstream link dash VIDEO_ID
```

### Get Embed Code

```bash
# Basic embed code
cfstream embed code VIDEO_ID

# Responsive embed with autoplay
cfstream embed code VIDEO_ID --responsive --autoplay --muted --loop
```

## Output Formats

All commands support multiple output formats:

```bash
# Table output (default, human-readable)
cfstream video list

# JSON output (for scripts)
cfstream video list --output json

# YAML output
cfstream video list --output yaml
```

## Global Flags

- `--output, -o` - Output format (table, json, yaml)
- `--quiet, -q` - Suppress non-essential output
- `--verbose, -v` - Verbose output
- `--version` - Show version

## Examples

### Upload and Share Workflow

```bash
# Upload a video
cfstream upload file presentation.mp4 --name "Q4 Presentation"

# Get the video ID from output (e.g., abc123def456)

# Generate a signed URL valid for 24 hours
cfstream link signed abc123def456 --duration 24h

# Or get an embed code
cfstream embed code abc123def456 --responsive
```

### Batch Operations with JSON

```bash
# List all videos as JSON
cfstream video list --output json > videos.json

# Get specific video details
cfstream video get abc123def456 --output json | jq '.name'
```

### Search and Filter

```bash
# Find videos by name
cfstream video list --search "tutorial"

# List only ready videos
cfstream video list --status ready --limit 50
```

## Features

### TUS Protocol for Resumable Uploads

The CLI uses the TUS protocol for file uploads, which provides:
- Resumable uploads for large files (>200MB)
- Chunked uploads (50MB chunks)
- Progress tracking with speed and ETA
- Automatic retry on network failures

### Configuration Flexibility

Configuration can be loaded from:
1. Config file (`~/.cfstream.yaml`)
2. Environment variables (override config file)
3. Default values

Environment variables take precedence over config file values.

### Error Handling

The CLI provides clear, actionable error messages:
- Configuration errors → Run `cfstream config init`
- Authentication errors → Check your API token
- Not found errors → Verify video ID
- Rate limit errors → Wait and retry

## Development

### Project Structure

```
cfstream/
├── cmd/                # Cobra commands
│   ├── root.go
│   ├── config.go
│   ├── upload.go
│   ├── video.go
│   ├── link.go
│   └── embed.go
├── internal/
│   ├── api/            # Cloudflare API client
│   ├── config/         # Configuration management
│   ├── output/         # Output formatters
│   └── upload/         # Upload progress tracking
├── main.go
├── PLAN.md            # Implementation plan
└── README.md
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/config/
```

### Building

```bash
# Build for current platform
go build -o cfstream

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o cfstream-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o cfstream-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o cfstream-windows-amd64.exe
```

## Implementation Status

✅ **Phase 1: Foundation**
- Project structure and dependencies
- Configuration management
- Output formatters (table, JSON, YAML)
- API client wrapper

✅ **Phase 2: Core Commands**
- Config commands (init, show)
- Video commands (list, get, update, delete)
- Link commands (preview, signed, thumbnail, HLS, DASH)
- Embed commands
- Upload commands with TUS protocol

See [PLAN.md](PLAN.md) for the complete implementation plan.

## License

MIT
