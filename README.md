# cfstream

[![CI](https://img.shields.io/github/actions/workflow/status/kljensen/cfstream/ci.yml?branch=main&style=for-the-badge&logo=github)](https://github.com/kljensen/cfstream/actions/workflows/ci.yml)
[![Go Reference](https://img.shields.io/badge/go-reference-007d9c?style=for-the-badge&logo=go)](https://pkg.go.dev/github.com/kljensen/cfstream)
[![Go Report Card](https://goreportcard.com/badge/github.com/kljensen/cfstream?style=for-the-badge)](https://goreportcard.com/report/github.com/kljensen/cfstream)
[![Go Version](https://img.shields.io/github/go-mod/go-version/kljensen/cfstream?style=for-the-badge&logo=go)](go.mod)
[![License](https://img.shields.io/github/license/kljensen/cfstream?style=for-the-badge)](LICENSE)

A minimal, fast command-line interface for managing [Cloudflare Stream](https://www.cloudflare.com/developer-platform/products/cloudflare-stream/) videos.

## Features

- 🎥 **Upload videos** with resumable TUS protocol for large files
- 📊 **Manage videos** - list, get, update, delete with rich filtering
- 🔗 **Generate links** - preview, signed URLs, thumbnails, HLS/DASH manifests
- 📦 **Embed codes** - responsive iframes with customization
- 🎨 **Multiple outputs** - table, JSON, or YAML formats
- ⚡ **Fast & efficient** - direct API integration with progress tracking

## Installation

```bash
go install github.com/kljensen/cfstream@latest
```

Or build from source:

```bash
git clone https://github.com/kljensen/cfstream.git
cd cfstream
go build -o cfstream
```

## Quick Start

### 1. Configure credentials

```bash
cfstream config init
```

Or use environment variables:

```bash
export CFSTREAM_ACCOUNT_ID=your_account_id
export CFSTREAM_API_TOKEN=your_api_token
```

### 2. Upload a video

```bash
cfstream upload file video.mp4 --name "My Video"
```

### 3. Get video links

```bash
# Preview URL
cfstream link preview VIDEO_ID

# Signed URL (expires in 24h)
cfstream link signed VIDEO_ID --duration 24h

# Thumbnail
cfstream link thumbnail VIDEO_ID --time 30s
```

### 4. Get embed code

```bash
cfstream embed code VIDEO_ID --responsive --autoplay --muted
```

## Commands

### Configuration

```bash
cfstream config init              # Interactive setup
cfstream config show              # Display current config
```

### Upload

```bash
cfstream upload file video.mp4    # Upload local file
cfstream upload url <url>         # Upload from URL
cfstream upload direct            # Generate direct upload URL
```

### Video Management

```bash
cfstream video list               # List all videos
cfstream video get VIDEO_ID       # Get video details
cfstream video update VIDEO_ID    # Update metadata
cfstream video delete VIDEO_ID    # Delete video
```

### Links

```bash
cfstream link preview VIDEO_ID    # Preview URL
cfstream link signed VIDEO_ID     # Signed URL
cfstream link thumbnail VIDEO_ID  # Thumbnail URL
cfstream link hls VIDEO_ID        # HLS manifest
cfstream link dash VIDEO_ID       # DASH manifest
```

### Embed

```bash
cfstream embed code VIDEO_ID      # Get iframe embed code
```

## Output Formats

Use `--output` or `-o` to change the output format:

```bash
cfstream video list --output json   # JSON output
cfstream video list --output yaml   # YAML output
cfstream video list                 # Table output (default)
```

## Global Flags

- `--output, -o` - Output format (table, json, yaml)
- `--quiet, -q` - Suppress non-essential output
- `--verbose, -v` - Verbose output
- `--help, -h` - Show help
- `--version` - Show version

## Examples

### Upload and share workflow

```bash
# Upload
cfstream upload file presentation.mp4 --name "Q4 Presentation"

# Generate signed URL (24h expiry)
cfstream link signed abc123def456 --duration 24h

# Or get embed code
cfstream embed code abc123def456 --responsive
```

### Batch operations with JSON

```bash
# Export all videos
cfstream video list --output json > videos.json

# Parse with jq
cfstream video get VIDEO_ID --output json | jq '.name'
```

### Search and filter

```bash
# Find by name
cfstream video list --search "tutorial"

# Filter by status
cfstream video list --status ready --limit 50
```

## Configuration

Configuration is loaded from (in order of precedence):

1. Environment variables (`CFSTREAM_*`)
2. Config file (`~/.config/cfstream/config.yaml`)
3. Defaults

### Environment Variables

- `CFSTREAM_ACCOUNT_ID` - Cloudflare account ID
- `CFSTREAM_API_TOKEN` - API token
- `CFSTREAM_OUTPUT` - Default output format

## Development

### Run tests

```bash
go test ./...
```

### Run linter

```bash
golangci-lint run
```

### Build

```bash
go build -o cfstream
```

## License

MIT
