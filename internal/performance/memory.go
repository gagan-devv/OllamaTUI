package performance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/gagan-devv/ollama-go/internal/session"
)

const (
	// MaxInMemoryMessages is the maximum number of messages to keep in memory
	MaxInMemoryMessages = 1000
	// MemoryWarningThresholdMB is the threshold for memory warnings in megabytes
	MemoryWarningThresholdMB = 500
)

// MemoryManager handles memory optimization for sessions
type MemoryManager struct {
	archivePath      string
	memoryLimitMB    int64
	maxInMemory      int
	mu               sync.RWMutex
	archivedSessions map[string]*ArchivedSession
}

// ArchivedSession tracks which messages are archived for a session
type ArchivedSession struct {
	SessionID        string    `json:"session_id"`
	TotalMessages    int       `json:"total_messages"`
	InMemoryMessages int       `json:"in_memory_messages"`
	ArchivedMessages int       `json:"archived_messages"`
	ArchiveFile      string    `json:"archive_file"`
	LastArchived     time.Time `json:"last_archived"`
}

// ArchivedMessages represents messages stored on disk
type ArchivedMessages struct {
	SessionID string            `json:"session_id"`
	Messages  []session.Message `json:"messages"`
	Archived  time.Time         `json:"archived"`
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager(archivePath string, memoryLimitMB int64) *MemoryManager {
	if memoryLimitMB <= 0 {
		memoryLimitMB = MemoryWarningThresholdMB
	}

	return &MemoryManager{
		archivePath:      archivePath,
		memoryLimitMB:    memoryLimitMB,
		maxInMemory:      MaxInMemoryMessages,
		archivedSessions: make(map[string]*ArchivedSession),
	}
}

// OptimizeSession optimizes memory usage for a session by archiving old messages
func (m *MemoryManager) OptimizeSession(sess *session.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(sess.Messages) <= m.maxInMemory {
		return nil // No optimization needed
	}

	// Calculate how many messages to archive
	toArchive := len(sess.Messages) - m.maxInMemory
	if toArchive <= 0 {
		return nil
	}

	// Extract messages to archive (oldest messages)
	messagesToArchive := sess.Messages[:toArchive]
	sess.Messages = sess.Messages[toArchive:]

	// Save archived messages to disk
	archiveFile := filepath.Join(m.archivePath, fmt.Sprintf("%s_archive.json", sess.ID))
	if err := m.saveArchivedMessages(sess.ID, messagesToArchive, archiveFile); err != nil {
		return fmt.Errorf("save archived messages: %w", err)
	}

	// Update tracking
	m.archivedSessions[sess.ID] = &ArchivedSession{
		SessionID:        sess.ID,
		TotalMessages:    len(messagesToArchive) + len(sess.Messages),
		InMemoryMessages: len(sess.Messages),
		ArchivedMessages: len(messagesToArchive),
		ArchiveFile:      archiveFile,
		LastArchived:     time.Now(),
	}

	return nil
}

// LoadArchivedMessages loads archived messages for a session
func (m *MemoryManager) LoadArchivedMessages(sessionID string) ([]session.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	archived, ok := m.archivedSessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("no archived messages for session %s", sessionID)
	}

	data, err := os.ReadFile(archived.ArchiveFile)
	if err != nil {
		return nil, fmt.Errorf("read archive file: %w", err)
	}

	var archivedMessages ArchivedMessages
	if err := json.Unmarshal(data, &archivedMessages); err != nil {
		return nil, fmt.Errorf("unmarshal archived messages: %w", err)
	}

	return archivedMessages.Messages, nil
}

// RestoreArchivedMessages restores archived messages back into a session
func (m *MemoryManager) RestoreArchivedMessages(sess *session.Session) error {
	messages, err := m.LoadArchivedMessages(sess.ID)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Prepend archived messages to current messages
	sess.Messages = append(messages, sess.Messages...)

	// Clear archive tracking
	delete(m.archivedSessions, sess.ID)

	return nil
}

// GetMemoryUsage returns current memory usage in megabytes
func (m *MemoryManager) GetMemoryUsage() int64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return int64(memStats.Alloc / 1024 / 1024)
}

// CheckMemoryThreshold checks if memory usage exceeds the threshold
func (m *MemoryManager) CheckMemoryThreshold() (bool, int64) {
	usage := m.GetMemoryUsage()
	return usage > m.memoryLimitMB, usage
}

// GetArchiveInfo returns archive information for a session
func (m *MemoryManager) GetArchiveInfo(sessionID string) (*ArchivedSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	archived, ok := m.archivedSessions[sessionID]
	return archived, ok
}

// ClearSession clears all messages from a session and removes archives
func (m *MemoryManager) ClearSession(sess *session.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear in-memory messages
	sess.Messages = []session.Message{}

	// Remove archive if it exists
	if archived, ok := m.archivedSessions[sess.ID]; ok {
		if err := os.Remove(archived.ArchiveFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove archive file: %w", err)
		}
		delete(m.archivedSessions, sess.ID)
	}

	return nil
}

// saveArchivedMessages saves messages to an archive file
func (m *MemoryManager) saveArchivedMessages(sessionID string, messages []session.Message, archiveFile string) error {
	// Ensure archive directory exists
	if err := os.MkdirAll(filepath.Dir(archiveFile), 0755); err != nil {
		return fmt.Errorf("create archive directory: %w", err)
	}

	// Load existing archived messages if file exists
	var existingMessages []session.Message
	if data, err := os.ReadFile(archiveFile); err == nil {
		var existing ArchivedMessages
		if err := json.Unmarshal(data, &existing); err == nil {
			existingMessages = existing.Messages
		}
	}

	// Append new messages to existing
	allMessages := append(existingMessages, messages...)

	archived := ArchivedMessages{
		SessionID: sessionID,
		Messages:  allMessages,
		Archived:  time.Now(),
	}

	data, err := json.MarshalIndent(archived, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal archived messages: %w", err)
	}

	if err := os.WriteFile(archiveFile, data, 0644); err != nil {
		return fmt.Errorf("write archive file: %w", err)
	}

	return nil
}

// GetSuggestedAction returns a suggested action based on memory usage
func (m *MemoryManager) GetSuggestedAction() string {
	exceeded, usage := m.CheckMemoryThreshold()
	if !exceeded {
		return ""
	}

	return fmt.Sprintf("Memory usage is high (%d MB). Consider archiving old sessions or clearing conversation history.", usage)
}
