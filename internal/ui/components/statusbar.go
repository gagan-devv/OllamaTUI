package components

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gagan-devv/ollama-go/internal/ui/theme"
)

// ConnectionStatus represents the connection state
type ConnectionStatus int

const (
	Connected ConnectionStatus = iota
	Disconnected
	Connecting
)

// StatusBarModel represents the status bar component
type StatusBarModel struct {
	width            int
	theme            theme.Theme
	connectionStatus ConnectionStatus
	modelName        string
	sessionName      string
	tokenCount       int
	showMetrics      bool
	currentTime      time.Time
	lastUpdate       time.Time
	narrowMode       bool  // For narrow terminal layout
}

// StatusBarUpdateMsg is sent to update status bar information
type StatusBarUpdateMsg struct {
	ConnectionStatus *ConnectionStatus
	ModelName        *string
	SessionName      *string
	TokenCount       *int
	ShowMetrics      *bool
}

// TickMsg is sent periodically to update the current time
type TickMsg time.Time

// NewStatusBar creates a new status bar component
func NewStatusBar(theme theme.Theme, showMetrics bool) *StatusBarModel {
	return &StatusBarModel{
		theme:            theme,
		connectionStatus: Connected,
		modelName:        "",  // Will be set via UpdateModel
		sessionName:      "Default",
		tokenCount:       0,
		showMetrics:      showMetrics,
		currentTime:      time.Now(),
		lastUpdate:       time.Now(),
	}
}

// Init initializes the status bar
func (m *StatusBarModel) Init() tea.Cmd {
	return m.tick()
}

// Update handles status bar updates
func (m *StatusBarModel) Update(msg tea.Msg) (*StatusBarModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
		
	case StatusBarUpdateMsg:
		// Update fields if provided
		if msg.ConnectionStatus != nil {
			m.connectionStatus = *msg.ConnectionStatus
		}
		if msg.ModelName != nil {
			m.modelName = *msg.ModelName
		}
		if msg.SessionName != nil {
			m.sessionName = *msg.SessionName
		}
		if msg.TokenCount != nil {
			m.tokenCount = *msg.TokenCount
		}
		if msg.ShowMetrics != nil {
			m.showMetrics = *msg.ShowMetrics
		}
		m.lastUpdate = time.Now()
		return m, nil
		
	case TickMsg:
		m.currentTime = time.Time(msg)
		return m, m.tick()
	}
	
	return m, nil
}

// View renders the status bar
func (m *StatusBarModel) View() string {
	if m.width == 0 {
		return ""
	}
	
	// Create status bar style
	statusStyle := lipgloss.NewStyle().
		Background(m.theme.StatusBarColor()).
		Foreground(lipgloss.Color("#FFFFFF")).
		Width(m.width).
		Padding(0, 1)
	
	// Build status components
	var components []string
	
	// Connection status
	connectionIcon := m.getConnectionIcon()
	components = append(components, connectionIcon)
	
	// In narrow mode, show only essential info (requirement 47.7)
	if m.narrowMode {
		// For very narrow terminals (< 50 cols), show minimal info
		if m.width < 50 {
			// Only show connection status and abbreviated model
			if m.modelName != "" {
				modelName := m.modelName
				if len(modelName) > 8 {
					modelName = modelName[:5] + "..."
				}
				components = append(components, modelName)
			}
		} else {
			// For narrow terminals (50-100 cols), show more info
			if m.modelName != "" {
				// Truncate model name if too long
				modelName := m.modelName
				if len(modelName) > 15 {
					modelName = modelName[:12] + "..."
				}
				components = append(components, modelName)
			}
			
			// Show token count if metrics enabled and space allows
			if m.showMetrics && m.tokenCount > 0 && m.width > 70 {
				components = append(components, fmt.Sprintf("T:%d", m.tokenCount))
			}
		}
	} else {
		// Full mode - show all components
		// Model name
		if m.modelName != "" {
			components = append(components, fmt.Sprintf("Model: %s", m.modelName))
		}
		
		// Session name
		if m.sessionName != "" {
			components = append(components, fmt.Sprintf("Session: %s", m.sessionName))
		}
		
		// Token count (if metrics enabled)
		if m.showMetrics && m.tokenCount > 0 {
			components = append(components, fmt.Sprintf("Tokens: %d", m.tokenCount))
		}
	}
	
	// Current time - adapt format based on width
	var timeStr string
	if m.width < 40 {
		timeStr = m.currentTime.Format("15:04") // Short format for very narrow terminals
	} else {
		timeStr = m.currentTime.Format("15:04:05") // Full format for wider terminals
	}
	
	// Join left components
	leftContent := strings.Join(components, " | ")
	
	// Calculate available space for content
	availableWidth := m.width - 2 // Account for padding
	timeWidth := len(timeStr)
	
	// Ensure we have space for time
	if len(leftContent)+timeWidth+3 > availableWidth {
		// Truncate left content to fit
		maxLeftWidth := availableWidth - timeWidth - 3
		if maxLeftWidth > 0 && len(leftContent) > maxLeftWidth {
			leftContent = leftContent[:maxLeftWidth-3] + "..."
		}
	}
	
	// Create the status bar content with time right-aligned
	content := leftContent
	spacesNeeded := availableWidth - len(leftContent) - timeWidth
	if spacesNeeded > 0 {
		content += strings.Repeat(" ", spacesNeeded) + timeStr
	} else {
		content += " " + timeStr
	}
	
	return statusStyle.Render(content)
}

// SetTheme updates the theme
func (m *StatusBarModel) SetTheme(theme theme.Theme) {
	m.theme = theme
}

// SetWidth updates the width
func (m *StatusBarModel) SetWidth(width int) {
	m.width = width
}

// UpdateConnection updates the connection status
func (m *StatusBarModel) UpdateConnection(status ConnectionStatus) tea.Cmd {
	return func() tea.Msg {
		return StatusBarUpdateMsg{
			ConnectionStatus: &status,
		}
	}
}

// UpdateModel updates the model name
func (m *StatusBarModel) UpdateModel(modelName string) tea.Cmd {
	return func() tea.Msg {
		return StatusBarUpdateMsg{
			ModelName: &modelName,
		}
	}
}

// UpdateSession updates the session name
func (m *StatusBarModel) UpdateSession(sessionName string) tea.Cmd {
	return func() tea.Msg {
		return StatusBarUpdateMsg{
			SessionName: &sessionName,
		}
	}
}

// UpdateTokenCount updates the token count
func (m *StatusBarModel) UpdateTokenCount(count int) tea.Cmd {
	return func() tea.Msg {
		return StatusBarUpdateMsg{
			TokenCount: &count,
		}
	}
}

// UpdateMetricsVisibility updates whether metrics are shown
func (m *StatusBarModel) UpdateMetricsVisibility(show bool) tea.Cmd {
	return func() tea.Msg {
		return StatusBarUpdateMsg{
			ShowMetrics: &show,
		}
	}
}

// tick returns a command that sends a TickMsg after 1 second
func (m *StatusBarModel) tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// SetNarrowMode enables or disables narrow mode for small terminals
// This adjusts the layout for terminals with limited width (requirement 47.7)
func (m *StatusBarModel) SetNarrowMode(narrow bool) {
	m.narrowMode = narrow
}

// IsNarrowMode returns whether narrow mode is currently enabled
func (m *StatusBarModel) IsNarrowMode() bool {
	return m.narrowMode
}

// getConnectionIcon returns the appropriate icon for the current connection status
// getConnectionIcon returns the appropriate icon for the current connection status
func (m *StatusBarModel) getConnectionIcon() string {
	switch m.connectionStatus {
	case Connected:
		return "🟢 Connected"
	case Connecting:
		return "🟡 Connecting"
	case Disconnected:
		return "🔴 Disconnected"
	default:
		return "? Unknown"
	}
}
