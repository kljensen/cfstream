package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"cfstream/internal/api"
	"cfstream/internal/config"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage cfstream configuration",
	Long:  `Initialize and display cfstream configuration settings.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize cfstream configuration",
	Long:  `Interactive setup for Cloudflare Stream API credentials and preferences.`,
	RunE:  runConfigInit,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long:  `Display current configuration values from file and environment variables.`,
	RunE:  runConfigShow,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	fmt.Println("Cloudflare Stream Configuration Setup")
	fmt.Println()

	cfg := &config.Config{}
	reader := bufio.NewReader(os.Stdin)

	// Prompt for Account ID
	fmt.Print("Enter Account ID: ")
	accountID, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read account ID: %w", err)
	}
	cfg.AccountID = strings.TrimSpace(accountID)

	// Prompt for API Token (masked)
	fmt.Print("Enter API Token: ")
	tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Print newline after masked input
	if err != nil {
		return fmt.Errorf("failed to read API token: %w", err)
	}
	cfg.APIToken = strings.TrimSpace(string(tokenBytes))

	// Prompt for default output format
	fmt.Print("Default output format (table/json/yaml) [table]: ")
	output, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read output format: %w", err)
	}
	output = strings.TrimSpace(output)
	if output == "" {
		output = "table"
	}
	cfg.DefaultOutput = output

	// Prompt for default signed URL duration
	fmt.Print("Default signed URL duration [1h]: ")
	duration, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read duration: %w", err)
	}
	duration = strings.TrimSpace(duration)
	if duration == "" {
		duration = "1h"
	}
	cfg.DefaultSignedDuration = duration

	fmt.Println()

	// Validate configuration
	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Test credentials by attempting to create client and list videos
	fmt.Println("Validating credentials...")
	client, err := api.NewClient(cfg.AccountID, cfg.APIToken)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Create context with timeout for validation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test API call
	_, err = client.ListVideos(ctx, nil)
	if err != nil {
		return fmt.Errorf("credential validation failed: %w", err)
	}

	fmt.Println("✓ Credentials validated successfully")
	fmt.Println()

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Configuration saved to %s\n", config.Path())
	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check which values come from environment
	envAccountID := os.Getenv("CFSTREAM_ACCOUNT_ID")
	envAPIToken := os.Getenv("CFSTREAM_API_TOKEN")
	envOutput := os.Getenv("CFSTREAM_OUTPUT")

	fmt.Println("Configuration:")

	// Display Account ID
	accountIDSource := ""
	if envAccountID != "" {
		accountIDSource = " (from env)"
	}
	fmt.Printf("  Account ID: %s%s\n", cfg.AccountID, accountIDSource)

	// Display masked API Token
	tokenSource := ""
	if envAPIToken != "" {
		tokenSource = " (from env)"
	}
	maskedToken := maskToken(cfg.APIToken)
	fmt.Printf("  API Token:  %s%s\n", maskedToken, tokenSource)

	// Display output format
	outputSource := ""
	if envOutput != "" {
		outputSource = " (from env)"
	}
	fmt.Printf("  Output:     %s%s\n", cfg.DefaultOutput, outputSource)

	// Display duration
	fmt.Printf("  Duration:   %s\n", cfg.DefaultSignedDuration)

	fmt.Println()
	fmt.Printf("Config file: %s\n", config.Path())

	return nil
}

// maskToken returns a masked version of the API token showing first 8 chars
func maskToken(token string) string {
	if token == "" {
		return "<not set>"
	}

	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}

	return token[:8] + strings.Repeat("*", len(token)-8)
}
