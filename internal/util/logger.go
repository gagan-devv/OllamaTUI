package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// LogLevelDebug is for detailed debugging information
	LogLevelDebug LogLevel = iota
	// LogLevelInfo is for general informational messages
	LogLevelInfo
	// LogLevelWarn is for warning messages
	LogLevelWarn
	// LogLevelError is for error messages
	LogLevelError
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging with levels
type Logger struct {
	mu          sync.Mutex
	level       LogLevel
	output      io.Writer
	logFile     *os.File
	logFilePath string
}

var (
	// defaultLogger is the global logger instance
	defaultLogger *Logger
	loggerOnce    sync.Once
)

// InitLogger initializes the global logger with the specified log directory
func InitLogger(logDir string, level LogLevel) error {
	var initErr error
	loggerOnce.Do(func() {
		// Ensure log directory exists
		if err := os.MkdirAll(logDir, 0755); err != nil {
			initErr = fmt.Errorf("failed to create log directory: %w", err)
			return
		}

		// Create log file with timestamp
		timestamp := time.Now().Format("2006-01-02")
		logFileName := fmt.Sprintf("ollama-go-%s.log", timestamp)
		logFilePath := filepath.Join(logDir, logFileName)

		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			initErr = fmt.Errorf("failed to open log file: %w", err)
			return
		}

		defaultLogger = &Logger{
			level:       level,
			output:      io.MultiWriter(os.Stderr, logFile),
			logFile:     logFile,
			logFilePath: logFilePath,
		}
	})

	return initErr
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if defaultLogger == nil {
		// Initialize with default settings if not already initialized
		_ = InitLogger("test_output/logs", LogLevelInfo)
	}
	return defaultLogger
}

// Close closes the log file
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// log writes a log message with the specified level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level.String(), message)

	if l.output != nil {
		_, _ = l.output.Write([]byte(logLine))
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LogLevelDebug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LogLevelInfo, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LogLevelWarn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LogLevelError, format, args...)
}

// LogError logs an AppError with full details
func (l *Logger) LogError(err *AppError) {
	if err == nil {
		return
	}

	l.Error("Error occurred: Code=%s, Category=%s, Message=%s", err.Code, err.Category, err.Message)
	if err.Cause != nil {
		l.Error("  Caused by: %v", err.Cause)
	}
	if err.Suggestion != "" {
		l.Debug("  Suggestion: %s", err.Suggestion)
	}
}

// Global logging functions for convenience

// Debug logs a debug message using the global logger
func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

// Info logs an info message using the global logger
func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

// Warn logs a warning message using the global logger
func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

// Error logs an error message using the global logger
func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

// LogError logs an AppError using the global logger
func LogError(err *AppError) {
	GetLogger().LogError(err)
}

// SetLogLevel sets the log level for the global logger
func SetLogLevel(level LogLevel) {
	GetLogger().SetLevel(level)
}

// CloseLogger closes the global logger
func CloseLogger() error {
	if defaultLogger != nil {
		return defaultLogger.Close()
	}
	return nil
}
