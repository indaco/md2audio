package processor

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/indaco/md2audio/internal/config"
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
		Voice:    "Kate",
		Rate:     180,
		Format:   "aiff",
		Prefix:   "test",
	}

	err := ProcessFile(mdFile, outputDir, cfg)
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
		Voice:    "Kate",
		Rate:     180,
		Format:   "aiff",
		Prefix:   "test",
	}

	// Should not error, but should return 0 sections processed
	err := ProcessFile(mdFile, outputDir, cfg)
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
		Voice:    "Kate",
		Rate:     180,
		Format:   "aiff",
		Prefix:   "test",
	}

	err := ProcessFile(mdFile, outputDir, cfg)
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
		InputDir:  tmpDir,
		Voice:     "Kate",
		Rate:      180,
		Format:    "aiff",
		Prefix:    "test",
		OutputDir: outputDir,
	}

	err := ProcessDirectory(cfg)
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
		Voice:     "Kate",
		Rate:      180,
		Format:    "aiff",
		Prefix:    "test",
	}

	err := ProcessDirectory(cfg)
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
		Voice:     "Kate",
		Rate:      180,
		Format:    "aiff",
		Prefix:    "test",
	}

	err := ProcessDirectory(cfg)
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
				Voice:    "Kate",
				Rate:     180,
				Format:   tt.format,
				Prefix:   "test",
			}

			err := ProcessFile(mdFile, outputDir, cfg)
			if err != nil {
				t.Errorf("ProcessFile() with %s format error = %v", tt.format, err)
			}
		})
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
