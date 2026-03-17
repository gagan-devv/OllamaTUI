package util

import (
	"fmt"
)

// ErrorCategory represents the category of an error
type ErrorCategory string

const (
	// ErrorCategoryNetwork represents network-related errors
	ErrorCategoryNetwork ErrorCategory = "Network"
	// ErrorCategoryAPI represents API-related errors
	ErrorCategoryAPI ErrorCategory = "API"
	// ErrorCategoryFileSystem represents file system errors
	ErrorCategoryFileSystem ErrorCategory = "FileSystem"
	// ErrorCategoryValidation represents validation errors
	ErrorCategoryValidation ErrorCategory = "Validation"
	// ErrorCategoryInternal represents internal application errors
	ErrorCategoryInternal ErrorCategory = "Internal"
	// ErrorCategoryPlugin represents plugin-related errors
	ErrorCategoryPlugin ErrorCategory = "Plugin"
)

// ErrorCode represents a specific error code
type ErrorCode string

const (
	// Network errors (E001-E003)
	ErrCodeConnectionFailed    ErrorCode = "E001"
	ErrCodeConnectionTimeout   ErrorCode = "E002"
	ErrCodeServiceUnavailable  ErrorCode = "E003"

	// API errors (E004-E008)
	ErrCodeModelNotFound       ErrorCode = "E004"
	ErrCodeInvalidModel        ErrorCode = "E005"
	ErrCodeAPIError            ErrorCode = "E006"
	ErrCodeStreamingFailed     ErrorCode = "E007"
	ErrCodeInvalidResponse     ErrorCode = "E008"

	// File system errors (E009-E012)
	ErrCodeFileNotFound        ErrorCode = "E009"
	ErrCodeFileReadError       ErrorCode = "E010"
	ErrCodeFileWriteError      ErrorCode = "E011"
	ErrCodePermissionDenied    ErrorCode = "E012"

	// Validation errors (E013-E016)
	ErrCodeInvalidConfig       ErrorCode = "E013"
	ErrCodeInvalidParameter    ErrorCode = "E014"
	ErrCodeInvalidInput        ErrorCode = "E015"
	ErrCodeValidationFailed    ErrorCode = "E016"

	// Internal errors (E017-E018)
	ErrCodeInternalError       ErrorCode = "E017"
	ErrCodeUnexpectedState     ErrorCode = "E018"

	// Plugin errors (E019)
	ErrCodePluginLoadFailed    ErrorCode = "E019"
)

// AppError represents an application error with context and suggestions
type AppError struct {
	Code       ErrorCode
	Category   ErrorCategory
	Message    string
	Suggestion string
	Cause      error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %s (caused by: %v)", e.Code, e.Category, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Category, e.Message)
}

// Unwrap returns the underlying cause
func (e *AppError) Unwrap() error {
	return e.Cause
}

// UserMessage returns a user-friendly error message with suggestion
func (e *AppError) UserMessage() string {
	msg := fmt.Sprintf("Error %s: %s", e.Code, e.Message)
	if e.Suggestion != "" {
		msg += fmt.Sprintf("\n\nSuggestion: %s", e.Suggestion)
	}
	return msg
}

// errorMessages maps error codes to user-friendly messages
var errorMessages = map[ErrorCode]string{
	// Network errors
	ErrCodeConnectionFailed:    "Failed to connect to Ollama service",
	ErrCodeConnectionTimeout:   "Connection to Ollama service timed out",
	ErrCodeServiceUnavailable:  "Ollama service is unavailable",

	// API errors
	ErrCodeModelNotFound:       "The requested model was not found",
	ErrCodeInvalidModel:        "The specified model is invalid",
	ErrCodeAPIError:            "An API error occurred",
	ErrCodeStreamingFailed:     "Streaming response failed",
	ErrCodeInvalidResponse:     "Received invalid response from API",

	// File system errors
	ErrCodeFileNotFound:        "File not found",
	ErrCodeFileReadError:       "Failed to read file",
	ErrCodeFileWriteError:      "Failed to write file",
	ErrCodePermissionDenied:    "Permission denied",

	// Validation errors
	ErrCodeInvalidConfig:       "Invalid configuration",
	ErrCodeInvalidParameter:    "Invalid parameter value",
	ErrCodeInvalidInput:        "Invalid input",
	ErrCodeValidationFailed:    "Validation failed",

	// Internal errors
	ErrCodeInternalError:       "An internal error occurred",
	ErrCodeUnexpectedState:     "Application reached an unexpected state",

	// Plugin errors
	ErrCodePluginLoadFailed:    "Failed to load plugin",
}

// NewAppError creates a new AppError with the given code and cause
func NewAppError(code ErrorCode, category ErrorCategory, cause error) *AppError {
	message, ok := errorMessages[code]
	if !ok {
		message = "An unknown error occurred"
	}

	return &AppError{
		Code:     code,
		Category: category,
		Message:  message,
		Cause:    cause,
	}
}

// NewAppErrorWithMessage creates a new AppError with a custom message
func NewAppErrorWithMessage(code ErrorCode, category ErrorCategory, message string, cause error) *AppError {
	return &AppError{
		Code:     code,
		Category: category,
		Message:  message,
		Cause:    cause,
	}
}

// WithSuggestion adds a suggestion to the error
func (e *AppError) WithSuggestion(suggestion string) *AppError {
	e.Suggestion = suggestion
	return e
}

// HandleNetworkError creates an AppError for network-related issues with troubleshooting suggestions
func HandleNetworkError(err error) *AppError {
	if err == nil {
		return nil
	}

	// Determine specific error code based on error type
	code := ErrCodeConnectionFailed
	suggestion := "Please check:\n" +
		"  1. Ollama service is running (run 'ollama serve')\n" +
		"  2. Network connection is stable\n" +
		"  3. Firewall settings allow connection\n" +
		"  4. Ollama is listening on the correct port (default: 11434)"

	errStr := err.Error()
	if contains(errStr, "timeout") {
		code = ErrCodeConnectionTimeout
		suggestion = "The connection timed out. Please check:\n" +
			"  1. Ollama service is responding\n" +
			"  2. Network latency is acceptable\n" +
			"  3. Try increasing timeout in configuration"
	} else if contains(errStr, "refused") || contains(errStr, "unavailable") {
		code = ErrCodeServiceUnavailable
		suggestion = "Ollama service is not running. Please:\n" +
			"  1. Start Ollama service: 'ollama serve'\n" +
			"  2. Verify service status\n" +
			"  3. Check service logs for errors"
	}

	return NewAppError(code, ErrorCategoryNetwork, err).WithSuggestion(suggestion)
}

// HandleAPIError creates an AppError for API-related issues with model suggestions
func HandleAPIError(err error, modelName string, availableModels []string) *AppError {
	if err == nil {
		return nil
	}

	code := ErrCodeAPIError
	suggestion := "Please check the API documentation or try again later."

	errStr := err.Error()
	if contains(errStr, "not found") || contains(errStr, "404") {
		code = ErrCodeModelNotFound
		suggestion = fmt.Sprintf("Model '%s' not found.", modelName)
		
		if len(availableModels) > 0 {
			suggestion += "\n\nAvailable models:\n"
			for _, model := range availableModels {
				suggestion += fmt.Sprintf("  - %s\n", model)
			}
			
			// Find closest match
			closest := findClosestMatch(modelName, availableModels)
			if closest != "" {
				suggestion += fmt.Sprintf("\nDid you mean '%s'?", closest)
			}
		} else {
			suggestion += "\n\nNo models are currently available. Download a model using:\n  ollama pull <model-name>"
		}
	} else if contains(errStr, "invalid") || contains(errStr, "400") {
		code = ErrCodeInvalidModel
		suggestion = fmt.Sprintf("Model '%s' is invalid. Please check the model name and try again.", modelName)
	} else if contains(errStr, "stream") {
		code = ErrCodeStreamingFailed
		suggestion = "Streaming failed. The partial response has been preserved.\n" +
			"Please check your connection and try again."
	}

	return NewAppError(code, ErrorCategoryAPI, err).WithSuggestion(suggestion)
}

// HandleFileSystemError creates an AppError for file system issues with path and permission details
func HandleFileSystemError(err error, filePath string, operation string) *AppError {
	if err == nil {
		return nil
	}

	code := ErrCodeFileReadError
	if operation == "write" || operation == "create" {
		code = ErrCodeFileWriteError
	}

	suggestion := fmt.Sprintf("File path: %s\n", filePath)
	errStr := err.Error()

	if contains(errStr, "not found") || contains(errStr, "no such file") {
		code = ErrCodeFileNotFound
		suggestion += "\nThe file does not exist. Please check:\n" +
			"  1. File path is correct\n" +
			"  2. File has not been moved or deleted\n" +
			"  3. You have permission to access the directory"
	} else if contains(errStr, "permission") || contains(errStr, "denied") {
		code = ErrCodePermissionDenied
		suggestion += "\nPermission denied. Please check:\n" +
			"  1. You have read/write permissions for this file\n" +
			"  2. The file is not locked by another process\n" +
			"  3. Directory permissions allow access"
	} else {
		suggestion += fmt.Sprintf("\nFailed to %s file. Please check:\n", operation) +
			"  1. File path is valid\n" +
			"  2. Disk space is available\n" +
			"  3. File system is not read-only"
	}

	return NewAppError(code, ErrorCategoryFileSystem, err).WithSuggestion(suggestion)
}

// HandleValidationError creates an AppError for validation issues with specific field errors
func HandleValidationError(field string, value interface{}, expectedFormat string) *AppError {
	code := ErrCodeValidationFailed
	message := fmt.Sprintf("Validation failed for field '%s'", field)
	
	suggestion := fmt.Sprintf("Field: %s\nProvided value: %v\nExpected format: %s\n", field, value, expectedFormat)
	
	// Provide specific guidance based on field type
	if contains(field, "temperature") {
		suggestion += "\nTemperature must be between 0.0 and 2.0.\n" +
			"  - Lower values (0.0-0.7): More focused and deterministic\n" +
			"  - Higher values (0.8-2.0): More creative and random"
	} else if contains(field, "top_p") {
		suggestion += "\nTop P must be between 0.0 and 1.0.\n" +
			"  - Controls diversity via nucleus sampling"
	} else if contains(field, "top_k") {
		suggestion += "\nTop K must be a positive integer.\n" +
			"  - Limits vocabulary to top K tokens"
	} else if contains(field, "color") {
		suggestion += "\nColor must be a valid hex code (e.g., #FF5733) or named color (e.g., red)."
	} else if contains(field, "path") {
		suggestion += "\nPath must be a valid file system path."
	}

	return NewAppErrorWithMessage(code, ErrorCategoryValidation, message, nil).WithSuggestion(suggestion)
}

// HandleConfigError creates an AppError for configuration issues
func HandleConfigError(err error, configKey string) *AppError {
	if err == nil {
		return nil
	}

	message := fmt.Sprintf("Invalid configuration for key '%s'", configKey)
	suggestion := fmt.Sprintf("Configuration key: %s\n", configKey) +
		"Please check your configuration file (~/.ollama-go/config.yaml) and ensure:\n" +
		"  1. The value is in the correct format\n" +
		"  2. Required fields are present\n" +
		"  3. YAML syntax is valid"

	return NewAppErrorWithMessage(ErrCodeInvalidConfig, ErrorCategoryValidation, message, err).WithSuggestion(suggestion)
}

// HandleInternalError creates an AppError for internal application errors
func HandleInternalError(err error, context string) *AppError {
	if err == nil {
		return nil
	}

	message := "An internal error occurred"
	if context != "" {
		message += fmt.Sprintf(" in %s", context)
	}

	suggestion := "This is an unexpected error. Please:\n" +
		"  1. Check the debug logs for more details\n" +
		"  2. Try restarting the application\n" +
		"  3. Report this issue if it persists"

	return NewAppErrorWithMessage(ErrCodeInternalError, ErrorCategoryInternal, message, err).WithSuggestion(suggestion)
}

// HandlePluginError creates an AppError for plugin-related issues
func HandlePluginError(err error, pluginName string) *AppError {
	if err == nil {
		return nil
	}

	message := fmt.Sprintf("Failed to load plugin '%s'", pluginName)
	suggestion := fmt.Sprintf("Plugin: %s\n", pluginName) +
		"Please check:\n" +
		"  1. Plugin file exists in ~/.ollama-go/plugins/\n" +
		"  2. Plugin is compatible with this version\n" +
		"  3. Plugin has correct permissions\n" +
		"  4. Plugin dependencies are satisfied"

	return NewAppErrorWithMessage(ErrCodePluginLoadFailed, ErrorCategoryPlugin, message, err).WithSuggestion(suggestion)
}

// Helper functions

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(s) > len(substr) && containsIgnoreCase(s, substr))
}

func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// findClosestMatch finds the closest matching string using simple edit distance
func findClosestMatch(target string, candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}

	minDist := len(target) + 1
	closest := ""

	for _, candidate := range candidates {
		dist := levenshteinDistance(toLower(target), toLower(candidate))
		if dist < minDist {
			minDist = dist
			closest = candidate
		}
	}

	// Only return if reasonably close (within 3 edits)
	if minDist <= 3 {
		return closest
	}
	return ""
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
