package components

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// InputHistory manages a persistent history of user inputs.
type InputHistory struct {
	entries  []string
	maxSize  int
	filePath string
	mu       sync.RWMutex
}

// historyFile represents the JSON structure for persisted history.
type historyFile struct {
	Version string   `json:"version"`
	Entries []string `json:"entries"`
}

// NewInputHistory creates a new input history manager.
// historyPath is the file path where history will be persisted.
// maxSize is the maximum number of entries to keep (e.g., 100).
func NewInputHistory(historyPath string, maxSize int) *InputHistory {
	h := &InputHistory{
		entries:  make([]string, 0, maxSize),
		maxSize:  maxSize,
		filePath: historyPath,
	}

	// Load existing history from disk
	h.load()

	return h
}

// Add adds a new entry to the history.
// The entry is added to the front of the list (most recent first).
// If the history exceeds maxSize, the oldest entry is removed.
func (h *InputHistory) Add(entry string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Don't add empty entries
	if entry == "" {
		return
	}

	// Don't add duplicates of the most recent entry
	if len(h.entries) > 0 && h.entries[0] == entry {
		return
	}

	// Add to front
	h.entries = append([]string{entry}, h.entries...)

	// Trim to maxSize
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[:h.maxSize]
	}

	// Persist to disk
	h.save()
}

// Get retrieves an entry by index.
// Index 0 is the most recent entry.
// Returns empty string if index is out of bounds.
func (h *InputHistory) Get(index int) string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if index < 0 || index >= len(h.entries) {
		return ""
	}

	return h.entries[index]
}

// Len returns the number of entries in the history.
func (h *InputHistory) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.entries)
}

// Clear removes all entries from the history.
func (h *InputHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = make([]string, 0, h.maxSize)
	h.save()
}

// load reads the history from disk.
func (h *InputHistory) load() {
	// Ensure directory exists
	dir := filepath.Dir(h.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return // Silently fail, history will start empty
	}

	// Read file
	data, err := os.ReadFile(h.filePath)
	if err != nil {
		return // File doesn't exist yet, start with empty history
	}

	// Parse JSON
	var hf historyFile
	if err := json.Unmarshal(data, &hf); err != nil {
		return // Corrupted file, start with empty history
	}

	// Load entries
	h.entries = hf.Entries
	if h.entries == nil {
		h.entries = make([]string, 0, h.maxSize)
	}

	// Trim to maxSize if needed
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[:h.maxSize]
	}
}

// save writes the history to disk.
func (h *InputHistory) save() {
	// Ensure directory exists
	dir := filepath.Dir(h.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return // Silently fail
	}

	// Create history file structure
	hf := historyFile{
		Version: "1.0",
		Entries: h.entries,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(hf, "", "  ")
	if err != nil {
		return // Silently fail
	}

	// Write to file
	os.WriteFile(h.filePath, data, 0644)
}
