package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DefaultTestFolder is the default name for the test artifacts directory
	DefaultTestFolder = "test_output"
)

// TestFolderManager handles creation and management of the test folder structure
type TestFolderManager struct {
	basePath string
}

// NewTestFolderManager creates a new test folder manager with the specified base path
func NewTestFolderManager(basePath string) *TestFolderManager {
	if basePath == "" {
		basePath = DefaultTestFolder
	}
	return &TestFolderManager{
		basePath: basePath,
	}
}

// EnsureTestFolder creates the test folder and all subdirectories if they don't exist
// Returns the absolute path to the test folder
func (m *TestFolderManager) EnsureTestFolder() (string, error) {
	// Get absolute path
	absPath, err := filepath.Abs(m.basePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Create main test folder
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create test folder: %w", err)
	}

	// Create subdirectories
	subdirs := []string{"history", "cache", "exports", "drafts", "logs"}
	for _, subdir := range subdirs {
		subdirPath := filepath.Join(absPath, subdir)
		if err := os.MkdirAll(subdirPath, 0755); err != nil {
			return "", fmt.Errorf("failed to create subdirectory %s: %w", subdir, err)
		}
	}

	return absPath, nil
}

// GetTestFolderPath returns the absolute path to the test folder
func (m *TestFolderManager) GetTestFolderPath() (string, error) {
	return filepath.Abs(m.basePath)
}

// GetSubdirectoryPath returns the absolute path to a specific subdirectory
func (m *TestFolderManager) GetSubdirectoryPath(subdir string) (string, error) {
	absPath, err := filepath.Abs(m.basePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}
	return filepath.Join(absPath, subdir), nil
}

// ValidatePath checks if a path is within the test folder
func (m *TestFolderManager) ValidatePath(path string) error {
	absTestPath, err := filepath.Abs(m.basePath)
	if err != nil {
		return fmt.Errorf("failed to resolve test folder path: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path is within test folder
	relPath, err := filepath.Rel(absTestPath, absPath)
	if err != nil {
		return fmt.Errorf("path validation failed: %w", err)
	}

	// Check if path tries to escape test folder using ../
	if len(relPath) > 0 && relPath[0] == '.' && len(relPath) > 1 && relPath[1] == '.' {
		return fmt.Errorf("path %s is outside test folder", path)
	}

	return nil
}


// EnsureGitignore adds the test folder to .gitignore if not already present
func (m *TestFolderManager) EnsureGitignore() error {
	gitignorePath := ".gitignore"
	
	// Read existing .gitignore if it exists
	content, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read .gitignore: %w", err)
	}

	// Check if test folder is already in .gitignore
	gitignoreContent := string(content)
	testFolderEntry := m.basePath + "/"
	
	// Check for exact match or with trailing slash
	if contains(gitignoreContent, m.basePath+"/") || contains(gitignoreContent, m.basePath+"\n") {
		return nil // Already present
	}

	// Append test folder to .gitignore
	var newContent string
	if len(content) > 0 && content[len(content)-1] != '\n' {
		newContent = gitignoreContent + "\n" + testFolderEntry + "\n"
	} else {
		newContent = gitignoreContent + testFolderEntry + "\n"
	}

	if err := os.WriteFile(gitignorePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}

	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
