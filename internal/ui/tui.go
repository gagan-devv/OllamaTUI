package ui

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/gagan-devv/ollama-go/internal/client"
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
	spinner        spinner.Model
	history        []api.Message
	client         *api.Client
	streamHandler  *client.StreamHandler
	modelName      string
	isThinking     bool
	err            error
	width          int
	height         int
	streamStarted  time.Time
	lastTokenTime  time.Time
	partialContent string // Buffer for current streaming message
}

func InitialModel(apiClient *api.Client, modelName, system string) Model {
	ta := textarea.New()
	ta.Placeholder = "Ask me anything..."
	ta.SetHeight(3)
	ta.Focus()

	vp := viewport.New(80, 20) // Give it a default size
	vp.SetContent("Ready.")

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		textarea:      ta,
		viewport:      vp,
		spinner:       s,
		client:        apiClient,
		streamHandler: client.NewStreamHandler(apiClient, modelName),
		modelName:     modelName,
		width:         80, // Fallback width
		height:        24, // Fallback height
		history:       []api.Message{{Role: "system", Content: system}},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		spCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 9 // Space for header/footer
		m.textarea.SetWidth(msg.Width)
		m.updateViewport()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
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
			return m, m.streamHandler.Stream(context.Background(), m.history)

		case tea.KeyEsc:
			m.textarea.Reset()
		}

	case client.StreamTokenMsg:
		// Record token receipt time for latency tracking
		m.lastTokenTime = time.Now()
		
		// Append token to partial content
		m.partialContent += msg.Token
		
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
		
		// Continue listening for more tokens
		return m, m.streamHandler.WaitForNextMsg()

	case client.StreamDoneMsg:
		// Stream completed successfully
		m.isThinking = false
		m.partialContent = ""
		m.updateViewport()
		return m, nil

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
		return m, nil
	}

	if !m.isThinking {
		m.textarea, tiCmd = m.textarea.Update(msg)
	}
	m.viewport, vpCmd = m.viewport.Update(msg)
	m.spinner, spCmd = m.spinner.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd, spCmd)
}

func (m *Model) updateViewport() {
	var sb strings.Builder
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(m.width-5),
	)

	for _, msg := range m.history {
		if msg.Role == "system" {
			continue
		}
		label := userStyle.Render("👤 You: ")
		if msg.Role == "assistant" {
			label = aiStyle.Render("🤖 AI: ")
		}

		rendered, _ := renderer.Render(msg.Content)
		sb.WriteString(label + "\n" + rendered + "\n")
	}
	m.viewport.SetContent(sb.String())
	m.viewport.GotoBottom()
}

// updateViewportIncremental updates only the last message without full redraw
func (m *Model) updateViewportIncremental() {
	var sb strings.Builder
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(m.width-5),
	)

	// Render all messages
	for _, msg := range m.history {
		if msg.Role == "system" {
			continue
		}
		label := userStyle.Render("👤 You: ")
		if msg.Role == "assistant" {
			label = aiStyle.Render("🤖 AI: ")
		}

		rendered, _ := renderer.Render(msg.Content)
		sb.WriteString(label + "\n" + rendered + "\n")
	}
	
	// Set content and scroll to bottom
	m.viewport.SetContent(sb.String())
	m.viewport.GotoBottom()
}

func (m Model) View() string {
	header := titleStyle.Render("Ollama TUI") + " | Model: " + m.modelName + "\n\n"

	var footer string
	if m.isThinking {
		footer = m.spinner.View() + " Thinking..."
	} else if m.err != nil {
		footer = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("Error: "+m.err.Error()) + "\n" + m.textarea.View()
	} else {
		footer = m.textarea.View()
	}

	return header + m.viewport.View() + "\n" + footer + "\n" + infoStyle.Render("Ctrl+C: Quit | Esc: Clear")
}
