package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestTestFolderCreatedOnFirstRun tests Property 64: Test Folder Created On First Run
// **Validates: Requirements 48.1**
func TestTestFolderCreatedOnFirstRun(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("test folder is created on first run", prop.ForAll(
		func(folderName string) bool {
			// Create a temporary directory for testing
			tempDir := t.TempDir()
			testPath := filepath.Join(tempDir, folderName)

			// Ensure the folder doesn't exist initially
			if _, err := os.Stat(testPath); !os.IsNotExist(err) {
				return false
			}

			// Create TestFolderManager and ensure folder
			tfm := NewTestFolderManager(testPath)
			absPath, err := tfm.EnsureTestFolder()
			if err != nil {
				t.Logf("EnsureTestFolder failed: %v", err)
				return false
			}

			// Verify the folder was created
			info, err := os.Stat(absPath)
			if err != nil {
				t.Logf("Folder not created: %v", err)
				return false
			}

			// Verify it's a directory
			if !info.IsDir() {
				t.Logf("Path exists but is not a directory")
				return false
			}

			return true
		},
		gen.Identifier().SuchThat(func(s string) bool {
			// Filter out empty strings and strings with path separators
			return s != "" && !filepath.IsAbs(s) && filepath.Clean(s) == s
		}),
	))

	properties.TestingRun(t)
}

// TestAllArtifactsInTestFolder tests Property 65: All Artifacts In Test Folder
// **Validates: Requirements 48.3, 48.4, 48.5, 48.6, 48.7**
func TestAllArtifactsInTestFolder(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("all artifact subdirectories are created in test folder", prop.ForAll(
		func(folderName string) bool {
			// Create a temporary directory for testing
			tempDir := t.TempDir()
			testPath := filepath.Join(tempDir, folderName)

			// Create TestFolderManager and ensure folder
			tfm := NewTestFolderManager(testPath)
			absPath, err := tfm.EnsureTestFolder()
			if err != nil {
				t.Logf("EnsureTestFolder failed: %v", err)
				return false
			}

			// Required subdirectories for different artifact types
			// 48.4: conversation history files
			// 48.5: exported files
			// 48.6: cache files
			// 48.7: debug logs
			requiredSubdirs := []string{"history", "cache", "exports", "drafts", "logs"}

			// Verify all subdirectories exist
			for _, subdir := range requiredSubdirs {
				subdirPath := filepath.Join(absPath, subdir)
				info, err := os.Stat(subdirPath)
				if err != nil {
					t.Logf("Subdirectory %s not created: %v", subdir, err)
					return false
				}
				if !info.IsDir() {
					t.Logf("Path %s exists but is not a directory", subdir)
					return false
				}
			}

			return true
		},
		gen.Identifier().SuchThat(func(s string) bool {
			return s != "" && !filepath.IsAbs(s) && filepath.Clean(s) == s
		}),
	))

	properties.TestingRun(t)
}

// TestDefaultTestFolderName verifies that the default folder name is "test_output"
// **Validates: Requirements 48.2**
func TestDefaultTestFolderName(t *testing.T) {
	tfm := NewTestFolderManager("")
	if tfm.basePath != DefaultTestFolder {
		t.Errorf("Expected default folder name %s, got %s", DefaultTestFolder, tfm.basePath)
	}
	
	if DefaultTestFolder != "test_output" {
		t.Errorf("Expected DefaultTestFolder to be 'test_output', got %s", DefaultTestFolder)
	}
}

// TestPathValidation tests that path validation prevents directory traversal
func TestPathValidation(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test_output")
	
	tfm := NewTestFolderManager(testPath)
	_, err := tfm.EnsureTestFolder()
	if err != nil {
		t.Fatalf("Failed to create test folder: %v", err)
	}

	// Test valid paths
	validPaths := []string{
		filepath.Join(testPath, "history", "session.json"),
		filepath.Join(testPath, "cache", "response.cache"),
		filepath.Join(testPath, "exports", "conversation.md"),
	}

	for _, path := range validPaths {
		if err := tfm.ValidatePath(path); err != nil {
			t.Errorf("Valid path %s failed validation: %v", path, err)
		}
	}

	// Test invalid paths (trying to escape test folder)
	invalidPaths := []string{
		filepath.Join(testPath, "..", "outside.txt"),
		filepath.Join(testPath, "history", "..", "..", "escape.txt"),
	}

	for _, path := range invalidPaths {
		if err := tfm.ValidatePath(path); err == nil {
			t.Errorf("Invalid path %s should have failed validation", path)
		}
	}
}

// TestGetSubdirectoryPath tests retrieving subdirectory paths
func TestGetSubdirectoryPath(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test_output")
	
	tfm := NewTestFolderManager(testPath)
	_, err := tfm.EnsureTestFolder()
	if err != nil {
		t.Fatalf("Failed to create test folder: %v", err)
	}

	subdirs := []string{"history", "cache", "exports", "drafts", "logs"}
	for _, subdir := range subdirs {
		path, err := tfm.GetSubdirectoryPath(subdir)
		if err != nil {
			t.Errorf("Failed to get path for %s: %v", subdir, err)
			continue
		}

		// Verify the path exists
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Subdirectory %s does not exist at %s: %v", subdir, path, err)
		}
	}
}
