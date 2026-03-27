package interactive

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true).Foreground(lipgloss.Color("86"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

// PromptItem represents a selectable prompt
type PromptItem struct {
	ID          string
	Title       string
	Summary     string
	IsFavorite  bool
	IsArchived  bool
	CategoryID  string
	CreatedAt   string
}

func (i PromptItem) FilterValue() string { return i.Title }

// PromptDelegate is the list item delegate
type PromptDelegate struct{}

func (d PromptDelegate) Height() int                             { return 2 }
func (d PromptDelegate) Spacing() int                            { return 1 }
func (d PromptDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d PromptDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(PromptItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s", i.Title)
	
	// Add indicators
	indicators := []string{}
	if i.IsFavorite {
		indicators = append(indicators, "⭐")
	}
	if i.IsArchived {
		indicators = append(indicators, "📦")
	}
	if len(indicators) > 0 {
		str = fmt.Sprintf("%s %s", str, strings.Join(indicators, " "))
	}

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("→ " + strings.Join(s, " "))
		}
	}

	// First line: title
	fmt.Fprint(w, fn(str))
	
	// Second line: summary (dimmed)
	if i.Summary != "" {
		summaryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(4)
		if index == m.Index() {
			summaryStyle = summaryStyle.PaddingLeft(4)
		}
		fmt.Fprintf(w, "\n%s", summaryStyle.Render(i.Summary))
	}
}

// PickerModel is the main model for the interactive picker
type PickerModel struct {
	list          list.Model
	choice        *PromptItem
	quitting      bool
	filterInput   textinput.Model
	filtering     bool
	multiSelect   bool
	selected      map[string]bool
	action        string // "show", "delete", "favorite", etc.
}

// NewPickerModel creates a new picker model
func NewPickerModel(items []PromptItem, action string, multiSelect bool) PickerModel {
	// Convert to list items
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	const defaultWidth = 80
	const listHeight = 20

	l := list.New(listItems, PromptDelegate{}, defaultWidth, listHeight)
	l.Title = getTitle(action, multiSelect)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	if multiSelect {
		l.AdditionalShortHelpKeys = func() []key.Binding {
			return []key.Binding{
				key.NewBinding(
					key.WithKeys("space"),
					key.WithHelp("space", "toggle selection"),
				),
			}
		}
	}

	ti := textinput.New()
	ti.Placeholder = "Filter prompts..."
	ti.Focus()

	return PickerModel{
		list:        l,
		filterInput: ti,
		multiSelect: multiSelect,
		selected:    make(map[string]bool),
		action:      action,
	}
}

func getTitle(action string, multiSelect bool) string {
	titles := map[string]string{
		"show":     "📖 Select Prompt to Show",
		"delete":   "🗑️  Select Prompt to Delete",
		"favorite": "⭐ Select Prompt to Favorite",
		"archive":  "📦 Select Prompt to Archive",
		"edit":     "✏️  Select Prompt to Edit",
	}

	title := titles[action]
	if title == "" {
		title = "Select Prompt"
	}

	if multiSelect {
		title += " (Multi-Select)"
	}

	return title
}

func (m PickerModel) Init() tea.Cmd {
	return nil
}

func (m PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.multiSelect {
				// Return all selected items
				m.quitting = true
				return m, tea.Quit
			} else {
				// Return single selected item
				i, ok := m.list.SelectedItem().(PromptItem)
				if ok {
					m.choice = &i
				}
				m.quitting = true
				return m, tea.Quit
			}

		case "space":
			if m.multiSelect {
				i, ok := m.list.SelectedItem().(PromptItem)
				if ok {
					m.selected[i.ID] = !m.selected[i.ID]
				}
			}

		case "esc":
			m.quitting = true
			return m, tea.Quit

		case "/":
			m.filtering = true
			m.filterInput.Focus()
			return m, textinput.Blink
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m PickerModel) View() string {
	if m.quitting {
		return ""
	}

	view := "\n" + m.list.View()

	if m.multiSelect && len(m.selected) > 0 {
		selectedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			PaddingLeft(4).
			PaddingTop(1)
		
		selectedIDs := []string{}
		for id, selected := range m.selected {
			if selected {
				selectedIDs = append(selectedIDs, id)
			}
		}
		view += "\n" + selectedStyle.Render(fmt.Sprintf("Selected: %d items", len(selectedIDs)))
	}

	return view
}

// GetChoice returns the selected prompt (single select mode)
func (m PickerModel) GetChoice() *PromptItem {
	return m.choice
}

// GetSelected returns all selected prompts (multi-select mode)
func (m PickerModel) GetSelected() []string {
	selected := []string{}
	for id, isSelected := range m.selected {
		if isSelected {
			selected = append(selected, id)
		}
	}
	return selected
}

// RunPicker runs the interactive picker and returns the selection
func RunPicker(items []PromptItem, action string, multiSelect bool) ([]string, error) {
	p := tea.NewProgram(NewPickerModel(items, action, multiSelect))
	
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	m := finalModel.(PickerModel)
	
	if multiSelect {
		return m.GetSelected(), nil
	} else {
		if m.GetChoice() != nil {
			return []string{m.GetChoice().ID}, nil
		}
		return []string{}, nil
	}
}
