package logger

import (
	"strings"
	"testing"
	"time"

	"github.com/indaco/md2audio/internal/testhelpers"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(logger *DefaultLogger, message string, args ...any) *LogEntry
		message  string
		args     []any
		expected string
	}{
		{
			name: "Default message",
			logFunc: func(logger *DefaultLogger, msg string, args ...any) *LogEntry {
				return logger.Default(msg, args...)
			},
			message:  "Operation completed successfully",
			expected: "Operation completed successfully\n",
		},
		{
			name: "Info message with args",
			logFunc: func(logger *DefaultLogger, msg string, args ...any) *LogEntry {
				return logger.Info(msg, args...)
			},
			message:  "Application started",
			args:     []any{"port:", 8080, "version:", "1.0.0"},
			expected: "ℹ Application started port: 8080 version: 1.0.0\n",
		},
		{
			name: "Success message",
			logFunc: func(logger *DefaultLogger, msg string, args ...any) *LogEntry {
				return logger.Success(msg, args...)
			},
			message:  "Operation completed",
			expected: "✔ Operation completed\n",
		},
		{
			name: "Warning message",
			logFunc: func(logger *DefaultLogger, msg string, args ...any) *LogEntry {
				return logger.Warning(msg, args...)
			},
			message:  "Low disk space",
			expected: "⚠ Low disk space\n",
		},
		{
			name: "Error message",
			logFunc: func(logger *DefaultLogger, msg string, args ...any) *LogEntry {
				return logger.Error(msg, args...)
			},
			message:  "Failed to connect to database",
			expected: "✘ Failed to connect to database\n",
		},
		{
			name: "Faint message (no icon)",
			logFunc: func(logger *DefaultLogger, msg string, args ...any) *LogEntry {
				return logger.Faint(msg, args...)
			},
			message:  "This is faint text without an icon",
			expected: "This is faint text without an icon\n",
		},
		{
			name: "Indented Success message",
			logFunc: func(logger *DefaultLogger, msg string, args ...any) *LogEntry {
				logger.WithIndent(true)
				entry := logger.Success(msg, args...)
				logger.WithIndent(false) // Reset after the test.
				return entry
			},
			message:  "Indented operation completed",
			expected: "  ✔ Indented operation completed\n",
		},
		{
			name: "Message with attributes",
			logFunc: func(logger *DefaultLogger, msg string, args ...any) *LogEntry {
				return logger.Success(msg).WithAttrs("items", 42, "duration", "1s")
			},
			message:  "Operation completed",
			expected: "✔ Operation completed\n  - items: 42\n  - duration: 1s\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewDefaultLogger() // Create a new logger instance per test.

			output, err := testhelpers.CaptureStdout(func() {
				tt.logFunc(logger, tt.message, tt.args...)
			})
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			if output != tt.expected {
				t.Errorf("Unexpected output:\nGot: %q\nWant: %q", output, tt.expected)
			}
		})
	}
}

func TestLoggerToJSON(t *testing.T) {
	logger := NewDefaultLogger()
	logger.WithTimestamp(true)

	tests := []struct {
		name        string
		logFunc     func(*DefaultLogger, string, ...any) *LogEntry
		message     string
		args        []any
		attributes  []any
		expectedKey string
	}{
		{
			name:        "Success with attributes",
			logFunc:     (*DefaultLogger).Success,
			message:     "Operation succeeded",
			args:        nil,
			attributes:  []any{"user", "JohnDoe", "action", "login"},
			expectedKey: "attributes",
		},
		{
			name:        "Error with attributes",
			logFunc:     (*DefaultLogger).Error,
			message:     "Database error",
			args:        nil,
			attributes:  []any{"code", 500, "reason", "timeout"},
			expectedKey: "attributes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := tt.logFunc(logger, tt.message, tt.args...)
			entry.WithAttrs(tt.attributes...)

			jsonData, err := entry.ToJSON()
			if err != nil {
				t.Fatalf("unexpected error serializing to JSON: %v", err)
			}

			if !strings.Contains(jsonData, tt.expectedKey) {
				t.Errorf("expected JSON to contain key %q, got: %s", tt.expectedKey, jsonData)
			}
		})
	}
}

func TestLoggerWithTimestamp(t *testing.T) {
	logger := NewDefaultLogger()
	logger.WithTimestamp(true)

	output, err := testhelpers.CaptureStdout(func() {
		logger.Info("Timestamped log")
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	if !strings.Contains(output, "Timestamped log") {
		t.Errorf("expected output to contain the log message, got: %q", output)
	}

	// Check for a partial timestamp (e.g., the current date in "2006-01-02" format).
	if !strings.Contains(output, time.Now().Format("2006-01-02")) {
		t.Errorf("expected output to contain the current date, got: %q", output)
	}
}

func TestLogAttrsWithIndent(t *testing.T) {
	tests := []struct {
		name           string
		indentEnabled  bool
		attrs          []any
		expectedOutput string
	}{
		{
			name:          "Attributes without indentation",
			indentEnabled: false,
			attrs:         []any{"key1", "value1", "key2", "value2"},
			expectedOutput: `✔ Testing attributes
  - key1: value1
  - key2: value2
`,
		},
		{
			name:          "Attributes with indentation",
			indentEnabled: true,
			attrs:         []any{"key1", "value1", "key2", "value2"},
			expectedOutput: `  ✔ Testing attributes
    - key1: value1
    - key2: value2
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewDefaultLogger()
			logger.WithTimestamp(false)
			logger.WithIndent(tt.indentEnabled)

			output, err := testhelpers.CaptureStdout(func() {
				logger.Success("Testing attributes").WithAttrs(tt.attrs...)
			})
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			if output != tt.expectedOutput {
				t.Errorf("Unexpected output:\nGot:\n%q\nWant:\n%q", output, tt.expectedOutput)
			}
		})
	}
}

func TestDebugLogging(t *testing.T) {
	tests := []struct {
		name           string
		debugEnabled   bool
		message        string
		args           []any
		expectedOutput string
	}{
		{
			name:           "Debug message when debug is enabled",
			debugEnabled:   true,
			message:        "Processing request",
			args:           []any{"id:", 123},
			expectedOutput: "🐛 Processing request id: 123\n",
		},
		{
			name:           "Debug message when debug is disabled",
			debugEnabled:   false,
			message:        "This should not appear",
			args:           []any{"value:", 456},
			expectedOutput: "", // No output expected
		},
		{
			name:           "Debug with no args",
			debugEnabled:   true,
			message:        "Simple debug message",
			expectedOutput: "🐛 Simple debug message\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewDefaultLogger()
			logger.SetDebug(tt.debugEnabled)

			output, err := testhelpers.CaptureStdout(func() {
				logger.Debug(tt.message, tt.args...)
			})
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			if output != tt.expectedOutput {
				t.Errorf("Unexpected output:\nGot:\n%q\nWant:\n%q", output, tt.expectedOutput)
			}
		})
	}
}

func TestSetDebug(t *testing.T) {
	logger := NewDefaultLogger()

	// Initially debug should be disabled
	output, err := testhelpers.CaptureStdout(func() {
		logger.Debug("Should not appear")
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}
	if output != "" {
		t.Errorf("Expected no output when debug is disabled, got: %q", output)
	}

	// Enable debug
	logger.SetDebug(true)
	output, err = testhelpers.CaptureStdout(func() {
		logger.Debug("Should appear")
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}
	if !strings.Contains(output, "Should appear") {
		t.Errorf("Expected debug message to appear, got: %q", output)
	}

	// Disable debug again
	logger.SetDebug(false)
	output, err = testhelpers.CaptureStdout(func() {
		logger.Debug("Should not appear again")
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}
	if output != "" {
		t.Errorf("Expected no output when debug is disabled, got: %q", output)
	}
}

func TestDebugWithAttributes(t *testing.T) {
	logger := NewDefaultLogger()
	logger.SetDebug(true)

	output, err := testhelpers.CaptureStdout(func() {
		logger.Debug("Debug with attributes").WithAttrs("key1", "value1", "count", 42)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	expectedOutput := `🐛 Debug with attributes
  - key1: value1
  - count: 42
`
	if output != expectedOutput {
		t.Errorf("Unexpected output:\nGot:\n%q\nWant:\n%q", output, expectedOutput)
	}
}

func TestDebugWithIndentation(t *testing.T) {
	logger := NewDefaultLogger()
	logger.SetDebug(true)
	logger.WithIndent(true)

	output, err := testhelpers.CaptureStdout(func() {
		logger.Debug("Indented debug message")
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	expectedOutput := "  🐛 Indented debug message\n"
	if output != expectedOutput {
		t.Errorf("Unexpected output:\nGot:\n%q\nWant:\n%q", output, expectedOutput)
	}

	logger.WithIndent(false) // Reset
}
