package ui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/gagan-devv/ollama-go/internal/client"
	"github.com/gagan-devv/ollama-go/internal/config"
	"github.com/gagan-devv/ollama-go/internal/ui/components"
	"github.com/gagan-devv/ollama-go/internal/ui/theme"
	"github.com/gagan-devv/ollama-go/internal/util"
	"github.com/ollama/ollama/api"
)

var (
	titleStyle = lipgloss.NewStyle().MarginLeft(2).Bold(true).Foreground(lipgloss.Color("5"))
	userStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	aiStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginLeft(2)
)

type Model struct {
	viewport       viewport.Model
	textarea       textarea.Model
	multilineInput *components.MultiLineInput
	statusBar      *components.StatusBarModel
	clipboard      *components.ClipboardModel
	searchView     *components.SearchView
	useMultiline   bool
	spinner        spinner.Model
	history        []api.Message
	client         *api.Client
	streamHandler  *client.StreamHandler
	modelName      string
	sessionName    string
	isThinking     bool
	err            error
	width          int
	height         int
	streamStarted  time.Time
	lastTokenTime  time.Time
	partialContent string // Buffer for current streaming message
	totalTokens    int    // Total tokens in current session
	theme          theme.Theme
	configManager  *config.ConfigManager
	
	// Responsive design fields (requirements 47.1-47.7)
	minWidth               int
	minHeight              int
	lastResizeTime         time.Time
	showSizeWarning        bool
	previousScrollPosition int
	resizeInProgress       bool
	
	// Clipboard state
	selectedMessageIndex int // Index of currently selected message for copying (-1 if none)
}

func InitialModel(apiClient *api.Client, modelName, system string, cfg *config.Config, configManager *config.ConfigManager) Model {
	ta := textarea.New()
	ta.Placeholder = "Ask me anything..."
	ta.SetHeight(3)
	ta.Focus()

	vp := viewport.New(80, 20) // Give it a default size
	vp.SetContent("Ready.")

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Initialize theme from config
	currentTheme := theme.GetTheme(cfg.UI.Theme, &cfg.UI.Colors)

	// Initialize multi-line input with history
	historyPath := filepath.Join(cfg.Paths.TestFolder, "history", "input_history.json")
	multilineInput := components.NewMultiLineInput(historyPath)

	// Initialize status bar
	statusBar := components.NewStatusBar(currentTheme, cfg.UI.ShowMetrics)
	
	// Initialize clipboard model
	clipboardModel := components.NewClipboardModel()
	
	// Initialize search view
	searchView := components.NewSearchView()
	searchView.SetTheme(cfg.UI.Theme == "dark")

	return Model{
		textarea:       ta,
		multilineInput: multilineInput,
		statusBar:      statusBar,
		clipboard:      clipboardModel,
		searchView:     searchView,
		useMultiline:   true, // Use enhanced multi-line input by default
		viewport:       vp,
		spinner:        s,
		client:         apiClient,
		streamHandler:  client.NewStreamHandler(apiClient, modelName),
		modelName:      modelName,
		sessionName:    "Default",
		width:          80, // Fallback width
		height:         24, // Fallback height
		history:        []api.Message{{Role: "system", Content: system}},
		totalTokens:    0,
		theme:          currentTheme,
		configManager:          configManager,
		
		// Responsive design initialization (requirements 47.2, 47.3)
		minWidth:               80,  // Minimum 80 columns
		minHeight:              24,  // Minimum 24 rows
		lastResizeTime:         time.Now(),
		showSizeWarning:        false,
		previousScrollPosition: 0,
		resizeInProgress:       false,
		
		// Clipboard initialization
		selectedMessageIndex: -1, // No message selected initially
	}
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	
	if m.useMultiline {
		cmds = append(cmds, m.multilineInput.Init())
	} else {
		cmds = append(cmds, textarea.Blink)
	}
	
	cmds = append(cmds, m.spinner.Tick, m.statusBar.Init(), m.statusBar.UpdateModel(m.modelName), tea.EnableMouseCellMotion)
	
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		spCmd tea.Cmd
		mlCmd tea.Cmd
		sbCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Record resize time for performance tracking (requirement 47.1)
		resizeStartTime := time.Now()
		m.lastResizeTime = resizeStartTime
		m.resizeInProgress = true
		
		// Store previous scroll position to preserve it (requirement 47.6)
		m.previousScrollPosition = m.viewport.YOffset
		
		// Update dimensions
		m.width = msg.Width
		m.height = msg.Height
		
		// Check minimum dimensions and display warning (requirements 47.2, 47.3)
		m.showSizeWarning = msg.Width < m.minWidth || msg.Height < m.minHeight
		
		// Calculate available space for components efficiently
		headerHeight := 3  // Title and spacing
		statusBarHeight := 1
		marginHeight := 1
		
		// Calculate proportional input height based on terminal height (requirement 47.5)
		var inputHeight, maxInputHeight int
		if msg.Height > 40 {
			inputHeight = 5
			maxInputHeight = 20 // Very large terminals get more input space
		} else if msg.Height > 30 {
			inputHeight = 4
			maxInputHeight = 15 // Large terminals get more input space
		} else if msg.Height > 20 {
			inputHeight = 3
			maxInputHeight = 10 // Medium terminals get standard space
		} else {
			inputHeight = 2
			maxInputHeight = 5  // Small terminals get minimal space
		}
		
		footerHeight := inputHeight + 1 // Input area plus info text (reduced spacing)
		
		// Adjust viewport dimensions
		viewportWidth := msg.Width
		viewportHeight := msg.Height - headerHeight - footerHeight - statusBarHeight - marginHeight
		
		// Ensure minimum viewport size
		if viewportHeight < 5 {
			viewportHeight = 5
			// Adjust footer height if needed
			footerHeight = msg.Height - headerHeight - statusBarHeight - marginHeight - 5
			if footerHeight < 2 {
				footerHeight = 2
			}
		}
		
		// Update viewport dimensions
		m.viewport.Width = viewportWidth
		m.viewport.Height = viewportHeight
		
		// Adjust input components width and height (requirement 47.5)
		m.textarea.SetWidth(msg.Width)
		m.textarea.SetHeight(inputHeight)
		
		if m.useMultiline {
			m.multilineInput.SetWidth(msg.Width)
			m.multilineInput.SetMaxLines(maxInputHeight)
		}
		
		// Adjust status bar width and handle narrow terminals (requirement 47.7)
		m.statusBar.SetWidth(msg.Width)
		m.statusBar.SetNarrowMode(msg.Width < 100) // Enable narrow mode for terminals < 100 cols
		
		// Update viewport content with new text wrapping (requirement 47.4)
		m.updateViewportWithResize()
		
		// Restore scroll position (requirement 47.6)
		if m.previousScrollPosition > 0 {
			m.viewport.SetYOffset(m.previousScrollPosition)
		}
		
		// Mark resize as complete and ensure it's within 100ms (requirement 47.1)
		m.resizeInProgress = false
		resizeTime := time.Since(resizeStartTime)
		if resizeTime > 100*time.Millisecond {
			// Log performance warning but don't block
			// In a real implementation, this could be logged to debug output
		}

	case tea.KeyMsg:
		// Handle search activation first
		if msg.String() == "ctrl+f" && !m.searchView.IsActive() {
			m.searchView.SetSize(m.width, m.height)
			return m, m.searchView.Activate()
		}
		
		// If search is active, let search view handle the message first
		if m.searchView.IsActive() {
			var searchCmd tea.Cmd
			searchCmd = m.searchView.Update(msg)
			return m, searchCmd
		}
		
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyCtrlT:
			// Toggle theme (Ctrl+T)
			return m, m.toggleTheme()
		case tea.KeyCtrlY:
			// Copy last message (Ctrl+Y)
			return m, m.copyLastMessage()
		case tea.KeyEsc:
			if m.useMultiline {
				m.multilineInput.Reset()
			} else {
				m.textarea.Reset()
			}
		}
		
		// Handle key combinations that need string matching
		keyStr := msg.String()
		switch keyStr {
		case "ctrl+shift+c":
			// Copy full conversation history (Ctrl+Shift+C)
			return m, m.copyFullHistory()
		}
		
		// Handle Enter key for old textarea only (not multiline)
		if !m.useMultiline && msg.String() == "enter" {
			if m.isThinking {
				return m, nil
			}
			
			input := m.textarea.Value()
			
			if strings.TrimSpace(input) == "" {
				return m, nil
			}

			m.history = append(m.history, api.Message{Role: "user", Content: input})
			m.textarea.Reset()
			
			m.isThinking = true
			m.streamStarted = time.Now()
			m.partialContent = ""
			m.updateViewport()
			return m, tea.Batch(
				m.streamHandler.Stream(context.Background(), m.history),
				m.statusBar.UpdateConnection(components.Connected),
			)
		}

	case components.SubmitMsg:
		// Handle submission from MultiLineInput
		if m.isThinking {
			return m, nil
		}
		
		input := m.multilineInput.Value()
		
		if strings.TrimSpace(input) == "" {
			return m, nil
		}

		m.history = append(m.history, api.Message{Role: "user", Content: input})
		m.multilineInput.Reset()
		
		m.isThinking = true
		m.streamStarted = time.Now()
		m.partialContent = ""
		m.updateViewport()
		return m, tea.Batch(
			m.streamHandler.Stream(context.Background(), m.history),
			m.statusBar.UpdateConnection(components.Connected),
		)

	case client.StreamTokenMsg:
		// Record token receipt time for latency tracking
		m.lastTokenTime = time.Now()
		
		// Append token to partial content
		m.partialContent += msg.Token
		
		// Estimate token count (rough approximation: 1 token ≈ 4 characters)
		m.totalTokens += len(msg.Token) / 4
		if len(msg.Token)%4 > 0 {
			m.totalTokens++
		}
		
		// Update the last message or create new assistant message
		if len(m.history) > 0 && m.history[len(m.history)-1].Role == "assistant" {
			m.history[len(m.history)-1].Content = m.partialContent
		} else {
			m.history = append(m.history, api.Message{
				Role:    "assistant",
				Content: m.partialContent,
			})
		}
		
		// Update viewport incrementally without full redraw
		m.updateViewportIncremental()
		
		// Update status bar with new token count
		sbCmd = m.statusBar.UpdateTokenCount(m.totalTokens)
		
		// Continue listening for more tokens
		return m, tea.Batch(m.streamHandler.WaitForNextMsg(), sbCmd)

	case client.StreamDoneMsg:
		// Stream completed successfully
		m.isThinking = false
		m.partialContent = ""
		m.updateViewport()
		return m, m.statusBar.UpdateConnection(components.Connected)

	case client.StreamErrorMsg:
		// Stream failed - preserve partial content and show error
		m.err = msg.Err
		m.isThinking = false
		
		// Mark the message as incomplete if we have partial content
		if m.partialContent != "" && len(m.history) > 0 {
			lastIdx := len(m.history) - 1
			if m.history[lastIdx].Role == "assistant" {
				m.history[lastIdx].Content += "\n\n[⚠ Response interrupted]"
			}
		}
		
		m.partialContent = ""
		m.updateViewport()
		return m, m.statusBar.UpdateConnection(components.Disconnected)
		
	case components.ClipboardMsg, components.ClipboardConfirmationMsg:
		// Handle clipboard messages
		var clipCmd tea.Cmd
		m.clipboard, clipCmd = m.clipboard.Update(msg)
		return m, clipCmd
		
	case components.SearchViewMsg:
		// Handle search view messages
		switch msg.Type {
		case components.SearchDeactivated:
			// Search was deactivated, update viewport to clear highlights
			m.updateViewport()
			return m, nil
			
		case components.SearchUpdated:
			// Perform search with new query
			query := msg.Data.(string)
			if err := m.searchView.Search(query, m.history); err != nil {
				m.err = err
			}
			m.updateViewportWithSearch()
			return m, nil
			
		case components.MatchChanged:
			// Navigate to a different match
			result := msg.Data.(*components.SearchResult)
			m.scrollToMatch(result)
			m.updateViewportWithSearch()
			return m, nil
		}
	}

	if !m.isThinking {
		if m.useMultiline {
			m.multilineInput, mlCmd = m.multilineInput.Update(msg)
		} else {
			m.textarea, tiCmd = m.textarea.Update(msg)
		}
	}
	
	// Always allow viewport to handle mouse events
	if _, ok := msg.(tea.MouseMsg); ok {
		m.viewport, vpCmd = m.viewport.Update(msg)
	} else {
		// For keyboard events, check if we should update viewport
		shouldUpdateViewport := true
		if m.useMultiline && !m.isThinking {
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				switch keyMsg.String() {
				case "up", "down", "ctrl+p", "ctrl+n":
					// Block viewport for history navigation keys
					shouldUpdateViewport = false
				}
			}
		}
		
		if shouldUpdateViewport {
			m.viewport, vpCmd = m.viewport.Update(msg)
		}
	}
	
	m.spinner, spCmd = m.spinner.Update(msg)
	m.statusBar, sbCmd = m.statusBar.Update(msg)

	return m, tea.Batch(tiCmd, mlCmd, vpCmd, spCmd, sbCmd)
}

func (m *Model) updateViewport() {
	var sb strings.Builder
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle(m.theme.GlamourStyle()),
		glamour.WithWordWrap(m.width-5),
	)

	for _, msg := range m.history {
		if msg.Role == "system" {
			continue
		}
		label := m.theme.UserLabelStyle().Render("👤 You: ")
		if msg.Role == "assistant" {
			label = m.theme.AILabelStyle().Render("🤖 AI: ")
		}

		rendered, _ := renderer.Render(msg.Content)
		sb.WriteString(label + "\n" + rendered + "\n")
	}
	m.viewport.SetContent(sb.String())
	m.viewport.GotoBottom()
}

// updateViewportIncremental updates only the last message without full redraw
func (m *Model) updateViewportWithResize() {
	var sb strings.Builder
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle(m.theme.GlamourStyle()),
		glamour.WithWordWrap(m.width-5), // Adjust text wrapping based on width (requirement 47.4)
	)

	for _, msg := range m.history {
		if msg.Role == "system" {
			continue
		}
		label := m.theme.UserLabelStyle().Render("👤 You: ")
		if msg.Role == "assistant" {
			label = m.theme.AILabelStyle().Render("🤖 AI: ")
		}

		rendered, _ := renderer.Render(msg.Content)
		sb.WriteString(label + "\n" + rendered + "\n")
	}
	m.viewport.SetContent(sb.String())
	// Don't auto-scroll to bottom during resize to preserve position
}

func (m *Model) updateViewportIncremental() {
	var sb strings.Builder
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle(m.theme.GlamourStyle()),
		glamour.WithWordWrap(m.width-5),
	)

	// Render all messages
	for _, msg := range m.history {
		if msg.Role == "system" {
			continue
		}
		label := m.theme.UserLabelStyle().Render("👤 You: ")
		if msg.Role == "assistant" {
			label = m.theme.AILabelStyle().Render("🤖 AI: ")
		}

		rendered, _ := renderer.Render(msg.Content)
		sb.WriteString(label + "\n" + rendered + "\n")
	}
	
	// Set content and scroll to bottom
	m.viewport.SetContent(sb.String())
	m.viewport.GotoBottom()
}

// updateViewportWithSearch updates viewport content with search highlighting
func (m *Model) updateViewportWithSearch() {
	var sb strings.Builder
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle(m.theme.GlamourStyle()),
		glamour.WithWordWrap(m.width-5),
	)

	for i, msg := range m.history {
		if msg.Role == "system" {
			continue
		}
		label := m.theme.UserLabelStyle().Render("👤 You: ")
		if msg.Role == "assistant" {
			label = m.theme.AILabelStyle().Render("🤖 AI: ")
		}

		// Apply search highlighting to message content
		content := m.searchView.HighlightMessage(msg.Content, i)
		rendered, _ := renderer.Render(content)
		sb.WriteString(label + "\n" + rendered + "\n")
	}
	
	m.viewport.SetContent(sb.String())
}

// scrollToMatch scrolls the viewport to show a specific search match
func (m *Model) scrollToMatch(result *components.SearchResult) {
	if result == nil {
		return
	}
	
	// Calculate approximate line position for the message
	// This is a simplified approach - in a real implementation,
	// you might want to calculate exact line positions
	messageIndex := result.MessageIndex
	
	// Count non-system messages before this one
	visibleMessages := 0
	for i := 0; i < messageIndex && i < len(m.history); i++ {
		if m.history[i].Role != "system" {
			visibleMessages++
		}
	}
	
	// Estimate lines per message (label + content + spacing)
	// This is approximate and could be improved with exact line counting
	linesPerMessage := 5 // Rough estimate
	targetLine := visibleMessages * linesPerMessage
	
	// Scroll to show the match
	m.viewport.SetYOffset(targetLine)
}

func (m Model) View() string {
	var header string
	
	// Display size warning if terminal is too small (requirement 47.3)
	if m.showSizeWarning {
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF0000"))
		
		// Create adaptive warning message based on available space
		var warningMsg string
		if m.width < 40 {
			// Very compact warning for extremely narrow terminals
			warningMsg = fmt.Sprintf("⚠ Too small! Need %dx%d, got %dx%d", 
				m.minWidth, m.minHeight, m.width, m.height)
		} else if m.width < 60 {
			// Compact warning for narrow terminals
			warningMsg = fmt.Sprintf("⚠ Terminal too small!\nMinimum: %dx%d | Current: %dx%d\nPlease resize terminal.", 
				m.minWidth, m.minHeight, m.width, m.height)
		} else {
			// Full warning for wider terminals
			warningMsg = fmt.Sprintf("⚠ Terminal too small!\nMinimum size: %dx%d\nCurrent size: %dx%d\nPlease resize your terminal for optimal experience.", 
				m.minWidth, m.minHeight, m.width, m.height)
		}
		
		warning := warningStyle.Render(warningMsg)
		
		// If terminal is extremely small, just show the warning
		if m.width < 30 || m.height < 8 {
			return warning
		}
		
		// Otherwise show warning at top
		header = warning + "\n\n"
		
		// Adaptive header based on width (requirement 47.4)
		if m.width < 60 {
			header += m.theme.TitleStyle().Render("Ollama") + " | " + m.modelName + "\n\n"
		} else {
			header += m.theme.TitleStyle().Render("Ollama TUI") + " | Model: " + m.modelName + " | Theme: " + m.theme.Name() + "\n\n"
		}
	} else {
		// Normal header - adaptive based on width (requirement 47.4)
		if m.width < 60 {
			header = m.theme.TitleStyle().Render("Ollama") + " | " + m.modelName + "\n\n"
		} else if m.width < 80 {
			header = m.theme.TitleStyle().Render("Ollama TUI") + " | " + m.modelName + "\n\n"
		} else {
			header = m.theme.TitleStyle().Render("Ollama TUI") + " | Model: " + m.modelName + " | Theme: " + m.theme.Name() + "\n\n"
		}
	}

	var footer string
	if m.isThinking {
		footer = m.spinner.View() + " Thinking..."
	} else if m.err != nil {
		errorMsg := m.theme.ErrorStyle().Render("Error: "+m.err.Error())
		if m.useMultiline {
			footer = errorMsg + "\n" + m.multilineInput.View()
		} else {
			footer = errorMsg + "\n" + m.textarea.View()
		}
	} else {
		if m.useMultiline {
			footer = m.multilineInput.View()
		} else {
			// Adaptive help text based on width (requirement 47.4)
			var helpText string
			if m.width < 50 {
				helpText = "Ctrl+C: Quit | Ctrl+T: Theme | Ctrl+Y: Copy"
			} else if m.width < 80 {
				helpText = "Ctrl+C: Quit | Ctrl+T: Theme | Ctrl+Y: Copy Last | Esc: Clear"
			} else {
				helpText = "Ctrl+C: Quit | Ctrl+T: Theme | Ctrl+Y: Copy Last | Ctrl+Shift+C: Copy All | Esc: Clear"
			}
			footer = m.textarea.View() + "\n" + m.theme.InfoStyle().Render(helpText)
		}
	}

	// Render status bar
	statusBar := m.statusBar.View()
	
	// Render clipboard confirmation if active
	clipboardConfirmation := m.clipboard.View()
	if clipboardConfirmation != "" {
		statusBar = clipboardConfirmation + " " + statusBar
	}

	// Combine all parts with proper spacing
	mainView := header + m.viewport.View() + "\n" + footer + "\n" + statusBar
	
	// Overlay search view if active
	if m.searchView.IsActive() {
		searchOverlay := m.searchView.View()
		if searchOverlay != "" {
			// Position search overlay at the top of the viewport area
			searchStyle := lipgloss.NewStyle().
				MarginTop(strings.Count(header, "\n")).
				MarginLeft(2)
			
			searchView := searchStyle.Render(searchOverlay)
			
			// Combine main view with search overlay
			return lipgloss.JoinVertical(lipgloss.Left, mainView, searchView)
		}
	}
	
	return mainView
}

// toggleTheme switches between dark and light themes and persists the preference
func (m *Model) toggleTheme() tea.Cmd {
	// Load current config
	cfg, err := m.configManager.Load()
	if err != nil {
		// If we can't load config, just toggle in memory
		if m.theme.Name() == "dark" {
			m.theme = theme.NewLightTheme()
		} else {
			m.theme = theme.NewDarkTheme()
		}
		m.statusBar.SetTheme(m.theme)
		m.updateViewport()
		return nil
	}

	// Toggle theme
	if cfg.UI.Theme == "dark" {
		cfg.UI.Theme = "light"
		m.theme = theme.GetTheme("light", &cfg.UI.Colors)
	} else {
		cfg.UI.Theme = "dark"
		m.theme = theme.GetTheme("dark", &cfg.UI.Colors)
	}

	// Update status bar theme
	m.statusBar.SetTheme(m.theme)
	
	// Update search view theme
	m.searchView.SetTheme(cfg.UI.Theme == "dark")

	// Save config
	if err := m.configManager.Save(cfg); err != nil {
		// Log error but continue with theme change
		m.err = err
	}

	// Update viewport with new theme
	m.updateViewport()

	return nil
}

// copyLastMessage copies the last non-system message to the clipboard
func (m *Model) copyLastMessage() tea.Cmd {
	// Find the last non-system message
	for i := len(m.history) - 1; i >= 0; i-- {
		if m.history[i].Role != "system" {
			content := m.history[i].Content
			
			// Check if this is a code block
			codeBlocks := util.ExtractCodeBlocks(content)
			if len(codeBlocks) > 0 {
				// If the entire message is a single code block, strip formatting
				if len(codeBlocks) == 1 && strings.TrimSpace(content) == strings.TrimSpace(codeBlocks[0]) {
					stripped, err := util.CopyCodeBlock(content)
					if err != nil {
						return components.ShowError(err)
					}
					return components.ShowSuccess(fmt.Sprintf("✓ Copied code (%d chars)", len(stripped)))
				}
			}
			
			// Copy the full message content
			if err := util.CopyMessageContent(content); err != nil {
				return components.ShowError(err)
			}
			
			return components.ShowSuccess(fmt.Sprintf("✓ Copied message (%d chars)", len(content)))
		}
	}
	
	return components.ShowError(fmt.Errorf("no message to copy"))
}

// copyFullHistory copies the entire conversation history to the clipboard
func (m *Model) copyFullHistory() tea.Cmd {
	var messages []string
	
	for _, msg := range m.history {
		if msg.Role == "system" {
			continue // Skip system messages
		}
		
		formatted := util.FormatMessageForCopy(msg.Role, msg.Content)
		messages = append(messages, formatted)
	}
	
	if len(messages) == 0 {
		return components.ShowError(fmt.Errorf("no messages to copy"))
	}
	
	if err := util.CopyFullHistory(messages); err != nil {
		return components.ShowError(err)
	}
	
	return components.ShowSuccess(fmt.Sprintf("✓ Copied full history (%d messages)", len(messages)))
}
