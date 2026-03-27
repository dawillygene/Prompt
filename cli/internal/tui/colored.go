package tui

import (
	"bufio"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/dawillygene/my-prompt-repository/internal/api"
	"github.com/dawillygene/my-prompt-repository/internal/config"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// ColoredUI provides a beautiful interactive TUI with colors using Lipgloss
type ColoredUI struct {
	client *api.Client
	config config.Config
	reader *bufio.Reader

	currentCategoryID   int
	currentCategoryName string
	listPage            int
	listPerPage         int
}

// NewColoredUI creates a new colored interactive UI
func NewColoredUI(cfg config.Config, client *api.Client) *ColoredUI {
	return &ColoredUI{
		client: client,
		config: cfg,
		reader: bufio.NewReader(os.Stdin),
		listPage: 1,
		listPerPage: 20,
	}
}

// Start begins the interactive session
func (ui *ColoredUI) Start() error {
	ui.printWelcome()

	for {
		ui.printPrompt()
		input, err := ui.reader.ReadString('\n')
		if err != nil {
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if !ui.handleCommand(input) {
			break
		}
	}

	ui.printGoodbye()
	return nil
}

// Color definitions
var (
	primaryColor   = lipgloss.Color("#7B68EE")   // Purple
	secondaryColor = lipgloss.Color("#20B2AA")   // Teal
	accentColor    = lipgloss.Color("#FF6B6B")   // Red
	successColor   = lipgloss.Color("#51CF66")   // Green
	mutedColor     = lipgloss.Color("#868E96")   // Gray
)

func (ui *ColoredUI) printWelcome() {
	title := lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("✨ PROMPT REPOSITORY ✨")
	userInfo := lipgloss.NewStyle().Foreground(secondaryColor).Render(fmt.Sprintf("🔓 Logged in • %s", ui.config.APIBase))
	divider := lipgloss.NewStyle().Foreground(mutedColor).Render(strings.Repeat("━", 60))

	fmt.Println("\n" + divider)
	fmt.Println(title)
	fmt.Println(userInfo)
	fmt.Println(divider + "\n")

	ui.printMenu()
}

func (ui *ColoredUI) printMenu() {
	menuItems := []struct {
		cmd  string
		name string
		icon string
	}{
		{"register", "Create new account", "🆕"},
		{"login", "Login to account", "🔑"},
		{"logout", "Logout", "🚪"},
		{"list", "List all prompts", "📚"},
		{"ls", "Alias for list", "📚"},
		{"pwd", "Show current path", "🧭"},
		{"cd <name|id|..|/>", "Navigate categories", "📁"},
		{"mkdir <name>", "Create category", "📂"},
		{"tree", "Show category tree", "🌳"},
		{"next | prev", "Navigate list pages", "📄"},
		{"add", "Add new prompt", "➕"},
		{"show <id>", "Show prompt details", "📖"},
		{"cat <id>", "Print prompt content", "📄"},
		{"edit <id>", "Edit prompt in $EDITOR", "✍️"},
		{"mv <id> <category|/>", "Move prompt to category", "📦"},
		{"copy <id>", "Copy prompt content", "📋"},
		{"delete <id>", "Delete a prompt", "🗑️"},
		{"trash", "List deleted prompts", "🗑️"},
		{"restore <id>", "Restore deleted prompt", "♻️"},
		{"purge <id>", "Permanent delete from trash", "🔥"},
		{"favorite <id>", "Toggle favorite", "⭐"},
		{"archive <id>", "Toggle archive", "📦"},
		{"search <keyword>", "Search prompts", "🔍"},
		{"clear", "Clear screen", "🧹"},
		{"categories", "List categories", "📁"},
		{"renamecat <id> <name>", "Rename category", "✏️"},
		{"rmdir <id>", "Delete category", "🗑️"},
		{"tags", "List tags", "🏷️"},
		{"tagadd <name>", "Create tag", "🏷️"},
		{"tagrename <id> <name>", "Rename tag", "✏️"},
		{"tagrm <id>", "Delete tag", "🗑️"},
		{"whoami", "Show user info", "👤"},
		{"help", "Show this menu", "❓"},
		{"exit", "Quit", "🚪"},
	}

	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("Available Commands:\n"))
	for _, item := range menuItems {
		cmdStyle := lipgloss.NewStyle().Bold(true).Background(secondaryColor).Foreground(lipgloss.Color("#FFF")).Padding(0, 1)
		descStyle := lipgloss.NewStyle().Foreground(mutedColor)
		fmt.Printf("%s %s  %s  %s\n", item.icon, cmdStyle.Render(item.cmd), descStyle.Render("•"), item.name)
	}
	fmt.Println()
}

func (ui *ColoredUI) printPrompt() {
	promptStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	pathStyle := lipgloss.NewStyle().Foreground(mutedColor)
	fmt.Print(pathStyle.Render(ui.currentPath()) + " ")
	fmt.Print(promptStyle.Render("(prompt)> "))
}

func (ui *ColoredUI) handleCommand(input string) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return true
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "exit", "quit":
		return false
	case "help":
		ui.printMenu()
	case "clear", "cls":
		ui.clearScreen()
	case "pwd":
		ui.printInfo(ui.currentPath())
	case "cd":
		if len(args) == 0 {
			ui.handleCD("/")
		} else {
			ui.handleCD(strings.Join(args, " "))
		}
	case "mkdir":
		if len(args) == 0 {
			ui.printError("Usage: mkdir <category-name>")
		} else {
			ui.handleMkdir(strings.Join(args, " "))
		}
	case "tree":
		ui.handleTree()
	case "register":
		ui.handleRegister()
	case "login":
		ui.handleLogin()
	case "logout":
		ui.handleLogout()
	case "list", "ls":
		ui.handleList(args)
	case "next":
		ui.listPage++
		ui.handleList(nil)
	case "prev":
		if ui.listPage > 1 {
			ui.listPage--
		}
		ui.handleList(nil)
	case "add":
		ui.handleAdd()
	case "show":
		if len(args) > 0 {
			ui.handleShow(args[0])
		} else {
			ui.printError("Usage: show <id-or-slug>")
		}
	case "cat":
		if len(args) > 0 {
			ui.handleCat(args[0])
		} else {
			ui.printError("Usage: cat <id-or-slug>")
		}
	case "edit":
		if len(args) > 0 {
			ui.handleEdit(args[0])
		} else {
			ui.printError("Usage: edit <id-or-slug>")
		}
	case "mv":
		if len(args) > 1 {
			ui.handleMove(args[0], strings.Join(args[1:], " "))
		} else {
			ui.printError("Usage: mv <id-or-slug> <category|/>")
		}
	case "copy":
		if len(args) > 0 {
			ui.handleCopy(args[0])
		} else {
			ui.printError("Usage: copy <id-or-slug>")
		}
	case "delete":
		if len(args) > 0 {
			ui.handleDelete(args[0])
		} else {
			ui.printError("Usage: delete <id-or-slug>")
		}
	case "trash":
		ui.handleTrash()
	case "restore":
		if len(args) > 0 {
			ui.handleRestore(args[0])
		} else {
			ui.printError("Usage: restore <id-or-slug>")
		}
	case "purge":
		if len(args) > 0 {
			ui.handlePurge(args[0])
		} else {
			ui.printError("Usage: purge <id-or-slug>")
		}
	case "favorite":
		if len(args) > 0 {
			ui.handleToggle(args[0], "/api/prompts/%s/favorite")
		} else {
			ui.printError("Usage: favorite <id-or-slug>")
		}
	case "archive":
		if len(args) > 0 {
			ui.handleToggle(args[0], "/api/prompts/%s/archive")
		} else {
			ui.printError("Usage: archive <id-or-slug>")
		}
	case "search":
		if len(args) > 0 {
			ui.handleSearch(strings.Join(args, " "))
		} else {
			ui.printError("Usage: search <keyword>")
		}
	case "categories":
		ui.listCategories()
	case "renamecat":
		if len(args) < 2 {
			ui.printError("Usage: renamecat <id-or-slug> <new-name>")
		} else {
			ui.handleRenameCategory(args[0], strings.Join(args[1:], " "))
		}
	case "rmdir":
		if len(args) < 1 {
			ui.printError("Usage: rmdir <id-or-slug>")
		} else {
			ui.handleDeleteCategory(args[0])
		}
	case "tags":
		ui.handleListTags()
	case "tagadd":
		if len(args) < 1 {
			ui.printError("Usage: tagadd <name>")
		} else {
			ui.handleCreateTag(strings.Join(args, " "))
		}
	case "tagrename":
		if len(args) < 2 {
			ui.printError("Usage: tagrename <id-or-slug> <new-name>")
		} else {
			ui.handleRenameTag(args[0], strings.Join(args[1:], " "))
		}
	case "tagrm":
		if len(args) < 1 {
			ui.printError("Usage: tagrm <id-or-slug>")
		} else {
			ui.handleDeleteTag(args[0])
		}
	case "whoami":
		ui.handleWhoami()
	default:
		ui.printError(fmt.Sprintf("Unknown command: %s (type 'help' for menu)", cmd))
	}

	return true
}

func (ui *ColoredUI) handleList(args []string) {
	ui.applyListOptions(args)

	path := "/api/prompts"
	queryParts := make([]string, 0, 3)
	if ui.currentCategoryID > 0 {
		queryParts = append(queryParts, "category_id="+strconv.Itoa(ui.currentCategoryID))
	}

	queryParts = append(queryParts, "page="+strconv.Itoa(ui.listPage))
	queryParts = append(queryParts, "per_page="+strconv.Itoa(ui.listPerPage))

	if len(queryParts) > 0 {
		path = path + "?" + strings.Join(queryParts, "&")
	}

	response, err := ui.client.Request("GET", path, nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	if ui.currentCategoryID == 0 {
		ui.listCategories()
	}

	data, ok := response["data"].([]interface{})
	if !ok || len(data) == 0 {
		if ui.listPage > 1 {
			ui.listPage--
		}
		if ui.currentCategoryID > 0 {
			ui.printInfo("No prompts found in this category")
		} else {
			ui.printInfo("No prompts found")
		}
		return
	}

	if meta, ok := response["meta"].(map[string]interface{}); ok {
		current := fmt.Sprintf("%v", meta["current_page"])
		last := fmt.Sprintf("%v", meta["last_page"])
		total := fmt.Sprintf("%v", meta["total"])
		ui.printInfo(fmt.Sprintf("Page %s/%s • Total %s", current, last, total))
	}

	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("\n📚 Your Prompts:\n"))
	for i, item := range data {
		p := item.(map[string]interface{})
		id := fmt.Sprintf("%.0f", p["id"].(float64))
		title := p["title"].(string)
		content := p["content"].(string)
		if len(content) > 40 {
			content = content[:40] + "..."
		}

		itemStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(1).
			Foreground(lipgloss.Color("#FFF"))

		if i == 0 {
			itemStyle = itemStyle.BorderForeground(primaryColor)
		}

		fmt.Println(itemStyle.Render(fmt.Sprintf("[%s] %s\n%s", id, title, content)))
	}
	fmt.Println()
}

func (ui *ColoredUI) handleAdd() {
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("\n➕ Add New Prompt\n"))

	fmt.Print(lipgloss.NewStyle().Foreground(secondaryColor).Render("Title: "))
	title, _ := ui.reader.ReadString('\n')
	title = strings.TrimSpace(title)

	fmt.Println(lipgloss.NewStyle().Foreground(secondaryColor).Render("Content (finish with a single '.' line):"))
	content := ui.readMultilineContent()

	fmt.Print(lipgloss.NewStyle().Foreground(secondaryColor).Render("Summary (optional): "))
	summary, _ := ui.reader.ReadString('\n')
	summary = strings.TrimSpace(summary)

	fmt.Print(lipgloss.NewStyle().Foreground(secondaryColor).Render("Visibility (private/public) [private]: "))
	visibility, _ := ui.reader.ReadString('\n')
	visibility = strings.TrimSpace(visibility)
	if visibility == "" {
		visibility = "private"
	}

	response, err := ui.client.Request("POST", "/api/prompts", map[string]interface{}{
		"title":      title,
		"content":    content,
		"summary":    summary,
		"visibility": visibility,
		"category_id": func() any {
			if ui.currentCategoryID > 0 {
				return ui.currentCategoryID
			}
			return nil
		}(),
	}, true)

	if err != nil {
		ui.printError(err.Error())
		return
	}

	if msg, ok := response["message"].(string); ok {
		ui.printSuccess(msg)
	} else {
		ui.printSuccess("Prompt created successfully!")
	}
}

func (ui *ColoredUI) readMultilineContent() string {
	lines := make([]string, 0, 8)
	for {
		fmt.Print(lipgloss.NewStyle().Foreground(mutedColor).Render("... "))
		line, err := ui.reader.ReadString('\n')
		if err != nil {
			break
		}

		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "." {
			break
		}
		lines = append(lines, trimmed)
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func (ui *ColoredUI) handleShow(id string) {
	response, err := ui.client.Request("GET", "/api/prompts/"+id, nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		ui.printError("Invalid response")
		return
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	fmt.Println("\n" + titleStyle.Render(fmt.Sprintf("📖 %v", data["title"])))

	metaStyle := lipgloss.NewStyle().Foreground(secondaryColor)
	fmt.Println(metaStyle.Render(fmt.Sprintf("ID: %v • Updated: %v • Visibility: %v\n", data["id"], data["updated_at"], data["visibility"])))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(1).
		Width(ui.contentBoxWidth())
	fmt.Println(boxStyle.Render(fmt.Sprintf("%v", data["content"])) + "\n")
}

func (ui *ColoredUI) handleCat(id string) {
	response, err := ui.client.Request("GET", "/api/prompts/"+id, nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		ui.printError("Invalid response")
		return
	}

	content, _ := data["content"].(string)
	fmt.Println(content)
}

func (ui *ColoredUI) handleEdit(id string) {
	response, err := ui.client.Request("GET", "/api/prompts/"+id, nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		ui.printError("Invalid response")
		return
	}

	content, _ := data["content"].(string)
	if strings.TrimSpace(content) == "" {
		content = "# Write prompt content here\n"
	}

	tempFile, err := os.CreateTemp("", "myprompt-*.md")
	if err != nil {
		ui.printError("Could not create temp file")
		return
	}
	tempPath := tempFile.Name()
	_ = tempFile.Close()
	defer func() { _ = os.Remove(tempPath) }()

	if err := os.WriteFile(tempPath, []byte(content), 0o600); err != nil {
		ui.printError("Could not write temp content")
		return
	}

	editor := os.Getenv("EDITOR")
	if strings.TrimSpace(editor) == "" {
		editor = "nano"
		if _, err := exec.LookPath(editor); err != nil {
			editor = "vi"
		}
	}

	cmd := exec.Command(editor, tempPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		ui.printError("Editor closed with error")
		return
	}

	updated, err := os.ReadFile(tempPath)
	if err != nil {
		ui.printError("Could not read edited content")
		return
	}

	trimmed := strings.TrimSpace(string(updated))
	if trimmed == "" {
		ui.printError("Refusing to save empty content")
		return
	}

	_, err = ui.client.Request("PUT", "/api/prompts/"+id, map[string]interface{}{
		"content": trimmed,
	}, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	ui.printSuccess("Prompt updated from editor")
}

func (ui *ColoredUI) handleMove(id string, categoryTarget string) {
	catID, catName, ok := ui.resolveCategoryTarget(categoryTarget)
	if !ok {
		return
	}

	var categoryValue any
	if catID > 0 {
		categoryValue = catID
	}

	_, err := ui.client.Request("PUT", "/api/prompts/"+id, map[string]interface{}{
		"category_id": categoryValue,
	}, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	if catID == 0 {
		ui.printSuccess("Prompt moved to /prompts")
		return
	}

	ui.printSuccess("Prompt moved to /prompts/" + catName)
}

func (ui *ColoredUI) handleDelete(id string) {
	fmt.Print(lipgloss.NewStyle().Foreground(accentColor).Render(fmt.Sprintf("Delete prompt %s? (y/n): ", id)))
	confirm, _ := ui.reader.ReadString('\n')
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(confirm)), "y") {
		ui.printInfo("Cancelled")
		return
	}

	_, err := ui.client.Request("DELETE", "/api/prompts/"+id, nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	ui.printSuccess("Prompt deleted!")
}

func (ui *ColoredUI) handleTrash() {
	response, err := ui.client.Request("GET", "/api/prompts/trash", nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	data, ok := response["data"].([]interface{})
	if !ok || len(data) == 0 {
		ui.printInfo("Trash is empty")
		return
	}

	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("\n🗑️ Trash:\n"))
	for _, item := range data {
		p, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		id := fmt.Sprintf("%.0f", p["id"].(float64))
		title := fmt.Sprintf("%v", p["title"])
		deletedAt := fmt.Sprintf("%v", p["deleted_at"])
		fmt.Printf("  [%s] %s (deleted: %s)\n", id, title, deletedAt)
	}
	fmt.Println()
}

func (ui *ColoredUI) handleRestore(id string) {
	_, err := ui.client.Request("POST", "/api/prompts/"+id+"/restore", map[string]interface{}{}, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	ui.printSuccess("Prompt restored")
}

func (ui *ColoredUI) handlePurge(id string) {
	fmt.Print(lipgloss.NewStyle().Foreground(accentColor).Render(fmt.Sprintf("Permanently delete prompt %s? (y/n): ", id)))
	confirm, _ := ui.reader.ReadString('\n')
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(confirm)), "y") {
		ui.printInfo("Cancelled")
		return
	}

	_, err := ui.client.Request("DELETE", "/api/prompts/"+id+"/force", nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	ui.printSuccess("Prompt permanently deleted")
}

func (ui *ColoredUI) handleCopy(id string) {
	response, err := ui.client.Request("GET", "/api/prompts/"+id, nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		ui.printError("Invalid response")
		return
	}

	content, ok := data["content"].(string)
	if !ok || strings.TrimSpace(content) == "" {
		ui.printError("Prompt content is empty")
		return
	}

	if ui.writeClipboard(content) {
		ui.printSuccess("Prompt content copied to clipboard")
		return
	}

	ui.printInfo("No clipboard tool found (tried wl-copy/xclip/xsel).")
	ui.printInfo("Showing content below so you can copy manually:")
	fmt.Println(content)
}

func (ui *ColoredUI) writeClipboard(content string) bool {
	commands := []struct {
		name string
		args []string
	}{
		{name: "wl-copy", args: []string{}},
		{name: "xclip", args: []string{"-selection", "clipboard"}},
		{name: "xsel", args: []string{"--clipboard", "--input"}},
	}

	for _, c := range commands {
		if _, err := exec.LookPath(c.name); err != nil {
			continue
		}

		cmd := exec.Command(c.name, c.args...)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			continue
		}

		if err := cmd.Start(); err != nil {
			_ = stdin.Close()
			continue
		}

		_, _ = stdin.Write([]byte(content))
		_ = stdin.Close()
		if err := cmd.Wait(); err == nil {
			return true
		}
	}

	return false
}

func (ui *ColoredUI) handleToggle(id, pattern string) {
	endpoint := fmt.Sprintf(pattern, id)
	_, err := ui.client.Request("POST", endpoint, map[string]interface{}{}, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	ui.printSuccess("Updated!")
}

func (ui *ColoredUI) handleSearch(keyword string) {
	response, err := ui.client.Request("GET", "/api/prompts?search="+keyword, nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	data, ok := response["data"].([]interface{})
	if !ok || len(data) == 0 {
		ui.printInfo(fmt.Sprintf("No results for '%s'", keyword))
		return
	}

	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render(fmt.Sprintf("\n🔍 Search Results for '%s':\n", keyword)))
	for _, item := range data {
		p := item.(map[string]interface{})
		fmt.Printf("%s %s\n", lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("[%.0f]", p["id"])), p["title"])
	}
	fmt.Println()
}

func (ui *ColoredUI) applyListOptions(args []string) {
	if len(args) == 0 {
		return
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--page":
			if i+1 < len(args) {
				if parsed, err := strconv.Atoi(args[i+1]); err == nil && parsed > 0 {
					ui.listPage = parsed
				}
				i++
			}
		case "--per-page", "--per_page":
			if i+1 < len(args) {
				if parsed, err := strconv.Atoi(args[i+1]); err == nil && parsed > 0 {
					if parsed > 100 {
						parsed = 100
					}
					ui.listPerPage = parsed
				}
				i++
			}
		}
	}
}

func (ui *ColoredUI) handleRegister() {
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("\n🆕 Create New Account\n"))

	fmt.Print(lipgloss.NewStyle().Foreground(secondaryColor).Render("Name: "))
	name, _ := ui.reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print(lipgloss.NewStyle().Foreground(secondaryColor).Render("Email: "))
	email, _ := ui.reader.ReadString('\n')
	email = strings.TrimSpace(email)

	password := ui.readPassword("Password: ")

	response, err := ui.client.Request("POST", "/api/register", map[string]interface{}{
		"name":     name,
		"email":    email,
		"password": password,
	}, false)

	if err != nil {
		ui.printError(err.Error())
		return
	}

	if token, ok := response["token"].(string); ok {
		ui.config.Token = token
		if err := config.Save(ui.config); err != nil {
			ui.printError("Could not save auth token locally")
		}
		ui.client = api.New(ui.config)
		ui.printSuccess(fmt.Sprintf("Welcome %s! You're now logged in.", name))
	} else {
		ui.printError("Failed to get authentication token")
	}
}

func (ui *ColoredUI) handleLogin() {
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("\n🔑 Login to Account\n"))

	fmt.Print(lipgloss.NewStyle().Foreground(secondaryColor).Render("Email: "))
	email, _ := ui.reader.ReadString('\n')
	email = strings.TrimSpace(email)

	password := ui.readPassword("Password: ")

	response, err := ui.client.Request("POST", "/api/login", map[string]interface{}{
		"email":    email,
		"password": password,
	}, false)

	if err != nil {
		ui.printError(err.Error())
		return
	}

	if token, ok := response["token"].(string); ok {
		ui.config.Token = token
		if err := config.Save(ui.config); err != nil {
			ui.printError("Could not save auth token locally")
		}
		ui.client = api.New(ui.config)
		user, _ := response["user"].(map[string]interface{})
		name := "User"
		if u, ok := user["name"]; ok {
			name = fmt.Sprintf("%v", u)
		}
		ui.printSuccess(fmt.Sprintf("Welcome back %s!", name))
	} else {
		ui.printError("Failed to authenticate")
	}
}

func (ui *ColoredUI) handleLogout() {
	_, _ = ui.client.Request("POST", "/api/logout", map[string]interface{}{}, true)
	ui.config.Token = ""
	if err := config.Save(ui.config); err != nil {
		ui.printError("Could not clear saved auth token")
	}
	ui.client = api.New(ui.config)
	ui.printSuccess("You've been logged out")
}

func (ui *ColoredUI) clearScreen() {
	fmt.Print("\033[H\033[2J")
	ui.printWelcome()
}

func (ui *ColoredUI) readPassword(label string) string {
	fmt.Print(lipgloss.NewStyle().Foreground(secondaryColor).Render(label))
	_ = exec.Command("stty", "-echo").Run()
	password, err := ui.reader.ReadString('\n')
	_ = exec.Command("stty", "echo").Run()
	fmt.Println()
	if err != nil {
		ui.printError("Could not read password")
		return ""
	}

	return strings.TrimSpace(password)
}

func (ui *ColoredUI) contentBoxWidth() int {
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if parsed, err := strconv.Atoi(cols); err == nil {
			if parsed > 20 {
				w := parsed - 8
				if w > 110 {
					return 110
				}
				if w < 50 {
					return 50
				}
				return w
			}
		}
	}

	return 90
}

func (ui *ColoredUI) handleWhoami() {
	response, err := ui.client.Request("GET", "/api/me", nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	user, ok := response["user"].(map[string]interface{})
	if !ok {
		ui.printError("Invalid response")
		return
	}

	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("\n👤 User Info:\n"))
	infoStyle := lipgloss.NewStyle().Foreground(secondaryColor)
	fmt.Printf("%s: %v\n", infoStyle.Render("Name"), user["name"])
	fmt.Printf("%s: %v\n", infoStyle.Render("Email"), user["email"])
	fmt.Printf("%s: %v\n", infoStyle.Render("ID"), user["id"])
	fmt.Printf("%s: %v\n\n", infoStyle.Render("Member Since"), user["created_at"])
}

func (ui *ColoredUI) currentPath() string {
	if ui.currentCategoryID <= 0 {
		return "/prompts"
	}
	return "/prompts/" + ui.currentCategoryName
}

func (ui *ColoredUI) handleCD(target string) {
	target = strings.TrimSpace(target)
	catID, catSlug, ok := ui.resolveCategoryTarget(target)
	if !ok {
		return
	}

	if catID == 0 {
		ui.currentCategoryID = 0
		ui.currentCategoryName = ""
		ui.printInfo("Moved to /prompts")
		return
	}

	ui.currentCategoryID = catID
	ui.currentCategoryName = catSlug
	ui.printSuccess("Moved to " + ui.currentPath())
}

func (ui *ColoredUI) resolveCategoryTarget(target string) (int, string, bool) {
	target = strings.TrimSpace(target)
	switch target {
	case "", "/", "~", "..":
		return 0, "", true
	}

	response, err := ui.client.Request("GET", "/api/categories", nil, true)
	if err != nil {
		ui.printError(err.Error())
		return 0, "", false
	}

	data, ok := response["data"].([]interface{})
	if !ok {
		ui.printError("Invalid categories response")
		return 0, "", false
	}

	needle := strings.ToLower(target)
	for _, item := range data {
		cat, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		idStr := fmt.Sprintf("%.0f", cat["id"])
		name := fmt.Sprintf("%v", cat["name"])
		slug := fmt.Sprintf("%v", cat["slug"])

		if needle == strings.ToLower(name) || needle == strings.ToLower(slug) || needle == idStr {
			id, _ := strconv.Atoi(idStr)
			return id, slug, true
		}
	}

	ui.printError("Category not found: " + target)
	return 0, "", false
}

func (ui *ColoredUI) handleMkdir(name string) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		ui.printError("Category name cannot be empty")
		return
	}

	// Ask if user wants to add a description
	fmt.Print(lipgloss.NewStyle().Foreground(accentColor).Render("Add description? (y/n): "))
	response, _ := ui.reader.ReadString('\n')
	var description string
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(response)), "y") {
		fmt.Print(lipgloss.NewStyle().Foreground(accentColor).Render("Description: "))
		desc, _ := ui.reader.ReadString('\n')
		description = strings.TrimSpace(desc)
	}

	payload := map[string]interface{}{
		"name": trimmed,
	}
	if description != "" {
		payload["description"] = description
	}

	res, err := ui.client.Request("POST", "/api/categories", payload, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	if msg, ok := res["message"].(string); ok {
		ui.printSuccess(msg)
	} else {
		ui.printSuccess("Category created")
	}
}

func (ui *ColoredUI) handleRenameCategory(ref string, newName string) {
	trimmed := strings.TrimSpace(newName)
	if trimmed == "" {
		ui.printError("Category name cannot be empty")
		return
	}

	// Ask if user wants to update description
	fmt.Print(lipgloss.NewStyle().Foreground(accentColor).Render("Update description? (y/n): "))
	response, _ := ui.reader.ReadString('\n')

	payload := map[string]interface{}{
		"name": trimmed,
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(response)), "y") {
		fmt.Print(lipgloss.NewStyle().Foreground(accentColor).Render("New description (or leave empty to remove): "))
		desc, _ := ui.reader.ReadString('\n')
		description := strings.TrimSpace(desc)
		payload["description"] = description
	}

	_, err := ui.client.Request("PUT", "/api/categories/"+ref, payload, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	ui.printSuccess("Category renamed")
}

func (ui *ColoredUI) handleDeleteCategory(ref string) {
	// First, check if category has prompts
	response, err := ui.client.Request("GET", "/api/prompts?category_id="+ref, nil, true)
	if err != nil {
		ui.printError("Could not check category prompts: " + err.Error())
		return
	}

	var promptCount int
	if data, ok := response["data"].([]interface{}); ok {
		promptCount = len(data)
	}

	if promptCount > 0 {
		msg := fmt.Sprintf("This category has %d prompt(s). Deleting it will unassign them. Continue? (y/n): ", promptCount)
		fmt.Print(lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render(msg))
	} else {
		fmt.Print(lipgloss.NewStyle().Foreground(accentColor).Render(fmt.Sprintf("Delete category %s? (y/n): ", ref)))
	}

	confirm, _ := ui.reader.ReadString('\n')
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(confirm)), "y") {
		ui.printInfo("Cancelled")
		return
	}

	_, err = ui.client.Request("DELETE", "/api/categories/"+ref, nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	ui.printSuccess("Category deleted")
	if ui.currentCategoryName == ref {
		ui.currentCategoryID = 0
		ui.currentCategoryName = ""
	}
}

func (ui *ColoredUI) handleListTags() {
	response, err := ui.client.Request("GET", "/api/tags", nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	data, ok := response["data"].([]interface{})
	if !ok || len(data) == 0 {
		ui.printInfo("No tags found")
		return
	}

	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("\n🏷️ Tags:\n"))
	for _, item := range data {
		tag, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		fmt.Printf("  [%.0f] %v (%v)\n", tag["id"], tag["name"], tag["slug"])
	}
	fmt.Println()
}

func (ui *ColoredUI) handleCreateTag(name string) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		ui.printError("Tag name cannot be empty")
		return
	}

	// Ask if user wants to add a description
	fmt.Print(lipgloss.NewStyle().Foreground(accentColor).Render("Add description? (y/n): "))
	response, _ := ui.reader.ReadString('\n')
	var description string
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(response)), "y") {
		fmt.Print(lipgloss.NewStyle().Foreground(accentColor).Render("Description: "))
		desc, _ := ui.reader.ReadString('\n')
		description = strings.TrimSpace(desc)
	}

	payload := map[string]interface{}{
		"name": trimmed,
	}
	if description != "" {
		payload["description"] = description
	}

	_, err := ui.client.Request("POST", "/api/tags", payload, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	ui.printSuccess("Tag created")
}

func (ui *ColoredUI) handleRenameTag(ref string, newName string) {
	trimmed := strings.TrimSpace(newName)
	if trimmed == "" {
		ui.printError("Tag name cannot be empty")
		return
	}

	// Ask if user wants to update description
	fmt.Print(lipgloss.NewStyle().Foreground(accentColor).Render("Update description? (y/n): "))
	response, _ := ui.reader.ReadString('\n')

	payload := map[string]interface{}{
		"name": trimmed,
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(response)), "y") {
		fmt.Print(lipgloss.NewStyle().Foreground(accentColor).Render("New description (or leave empty to remove): "))
		desc, _ := ui.reader.ReadString('\n')
		description := strings.TrimSpace(desc)
		payload["description"] = description
	}

	_, err := ui.client.Request("PUT", "/api/tags/"+ref, payload, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	ui.printSuccess("Tag renamed")
}

func (ui *ColoredUI) handleDeleteTag(ref string) {
	fmt.Print(lipgloss.NewStyle().Foreground(accentColor).Render(fmt.Sprintf("Delete tag %s? (y/n): ", ref)))
	confirm, _ := ui.reader.ReadString('\n')
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(confirm)), "y") {
		ui.printInfo("Cancelled")
		return
	}

	_, err := ui.client.Request("DELETE", "/api/tags/"+ref, nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	ui.printSuccess("Tag deleted")
}

func (ui *ColoredUI) listCategories() {
	response, err := ui.client.Request("GET", "/api/categories", nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	data, ok := response["data"].([]interface{})
	if !ok || len(data) == 0 {
		ui.printInfo("No categories found")
		return
	}

	type catRow struct {
		id   int
		name string
		slug string
	}

	rows := make([]catRow, 0, len(data))
	for _, item := range data {
		cat, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		id, _ := strconv.Atoi(fmt.Sprintf("%.0f", cat["id"]))
		rows = append(rows, catRow{
			id:   id,
			name: fmt.Sprintf("%v", cat["name"]),
			slug: fmt.Sprintf("%v", cat["slug"]),
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].name) < strings.ToLower(rows[j].name)
	})

	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("\n📁 Categories:\n"))
	for _, row := range rows {
		fmt.Printf("  [%d] %s (%s)\n", row.id, row.name, row.slug)
	}
	fmt.Println()
}

func (ui *ColoredUI) handleTree() {
	catResp, err := ui.client.Request("GET", "/api/categories", nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	promptResp, err := ui.client.Request("GET", "/api/prompts", nil, true)
	if err != nil {
		ui.printError(err.Error())
		return
	}

	catData, _ := catResp["data"].([]interface{})
	promptData, _ := promptResp["data"].([]interface{})

	counts := map[int]int{}
	rootCount := 0
	for _, item := range promptData {
		p, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if p["category_id"] == nil {
			rootCount++
			continue
		}

		switch val := p["category_id"].(type) {
		case float64:
			counts[int(val)]++
		case int:
			counts[val]++
		}
	}

	type catRow struct {
		id   int
		name string
		slug string
	}

	rows := make([]catRow, 0, len(catData))
	for _, item := range catData {
		cat, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		id, _ := strconv.Atoi(fmt.Sprintf("%.0f", cat["id"]))
		rows = append(rows, catRow{
			id:   id,
			name: fmt.Sprintf("%v", cat["name"]),
			slug: fmt.Sprintf("%v", cat["slug"]),
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].name) < strings.ToLower(rows[j].name)
	})

	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("\n🌳 Prompt Tree\n"))
	fmt.Printf("/prompts (%d)\n", rootCount)
	for _, row := range rows {
		fmt.Printf("├── %s [%d] (%d)\n", row.slug, row.id, counts[row.id])
	}
	fmt.Println()
}

func (ui *ColoredUI) printSuccess(msg string) {
	fmt.Println(lipgloss.NewStyle().Foreground(successColor).Render("✅ " + msg))
}

func (ui *ColoredUI) printError(msg string) {
	fmt.Println(lipgloss.NewStyle().Foreground(accentColor).Render("❌ " + msg))
}

func (ui *ColoredUI) printInfo(msg string) {
	fmt.Println(lipgloss.NewStyle().Foreground(mutedColor).Render("ℹ️ " + msg))
}

func (ui *ColoredUI) printGoodbye() {
	goodbye := lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("👋 Goodbye!")
	fmt.Println("\n" + goodbye + "\n")
}
