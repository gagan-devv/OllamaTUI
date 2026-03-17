package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/gagan-devv/ollama-go/internal/config"
)

// Theme defines the interface for UI themes
type Theme interface {
	// Name returns the theme name
	Name() string
	
	// User message styling
	UserMessageColor() lipgloss.Color
	UserLabelStyle() lipgloss.Style
	
	// AI message styling
	AIMessageColor() lipgloss.Color
	AILabelStyle() lipgloss.Style
	
	// UI element styling
	BackgroundColor() lipgloss.Color
	BorderColor() lipgloss.Color
	StatusBarColor() lipgloss.Color
	TitleStyle() lipgloss.Style
	InfoStyle() lipgloss.Style
	ErrorStyle() lipgloss.Style
	
	// Glamour style for markdown rendering
	GlamourStyle() string
}

// BaseTheme provides common theme functionality
type BaseTheme struct {
	name         string
	glamourStyle string
	colors       config.ColorScheme
}

// ApplyColorScheme creates a new theme with custom colors
func (t *BaseTheme) ApplyColorScheme(scheme config.ColorScheme) {
	if scheme.UserMessage != "" {
		t.colors.UserMessage = scheme.UserMessage
	}
	if scheme.AIMessage != "" {
		t.colors.AIMessage = scheme.AIMessage
	}
	if scheme.Background != "" {
		t.colors.Background = scheme.Background
	}
	if scheme.Border != "" {
		t.colors.Border = scheme.Border
	}
	if scheme.StatusBar != "" {
		t.colors.StatusBar = scheme.StatusBar
	}
}

// GetTheme returns a theme by name with optional color customization
func GetTheme(name string, customColors *config.ColorScheme) Theme {
	var t Theme
	
	switch name {
	case "light":
		t = NewLightTheme()
	case "dark":
		fallthrough
	default:
		t = NewDarkTheme()
	}
	
	// Apply custom colors if provided
	if customColors != nil {
		if bt, ok := t.(*DarkTheme); ok {
			bt.BaseTheme.ApplyColorScheme(*customColors)
		} else if bt, ok := t.(*LightTheme); ok {
			bt.BaseTheme.ApplyColorScheme(*customColors)
		}
	}
	
	return t
}
