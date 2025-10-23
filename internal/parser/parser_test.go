package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMarkdownFile(t *testing.T) {
	tests := []struct {
		name              string
		markdown          string
		expectedCount     int
		expectedTitles    []string
		expectedTiming    []bool
		expectedDurations []float64
	}{
		{
			name: "basic sections without timing",
			markdown: `# H1 Title

## Section 1

This is content for section 1.

## Section 2

This is content for section 2.`,
			expectedCount:     2,
			expectedTitles:    []string{"Section 1", "Section 2"},
			expectedTiming:    []bool{false, false},
			expectedDurations: []float64{0, 0},
		},
		{
			name: "sections with timing annotations",
			markdown: `## Scene 1: Introduction (8s)

Content for scene 1.

## Scene 2: Main Demo (12s)

Content for scene 2.

## Scene 3: Conclusion

No timing for scene 3.`,
			expectedCount:     3,
			expectedTitles:    []string{"Scene 1: Introduction", "Scene 2: Main Demo", "Scene 3: Conclusion"},
			expectedTiming:    []bool{true, true, false},
			expectedDurations: []float64{8.0, 12.0, 0},
		},
		{
			name: "timing with range format",
			markdown: `## Scene 1 (0-8s)

Content here.

## Scene 2 (5-12.5s)

More content.`,
			expectedCount:     2,
			expectedTitles:    []string{"Scene 1", "Scene 2"},
			expectedTiming:    []bool{true, true},
			expectedDurations: []float64{8.0, 12.5},
		},
		{
			name: "timing with seconds spelled out",
			markdown: `## Test Section (15 seconds)

Content.`,
			expectedCount:     1,
			expectedTitles:    []string{"Test Section"},
			expectedTiming:    []bool{true},
			expectedDurations: []float64{15.0},
		},
		{
			name: "empty sections skipped",
			markdown: `## Section 1

This has content.

## Section 2

## Section 3

This also has content.`,
			expectedCount:     2,
			expectedTitles:    []string{"Section 1", "Section 3"},
			expectedTiming:    []bool{false, false},
			expectedDurations: []float64{0, 0},
		},
		{
			name: "no H2 sections",
			markdown: `# H1 Title

This is just regular content.

### H3 Subtitle

More content.`,
			expectedCount:     0,
			expectedTitles:    []string{},
			expectedTiming:    []bool{},
			expectedDurations: []float64{},
		},
		{
			name: "mixed content with markdown formatting",
			markdown: `## Test Section (10s)

This has **bold** and *italic* text with [links](https://example.com).

Also has ` + "`code`" + ` blocks.`,
			expectedCount:     1,
			expectedTitles:    []string{"Test Section"},
			expectedTiming:    []bool{true},
			expectedDurations: []float64{10.0},
		},
		{
			name: "decimal timing",
			markdown: `## Scene (8.5s)

Content.`,
			expectedCount:     1,
			expectedTitles:    []string{"Scene"},
			expectedTiming:    []bool{true},
			expectedDurations: []float64{8.5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(tmpFile, []byte(tt.markdown), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			// Parse the file
			sections, err := ParseMarkdownFile(tmpFile)
			if err != nil {
				t.Fatalf("ParseMarkdownFile() error = %v", err)
			}

			// Check section count
			if len(sections) != tt.expectedCount {
				t.Errorf("Expected %d sections, got %d", tt.expectedCount, len(sections))
			}

			// Check each section
			for i := 0; i < len(sections) && i < len(tt.expectedTitles); i++ {
				if sections[i].Title != tt.expectedTitles[i] {
					t.Errorf("Section %d: expected title %q, got %q", i, tt.expectedTitles[i], sections[i].Title)
				}

				if sections[i].HasTiming != tt.expectedTiming[i] {
					t.Errorf("Section %d: expected HasTiming=%v, got %v", i, tt.expectedTiming[i], sections[i].HasTiming)
				}

				if sections[i].Duration != tt.expectedDurations[i] {
					t.Errorf("Section %d: expected duration=%.1f, got %.1f", i, tt.expectedDurations[i], sections[i].Duration)
				}

				if sections[i].Content == "" {
					t.Errorf("Section %d: content should not be empty", i)
				}
			}
		})
	}
}

func TestParseMarkdownFileNotFound(t *testing.T) {
	_, err := ParseMarkdownFile("/nonexistent/file.md")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestParseMarkdownFileEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.md")
	if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	sections, err := ParseMarkdownFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseMarkdownFile() error = %v", err)
	}

	if len(sections) != 0 {
		t.Errorf("Expected 0 sections from empty file, got %d", len(sections))
	}
}

func TestSectionStructure(t *testing.T) {
	markdown := `## Test Section (10s)

This is test content.`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(tmpFile, []byte(markdown), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	sections, err := ParseMarkdownFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseMarkdownFile() error = %v", err)
	}

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(sections))
	}

	section := sections[0]

	// Verify all fields
	if section.Title == "" {
		t.Error("Section.Title should not be empty")
	}
	if section.Content == "" {
		t.Error("Section.Content should not be empty")
	}
	if !section.HasTiming {
		t.Error("Section.HasTiming should be true")
	}
	if section.Duration != 10.0 {
		t.Errorf("Expected duration 10.0, got %.1f", section.Duration)
	}
}

func TestFindMarkdownFiles(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"root.md":            "## Root\nContent",
		"sub/file1.md":       "## File 1\nContent",
		"sub/file2.md":       "## File 2\nContent",
		"sub/deep/nested.md": "## Nested\nContent",
		"other/readme.txt":   "Not a markdown file",
		"sub/.hidden.md":     "## Hidden\nContent",
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	// Find markdown files
	mdFiles, err := FindMarkdownFiles(tmpDir)
	if err != nil {
		t.Fatalf("FindMarkdownFiles() error = %v", err)
	}

	// Should find 5 .md files (excluding .txt)
	expectedCount := 5
	if len(mdFiles) != expectedCount {
		t.Errorf("Expected %d markdown files, got %d", expectedCount, len(mdFiles))
	}

	// Verify structure of found files
	for _, mf := range mdFiles {
		if mf.AbsPath == "" {
			t.Error("AbsPath should not be empty")
		}
		if mf.RelPath == "" {
			t.Error("RelPath should not be empty")
		}
		if mf.FileName == "" {
			t.Error("FileName should not be empty")
		}
		if !filepath.IsAbs(mf.AbsPath) {
			t.Errorf("AbsPath should be absolute: %s", mf.AbsPath)
		}
	}
}

func TestMarkdownFileGetOutputDir(t *testing.T) {
	tests := []struct {
		name          string
		relPath       string
		fileName      string
		baseOutputDir string
		expected      string
	}{
		{
			name:          "root level file",
			relPath:       "intro.md",
			fileName:      "intro",
			baseOutputDir: "/output",
			expected:      "/output/intro",
		},
		{
			name:          "single subdirectory",
			relPath:       "chapter1/content.md",
			fileName:      "content",
			baseOutputDir: "/output",
			expected:      "/output/chapter1/content",
		},
		{
			name:          "deeply nested",
			relPath:       "docs/api/v1/endpoints.md",
			fileName:      "endpoints",
			baseOutputDir: "/output",
			expected:      "/output/docs/api/v1/endpoints",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf := MarkdownFile{
				RelPath:  tt.relPath,
				FileName: tt.fileName,
			}

			result := mf.GetOutputDir(tt.baseOutputDir)
			if result != tt.expected {
				t.Errorf("GetOutputDir() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFindMarkdownFilesEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	mdFiles, err := FindMarkdownFiles(tmpDir)
	if err != nil {
		t.Fatalf("FindMarkdownFiles() error = %v", err)
	}

	if len(mdFiles) != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", len(mdFiles))
	}
}

func TestFindMarkdownFilesNonExistentDirectory(t *testing.T) {
	_, err := FindMarkdownFiles("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Expected error for nonexistent directory, got nil")
	}
}
