package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gagan-devv/ollama-go/internal/session"
)

// Exporter handles exporting sessions to various formats
type Exporter struct {
	testFolderManager *TestFolderManager
}

// NewExporter creates a new exporter
func NewExporter(testFolderManager *TestFolderManager) *Exporter {
	return &Exporter{
		testFolderManager: testFolderManager,
	}
}

// ExportToMarkdown exports a session to Markdown format
func (e *Exporter) ExportToMarkdown(sess *session.Session) (string, error) {
	// Ensure exports directory exists
	if _, err := e.testFolderManager.EnsureTestFolder(); err != nil {
		return "", fmt.Errorf("ensure test folder: %w", err)
	}

	exportsPath, err := e.testFolderManager.GetSubdirectoryPath("exports")
	if err != nil {
		return "", fmt.Errorf("get exports path: %w", err)
	}

	// Create filename
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.md", sess.Name, timestamp)
	filename = sanitizeFilename(filename)
	filePath := filepath.Join(exportsPath, filename)

	// Build markdown content
	var content strings.Builder
	
	// Header
	content.WriteString(fmt.Sprintf("# %s\n\n", sess.Name))
	content.WriteString(fmt.Sprintf("**Session ID:** %s\n\n", sess.ID))
	content.WriteString(fmt.Sprintf("**Model:** %s\n\n", sess.Model))
	content.WriteString(fmt.Sprintf("**Created:** %s\n\n", sess.CreatedAt.Format(time.RFC3339)))
	content.WriteString(fmt.Sprintf("**Updated:** %s\n\n", sess.UpdatedAt.Format(time.RFC3339)))
	
	if len(sess.Tags) > 0 {
		content.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(sess.Tags, ", ")))
	}
	
	content.WriteString("---\n\n")

	// Messages
	for i, msg := range sess.Messages {
		// Message header
		role := strings.Title(msg.Role)
		content.WriteString(fmt.Sprintf("## Message %d - %s\n\n", i+1, role))
		content.WriteString(fmt.Sprintf("**Timestamp:** %s\n\n", msg.Timestamp.Format(time.RFC3339)))
		
		if msg.Model != "" {
			content.WriteString(fmt.Sprintf("**Model:** %s\n\n", msg.Model))
		}

		// Message content
		content.WriteString(msg.Content)
		content.WriteString("\n\n")

		// Metrics if available
		if msg.Metrics != nil {
			content.WriteString(fmt.Sprintf("*Response time: %.2fs | Tokens: %d | Speed: %.2f tokens/s*\n\n",
				msg.Metrics.ResponseTime.Seconds(),
				msg.Metrics.TokenCount,
				msg.Metrics.TokensPerSec))
		}

		content.WriteString("---\n\n")
	}

	// Write to file
	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		return "", fmt.Errorf("write markdown file: %w", err)
	}

	return filePath, nil
}

// ExportToPlainText exports a session to plain text format
func (e *Exporter) ExportToPlainText(sess *session.Session) (string, error) {
	// Ensure exports directory exists
	if _, err := e.testFolderManager.EnsureTestFolder(); err != nil {
		return "", fmt.Errorf("ensure test folder: %w", err)
	}

	exportsPath, err := e.testFolderManager.GetSubdirectoryPath("exports")
	if err != nil {
		return "", fmt.Errorf("get exports path: %w", err)
	}

	// Create filename
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.txt", sess.Name, timestamp)
	filename = sanitizeFilename(filename)
	filePath := filepath.Join(exportsPath, filename)

	// Build plain text content
	var content strings.Builder
	
	// Header
	content.WriteString(fmt.Sprintf("Session: %s\n", sess.Name))
	content.WriteString(fmt.Sprintf("ID: %s\n", sess.ID))
	content.WriteString(fmt.Sprintf("Model: %s\n", sess.Model))
	content.WriteString(fmt.Sprintf("Created: %s\n", sess.CreatedAt.Format(time.RFC3339)))
	content.WriteString(fmt.Sprintf("Updated: %s\n", sess.UpdatedAt.Format(time.RFC3339)))
	
	if len(sess.Tags) > 0 {
		content.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(sess.Tags, ", ")))
	}
	
	content.WriteString("\n")
	content.WriteString(strings.Repeat("=", 80))
	content.WriteString("\n\n")

	// Messages
	for i, msg := range sess.Messages {
		// Message header
		role := strings.ToUpper(msg.Role)
		content.WriteString(fmt.Sprintf("[%s - %s]\n", role, msg.Timestamp.Format(time.RFC3339)))
		
		if msg.Model != "" {
			content.WriteString(fmt.Sprintf("Model: %s\n", msg.Model))
		}

		content.WriteString("\n")
		content.WriteString(msg.Content)
		content.WriteString("\n")

		// Metrics if available
		if msg.Metrics != nil {
			content.WriteString(fmt.Sprintf("\n(Response time: %.2fs | Tokens: %d | Speed: %.2f tokens/s)\n",
				msg.Metrics.ResponseTime.Seconds(),
				msg.Metrics.TokenCount,
				msg.Metrics.TokensPerSec))
		}

		if i < len(sess.Messages)-1 {
			content.WriteString("\n")
			content.WriteString(strings.Repeat("-", 80))
			content.WriteString("\n\n")
		}
	}

	// Write to file
	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		return "", fmt.Errorf("write text file: %w", err)
	}

	return filePath, nil
}

// ExportToJSON exports a session to JSON format
func (e *Exporter) ExportToJSON(sess *session.Session) (string, error) {
	// Ensure exports directory exists
	if _, err := e.testFolderManager.EnsureTestFolder(); err != nil {
		return "", fmt.Errorf("ensure test folder: %w", err)
	}

	exportsPath, err := e.testFolderManager.GetSubdirectoryPath("exports")
	if err != nil {
		return "", fmt.Errorf("get exports path: %w", err)
	}

	// Create filename
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.json", sess.Name, timestamp)
	filename = sanitizeFilename(filename)
	filePath := filepath.Join(exportsPath, filename)

	// Marshal session to JSON with indentation
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal session to JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("write JSON file: %w", err)
	}

	return filePath, nil
}

// sanitizeFilename removes or replaces characters that are invalid in filenames
func sanitizeFilename(filename string) string {
	// Replace invalid characters with underscores
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
