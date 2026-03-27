package shell

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dawillygene/my-prompt-repository/internal/api"
	"github.com/dawillygene/my-prompt-repository/internal/config"
)

// Mode represents the current shell mode
type Mode int

const (
	ModeInteractive Mode = iota
	ModePlan
)

func (m Mode) String() string {
	switch m {
	case ModeInteractive:
		return "interactive"
	case ModePlan:
		return "plan"
	default:
		return "unknown"
	}
}

// InputMode represents the current input mode
type InputMode int

const (
	InputNormal InputMode = iota
	InputAddTitle
	InputAddContent
	InputAddSummary
)

// Model is the main shell model
type Model struct {
	config      config.Config
	client      *api.Client
	input       textinput.Model
	mode        Mode
	inputMode   InputMode // Current input mode (normal, adding title, etc.)
	history     []string
	historyIdx  int
	suggestions []Suggestion
	suggestIdx  int
	showSuggest bool
	output      []OutputLine
	width       int
	height      int
	quitting    bool
	showBanner  bool
	err         error
	// Category navigation
	currentCat   string  // Current category name (empty = root)
	currentCatID float64 // Current category ID (0 = root)
	// Pending category change (from cd command)
	pendingCat   string
	pendingCatID float64
	pendingCd    bool
	// Cached data for autocompletion
	cachedPrompts    []map[string]interface{}
	cachedCategories []map[string]interface{}
	cachedTags       []map[string]interface{}
	// Interactive add state
	newPromptTitle   string
	newPromptContent string
	newPromptSummary string
}

// Suggestion represents a command suggestion
type Suggestion struct {
	Command     string
	Description string
	Category    string
}

// OutputLine represents a line of output
type OutputLine struct {
	Text  string
	Style lipgloss.Style
}

// NewModel creates a new shell model
func NewModel(cfg config.Config, client *api.Client, showBanner bool) Model {
	ti := textinput.New()
	ti.Placeholder = "Type a command or /help for options..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 60
	ti.Prompt = ""

	return Model{
		config:      cfg,
		client:      client,
		input:       ti,
		mode:        ModeInteractive,
		history:     []string{},
		historyIdx:  -1,
		suggestions: getAllSuggestions(),
		suggestIdx:  0,
		showSuggest: false,
		output:      []OutputLine{},
		width:       80,
		height:      24,
		showBanner:  showBanner,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// getAllSuggestions returns all available commands
func getAllSuggestions() []Suggestion {
	return []Suggestion{
		// Navigation (like Linux filesystem)
		{"ls", "List prompts in current category", "Navigation"},
		{"ls -a", "List all including archived", "Navigation"},
		{"ls -l", "List with details", "Navigation"},
		{"cd", "Go to category (cd <name> or cd ..)", "Navigation"},
		{"pwd", "Show current category path", "Navigation"},
		{"tree", "Show all categories & prompts", "Navigation"},
		
		// File operations (prompts = files)
		{"cat", "Show prompt content (cat <id>)", "Prompts"},
		{"touch", "Quick create prompt (touch <title>)", "Prompts"},
		{"add", "Create prompt interactively", "Prompts"},
		{"edit", "Edit in $EDITOR (edit <id>)", "Prompts"},
		{"rm", "Delete prompt (rm <id>)", "Prompts"},
		{"mv", "Move to category (mv <id> <cat>)", "Prompts"},
		{"cp", "Copy prompt (cp <id> <newname>)", "Prompts"},
		{"copy", "Copy content to clipboard (copy <id>)", "Prompts"},
		
		// Directory operations (categories = directories)
		{"mkdir", "Create category (mkdir <name>)", "Categories"},
		{"rmdir", "Delete category (rmdir <name>)", "Categories"},
		
		// Search
		{"find", "Search prompts (find <keyword>)", "Search"},
		{"grep", "Search in content (grep <pattern>)", "Search"},
		
		// Favorites & Archive
		{"star", "Toggle favorite (star <id>)", "Organize"},
		{"archive", "Archive prompt (archive <id>)", "Organize"},
		
		// Version control
		{"history", "Show versions (history <id>)", "Versions"},
		{"diff", "Compare versions (diff <id> <v1> <v2>)", "Versions"},
		
		// Sync & Export
		{"export", "Export prompts to JSON", "Sync"},
		{"import", "Import from JSON file", "Sync"},
		{"sync", "Sync with server", "Sync"},
		
		// Tags
		{"tag", "Manage tags (tag add/rm <id> <tag>)", "Tags"},
		{"tags", "List all tags", "Tags"},
		
		// Auth
		{"login", "Log in (login <email> <pass>)", "Auth"},
		{"logout", "Log out", "Auth"},
		{"register", "Create account", "Auth"},
		{"whoami", "Show current user", "Auth"},
		
		// Session
		{"help", "Show this help", "Help"},
		{"man", "Show help (alias)", "Help"},
		{"clear", "Clear screen", "Session"},
		{"exit", "Exit CLI", "Session"},
		{"q", "Exit CLI (shortcut)", "Session"},
	}
}

// filterSuggestions filters suggestions based on input
func filterSuggestions(input string, all []Suggestion) []Suggestion {
	if input == "" {
		return nil
	}
	
	input = strings.ToLower(input)
	// Strip leading / if present
	if strings.HasPrefix(input, "/") {
		input = strings.TrimPrefix(input, "/")
	}
	
	var filtered []Suggestion
	
	for _, s := range all {
		if strings.HasPrefix(strings.ToLower(s.Command), input) {
			filtered = append(filtered, s)
		}
	}
	
	return filtered
}
