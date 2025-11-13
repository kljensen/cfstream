package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"cfstream/internal/api"
	"cfstream/internal/config"
	"cfstream/internal/output"
	"cfstream/internal/upload"
)

var (
	uploadName     string
	uploadMetadata string
	uploadExpires  string
	maxDuration    int
)

// uploadCmd represents the upload command.
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload videos to Cloudflare Stream",
	Long: `Upload videos to Cloudflare Stream using various methods:

- upload file <path> - Upload a local video file
- upload url <url>  - Upload from a URL
- upload direct     - Generate a direct upload URL`,
}

// uploadFileCmd uploads a local video file.
var uploadFileCmd = &cobra.Command{
	Use:   "file <path>",
	Short: "Upload a local video file",
	Long: `Upload a local video file to Cloudflare Stream using multipart/form-data.

This command uploads a video file with support for progress tracking.
The upload uses standard multipart/form-data encoding.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Validate file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filePath)
		}

		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Create API client
		client, err := api.NewClient(cfg.AccountID, cfg.APIToken)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		// Parse metadata if provided
		var metadata map[string]interface{}
		if uploadMetadata != "" {
			if err := json.Unmarshal([]byte(uploadMetadata), &metadata); err != nil {
				return fmt.Errorf("invalid metadata JSON: %w", err)
			}
		}

		// Prepare upload options
		opts := &api.UploadOptions{
			Name:              uploadName,
			Metadata:          metadata,
			RequireSignedURLs: true,
		}

		// If name not provided, use filename
		if opts.Name == "" {
			opts.Name = filepath.Base(filePath)
		}

		// Get file size for progress tracking
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("failed to get file info: %w", err)
		}

		if !quiet {
			fmt.Printf("Uploading %s (%s)...\n", filepath.Base(filePath), upload.FormatBytes(fileInfo.Size()))
		}

		// Create progress tracker
		progressTracker := upload.NewProgressTracker(fileInfo.Size(), filepath.Base(filePath), quiet)

		// Create progress channel
		progressCh := make(chan api.UploadProgress, 10)
		go func() {
			for progress := range progressCh {
				progressTracker.Update(progress)
			}
		}()

		// Upload file
		ctx := context.Background()
		video, err := client.UploadFile(ctx, filePath, opts, progressCh)
		close(progressCh)
		progressTracker.Finish()

		if err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}

		if !quiet {
			fmt.Println("Upload complete")
			fmt.Printf("Video ID: %s\n", video.UID)
			fmt.Printf("Status: %s\n", video.Status)
			if video.Preview != "" {
				fmt.Printf("Preview: %s\n", video.Preview)
			}
		}

		// Poll for processing status if not quiet
		if !quiet && !video.ReadyToStream {
			fmt.Println("\nProcessing video...")
			if err := pollVideoStatus(ctx, client, video.UID); err != nil {
				fmt.Printf("Warning: failed to check video status: %v\n", err)
			}
		}

		// Output video details in requested format
		if outputFormat != outputFormatTable {
			formatter, err := output.NewFormatter(outputFormat)
			if err != nil {
				return err
			}
			return formatter.FormatSingle(os.Stdout, video)
		}

		return nil
	},
}

// uploadURLCmd uploads a video from a URL.
var uploadURLCmd = &cobra.Command{
	Use:   "url <url>",
	Short: "Upload a video from a URL",
	Long: `Upload a video from a URL to Cloudflare Stream.

Cloudflare will download the video from the provided URL and process it.
Processing happens asynchronously, so the command returns immediately with
a video ID.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		videoURL := args[0]

		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Create API client
		client, err := api.NewClient(cfg.AccountID, cfg.APIToken)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		// Parse metadata if provided
		var metadata map[string]interface{}
		if uploadMetadata != "" {
			if err := json.Unmarshal([]byte(uploadMetadata), &metadata); err != nil {
				return fmt.Errorf("invalid metadata JSON: %w", err)
			}
		}

		// Prepare upload options
		opts := &api.UploadOptions{
			Name:              uploadName,
			Metadata:          metadata,
			RequireSignedURLs: true,
		}

		if !quiet {
			fmt.Printf("Uploading from URL: %s\n", videoURL)
		}

		// Upload from URL
		ctx := context.Background()
		video, err := client.UploadFromURL(ctx, videoURL, opts)
		if err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}

		if !quiet {
			fmt.Println("Upload initiated")
			fmt.Printf("Video ID: %s\n", video.UID)
			fmt.Printf("Status: %s\n", video.Status)
			if video.Preview != "" {
				fmt.Printf("Preview: %s\n", video.Preview)
			}
			fmt.Println("\nNote: Video processing happens asynchronously. Use 'cfstream video get' to check status.")
		}

		// Output video details in requested format
		if outputFormat != outputFormatTable {
			formatter, err := output.NewFormatter(outputFormat)
			if err != nil {
				return err
			}
			return formatter.FormatSingle(os.Stdout, video)
		}

		return nil
	},
}

// uploadDirectCmd generates a direct upload URL.
var uploadDirectCmd = &cobra.Command{
	Use:   "direct",
	Short: "Generate a direct upload URL",
	Long: `Generate a direct upload URL for end users to upload videos directly.

This is useful when you want to allow users to upload videos directly to
Cloudflare Stream without going through your server. The URL is time-limited
and can be configured with upload constraints.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Create API client
		client, err := api.NewClient(cfg.AccountID, cfg.APIToken)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		// Parse expiry if provided
		var expiry *time.Time
		if uploadExpires != "" {
			duration, err := time.ParseDuration(uploadExpires)
			if err != nil {
				return fmt.Errorf("invalid expiry duration: %w", err)
			}
			expiryTime := time.Now().Add(duration)
			expiry = &expiryTime
		}

		// Prepare options
		opts := &api.DirectUploadOptions{
			MaxDurationSeconds: maxDuration,
			Expiry:             expiry,
			RequireSignedURLs:  true,
		}

		// Create direct upload URL
		ctx := context.Background()
		result, err := client.CreateDirectUploadURL(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to create direct upload URL: %w", err)
		}

		if !quiet {
			fmt.Println("Direct upload URL created")
			fmt.Printf("Video ID: %s\n", result.UID)
			fmt.Printf("Upload URL: %s\n", result.UploadURL)
			if !result.Expiry.IsZero() {
				fmt.Printf("Expires: %s\n", result.Expiry.Format(time.RFC3339))
			}
		}

		// Output result in requested format
		if outputFormat != outputFormatTable {
			formatter, err := output.NewFormatter(outputFormat)
			if err != nil {
				return err
			}
			return formatter.FormatSingle(os.Stdout, result)
		}

		return nil
	},
}

// pollVideoStatus polls the video status until it's ready to stream.
func pollVideoStatus(ctx context.Context, client api.Client, videoID string) error {
	const maxAttempts = 60
	const pollInterval = 5 * time.Second

	for i := 0; i < maxAttempts; i++ {
		time.Sleep(pollInterval)

		video, err := client.GetVideo(ctx, videoID)
		if err != nil {
			return err
		}

		if video.ReadyToStream {
			fmt.Println("Video ready for streaming")
			return nil
		}

		if video.Status == "error" {
			return fmt.Errorf("video processing failed: %s", video.StatusDetails)
		}

		if !quiet {
			fmt.Printf("Status: %s", video.Status)
			if video.StatusDetails != "" {
				fmt.Printf(" (%s)", video.StatusDetails)
			}
			fmt.Println()
		}
	}

	fmt.Println("Video is still processing. Use 'cfstream video get' to check status.")
	return nil
}

func init() {
	// Add subcommands
	uploadCmd.AddCommand(uploadFileCmd)
	uploadCmd.AddCommand(uploadURLCmd)
	uploadCmd.AddCommand(uploadDirectCmd)

	// Flags for file and url uploads
	uploadFileCmd.Flags().StringVar(&uploadName, "name", "", "video name (defaults to filename)")
	uploadFileCmd.Flags().StringVar(&uploadMetadata, "metadata", "", "video metadata as JSON")

	uploadURLCmd.Flags().StringVar(&uploadName, "name", "", "video name")
	uploadURLCmd.Flags().StringVar(&uploadMetadata, "metadata", "", "video metadata as JSON")

	// Flags for direct upload
	uploadDirectCmd.Flags().StringVar(&uploadExpires, "expires", "1h", "expiration duration (e.g., 1h, 30m)")
	uploadDirectCmd.Flags().IntVar(&maxDuration, "max-duration", 0, "maximum video duration in seconds")
}
