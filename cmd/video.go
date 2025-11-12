package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"cfstream/internal/api"
	"cfstream/internal/config"
	"cfstream/internal/output"

	"github.com/spf13/cobra"
)

var videoCmd = &cobra.Command{
	Use:   "video",
	Short: "Manage Cloudflare Stream videos",
	Long:  `List, get, delete, and update Cloudflare Stream videos.`,
}

var videoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List videos",
	Long:  `List videos from Cloudflare Stream with optional filtering.`,
	RunE:  runVideoList,
}

var videoGetCmd = &cobra.Command{
	Use:   "get <video-id>",
	Short: "Get video details",
	Long:  `Get details for a specific video by ID.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runVideoGet,
}

var videoDeleteCmd = &cobra.Command{
	Use:   "delete <video-id>",
	Short: "Delete a video",
	Long:  `Delete a video from Cloudflare Stream.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runVideoDelete,
}

var videoUpdateCmd = &cobra.Command{
	Use:   "update <video-id>",
	Short: "Update video metadata",
	Long:  `Update metadata for a specific video.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runVideoUpdate,
}

var (
	// List flags
	listSearch string
	listLimit  int
	listAfter  string
	listStatus string

	// Delete flags
	deleteYes bool

	// Update flags
	updateName              string
	updateMetadata          string
	updateRequireSignedURLs string
)

func init() {
	rootCmd.AddCommand(videoCmd)
	videoCmd.AddCommand(videoListCmd)
	videoCmd.AddCommand(videoGetCmd)
	videoCmd.AddCommand(videoDeleteCmd)
	videoCmd.AddCommand(videoUpdateCmd)

	// List command flags
	videoListCmd.Flags().StringVar(&listSearch, "search", "", "search by video name")
	videoListCmd.Flags().IntVar(&listLimit, "limit", 50, "number of videos to return")
	videoListCmd.Flags().StringVar(&listAfter, "after", "", "cursor for pagination")
	videoListCmd.Flags().StringVar(&listStatus, "status", "", "filter by status (ready, processing, error)")

	// Delete command flags
	videoDeleteCmd.Flags().BoolVarP(&deleteYes, "yes", "y", false, "skip confirmation")

	// Update command flags
	videoUpdateCmd.Flags().StringVar(&updateName, "name", "", "new name for the video")
	videoUpdateCmd.Flags().StringVar(&updateMetadata, "metadata", "", "JSON string of metadata key-value pairs")
	videoUpdateCmd.Flags().StringVar(&updateRequireSignedURLs, "require-signed", "", "require signed URLs (true/false)")
}

func runVideoList(cmd *cobra.Command, args []string) error {
	client, err := createClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := &api.ListOptions{
		Search: listSearch,
		Status: listStatus,
	}

	videos, err := client.ListVideos(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list videos: %w", err)
	}

	if len(videos) == 0 {
		if !quiet {
			fmt.Println("No videos found")
		}
		return nil
	}

	// Create formatter
	formatter, err := output.NewFormatter(outputFormat)
	if err != nil {
		return err
	}

	// Format and display videos
	headers := []string{"UID", "Name", "Status", "Duration", "Created"}
	if err := formatter.FormatList(os.Stdout, headers, videos); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	return nil
}

func runVideoGet(cmd *cobra.Command, args []string) error {
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

	// Create formatter
	formatter, err := output.NewFormatter(outputFormat)
	if err != nil {
		return err
	}

	// Format and display video
	if err := formatter.FormatSingle(os.Stdout, video); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	return nil
}

func runVideoDelete(cmd *cobra.Command, args []string) error {
	videoID := args[0]

	// Confirm deletion unless --yes flag is provided
	if !deleteYes {
		fmt.Printf("Are you sure you want to delete video %s? (y/N): ", videoID)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	client, err := createClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.DeleteVideo(ctx, videoID); err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}

	if !quiet {
		fmt.Printf("Video %s deleted successfully\n", videoID)
	}

	return nil
}

func runVideoUpdate(cmd *cobra.Command, args []string) error {
	videoID := args[0]

	// Validate that at least one update option is provided
	if updateName == "" && updateMetadata == "" && updateRequireSignedURLs == "" {
		return fmt.Errorf("at least one of --name, --metadata, or --require-signed must be provided")
	}

	// Build update options
	opts := &api.UpdateOptions{
		Meta: make(map[string]interface{}),
	}

	// Handle name flag
	if updateName != "" {
		opts.Meta["name"] = updateName
	}

	// Handle metadata flag
	if updateMetadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(updateMetadata), &metadata); err != nil {
			return fmt.Errorf("invalid metadata JSON: %w", err)
		}
		// Merge metadata into opts.Meta
		for k, v := range metadata {
			opts.Meta[k] = v
		}
	}

	// Handle requireSignedURLs flag
	if updateRequireSignedURLs != "" {
		switch strings.ToLower(updateRequireSignedURLs) {
		case "true", "yes", "1":
			opts.RequireSignedURLs = &[]bool{true}[0]
		case "false", "no", "0":
			opts.RequireSignedURLs = &[]bool{false}[0]
		default:
			return fmt.Errorf("invalid value for --require-signed: %s (use true or false)", updateRequireSignedURLs)
		}
	}

	// Clean up empty Meta if only requireSignedURLs was set
	if len(opts.Meta) == 0 {
		opts.Meta = nil
	}

	client, err := createClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	video, err := client.UpdateVideo(ctx, videoID, opts)
	if err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}

	if !quiet {
		fmt.Println("Video updated successfully")
	}

	// Create formatter
	formatter, err := output.NewFormatter(outputFormat)
	if err != nil {
		return err
	}

	// Format and display updated video
	if err := formatter.FormatSingle(os.Stdout, video); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	return nil
}

// createClient creates an API client from configuration
func createClient() (api.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	if cfg.AccountID == "" {
		return nil, fmt.Errorf("account ID not configured (run 'cfstream config init')")
	}
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("API token not configured (run 'cfstream config init')")
	}

	client, err := api.NewClient(cfg.AccountID, cfg.APIToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	return client, nil
}
