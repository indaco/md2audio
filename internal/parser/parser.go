// Package parser provides markdown file parsing and section extraction functionality.
// It extracts H2 sections from markdown files, parses timing annotations,
// and discovers markdown files in directory trees.
//
// Key features:
//   - H2 section extraction from markdown files
//   - Timing annotation parsing (e.g., "## Scene 1 (5s)")
//   - Recursive markdown file discovery
//   - Input validation (file size, path safety)
//   - Directory structure mirroring for batch processing
package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/indaco/md2audio/internal/text"
)

const (
	// MaxFileSize is the maximum allowed markdown file size (10MB)
	MaxFileSize = 10 * 1024 * 1024 // 10MB should be more than enough for any reasonable markdown
)

// Pre-compiled regular expressions for performance
var (
	// Pattern to match H2 headers (##)
	h2Pattern = regexp.MustCompile(`^##\s+(.+)$`)

	// Pattern to extract timing from title: (0-8s) or (10s) or (8 seconds)
	timingPattern = regexp.MustCompile(`\((\d+(?:\.\d+)?)\s*(?:-\s*(\d+(?:\.\d+)?))?\s*s(?:ec(?:ond)?s?)?\)`)
)

// Section represents a markdown section with title and content
type Section struct {
	Title     string
	Content   string
	Duration  float64 // Target duration in seconds
	HasTiming bool    // Whether timing was specified
}

// validateMarkdownFile validates that a file is safe to read
func validateMarkdownFile(filename string) error {
	// Get file info
	info, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file: %s", filename)
	}

	// Check file size
	if info.Size() > MaxFileSize {
		return fmt.Errorf("file too large: %d bytes (max: %d bytes)", info.Size(), MaxFileSize)
	}

	// Validate file extension
	if filepath.Ext(filename) != ".md" {
		return fmt.Errorf("not a markdown file: %s", filename)
	}

	// Clean the path to prevent path traversal
	cleanPath, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Ensure the cleaned path still points to the same file
	// This prevents symlink attacks
	if cleanInfo, err := os.Stat(cleanPath); err != nil {
		return fmt.Errorf("failed to validate resolved path: %w", err)
	} else if !os.SameFile(info, cleanInfo) {
		return fmt.Errorf("path resolution mismatch (possible symlink attack)")
	}

	return nil
}

// parseTimingAnnotation extracts timing information from a title string.
// Returns the parsed duration, whether timing was found, and the title without timing.
func parseTimingAnnotation(titleWithTiming string) (duration float64, hasTiming bool, cleanTitle string) {
	timingMatch := timingPattern.FindStringSubmatch(titleWithTiming)
	if timingMatch == nil {
		return 0, false, titleWithTiming
	}

	// Try range format first: (0-8s) - use end time
	if len(timingMatch) >= 3 && timingMatch[2] != "" {
		if dur, err := parseFloat(timingMatch[2]); err == nil {
			cleanTitle = strings.TrimSpace(timingPattern.ReplaceAllString(titleWithTiming, ""))
			return dur, true, cleanTitle
		}
	}

	// Try single value format: (8s) or (10s)
	if len(timingMatch) >= 2 {
		if dur, err := parseFloat(timingMatch[1]); err == nil {
			cleanTitle = strings.TrimSpace(timingPattern.ReplaceAllString(titleWithTiming, ""))
			return dur, true, cleanTitle
		}
	}

	return 0, false, titleWithTiming
}

// saveSection saves a section with cleaned content to the sections slice.
// Returns the updated sections slice.
func saveSection(sections []Section, section *Section, contentLines []string) []Section {
	if section == nil {
		return sections
	}

	sectionText := strings.Join(contentLines, "\n")
	sectionText = text.CleanMarkdown(sectionText)
	if sectionText != "" {
		section.Content = sectionText
		sections = append(sections, *section)
	}

	return sections
}

// ParseMarkdownFile parses a markdown file and extracts H2 sections
func ParseMarkdownFile(filename string) ([]Section, error) {
	// Validate file before reading
	if err := validateMarkdownFile(filename); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	var sections []Section
	var currentSection *Section
	var contentLines []string

	for _, line := range lines {
		if match := h2Pattern.FindStringSubmatch(line); match != nil {
			// Save previous section if exists
			sections = saveSection(sections, currentSection, contentLines)

			// Start new section
			titleWithTiming := strings.TrimSpace(match[1])
			duration, hasTiming, cleanTitle := parseTimingAnnotation(titleWithTiming)

			currentSection = &Section{
				Title:     cleanTitle,
				Duration:  duration,
				HasTiming: hasTiming,
			}

			// Reset content lines for new section
			contentLines = []string{}
		} else if currentSection != nil {
			// Add line to current section content
			contentLines = append(contentLines, line)
		}
	}

	// Save last section
	sections = saveSection(sections, currentSection, contentLines)

	return sections, nil
}

// parseFloat parses a string to float64
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// MarkdownFile represents a discovered markdown file with its relative path
type MarkdownFile struct {
	AbsPath  string // Absolute path to the file
	RelPath  string // Relative path from base directory
	BaseDir  string // Base directory that was scanned
	FileName string // Just the filename without extension
}

// FindMarkdownFiles recursively finds all .md files in the given directory
func FindMarkdownFiles(baseDir string) ([]MarkdownFile, error) {
	var files []MarkdownFile

	// Get absolute path of base directory
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Walk the directory tree
	err = filepath.Walk(absBaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file has .md extension
		if filepath.Ext(path) == ".md" {
			// Get relative path from base directory
			relPath, err := filepath.Rel(absBaseDir, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}

			// Get filename without extension
			fileName := strings.TrimSuffix(filepath.Base(path), ".md")

			files = append(files, MarkdownFile{
				AbsPath:  path,
				RelPath:  relPath,
				BaseDir:  absBaseDir,
				FileName: fileName,
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}

// GetOutputDir returns the output directory path for this markdown file
// It creates a mirror structure based on the relative path
func (mf MarkdownFile) GetOutputDir(baseOutputDir string) string {
	// Get the directory containing the markdown file (relative to base)
	relDir := filepath.Dir(mf.RelPath)

	// If the file is in the root, use the filename as the directory
	if relDir == "." {
		return filepath.Join(baseOutputDir, mf.FileName)
	}

	// Otherwise, append the directory path and filename
	return filepath.Join(baseOutputDir, relDir, mf.FileName)
}
