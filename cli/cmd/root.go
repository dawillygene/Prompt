package cmd

import (
	"fmt"

	"github.com/dawillygene/my-prompt-repository/internal/api"
	"github.com/dawillygene/my-prompt-repository/internal/config"
	"github.com/dawillygene/my-prompt-repository/internal/shell"
	"github.com/spf13/cobra"
)

var (
	jsonMode   bool
	cfgFile    string
	cfg        config.Config
	client     *api.Client
	showBanner bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "prompt",
	Short: "AI-powered prompt management CLI",
	Long: `Prompt - A GitHub Copilot-style prompt management CLI

A terminal-first system for managing, searching, versioning, and syncing 
your prompts. Features include:

  • Interactive shell with slash commands
  • Real-time autocomplete suggestions
  • Prompt CRUD operations with categories and tags
  • Full-text search and filtering
  • Favorites and archiving
  • Mode switching (Interactive ↔ Plan)

Launch 'prompt' for the interactive shell, or use individual 
commands for quick operations.`,
	Version: "1.0.0",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for commands that don't need it
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}

		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client = api.New(cfg)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no subcommand, launch interactive shell
		return shell.Start(cfg, client, showBanner)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&jsonMode, "json", false, "output machine-readable JSON")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.prompt.json)")
	rootCmd.Flags().BoolVar(&showBanner, "banner", true, "show animated banner on startup")

	// Hide completion command from main help (still accessible)
	rootCmd.CompletionOptions.HiddenDefaultCmd = false
}

// Helper functions for subcommands
func getClient() *api.Client {
	return client
}

func getConfig() config.Config {
	return cfg
}

func isJSONMode() bool {
	return jsonMode
}

func setConfig(newCfg config.Config) {
	cfg = newCfg
}
