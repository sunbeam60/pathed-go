package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// modal represents a dialog that blocks input to the rest of the program
type modal struct {
	question string
	options  []string
	selected int
	onSelect func(index int) tea.Cmd
}

func newModal(question string, options []string, onSelect func(index int) tea.Cmd) *modal {
	return &modal{
		question: question,
		options:  options,
		selected: 0,
		onSelect: onSelect,
	}
}

// Update handles input for the modal and returns the result
// Returns: updated modal (nil if dismissed), tea.Cmd to execute, whether modal was actioned
func (m *modal) Update(msg tea.KeyMsg) (*modal, tea.Cmd, bool) {
	switch msg.String() {
	case "left", "h":
		if m.selected > 0 {
			m.selected--
		}
	case "right", "l":
		if m.selected < len(m.options)-1 {
			m.selected++
		}
	case "enter":
		if m.onSelect != nil {
			return nil, m.onSelect(m.selected), true
		}
		return nil, nil, true
	case "esc":
		return nil, nil, true
	case "y":
		// Shortcut for Yes (first option)
		if m.onSelect != nil {
			return nil, m.onSelect(0), true
		}
		return nil, nil, true
	case "n":
		// Shortcut for No (second option)
		if len(m.options) > 1 && m.onSelect != nil {
			return nil, m.onSelect(1), true
		}
		return nil, nil, true
	}
	return m, nil, false
}

// View renders the modal
func (m *modal) View() string {
	var opts []string
	for i, opt := range m.options {
		// Underline first letter as shortcut hint
		styledOpt := ansiUnderline + string(opt[0]) + ansiNoUnder + opt[1:]
		if i == m.selected {
			opts = append(opts, "["+styledOpt+"]")
		} else {
			opts = append(opts, " "+styledOpt+" ")
		}
	}
	return fmt.Sprintf("%s  %s  (Esc to cancel)", m.question, strings.Join(opts, "  "))
}
