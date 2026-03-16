package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// SessionManager manages multiple chat sessions
type SessionManager interface {
	// Session lifecycle
	CreateSession(name string, config SessionConfig) (*Session, error)
	GetSession(id string) (*Session, error)
	ListSessions() ([]*SessionMetadata, error)
	DeleteSession(id string) error

	// Session operations
	SwitchSession(id string) error
	RenameSession(id string, newName string) error
	SaveSession(session *Session) error
	LoadSession(id string) (*Session, error)

	// Tagging and organization
	AddTag(sessionID string, tag string) error
	RemoveTag(sessionID string, tag string) error
	FindByTag(tag string) ([]*SessionMetadata, error)

	// Bookmarking
	AddBookmark(sessionID string, messageIndex int) error
	RemoveBookmark(sessionID string, messageIndex int) error
	GetBookmarks(sessionID string) ([]int, error)

	// Current session
	GetCurrentSession() (*Session, error)
	SetCurrentSession(id string) error
}

// Storage defines the interface for session persistence
type Storage interface {
	SaveSession(session *Session) error
	LoadSession(id string) (*Session, error)
	DeleteSession(id string) error
	ListSessions() ([]*SessionMetadata, error)
}

// DefaultSessionManager is the default implementation of SessionManager
type DefaultSessionManager struct {
	storage        Storage
	sessions       map[string]*Session
	currentSession *Session
	mu             sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager(storage Storage) SessionManager {
	return &DefaultSessionManager{
		storage:  storage,
		sessions: make(map[string]*Session),
	}
}

// generateSessionID generates a unique session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate session ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSession creates a new session with the given name and configuration
func (m *DefaultSessionManager) CreateSession(name string, config SessionConfig) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &Session{
		ID:           id,
		Name:         name,
		Messages:     []Message{},
		Model:        config.Model,
		SystemPrompt: config.SystemPrompt,
		Parameters:   config.Parameters,
		Tags:         []string{},
		Bookmarks:    []int{},
		CreatedAt:    now,
		UpdatedAt:    now,
		Metadata:     make(map[string]interface{}),
	}

	m.sessions[id] = session

	// Save to storage
	if err := m.storage.SaveSession(session); err != nil {
		delete(m.sessions, id)
		return nil, fmt.Errorf("save session: %w", err)
	}

	return session, nil
}

// GetSession retrieves a session by ID
func (m *DefaultSessionManager) GetSession(id string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if session, ok := m.sessions[id]; ok {
		return session, nil
	}

	// Try loading from storage
	session, err := m.storage.LoadSession(id)
	if err != nil {
		return nil, fmt.Errorf("load session: %w", err)
	}

	m.mu.RUnlock()
	m.mu.Lock()
	m.sessions[id] = session
	m.mu.Unlock()
	m.mu.RLock()

	return session, nil
}

// ListSessions returns metadata for all sessions
func (m *DefaultSessionManager) ListSessions() ([]*SessionMetadata, error) {
	return m.storage.ListSessions()
}

// DeleteSession deletes a session by ID
func (m *DefaultSessionManager) DeleteSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove from memory
	delete(m.sessions, id)

	// Remove from storage
	if err := m.storage.DeleteSession(id); err != nil {
		return fmt.Errorf("delete session from storage: %w", err)
	}

	// Clear current session if it was deleted
	if m.currentSession != nil && m.currentSession.ID == id {
		m.currentSession = nil
	}

	return nil
}

// SwitchSession switches to a different session
func (m *DefaultSessionManager) SwitchSession(id string) error {
	session, err := m.GetSession(id)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Save current session if it exists
	if m.currentSession != nil {
		m.currentSession.UpdatedAt = time.Now()
		if err := m.storage.SaveSession(m.currentSession); err != nil {
			return fmt.Errorf("save current session: %w", err)
		}
	}

	m.currentSession = session
	return nil
}

// RenameSession renames a session
func (m *DefaultSessionManager) RenameSession(id string, newName string) error {
	session, err := m.GetSession(id)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	session.Name = newName
	session.UpdatedAt = time.Now()

	if err := m.storage.SaveSession(session); err != nil {
		return fmt.Errorf("save session: %w", err)
	}

	return nil
}

// SaveSession saves a session to storage
func (m *DefaultSessionManager) SaveSession(session *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session.UpdatedAt = time.Now()
	if err := m.storage.SaveSession(session); err != nil {
		return fmt.Errorf("save session: %w", err)
	}

	return nil
}

// LoadSession loads a session from storage
func (m *DefaultSessionManager) LoadSession(id string) (*Session, error) {
	return m.GetSession(id)
}

// AddTag adds a tag to a session
func (m *DefaultSessionManager) AddTag(sessionID string, tag string) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if tag already exists
	for _, t := range session.Tags {
		if t == tag {
			return nil // Tag already exists
		}
	}

	session.Tags = append(session.Tags, tag)
	session.UpdatedAt = time.Now()

	if err := m.storage.SaveSession(session); err != nil {
		return fmt.Errorf("save session: %w", err)
	}

	return nil
}

// RemoveTag removes a tag from a session
func (m *DefaultSessionManager) RemoveTag(sessionID string, tag string) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Find and remove the tag
	for i, t := range session.Tags {
		if t == tag {
			session.Tags = append(session.Tags[:i], session.Tags[i+1:]...)
			break
		}
	}

	session.UpdatedAt = time.Now()

	if err := m.storage.SaveSession(session); err != nil {
		return fmt.Errorf("save session: %w", err)
	}

	return nil
}

// FindByTag finds all sessions with a specific tag
func (m *DefaultSessionManager) FindByTag(tag string) ([]*SessionMetadata, error) {
	allSessions, err := m.ListSessions()
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	var result []*SessionMetadata
	for _, metadata := range allSessions {
		for _, t := range metadata.Tags {
			if t == tag {
				result = append(result, metadata)
				break
			}
		}
	}

	return result, nil
}

// AddBookmark adds a bookmark to a message in a session
func (m *DefaultSessionManager) AddBookmark(sessionID string, messageIndex int) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate message index
	if messageIndex < 0 || messageIndex >= len(session.Messages) {
		return fmt.Errorf("invalid message index: %d", messageIndex)
	}

	// Check if bookmark already exists
	for _, b := range session.Bookmarks {
		if b == messageIndex {
			return nil // Bookmark already exists
		}
	}

	session.Bookmarks = append(session.Bookmarks, messageIndex)
	session.UpdatedAt = time.Now()

	if err := m.storage.SaveSession(session); err != nil {
		return fmt.Errorf("save session: %w", err)
	}

	return nil
}

// RemoveBookmark removes a bookmark from a session
func (m *DefaultSessionManager) RemoveBookmark(sessionID string, messageIndex int) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Find and remove the bookmark
	for i, b := range session.Bookmarks {
		if b == messageIndex {
			session.Bookmarks = append(session.Bookmarks[:i], session.Bookmarks[i+1:]...)
			break
		}
	}

	session.UpdatedAt = time.Now()

	if err := m.storage.SaveSession(session); err != nil {
		return fmt.Errorf("save session: %w", err)
	}

	return nil
}

// GetBookmarks returns all bookmarks for a session
func (m *DefaultSessionManager) GetBookmarks(sessionID string) ([]int, error) {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return session.Bookmarks, nil
}

// GetCurrentSession returns the current active session
func (m *DefaultSessionManager) GetCurrentSession() (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentSession == nil {
		return nil, fmt.Errorf("no active session")
	}

	return m.currentSession, nil
}

// SetCurrentSession sets the current active session
func (m *DefaultSessionManager) SetCurrentSession(id string) error {
	return m.SwitchSession(id)
}
