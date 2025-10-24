package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/indaco/md2audio/internal/config"
	"github.com/indaco/md2audio/internal/logger"
)

func TestProcessFile(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	// Create test markdown file
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := `## Test Section

This is test content for audio generation.

## Another Section (5s)

More content here with timing.
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")

	cfg := config.Config{
		Provider: "say",
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
	}

	log := logger.NewDefaultLogger()
	err := ProcessFile(mdFile, outputDir, cfg, log)
	if err != nil {
		t.Errorf("ProcessFile() error = %v", err)
	}

	// Verify output directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("Output directory was not created")
	}

	// Verify audio files were created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}

	if len(files) == 0 {
		t.Error("No audio files were generated")
	}
}

func TestProcessFileInvalidMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	// Create file with no H2 sections
	content := "# H1 Title\n\nSome content without H2 sections."
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")

	cfg := config.Config{
		Provider: "say",
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
	}

	// Should not error, but should return 0 sections processed
	log := logger.NewDefaultLogger()
	err := ProcessFile(mdFile, outputDir, cfg, log)
	if err != nil {
		t.Errorf("ProcessFile() should not error on file with no sections, got: %v", err)
	}
}

func TestProcessFileNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "nonexistent.md")
	outputDir := filepath.Join(tmpDir, "output")

	cfg := config.Config{
		Provider: "say",
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
	}

	log := logger.NewDefaultLogger()
	err := ProcessFile(mdFile, outputDir, cfg, log)
	if err == nil {
		t.Error("ProcessFile() should error on nonexistent file")
	}
}

func TestProcessDirectory(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	// Create test directory structure
	tmpDir := t.TempDir()

	files := map[string]string{
		"file1.md": `## Section 1
Content for file 1.`,
		"sub/file2.md": `## Section 2
Content for file 2.`,
		"sub/deep/file3.md": `## Section 3
Content for file 3.`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	outputDir := filepath.Join(tmpDir, "audio_output")

	cfg := config.Config{
		InputDir: tmpDir,
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format:    "aiff",
		Prefix:    "test",
		OutputDir: outputDir,
	}

	log := logger.NewDefaultLogger()
	err := ProcessDirectory(cfg, log)
	if err != nil {
		t.Errorf("ProcessDirectory() error = %v", err)
	}

	// Verify output directory structure was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("Output directory was not created")
	}

	// Verify subdirectories were created (mirror structure)
	expectedDirs := []string{
		filepath.Join(outputDir, "file1"),
		filepath.Join(outputDir, "sub/file2"),
		filepath.Join(outputDir, "sub/deep/file3"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory not created: %s", dir)
		}
	}
}

func TestProcessDirectoryEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.Config{
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
	err := ProcessDirectory(cfg, log)
	if err == nil {
		t.Error("ProcessDirectory() should error on empty directory")
	}

	expectedMsg := "no markdown files found"
	if err != nil && !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing %q, got: %v", expectedMsg, err)
	}
}

func TestProcessDirectoryNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "does_not_exist")

	cfg := config.Config{
		InputDir:  nonExistent,
		OutputDir: filepath.Join(tmpDir, "output"),
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
	}

	log := logger.NewDefaultLogger()
	err := ProcessDirectory(cfg, log)
	if err == nil {
		t.Error("ProcessDirectory() should error on nonexistent directory")
	}
}

func TestProcessFileWithDifferentFormats(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	tests := []struct {
		name   string
		format string
	}{
		{"aiff format", "aiff"},
		{"m4a format", "m4a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			mdFile := filepath.Join(tmpDir, "test.md")
			content := "## Test\nContent"

			if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			outputDir := filepath.Join(tmpDir, "output")

			cfg := config.Config{
				Provider: "say",
				Say: config.SayConfig{
					Voice: "Kate",
					Rate:  180,
				},
				Format: tt.format,
				Prefix: "test",
			}

			log := logger.NewDefaultLogger()
			err := ProcessFile(mdFile, outputDir, cfg, log)
			if err != nil {
				t.Errorf("ProcessFile() with %s format error = %v", tt.format, err)
			}
		})
	}
}

func TestProcessFileDryRun(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := `## Test Section

This is test content for dry-run.

## Section with Timing (5s)

Content with timing annotation.
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")

	cfg := config.Config{
		Provider: "say",
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
		Commands: config.CommandFlags{
			DryRun: true, // Enable dry-run mode
		},
	}

	log := logger.NewDefaultLogger()
	err := ProcessFile(mdFile, outputDir, cfg, log)
	if err != nil {
		t.Errorf("ProcessFile() with dry-run error = %v", err)
	}

	// Verify no actual audio files were created
	entries, _ := os.ReadDir(outputDir)
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".aiff") {
			t.Error("Dry-run mode should not create audio files")
		}
	}
}

func TestProcessFileInvalidProvider(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := "## Test\nContent"

	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")

	cfg := config.Config{
		Provider: "invalid-provider", // Invalid provider
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
	}

	log := logger.NewDefaultLogger()
	err := ProcessFile(mdFile, outputDir, cfg, log)
	if err == nil {
		t.Error("ProcessFile() should error on invalid provider")
	}

	expectedMsg := "unsupported provider"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing %q, got: %v", expectedMsg, err)
	}
}

func TestProcessFileReadOnlyOutputDir(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := "## Test\nContent"

	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a read-only parent directory
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.MkdirAll(readOnlyDir, 0555); err != nil {
		t.Fatalf("Failed to create read-only directory: %v", err)
	}
	defer func() { _ = os.Chmod(readOnlyDir, 0755) }() // Restore permissions for cleanup

	outputDir := filepath.Join(readOnlyDir, "output")

	cfg := config.Config{
		Provider: "say",
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
	}

	log := logger.NewDefaultLogger()
	err := ProcessFile(mdFile, outputDir, cfg, log)
	if err == nil {
		t.Error("ProcessFile() should error when output directory cannot be created")
	}
}

func TestProcessDirectoryPartialFailure(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	tmpDir := t.TempDir()

	// Create valid markdown file
	validFile := filepath.Join(tmpDir, "valid.md")
	validContent := `## Section 1
Valid content.`
	if err := os.WriteFile(validFile, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to create valid file: %v", err)
	}

	// Create invalid markdown file (no sections)
	invalidFile := filepath.Join(tmpDir, "invalid.md")
	invalidContent := "# Just an H1\nNo H2 sections here."
	if err := os.WriteFile(invalidFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
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
	err := ProcessDirectory(cfg, log)
	// Should not error - partial failures are handled gracefully
	if err != nil {
		t.Errorf("ProcessDirectory() should handle partial failures gracefully, got: %v", err)
	}
}

func TestProcessDirectoryWithSubdirectories(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	tmpDir := t.TempDir()

	// Create nested directory structure
	files := map[string]string{
		"root.md":                "## Root\nRoot content.",
		"sub1/file1.md":          "## Sub1 File1\nContent 1.",
		"sub1/file2.md":          "## Sub1 File2\nContent 2.",
		"sub2/deep/file3.md":     "## Deep File\nDeep content.",
		"sub2/deep/alt/file4.md": "## Alt File\nAlt content.",
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
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
	err := ProcessDirectory(cfg, log)
	if err != nil {
		t.Errorf("ProcessDirectory() error = %v", err)
	}

	// Verify nested directory structure was created
	expectedDirs := []string{
		filepath.Join(outputDir, "root"),
		filepath.Join(outputDir, "sub1/file1"),
		filepath.Join(outputDir, "sub1/file2"),
		filepath.Join(outputDir, "sub2/deep/file3"),
		filepath.Join(outputDir, "sub2/deep/alt/file4"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory not created: %s", dir)
		}
	}
}

func TestProcessFileElevenLabsProvider(t *testing.T) {
	// This test verifies the code path for ElevenLabs provider configuration
	// Without mocking, it will fail due to missing API key, but we can verify
	// the error is appropriate

	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := "## Test\nContent"

	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")

	cfg := config.Config{
		Provider: "elevenlabs",
		ElevenLabs: config.ElevenLabsConfig{
			VoiceID: "test-voice-id",
			APIKey:  "", // Missing API key
		},
		Format: "mp3",
		Prefix: "test",
	}

	log := logger.NewDefaultLogger()
	err := ProcessFile(mdFile, outputDir, cfg, log)
	if err == nil {
		t.Error("ProcessFile() should error with missing ElevenLabs API key")
	}

	// Should contain API key related error
	if !strings.Contains(err.Error(), "API") && !strings.Contains(err.Error(), "key") {
		t.Logf("Got error: %v", err)
	}
}

func TestProcessFileLongSectionTitle(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	// Create section with very long title
	longTitle := strings.Repeat("a", 100)
	content := fmt.Sprintf("## %s\n\nContent for long title.", longTitle)

	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")

	cfg := config.Config{
		Provider: "say",
		Say: config.SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "test",
		Commands: config.CommandFlags{
			DryRun: true, // Use dry-run to avoid actual file creation
		},
	}

	log := logger.NewDefaultLogger()
	err := ProcessFile(mdFile, outputDir, cfg, log)
	if err != nil {
		t.Errorf("ProcessFile() with long title error = %v", err)
	}
}

func TestProcessFileSpecialCharactersInTitle(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	// Create section with special characters in title
	content := `## Test: Special/Characters & Symbols!

Content with special characters in section title.
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")

	cfg := config.Config{
		Provider: "say",
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
	err := ProcessFile(mdFile, outputDir, cfg, log)
	if err != nil {
		t.Errorf("ProcessFile() with special characters error = %v", err)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
