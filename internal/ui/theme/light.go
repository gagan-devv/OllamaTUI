package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/gagan-devv/ollama-go/internal/config"
)

// LightTheme implements a light theme with high-contrast colors
type LightTheme struct {
	BaseTheme
}

// NewLightTheme creates a new light theme instance
func NewLightTheme() *LightTheme {
	return &LightTheme{
		BaseTheme: BaseTheme{
			name:         "light",
			glamourStyle: "light",
			colors: config.ColorScheme{
				UserMessage: "#0066CC", // Deep blue for high contrast
				AIMessage:   "#CC6600", // Deep orange for high contrast
				Background:  "#FFFFFF", // White
				Border:      "#CCCCCC", // Light gray
				StatusBar:   "#F5F5F5", // Very light gray
			},
		},
	}
}

func (t *LightTheme) Name() string {
	return t.name
}

func (t *LightTheme) UserMessageColor() lipgloss.Color {
	return lipgloss.Color(t.colors.UserMessage)
}

func (t *LightTheme) UserLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.UserMessageColor()).
		Bold(true)
}

func (t *LightTheme) AIMessageColor() lipgloss.Color {
	return lipgloss.Color(t.colors.AIMessage)
}

func (t *LightTheme) AILabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.AIMessageColor()).
		Bold(true)
}

func (t *LightTheme) BackgroundColor() lipgloss.Color {
	return lipgloss.Color(t.colors.Background)
}

func (t *LightTheme) BorderColor() lipgloss.Color {
	return lipgloss.Color(t.colors.Border)
}

func (t *LightTheme) StatusBarColor() lipgloss.Color {
	return lipgloss.Color(t.colors.StatusBar)
}

func (t *LightTheme) TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		MarginLeft(2).
		Bold(true).
		Foreground(lipgloss.Color("5")) // Purple accent
}

func (t *LightTheme) InfoStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")). // Dark gray for contrast
		MarginLeft(2)
}

func (t *LightTheme) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("1")) // Dark red for contrast
}

func (t *LightTheme) GlamourStyle() string {
	return t.glamourStyle
}
