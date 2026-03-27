package completion

import (
	"fmt"
	"strings"
	"time"

	"github.com/dawillygene/my-prompt-repository/internal/api"
	"github.com/spf13/cobra"
)

var (
	// Global cache with 5-minute TTL
	cache = NewCache(5 * time.Minute)
)

// PromptCompleter provides completion for prompt IDs and slugs
type PromptCompleter struct {
	client *api.Client
}

// NewPromptCompleter creates a new prompt completer
func NewPromptCompleter(client *api.Client) *PromptCompleter {
	return &PromptCompleter{client: client}
}

// Complete returns prompt ID/slug completions
func (pc *PromptCompleter) Complete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Check cache first
	if cached, ok := cache.Get("prompts"); ok {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	// Fetch from API
	prompts, err := pc.fetchPrompts()
	if err != nil {
		// Return empty on error, don't break completion
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}

	// Cache results
	cache.Set("prompts", prompts)

	return filterCompletions(prompts, toComplete), cobra.ShellCompDirectiveNoFileComp
}

func (pc *PromptCompleter) fetchPrompts() ([]string, error) {
	response, err := pc.client.Request("GET", "/api/prompts", nil, true)
	if err != nil {
		return nil, err
	}

	data, ok := response["data"]
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	prompts, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	var completions []string
	for _, p := range prompts {
		prompt, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		// Add slug (preferred for completion)
		if slug, ok := prompt["slug"].(string); ok && slug != "" {
			title, _ := prompt["title"].(string)
			completions = append(completions, fmt.Sprintf("%s\t%s", slug, title))
		}

		// Also add ID as fallback
		if id, ok := prompt["id"].(float64); ok {
			title, _ := prompt["title"].(string)
			completions = append(completions, fmt.Sprintf("%d\t%s", int(id), title))
		}
	}

	return completions, nil
}

// CategoryCompleter provides completion for category IDs and slugs
type CategoryCompleter struct {
	client *api.Client
}

// NewCategoryCompleter creates a new category completer
func NewCategoryCompleter(client *api.Client) *CategoryCompleter {
	return &CategoryCompleter{client: client}
}

// Complete returns category ID/slug completions
func (cc *CategoryCompleter) Complete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Check cache first
	if cached, ok := cache.Get("categories"); ok {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	// Fetch from API
	categories, err := cc.fetchCategories()
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}

	// Cache results
	cache.Set("categories", categories)

	return filterCompletions(categories, toComplete), cobra.ShellCompDirectiveNoFileComp
}

func (cc *CategoryCompleter) fetchCategories() ([]string, error) {
	response, err := cc.client.Request("GET", "/api/categories", nil, true)
	if err != nil {
		return nil, err
	}

	data, ok := response["data"]
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	categories, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	var completions []string
	for _, c := range categories {
		category, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		if slug, ok := category["slug"].(string); ok && slug != "" {
			name, _ := category["name"].(string)
			completions = append(completions, fmt.Sprintf("%s\t%s", slug, name))
		}

		if id, ok := category["id"].(float64); ok {
			name, _ := category["name"].(string)
			completions = append(completions, fmt.Sprintf("%d\t%s", int(id), name))
		}
	}

	return completions, nil
}

// TagCompleter provides completion for tag IDs and slugs
type TagCompleter struct {
	client *api.Client
}

// NewTagCompleter creates a new tag completer
func NewTagCompleter(client *api.Client) *TagCompleter {
	return &TagCompleter{client: client}
}

// Complete returns tag ID/slug completions
func (tc *TagCompleter) Complete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Check cache first
	if cached, ok := cache.Get("tags"); ok {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	// Fetch from API
	tags, err := tc.fetchTags()
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}

	// Cache results
	cache.Set("tags", tags)

	return filterCompletions(tags, toComplete), cobra.ShellCompDirectiveNoFileComp
}

func (tc *TagCompleter) fetchTags() ([]string, error) {
	response, err := tc.client.Request("GET", "/api/tags", nil, true)
	if err != nil {
		return nil, err
	}

	data, ok := response["data"]
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	tags, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format")
	}

	var completions []string
	for _, t := range tags {
		tag, ok := t.(map[string]interface{})
		if !ok {
			continue
		}

		if slug, ok := tag["slug"].(string); ok && slug != "" {
			name, _ := tag["name"].(string)
			completions = append(completions, fmt.Sprintf("%s\t%s", slug, name))
		}

		if id, ok := tag["id"].(float64); ok {
			name, _ := tag["name"].(string)
			completions = append(completions, fmt.Sprintf("%d\t%s", int(id), name))
		}
	}

	return completions, nil
}

// filterCompletions filters completions based on the toComplete prefix
func filterCompletions(completions []string, toComplete string) []string {
	if toComplete == "" {
		return completions
	}

	var filtered []string
	for _, c := range completions {
		// Extract just the value part (before tab) for comparison
		value := strings.Split(c, "\t")[0]
		if strings.HasPrefix(value, toComplete) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// InvalidateCache clears the completion cache
func InvalidateCache() {
	cache.Clear()
}

// InvalidateCacheKey clears a specific cache key
func InvalidateCacheKey(key string) {
	cache.Invalidate(key)
}
