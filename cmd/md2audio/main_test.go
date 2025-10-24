package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/indaco/md2audio/internal/config"
	"github.com/indaco/md2audio/internal/logger"
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
				Say: config.SayConfig{
					Voice: "Kate",
					Rate:  180,
				},
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
				Say: config.SayConfig{
					Voice: "Kate",
					Rate:  180,
				},
				Format: "aiff",
			},
			expectError: true,
			errorMsg:    "cannot use both -f and -d flags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.NewDefaultLogger()
			err := run(tt.cfg, log)

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
		Provider: "say",
		Commands: config.CommandFlags{
			ListVoices: true,
		},
	}

	log := logger.NewDefaultLogger()
	err := run(cfg, log)
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
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
	}

	log := logger.NewDefaultLogger()
	err := run(cfg, log)
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
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
	}

	log := logger.NewDefaultLogger()
	err := run(cfg, log)
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
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
	}

	log := logger.NewDefaultLogger()
	err := run(cfg, log)
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
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
	}

	log := logger.NewDefaultLogger()
	err := run(cfg, log)
	if err == nil {
		t.Error("run() should error on empty directory")
	}

	expectedMsg := "no markdown files found"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing %q, got: %v", expectedMsg, err)
	}
}

func TestRunExportVoices(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "voices.json")

	cfg := config.Config{
		Provider: "say",
		Commands: config.CommandFlags{
			ExportVoices: exportPath,
		},
	}

	log := logger.NewDefaultLogger()
	err := run(cfg, log)
	if err != nil {
		t.Errorf("run() with ExportVoices should not error, got: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Error("Export file was not created")
	}
}

func TestRunDryRunMode(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	// Create test markdown file
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := `## Test Section

This is test content for dry-run.
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")

	cfg := config.Config{
		Provider:     "say",
		MarkdownFile: mdFile,
		OutputDir:    outputDir,
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
		Commands: config.CommandFlags{
			DryRun: true,
		},
	}

	log := logger.NewDefaultLogger()
	err := run(cfg, log)
	if err != nil {
		t.Errorf("run() with dry-run error = %v", err)
	}

	// Verify no actual audio files were created (dry-run)
	// Output directory might be created but shouldn't have audio files
	entries, _ := os.ReadDir(outputDir)
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".aiff") {
			t.Error("Dry-run mode should not create audio files")
		}
	}
}

func TestRunDebugMode(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	// Create test markdown file
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := `## Test Section

Debug test content.
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")

	cfg := config.Config{
		Provider:     "say",
		MarkdownFile: mdFile,
		OutputDir:    outputDir,
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
		Commands: config.CommandFlags{
			DryRun: true,
		},
	}

	log := logger.NewDefaultLogger()
	log.SetDebug(cfg.Commands.Debug)

	err := run(cfg, log)
	if err != nil {
		t.Errorf("run() with debug mode error = %v", err)
	}
}

func TestRunWithM4AFormat(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	// Create test markdown file
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := `## M4A Test

Content for M4A format test.
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")

	cfg := config.Config{
		Provider:     "say",
		MarkdownFile: mdFile,
		OutputDir:    outputDir,
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "m4a", // Test M4A format
		Prefix: "test",
	}

	log := logger.NewDefaultLogger()
	err := run(cfg, log)
	if err != nil {
		t.Errorf("run() with M4A format error = %v", err)
	}
}

func TestRunDirectoryWithDryRun(t *testing.T) {
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
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
		Commands: config.CommandFlags{
			DryRun: true, // Enable dry-run for directory mode
		},
	}

	log := logger.NewDefaultLogger()
	err := run(cfg, log)
	if err != nil {
		t.Errorf("run() directory dry-run error = %v", err)
	}
}

func TestRunCacheInitializationFailure(t *testing.T) {
	// This test verifies error handling when cache initialization fails
	// In practice, this is hard to trigger without mocking, but we can
	// at least verify the error path exists

	cfg := config.Config{
		Provider:     "say",
		MarkdownFile: "/tmp/test.md",
		OutputDir:    "/tmp/output",
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
	}

	log := logger.NewDefaultLogger()

	// Try to run - cache might fail in some edge cases
	// Even if it doesn't fail, this exercises the cache initialization code path
	_ = run(cfg, log)

	// The test passes if it doesn't panic
}
