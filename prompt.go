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
	key := msg.String()
	switch key {
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
	default:
		// Check if key matches first letter of any option (case-insensitive)
		if len(key) == 1 {
			keyLower := strings.ToLower(key)
			for i, opt := range p.options {
				if len(opt) > 0 && strings.ToLower(string(opt[0])) == keyLower {
					if p.onSelect != nil {
						return nil, p.onSelect(i), true
					}
					return nil, nil, true
				}
			}
		}
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
