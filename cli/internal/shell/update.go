package shell

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dawillygene/my-prompt-repository/internal/config"
)

// Message types
type (
	commandResultMsg struct {
		output      []OutputLine
		err         error
		startAdd    bool    // Flag to start interactive add
		clearScreen bool    // Flag to clear screen
		newCat      string  // New category name (for cd)
		newCatID    float64 // New category ID (for cd)
		changeCat   bool    // Flag indicating category change
	}
	clearScreenMsg  struct{}
	addCompleteMsg  struct {
		success bool
		id      string
		title   string
		err     error
	}
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = msg.Width - 20
		return m, nil

	case tea.KeyMsg:
		// Handle ESC to cancel interactive add
		if msg.Type == tea.KeyEsc && m.inputMode != InputNormal {
			m.inputMode = InputNormal
			m.newPromptTitle = ""
			m.newPromptContent = ""
			m.newPromptSummary = ""
			m.output = append(m.output, OutputLine{Text: "Cancelled.", Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#F85149"))})
			return m, nil
		}

		// Handle Enter during interactive add
		if msg.Type == tea.KeyEnter && m.inputMode != InputNormal {
			return m.handleAddInput()
		}

		// Handle Ctrl+C
		if msg.Type == tea.KeyCtrlC {
			if m.inputMode != InputNormal {
				m.inputMode = InputNormal
				m.newPromptTitle = ""
				m.newPromptContent = ""
				m.newPromptSummary = ""
				m.output = append(m.output, OutputLine{Text: "Cancelled.", Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#F85149"))})
				m.input.SetValue("")
				return m, nil
			}
			if m.input.Value() != "" {
				m.input.SetValue("")
				m.showSuggest = false
				m.suggestIdx = 0
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		}

		// Normal mode key handling
		if m.inputMode == InputNormal {
			switch msg.Type {
			case tea.KeyCtrlD:
				m.quitting = true
				return m, tea.Quit

			case tea.KeyCtrlL:
				m.output = []OutputLine{}
				return m, nil

			case tea.KeyTab, tea.KeyRight:
				// Autocomplete
				val := m.input.Value()
				if len(val) > 0 {
					filtered := filterSuggestions(val, m.suggestions)
					if len(filtered) > 0 {
						m.input.SetValue(filtered[m.suggestIdx%len(filtered)].Command + " ")
						m.input.CursorEnd()
						m.showSuggest = false
						m.suggestIdx = 0
					}
				}
				return m, nil

			case tea.KeyShiftTab:
				if m.mode == ModeInteractive {
					m.mode = ModePlan
				} else {
					m.mode = ModeInteractive
				}
				return m, nil

			case tea.KeyUp:
				val := m.input.Value()
				if m.showSuggest && len(val) > 0 {
					filtered := filterSuggestions(val, m.suggestions)
					if len(filtered) > 0 {
						if m.suggestIdx > 0 {
							m.suggestIdx--
						} else {
							m.suggestIdx = len(filtered) - 1
						}
					}
					return m, nil
				}
				if len(m.history) > 0 {
					if m.historyIdx < len(m.history)-1 {
						m.historyIdx++
					}
					m.input.SetValue(m.history[len(m.history)-1-m.historyIdx])
					m.input.CursorEnd()
				}
				return m, nil

			case tea.KeyDown:
				val := m.input.Value()
				if m.showSuggest && len(val) > 0 {
					filtered := filterSuggestions(val, m.suggestions)
					if len(filtered) > 0 {
						if m.suggestIdx < len(filtered)-1 {
							m.suggestIdx++
						} else {
							m.suggestIdx = 0
						}
					}
					return m, nil
				}
				if m.historyIdx > 0 {
					m.historyIdx--
					m.input.SetValue(m.history[len(m.history)-1-m.historyIdx])
					m.input.CursorEnd()
				} else if m.historyIdx == 0 {
					m.historyIdx = -1
					m.input.SetValue("")
				}
				return m, nil

			case tea.KeyEnter:
				input := strings.TrimSpace(m.input.Value())
				if input == "" {
					return m, nil
				}

				m.history = append(m.history, input)
				m.historyIdx = -1
				m.input.SetValue("")
				m.showSuggest = false
				m.suggestIdx = 0

				return m, m.executeCommand(input)

			case tea.KeyEsc:
				m.showSuggest = false
				m.suggestIdx = 0
				return m, nil
			}
		}

	case commandResultMsg:
		// Check if we should clear screen
		if msg.clearScreen {
			m.output = []OutputLine{}
			return m, nil
		}
		// Check if we should change category
		if msg.changeCat {
			m.currentCat = msg.newCat
			m.currentCatID = msg.newCatID
		}
		m.output = append(m.output, msg.output...)
		if msg.err != nil {
			m.err = msg.err
		}
		// Check if we should start interactive add
		if msg.startAdd {
			m.inputMode = InputAddTitle
			m.newPromptTitle = ""
			m.newPromptContent = ""
			m.newPromptSummary = ""
		}
		return m, nil

	case addCompleteMsg:
		if msg.err != nil {
			m.output = append(m.output, OutputLine{
				Text:  fmt.Sprintf("Error: %v", msg.err),
				Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#F85149")),
			})
		} else if msg.success {
			m.output = append(m.output, OutputLine{
				Text:  fmt.Sprintf("✓ Created: %s (id: %s)", msg.title, msg.id),
				Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#3FB950")),
			})
		}
		return m, nil

	case clearScreenMsg:
		m.output = []OutputLine{}
		return m, nil
	}

	// Update text input
	prevVal := m.input.Value()
	m.input, cmd = m.input.Update(msg)
	newVal := m.input.Value()

	if prevVal != newVal {
		m.suggestIdx = 0
	}

	// Show suggestions
	if m.inputMode == InputNormal && len(newVal) > 0 {
		filtered := filterSuggestions(newVal, m.suggestions)
		m.showSuggest = len(filtered) > 0
	} else {
		m.showSuggest = false
	}

	return m, cmd
}

// handleAddInput handles input during interactive add
func (m *Model) handleAddInput() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.input.Value())
	m.input.SetValue("")

	switch m.inputMode {
	case InputAddTitle:
		if input == "" {
			m.output = append(m.output, OutputLine{
				Text:  "Title cannot be empty. Try again:",
				Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#F85149")),
			})
			return m, nil
		}
		m.newPromptTitle = input
		m.inputMode = InputAddContent
		m.output = append(m.output, OutputLine{
			Text:  fmt.Sprintf("Title: %s", input),
			Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#8B949E")),
		})
		return m, nil

	case InputAddContent:
		if input == "" {
			m.output = append(m.output, OutputLine{
				Text:  "Content cannot be empty. Try again:",
				Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#F85149")),
			})
			return m, nil
		}
		m.newPromptContent = input
		m.inputMode = InputAddSummary
		m.output = append(m.output, OutputLine{
			Text:  "Content: " + truncate(input, 50),
			Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#8B949E")),
		})
		return m, nil

	case InputAddSummary:
		m.newPromptSummary = input
		m.inputMode = InputNormal

		// Create the prompt
		return m, m.createPrompt()
	}

	return m, nil
}

// createPrompt sends the prompt to the API
func (m *Model) createPrompt() tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return addCompleteMsg{err: fmt.Errorf("not connected")}
		}

		payload := map[string]any{
			"title":   m.newPromptTitle,
			"content": m.newPromptContent,
		}
		if m.newPromptSummary != "" {
			payload["summary"] = m.newPromptSummary
		}
		if m.currentCatID > 0 {
			payload["category_id"] = m.currentCatID
		}

		response, err := m.client.Request("POST", "/api/prompts", payload, true)
		if err != nil {
			return addCompleteMsg{err: err}
		}

		// Clear the temp values
		title := m.newPromptTitle
		m.newPromptTitle = ""
		m.newPromptContent = ""
		m.newPromptSummary = ""

		// Get the ID
		id := "?"
		if data, ok := response["data"].(map[string]interface{}); ok {
			if idFloat, ok := data["id"].(float64); ok {
				id = fmt.Sprintf("%.0f", idFloat)
			}
		}

		// Save config if needed
		config.Save(m.config)

		return addCompleteMsg{success: true, id: id, title: title}
	}
}

// truncate truncates a string to max length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// executeCommand handles command execution
func (m *Model) executeCommand(input string) tea.Cmd {
	return func() tea.Msg {
		output, err := m.runCommand(input)
		
		// Check command flags
		startAdd := false
		clearScreen := false
		changeCat := false
		newCat := ""
		var newCatID float64 = 0
		
		parts := strings.Fields(input)
		if len(parts) > 0 {
			cmd := strings.ToLower(parts[0])
			if cmd == "add" && err == nil {
				startAdd = true
			}
			if cmd == "clear" || cmd == "cls" {
				clearScreen = true
			}
			if cmd == "cd" && m.pendingCd {
				changeCat = true
				newCat = m.pendingCat
				newCatID = m.pendingCatID
				m.pendingCd = false // Reset the flag
			}
		}
		
		return commandResultMsg{
			output:      output,
			err:         err,
			startAdd:    startAdd,
			clearScreen: clearScreen,
			changeCat:   changeCat,
			newCat:      newCat,
			newCatID:    newCatID,
		}
	}
}
