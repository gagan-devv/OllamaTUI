package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gagan-bhullar-tech/ollama-go/internal/session"
)

// SessionStorage implements session persistence to the filesystem
type SessionStorage struct {
	testFolderManager *TestFolderManager
}

// NewSessionStorage creates a new session storage
func NewSessionStorage(testFolderManager *TestFolderManager) *SessionStorage {
	return &SessionStorage{
		testFolderManager: testFolderManager,
	}
}

// SaveSession saves a session to the history directory as JSON
func (s *SessionStorage) SaveSession(sess *session.Session) error {
	// Ensure test folder exists
	if _, err := s.testFolderManager.EnsureTestFolder(); err != nil {
		return fmt.Errorf("ensure test folder: %w", err)
	}

	// Get history directory path
	historyPath, err := s.testFolderManager.GetSubdirectoryPath("history")
	if err != nil {
		return fmt.Errorf("get history path: %w", err)
	}

	// Create session file path
	filename := fmt.Sprintf("%s.json", sess.ID)
	filePath := filepath.Join(historyPath, filename)

	// Marshal session to JSON
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write session file: %w", err)
	}

	return nil
}

// LoadSession loads a session from the history directory
func (s *SessionStorage) LoadSession(id string) (*session.Session, error) {
	// Get history directory path
	historyPath, err := s.testFolderManager.GetSubdirectoryPath("history")
	if err != nil {
		return nil, fmt.Errorf("get history path: %w", err)
	}

	// Create session file path
	filename := fmt.Sprintf("%s.json", id)
	filePath := filepath.Join(historyPath, filename)

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, fmt.Errorf("read session file: %w", err)
	}

	// Unmarshal session
	var sess session.Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return &sess, nil
}

// DeleteSession deletes a session from storage with confirmation
func (s *SessionStorage) DeleteSession(id string) error {
	// Get history directory path
	historyPath, err := s.testFolderManager.GetSubdirectoryPath("history")
	if err != nil {
		return fmt.Errorf("get history path: %w", err)
	}

	// Create session file path
	filename := fmt.Sprintf("%s.json", id)
	filePath := filepath.Join(historyPath, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("session not found: %s", id)
	}

	// Delete file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("delete session file: %w", err)
	}

	return nil
}

// ListSessions returns metadata for all sessions in storage
func (s *SessionStorage) ListSessions() ([]*session.SessionMetadata, error) {
	// Get history directory path
	historyPath, err := s.testFolderManager.GetSubdirectoryPath("history")
	if err != nil {
		return nil, fmt.Errorf("get history path: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(historyPath, 0755); err != nil {
		return nil, fmt.Errorf("create history directory: %w", err)
	}

	// Read directory
	entries, err := os.ReadDir(historyPath)
	if err != nil {
		return nil, fmt.Errorf("read history directory: %w", err)
	}

	var sessions []*session.SessionMetadata
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process JSON files
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Extract session ID from filename
		id := strings.TrimSuffix(entry.Name(), ".json")

		// Load session to get metadata
		sess, err := s.LoadSession(id)
		if err != nil {
			// Skip corrupted sessions
			continue
		}

		metadata := &session.SessionMetadata{
			ID:           sess.ID,
			Name:         sess.Name,
			Model:        sess.Model,
			Tags:         sess.Tags,
			MessageCount: len(sess.Messages),
			CreatedAt:    sess.CreatedAt,
			UpdatedAt:    sess.UpdatedAt,
		}

		sessions = append(sessions, metadata)
	}

	return sessions, nil
}
