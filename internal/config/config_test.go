package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear environment variables
	clearEnv(t)

	// Use temporary XDG_CONFIG_HOME to isolate test
	tempDir := t.TempDir()
	oldXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if oldXDGConfig != "" {
			os.Setenv("XDG_CONFIG_HOME", oldXDGConfig)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
		xdg.Reload()
	}()
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	xdg.Reload()

	// Load config without file or env vars
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check defaults are set
	assert.Equal(t, "", cfg.AccountID)
	assert.Equal(t, "", cfg.APIToken)
	assert.Equal(t, "table", cfg.DefaultOutput)
	assert.Equal(t, "1h", cfg.DefaultSignedDuration)
}

func TestLoad_FromEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected Config
	}{
		{
			name: "all environment variables",
			envVars: map[string]string{
				"CFSTREAM_ACCOUNT_ID": "test-account-123",
				"CFSTREAM_API_TOKEN":  "test-token-xyz",
				"CFSTREAM_OUTPUT":     "json",
			},
			expected: Config{
				AccountID:             "test-account-123",
				APIToken:              "test-token-xyz",
				DefaultOutput:         "json",
				DefaultSignedDuration: "1h",
			},
		},
		{
			name: "partial environment variables",
			envVars: map[string]string{
				"CFSTREAM_ACCOUNT_ID": "partial-account",
			},
			expected: Config{
				AccountID:             "partial-account",
				APIToken:              "",
				DefaultOutput:         "table",
				DefaultSignedDuration: "1h",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			setEnv(t, tt.envVars)

			// Use temporary XDG_CONFIG_HOME to isolate test
			tempDir := t.TempDir()
			oldXDGConfig := os.Getenv("XDG_CONFIG_HOME")
			defer func() {
				if oldXDGConfig != "" {
					os.Setenv("XDG_CONFIG_HOME", oldXDGConfig)
				} else {
					os.Unsetenv("XDG_CONFIG_HOME")
				}
				xdg.Reload()
			}()
			os.Setenv("XDG_CONFIG_HOME", tempDir)
			xdg.Reload()

			cfg, err := Load()
			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.Equal(t, tt.expected.AccountID, cfg.AccountID)
			assert.Equal(t, tt.expected.APIToken, cfg.APIToken)
			assert.Equal(t, tt.expected.DefaultOutput, cfg.DefaultOutput)
			assert.Equal(t, tt.expected.DefaultSignedDuration, cfg.DefaultSignedDuration)
		})
	}
}

func TestSave_Success(t *testing.T) {
	// Clear environment variables first
	clearEnv(t)

	// Create temporary directory for config file
	tempDir := t.TempDir()
	tempConfigPath := filepath.Join(tempDir, "cfstream", "config.yaml")

	// Override XDG_CONFIG_HOME for this test
	oldXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		// Restore after test
		if oldXDGConfig != "" {
			os.Setenv("XDG_CONFIG_HOME", oldXDGConfig)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
		xdg.Reload()
	}()
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	xdg.Reload()

	cfg := &Config{
		AccountID:             "save-account-123",
		APIToken:              "save-token-xyz",
		DefaultOutput:         "yaml",
		DefaultSignedDuration: "2h",
	}

	err := Save(cfg)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(tempConfigPath)
	require.NoError(t, err)

	// Load config and verify values
	loadedCfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, cfg.AccountID, loadedCfg.AccountID)
	assert.Equal(t, cfg.APIToken, loadedCfg.APIToken)
	assert.Equal(t, cfg.DefaultOutput, loadedCfg.DefaultOutput)
	assert.Equal(t, cfg.DefaultSignedDuration, loadedCfg.DefaultSignedDuration)
}

func TestSave_NilConfig(t *testing.T) {
	err := Save(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")
}

func TestPath(t *testing.T) {
	path := Path()
	assert.NotEmpty(t, path)
	assert.True(t, filepath.IsAbs(path))
	assert.Contains(t, path, filepath.Join("cfstream", "config.yaml"))
}

func TestValidate_Success(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "valid config with all fields",
			config: Config{
				AccountID:             "valid-account",
				APIToken:              "valid-token",
				DefaultOutput:         "table",
				DefaultSignedDuration: "1h",
			},
		},
		{
			name: "valid config with json output",
			config: Config{
				AccountID:             "valid-account",
				APIToken:              "valid-token",
				DefaultOutput:         "json",
				DefaultSignedDuration: "30m",
			},
		},
		{
			name: "valid config with yaml output",
			config: Config{
				AccountID:             "valid-account",
				APIToken:              "valid-token",
				DefaultOutput:         "yaml",
				DefaultSignedDuration: "1h30m",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(&tt.config)
			assert.NoError(t, err)
		})
	}
}

func TestValidate_Errors(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError string
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: "config cannot be nil",
		},
		{
			name: "missing account_id",
			config: &Config{
				APIToken:              "token",
				DefaultOutput:         "table",
				DefaultSignedDuration: "1h",
			},
			expectError: "account_id is required",
		},
		{
			name: "empty account_id",
			config: &Config{
				AccountID:             "   ",
				APIToken:              "token",
				DefaultOutput:         "table",
				DefaultSignedDuration: "1h",
			},
			expectError: "account_id is required",
		},
		{
			name: "missing api_token",
			config: &Config{
				AccountID:             "account",
				DefaultOutput:         "table",
				DefaultSignedDuration: "1h",
			},
			expectError: "api_token is required",
		},
		{
			name: "empty api_token",
			config: &Config{
				AccountID:             "account",
				APIToken:              "   ",
				DefaultOutput:         "table",
				DefaultSignedDuration: "1h",
			},
			expectError: "api_token is required",
		},
		{
			name: "invalid output format",
			config: &Config{
				AccountID:             "account",
				APIToken:              "token",
				DefaultOutput:         "xml",
				DefaultSignedDuration: "1h",
			},
			expectError: "default_output must be one of: table, json, yaml",
		},
		{
			name: "invalid signed duration",
			config: &Config{
				AccountID:             "account",
				APIToken:              "token",
				DefaultOutput:         "table",
				DefaultSignedDuration: "invalid",
			},
			expectError: "default_signed_duration must be a valid duration string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.config)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

func TestValidate_DefaultValues(t *testing.T) {
	cfg := &Config{
		AccountID: "account",
		APIToken:  "token",
		// DefaultOutput and DefaultSignedDuration are empty
	}

	err := Validate(cfg)
	require.NoError(t, err)

	// Validate should set defaults
	assert.Equal(t, "table", cfg.DefaultOutput)
	assert.Equal(t, "1h", cfg.DefaultSignedDuration)
}

// Helper function to clear environment variables
func clearEnv(t *testing.T) {
	t.Helper()
	envVars := []string{
		"CFSTREAM_ACCOUNT_ID",
		"CFSTREAM_API_TOKEN",
		"CFSTREAM_OUTPUT",
	}
	for _, key := range envVars {
		os.Unsetenv(key)
	}
}

// Helper function to set environment variables
func setEnv(t *testing.T, envVars map[string]string) {
	t.Helper()
	for key, value := range envVars {
		os.Setenv(key, value)
	}
}
