/*
Copyright 2025 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package form

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TextInput is a simple text input component
type TextInput struct {
	Value       string
	Placeholder string
	Label       string
	Focused     bool
	CursorPos   int
	ErrorMsg    string
	Validator   func(string) error
}

// NewTextInput creates a new text input
func NewTextInput(label, placeholder string, validator func(string) error) *TextInput {
	return &TextInput{
		Label:       label,
		Placeholder: placeholder,
		Validator:   validator,
		Focused:     false,
	}
}

// Update handles input messages
func (ti *TextInput) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			ti.Value += string(msg.Runes)
			ti.CursorPos += len(msg.Runes)
		case tea.KeyBackspace:
			if ti.CursorPos > 0 {
				ti.Value = ti.Value[:ti.CursorPos-1] + ti.Value[ti.CursorPos:]
				ti.CursorPos--
				ti.ErrorMsg = ""
			}
		}
	}
	return nil
}

// View renders the text input
func (ti *TextInput) View() string {
	style := lipgloss.NewStyle()
	if ti.ErrorMsg != "" {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	}

	displayValue := ti.Value
	if displayValue == "" {
		displayValue = ti.Placeholder
	}

	return style.Render(ti.Label + ": " + displayValue)
}

// Validate validates the current input
func (ti *TextInput) Validate() error {
	if ti.Validator != nil {
		return ti.Validator(ti.Value)
	}
	return nil
}

// MultiSelect is a multi-select choice component
type MultiSelect struct {
	Options   []Option
	Selected  map[string]bool
	Label     string
	CursorPos int
	Focused   bool
	ErrorMsg  string
	Validator func([]string) error
}

// Option represents a single option in a multi-select
type Option struct {
	Label string
	Value string
}

// NewMultiSelect creates a new multi-select component
func NewMultiSelect(label string, options []Option, validator func([]string) error) *MultiSelect {
	ms := &MultiSelect{
		Label:     label,
		Options:   options,
		Selected:  make(map[string]bool),
		Focused:   false,
		Validator: validator,
	}
	return ms
}

// Update handles input messages
func (ms *MultiSelect) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if ms.CursorPos > 0 {
				ms.CursorPos--
			}
		case "down":
			if ms.CursorPos < len(ms.Options)-1 {
				ms.CursorPos++
			}
		case " ":
			ms.Selected[ms.Options[ms.CursorPos].Value] = !ms.Selected[ms.Options[ms.CursorPos].Value]
			ms.ErrorMsg = ""
		}
	}
	return nil
}

// View renders the multi-select
func (ms *MultiSelect) View() string {
	style := lipgloss.NewStyle()
	if ms.ErrorMsg != "" {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	}

	var output strings.Builder
	output.WriteString(style.Render(ms.Label) + "\n")

	for i, opt := range ms.Options {
		cursor := " "
		if i == ms.CursorPos {
			cursor = "›"
		}

		checked := " "
		if ms.Selected[opt.Value] {
			checked = "✓"
		}

		output.WriteString("  " + cursor + " [" + checked + "] " + opt.Label + "\n")
	}

	return output.String()
}

// GetSelected returns the list of selected values
func (ms *MultiSelect) GetSelected() []string {
	var result []string
	for _, opt := range ms.Options {
		if ms.Selected[opt.Value] {
			result = append(result, opt.Value)
		}
	}
	return result
}

// Validate validates the current selection
func (ms *MultiSelect) Validate() error {
	selected := ms.GetSelected()
	if ms.Validator != nil {
		return ms.Validator(selected)
	}
	return nil
}

// SingleSelect is a single-select choice component
type SingleSelect struct {
	Options   []Option
	Selected  string
	Label     string
	CursorPos int
	Focused   bool
}

// NewSingleSelect creates a new single-select component
func NewSingleSelect(label string, options []Option) *SingleSelect {
	return &SingleSelect{
		Label:   label,
		Options: options,
		Focused: false,
	}
}

// Update handles input messages
func (ss *SingleSelect) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if ss.CursorPos > 0 {
				ss.CursorPos--
			}
		case "down":
			if ss.CursorPos < len(ss.Options)-1 {
				ss.CursorPos++
			}
		}
	}
	return nil
}

// View renders the single-select
func (ss *SingleSelect) View() string {
	var output strings.Builder
	output.WriteString(ss.Label + "\n")

	for i, opt := range ss.Options {
		cursor := " "
		if i == ss.CursorPos {
			cursor = "›"
		}
		output.WriteString("  " + cursor + " " + opt.Label + "\n")
	}

	return output.String()
}

// GetSelected returns the selected value
func (ss *SingleSelect) GetSelected() string {
	if ss.CursorPos >= 0 && ss.CursorPos < len(ss.Options) {
		return ss.Options[ss.CursorPos].Value
	}
	return ""
}

// Confirmation is a yes/no confirmation component
type Confirmation struct {
	Label     string
	Value     bool
	CursorPos int // 0 for Yes, 1 for No
	Focused   bool
}

// NewConfirmation creates a new confirmation component
func NewConfirmation(label string) *Confirmation {
	return &Confirmation{
		Label:     label,
		CursorPos: 0,
		Focused:   false,
	}
}

// Update handles input messages
func (c *Confirmation) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "up":
			if c.CursorPos > 0 {
				c.CursorPos--
			}
		case "right", "down":
			if c.CursorPos < 1 {
				c.CursorPos++
			}
		}
	}
	c.Value = c.CursorPos == 0
	return nil
}

// View renders the confirmation
func (c *Confirmation) View() string {
	yes := " Yes "
	no := " No  "

	if c.CursorPos == 0 {
		yes = "> Yes <"
	} else {
		no = "> No  <"
	}

	return c.Label + ": " + yes + " " + no
}
