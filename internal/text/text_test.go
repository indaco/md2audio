package text

import "testing"

func TestCleanMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic text unchanged",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "removes multiple newlines",
			input:    "Line 1\n\n\nLine 2",
			expected: "Line 1 Line 2",
		},
		{
			name:     "removes markdown links",
			input:    "Check out [this link](https://example.com)",
			expected: "Check out this link",
		},
		{
			name:     "removes bold markers",
			input:    "This is **bold text** here",
			expected: "This is bold text here",
		},
		{
			name:     "removes italic markers",
			input:    "This is *italic text* here",
			expected: "This is italic text here",
		},
		{
			name:     "removes italic underscore markers",
			input:    "This is _italic text_ here",
			expected: "This is italic text here",
		},
		{
			name:     "removes code blocks",
			input:    "Run `npm install` command",
			expected: "Run  command",
		},
		{
			name:     "removes extra whitespace",
			input:    "Too    many     spaces",
			expected: "Too many spaces",
		},
		{
			name:     "complex markdown",
			input:    "This is **bold** and *italic* with [a link](https://example.com) and `code`\n\nNew paragraph",
			expected: "This is bold and italic with a link and  New paragraph",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \n\n   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("CleanMarkdown(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic lowercase conversion",
			input:    "HelloWorld",
			expected: "helloworld",
		},
		{
			name:     "spaces to underscores",
			input:    "My File Name",
			expected: "my_file_name",
		},
		{
			name:     "removes special characters",
			input:    "File@Name#With$Special%Chars",
			expected: "filenamewithspecialchars",
		},
		{
			name:     "keeps hyphens",
			input:    "my-file-name",
			expected: "my-file-name",
		},
		{
			name:     "scene with colon and parens",
			input:    "Scene 1: Introduction (8s)",
			expected: "scene_1_introduction_8s",
		},
		{
			name:     "multiple spaces collapsed",
			input:    "Too    Many     Spaces",
			expected: "too_many_spaces",
		},
		{
			name:     "long filename truncated",
			input:    "This is a very long filename that should be truncated to fifty characters max",
			expected: "this_is_a_very_long_filename_that_should_be_trunca",
		},
		{
			name:     "unicode characters removed",
			input:    "File with émojis 🎉",
			expected: "file_with_mojis_",
		},
		{
			name:     "already clean filename",
			input:    "clean_filename_123",
			expected: "clean_filename_123",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "numbers and underscores",
			input:    "Section_01_Test_123",
			expected: "section_01_test_123",
		},
		{
			name:     "dots and slashes removed",
			input:    "path/to/file.txt",
			expected: "pathtofiletxt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilenameLength(t *testing.T) {
	longInput := "a" // Start with one character
	for range 100 {
		longInput += "a"
	}

	result := SanitizeFilename(longInput)
	if len(result) > 50 {
		t.Errorf("SanitizeFilename should truncate to 50 chars, got %d chars: %q", len(result), result)
	}
}
