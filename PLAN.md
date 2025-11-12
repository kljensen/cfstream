# Cloudflare Stream CLI - Implementation Plan

## Overview

`cfstream` - A command-line interface for managing Cloudflare Stream videos built with Go and Cobra.

## Core Requirements

1. Upload videos to Cloudflare Stream and get identifiers back
2. Get links (long-lived or signed short-lived) to videos
3. Get embed codes for videos
4. Manage videos (list, get details, update, delete)

## Command Structure

```
cfstream
├── config
│   ├── init           # Interactive setup for API token & account ID
│   └── show           # Display current configuration
├── upload
│   ├── file           # Upload from local file (TUS protocol)
│   ├── url            # Upload from URL
│   └── direct         # Generate direct creator upload URL
├── video
│   ├── list           # List videos with pagination
│   ├── get            # Get video details
│   ├── update         # Update video metadata
│   ├── delete         # Delete video
│   └── search         # Search videos by name
├── link
│   ├── preview        # Get preview URL
│   ├── thumbnail      # Get thumbnail URL
│   ├── hls            # Get HLS manifest URL
│   ├── dash           # Get DASH manifest URL
│   └── signed         # Generate signed token URL
├── embed
│   └── code           # Get HTML embed code
├── download
│   └── create         # Create downloadable version
└── usage              # Show storage usage stats
```

## Configuration

**Config File:** `~/.cfstream.yaml`

```yaml
account_id: "abc123..."
api_token: "xyz789..."
default_output: "table"  # table, json, yaml
default_signed_duration: "1h"
```

**Environment Variables:**
- `CFSTREAM_ACCOUNT_ID`
- `CFSTREAM_API_TOKEN`
- `CFSTREAM_OUTPUT`

## Project Structure

```
cfstream/
├── cmd/
│   ├── root.go              # Root command setup
│   ├── config.go            # Config init/show commands
│   ├── upload.go            # Upload commands
│   ├── video.go             # Video management commands
│   ├── link.go              # Link generation commands
│   ├── embed.go             # Embed code commands
│   └── usage.go             # Usage stats command
├── internal/
│   ├── api/
│   │   ├── client.go        # Cloudflare API client wrapper
│   │   ├── video.go         # Video operations
│   │   ├── upload.go        # Upload operations
│   │   ├── errors.go        # Error handling/mapping
│   │   └── client_test.go   # Tests with mocks
│   ├── config/
│   │   ├── config.go        # Config struct and loading
│   │   ├── validate.go      # Config validation
│   │   └── config_test.go   # Config tests
│   ├── output/
│   │   ├── formatter.go     # Output format switching
│   │   ├── table.go         # Table rendering
│   │   ├── json.go          # JSON rendering
│   │   └── formatter_test.go
│   └── upload/
│       ├── progress.go      # Upload progress tracking
│       └── progress_test.go
├── go.mod
├── go.sum
├── main.go                  # Entry point
├── PLAN.md                  # This file
└── README.md
```

## Dependencies

### Required
```go
github.com/spf13/cobra           // CLI framework
github.com/spf13/viper           // Configuration management
github.com/cloudflare/cloudflare-go/v3  // Official Cloudflare SDK
github.com/olekukonko/tablewriter      // Table output
github.com/schollz/progressbar/v3      // Progress bars
```

### Testing
```go
github.com/stretchr/testify      // Testing utilities and mocks
```

## Implementation Phases

### Phase 1: Foundation ✓
- [x] Write plan to PLAN.md
- [ ] Project setup with Go modules
- [ ] Basic Cobra structure (root command)
- [ ] Config management (`config init`, `config show`)
- [ ] API client wrapper initialization
- [ ] Output formatter (table and JSON)
- [ ] Tests for config and formatter packages

### Phase 2: Core Video Operations
- [ ] `video list` - Proves API connectivity
- [ ] `video get` - Basic resource fetching
- [ ] `video delete` - Includes confirmation UX
- [ ] `video update` - Metadata updates
- [ ] Tests with mocked API client

### Phase 3: Upload (TUS Protocol)
- [ ] `upload file` - Local file with progress (uses TUS)
- [ ] `upload url` - URL-based upload
- [ ] `upload direct` - Direct upload URL generation
- [ ] Tests for upload logic

### Phase 4: Links & Embeds
- [ ] `link preview` - Most common use case
- [ ] `link signed` - Token generation
- [ ] `link thumbnail/hls/dash` - Other link types
- [ ] `embed code` - HTML embed generation
- [ ] Tests for link generation

### Phase 5: Polish
- [ ] `video search` - Search functionality
- [ ] `usage` - Storage stats
- [ ] Better error messages
- [ ] Documentation and examples
- [ ] Integration testing

## Key Design Principles

1. **TUS Protocol**: Use Cloudflare SDK's TUS implementation for resumable uploads
2. **Small Functions**: Keep functions focused and simple
3. **Idiomatic Go**: Follow Go best practices and conventions
4. **Testable Code**: Mock external dependencies (API calls)
5. **Clear Errors**: Helpful error messages with actionable guidance
6. **Table Default**: Human-friendly output by default, JSON for scripts

## Testing Strategy

### Unit Tests
- Config loading and validation (with test fixtures)
- Output formatters (table/JSON rendering)
- API client wrapper (with mocked HTTP/SDK)
- Progress tracking

### Mocking Approach
- Mock Cloudflare SDK client interface
- Use `testify/mock` for generating mocks
- Test files live alongside implementation (`*_test.go`)

### Integration Tests (Future)
- Real API calls against test account
- Upload/download workflows
- Token generation and validation

## TUS Protocol Implementation

The Cloudflare Go SDK handles TUS protocol automatically:
- Resumable uploads for interrupted connections
- Chunked uploads for large files (>200MB)
- Progress callbacks for CLI progress bar
- Automatic retry logic

Our wrapper adds:
- Progress bar display
- Error handling and user-friendly messages
- Status polling after upload completes

## Output Examples

### Table Output (Default)
```
ID              NAME              STATUS   DURATION   CREATED
abc123          My Video          ready    00:05:32   2 hours ago
def456          Another Video     ready    00:12:15   1 day ago
```

### JSON Output
```json
{
  "uid": "abc123",
  "name": "My Video",
  "status": "ready",
  "duration": 332,
  "created": "2025-11-11T10:00:00Z"
}
```

### Upload Progress
```
Uploading video.mp4...
[████████████████████████████████] 100% | 1.2 GB | 45 MB/s
✓ Upload complete

Video ID: abc123def456
Status: Processing
Preview: https://customer-xyz.cloudflarestream.com/abc123def456/manifest/video.m3u8
```

## Error Handling

Exit codes:
- `0` - Success
- `1` - General error
- `2` - Usage error

Error messages include:
- What went wrong
- How to fix it
- Relevant commands or documentation links

## Next Steps

1. Initialize Go module
2. Set up basic Cobra structure
3. Implement config package with tests
4. Build API client wrapper with mocks
5. Implement output formatters
6. Add commands incrementally

---

**Status:** Planning complete, ready for implementation
**Last Updated:** 2025-11-11
