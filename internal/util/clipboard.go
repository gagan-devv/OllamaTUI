package util

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
)

// CopyToClipboard copies the given content to the system clipboard.
// Returns an error if clipboard access fails.
func CopyToClipboard(content string) error {
	if err := clipboard.WriteAll(content); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}
	return nil
}

// CopyMessageContent copies a message's content to the clipboard.
// This preserves the exact message content without modification.
func CopyMessageContent(content string) error {
	return CopyToClipboard(content)
}

// CopyCodeBlock copies a code block to the clipboard, stripping markdown
// formatting markers (backticks and language tags).
func CopyCodeBlock(codeBlock string) (string, error) {
	stripped := StripCodeBlockFormatting(codeBlock)
	if err := CopyToClipboard(stripped); err != nil {
		return "", err
	}
	return stripped, nil
}

// StripCodeBlockFormatting removes markdown code block formatting markers.
// It removes:
// - Opening backticks with optional language identifier (```language)
// - Closing backticks (```)
// - Leading/trailing whitespace
func StripCodeBlockFormatting(codeBlock string) string {
	// Remove opening code fence with optional language identifier
	// Pattern: ```language or ``` at the start
	openingPattern := regexp.MustCompile(`^` + "```" + `[a-zA-Z0-9_+-]*\s*\n?`)
	result := openingPattern.ReplaceAllString(codeBlock, "")
	
	// Remove closing code fence
	// Pattern: ``` at the end (possibly with trailing whitespace)
	closingPattern := regexp.MustCompile(`\n?` + "```" + `\s*$`)
	result = closingPattern.ReplaceAllString(result, "")
	
	// Trim any remaining leading/trailing whitespace
	result = strings.TrimSpace(result)
	
	return result
}

// CopyFullHistory copies the entire conversation history to the clipboard.
// Each message is formatted with a role prefix and separated by newlines.
func CopyFullHistory(messages []string) error {
	fullHistory := strings.Join(messages, "\n\n")
	return CopyToClipboard(fullHistory)
}

// ExtractCodeBlocks finds all code blocks in the given content and returns them.
// This is useful for identifying code blocks that can be copied separately.
func ExtractCodeBlocks(content string) []string {
	// Pattern to match code blocks: ```language\ncode\n```
	pattern := regexp.MustCompile("```[a-zA-Z0-9_+-]*\\s*\\n([\\s\\S]*?)\\n```")
	matches := pattern.FindAllString(content, -1)
	return matches
}

// FormatMessageForCopy formats a message for clipboard copying.
// It includes the role (User/AI) and the content.
func FormatMessageForCopy(role, content string) string {
	var prefix string
	switch role {
	case "user":
		prefix = "👤 You:"
	case "assistant":
		prefix = "🤖 AI:"
	case "system":
		prefix = "⚙️ System:"
	default:
		prefix = role + ":"
	}
	
	return fmt.Sprintf("%s\n%s", prefix, content)
}
