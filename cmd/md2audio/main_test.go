package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/indaco/md2audio/internal/config"
)

func TestRunValidation(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing both file and directory",
			cfg: config.Config{
				Voice:  "Kate",
				Rate:   180,
				Format: "aiff",
			},
			expectError: true,
			errorMsg:    "either -f (file) or -d (directory) is required",
		},
		{
			name: "both file and directory specified",
			cfg: config.Config{
				MarkdownFile: "test.md",
				InputDir:     "./docs",
				Voice:        "Kate",
				Rate:         180,
				Format:       "aiff",
			},
			expectError: true,
			errorMsg:    "cannot use both -f and -d flags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.cfg)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestRunListVoices(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	cfg := config.Config{
		Provider:   "say",
		ListVoices: true,
	}

	err := run(cfg)
	if err != nil {
		t.Errorf("run() with ListVoices should not error, got: %v", err)
	}
}

func TestRunProcessFile(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	// Create test markdown file
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := `## Test Section

This is test content for audio generation.
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")

	cfg := config.Config{
		Provider:     "say",
		MarkdownFile: mdFile,
		OutputDir:    outputDir,
		Voice:        "Kate",
		Rate:         180,
		Format:       "aiff",
		Prefix:       "test",
	}

	err := run(cfg)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	// Verify output directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("Output directory was not created")
	}
}

func TestRunProcessDirectory(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	// Create test directory structure
	tmpDir := t.TempDir()

	mdFile := filepath.Join(tmpDir, "test.md")
	content := `## Section 1
Content for test.
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "audio_output")

	cfg := config.Config{
		Provider:  "say",
		InputDir:  tmpDir,
		OutputDir: outputDir,
		Voice:     "Kate",
		Rate:      180,
		Format:    "aiff",
		Prefix:    "test",
	}

	err := run(cfg)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	// Verify output directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("Output directory was not created")
	}
}

func TestRunNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "nonexistent.md")

	cfg := config.Config{
		Provider:     "say",
		MarkdownFile: nonExistent,
		OutputDir:    filepath.Join(tmpDir, "output"),
		Voice:        "Kate",
		Rate:         180,
		Format:       "aiff",
		Prefix:       "test",
	}

	err := run(cfg)
	if err == nil {
		t.Error("run() should error on nonexistent file")
	}
}

func TestRunEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.Config{
		Provider:  "say",
		InputDir:  tmpDir,
		OutputDir: filepath.Join(tmpDir, "output"),
		Voice:     "Kate",
		Rate:      180,
		Format:    "aiff",
		Prefix:    "test",
	}

	err := run(cfg)
	if err == nil {
		t.Error("run() should error on empty directory")
	}

	expectedMsg := "no markdown files found"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing %q, got: %v", expectedMsg, err)
	}
}
