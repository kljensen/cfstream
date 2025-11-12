// Package config manages configuration loading and persistence for cfstream CLI.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

// Config holds the configuration for cfstream CLI.
type Config struct {
	AccountID             string `mapstructure:"account_id"`
	APIToken              string `mapstructure:"api_token"`
	DefaultOutput         string `mapstructure:"default_output"`
	DefaultSignedDuration string `mapstructure:"default_signed_duration"`
}

// Load reads configuration from file and environment variables.
// Environment variables take precedence over config file values.
// Returns a Config with default values if no configuration exists.
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("default_output", "table")
	v.SetDefault("default_signed_duration", "1h")

	// Configure file location
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(filepath.Join(xdg.ConfigHome, "cfstream"))

	// Read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		// Ignore file not found errors and permission errors, we'll use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Also check for generic file system errors (file doesn't exist, permission denied)
			if !os.IsNotExist(err) && !os.IsPermission(err) {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		}
	}

	// Environment variables override config file
	v.BindEnv("account_id", "CFSTREAM_ACCOUNT_ID")
	v.BindEnv("api_token", "CFSTREAM_API_TOKEN")
	v.BindEnv("default_output", "CFSTREAM_OUTPUT")

	// Create config struct
	cfg := &Config{
		AccountID:             v.GetString("account_id"),
		APIToken:              v.GetString("api_token"),
		DefaultOutput:         v.GetString("default_output"),
		DefaultSignedDuration: v.GetString("default_signed_duration"),
	}

	return cfg, nil
}

// Save writes the configuration to the config file.
func Save(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Ensure config directory exists
	configPath := Path()
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create viper instance and set values
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	v.Set("account_id", cfg.AccountID)
	v.Set("api_token", cfg.APIToken)
	v.Set("default_output", cfg.DefaultOutput)
	v.Set("default_signed_duration", cfg.DefaultSignedDuration)

	// Write config file
	if err := v.WriteConfig(); err != nil {
		// If file doesn't exist, create it
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := v.SafeWriteConfig(); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Path returns the full path to the config file.
func Path() string {
	return filepath.Join(xdg.ConfigHome, "cfstream", "config.yaml")
}
