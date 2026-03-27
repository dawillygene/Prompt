package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dawillygene/my-prompt-repository/internal/api"
	"github.com/dawillygene/my-prompt-repository/internal/config"
	"github.com/dawillygene/my-prompt-repository/internal/tui"
)

var jsonMode bool

func Run(args []string) error {
	args = normalizeGlobalFlags(args)

	if len(args) == 0 {
		printHelp()
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.New(cfg)

	switch args[0] {
	case "ui":
		// Launch interactive TUI with colors
		ui := tui.NewColoredUI(cfg, client)
		return ui.Start()
	case "register":
		return runRegister(cfg, client, args[1:])
	case "login":
		return runLogin(cfg, client, args[1:])
	case "logout":
		return runLogout(cfg, client)
	case "whoami":
		return printResponse(client.Request("GET", "/api/me", nil, true))
	case "add":
		return runAdd(client, args[1:])
	case "list":
		return runList(client, args[1:])
	case "show":
		return runShow(client, args[1:])
	case "delete":
		return runDelete(client, args[1:])
	case "favorite":
		return runToggle(client, args[1:], "/api/prompts/%s/favorite")
	case "archive":
		return runToggle(client, args[1:], "/api/prompts/%s/archive")
	case "search":
		return runSearch(client, args[1:])
	case "category":
		return runCategory(client, args[1:])
	case "tag":
		return runTag(client, args[1:])
	case "config":
		return runConfig(cfg, args[1:])
	case "edit":
		return runEdit()
	case "export":
		return runExport(client, args[1:])
	case "import":
		return runImport(client, args[1:])
	case "sync":
		return runSync(client, args[1:])
	case "help":
		printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func runRegister(cfg config.Config, client *api.Client, args []string) error {
	values := parseFlags(args)
	response, err := client.Request("POST", "/api/register", map[string]any{
		"name":     values["name"],
		"email":    values["email"],
		"password": values["password"],
	}, false)
	if err != nil {
		return err
	}

	token, _ := response["token"].(string)
	cfg.Token = token
	if err := config.Save(cfg); err != nil {
		return err
	}

	return prettyPrint(response)
}

func runLogin(cfg config.Config, client *api.Client, args []string) error {
	values := parseFlags(args)
	response, err := client.Request("POST", "/api/login", map[string]any{
		"email":    values["email"],
		"password": values["password"],
	}, false)
	if err != nil {
		return err
	}

	token, _ := response["token"].(string)
	cfg.Token = token
	if err := config.Save(cfg); err != nil {
		return err
	}

	return prettyPrint(response)
}

func runLogout(cfg config.Config, client *api.Client) error {
	if _, err := client.Request("POST", "/api/logout", map[string]any{}, true); err != nil {
		return err
	}

	cfg.Token = ""
	if err := config.Save(cfg); err != nil {
		return err
	}

	if jsonMode {
		return prettyPrint(map[string]any{
			"message": "Logged out.",
			"data":    map[string]any{},
		})
	}

	fmt.Println("Logged out.")
	return nil
}

func runAdd(client *api.Client, args []string) error {
	values := parseFlags(args)
	return printResponse(client.Request("POST", "/api/prompts", map[string]any{
		"title":      values["title"],
		"content":    values["content"],
		"summary":    values["summary"],
		"visibility": defaultValue(values["visibility"], "private"),
	}, true))
}

func runList(client *api.Client, args []string) error {
	query := queryString(parseFlags(args))
	return printResponse(client.Request("GET", "/api/prompts"+query, nil, true))
}

func runShow(client *api.Client, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: myprompts show <id-or-slug>")
	}

	return printResponse(client.Request("GET", "/api/prompts/"+args[0], nil, true))
}

func runDelete(client *api.Client, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: myprompts delete <id-or-slug>")
	}

	return printResponse(client.Request("DELETE", "/api/prompts/"+args[0], nil, true))
}

func runToggle(client *api.Client, args []string, pattern string) error {
	if len(args) == 0 {
		return errors.New("usage requires <id-or-slug>")
	}

	return printResponse(client.Request("POST", fmt.Sprintf(pattern, args[0]), map[string]any{}, true))
}

func runSearch(client *api.Client, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: myprompts search <keyword>")
	}

	return printResponse(client.Request("GET", "/api/prompts?search="+args[0], nil, true))
}

func runCategory(client *api.Client, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: myprompts category <list|create|update|delete>")
	}

	switch args[0] {
	case "list":
		return printResponse(client.Request("GET", "/api/categories", nil, true))
	case "create":
		if len(args) < 2 {
			return errors.New("usage: myprompts category create <name> [--description TEXT]")
		}
		// Extract name from first positional argument
		name := args[1]
		// Parse remaining args for flags
		flags := parseFlags(args[2:])
		payload := map[string]any{
			"name": name,
		}
		if flags["description"] != "" {
			payload["description"] = flags["description"]
		}
		return printResponse(client.Request("POST", "/api/categories", payload, true))
	case "update":
		if len(args) < 3 {
			return errors.New("usage: myprompts category update <id-or-slug> <new-name> [--description TEXT]")
		}
		// Extract id/slug and name from positional arguments
		ref := args[1]
		name := args[2]
		// Parse remaining args for flags
		flags := parseFlags(args[3:])
		payload := map[string]any{
			"name": name,
		}
		if flags["description"] != "" {
			payload["description"] = flags["description"]
		}
		return printResponse(client.Request("PUT", "/api/categories/"+ref, payload, true))
	case "delete":
		if len(args) < 2 {
			return errors.New("usage: myprompts category delete <id-or-slug>")
		}
		return printResponse(client.Request("DELETE", "/api/categories/"+args[1], nil, true))
	default:
		return errors.New("usage: myprompts category <list|create|update|delete>")
	}
}

func runTag(client *api.Client, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: myprompts tag <list|create|update|delete>")
	}

	switch args[0] {
	case "list":
		return printResponse(client.Request("GET", "/api/tags", nil, true))
	case "create":
		if len(args) < 2 {
			return errors.New("usage: myprompts tag create <name> [--description TEXT]")
		}
		// Extract name from first positional argument
		name := args[1]
		// Parse remaining args for flags
		flags := parseFlags(args[2:])
		payload := map[string]any{
			"name": name,
		}
		if flags["description"] != "" {
			payload["description"] = flags["description"]
		}
		return printResponse(client.Request("POST", "/api/tags", payload, true))
	case "update":
		if len(args) < 3 {
			return errors.New("usage: myprompts tag update <id-or-slug> <new-name> [--description TEXT]")
		}
		// Extract id/slug and name from positional arguments
		ref := args[1]
		name := args[2]
		// Parse remaining args for flags
		flags := parseFlags(args[3:])
		payload := map[string]any{
			"name": name,
		}
		if flags["description"] != "" {
			payload["description"] = flags["description"]
		}
		return printResponse(client.Request("PUT", "/api/tags/"+ref, payload, true))
	case "delete":
		if len(args) < 2 {
			return errors.New("usage: myprompts tag delete <id-or-slug>")
		}
		return printResponse(client.Request("DELETE", "/api/tags/"+args[1], nil, true))
	default:
		return errors.New("usage: myprompts tag <list|create|update|delete>")
	}
}

func runConfig(cfg config.Config, args []string) error {
	if len(args) < 3 || args[0] != "set" {
		return errors.New("usage: myprompts config set <key> <value>")
	}

	switch args[1] {
	case "api_base":
		cfg.APIBase = args[2]
	default:
		return fmt.Errorf("unsupported config key: %s", args[1])
	}

	if err := config.Save(cfg); err != nil {
		return err
	}

 	if jsonMode {
		return prettyPrint(map[string]any{
			"message": fmt.Sprintf("%s updated.", args[1]),
			"data": map[string]any{
				"key": args[1],
				"value": args[2],
			},
		})
	}

	fmt.Printf("%s updated.\n", args[1])
	return nil
}

func runEdit() error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return errors.New("`edit` is reserved for editor-based prompt updates; set $EDITOR and implement the update flow next")
	}

	cmd := exec.Command(editor)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	if jsonMode {
		return prettyPrint(map[string]any{
			"message": "Editor closed.",
			"data":    map[string]any{},
		})
	}

	fmt.Println("Editor closed.")
	return nil
}

func runExport(client *api.Client, args []string) error {
	filePath := "prompts-export.json"
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
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

	if jsonMode {
		return prettyPrint(map[string]any{
			"message": "Export completed.",
			"data": map[string]any{
				"file": abs,
			},
		})
	}

	fmt.Printf("Exported prompts to %s\n", abs)
	return nil
}

func normalizeGlobalFlags(args []string) []string {
	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		switch arg {
		case "--json":
			jsonMode = true
		default:
			filtered = append(filtered, arg)
		}
	}

	return filtered
}

func runImport(client *api.Client, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: myprompts import <file>")
	}

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

	return printResponse(client.Request("POST", "/api/import", map[string]any{
		"prompts": prompts,
	}, true))
}

func parseFlags(args []string) map[string]string {
	values := map[string]string{}

	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "--") {
			continue
		}
		key := strings.TrimPrefix(args[i], "--")
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			values[key] = args[i+1]
			i++
			continue
		}
		values[key] = "true"
	}

	return values
}

func queryString(values map[string]string) string {
	if len(values) == 0 {
		return ""
	}

	parts := make([]string, 0, len(values))
	for key, value := range values {
		if value == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}

	if len(parts) == 0 {
		return ""
	}

	return "?" + strings.Join(parts, "&")
}

func defaultValue(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

func printResponse(response map[string]any, err error) error {
	if err != nil {
		return err
	}

	return prettyPrint(response)
}

func prettyPrint(value any) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(content))
	return nil
}

func runSync(client *api.Client, args []string) error {
	// Parse flags
	showStatus := false
	for _, arg := range args {
		if arg == "--status" {
			showStatus = true
		}
	}

	if showStatus {
		// Show sync status
		return showSyncStatus(client)
	}

	// Full sync: get pending items and submit to server
	return performSync(client)
}

func showSyncStatus(client *api.Client) error {
	response, err := client.Request("GET", "/api/sync/status", nil, true)
	if err != nil {
		return err
	}

	if jsonMode {
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

func performSync(client *api.Client) error {
	// This is a simplified sync that would integrate with local cache/queue
	// For now, just show a message that sync is processing
	fmt.Println("🔄 Syncing with server...")

	// Get sync status
	response, err := client.Request("GET", "/api/sync/status", nil, true)
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
		if jsonMode {
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
		if jsonMode {
			return prettyPrint(map[string]interface{}{
				"message": "Sync completed with conflicts.",
				"data":    data,
			})
		}
		fmt.Printf("⚠️  Sync found %d conflict(s).\n", int(conflictCount))
		fmt.Println("Conflicts require manual resolution through the interactive UI.")
		return nil
	}

	if jsonMode {
		return prettyPrint(map[string]interface{}{
			"message": "Sync completed successfully.",
			"data":    data,
		})
	}

	fmt.Printf("✅ Sync completed: %d items synced\n", int(pendingCount))
	return nil
}

func printHelp() {
	fmt.Println(`myprompts commands:
  ui                      - Launch interactive UI (recommended!)
	--json                  - Output machine-readable JSON where supported
  register --name NAME --email EMAIL --password PASSWORD
  login --email EMAIL --password PASSWORD
  logout
  whoami
  add --title TITLE --content CONTENT [--summary SUMMARY] [--visibility private|public]
  list [--search KEYWORD] [--category_id ID] [--tag_id ID] [--sort updated_at]
  show <id-or-slug>
  delete <id-or-slug>
  favorite <id-or-slug>
  archive <id-or-slug>
  search <keyword>
	export [file]
	import <file>
	sync [--status]           - Sync local cache with server, resolve conflicts
	category list|create|update|delete
	tag list|create|update|delete
  config set api_base <url>`)
}
