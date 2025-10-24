// Package text provides text processing utilities for markdown content.
// It includes functions for cleaning markdown formatting and sanitizing filenames.
//
// Key features:
//   - Markdown formatting removal for TTS compatibility
//   - Safe filename generation from section titles
//   - Pre-compiled regex patterns for performance
package text

import (
	"regexp"
	"strings"
)

// Pre-compiled regular expressions for performance
// These are compiled once at package initialization instead of on every function call
var (
	// Markdown cleaning patterns
	newlinePattern      = regexp.MustCompile(`\n+`)
	whitespacePattern   = regexp.MustCompile(`\s+`)
	markdownLinkPattern = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
	boldItalicPattern   = regexp.MustCompile(`[*_]{1,2}([^*_]+)[*_]{1,2}`)
	codeBlockPattern    = regexp.MustCompile("`[^`]+`")

	// Filename sanitization patterns
	invalidCharsPattern = regexp.MustCompile(`[^\w\s-]`)
)

// CleanMarkdown removes markdown formatting from text for speech synthesis
func CleanMarkdown(text string) string {
	// Remove extra whitespace and newlines
	text = newlinePattern.ReplaceAllString(text, " ")
	text = whitespacePattern.ReplaceAllString(text, " ")

	// Remove markdown links [text](url) -> text
	text = markdownLinkPattern.ReplaceAllString(text, "$1")

	// Remove bold/italic markers
	text = boldItalicPattern.ReplaceAllString(text, "$1")

	// Remove code blocks
	text = codeBlockPattern.ReplaceAllString(text, "")

	return strings.TrimSpace(text)
}

// SanitizeFilename converts a title into a safe filename
func SanitizeFilename(title string) string {
	// Remove or replace invalid characters
	filename := invalidCharsPattern.ReplaceAllString(title, "")

	// Replace spaces with underscores
	filename = whitespacePattern.ReplaceAllString(filename, "_")

	filename = strings.ToLower(filename)

	// Limit length
	if len(filename) > 50 {
		filename = filename[:50]
	}

	return filename
}
