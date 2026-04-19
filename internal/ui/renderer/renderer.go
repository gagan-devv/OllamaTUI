package renderer

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

// Renderer handles markdown rendering with syntax highlighting
type Renderer struct {
	glamourStyle string
	width        int
}

// NewRenderer creates a new markdown renderer
func NewRenderer(glamourStyle string, width int) *Renderer {
	return &Renderer{
		glamourStyle: glamourStyle,
		width:        width,
	}
}

// Render renders markdown content with syntax highlighting
// Supports 20+ programming languages via Chroma (built into Glamour)
// Auto-detects language when not specified
// Uses theme-aware colors based on glamourStyle
func Render(content string, glamourStyle string, width int) (string, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(glamourStyle),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", err
	}

	return renderer.Render(content)
}

// RenderWithRenderer renders markdown using an existing renderer instance
func (r *Renderer) Render(content string) (string, error) {
	return Render(content, r.glamourStyle, r.width)
}

// HasSyntaxHighlighting checks if rendered output contains ANSI color codes
// indicating syntax highlighting was applied
func HasSyntaxHighlighting(rendered string) bool {
	// ANSI color codes start with ESC[ (escape sequence)
	// Common patterns: \x1b[, \033[, or \u001b[
	return strings.Contains(rendered, "\x1b[") || 
	       strings.Contains(rendered, "\033[") ||
	       strings.Contains(rendered, "\u001b[")
}

// ExtractCodeBlocks extracts code blocks from markdown content
func ExtractCodeBlocks(markdown string) []string {
	var blocks []string
	lines := strings.Split(markdown, "\n")
	
	var inBlock bool
	var currentBlock strings.Builder
	
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inBlock {
				// End of block
				blocks = append(blocks, currentBlock.String())
				currentBlock.Reset()
				inBlock = false
			} else {
				// Start of block
				inBlock = true
			}
		} else if inBlock {
			currentBlock.WriteString(line)
			currentBlock.WriteString("\n")
		}
	}
	
	return blocks
}
