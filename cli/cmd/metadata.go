package cmd

import (
	"github.com/dawillygene/my-prompt-repository/internal/completion"
	"github.com/spf13/cobra"
)

var categoryCmd = &cobra.Command{
	Use:   "category",
	Short: "Manage prompt categories",
	Long:  `Create, list, update, and delete prompt categories.`,
	Example: `  prompt category list
  prompt category create "Development" --description "Dev prompts"
  prompt category update tech "Technology" --description "Updated desc"
  prompt category delete old-category`,
}

var categoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		response, err := client.Request("GET", "/api/categories", nil, true)
		if err != nil {
			return err
		}
		return prettyPrint(response)
	},
}

var (
	categoryDesc string
)

var categoryCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new category",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		payload := map[string]any{
			"name": args[0],
		}
		if categoryDesc != "" {
			payload["description"] = categoryDesc
		}
		response, err := client.Request("POST", "/api/categories", payload, true)
		if err != nil {
			return err
		}
		// Invalidate cache after creating
		completion.InvalidateCacheKey("categories")
		return prettyPrint(response)
	},
}

var categoryUpdateCmd = &cobra.Command{
	Use:   "update <id-or-slug> <new-name>",
	Short: "Update a category",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		payload := map[string]any{
			"name": args[1],
		}
		if categoryDesc != "" {
			payload["description"] = categoryDesc
		}
		response, err := client.Request("PUT", "/api/categories/"+args[0], payload, true)
		if err != nil {
			return err
		}
		// Invalidate cache after updating
		completion.InvalidateCacheKey("categories")
		return prettyPrint(response)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Only complete the first argument (category ID)
		if len(args) == 0 && client != nil {
			completer := completion.NewCategoryCompleter(client)
			return completer.Complete(cmd, args, toComplete)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

var categoryDeleteCmd = &cobra.Command{
	Use:   "delete <id-or-slug>",
	Short: "Delete a category",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		response, err := client.Request("DELETE", "/api/categories/"+args[0], nil, true)
		if err != nil {
			return err
		}
		// Invalidate cache after deleting
		completion.InvalidateCacheKey("categories")
		return prettyPrint(response)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if client == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		completer := completion.NewCategoryCompleter(client)
		return completer.Complete(cmd, args, toComplete)
	},
}

// Tag commands
var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage prompt tags",
	Long:  `Create, list, update, and delete prompt tags.`,
	Example: `  prompt tag list
  prompt tag create "python" --description "Python related"
  prompt tag update go "golang" --description "Go language"
  prompt tag delete old-tag`,
}

var tagListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		response, err := client.Request("GET", "/api/tags", nil, true)
		if err != nil {
			return err
		}
		return prettyPrint(response)
	},
}

var (
	tagDesc string
)

var tagCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new tag",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		payload := map[string]any{
			"name": args[0],
		}
		if tagDesc != "" {
			payload["description"] = tagDesc
		}
		response, err := client.Request("POST", "/api/tags", payload, true)
		if err != nil {
			return err
		}
		// Invalidate cache after creating
		completion.InvalidateCacheKey("tags")
		return prettyPrint(response)
	},
}

var tagUpdateCmd = &cobra.Command{
	Use:   "update <id-or-slug> <new-name>",
	Short: "Update a tag",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		payload := map[string]any{
			"name": args[1],
		}
		if tagDesc != "" {
			payload["description"] = tagDesc
		}
		response, err := client.Request("PUT", "/api/tags/"+args[0], payload, true)
		if err != nil {
			return err
		}
		// Invalidate cache after updating
		completion.InvalidateCacheKey("tags")
		return prettyPrint(response)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Only complete the first argument (tag ID)
		if len(args) == 0 && client != nil {
			completer := completion.NewTagCompleter(client)
			return completer.Complete(cmd, args, toComplete)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

var tagDeleteCmd = &cobra.Command{
	Use:   "delete <id-or-slug>",
	Short: "Delete a tag",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		response, err := client.Request("DELETE", "/api/tags/"+args[0], nil, true)
		if err != nil {
			return err
		}
		// Invalidate cache after deleting
		completion.InvalidateCacheKey("tags")
		return prettyPrint(response)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if client == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		completer := completion.NewTagCompleter(client)
		return completer.Complete(cmd, args, toComplete)
	},
}

func init() {
	// Category commands
	rootCmd.AddCommand(categoryCmd)
	categoryCmd.AddCommand(categoryListCmd)
	categoryCmd.AddCommand(categoryCreateCmd)
	categoryCmd.AddCommand(categoryUpdateCmd)
	categoryCmd.AddCommand(categoryDeleteCmd)

	categoryCreateCmd.Flags().StringVar(&categoryDesc, "description", "", "category description")
	categoryUpdateCmd.Flags().StringVar(&categoryDesc, "description", "", "category description")

	// Tag commands
	rootCmd.AddCommand(tagCmd)
	tagCmd.AddCommand(tagListCmd)
	tagCmd.AddCommand(tagCreateCmd)
	tagCmd.AddCommand(tagUpdateCmd)
	tagCmd.AddCommand(tagDeleteCmd)

	tagCreateCmd.Flags().StringVar(&tagDesc, "description", "", "tag description")
	tagUpdateCmd.Flags().StringVar(&tagDesc, "description", "", "tag description")
}
