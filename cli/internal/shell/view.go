package shell

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette (GitHub Copilot-style)
var (
	primaryColor   = lipgloss.Color("#58A6FF") // Blue
	secondaryColor = lipgloss.Color("#8B949E") // Gray
	successColor   = lipgloss.Color("#3FB950") // Green
	warningColor   = lipgloss.Color("#D29922") // Yellow
	errorColor     = lipgloss.Color("#F85149") // Red
	accentColor    = lipgloss.Color("#A371F7") // Purple
	mutedColor     = lipgloss.Color("#6E7681") // Dark gray
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

	accentStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor)

	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor)

	suggestionStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	selectedSuggestionStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	inlineHintStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	bannerStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	dividerStyle = lipgloss.NewStyle().
			Foreground(mutedColor)
)

func (m Model) View() string {
	if m.quitting {
		return successStyle.Render("👋 Goodbye!\n\n")
	}

	var b strings.Builder

	// Banner (only on start)
	if m.showBanner && len(m.output) == 0 {
		b.WriteString(m.renderBanner())
	}

	// Output area - show ALL output (terminal handles scrolling)
	for _, line := range m.output {
		b.WriteString(line.Style.Render(line.Text))
		b.WriteString("\n")
	}

	// Input line with inline completion hint
	b.WriteString(m.renderInputLine())

	return b.String()
}

func (m Model) renderBanner() string {
	return bannerStyle.Render("\n  PROMPT CLI") + " " + helpDescStyle.Render("v1.0.0") + "\n" +
		helpDescStyle.Render("  Type 'help' for commands. Use Tab for autocomplete.\n\n")
}

func (m Model) renderInputLine() string {
	var b strings.Builder

	// Different prompts based on input mode
	switch m.inputMode {
	case InputAddTitle:
		b.WriteString(accentStyle.Render("Title: "))
		b.WriteString(m.input.Value())
		b.WriteString("\n")
		return b.String()
	case InputAddContent:
		b.WriteString(accentStyle.Render("Content: "))
		b.WriteString(m.input.Value())
		b.WriteString("\n")
		b.WriteString(helpDescStyle.Render("  (Enter to save, or type more)\n"))
		return b.String()
	case InputAddSummary:
		b.WriteString(accentStyle.Render("Summary (optional, Enter to skip): "))
		b.WriteString(m.input.Value())
		b.WriteString("\n")
		return b.String()
	}

	// Normal mode - prompt with current category path
	path := "~"
	if m.currentCat != "" {
		path = "~/" + m.currentCat
	}
	b.WriteString(promptStyle.Render(path + " ❯ "))

	// Current input
	input := m.input.Value()
	b.WriteString(input)

	// Inline completion hint (ghost text)
	if len(input) > 0 {
		filtered := filterSuggestions(input, m.suggestions)
		if len(filtered) > 0 {
			best := filtered[m.suggestIdx%len(filtered)]
			if len(best.Command) > len(input) {
				hint := best.Command[len(input):]
				b.WriteString(inlineHintStyle.Render(hint))
			}
		}
	}

	b.WriteString("\n")

	// Show dropdown only if multiple suggestions
	if m.showSuggest && len(input) >= 1 && len(input) < 6 {
		filtered := filterSuggestions(input, m.suggestions)
		if len(filtered) > 1 && len(filtered) <= 12 {
			b.WriteString(m.renderSuggestions(filtered))
		}
	}

	return b.String()
}

func (m Model) renderSuggestions(filtered []Suggestion) string {
	var b strings.Builder

	// Show up to 6 suggestions
	maxShow := 6
	if len(filtered) < maxShow {
		maxShow = len(filtered)
	}

	for i := 0; i < maxShow; i++ {
		s := filtered[i]

		if i == m.suggestIdx%len(filtered) {
			b.WriteString("  " + selectedSuggestionStyle.Render("▸ "+s.Command))
			b.WriteString("  " + helpDescStyle.Render(s.Description))
		} else {
			b.WriteString("    " + suggestionStyle.Render(s.Command))
			b.WriteString("  " + suggestionStyle.Render(s.Description))
		}
		b.WriteString("\n")
	}

	if len(filtered) > maxShow {
		b.WriteString(suggestionStyle.Render(fmt.Sprintf("    ... %d more (keep typing)\n", len(filtered)-maxShow)))
	}

	return b.String()
}
