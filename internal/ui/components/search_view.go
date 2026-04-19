package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ollama/ollama/api"
)

// SearchMode represents the current search state
type SearchMode int

const (
	SearchModeInactive SearchMode = iota
	SearchModeActive
)

// SearchView handles the search UI component
type SearchView struct {
	mode           SearchMode
	input          textinput.Model
	engine         *SearchEngine
	caseSensitive  bool
	width          int
	height         int
	
	// Styles
	containerStyle    lipgloss.Style
	inputStyle        lipgloss.Style
	statusStyle       lipgloss.Style
	highlightStyle    lipgloss.Style
	activeMatchStyle  lipgloss.Style
}

// SearchViewMsg represents search-related messages
type SearchViewMsg struct {
	Type string
	Data interface{}
}

// Message types
const (
	SearchActivated   = "search_activated"
	SearchDeactivated = "search_deactivated"
	SearchUpdated     = "search_updated"
	MatchChanged      = "match_changed"
)

// NewSearchView creates a new search view
func NewSearchView() *SearchView {
	input := textinput.New()
	input.Placeholder = "Search conversation..."
	input.CharLimit = 256

	return &SearchView{
		mode:   SearchModeInactive,
		input:  input,
		engine: NewSearchEngine(),
		
		// Default styles
		containerStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		
		inputStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		
		statusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Italic(true),
		
		highlightStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("220")).
			Foreground(lipgloss.Color("16")),
		
		activeMatchStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("196")).
			Foreground(lipgloss.Color("15")).
			Bold(true),
	}
}

// SetSize updates the search view dimensions
func (sv *SearchView) SetSize(width, height int) {
	sv.width = width
	sv.height = height
	sv.input.Width = width - 4 // Account for border and padding
}

// SetTheme updates the search view styles based on theme
func (sv *SearchView) SetTheme(isDark bool) {
	if isDark {
		sv.containerStyle = sv.containerStyle.
			BorderForeground(lipgloss.Color("240"))
		
		sv.inputStyle = sv.inputStyle.
			Foreground(lipgloss.Color("252"))
		
		sv.statusStyle = sv.statusStyle.
			Foreground(lipgloss.Color("243"))
		
		sv.highlightStyle = sv.highlightStyle.
			Background(lipgloss.Color("220")).
			Foreground(lipgloss.Color("16"))
		
		sv.activeMatchStyle = sv.activeMatchStyle.
			Background(lipgloss.Color("196")).
			Foreground(lipgloss.Color("15"))
	} else {
		sv.containerStyle = sv.containerStyle.
			BorderForeground(lipgloss.Color("240"))
		
		sv.inputStyle = sv.inputStyle.
			Foreground(lipgloss.Color("16"))
		
		sv.statusStyle = sv.statusStyle.
			Foreground(lipgloss.Color("102"))
		
		sv.highlightStyle = sv.highlightStyle.
			Background(lipgloss.Color("226")).
			Foreground(lipgloss.Color("16"))
		
		sv.activeMatchStyle = sv.activeMatchStyle.
			Background(lipgloss.Color("160")).
			Foreground(lipgloss.Color("15"))
	}
}

// Activate enables search mode
func (sv *SearchView) Activate() tea.Cmd {
	sv.mode = SearchModeActive
	sv.input.Focus()
	return textinput.Blink
}

// Deactivate disables search mode and clears highlights
func (sv *SearchView) Deactivate() tea.Cmd {
	sv.mode = SearchModeInactive
	sv.input.Blur()
	sv.engine.ClearHighlights()
	return func() tea.Msg {
		return SearchViewMsg{Type: SearchDeactivated}
	}
}

// IsActive returns true if search mode is active
func (sv *SearchView) IsActive() bool {
	return sv.mode == SearchModeActive
}

// Update handles search view updates
func (sv *SearchView) Update(msg tea.Msg) tea.Cmd {
	if sv.mode != SearchModeActive {
		return nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "escape":
			return sv.Deactivate()
		
		case "enter", "ctrl+f":
			// Perform search with current input
			return sv.performSearch()
		
		case "ctrl+n", "f3":
			// Next match
			return sv.nextMatch()
		
		case "ctrl+p", "shift+f3":
			// Previous match
			return sv.previousMatch()
		
		case "ctrl+i":
			// Toggle case sensitivity
			sv.caseSensitive = !sv.caseSensitive
			return sv.performSearch()
		
		default:
			// Update input and perform search
			sv.input, cmd = sv.input.Update(msg)
			
			// Perform search on input change
			searchCmd := sv.performSearch()
			return tea.Batch(cmd, searchCmd)
		}
	
	default:
		sv.input, cmd = sv.input.Update(msg)
	}

	return cmd
}

// performSearch executes a search with the current input
func (sv *SearchView) performSearch() tea.Cmd {
	return func() tea.Msg {
		return SearchViewMsg{
			Type: SearchUpdated,
			Data: sv.input.Value(),
		}
	}
}

// nextMatch moves to the next search result
func (sv *SearchView) nextMatch() tea.Cmd {
	result := sv.engine.NextMatch()
	if result != nil {
		return func() tea.Msg {
			return SearchViewMsg{
				Type: MatchChanged,
				Data: result,
			}
		}
	}
	return nil
}

// previousMatch moves to the previous search result
func (sv *SearchView) previousMatch() tea.Cmd {
	result := sv.engine.PreviousMatch()
	if result != nil {
		return func() tea.Msg {
			return SearchViewMsg{
				Type: MatchChanged,
				Data: result,
			}
		}
	}
	return nil
}

// Search performs a search across messages
func (sv *SearchView) Search(query string, messages []api.Message) error {
	options := SearchOptions{
		CaseSensitive: sv.caseSensitive,
		WholeWord:     false,
		Regex:         false,
	}
	
	return sv.engine.Search(query, messages, options)
}

// HighlightMessage applies search highlighting to a message
func (sv *SearchView) HighlightMessage(content string, msgIdx int) string {
	if sv.mode != SearchModeActive || !sv.engine.HasResults() {
		return content
	}
	
	return sv.engine.HighlightText(content, msgIdx, sv.highlightStyle, sv.activeMatchStyle)
}

// GetCurrentMatch returns the current search match
func (sv *SearchView) GetCurrentMatch() *SearchResult {
	return sv.engine.GetCurrentMatch()
}

// GetMatchInfo returns match count and current position
func (sv *SearchView) GetMatchInfo() (current, total int) {
	return sv.engine.GetCurrentMatchIndex(), sv.engine.GetMatchCount()
}

// View renders the search view
func (sv *SearchView) View() string {
	if sv.mode != SearchModeActive {
		return ""
	}

	// Input field
	inputView := sv.inputStyle.Render(sv.input.View())
	
	// Status line with match info and options
	var status string
	current, total := sv.GetMatchInfo()
	
	if total > 0 {
		status = fmt.Sprintf("%d/%d matches", current, total)
	} else if sv.input.Value() != "" {
		status = "No matches"
	} else {
		status = "Type to search"
	}
	
	// Add case sensitivity indicator
	if sv.caseSensitive {
		status += " [Aa]"
	} else {
		status += " [aa]"
	}
	
	statusView := sv.statusStyle.Render(status)
	
	// Help text
	help := sv.statusStyle.Render("ESC: exit • Enter: search • Ctrl+N/P: next/prev • Ctrl+I: case")
	
	// Combine all elements
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		inputView,
		statusView,
		help,
	)
	
	return sv.containerStyle.Render(content)
}

// GetEngine returns the search engine for external access
func (sv *SearchView) GetEngine() *SearchEngine {
	return sv.engine
}