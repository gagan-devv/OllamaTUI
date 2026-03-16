package session

import (
	"time"
)

// Session represents a chat session with its configuration and message history
type Session struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Messages     []Message              `json:"messages"`
	Model        string                 `json:"model"`
	SystemPrompt string                 `json:"system_prompt"`
	Parameters   ModelParameters        `json:"parameters"`
	Tags         []string               `json:"tags"`
	Bookmarks    []int                  `json:"bookmarks"` // Message indices
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Message represents a single message in a conversation
type Message struct {
	ID          string          `json:"id"`
	Role        string          `json:"role"` // "user", "assistant", "system"
	Content     string          `json:"content"`
	Timestamp   time.Time       `json:"timestamp"`
	Model       string          `json:"model"` // Track which model generated this
	Metrics     *MessageMetrics `json:"metrics,omitempty"`
	Attachments []Attachment    `json:"attachments,omitempty"`
}

// MessageMetrics contains performance metrics for a message
type MessageMetrics struct {
	TokenCount   int           `json:"token_count"`
	ResponseTime time.Duration `json:"response_time"`
	TokensPerSec float64       `json:"tokens_per_sec"`
}

// Attachment represents a file or image attached to a message
type Attachment struct {
	Type     string `json:"type"` // "file", "image"
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	Data     []byte `json:"data,omitempty"` // For images
}

// ModelParameters contains model configuration parameters
type ModelParameters struct {
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
	TopK        int     `json:"top_k"`
}

// SessionConfig contains configuration for creating a new session
type SessionConfig struct {
	Model        string
	SystemPrompt string
	Parameters   ModelParameters
}

// SessionMetadata contains summary information about a session
type SessionMetadata struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Model        string    `json:"model"`
	Tags         []string  `json:"tags"`
	MessageCount int       `json:"message_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
