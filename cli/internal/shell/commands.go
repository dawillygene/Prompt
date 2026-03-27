package shell

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/lipgloss"
	"github.com/dawillygene/my-prompt-repository/internal/config"
)

// runCommand executes a command and returns the output
func (m *Model) runCommand(input string) ([]OutputLine, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}

	parts := strings.Fields(input)
	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	// Session
	case "help", "man", "?":
		return m.cmdHelp(), nil
	case "exit", "quit", "q":
		m.quitting = true
		return nil, nil
	case "clear", "cls":
		// clearScreen flag is handled in executeCommand/Update
		return nil, nil

	// Navigation
	case "ls":
		return m.cmdLs(args), nil
	case "cd":
		return m.cmdCd(args), nil
	case "pwd":
		return m.cmdPwd(), nil
	case "tree":
		return m.cmdTree(), nil

	// Prompt operations (file-like)
	case "cat":
		return m.cmdCat(args), nil
	case "touch":
		return m.cmdTouch(args), nil
	case "add":
		return m.cmdAdd(args), nil
	case "edit", "vi", "vim", "nano":
		return m.cmdEdit(args), nil
	case "rm":
		return m.cmdRm(args), nil
	case "mv":
		return m.cmdMv(args), nil
	case "cp":
		return m.cmdCp(args), nil
	case "copy":
		return m.cmdCopy(args), nil

	// Category operations (directory-like)
	case "mkdir":
		return m.cmdMkdir(args), nil
	case "rmdir":
		return m.cmdRmdir(args), nil

	// Search
	case "find", "search":
		return m.cmdFind(args), nil
	case "grep":
		return m.cmdGrep(args), nil

	// Favorites & Archive
	case "star", "fav", "favorite":
		return m.cmdStar(args), nil
	case "archive":
		return m.cmdArchive(args), nil

	// Version control
	case "history", "log":
		return m.cmdHistory(args), nil
	case "diff":
		return m.cmdDiff(args), nil

	// Tags
	case "tag":
		return m.cmdTag(args), nil
	case "tags":
		return m.cmdTags(), nil

	// Sync & Export
	case "export":
		return m.cmdExport(), nil
	case "import":
		return m.cmdImport(args), nil
	case "sync":
		return m.cmdSync(), nil

	// Auth
	case "login":
		return m.cmdLogin(args), nil
	case "logout":
		return m.cmdLogout(), nil
	case "register":
		return m.cmdRegister(args), nil
	case "whoami":
		return m.cmdWhoami(), nil

	// Config
	case "config":
		return m.cmdConfig(args), nil

	default:
		return []OutputLine{{
			Text:  fmt.Sprintf("Command not found: %s. Type 'help' for commands.", cmd),
			Style: errorStyle,
		}}, nil
	}
}

// ============ HELP ============

func (m *Model) cmdHelp() []OutputLine {
	path := "~"
	if m.currentCat != "" {
		path = "~/" + m.currentCat
	}
	return []OutputLine{
		{Text: "", Style: lipgloss.NewStyle()},
		{Text: fmt.Sprintf("PROMPT CLI - Current: %s", path), Style: titleStyle},
		{Text: "", Style: lipgloss.NewStyle()},
		{Text: "NAVIGATION", Style: accentStyle},
		{Text: "  ls [-a|-l]        List prompts", Style: helpDescStyle},
		{Text: "  cd <category>     Enter category (cd .. to go back)", Style: helpDescStyle},
		{Text: "  pwd               Show current path", Style: helpDescStyle},
		{Text: "  tree              Show all categories & prompts", Style: helpDescStyle},
		{Text: "", Style: lipgloss.NewStyle()},
		{Text: "PROMPTS", Style: accentStyle},
		{Text: "  cat <id>          Show prompt content", Style: helpDescStyle},
		{Text: "  touch <title>     Quick create prompt", Style: helpDescStyle},
		{Text: "  add               Create prompt (interactive)", Style: helpDescStyle},
		{Text: "  edit <id>         Edit in $EDITOR", Style: helpDescStyle},
		{Text: "  rm <id>           Delete prompt", Style: helpDescStyle},
		{Text: "  mv <id> <cat>     Move to category", Style: helpDescStyle},
		{Text: "  cp <id> <name>    Copy prompt", Style: helpDescStyle},
		{Text: "", Style: lipgloss.NewStyle()},
		{Text: "CATEGORIES", Style: accentStyle},
		{Text: "  mkdir <name>      Create category", Style: helpDescStyle},
		{Text: "  rmdir <name>      Delete category", Style: helpDescStyle},
		{Text: "", Style: lipgloss.NewStyle()},
		{Text: "SEARCH & ORGANIZE", Style: accentStyle},
		{Text: "  find <keyword>    Search prompts", Style: helpDescStyle},
		{Text: "  grep <pattern>    Search in content", Style: helpDescStyle},
		{Text: "  star <id>         Toggle favorite", Style: helpDescStyle},
		{Text: "  archive <id>      Archive/unarchive", Style: helpDescStyle},
		{Text: "  tag <id> <tag>    Add tag to prompt", Style: helpDescStyle},
		{Text: "", Style: lipgloss.NewStyle()},
		{Text: "SYNC", Style: accentStyle},
		{Text: "  export            Export to JSON", Style: helpDescStyle},
		{Text: "  import <file>     Import from JSON", Style: helpDescStyle},
		{Text: "  sync              Sync with server", Style: helpDescStyle},
		{Text: "", Style: lipgloss.NewStyle()},
		{Text: "AUTH", Style: accentStyle},
		{Text: "  login <email> <pass>", Style: helpDescStyle},
		{Text: "  register <name> <email> <pass>", Style: helpDescStyle},
		{Text: "  logout / whoami", Style: helpDescStyle},
		{Text: "", Style: lipgloss.NewStyle()},
		{Text: "Tab=autocomplete  ↑↓=history  q=quit", Style: helpDescStyle},
		{Text: "", Style: lipgloss.NewStyle()},
	}
}

// ============ NAVIGATION ============

func (m *Model) cmdLs(args []string) []OutputLine {
	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	showAll := false
	showLong := false
	for _, arg := range args {
		if arg == "-a" || arg == "--all" {
			showAll = true
		}
		if arg == "-l" || arg == "--long" {
			showLong = true
		}
	}

	// Build query
	query := "/api/prompts"
	if m.currentCatID > 0 {
		query = fmt.Sprintf("/api/prompts?category_id=%.0f", m.currentCatID)
	}
	if !showAll {
		if strings.Contains(query, "?") {
			query += "&is_archived=false"
		} else {
			query += "?is_archived=false"
		}
	}

	response, err := m.client.Request("GET", query, nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	data, ok := response["data"].([]interface{})
	if !ok || len(data) == 0 {
		return []OutputLine{{Text: "(empty)", Style: helpDescStyle}}
	}

	// Cache for autocomplete
	m.cachedPrompts = make([]map[string]interface{}, 0)

	var lines []OutputLine
	for _, p := range data {
		prompt, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		m.cachedPrompts = append(m.cachedPrompts, prompt)

		id := ""
		if idFloat, ok := prompt["id"].(float64); ok {
			id = fmt.Sprintf("%.0f", idFloat)
		}
		title, _ := prompt["title"].(string)
		isFav, _ := prompt["is_favorite"].(bool)
		isArchived, _ := prompt["is_archived"].(bool)

		prefix := "  "
		if isFav {
			prefix = "⭐"
		}
		if isArchived {
			prefix = "📦"
		}

		if showLong {
			createdAt, _ := prompt["created_at"].(string)
			cat, _ := prompt["category"].(map[string]interface{})
			catName := "-"
			if cat != nil {
				catName, _ = cat["name"].(string)
			}
			lines = append(lines, OutputLine{
				Text:  fmt.Sprintf("%s %4s  %-12s  %s  %s", prefix, id, catName, createdAt[:10], title),
				Style: lipgloss.NewStyle(),
			})
		} else {
			lines = append(lines, OutputLine{
				Text:  fmt.Sprintf("%s %s\t%s", prefix, id, title),
				Style: lipgloss.NewStyle(),
			})
		}
	}

	return lines
}

func (m *Model) cmdCd(args []string) []OutputLine {
	if len(args) == 0 || args[0] == "~" || args[0] == "/" {
		m.pendingCat = ""
		m.pendingCatID = 0
		m.pendingCd = true
		return []OutputLine{{Text: "~", Style: successStyle}}
	}

	target := args[0]

	if target == ".." {
		m.pendingCat = ""
		m.pendingCatID = 0
		m.pendingCd = true
		return []OutputLine{{Text: "~", Style: successStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	// Fetch categories
	response, err := m.client.Request("GET", "/api/categories", nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	data, ok := response["data"].([]interface{})
	if !ok {
		return []OutputLine{{Text: "No categories found", Style: errorStyle}}
	}

	// Find matching category
	for _, c := range data {
		cat, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := cat["name"].(string)
		slug, _ := cat["slug"].(string)
		id, _ := cat["id"].(float64)

		if strings.EqualFold(name, target) || strings.EqualFold(slug, target) || fmt.Sprintf("%.0f", id) == target {
			m.pendingCat = name
			m.pendingCatID = id
			m.pendingCd = true
			return []OutputLine{{Text: fmt.Sprintf("~/%s", name), Style: successStyle}}
		}
	}

	return []OutputLine{{Text: fmt.Sprintf("Category not found: %s", target), Style: errorStyle}}
}

func (m *Model) cmdPwd() []OutputLine {
	if m.currentCat == "" {
		return []OutputLine{{Text: "~", Style: titleStyle}}
	}
	return []OutputLine{{Text: fmt.Sprintf("~/%s", m.currentCat), Style: titleStyle}}
}

func (m *Model) cmdTree() []OutputLine {
	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	catResp, err := m.client.Request("GET", "/api/categories", nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	promptResp, err := m.client.Request("GET", "/api/prompts", nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	categories, _ := catResp["data"].([]interface{})
	prompts, _ := promptResp["data"].([]interface{})

	lines := []OutputLine{
		{Text: "📂 ~", Style: titleStyle},
	}

	// Group prompts by category
	grouped := make(map[float64][]map[string]interface{})
	uncategorized := []map[string]interface{}{}

	for _, p := range prompts {
		prompt, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		catID, ok := prompt["category_id"].(float64)
		if ok && catID > 0 {
			grouped[catID] = append(grouped[catID], prompt)
		} else {
			uncategorized = append(uncategorized, prompt)
		}
	}

	// Show categories
	for i, c := range categories {
		cat, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		catID, _ := cat["id"].(float64)
		name, _ := cat["name"].(string)
		catPrompts := grouped[catID]

		isLast := i == len(categories)-1 && len(uncategorized) == 0
		prefix := "├──"
		if isLast {
			prefix = "└──"
		}

		lines = append(lines, OutputLine{
			Text:  fmt.Sprintf("%s 📁 %s/", prefix, name),
			Style: accentStyle,
		})

		childPrefix := "│   "
		if isLast {
			childPrefix = "    "
		}

		for j, p := range catPrompts {
			title, _ := p["title"].(string)
			id := ""
			if idFloat, ok := p["id"].(float64); ok {
				id = fmt.Sprintf("%.0f", idFloat)
			}
			isFav, _ := p["is_favorite"].(bool)
			icon := "📄"
			if isFav {
				icon = "⭐"
			}

			itemPrefix := childPrefix + "├──"
			if j == len(catPrompts)-1 {
				itemPrefix = childPrefix + "└──"
			}
			lines = append(lines, OutputLine{
				Text:  fmt.Sprintf("%s %s %s (%s)", itemPrefix, icon, title, id),
				Style: helpDescStyle,
			})
		}
	}

	// Show uncategorized
	if len(uncategorized) > 0 {
		for j, p := range uncategorized {
			title, _ := p["title"].(string)
			id := ""
			if idFloat, ok := p["id"].(float64); ok {
				id = fmt.Sprintf("%.0f", idFloat)
			}
			isFav, _ := p["is_favorite"].(bool)
			icon := "📄"
			if isFav {
				icon = "⭐"
			}

			prefix := "├──"
			if j == len(uncategorized)-1 {
				prefix = "└──"
			}
			lines = append(lines, OutputLine{
				Text:  fmt.Sprintf("%s %s %s (%s)", prefix, icon, title, id),
				Style: helpDescStyle,
			})
		}
	}

	if len(prompts) == 0 && len(categories) == 0 {
		lines = append(lines, OutputLine{Text: "└── (empty)", Style: helpDescStyle})
	}

	return lines
}

// ============ PROMPT OPERATIONS ============

func (m *Model) cmdCat(args []string) []OutputLine {
	if len(args) == 0 {
		return []OutputLine{{Text: "Usage: cat <id>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	response, err := m.client.Request("GET", "/api/prompts/"+args[0], nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return []OutputLine{{Text: "Prompt not found", Style: errorStyle}}
	}

	title, _ := data["title"].(string)
	content, _ := data["content"].(string)
	summary, _ := data["summary"].(string)

	lines := []OutputLine{
		{Text: "", Style: lipgloss.NewStyle()},
		{Text: title, Style: titleStyle},
	}

	if summary != "" {
		lines = append(lines, OutputLine{Text: summary, Style: helpDescStyle})
	}

	lines = append(lines, OutputLine{Text: "", Style: lipgloss.NewStyle()})
	
	// Split content into lines
	for _, line := range strings.Split(content, "\n") {
		lines = append(lines, OutputLine{Text: line, Style: lipgloss.NewStyle()})
	}

	return lines
}

func (m *Model) cmdTouch(args []string) []OutputLine {
	if len(args) == 0 {
		return []OutputLine{{Text: "Usage: touch <title>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	title := strings.Join(args, " ")
	payload := map[string]any{
		"title":   title,
		"content": "",
	}

	if m.currentCatID > 0 {
		payload["category_id"] = m.currentCatID
	}

	response, err := m.client.Request("POST", "/api/prompts", payload, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	data, ok := response["data"].(map[string]interface{})
	if ok {
		if id, ok := data["id"].(float64); ok {
			return []OutputLine{{Text: fmt.Sprintf("✓ Created: %s (id: %.0f)", title, id), Style: successStyle}}
		}
	}

	return []OutputLine{{Text: fmt.Sprintf("✓ Created: %s", title), Style: successStyle}}
}

func (m *Model) cmdAdd(args []string) []OutputLine {
	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	// Return instructions - the actual mode switch happens via startAdd flag
	return []OutputLine{
		{Text: "", Style: lipgloss.NewStyle()},
		{Text: "Creating new prompt (Esc to cancel)", Style: accentStyle},
		{Text: "", Style: lipgloss.NewStyle()},
	}
}

func (m *Model) cmdEdit(args []string) []OutputLine {
	if len(args) == 0 {
		return []OutputLine{{Text: "Usage: edit <id>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	// Get current prompt
	response, err := m.client.Request("GET", "/api/prompts/"+args[0], nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return []OutputLine{{Text: "Prompt not found", Style: errorStyle}}
	}

	title, _ := data["title"].(string)
	content, _ := data["content"].(string)

	// Create temp file
	tmpFile, err := os.CreateTemp("", "prompt-*.md")
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	tmpFile.WriteString(fmt.Sprintf("# %s\n\n%s", title, content))
	tmpFile.Close()

	// Open editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Editor error: %v", err), Style: errorStyle}}
	}

	// Read updated content
	newContent, err := os.ReadFile(tmpPath)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	// Update via API
	_, err = m.client.Request("PUT", "/api/prompts/"+args[0], map[string]any{
		"content": string(newContent),
	}, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	return []OutputLine{{Text: fmt.Sprintf("✓ Updated prompt %s", args[0]), Style: successStyle}}
}

func (m *Model) cmdRm(args []string) []OutputLine {
	if len(args) == 0 {
		return []OutputLine{{Text: "Usage: rm <id>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	_, err := m.client.Request("DELETE", "/api/prompts/"+args[0], nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	return []OutputLine{{Text: fmt.Sprintf("✓ Deleted prompt %s", args[0]), Style: successStyle}}
}

func (m *Model) cmdMv(args []string) []OutputLine {
	if len(args) < 2 {
		return []OutputLine{{Text: "Usage: mv <id> <category>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	promptID := args[0]
	targetCat := args[1]

	// Find category ID
	catResp, err := m.client.Request("GET", "/api/categories", nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	categories, _ := catResp["data"].([]interface{})
	var catID float64 = 0

	for _, c := range categories {
		cat, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := cat["name"].(string)
		id, _ := cat["id"].(float64)
		if strings.EqualFold(name, targetCat) || fmt.Sprintf("%.0f", id) == targetCat {
			catID = id
			break
		}
	}

	if catID == 0 {
		return []OutputLine{{Text: fmt.Sprintf("Category not found: %s", targetCat), Style: errorStyle}}
	}

	_, err = m.client.Request("PUT", "/api/prompts/"+promptID, map[string]any{
		"category_id": catID,
	}, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	return []OutputLine{{Text: fmt.Sprintf("✓ Moved prompt %s to %s", promptID, targetCat), Style: successStyle}}
}

func (m *Model) cmdCp(args []string) []OutputLine {
	if len(args) < 2 {
		return []OutputLine{{Text: "Usage: cp <id> <new-title>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	promptID := args[0]
	newTitle := strings.Join(args[1:], " ")

	// Get original prompt
	response, err := m.client.Request("GET", "/api/prompts/"+promptID, nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return []OutputLine{{Text: "Prompt not found", Style: errorStyle}}
	}

	content, _ := data["content"].(string)
	summary, _ := data["summary"].(string)
	catID, _ := data["category_id"].(float64)

	// Create copy
	payload := map[string]any{
		"title":   newTitle,
		"content": content,
		"summary": summary,
	}
	if catID > 0 {
		payload["category_id"] = catID
	}

	_, err = m.client.Request("POST", "/api/prompts", payload, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	return []OutputLine{{Text: fmt.Sprintf("✓ Copied to: %s", newTitle), Style: successStyle}}
}

// cmdCopy copies prompt content to clipboard
func (m *Model) cmdCopy(args []string) []OutputLine {
	if len(args) == 0 {
		return []OutputLine{{Text: "Usage: copy <id>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	promptID := args[0]

	// Get prompt
	response, err := m.client.Request("GET", "/api/prompts/"+promptID, nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return []OutputLine{{Text: "Prompt not found", Style: errorStyle}}
	}

	content, _ := data["content"].(string)
	title, _ := data["title"].(string)

	// Copy to clipboard
	err = clipboard.WriteAll(content)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error copying to clipboard: %v", err), Style: errorStyle}}
	}

	return []OutputLine{{Text: fmt.Sprintf("✓ Copied \"%s\" to clipboard", title), Style: successStyle}}
}

// ============ CATEGORY OPERATIONS ============

func (m *Model) cmdMkdir(args []string) []OutputLine {
	if len(args) == 0 {
		return []OutputLine{{Text: "Usage: mkdir <name>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	name := strings.Join(args, " ")
	_, err := m.client.Request("POST", "/api/categories", map[string]any{"name": name}, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	return []OutputLine{{Text: fmt.Sprintf("✓ Created category: %s", name), Style: successStyle}}
}

func (m *Model) cmdRmdir(args []string) []OutputLine {
	if len(args) == 0 {
		return []OutputLine{{Text: "Usage: rmdir <name>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	target := args[0]

	// Find category
	catResp, err := m.client.Request("GET", "/api/categories", nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	categories, _ := catResp["data"].([]interface{})
	var catID string

	for _, c := range categories {
		cat, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := cat["name"].(string)
		id, _ := cat["id"].(float64)
		if strings.EqualFold(name, target) || fmt.Sprintf("%.0f", id) == target {
			catID = fmt.Sprintf("%.0f", id)
			break
		}
	}

	if catID == "" {
		return []OutputLine{{Text: fmt.Sprintf("Category not found: %s", target), Style: errorStyle}}
	}

	_, err = m.client.Request("DELETE", "/api/categories/"+catID, nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	return []OutputLine{{Text: fmt.Sprintf("✓ Deleted category: %s", target), Style: successStyle}}
}

func (m *Model) cmdCategories() []OutputLine {
	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	response, err := m.client.Request("GET", "/api/categories", nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	data, ok := response["data"].([]interface{})
	if !ok || len(data) == 0 {
		return []OutputLine{{Text: "No categories", Style: helpDescStyle}}
	}

	// Cache for autocomplete
	m.cachedCategories = make([]map[string]interface{}, 0)

	lines := []OutputLine{{Text: "📁 Categories", Style: titleStyle}}
	for _, c := range data {
		cat, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		m.cachedCategories = append(m.cachedCategories, cat)

		id := ""
		if idFloat, ok := cat["id"].(float64); ok {
			id = fmt.Sprintf("%.0f", idFloat)
		}
		name, _ := cat["name"].(string)
		lines = append(lines, OutputLine{
			Text:  fmt.Sprintf("  %s. %s/", id, name),
			Style: helpDescStyle,
		})
	}

	return lines
}

// ============ SEARCH ============

func (m *Model) cmdFind(args []string) []OutputLine {
	if len(args) == 0 {
		return []OutputLine{{Text: "Usage: find <keyword>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	query := strings.Join(args, " ")
	response, err := m.client.Request("GET", "/api/prompts?search="+query, nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	data, ok := response["data"].([]interface{})
	if !ok || len(data) == 0 {
		return []OutputLine{{Text: fmt.Sprintf("No results for: %s", query), Style: helpDescStyle}}
	}

	lines := []OutputLine{{Text: fmt.Sprintf("Results for '%s':", query), Style: titleStyle}}
	for _, p := range data {
		prompt, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		id := ""
		if idFloat, ok := prompt["id"].(float64); ok {
			id = fmt.Sprintf("%.0f", idFloat)
		}
		title, _ := prompt["title"].(string)
		lines = append(lines, OutputLine{
			Text:  fmt.Sprintf("  %s  %s", id, title),
			Style: helpDescStyle,
		})
	}

	return lines
}

func (m *Model) cmdGrep(args []string) []OutputLine {
	if len(args) == 0 {
		return []OutputLine{{Text: "Usage: grep <pattern>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	pattern := strings.ToLower(strings.Join(args, " "))
	response, err := m.client.Request("GET", "/api/prompts", nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	data, ok := response["data"].([]interface{})
	if !ok {
		return []OutputLine{{Text: "No prompts", Style: helpDescStyle}}
	}

	var lines []OutputLine
	for _, p := range data {
		prompt, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		content, _ := prompt["content"].(string)
		title, _ := prompt["title"].(string)

		if strings.Contains(strings.ToLower(content), pattern) {
			id := ""
			if idFloat, ok := prompt["id"].(float64); ok {
				id = fmt.Sprintf("%.0f", idFloat)
			}
			lines = append(lines, OutputLine{
				Text:  fmt.Sprintf("  %s  %s", id, title),
				Style: helpDescStyle,
			})
		}
	}

	if len(lines) == 0 {
		return []OutputLine{{Text: fmt.Sprintf("No matches for: %s", pattern), Style: helpDescStyle}}
	}

	return append([]OutputLine{{Text: fmt.Sprintf("Matches for '%s':", pattern), Style: titleStyle}}, lines...)
}

// ============ FAVORITES & ARCHIVE ============

func (m *Model) cmdStar(args []string) []OutputLine {
	if len(args) == 0 {
		return []OutputLine{{Text: "Usage: star <id>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	_, err := m.client.Request("POST", fmt.Sprintf("/api/prompts/%s/favorite", args[0]), map[string]any{}, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	return []OutputLine{{Text: fmt.Sprintf("⭐ Toggled star for prompt %s", args[0]), Style: successStyle}}
}

func (m *Model) cmdArchive(args []string) []OutputLine {
	if len(args) == 0 {
		return []OutputLine{{Text: "Usage: archive <id>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	_, err := m.client.Request("POST", fmt.Sprintf("/api/prompts/%s/archive", args[0]), map[string]any{}, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	return []OutputLine{{Text: fmt.Sprintf("📦 Toggled archive for prompt %s", args[0]), Style: successStyle}}
}

// ============ VERSION CONTROL ============

func (m *Model) cmdHistory(args []string) []OutputLine {
	if len(args) == 0 {
		return []OutputLine{{Text: "Usage: history <id>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	response, err := m.client.Request("GET", fmt.Sprintf("/api/prompts/%s/versions", args[0]), nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	data, ok := response["data"].([]interface{})
	if !ok || len(data) == 0 {
		return []OutputLine{{Text: "No version history", Style: helpDescStyle}}
	}

	lines := []OutputLine{{Text: fmt.Sprintf("History for prompt %s:", args[0]), Style: titleStyle}}
	for _, v := range data {
		ver, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		verNum, _ := ver["version_number"].(float64)
		createdAt, _ := ver["created_at"].(string)
		lines = append(lines, OutputLine{
			Text:  fmt.Sprintf("  v%.0f  %s", verNum, createdAt),
			Style: helpDescStyle,
		})
	}

	return lines
}

func (m *Model) cmdDiff(args []string) []OutputLine {
	return []OutputLine{{Text: "diff: not yet implemented", Style: helpDescStyle}}
}

// ============ TAGS ============

func (m *Model) cmdTag(args []string) []OutputLine {
	if len(args) < 2 {
		return []OutputLine{{Text: "Usage: tag <id> <tagname>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	promptID := args[0]
	tagName := strings.Join(args[1:], " ")

	// This would need an API endpoint to add tag to prompt
	_, err := m.client.Request("POST", fmt.Sprintf("/api/prompts/%s/tags", promptID), map[string]any{
		"name": tagName,
	}, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	return []OutputLine{{Text: fmt.Sprintf("✓ Tagged prompt %s with '%s'", promptID, tagName), Style: successStyle}}
}

func (m *Model) cmdTags() []OutputLine {
	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	response, err := m.client.Request("GET", "/api/tags", nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	data, ok := response["data"].([]interface{})
	if !ok || len(data) == 0 {
		return []OutputLine{{Text: "No tags", Style: helpDescStyle}}
	}

	// Cache for autocomplete
	m.cachedTags = make([]map[string]interface{}, 0)

	lines := []OutputLine{{Text: "🏷️  Tags", Style: titleStyle}}
	for _, t := range data {
		tag, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		m.cachedTags = append(m.cachedTags, tag)

		name, _ := tag["name"].(string)
		lines = append(lines, OutputLine{
			Text:  fmt.Sprintf("  #%s", name),
			Style: helpDescStyle,
		})
	}

	return lines
}

// ============ SYNC & EXPORT ============

func (m *Model) cmdExport() []OutputLine {
	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	response, err := m.client.Request("GET", "/api/export", nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	data, _ := json.MarshalIndent(response, "", "  ")
	filename := fmt.Sprintf("prompts-%s.json", time.Now().Format("2006-01-02"))
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	return []OutputLine{{Text: fmt.Sprintf("✓ Exported to %s", filename), Style: successStyle}}
}

func (m *Model) cmdImport(args []string) []OutputLine {
	if len(args) == 0 {
		return []OutputLine{{Text: "Usage: import <file.json>", Style: errorStyle}}
	}

	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	data, err := os.ReadFile(args[0])
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	var importData map[string]any
	if err := json.Unmarshal(data, &importData); err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	_, err = m.client.Request("POST", "/api/import", importData, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	return []OutputLine{{Text: fmt.Sprintf("✓ Imported from %s", args[0]), Style: successStyle}}
}

func (m *Model) cmdSync() []OutputLine {
	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	_, err := m.client.Request("POST", "/api/sync", map[string]any{}, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	return []OutputLine{{Text: "✓ Synced with server", Style: successStyle}}
}

// ============ AUTH ============

func (m *Model) cmdLogin(args []string) []OutputLine {
	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	if len(args) < 2 {
		return []OutputLine{
			{Text: "Usage: login <email> <password>", Style: helpDescStyle},
		}
	}

	response, err := m.client.Request("POST", "/api/login", map[string]any{
		"email":    args[0],
		"password": args[1],
	}, false)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	token, _ := response["token"].(string)
	if token == "" {
		if data, ok := response["data"].(map[string]any); ok {
			token, _ = data["token"].(string)
		}
	}
	if token == "" {
		if access, ok := response["access_token"].(string); ok {
			token = access
		}
	}

	if token == "" {
		return []OutputLine{{Text: "Login failed: no token", Style: errorStyle}}
	}

	m.config.Token = token
	config.Save(m.config)
	m.client.SetToken(token)

	return []OutputLine{{Text: "✓ Logged in!", Style: successStyle}}
}

func (m *Model) cmdLogout() []OutputLine {
	m.config.Token = ""
	config.Save(m.config)
	m.client.SetToken("")
	return []OutputLine{{Text: "✓ Logged out", Style: successStyle}}
}

func (m *Model) cmdRegister(args []string) []OutputLine {
	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	if len(args) < 3 {
		return []OutputLine{
			{Text: "Usage: register <name> <email> <password>", Style: helpDescStyle},
		}
	}

	response, err := m.client.Request("POST", "/api/register", map[string]any{
		"name":                  args[0],
		"email":                 args[1],
		"password":              args[2],
		"password_confirmation": args[2],
	}, false)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	token, _ := response["token"].(string)
	if token == "" {
		if data, ok := response["data"].(map[string]any); ok {
			token, _ = data["token"].(string)
		}
	}

	if token != "" {
		m.config.Token = token
		config.Save(m.config)
		m.client.SetToken(token)
		return []OutputLine{{Text: "✓ Registered and logged in!", Style: successStyle}}
	}

	return []OutputLine{{Text: "✓ Registered! Now use: login <email> <password>", Style: successStyle}}
}

func (m *Model) cmdWhoami() []OutputLine {
	if m.client == nil {
		return []OutputLine{{Text: "Error: not connected", Style: errorStyle}}
	}

	response, err := m.client.Request("GET", "/api/user", nil, true)
	if err != nil {
		return []OutputLine{{Text: fmt.Sprintf("Error: %v", err), Style: errorStyle}}
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return []OutputLine{{Text: "Not logged in", Style: errorStyle}}
	}

	name, _ := data["name"].(string)
	email, _ := data["email"].(string)

	return []OutputLine{
		{Text: fmt.Sprintf("👤 %s", name), Style: titleStyle},
		{Text: fmt.Sprintf("   %s", email), Style: helpDescStyle},
	}
}

// ============ CONFIG ============

func (m *Model) cmdConfig(args []string) []OutputLine {
	if len(args) == 0 {
		return []OutputLine{
			{Text: "Configuration", Style: titleStyle},
			{Text: fmt.Sprintf("  api_base: %s", m.config.APIBase), Style: helpDescStyle},
			{Text: fmt.Sprintf("  logged_in: %v", m.config.Token != ""), Style: helpDescStyle},
		}
	}

	if len(args) >= 3 && args[0] == "set" {
		key := args[1]
		value := strings.Join(args[2:], " ")
		switch key {
		case "api_base", "url":
			m.config.APIBase = value
			config.Save(m.config)
			return []OutputLine{{Text: fmt.Sprintf("✓ api_base = %s", value), Style: successStyle}}
		}
		return []OutputLine{{Text: fmt.Sprintf("Unknown key: %s", key), Style: errorStyle}}
	}

	return []OutputLine{{Text: "Usage: config set <key> <value>", Style: helpDescStyle}}
}
