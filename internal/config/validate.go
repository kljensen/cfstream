package config

import (
	"fmt"
	"strings"
	"time"
)

// Validate checks if the configuration has all required fields and valid values.
func Validate(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Check required fields
	if strings.TrimSpace(cfg.AccountID) == "" {
		return fmt.Errorf("account_id is required")
	}

	if strings.TrimSpace(cfg.APIToken) == "" {
		return fmt.Errorf("api_token is required")
	}

	// Validate output format
	validOutputs := map[string]bool{
		"table": true,
		"json":  true,
		"yaml":  true,
	}

	output := strings.ToLower(strings.TrimSpace(cfg.DefaultOutput))
	if output == "" {
		output = "table" // Default value
		cfg.DefaultOutput = output
	}

	if !validOutputs[output] {
		return fmt.Errorf("default_output must be one of: table, json, yaml (got: %s)", cfg.DefaultOutput)
	}

	// Validate signed duration
	duration := strings.TrimSpace(cfg.DefaultSignedDuration)
	if duration == "" {
		duration = "1h" // Default value
		cfg.DefaultSignedDuration = duration
	}

	if _, err := time.ParseDuration(cfg.DefaultSignedDuration); err != nil {
		return fmt.Errorf("default_signed_duration must be a valid duration string (e.g., 1h, 30m, 1h30m): %w", err)
	}

	return nil
}
