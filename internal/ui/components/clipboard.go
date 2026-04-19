package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ClipboardMsg represents a clipboard operation message
type ClipboardMsg struct {
	Success bool
	Message string
	Error   error
}

// ClipboardConfirmationMsg is sent to hide the confirmation after a delay
type ClipboardConfirmationMsg struct{}

// ClipboardModel manages clipboard operation state and confirmation display
type ClipboardModel struct {
	showConfirmation bool
	confirmationMsg  string
	isError          bool
	confirmationTime time.Time
	displayDuration  time.Duration
}

// NewClipboardModel creates a new clipboard model
func NewClipboardModel() *ClipboardModel {
	return &ClipboardModel{
		showConfirmation: false,
		displayDuration:  2 * time.Second, // Show confirmation for 2 seconds
	}
}

// Update handles clipboard-related messages
func (m *ClipboardModel) Update(msg tea.Msg) (*ClipboardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ClipboardMsg:
		m.showConfirmation = true
		m.confirmationTime = time.Now()
		m.isError = !msg.Success
		
		if msg.Success {
			m.confirmationMsg = msg.Message
		} else {
			if msg.Error != nil {
				m.confirmationMsg = "Clipboard error: " + msg.Error.Error()
			} else {
				m.confirmationMsg = "Failed to copy to clipboard"
			}
		}
		
		// Schedule hiding the confirmation
		return m, tea.Tick(m.displayDuration, func(t time.Time) tea.Msg {
			return ClipboardConfirmationMsg{}
		})
		
	case ClipboardConfirmationMsg:
		// Hide confirmation if enough time has passed
		if time.Since(m.confirmationTime) >= m.displayDuration {
			m.showConfirmation = false
			m.confirmationMsg = ""
		}
	}
	
	return m, nil
}

// View renders the clipboard confirmation indicator
func (m *ClipboardModel) View() string {
	if !m.showConfirmation {
		return ""
	}
	
	var style lipgloss.Style
	if m.isError {
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Background(lipgloss.Color("#3A0000")).
			Padding(0, 1).
			Bold(true)
	} else {
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Background(lipgloss.Color("#003A00")).
			Padding(0, 1).
			Bold(true)
	}
	
	return style.Render(m.confirmationMsg)
}

// IsShowing returns true if the confirmation is currently visible
func (m *ClipboardModel) IsShowing() bool {
	return m.showConfirmation
}

// ShowSuccess displays a success confirmation message
func ShowSuccess(message string) tea.Cmd {
	return func() tea.Msg {
		return ClipboardMsg{
			Success: true,
			Message: message,
		}
	}
}

// ShowError displays an error message
func ShowError(err error) tea.Cmd {
	return func() tea.Msg {
		return ClipboardMsg{
			Success: false,
			Error:   err,
		}
	}
}
