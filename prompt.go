package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// prompt represents a dialog that blocks input to the rest of the program
type prompt struct {
	question string
	options  []string
	selected int
	onSelect func(index int) tea.Cmd
}

func newPrompt(question string, options []string, onSelect func(index int) tea.Cmd) *prompt {
	return &prompt{
		question: question,
		options:  options,
		selected: 0,
		onSelect: onSelect,
	}
}

// Update handles input for the prompt and returns the result
// Returns: updated prompt (nil if dismissed), tea.Cmd to execute, whether prompt was actioned
func (p *prompt) Update(msg tea.KeyMsg) (*prompt, tea.Cmd, bool) {
	switch msg.String() {
	case "left", "h":
		if p.selected > 0 {
			p.selected--
		}
	case "right", "l":
		if p.selected < len(p.options)-1 {
			p.selected++
		}
	case "enter":
		if p.onSelect != nil {
			return nil, p.onSelect(p.selected), true
		}
		return nil, nil, true
	case "esc":
		return nil, nil, true
	case "y":
		// Shortcut for Yes (first option)
		if p.onSelect != nil {
			return nil, p.onSelect(0), true
		}
		return nil, nil, true
	case "n":
		// Shortcut for No (second option)
		if len(p.options) > 1 && p.onSelect != nil {
			return nil, p.onSelect(1), true
		}
		return nil, nil, true
	}
	return p, nil, false
}

// View renders the prompt
func (p *prompt) View() string {
	var opts []string
	for i, opt := range p.options {
		// Underline first letter as shortcut hint
		styledOpt := ansiUnderline + string(opt[0]) + ansiNoUnder + opt[1:]
		if i == p.selected {
			opts = append(opts, "["+styledOpt+"]")
		} else {
			opts = append(opts, " "+styledOpt+" ")
		}
	}
	return fmt.Sprintf("%s  %s  (Esc to cancel)", p.question, strings.Join(opts, "  "))
}
