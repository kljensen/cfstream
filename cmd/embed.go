package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"cfstream/internal/api"
	"cfstream/internal/config"
)

var embedCmd = &cobra.Command{
	Use:   "embed",
	Short: "Get video embed code",
	Long:  `Get HTML embed code for videos.`,
}

var embedCodeCmd = &cobra.Command{
	Use:   "code <video-id>",
	Short: "Get HTML embed code",
	Long:  `Get HTML iframe embed code for a video.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEmbedCode,
}

var (
	embedResponsive bool
	embedAutoplay   bool
	embedMuted      bool
	embedLoop       bool
	embedControls   bool
	embedDuration   string
)

func init() {
	rootCmd.AddCommand(embedCmd)
	embedCmd.AddCommand(embedCodeCmd)

	// Embed code flags
	embedCodeCmd.Flags().BoolVar(&embedResponsive, "responsive", false, "make iframe responsive")
	embedCodeCmd.Flags().BoolVar(&embedAutoplay, "autoplay", false, "enable autoplay")
	embedCodeCmd.Flags().BoolVar(&embedMuted, "muted", false, "start muted")
	embedCodeCmd.Flags().BoolVar(&embedLoop, "loop", false, "loop video")
	embedCodeCmd.Flags().BoolVar(&embedControls, "controls", true, "show controls")
	embedCodeCmd.Flags().StringVar(&embedDuration, "duration", "", "signed URL duration (e.g., 1h, 24h) - required for private videos")
}

func runEmbedCode(cmd *cobra.Command, args []string) error {
	videoID := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nRun 'cfstream config init' to configure credentials", err)
	}

	client, err := api.NewClient(cfg.AccountID, cfg.APIToken)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get video to check if it requires signed URLs
	video, err := client.GetVideo(ctx, videoID)
	if err != nil {
		return fmt.Errorf("failed to get video: %w", err)
	}

	var signedToken string

	// If video requires signed URLs, generate token
	if video.RequireSignedURLs {
		// Determine duration
		duration := embedDuration
		if duration == "" {
			// Try config default
			if cfg.DefaultSignedDuration != "" {
				duration = cfg.DefaultSignedDuration
			} else {
				return fmt.Errorf("this video is private and requires a signed URL\n\nUse: cfstream embed code %s --duration 24h", videoID)
			}
		}

		// Parse duration
		d, err := time.ParseDuration(duration)
		if err != nil {
			return fmt.Errorf("invalid duration format: %w\nExample: --duration 24h", err)
		}

		// Generate signed token (calculate absolute expiration timestamp)
		expirationTime := time.Now().Unix() + int64(d.Seconds())
		token, err := client.GetSignedToken(ctx, videoID, expirationTime)
		if err != nil {
			return fmt.Errorf("failed to generate signed token: %w", err)
		}
		signedToken = token
	}

	// Build embed options
	opts := &api.EmbedOptions{
		Responsive:  embedResponsive,
		Autoplay:    embedAutoplay,
		Muted:       embedMuted,
		Loop:        embedLoop,
		Controls:    embedControls,
		SignedToken: signedToken,
	}

	// Get embed code
	embedCode, err := client.GetEmbedCode(ctx, videoID, opts)
	if err != nil {
		return fmt.Errorf("failed to get embed code: %w", err)
	}

	if outputFormat == outputFormatJSON {
		result := map[string]string{
			"html": embedCode,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	fmt.Println(embedCode)
	return nil
}
