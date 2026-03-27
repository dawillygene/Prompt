package shell

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dawillygene/my-prompt-repository/internal/api"
	"github.com/dawillygene/my-prompt-repository/internal/config"
)

// Start launches the interactive shell
func Start(cfg config.Config, client *api.Client, showBanner bool) error {
	model := NewModel(cfg, client, showBanner)

	// Don't use alt screen - allows scrollback in terminal
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running shell: %w", err)
	}

	// Check for errors in the final model
	if m, ok := finalModel.(Model); ok && m.err != nil {
		fmt.Fprintln(os.Stderr, m.err)
	}

	return nil
}
