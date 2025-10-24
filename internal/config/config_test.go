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
		Say: SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format: "aiff",
		Prefix: "section",
	}

	// Verify default values
	if cfg.OutputDir != "./audio_sections" {
		t.Errorf("Expected default OutputDir './audio_sections', got %q", cfg.OutputDir)
	}
	if cfg.Say.Voice != "Kate" {
		t.Errorf("Expected default Voice 'Kate', got %q", cfg.Say.Voice)
	}
	if cfg.Say.Rate != 180 {
		t.Errorf("Expected default Rate 180, got %d", cfg.Say.Rate)
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
				Provider:     "say",
				Say: SayConfig{
					Voice: "Kate",
					Rate:  180,
				},
				Format:    "aiff",
				OutputDir: "./audio",
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
				Provider:     "say",
				Say: SayConfig{
					Voice: "Samantha",
					Rate:  170,
				},
				Format:    "m4a",
				OutputDir: "./output",
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
		Provider:     "say",
		Say: SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format:    "aiff",
		OutputDir: "./audio",
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
				Provider:     "say",
			},
			expectError: false,
		},
		{
			name: "valid directory mode",
			config: Config{
				InputDir: "./docs",
				Provider: "say",
			},
			expectError: false,
		},
		{
			name: "both file and directory",
			config: Config{
				MarkdownFile: "test.md",
				InputDir:     "./docs",
			},
			expectError: true,
			errorMsg:    "cannot use both -f and -d flags",
		},
		{
			name: "neither file nor directory",
			config: Config{
				MarkdownFile: "",
				InputDir:     "",
			},
			expectError: true,
			errorMsg:    "either -f (file) or -d (directory) is required",
		},
		{
			name: "list voices mode ignores missing input",
			config: Config{
				MarkdownFile: "",
				InputDir:     "",
				Provider:     "say",
				Commands: CommandFlags{
					ListVoices: true,
				},
			},
			expectError: false,
		},
		{
			name: "valid say provider",
			config: Config{
				MarkdownFile: "test.md",
				Provider:     "say",
			},
			expectError: false,
		},
		{
			name: "valid elevenlabs provider with voice ID",
			config: Config{
				MarkdownFile: "test.md",
				Provider:     "elevenlabs",
				ElevenLabs: ElevenLabsConfig{
					VoiceID: "21m00Tcm4TlvDq8ikWAM",
				},
			},
			expectError: false,
		},
		{
			name: "elevenlabs provider without voice ID",
			config: Config{
				MarkdownFile: "test.md",
				Provider:     "elevenlabs",
			},
			expectError: true,
			errorMsg:    "ElevenLabs voice ID is required",
		},
		{
			name: "invalid provider name",
			config: Config{
				MarkdownFile: "test.md",
				Provider:     "invalid-provider",
			},
			expectError: true,
			errorMsg:    "invalid provider",
		},
		{
			name: "elevenlabs list voices without voice ID is ok",
			config: Config{
				Provider: "elevenlabs",
				Commands: CommandFlags{
					ListVoices: true,
				},
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
			if cfg.Say.Voice != tt.expectedVoice {
				t.Errorf("Voice = %q, want %q", cfg.Say.Voice, tt.expectedVoice)
			}
			if cfg.Say.Rate != tt.expectedRate {
				t.Errorf("Rate = %d, want %d", cfg.Say.Rate, tt.expectedRate)
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
	if cfg.Say.Voice != "Kate" {
		t.Errorf("Voice = %q, want %q for unknown preset", cfg.Say.Voice, "Kate")
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
	if cfg.Say.Voice != "Kate" {
		t.Errorf("Voice = %q, want %q when no voice specified", cfg.Say.Voice, "Kate")
	}

	// Should print default message
	if !strings.Contains(output, "No voice specified") && !strings.Contains(output, "using default: Kate") {
		t.Errorf("Expected message about default voice, got: %s", output)
	}
}

// TestGetEnvFloat tests the getEnvFloat helper function
func TestGetEnvFloat(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue float64
		expected     float64
	}{
		{
			name:         "valid float value",
			envKey:       "TEST_FLOAT",
			envValue:     "0.75",
			defaultValue: 0.5,
			expected:     0.75,
		},
		{
			name:         "integer value",
			envKey:       "TEST_FLOAT_INT",
			envValue:     "1",
			defaultValue: 0.5,
			expected:     1.0,
		},
		{
			name:         "empty value uses default",
			envKey:       "TEST_FLOAT_EMPTY",
			envValue:     "",
			defaultValue: 0.5,
			expected:     0.5,
		},
		{
			name:         "invalid value uses default",
			envKey:       "TEST_FLOAT_INVALID",
			envValue:     "not-a-number",
			defaultValue: 0.5,
			expected:     0.5,
		},
		{
			name:         "unset variable uses default",
			envKey:       "TEST_FLOAT_UNSET",
			envValue:     "",
			defaultValue: 0.25,
			expected:     0.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set or unset env var
			if tt.envValue != "" {
				if err := os.Setenv(tt.envKey, tt.envValue); err != nil {
					t.Fatalf("Failed to set env var: %v", err)
				}
				defer func() {
					if err := os.Unsetenv(tt.envKey); err != nil {
						t.Errorf("Failed to unset env var: %v", err)
					}
				}()
			} else {
				if err := os.Unsetenv(tt.envKey); err != nil {
					t.Fatalf("Failed to unset env var: %v", err)
				}
			}

			result := getEnvFloat(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvFloat() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetEnvBool tests the getEnvBool helper function
func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "true value",
			envKey:       "TEST_BOOL_TRUE",
			envValue:     "true",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "false value",
			envKey:       "TEST_BOOL_FALSE",
			envValue:     "false",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "1 for true",
			envKey:       "TEST_BOOL_ONE",
			envValue:     "1",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "0 for false",
			envKey:       "TEST_BOOL_ZERO",
			envValue:     "0",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "empty value uses default",
			envKey:       "TEST_BOOL_EMPTY",
			envValue:     "",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "invalid value uses default",
			envKey:       "TEST_BOOL_INVALID",
			envValue:     "not-a-bool",
			defaultValue: false,
			expected:     false,
		},
		{
			name:         "unset variable uses default",
			envKey:       "TEST_BOOL_UNSET",
			envValue:     "",
			defaultValue: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set or unset env var
			if tt.envValue != "" {
				if err := os.Setenv(tt.envKey, tt.envValue); err != nil {
					t.Fatalf("Failed to set env var: %v", err)
				}
				defer func() {
					if err := os.Unsetenv(tt.envKey); err != nil {
						t.Errorf("Failed to unset env var: %v", err)
					}
				}()
			} else {
				if err := os.Unsetenv(tt.envKey); err != nil {
					t.Fatalf("Failed to unset env var: %v", err)
				}
			}

			result := getEnvBool(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvBool() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestMaskSecret tests the maskSecret helper function
func TestMaskSecret(t *testing.T) {
	tests := []struct {
		name     string
		secret   string
		expected string
	}{
		{
			name:     "empty string",
			secret:   "",
			expected: "[not set]",
		},
		{
			name:     "short string (8 chars or less)",
			secret:   "short",
			expected: "****",
		},
		{
			name:     "exactly 8 chars",
			secret:   "12345678",
			expected: "****",
		},
		{
			name:     "9 chars shows first and last 4",
			secret:   "123456789",
			expected: "1234****6789",
		},
		{
			name:     "typical API key",
			secret:   "fake_api_key_1234567890abcdefgh",
			expected: "fake****efgh",
		},
		{
			name:     "long key",
			secret:   "very_long_api_key_with_many_characters_here",
			expected: "very****here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskSecret(tt.secret)
			if result != tt.expected {
				t.Errorf("maskSecret(%q) = %q, want %q", tt.secret, result, tt.expected)
			}
		})
	}
}

// TestConfigPrintElevenLabs tests Print with ElevenLabs provider
func TestConfigPrintElevenLabs(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		expectMasked  bool
		expectedParts []string
	}{
		{
			name: "ElevenLabs with API key",
			config: Config{
				MarkdownFile: "test.md",
				Provider:     "elevenlabs",
				ElevenLabs: ElevenLabsConfig{
					VoiceID: "21m00Tcm4TlvDq8ikWAM",
					Model:   "eleven_multilingual_v2",
					APIKey:  "fake_api_key_1234567890abcdefgh",
				},
				Format:    "mp3",
				OutputDir: "./audio",
			},
			expectMasked: true,
			expectedParts: []string{
				"Configuration:",
				"elevenlabs",
				"21m00Tcm4TlvDq8ikWAM",
				"eleven_multilingual_v2",
				"fake****efgh", // Masked API key
				"mp3",
			},
		},
		{
			name: "ElevenLabs without API key shown",
			config: Config{
				MarkdownFile: "test.md",
				Provider:     "elevenlabs",
				ElevenLabs: ElevenLabsConfig{
					VoiceID: "21m00Tcm4TlvDq8ikWAM",
					Model:   "eleven_multilingual_v2",
					APIKey:  "", // No API key
				},
				Format:    "mp3",
				OutputDir: "./audio",
			},
			expectMasked: false,
			expectedParts: []string{
				"Configuration:",
				"elevenlabs",
				"21m00Tcm4TlvDq8ikWAM",
				"eleven_multilingual_v2",
				"mp3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			tt.config.Print()

			if err := w.Close(); err != nil {
				t.Fatalf("Failed to close pipe: %v", err)
			}
			os.Stdout = old

			var buf bytes.Buffer
			if _, err := io.Copy(&buf, r); err != nil {
				t.Fatalf("Failed to read output: %v", err)
			}
			output := buf.String()

			// Check for expected parts
			for _, part := range tt.expectedParts {
				if !strings.Contains(output, part) {
					t.Errorf("Expected output to contain %q, got:\n%s", part, output)
				}
			}

			// If we expect masking, ensure full API key is NOT in output
			if tt.expectMasked && tt.config.ElevenLabs.APIKey != "" {
				if strings.Contains(output, tt.config.ElevenLabs.APIKey) {
					t.Errorf("Full API key should not appear in output, got:\n%s", output)
				}
			}
		})
	}
}

// TestParseWithVersion tests Parse with version flag
func TestParseWithVersion(t *testing.T) {
	// Save original os.Args and flag.CommandLine
	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
	}()

	// Reset flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Set test args with version flag
	os.Args = []string{"cmd", "-version"}

	cfg := Parse()

	// Should return early with version flag set
	if !cfg.Commands.Version {
		t.Error("Expected Version flag to be true")
	}

	// Provider should have flag default value (set before early return)
	if cfg.Provider != "say" {
		t.Errorf("Expected Provider 'say' (flag default), got %q", cfg.Provider)
	}

	// Voice should be empty since voice selection logic is skipped
	if cfg.Say.Voice != "" {
		t.Errorf("Expected empty Voice (skipped initialization), got %q", cfg.Say.Voice)
	}
}

// TestParseElevenLabsDefaultVoice tests Parse with ElevenLabs and no voice ID
func TestParseElevenLabsDefaultVoice(t *testing.T) {
	// Save original os.Args and flag.CommandLine
	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
	}()

	// Reset flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Set test args with ElevenLabs provider but no voice ID
	os.Args = []string{"cmd", "-provider", "elevenlabs", "-f", "test.md"}

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

	// Should set default ElevenLabs voice
	if cfg.ElevenLabs.VoiceID != DefaultElevenLabsVoiceID {
		t.Errorf("Expected default voice %q, got %q", DefaultElevenLabsVoiceID, cfg.ElevenLabs.VoiceID)
	}

	// Should print default message
	if !strings.Contains(output, "No ElevenLabs voice specified") {
		t.Errorf("Expected message about default voice, got: %s", output)
	}
}

// restoreEnvVar restores an environment variable to its original value or unsets it
func restoreEnvVar(key, oldValue string) {
	if oldValue != "" {
		_ = os.Setenv(key, oldValue)
	} else {
		_ = os.Unsetenv(key)
	}
}

// TestParseElevenLabsWithEnvVars tests ElevenLabs voice settings from env vars
func TestParseElevenLabsWithEnvVars(t *testing.T) {
	// Save original env vars
	envVars := map[string]string{
		"ELEVENLABS_STABILITY":         os.Getenv("ELEVENLABS_STABILITY"),
		"ELEVENLABS_SIMILARITY_BOOST":  os.Getenv("ELEVENLABS_SIMILARITY_BOOST"),
		"ELEVENLABS_STYLE":             os.Getenv("ELEVENLABS_STYLE"),
		"ELEVENLABS_USE_SPEAKER_BOOST": os.Getenv("ELEVENLABS_USE_SPEAKER_BOOST"),
		"ELEVENLABS_SPEED":             os.Getenv("ELEVENLABS_SPEED"),
	}
	defer func() {
		for key, value := range envVars {
			restoreEnvVar(key, value)
		}
	}()

	// Set custom env vars
	if err := os.Setenv("ELEVENLABS_STABILITY", "0.75"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	if err := os.Setenv("ELEVENLABS_SIMILARITY_BOOST", "0.85"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	if err := os.Setenv("ELEVENLABS_STYLE", "0.3"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	if err := os.Setenv("ELEVENLABS_USE_SPEAKER_BOOST", "false"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	if err := os.Setenv("ELEVENLABS_SPEED", "1.1"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}

	// Save original os.Args and flag.CommandLine
	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
	}()

	// Reset flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Set test args with ElevenLabs provider
	os.Args = []string{"cmd", "-provider", "elevenlabs", "-elevenlabs-voice-id", "test-voice", "-f", "test.md"}

	// Capture stdout
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	cfg := Parse()

	_ = w.Close()
	os.Stdout = old

	// Verify env vars were loaded
	if cfg.ElevenLabs.VoiceSettings.Stability != 0.75 {
		t.Errorf("Expected Stability 0.75, got %v", cfg.ElevenLabs.VoiceSettings.Stability)
	}
	if cfg.ElevenLabs.VoiceSettings.SimilarityBoost != 0.85 {
		t.Errorf("Expected SimilarityBoost 0.85, got %v", cfg.ElevenLabs.VoiceSettings.SimilarityBoost)
	}
	if cfg.ElevenLabs.VoiceSettings.Style != 0.3 {
		t.Errorf("Expected Style 0.3, got %v", cfg.ElevenLabs.VoiceSettings.Style)
	}
	if cfg.ElevenLabs.VoiceSettings.UseSpeakerBoost != false {
		t.Errorf("Expected UseSpeakerBoost false, got %v", cfg.ElevenLabs.VoiceSettings.UseSpeakerBoost)
	}
	if cfg.ElevenLabs.VoiceSettings.Speed != 1.1 {
		t.Errorf("Expected Speed 1.1, got %v", cfg.ElevenLabs.VoiceSettings.Speed)
	}
}

// TestParseListVoicesWithSayProvider tests list-voices with say provider
func TestParseListVoicesWithSayProvider(t *testing.T) {
	// Save original os.Args and flag.CommandLine
	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
	}()

	// Reset flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Set test args with list-voices flag
	os.Args = []string{"cmd", "-list-voices"}

	// Capture stdout
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	cfg := Parse()

	_ = w.Close()
	os.Stdout = old

	// Should have list-voices set
	if !cfg.Commands.ListVoices {
		t.Error("Expected ListVoices to be true")
	}

	// Voice should not be set when listing voices
	if cfg.Say.Voice != "" {
		t.Errorf("Expected empty Voice when listing, got %q", cfg.Say.Voice)
	}
}

// TestParseElevenLabsListVoices tests ElevenLabs with list-voices (should not require voice ID)
func TestParseElevenLabsListVoices(t *testing.T) {
	// Save original os.Args and flag.CommandLine
	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
	}()

	// Reset flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Set test args with ElevenLabs and list-voices (no voice ID required)
	os.Args = []string{"cmd", "-provider", "elevenlabs", "-list-voices"}

	// Capture stdout
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	cfg := Parse()

	_ = w.Close()
	os.Stdout = old

	// Should not set default voice when listing
	if cfg.ElevenLabs.VoiceID != "" {
		t.Errorf("Expected empty voice ID when listing, got %q", cfg.ElevenLabs.VoiceID)
	}

	// Should set ListVoices
	if !cfg.Commands.ListVoices {
		t.Error("Expected ListVoices to be true")
	}
}

// TestParseWithDryRunFlag tests Parse with dry-run flag
func TestParseWithDryRunFlag(t *testing.T) {
	// Save original os.Args and flag.CommandLine
	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
	}()

	// Reset flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Set test args with dry-run flag
	os.Args = []string{"cmd", "-f", "test.md", "-dry-run"}

	// Capture stdout
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	cfg := Parse()

	_ = w.Close()
	os.Stdout = old

	// Should have DryRun set
	if !cfg.Commands.DryRun {
		t.Error("Expected DryRun to be true")
	}
}

// TestConfigPrintDirectoryMode tests Print with directory mode
func TestConfigPrintDirectoryMode(t *testing.T) {
	cfg := Config{
		InputDir: "./docs",
		Provider: "say",
		Say: SayConfig{
			Voice: "Kate",
			Rate:  180,
		},
		Format:    "aiff",
		OutputDir: "./audio",
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

	// Read captured output
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	output := buf.String()

	// Should contain "Input directory" instead of "Markdown file"
	if !strings.Contains(output, "Input directory") {
		t.Errorf("Expected output to contain 'Input directory', got:\n%s", output)
	}
	if strings.Contains(output, "Markdown file") {
		t.Errorf("Should not contain 'Markdown file' in directory mode, got:\n%s", output)
	}
}
