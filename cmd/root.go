// Package cmd implements the CLI commands for cfstream.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	// Import dependencies to ensure they're in go.mod
	_ "github.com/cloudflare/cloudflare-go/v3"
	_ "github.com/olekukonko/tablewriter"
	_ "github.com/schollz/progressbar/v3"
	_ "github.com/stretchr/testify/assert"
)

const version = "0.1.0"

var (
	// Global flags
	outputFormat string
	quiet        bool
	verbose      bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cfstream",
	Short: "Cloudflare Stream management CLI",
	Long: `cfstream is a command-line interface for managing Cloudflare Stream videos.

Upload videos, manage metadata, generate links, and retrieve embed codes
for your Cloudflare Stream account.`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(uploadCmd)

	// Global flags available to all commands
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format (table, json, yaml)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper for config file support
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))

	// Version template
	rootCmd.SetVersionTemplate(fmt.Sprintf("cfstream version %s\n", version))
}
