package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/gagan-devv/ollama-go/internal/config"
)

// DarkTheme implements a dark theme with eye-strain reducing colors
type DarkTheme struct {
	BaseTheme
}

// NewDarkTheme creates a new dark theme instance
func NewDarkTheme() *DarkTheme {
	return &DarkTheme{
		BaseTheme: BaseTheme{
			name:         "dark",
			glamourStyle: "dark",
			colors: config.ColorScheme{
				UserMessage: "#5FAFD7", // Soft blue
				AIMessage:   "#FFD787", // Warm yellow
				Background:  "#1E1E1E", // Dark gray
				Border:      "#3E3E3E", // Medium gray
				StatusBar:   "#2E2E2E", // Slightly lighter than background
			},
		},
	}
}

func (t *DarkTheme) Name() string {
	return t.name
}

func (t *DarkTheme) UserMessageColor() lipgloss.Color {
	return lipgloss.Color(t.colors.UserMessage)
}

func (t *DarkTheme) UserLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.UserMessageColor()).
		Bold(true)
}

func (t *DarkTheme) AIMessageColor() lipgloss.Color {
	return lipgloss.Color(t.colors.AIMessage)
}

func (t *DarkTheme) AILabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.AIMessageColor()).
		Bold(true)
}

func (t *DarkTheme) BackgroundColor() lipgloss.Color {
	return lipgloss.Color(t.colors.Background)
}

func (t *DarkTheme) BorderColor() lipgloss.Color {
	return lipgloss.Color(t.colors.Border)
}

func (t *DarkTheme) StatusBarColor() lipgloss.Color {
	return lipgloss.Color(t.colors.StatusBar)
}

func (t *DarkTheme) TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		MarginLeft(2).
		Bold(true).
		Foreground(lipgloss.Color("5")) // Purple accent
}

func (t *DarkTheme) InfoStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")). // Dim gray
		MarginLeft(2)
}

func (t *DarkTheme) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")) // Bright red
}

func (t *DarkTheme) GlamourStyle() string {
	return t.glamourStyle
}
