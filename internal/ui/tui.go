package ui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/ollama/ollama/api"
)

var (
	titleStyle = lipgloss.NewStyle().MarginLeft(2).Bold(true).Foreground(lipgloss.Color("5"))
	userStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	aiStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginLeft(2)
)

// Messages
type streamMsg string
type errMsg error
type doneMsg bool

type Model struct {
	viewport    viewport.Model
	textarea    textarea.Model
	spinner     spinner.Model
	history     []api.Message
	client      *api.Client
	modelName   string
	isThinking  bool
	err         error
	width       int
	height      int
}

func InitialModel(client *api.Client, modelName, system string) Model {
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
		textarea:  ta,
		viewport:  vp,
		spinner:   s,
		client:    client,
		modelName: modelName,
		width:     80, // Fallback width
		height:    24, // Fallback height
		history:   []api.Message{{Role: "system", Content: system}},
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
			m.updateViewport()
			return m, m.fetchAIResponse()

		case tea.KeyEsc:
			m.textarea.Reset()
		}

	case streamMsg:
		last := len(m.history) - 1
		if m.history[last].Role == "assistant" {
			m.history[last].Content += string(msg)
		} else {
			m.history = append(m.history, api.Message{Role: "assistant", Content: string(msg)})
		}
		m.updateViewport()

	case doneMsg:
		m.isThinking = false
		return m, nil

	case errMsg:
		m.err = msg
		m.isThinking = false
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
		if msg.Role == "system" { continue }
		label := userStyle.Render("👤 You: ")
		if msg.Role == "assistant" { label = aiStyle.Render("🤖 AI: ") }
		
		rendered, _ := renderer.Render(msg.Content)
		sb.WriteString(label + "\n" + rendered + "\n")
	}
	m.viewport.SetContent(sb.String())
	m.viewport.GotoBottom()
}

func (m Model) View() string {
	header := titleStyle.Render("Ollama TUI") + " | Model: " + m.modelName + "\n\n"
	
	var footer string
	if m.isThinking {
		footer = m.spinner.View() + " Thinking..."
	} else {
		footer = m.textarea.View()
	}

	return header + m.viewport.View() + "\n" + footer + "\n" + infoStyle.Render("Ctrl+C: Quit | Esc: Clear")
}

func (m Model) fetchAIResponse() tea.Cmd {
	return func() tea.Msg {
		// Note: To do true word-by-word streaming in Bubble Tea, 
		// you usually need to pass the program handle. 
		// For now, this collects and updates via Update loop.
		var full strings.Builder
		err := m.client.Chat(context.Background(), &api.ChatRequest{
			Model:    m.modelName,
			Messages: m.history,
		}, func(resp api.ChatResponse) error {
			full.WriteString(resp.Message.Content)
			return nil
		})
		if err != nil { return errMsg(err) }
		return doneMsg(true)
	}
}