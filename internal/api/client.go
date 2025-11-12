package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/cloudflare/cloudflare-go/v3/stream"
)

// Client defines the interface for interacting with Cloudflare Stream API.
type Client interface {
	// ListVideos retrieves a list of videos with optional filtering.
	ListVideos(ctx context.Context, opts *ListOptions) ([]Video, error)

	// GetVideo retrieves details for a specific video by ID.
	GetVideo(ctx context.Context, videoID string) (*Video, error)

	// DeleteVideo deletes a video by ID.
	DeleteVideo(ctx context.Context, videoID string) error

	// UpdateVideo updates video metadata.
	UpdateVideo(ctx context.Context, videoID string, opts *UpdateOptions) (*Video, error)

	// GetSignedToken generates a signed token for a video.
	GetSignedToken(ctx context.Context, videoID string, duration int64) (string, error)

	// GetEmbedCode returns the HTML embed code for a video.
	GetEmbedCode(ctx context.Context, videoID string, opts *EmbedOptions) (string, error)

	// UploadFile uploads a video file using multipart/form-data.
	UploadFile(ctx context.Context, filePath string, opts *UploadOptions, progressCh chan<- UploadProgress) (*Video, error)

	// UploadFromURL uploads a video from a URL.
	UploadFromURL(ctx context.Context, url string, opts *UploadOptions) (*Video, error)

	// CreateDirectUploadURL generates a direct upload URL for end users.
	CreateDirectUploadURL(ctx context.Context, opts *DirectUploadOptions) (*DirectUploadResult, error)
}

// ClientImpl implements the Client interface using the Cloudflare SDK.
type ClientImpl struct {
	sdk       *cloudflare.Client
	accountID string
	apiToken  string
}

// NewClient creates a new Cloudflare Stream API client.
func NewClient(accountID, apiToken string) (Client, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}
	if apiToken == "" {
		return nil, fmt.Errorf("API token is required")
	}

	sdk := cloudflare.NewClient(
		option.WithAPIToken(apiToken),
	)

	return &ClientImpl{
		sdk:       sdk,
		accountID: accountID,
		apiToken:  apiToken,
	}, nil
}

// ListVideos retrieves a list of videos with optional filtering.
func (c *ClientImpl) ListVideos(ctx context.Context, opts *ListOptions) ([]Video, error) {
	params := stream.StreamListParams{
		AccountID: cloudflare.F(c.accountID),
	}

	// Apply options if provided
	if opts != nil {
		if opts.Search != "" {
			params.Search = cloudflare.F(opts.Search)
		}
		if opts.Creator != "" {
			params.Creator = cloudflare.F(opts.Creator)
		}
		if opts.Start != nil {
			params.Start = cloudflare.F(*opts.Start)
		}
		if opts.End != nil {
			params.End = cloudflare.F(*opts.End)
		}
		if opts.Asc {
			params.Asc = cloudflare.F(true)
		}
	}

	page, err := c.sdk.Stream.List(ctx, params)
	if err != nil {
		return nil, WrapError(err)
	}

	// Extract videos from page
	videos := page.Result
	return VideosFromSDK(videos), nil
}

// GetVideo retrieves details for a specific video by ID.
func (c *ClientImpl) GetVideo(ctx context.Context, videoID string) (*Video, error) {
	if videoID == "" {
		return nil, fmt.Errorf("%w: video ID cannot be empty", ErrInvalidInput)
	}

	params := stream.StreamGetParams{
		AccountID: cloudflare.F(c.accountID),
	}

	video, err := c.sdk.Stream.Get(ctx, videoID, params)
	if err != nil {
		return nil, WrapError(err)
	}

	return VideoFromSDK(video), nil
}

// DeleteVideo deletes a video by ID.
func (c *ClientImpl) DeleteVideo(ctx context.Context, videoID string) error {
	if videoID == "" {
		return fmt.Errorf("%w: video ID cannot be empty", ErrInvalidInput)
	}

	params := stream.StreamDeleteParams{
		AccountID: cloudflare.F(c.accountID),
	}

	err := c.sdk.Stream.Delete(ctx, videoID, params)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// UpdateVideo updates video metadata.
func (c *ClientImpl) UpdateVideo(ctx context.Context, videoID string, opts *UpdateOptions) (*Video, error) {
	if videoID == "" {
		return nil, fmt.Errorf("%w: video ID cannot be empty", ErrInvalidInput)
	}
	if opts == nil {
		return nil, fmt.Errorf("%w: update options cannot be nil", ErrInvalidInput)
	}

	// Build the request body
	body := make(map[string]interface{})
	if opts.Meta != nil {
		body["meta"] = opts.Meta
	}
	if opts.RequireSignedURLs != nil {
		body["requireSignedURLs"] = *opts.RequireSignedURLs
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Make direct HTTP request to update video
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/stream/%s", c.accountID, videoID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var apiResp struct {
		Result  stream.Video `json:"result"`
		Success bool         `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		if len(apiResp.Errors) > 0 {
			return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("API request failed")
	}

	return VideoFromSDK(&apiResp.Result), nil
}

// GetSignedToken generates a signed token for a video.
func (c *ClientImpl) GetSignedToken(ctx context.Context, videoID string, duration int64) (string, error) {
	if videoID == "" {
		return "", fmt.Errorf("%w: video ID cannot be empty", ErrInvalidInput)
	}

	// Build request body with expiration time
	body := make(map[string]interface{})
	if duration > 0 {
		body["exp"] = duration
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Make HTTP request to create token
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/stream/%s/token", c.accountID, videoID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var apiResp struct {
		Result struct {
			Token string `json:"token"`
		} `json:"result"`
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		if len(apiResp.Errors) > 0 {
			return "", fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
		}
		return "", fmt.Errorf("API request failed")
	}

	return apiResp.Result.Token, nil
}

// GetEmbedCode returns the HTML embed code for a video.
func (c *ClientImpl) GetEmbedCode(ctx context.Context, videoID string, opts *EmbedOptions) (string, error) {
	if videoID == "" {
		return "", fmt.Errorf("%w: video ID cannot be empty", ErrInvalidInput)
	}

	// Get video details to extract customer code
	video, err := c.GetVideo(ctx, videoID)
	if err != nil {
		return "", fmt.Errorf("failed to get video details: %w", err)
	}

	// Extract customer code from preview URL
	customerCode, err := extractCustomerCode(video.Preview)
	if err != nil {
		return "", fmt.Errorf("failed to extract customer code: %w", err)
	}

	// Build iframe URL with query parameters
	iframeURL := fmt.Sprintf("https://customer-%s.cloudflarestream.com/%s/iframe", customerCode, videoID)

	queryParams := make([]string, 0)
	if opts != nil {
		// Add signed token first if present
		if opts.SignedToken != "" {
			queryParams = append(queryParams, fmt.Sprintf("token=%s", opts.SignedToken))
		}
		if opts.Autoplay {
			queryParams = append(queryParams, "autoplay=true")
		}
		if opts.Muted {
			queryParams = append(queryParams, "muted=true")
		}
		if opts.Loop {
			queryParams = append(queryParams, "loop=true")
		}
		if !opts.Controls {
			queryParams = append(queryParams, "controls=false")
		}
	}

	if len(queryParams) > 0 {
		iframeURL += "?" + strings.Join(queryParams, "&")
	}

	// Build iframe HTML
	style := "border: none;"
	if opts != nil && opts.Responsive {
		// Responsive style with 16:9 aspect ratio
		return fmt.Sprintf(`<div style="position: relative; padding-top: 56.25%%;">
  <iframe
    src="%s"
    style="border: none; position: absolute; top: 0; left: 0; height: 100%%; width: 100%%;"
    allow="accelerometer; gyroscope; autoplay; encrypted-media; picture-in-picture;"
    allowfullscreen="true">
  </iframe>
</div>`, iframeURL), nil
	}

	return fmt.Sprintf(`<iframe
  src="%s"
  style="%s"
  height="720"
  width="1280"
  allow="accelerometer; gyroscope; autoplay; encrypted-media; picture-in-picture;"
  allowfullscreen="true">
</iframe>`, iframeURL, style), nil
}

// extractCustomerCode extracts the customer code from a preview URL.
func extractCustomerCode(previewURL string) (string, error) {
	if previewURL == "" {
		return "", fmt.Errorf("preview URL is empty")
	}

	// URL format: https://customer-{code}.cloudflarestream.com/{videoID}/manifest/video.m3u8
	parts := strings.Split(previewURL, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid preview URL format")
	}

	// Extract customer code from subdomain
	subdomain := parts[0]
	prefix := "https://customer-"
	if !strings.HasPrefix(subdomain, prefix) {
		return "", fmt.Errorf("invalid preview URL format: missing customer prefix")
	}

	code := strings.TrimPrefix(subdomain, prefix)
	if code == "" {
		return "", fmt.Errorf("customer code is empty")
	}

	return code, nil
}

// CreateDirectUploadURL generates a direct upload URL for end users.
func (c *ClientImpl) CreateDirectUploadURL(ctx context.Context, opts *DirectUploadOptions) (*DirectUploadResult, error) {
	if opts == nil {
		opts = &DirectUploadOptions{}
	}

	// Build request body
	body := make(map[string]interface{})
	if opts.MaxDurationSeconds > 0 {
		body["maxDurationSeconds"] = opts.MaxDurationSeconds
	}
	if opts.Expiry != nil {
		body["expiry"] = opts.Expiry.Format(time.RFC3339)
	}
	if opts.RequireSignedURLs {
		body["requireSignedURLs"] = true
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Make HTTP request
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/stream/direct_upload", c.accountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var apiResp struct {
		Result struct {
			UploadURL string `json:"uploadURL"`
			UID       string `json:"uid"`
		} `json:"result"`
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		if len(apiResp.Errors) > 0 {
			return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("API request failed")
	}

	result := &DirectUploadResult{
		UploadURL: apiResp.Result.UploadURL,
		UID:       apiResp.Result.UID,
	}

	if opts.Expiry != nil {
		result.Expiry = *opts.Expiry
	}

	return result, nil
}

// UploadFromURL uploads a video from a URL.
func (c *ClientImpl) UploadFromURL(ctx context.Context, url string, opts *UploadOptions) (*Video, error) {
	if url == "" {
		return nil, fmt.Errorf("%w: URL cannot be empty", ErrInvalidInput)
	}
	if opts == nil {
		opts = &UploadOptions{}
	}

	// Build request body
	body := make(map[string]interface{})
	body["url"] = url
	body["requireSignedURLs"] = true

	// Add metadata if provided
	meta := make(map[string]interface{})
	if opts.Name != "" {
		meta["name"] = opts.Name
	}
	if opts.Metadata != nil {
		for k, v := range opts.Metadata {
			meta[k] = v
		}
	}
	if len(meta) > 0 {
		body["meta"] = meta
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Make HTTP request
	apiURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/stream/copy", c.accountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var apiResp struct {
		Result  stream.Video `json:"result"`
		Success bool         `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		if len(apiResp.Errors) > 0 {
			return nil, fmt.Errorf("API error: %s", apiResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("API request failed")
	}

	return VideoFromSDK(&apiResp.Result), nil
}

// UploadFile uploads a video file using multipart/form-data or TUS protocol.
func (c *ClientImpl) UploadFile(ctx context.Context, filePath string, opts *UploadOptions, progressCh chan<- UploadProgress) (*Video, error) {
	if filePath == "" {
		return nil, fmt.Errorf("%w: file path cannot be empty", ErrInvalidInput)
	}
	if opts == nil {
		opts = &UploadOptions{}
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	// Choose upload method based on file size
	const tusThreshold = 200 * 1024 * 1024 // 200 MB

	if fileSize >= tusThreshold {
		// Use TUS for large files
		tusURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/stream", c.accountID)
		videoID, err := c.tusUploadDirect(ctx, tusURL, file, fileSize, opts, progressCh)
		if err != nil {
			return nil, fmt.Errorf("TUS upload failed: %w", err)
		}

		// Get the video details
		video, err := c.GetVideo(ctx, videoID)
		if err != nil {
			return nil, fmt.Errorf("failed to get video details: %w", err)
		}

		return video, nil
	}

	// For smaller files, use direct upload URL with multipart
	directOpts := &DirectUploadOptions{
		MaxDurationSeconds: 21600, // 6 hours max video duration
		RequireSignedURLs:  true,
	}
	directResult, err := c.CreateDirectUploadURL(ctx, directOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create direct upload URL: %w", err)
	}

	// Upload using multipart/form-data
	if err := c.multipartUpload(ctx, directResult.UploadURL, file, fileSize, opts, progressCh); err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	// Get the video details
	video, err := c.GetVideo(ctx, directResult.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video details: %w", err)
	}

	return video, nil
}

// multipartUpload performs a multipart/form-data upload.
func (c *ClientImpl) multipartUpload(ctx context.Context, uploadURL string, file *os.File, fileSize int64, opts *UploadOptions, progressCh chan<- UploadProgress) error {
	// Create a pipe for streaming the multipart data
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// Start writing the multipart data in a goroutine
	go func() {
		defer pw.Close()
		defer writer.Close()

		// Add the file field
		part, err := writer.CreateFormFile("file", file.Name())
		if err != nil {
			pw.CloseWithError(err)
			return
		}

		// Copy file to part with progress tracking
		buffer := make([]byte, 1024*1024) // 1MB buffer
		var written int64
		for {
			n, err := file.Read(buffer)
			if n > 0 {
				_, writeErr := part.Write(buffer[:n])
				if writeErr != nil {
					pw.CloseWithError(writeErr)
					return
				}
				written += int64(n)

				// Send progress update
				if progressCh != nil {
					select {
					case progressCh <- UploadProgress{BytesSent: written, BytesTotal: fileSize}:
					default:
					}
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				pw.CloseWithError(err)
				return
			}
		}
	}()

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, pr)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// tusUploadDirect uploads directly to the Stream TUS endpoint (for large files).
func (c *ClientImpl) tusUploadDirect(ctx context.Context, tusURL string, file *os.File, fileSize int64, opts *UploadOptions, progressCh chan<- UploadProgress) (string, error) {
	// Build Upload-Metadata header
	var metadataParts []string
	if opts.Name != "" {
		encoded := fmt.Sprintf("name %s", base64.StdEncoding.EncodeToString([]byte(opts.Name)))
		metadataParts = append(metadataParts, encoded)
	}
	uploadMetadata := strings.Join(metadataParts, ",")

	// Create initial TUS request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tusURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create TUS request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Length", fmt.Sprintf("%d", fileSize))
	if uploadMetadata != "" {
		req.Header.Set("Upload-Metadata", uploadMetadata)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to initiate TUS upload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("TUS upload initiation failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Get upload URL from Location header
	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("TUS upload location not returned")
	}

	// Extract video ID from Location header
	// Location format: https://api.cloudflare.com/client/v4/accounts/{account_id}/stream/{video_id}
	locationParts := strings.Split(location, "/")
	if len(locationParts) == 0 {
		return "", fmt.Errorf("failed to extract video ID from location header")
	}
	videoID := locationParts[len(locationParts)-1]

	// Upload file in chunks (50 MB)
	const chunkSize = 50 * 1024 * 1024
	buffer := make([]byte, chunkSize)
	var offset int64

	for {
		n, err := file.Read(buffer)
		if n == 0 && err == io.EOF {
			break
		}
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to read file: %w", err)
		}

		// Upload chunk
		chunkReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, location, bytes.NewReader(buffer[:n]))
		if err != nil {
			return "", fmt.Errorf("failed to create chunk request: %w", err)
		}

		chunkReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
		chunkReq.Header.Set("Tus-Resumable", "1.0.0")
		chunkReq.Header.Set("Upload-Offset", fmt.Sprintf("%d", offset))
		chunkReq.Header.Set("Content-Type", "application/offset+octet-stream")
		chunkReq.Header.Set("Content-Length", fmt.Sprintf("%d", n))

		chunkResp, err := client.Do(chunkReq)
		if err != nil {
			return "", fmt.Errorf("chunk upload failed: %w", err)
		}
		defer chunkResp.Body.Close()

		if chunkResp.StatusCode != http.StatusNoContent {
			body, _ := io.ReadAll(chunkResp.Body)
			return "", fmt.Errorf("chunk upload failed with status %d: %s", chunkResp.StatusCode, string(body))
		}

		offset += int64(n)

		// Send progress update
		if progressCh != nil {
			select {
			case progressCh <- UploadProgress{BytesSent: offset, BytesTotal: fileSize}:
			default:
			}
		}

		if err == io.EOF {
			break
		}
	}

	return videoID, nil
}
