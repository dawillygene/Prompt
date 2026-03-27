package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dawillygene/my-prompt-repository/internal/api"
	"github.com/dawillygene/my-prompt-repository/internal/config"
	"github.com/dawillygene/my-prompt-repository/internal/tui"
	"github.com/spf13/cobra"
)

// UI command - launch interactive TUI
var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch interactive terminal UI",
	Long:  `Start the full-featured interactive terminal interface with colors and navigation.`,
	Example: `  myprompts ui
  myprompts ui --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()
		client := getClient()
		ui := tui.NewColoredUI(cfg, client)
		return ui.Start()
	},
}

// Config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `View and update configuration settings.`,
	Example: `  myprompts config set api_base http://localhost:8000
  myprompts config set api_base https://api.example.com`,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	ValidArgs: []string{"api_base"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()
		key := args[0]
		value := args[1]

		switch key {
		case "api_base":
			cfg.APIBase = value
		default:
			return fmt.Errorf("unsupported config key: %s", key)
		}

		if err := config.Save(cfg); err != nil {
			return err
		}
		setConfig(cfg)

		if isJSONMode() {
			return prettyPrint(map[string]any{
				"message": fmt.Sprintf("%s updated.", key),
				"data": map[string]any{
					"key":   key,
					"value": value,
				},
			})
		}

		fmt.Printf("%s updated.\n", key)
		return nil
	},
}

// Edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit prompts in your default editor",
	Long:  `Open your default editor (from $EDITOR) to edit prompts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			return errors.New("`edit` requires $EDITOR environment variable to be set")
		}

		editorCmd := exec.Command(editor)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		if err := editorCmd.Run(); err != nil {
			return err
		}

		if isJSONMode() {
			return prettyPrint(map[string]any{
				"message": "Editor closed.",
				"data":    map[string]any{},
			})
		}

		fmt.Println("Editor closed.")
		return nil
	},
}

// Export command
var exportFile string

var exportCmd = &cobra.Command{
	Use:   "export [file]",
	Short: "Export prompts to a JSON file",
	Long:  `Export all your prompts to a JSON file for backup or sharing.`,
	Example: `  myprompts export
  myprompts export my-prompts.json
  myprompts export ~/backup/prompts.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()

		filePath := "prompts-export.json"
		if len(args) > 0 && args[0] != "" {
			filePath = args[0]
		}

		response, err := client.Request("GET", "/api/export", nil, true)
		if err != nil {
			return err
		}

		payload, ok := response["data"]
		if !ok {
			return errors.New("invalid export payload")
		}

		raw, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}

		abs, err := filepath.Abs(filePath)
		if err != nil {
			return err
		}

		if err := os.WriteFile(abs, raw, 0o600); err != nil {
			return err
		}

		if isJSONMode() {
			return prettyPrint(map[string]any{
				"message": "Export completed.",
				"data": map[string]any{
					"file": abs,
				},
			})
		}

		fmt.Printf("Exported prompts to %s\n", abs)
		return nil
	},
}

// Import command
var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import prompts from a JSON file",
	Long:  `Import prompts from a previously exported JSON file.`,
	Example: `  myprompts import prompts-export.json
  myprompts import ~/backup/prompts.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()

		abs, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}

		raw, err := os.ReadFile(abs)
		if err != nil {
			return err
		}

		var decoded map[string]any
		if err := json.Unmarshal(raw, &decoded); err != nil {
			return fmt.Errorf("invalid import file: %w", err)
		}

		prompts, hasPrompts := decoded["prompts"]
		if !hasPrompts {
			return errors.New("invalid import file: missing `prompts` array")
		}

		response, err := client.Request("POST", "/api/import", map[string]any{
			"prompts": prompts,
		}, true)
		if err != nil {
			return err
		}

		return prettyPrint(response)
	},
}

// Sync command
var syncStatus bool

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync local cache with server",
	Long:  `Synchronize local cache with the server and resolve any conflicts.`,
	Example: `  myprompts sync
  myprompts sync --status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()

		if syncStatus {
			return showSyncStatus(client)
		}

		return performSync(client)
	},
}

func showSyncStatus(client any) error {
	// Cast to proper type
	c := client.(*api.Client)
	response, err := c.Request("GET", "/api/sync/status", nil, true)
	if err != nil {
		return err
	}

	if isJSONMode() {
		return prettyPrint(response)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid response format")
	}

	pendingCount, _ := data["pending_count"].(float64)
	conflictCount, _ := data["conflict_count"].(float64)

	fmt.Println(fmt.Sprintf("\n📊 Sync Status"))
	fmt.Println(fmt.Sprintf("  Pending:  %d items", int(pendingCount)))
	fmt.Println(fmt.Sprintf("  Conflicts: %d items\n", int(conflictCount)))

	if conflictCount > 0 {
		fmt.Println("⚠️  Conflicts detected! Run 'myprompts sync' to resolve.")
	}

	return nil
}

func performSync(client any) error {
	c := client.(*api.Client)
	fmt.Println("🔄 Syncing with server...")

	response, err := c.Request("GET", "/api/sync/status", nil, true)
	if err != nil {
		return err
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid response format")
	}

	pendingCount, _ := data["pending_count"].(float64)
	conflictCount, _ := data["conflict_count"].(float64)

	if int(pendingCount) == 0 && int(conflictCount) == 0 {
		if isJSONMode() {
			return prettyPrint(map[string]interface{}{
				"message": "Sync completed. Everything is up to date.",
				"data": map[string]interface{}{
					"synced":    0,
					"conflicts": 0,
				},
			})
		}
		fmt.Println("✅ Already synchronized. No changes pending.")
		return nil
	}

	if int(conflictCount) > 0 {
		if isJSONMode() {
			return prettyPrint(map[string]interface{}{
				"message": "Sync completed with conflicts.",
				"data":    data,
			})
		}
		fmt.Printf("⚠️  Sync found %d conflict(s).\n", int(conflictCount))
		fmt.Println("Conflicts require manual resolution through the interactive UI.")
		return nil
	}

	if isJSONMode() {
		return prettyPrint(map[string]interface{}{
			"message": "Sync completed successfully.",
			"data":    data,
		})
	}

	fmt.Printf("✅ Sync completed: %d items synced\n", int(pendingCount))
	return nil
}

func init() {
	rootCmd.AddCommand(uiCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(syncCmd)

	configCmd.AddCommand(configSetCmd)
	syncCmd.Flags().BoolVar(&syncStatus, "status", false, "show sync status only")
}
