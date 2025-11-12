package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"cfstream/internal/config"

	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Get video links",
	Long:  `Get various types of links for videos (preview, signed, thumbnails, HLS, DASH).`,
}

var linkPreviewCmd = &cobra.Command{
	Use:   "preview <video-id>",
	Short: "Get preview URL",
	Long:  `Get the preview/HLS manifest URL for a video.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runLinkPreview,
}

var linkSignedCmd = &cobra.Command{
	Use:   "signed <video-id>",
	Short: "Get signed URL",
	Long:  `Generate a signed (short-lived) URL for a video.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runLinkSigned,
}

var linkThumbnailCmd = &cobra.Command{
	Use:   "thumbnail <video-id>",
	Short: "Get thumbnail URL",
	Long:  `Get thumbnail URL for a video.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runLinkThumbnail,
}

var linkHLSCmd = &cobra.Command{
	Use:   "hls <video-id>",
	Short: "Get HLS manifest URL",
	Long:  `Get HLS manifest URL for a video (same as preview).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runLinkPreview,
}

var linkDASHCmd = &cobra.Command{
	Use:   "dash <video-id>",
	Short: "Get DASH manifest URL",
	Long:  `Get DASH manifest URL for a video.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runLinkDASH,
}

var (
	signedDuration string
	thumbnailTime  string
)

func init() {
	rootCmd.AddCommand(linkCmd)
	linkCmd.AddCommand(linkPreviewCmd)
	linkCmd.AddCommand(linkSignedCmd)
	linkCmd.AddCommand(linkThumbnailCmd)
	linkCmd.AddCommand(linkHLSCmd)
	linkCmd.AddCommand(linkDASHCmd)

	// Signed command flags
	linkSignedCmd.Flags().StringVar(&signedDuration, "duration", "", "token duration (e.g., 1h, 30m, 2h30m)")

	// Thumbnail command flags
	linkThumbnailCmd.Flags().StringVar(&thumbnailTime, "time", "", "timestamp for thumbnail (e.g., 10s, 1m30s)")
}

func runLinkPreview(cmd *cobra.Command, args []string) error {
	videoID := args[0]

	client, err := createClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	video, err := client.GetVideo(ctx, videoID)
	if err != nil {
		return fmt.Errorf("failed to get video: %w", err)
	}

	// Check if video requires signed URLs
	if video.RequireSignedURLs {
		return fmt.Errorf("this video is private and requires a signed URL\n\nUse: cfstream link signed %s --duration 24h", videoID)
	}

	if outputFormat == "json" {
		result := map[string]string{
			"url": video.Preview,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	fmt.Println(video.Preview)
	return nil
}

func runLinkSigned(cmd *cobra.Command, args []string) error {
	videoID := args[0]

	// Parse duration
	var durationSeconds int64
	if signedDuration != "" {
		duration, err := time.ParseDuration(signedDuration)
		if err != nil {
			return fmt.Errorf("invalid duration format: %w", err)
		}
		durationSeconds = time.Now().Unix() + int64(duration.Seconds())
	} else {
		// Use default duration from config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		duration, err := time.ParseDuration(cfg.DefaultSignedDuration)
		if err != nil {
			return fmt.Errorf("invalid default duration in config: %w", err)
		}
		durationSeconds = time.Now().Unix() + int64(duration.Seconds())
	}

	client, err := createClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get video to extract customer code
	video, err := client.GetVideo(ctx, videoID)
	if err != nil {
		return fmt.Errorf("failed to get video: %w", err)
	}

	// Generate signed token
	token, err := client.GetSignedToken(ctx, videoID, durationSeconds)
	if err != nil {
		return fmt.Errorf("failed to generate signed token: %w", err)
	}

	// Extract customer code from preview URL
	customerCode, err := extractCustomerCodeFromURL(video.Preview)
	if err != nil {
		return fmt.Errorf("failed to extract customer code: %w", err)
	}

	// Construct signed URL
	signedURL := fmt.Sprintf("https://customer-%s.cloudflarestream.com/%s/watch?token=%s", customerCode, videoID, token)

	if outputFormat == "json" {
		result := map[string]string{
			"url":   signedURL,
			"token": token,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	fmt.Println(signedURL)
	return nil
}

func runLinkThumbnail(cmd *cobra.Command, args []string) error {
	videoID := args[0]

	client, err := createClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	video, err := client.GetVideo(ctx, videoID)
	if err != nil {
		return fmt.Errorf("failed to get video: %w", err)
	}

	thumbnailURL := video.Thumbnail

	// Add time parameter if specified
	if thumbnailTime != "" {
		// Parse time duration
		duration, err := time.ParseDuration(thumbnailTime)
		if err != nil {
			return fmt.Errorf("invalid time format: %w", err)
		}

		// Convert to seconds
		seconds := duration.Seconds()

		// Extract customer code from preview URL
		customerCode, err := extractCustomerCodeFromURL(video.Preview)
		if err != nil {
			return fmt.Errorf("failed to extract customer code: %w", err)
		}

		// Construct thumbnail URL with time parameter
		thumbnailURL = fmt.Sprintf("https://customer-%s.cloudflarestream.com/%s/thumbnails/thumbnail.jpg?time=%.0fs", customerCode, videoID, seconds)
	}

	if outputFormat == "json" {
		result := map[string]string{
			"url": thumbnailURL,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	fmt.Println(thumbnailURL)
	return nil
}

func runLinkDASH(cmd *cobra.Command, args []string) error {
	videoID := args[0]

	client, err := createClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	video, err := client.GetVideo(ctx, videoID)
	if err != nil {
		return fmt.Errorf("failed to get video: %w", err)
	}

	// Check if video requires signed URLs
	if video.RequireSignedURLs {
		return fmt.Errorf("this video is private and requires a signed URL\n\nUse: cfstream link signed %s --duration 24h", videoID)
	}

	// Extract customer code from preview URL
	customerCode, err := extractCustomerCodeFromURL(video.Preview)
	if err != nil {
		return fmt.Errorf("failed to extract customer code: %w", err)
	}

	// Construct DASH URL
	dashURL := fmt.Sprintf("https://customer-%s.cloudflarestream.com/%s/manifest/video.mpd", customerCode, videoID)

	if outputFormat == "json" {
		result := map[string]string{
			"url": dashURL,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	fmt.Println(dashURL)
	return nil
}

// extractCustomerCodeFromURL extracts the customer code from a Cloudflare Stream URL.
func extractCustomerCodeFromURL(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("URL is empty")
	}

	// URL format: https://customer-{code}.cloudflarestream.com/...
	parts := strings.Split(url, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid URL format")
	}

	// Extract customer code from subdomain
	subdomain := parts[0]
	prefix := "https://customer-"
	if !strings.HasPrefix(subdomain, prefix) {
		return "", fmt.Errorf("invalid URL format: missing customer prefix")
	}

	code := strings.TrimPrefix(subdomain, prefix)
	if code == "" {
		return "", fmt.Errorf("customer code is empty")
	}

	return code, nil
}
