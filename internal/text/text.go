package text

import (
	"regexp"
	"strings"
)

// CleanMarkdown removes markdown formatting from text for speech synthesis
func CleanMarkdown(text string) string {
	// Remove extra whitespace and newlines
	text = regexp.MustCompile(`\n+`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Remove markdown links [text](url) -> text
	text = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`).ReplaceAllString(text, "$1")

	// Remove bold/italic markers
	text = regexp.MustCompile(`[*_]{1,2}([^*_]+)[*_]{1,2}`).ReplaceAllString(text, "$1")

	// Remove code blocks
	text = regexp.MustCompile("`[^`]+`").ReplaceAllString(text, "")

	return strings.TrimSpace(text)
}

// SanitizeFilename converts a title into a safe filename
func SanitizeFilename(title string) string {
	// Remove or replace invalid characters
	reg := regexp.MustCompile(`[^\w\s-]`)
	filename := reg.ReplaceAllString(title, "")

	// Replace spaces with underscores
	filename = regexp.MustCompile(`\s+`).ReplaceAllString(filename, "_")

	filename = strings.ToLower(filename)

	// Limit length
	if len(filename) > 50 {
		filename = filename[:50]
	}

	return filename
}
