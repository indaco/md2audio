package config

import (
	"bytes"
	"flag"
	"io"
	"os"
	"strings"
	"testing"
)

func TestVoicePresets(t *testing.T) {
	tests := []struct {
		preset   string
		expected string
	}{
		{"british-female", "Kate"},
		{"british-male", "Daniel"},
		{"us-female", "Samantha"},
		{"us-male", "Alex"},
		{"australian-female", "Karen"},
		{"indian-female", "Veena"},
	}

	for _, tt := range tests {
		t.Run(tt.preset, func(t *testing.T) {
			voice, ok := VoicePresets[tt.preset]
			if !ok {
				t.Errorf("Preset %q not found in VoicePresets", tt.preset)
			}
			if voice != tt.expected {
				t.Errorf("VoicePresets[%q] = %q, want %q", tt.preset, voice, tt.expected)
			}
		})
	}
}

func TestVoicePresetsCount(t *testing.T) {
	expectedCount := 6
	if len(VoicePresets) != expectedCount {
		t.Errorf("Expected %d voice presets, got %d", expectedCount, len(VoicePresets))
	}
}

func TestVoicePresetsAllUnique(t *testing.T) {
	seen := make(map[string]bool)
	for preset, voice := range VoicePresets {
		if seen[voice] {
			t.Errorf("Duplicate voice %q found (preset: %q)", voice, preset)
		}
		seen[voice] = true
	}
}

func TestConfigStructDefaults(t *testing.T) {
	cfg := Config{
		MarkdownFile: "test.md",
		OutputDir:    "./audio_sections",
		Voice:        "Kate",
		Rate:         180,
		Format:       "aiff",
		Prefix:       "section",
		ListVoices:   false,
	}

	// Verify default values
	if cfg.OutputDir != "./audio_sections" {
		t.Errorf("Expected default OutputDir './audio_sections', got %q", cfg.OutputDir)
	}
	if cfg.Voice != "Kate" {
		t.Errorf("Expected default Voice 'Kate', got %q", cfg.Voice)
	}
	if cfg.Rate != 180 {
		t.Errorf("Expected default Rate 180, got %d", cfg.Rate)
	}
	if cfg.Format != "aiff" {
		t.Errorf("Expected default Format 'aiff', got %q", cfg.Format)
	}
	if cfg.Prefix != "section" {
		t.Errorf("Expected default Prefix 'section', got %q", cfg.Prefix)
	}
}

func TestConfigPrint(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		checks []string // Strings that should appear in output
	}{
		{
			name: "standard config",
			config: Config{
				MarkdownFile: "test.md",
				Voice:        "Kate",
				Rate:         180,
				Format:       "aiff",
				OutputDir:    "./audio",
			},
			checks: []string{
				"Configuration:",
				"test.md",
				"Kate",
				"180",
				"aiff",
				"./audio",
			},
		},
		{
			name: "different voice and format",
			config: Config{
				MarkdownFile: "script.md",
				Voice:        "Samantha",
				Rate:         170,
				Format:       "m4a",
				OutputDir:    "./output",
			},
			checks: []string{
				"script.md",
				"Samantha",
				"170",
				"m4a",
				"./output",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call Print
			tt.config.Print()

			// Restore stdout
			if err := w.Close(); err != nil {
				t.Fatalf("Failed to close pipe: %v", err)
			}
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, r); err != nil {
				t.Fatalf("Failed to read output: %v", err)
			}
			output := buf.String()

			// Check for expected strings
			for _, check := range tt.checks {
				if !bytes.Contains(buf.Bytes(), []byte(check)) {
					t.Errorf("Expected output to contain %q, got:\n%s", check, output)
				}
			}
		})
	}
}

func TestConfigPrintFormat(t *testing.T) {
	cfg := Config{
		MarkdownFile: "test.md",
		Voice:        "Kate",
		Rate:         180,
		Format:       "aiff",
		OutputDir:    "./audio",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg.Print()

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close pipe: %v", err)
	}
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	output := buf.String()

	// Verify output starts with "Configuration:"
	if !bytes.HasPrefix(buf.Bytes(), []byte("\nConfiguration:")) {
		t.Errorf("Expected output to start with '\\nConfiguration:', got:\n%s", output)
	}

	// Verify output has proper formatting
	expectedFields := []string{
		"Markdown file:",
		"Voice:",
		"Rate:",
		"Format:",
		"Output directory:",
	}

	for _, field := range expectedFields {
		if !bytes.Contains(buf.Bytes(), []byte(field)) {
			t.Errorf("Expected output to contain field %q, got:\n%s", field, output)
		}
	}
}

func TestVoicePresetsAreMacOSVoices(t *testing.T) {
	// This test verifies that all preset voices are valid macOS voice names
	// These are well-known voice names that should exist on macOS systems
	expectedVoices := map[string]bool{
		"Kate":     true, // British English
		"Daniel":   true, // British English
		"Samantha": true, // US English
		"Alex":     true, // US English
		"Karen":    true, // Australian English
		"Veena":    true, // Indian English
	}

	for preset, voice := range VoicePresets {
		if !expectedVoices[voice] {
			t.Errorf("Preset %q maps to unexpected voice %q", preset, voice)
		}
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid file mode",
			config: Config{
				MarkdownFile: "test.md",
				ListVoices:   false,
			},
			expectError: false,
		},
		{
			name: "valid directory mode",
			config: Config{
				InputDir:   "./docs",
				ListVoices: false,
			},
			expectError: false,
		},
		{
			name: "both file and directory",
			config: Config{
				MarkdownFile: "test.md",
				InputDir:     "./docs",
				ListVoices:   false,
			},
			expectError: true,
			errorMsg:    "cannot use both -f and -d flags",
		},
		{
			name: "neither file nor directory",
			config: Config{
				MarkdownFile: "",
				InputDir:     "",
				ListVoices:   false,
			},
			expectError: true,
			errorMsg:    "either -f (file) or -d (directory) is required",
		},
		{
			name: "list voices mode ignores missing input",
			config: Config{
				MarkdownFile: "",
				InputDir:     "",
				ListVoices:   true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

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

func TestConfigIsDirectoryMode(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool
	}{
		{
			name:     "directory mode",
			config:   Config{InputDir: "./docs"},
			expected: true,
		},
		{
			name:     "file mode",
			config:   Config{MarkdownFile: "test.md"},
			expected: false,
		},
		{
			name:     "empty config",
			config:   Config{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsDirectoryMode()
			if result != tt.expected {
				t.Errorf("IsDirectoryMode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedVoice  string
		expectedRate   int
		expectedFormat string
		expectedPrefix string
		expectedFile   string
		expectedDir    string
		expectedOutput string
	}{
		{
			name:           "with british-female preset",
			args:           []string{"cmd", "-f", "test.md", "-p", "british-female"},
			expectedVoice:  "Kate",
			expectedRate:   180,
			expectedFormat: "aiff",
			expectedPrefix: "section",
			expectedFile:   "test.md",
			expectedOutput: "./audio_sections",
		},
		{
			name:           "with us-male preset",
			args:           []string{"cmd", "-f", "test.md", "-p", "us-male"},
			expectedVoice:  "Alex",
			expectedRate:   180,
			expectedFormat: "aiff",
			expectedPrefix: "section",
			expectedFile:   "test.md",
			expectedOutput: "./audio_sections",
		},
		{
			name:           "explicit voice overrides preset",
			args:           []string{"cmd", "-f", "test.md", "-p", "british-female", "-v", "Samantha"},
			expectedVoice:  "Samantha",
			expectedRate:   180,
			expectedFormat: "aiff",
			expectedPrefix: "section",
			expectedFile:   "test.md",
			expectedOutput: "./audio_sections",
		},
		{
			name:           "directory mode",
			args:           []string{"cmd", "-d", "./docs", "-p", "us-female"},
			expectedVoice:  "Samantha",
			expectedRate:   180,
			expectedFormat: "aiff",
			expectedPrefix: "section",
			expectedDir:    "./docs",
			expectedOutput: "./audio_sections",
		},
		{
			name:           "custom rate and format",
			args:           []string{"cmd", "-f", "test.md", "-p", "british-male", "-r", "150", "-format", "m4a"},
			expectedVoice:  "Daniel",
			expectedRate:   150,
			expectedFormat: "m4a",
			expectedPrefix: "section",
			expectedFile:   "test.md",
			expectedOutput: "./audio_sections",
		},
		{
			name:           "custom output and prefix",
			args:           []string{"cmd", "-f", "test.md", "-p", "australian-female", "-o", "./output", "-prefix", "audio"},
			expectedVoice:  "Karen",
			expectedRate:   180,
			expectedFormat: "aiff",
			expectedPrefix: "audio",
			expectedFile:   "test.md",
			expectedOutput: "./output",
		},
		{
			name:           "indian-female preset",
			args:           []string{"cmd", "-f", "test.md", "-p", "indian-female"},
			expectedVoice:  "Veena",
			expectedRate:   180,
			expectedFormat: "aiff",
			expectedPrefix: "section",
			expectedFile:   "test.md",
			expectedOutput: "./audio_sections",
		},
		{
			name:           "explicit voice only",
			args:           []string{"cmd", "-f", "test.md", "-v", "Daniel"},
			expectedVoice:  "Daniel",
			expectedRate:   180,
			expectedFormat: "aiff",
			expectedPrefix: "section",
			expectedFile:   "test.md",
			expectedOutput: "./audio_sections",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original os.Args and flag.CommandLine
			oldArgs := os.Args
			oldCommandLine := flag.CommandLine
			defer func() {
				os.Args = oldArgs
				flag.CommandLine = oldCommandLine
			}()

			// Reset flag.CommandLine for each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Set test args
			os.Args = tt.args

			// Capture stdout to suppress "No voice specified" messages
			old := os.Stdout
			_, w, _ := os.Pipe()
			os.Stdout = w

			// Parse config
			cfg := Parse()

			// Restore stdout
			_ = w.Close()
			os.Stdout = old

			// Verify results
			if cfg.Voice != tt.expectedVoice {
				t.Errorf("Voice = %q, want %q", cfg.Voice, tt.expectedVoice)
			}
			if cfg.Rate != tt.expectedRate {
				t.Errorf("Rate = %d, want %d", cfg.Rate, tt.expectedRate)
			}
			if cfg.Format != tt.expectedFormat {
				t.Errorf("Format = %q, want %q", cfg.Format, tt.expectedFormat)
			}
			if cfg.Prefix != tt.expectedPrefix {
				t.Errorf("Prefix = %q, want %q", cfg.Prefix, tt.expectedPrefix)
			}
			if cfg.MarkdownFile != tt.expectedFile {
				t.Errorf("MarkdownFile = %q, want %q", cfg.MarkdownFile, tt.expectedFile)
			}
			if cfg.InputDir != tt.expectedDir {
				t.Errorf("InputDir = %q, want %q", cfg.InputDir, tt.expectedDir)
			}
			if cfg.OutputDir != tt.expectedOutput {
				t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, tt.expectedOutput)
			}
		})
	}
}

func TestParseUnknownPreset(t *testing.T) {
	// Save original os.Args and flag.CommandLine
	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
	}()

	// Reset flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Set test args with unknown preset
	os.Args = []string{"cmd", "-f", "test.md", "-p", "unknown-preset"}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := Parse()

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close pipe: %v", err)
	}
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	output := buf.String()

	// Should default to Kate for unknown preset
	if cfg.Voice != "Kate" {
		t.Errorf("Voice = %q, want %q for unknown preset", cfg.Voice, "Kate")
	}

	// Should print warning message
	if !strings.Contains(output, "Unknown preset") {
		t.Errorf("Expected warning about unknown preset, got: %s", output)
	}
}

func TestParseNoVoiceSpecified(t *testing.T) {
	// Save original os.Args and flag.CommandLine
	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
	}()

	// Reset flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Set test args without voice or preset
	os.Args = []string{"cmd", "-f", "test.md"}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := Parse()

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close pipe: %v", err)
	}
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	output := buf.String()

	// Should default to Kate when no voice specified
	if cfg.Voice != "Kate" {
		t.Errorf("Voice = %q, want %q when no voice specified", cfg.Voice, "Kate")
	}

	// Should print default message
	if !strings.Contains(output, "No voice specified") && !strings.Contains(output, "using default: Kate") {
		t.Errorf("Expected message about default voice, got: %s", output)
	}
}
