// Package api provides a simplified wrapper around the Cloudflare Stream API.
package api

import (
	"time"

	"github.com/cloudflare/cloudflare-go/v3/stream"
)

// Video represents a Cloudflare Stream video with simplified fields for CLI usage.
type Video struct {
	UID               string
	Name              string
	Status            string
	StatusDetails     string
	Duration          float64
	Created           time.Time
	Modified          time.Time
	ReadyToStream     bool
	RequireSignedURLs bool
	Preview           string
	Thumbnail         string
	Creator           string
	Meta              map[string]interface{}
}

// ListOptions contains parameters for listing videos.
type ListOptions struct {
	Search  string
	Creator string
	Start   *time.Time
	End     *time.Time
	Status  string
	Asc     bool
}

// UpdateOptions contains parameters for updating a video.
type UpdateOptions struct {
	Meta              map[string]interface{}
	RequireSignedURLs *bool // Pointer to allow nil (optional)
}

// EmbedOptions contains parameters for customizing embed code.
type EmbedOptions struct {
	Responsive  bool
	Autoplay    bool
	Muted       bool
	Loop        bool
	Controls    bool
	SignedToken string
}

// UploadOptions contains parameters for uploading a video.
type UploadOptions struct {
	Name              string
	Metadata          map[string]interface{}
	RequireSignedURLs bool
}

// DirectUploadOptions contains parameters for creating a direct upload URL.
type DirectUploadOptions struct {
	MaxDurationSeconds int
	Expiry             *time.Time
	RequireSignedURLs  bool
}

// DirectUploadResult contains the response from creating a direct upload URL.
type DirectUploadResult struct {
	UploadURL string
	UID       string
	Expiry    time.Time
}

// UploadProgress represents the current state of an upload.
type UploadProgress struct {
	BytesSent  int64
	BytesTotal int64
}

// VideoFromSDK converts a Cloudflare SDK Video to our simplified Video type.
func VideoFromSDK(v *stream.Video) *Video {
	if v == nil {
		return nil
	}

	video := &Video{
		UID:               v.UID,
		Duration:          v.Duration,
		Created:           v.Created,
		Modified:          v.Modified,
		ReadyToStream:     v.ReadyToStream,
		RequireSignedURLs: v.RequireSignedURLs,
		Preview:           v.Preview,
		Thumbnail:         v.Thumbnail,
		Creator:           v.Creator,
	}

	// Extract status information
	video.Status = string(v.Status.State)
	if v.Status.ErrorReasonText != "" {
		video.StatusDetails = v.Status.ErrorReasonText
	} else if v.Status.PctComplete != "" {
		video.StatusDetails = v.Status.PctComplete + "% complete"
	}

	// Extract name from meta if available
	if metaMap, ok := v.Meta.(map[string]interface{}); ok {
		if name, ok := metaMap["name"].(string); ok && name != "" {
			video.Name = name
		}
		video.Meta = metaMap
	}

	// Fallback to UID if no name
	if video.Name == "" {
		video.Name = v.UID
	}

	return video
}

// VideosFromSDK converts a slice of SDK videos to our simplified type.
func VideosFromSDK(videos []stream.Video) []Video {
	result := make([]Video, 0, len(videos))
	for _, v := range videos {
		if converted := VideoFromSDK(&v); converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}
