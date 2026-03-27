package cmd

import (
	"fmt"

	"github.com/dawillygene/my-prompt-repository/internal/completion"
	"github.com/dawillygene/my-prompt-repository/internal/interactive"
	"github.com/spf13/cobra"
)

var (
	addTitle      string
	addContent    string
	addSummary    string
	addVisibility string
	listSearch    string
	listCategory  string
	listTag       string
	listSort      string
	interactiveMode bool
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new prompt",
	Long:  `Create a new prompt with title, content, and optional metadata.`,
	Example: `  prompt add --title "Code Review" --content "Review this code carefully"
  prompt add --title "API Docs" --content "Document the API" --summary "API documentation helper"
  prompt add --title "Private Note" --content "Secret" --visibility private`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		response, err := client.Request("POST", "/api/prompts", map[string]any{
			"title":      addTitle,
			"content":    addContent,
			"summary":    addSummary,
			"visibility": defaultValue(addVisibility, "private"),
		}, true)
		if err != nil {
			return err
		}
		// Invalidate prompt cache after adding
		completion.InvalidateCacheKey("prompts")
		return prettyPrint(response)
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all prompts",
	Long:    `Display a list of your prompts with optional filtering.`,
	Example: `  prompt list
  prompt list --search "code review"
  prompt list --category tech --sort updated_at
  prompt ls --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		
		// Build query string
		query := ""
		params := map[string]string{
			"search":      listSearch,
			"category_id": listCategory,
			"tag_id":      listTag,
			"sort":        listSort,
		}
		
		first := true
		for key, value := range params {
			if value != "" {
				if first {
					query += "?"
					first = false
				} else {
					query += "&"
				}
				query += fmt.Sprintf("%s=%s", key, value)
			}
		}

		response, err := client.Request("GET", "/api/prompts"+query, nil, true)
		if err != nil {
			return err
		}
		return prettyPrint(response)
	},
}

var showCmd = &cobra.Command{
	Use:   "show [id-or-slug]",
	Short: "Show a specific prompt",
	Long:  `Display detailed information about a prompt by its ID or slug. Use --interactive for fuzzy search.`,
	Example: `  prompt show 123
  prompt show my-first-prompt
  prompt show --interactive
  prompt show -i --json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if interactiveMode {
			return nil // No args required in interactive mode
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		
		var promptID string
		
		if interactiveMode {
			// Fetch prompts and show interactive picker
			response, err := client.Request("GET", "/api/prompts", nil, true)
			if err != nil {
				return err
			}
			
			items := promptsToPickerItems(response)
			if len(items) == 0 {
				return fmt.Errorf("no prompts found")
			}
			
			selected, err := interactive.RunPicker(items, "show", false)
			if err != nil {
				return err
			}
			
			if len(selected) == 0 {
				return fmt.Errorf("no prompt selected")
			}
			
			promptID = selected[0]
		} else {
			promptID = args[0]
		}
		
		response, err := client.Request("GET", "/api/prompts/"+promptID, nil, true)
		if err != nil {
			return err
		}
		return prettyPrint(response)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if client == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		completer := completion.NewPromptCompleter(client)
		return completer.Complete(cmd, args, toComplete)
	},
}

var deleteCmd = &cobra.Command{
	Use:     "delete [id-or-slug]",
	Aliases: []string{"rm"},
	Short:   "Delete a prompt",
	Long:    `Permanently delete a prompt by its ID or slug. Use --interactive for fuzzy search.`,
	Example: `  prompt delete 123
  prompt rm old-prompt
  prompt delete --interactive
  prompt rm -i`,
	Args: func(cmd *cobra.Command, args []string) error {
		if interactiveMode {
			return nil
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		
		var promptID string
		
		if interactiveMode {
			response, err := client.Request("GET", "/api/prompts", nil, true)
			if err != nil {
				return err
			}
			
			items := promptsToPickerItems(response)
			if len(items) == 0 {
				return fmt.Errorf("no prompts found")
			}
			
			selected, err := interactive.RunPicker(items, "delete", false)
			if err != nil {
				return err
			}
			
			if len(selected) == 0 {
				return fmt.Errorf("no prompt selected")
			}
			
			promptID = selected[0]
		} else {
			promptID = args[0]
		}
		
		response, err := client.Request("DELETE", "/api/prompts/"+promptID, nil, true)
		if err != nil {
			return err
		}
		// Invalidate prompt cache after deleting
		completion.InvalidateCacheKey("prompts")
		return prettyPrint(response)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if client == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		completer := completion.NewPromptCompleter(client)
		return completer.Complete(cmd, args, toComplete)
	},
}

var searchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search prompts by keyword",
	Long:  `Search your prompts by keyword in title, content, or tags.`,
	Example: `  prompt search "code review"
  prompt search API --json`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		keyword := args[0]
		response, err := client.Request("GET", "/api/prompts?search="+keyword, nil, true)
		if err != nil {
			return err
		}
		return prettyPrint(response)
	},
}

var favoriteCmd = &cobra.Command{
	Use:     "favorite [id-or-slug]",
	Aliases: []string{"fav"},
	Short:   "Toggle favorite status of a prompt",
	Long:    `Mark or unmark a prompt as favorite. Use --interactive for fuzzy search.`,
	Example: `  prompt favorite 123
  prompt fav my-best-prompt
  prompt fav --interactive
  prompt fav -i`,
	Args: func(cmd *cobra.Command, args []string) error {
		if interactiveMode {
			return nil
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		
		var promptID string
		
		if interactiveMode {
			response, err := client.Request("GET", "/api/prompts", nil, true)
			if err != nil {
				return err
			}
			
			items := promptsToPickerItems(response)
			if len(items) == 0 {
				return fmt.Errorf("no prompts found")
			}
			
			selected, err := interactive.RunPicker(items, "favorite", false)
			if err != nil {
				return err
			}
			
			if len(selected) == 0 {
				return fmt.Errorf("no prompt selected")
			}
			
			promptID = selected[0]
		} else {
			promptID = args[0]
		}
		
		response, err := client.Request("POST", fmt.Sprintf("/api/prompts/%s/favorite", promptID), map[string]any{}, true)
		if err != nil {
			return err
		}
		return prettyPrint(response)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if client == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		completer := completion.NewPromptCompleter(client)
		return completer.Complete(cmd, args, toComplete)
	},
}

var archiveCmd = &cobra.Command{
	Use:     "archive <id-or-slug>",
	Aliases: []string{"arch"},
	Short:   "Toggle archive status of a prompt",
	Long:    `Archive or unarchive a prompt.`,
	Example: `  prompt archive 123
  prompt arch old-prompt`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		response, err := client.Request("POST", fmt.Sprintf("/api/prompts/%s/archive", args[0]), map[string]any{}, true)
		if err != nil {
			return err
		}
		return prettyPrint(response)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if client == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		completer := completion.NewPromptCompleter(client)
		return completer.Complete(cmd, args, toComplete)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(favoriteCmd)
	rootCmd.AddCommand(archiveCmd)

	// Add command flags
	addCmd.Flags().StringVar(&addTitle, "title", "", "prompt title (required)")
	addCmd.Flags().StringVar(&addContent, "content", "", "prompt content (required)")
	addCmd.Flags().StringVar(&addSummary, "summary", "", "optional summary")
	addCmd.Flags().StringVar(&addVisibility, "visibility", "private", "visibility: private or public")
	addCmd.MarkFlagRequired("title")
	addCmd.MarkFlagRequired("content")

	// List command flags
	listCmd.Flags().StringVar(&listSearch, "search", "", "search keyword")
	listCmd.Flags().StringVar(&listCategory, "category-id", "", "filter by category ID")
	listCmd.Flags().StringVar(&listTag, "tag-id", "", "filter by tag ID")
	listCmd.Flags().StringVar(&listSort, "sort", "", "sort field (e.g., updated_at)")

	// Interactive flags
	showCmd.Flags().BoolVarP(&interactiveMode, "interactive", "i", false, "interactive fuzzy search mode")
	deleteCmd.Flags().BoolVarP(&interactiveMode, "interactive", "i", false, "interactive fuzzy search mode")
	favoriteCmd.Flags().BoolVarP(&interactiveMode, "interactive", "i", false, "interactive fuzzy search mode")
	archiveCmd.Flags().BoolVarP(&interactiveMode, "interactive", "i", false, "interactive fuzzy search mode")
}

// Helper function to convert API response to picker items
func promptsToPickerItems(response map[string]any) []interactive.PromptItem {
	data, ok := response["data"]
	if !ok {
		return []interactive.PromptItem{}
	}

	prompts, ok := data.([]interface{})
	if !ok {
		return []interactive.PromptItem{}
	}

	items := make([]interactive.PromptItem, 0, len(prompts))
	for _, p := range prompts {
		prompt, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		id := ""
		if idFloat, ok := prompt["id"].(float64); ok {
			id = fmt.Sprintf("%.0f", idFloat)
		} else if idStr, ok := prompt["id"].(string); ok {
			id = idStr
		}

		title, _ := prompt["title"].(string)
		summary, _ := prompt["summary"].(string)
		isFavorite, _ := prompt["is_favorite"].(bool)
		isArchived, _ := prompt["is_archived"].(bool)

		items = append(items, interactive.PromptItem{
			ID:         id,
			Title:      title,
			Summary:    summary,
			IsFavorite: isFavorite,
			IsArchived: isArchived,
		})
	}

	return items
}

func defaultValue(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
