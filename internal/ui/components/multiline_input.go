package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	maxVisibleLines = 10
	minHeight       = 3
)

// SubmitMsg is sent when the user presses Enter to submit.
type SubmitMsg struct{}

// MultiLineInput is an enhanced input component with multi-line support,
// auto-expansion, and input history navigation.
type MultiLineInput struct {
	textarea      textarea.Model
	history       *InputHistory
	historyIndex  int
	currentDraft  string
	width         int
	maxLines      int
	infoStyle     lipgloss.Style
}

// NewMultiLineInput creates a new multi-line input component.
func NewMultiLineInput(historyPath string) *MultiLineInput {
	ta := textarea.New()
	ta.Placeholder = "Ask me anything... (Alt+Enter for newline, Enter to submit)"
	ta.SetHeight(minHeight)
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.Focus()

	// Disable ALL default keybindings that we want to handle ourselves
	ta.KeyMap.InsertNewline.SetEnabled(false)
	
	// Try to prevent textarea from handling any Enter keys
	ta.KeyMap.LineEnd.SetEnabled(true)
	ta.KeyMap.LineStart.SetEnabled(true)
	
	history := NewInputHistory(historyPath, 100)

	return &MultiLineInput{
		textarea:     ta,
		history:      history,
		historyIndex: -1,
		maxLines:     maxVisibleLines,
		infoStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
	}
}

// Init initializes the component.
func (m *MultiLineInput) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages and updates the component state.
func (m *MultiLineInput) Update(msg tea.Msg) (*MultiLineInput, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyStr := msg.String()
		
		// Handle Alt+Enter or Ctrl+J for newline insertion
		if keyStr == "alt+enter" || keyStr == "ctrl+j" || (msg.Type == tea.KeyEnter && msg.Alt) {
			// Insert newline manually and preserve indentation
			currentLine := m.getCurrentLine()
			indentation := m.getIndentation(currentLine)
			
			// Manually insert newline
			currentValue := m.textarea.Value()
			m.textarea.SetValue(currentValue + "\n" + indentation)
			
			// Auto-expand height
			m.adjustHeight()
			return m, nil
		}
		
		// Handle plain Enter for submission
		if msg.Type == tea.KeyEnter && !msg.Alt {
			return m, func() tea.Msg { return SubmitMsg{} }
		}
		
		switch keyStr {
		case "up", "ctrl+p":
			// Navigate history backward with Up arrow or Ctrl+P
			if m.historyIndex < m.history.Len()-1 {
				// Save current draft if we're starting to navigate history
				if m.historyIndex == -1 {
					m.currentDraft = m.textarea.Value()
				}
				
				m.historyIndex++
				entry := m.history.Get(m.historyIndex)
				m.textarea.SetValue(entry)
				m.adjustHeight()
				return m, nil
			}
			return m, nil

		case "down", "ctrl+n":
			// Navigate history forward with Down arrow or Ctrl+N
			if m.historyIndex > -1 {
				m.historyIndex--
				
				if m.historyIndex == -1 {
					// Restore draft
					m.textarea.SetValue(m.currentDraft)
				} else {
					entry := m.history.Get(m.historyIndex)
					m.textarea.SetValue(entry)
				}
				
				m.adjustHeight()
				return m, nil
			}
			return m, nil
		}
	}

	// Default textarea update
	m.textarea, cmd = m.textarea.Update(msg)
	m.adjustHeight()
	
	return m, cmd
}

// View renders the component.
func (m *MultiLineInput) View() string {
	return m.textarea.View() + "\n" + 
		m.infoStyle.Render("Enter: Submit | Alt+Enter or Ctrl+J: New line | ↑↓: History")
}

// Value returns the current input value.
func (m *MultiLineInput) Value() string {
	return m.textarea.Value()
}

// Reset clears the input and adds the value to history.
func (m *MultiLineInput) Reset() {
	value := m.textarea.Value()
	if strings.TrimSpace(value) != "" {
		m.history.Add(value)
	}
	
	m.textarea.Reset()
	m.textarea.SetHeight(minHeight)
	m.historyIndex = -1
	m.currentDraft = ""
}

// SetWidth sets the width of the input.
func (m *MultiLineInput) SetWidth(width int) {
	m.width = width
	m.textarea.SetWidth(width)
}

// Focus focuses the input.
func (m *MultiLineInput) Focus() tea.Cmd {
	return m.textarea.Focus()
}

// Blur removes focus from the input.
func (m *MultiLineInput) Blur() {
	m.textarea.Blur()
}

// Focused returns whether the input is focused.
func (m *MultiLineInput) Focused() bool {
	return m.textarea.Focused()
}

// IsInHistoryMode returns true if the user is currently navigating history.
func (m *MultiLineInput) IsInHistoryMode() bool {
	return m.historyIndex >= 0
}

// adjustHeight adjusts the textarea height based on content.
func (m *MultiLineInput) adjustHeight() {
	lines := strings.Count(m.textarea.Value(), "\n") + 1
	
	if lines <= minHeight {
		m.textarea.SetHeight(minHeight)
	} else if lines <= m.maxLines {
		// Auto-expand up to maxLines
		m.textarea.SetHeight(lines)
	} else {
		// Make scrollable beyond maxLines
		m.textarea.SetHeight(m.maxLines)
	}
}

// isAtStart returns true if the cursor is at the start of the input.
func (m *MultiLineInput) isAtStart() bool {
	// For history navigation, we'll check if the input is empty or
	// if we're already navigating history
	value := m.textarea.Value()
	return value == "" || m.historyIndex >= 0
}

// getCurrentLine returns the line where the cursor is currently positioned.
func (m *MultiLineInput) getCurrentLine() string {
	value := m.textarea.Value()
	if value == "" {
		return ""
	}
	
	// Get the last line (where cursor typically is when typing)
	lines := strings.Split(value, "\n")
	if len(lines) > 0 {
		return lines[len(lines)-1]
	}
	
	return ""
}

// SetMaxLines sets the maximum number of visible lines for the input area
// This is used for proportional scaling based on terminal height (requirement 47.5)
func (m *MultiLineInput) SetMaxLines(maxLines int) {
	if maxLines < minHeight {
		maxLines = minHeight
	}
	m.maxLines = maxLines
	m.adjustHeight() // Readjust height with new max
}

// SetHeight sets a specific height for the input area
func (m *MultiLineInput) SetHeight(height int) {
	if height < minHeight {
		height = minHeight
	}
	if height > m.maxLines {
		height = m.maxLines
	}
	m.textarea.SetHeight(height)
}

// GetCurrentHeight returns the current height of the input area
func (m *MultiLineInput) GetCurrentHeight() int {
	return m.textarea.Height()
}
// getIndentation extracts the indentation (leading whitespace) from a line
func (m *MultiLineInput) getIndentation(line string) string {
	var indentation strings.Builder
	for _, char := range line {
		if char == ' ' || char == '\t' {
			indentation.WriteRune(char)
		} else {
			break
		}
	}
	return indentation.String()
}
