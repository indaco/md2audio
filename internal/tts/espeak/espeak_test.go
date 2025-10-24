package espeak

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/indaco/md2audio/internal/tts"
)

func TestNewProvider(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test")
	}

	provider, err := NewProvider()
	if err != nil {
		t.Fatalf("Failed to create espeak provider: %v", err)
	}

	if provider == nil {
		t.Error("Expected non-nil provider")
	}

	if provider.Name() != "espeak" {
		t.Errorf("Expected provider name 'espeak', got %q", provider.Name())
	}
}

func TestNewProviderNonLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Skipping non-Linux test")
	}

	_, err := NewProvider()
	if err == nil {
		t.Error("Expected error on non-Linux platform")
	}

	// Verify error message is helpful
	expectedMsg := "espeak provider is only available on Linux"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message containing %q, got %q", expectedMsg, err.Error())
	}
}

// TestProviderName tests the Name() method
func TestProviderName(t *testing.T) {
	// Create a provider instance (even though NewProvider would fail on non-Linux)
	p := &Provider{}

	if p.Name() != "espeak" {
		t.Errorf("Provider.Name() = %q, want %q", p.Name(), "espeak")
	}
}

func TestMapVoiceToEspeak(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		note     string
	}{
		// British voices
		{"Kate", "en-gb", "British female"},
		{"Daniel", "en-gb", "British male"},
		{"Oliver", "en-gb", "British male alt"},
		{"Serena", "en-gb", "British female alt"},

		// US voices
		{"Samantha", "en-us", "US female"},
		{"Alex", "en-us", "US male"},
		{"Tom", "en-us", "US male alt"},
		{"Fiona", "en-us", "US female alt"},

		// Australian voices
		{"Karen", "en-au", "Australian female"},

		// Indian voices
		{"Veena", "en-in", "Indian female"},

		// Other language voices
		{"Thomas", "fr", "French"},
		{"Anna", "de", "German"},
		{"Monica", "es", "Spanish"},
		{"Alice", "it", "Italian"},
		{"Joana", "pt-pt", "Portuguese"},

		// Direct espeak codes (pass-through)
		{"en-us", "en-us", "Direct code"},
		{"en-gb", "en-gb", "Direct code"},
		{"en-au", "en-au", "Direct code"},
		{"fr", "fr", "Direct language code"},
		{"de", "de", "Direct language code"},
		{"pt-pt", "pt-pt", "Direct region code"},

		// Edge cases
		{"UnknownVoice", "en-us", "Unknown defaults to en-us"},
		{"", "en-us", "Empty string defaults to en-us"},
		{"random123", "en-us", "Random string defaults to en-us"},
		{"KATE", "en-us", "Case sensitive - no match"},
		{"kate", "en-us", "Lowercase - no match"},
	}

	for _, tt := range tests {
		t.Run(tt.input+"_"+tt.note, func(t *testing.T) {
			result := mapVoiceToEspeak(tt.input)
			if result != tt.expected {
				t.Errorf("mapVoiceToEspeak(%q) = %q, want %q (%s)", tt.input, result, tt.expected, tt.note)
			}
		})
	}
}

// TestMapVoiceToEspeakAllPresets ensures all voice presets are mapped
func TestMapVoiceToEspeakAllPresets(t *testing.T) {
	// Test that all common macOS voices have mappings
	macOSVoices := map[string]string{
		"Kate":     "en-gb",
		"Daniel":   "en-gb",
		"Samantha": "en-us",
		"Alex":     "en-us",
		"Karen":    "en-au",
		"Veena":    "en-in",
	}

	for voice, expectedLang := range macOSVoices {
		result := mapVoiceToEspeak(voice)
		if result != expectedLang {
			t.Errorf("Voice preset %q mapped to %q, want %q", voice, result, expectedLang)
		}
	}
}

// TestMapVoiceToEspeakRegexPattern tests the regex pattern matching
func TestMapVoiceToEspeakRegexPattern(t *testing.T) {
	tests := []struct {
		input       string
		shouldMatch bool
		description string
	}{
		{"en", true, "Two letter language code"},
		{"fr", true, "Two letter language code"},
		{"de", true, "Two letter language code"},
		{"en-us", true, "Language with region"},
		{"en-gb", true, "Language with region"},
		{"pt-pt", true, "Language with region"},
		{"fr-ca", true, "Language with region"},
		{"eng", false, "Three letter code"},
		{"en-USA", false, "Uppercase region"},
		{"EN-us", false, "Uppercase language"},
		{"en_us", false, "Underscore separator"},
		{"en-", false, "Trailing dash"},
		{"-gb", false, "Leading dash"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapVoiceToEspeak(tt.input)

			if tt.shouldMatch {
				// Should return the input as-is
				if result != tt.input {
					t.Errorf("Pattern %q should match and return %q, got %q", tt.input, tt.input, result)
				}
			} else {
				// Should return default
				if result != "en-us" {
					t.Errorf("Pattern %q should not match and return en-us, got %q", tt.input, result)
				}
			}
		})
	}
}

func TestListVoices(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test")
	}

	provider, err := NewProvider()
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	ctx := context.Background()
	voices, err := provider.ListVoices(ctx)
	if err != nil {
		t.Fatalf("ListVoices() failed: %v", err)
	}

	if len(voices) == 0 {
		t.Error("Expected at least one voice")
	}

	// Verify voice structure
	for i, voice := range voices {
		if voice.ID == "" {
			t.Errorf("Voice %d has empty ID", i)
		}
		if voice.Name == "" {
			t.Errorf("Voice %d has empty Name", i)
		}
		if voice.Language == "" {
			t.Errorf("Voice %d has empty Language", i)
		}
	}
}

func TestGenerate(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test")
	}

	provider, err := NewProvider()
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.wav")

	ctx := context.Background()
	req := tts.GenerateRequest{
		Text:       "Hello, this is a test.",
		Voice:      "en-us",
		OutputPath: outputPath,
		Format:     "wav",
	}

	result, err := provider.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	if result != outputPath {
		t.Errorf("Expected output path %q, got %q", outputPath, result)
	}

	// Verify file was created
	if _, err := os.Stat(result); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}

func TestGenerateWithMacOSVoice(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test")
	}

	provider, err := NewProvider()
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_kate.wav")

	ctx := context.Background()
	// Use macOS voice name - should be mapped to en-gb
	req := tts.GenerateRequest{
		Text:       "Testing voice mapping.",
		Voice:      "Kate",
		OutputPath: outputPath,
		Format:     "wav",
	}

	result, err := provider.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate() with macOS voice failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(result); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}

func TestGenerateWithRate(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test")
	}

	provider, err := NewProvider()
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_rate.wav")

	ctx := context.Background()
	rate := 150
	req := tts.GenerateRequest{
		Text:       "Testing custom rate.",
		Voice:      "en-us",
		OutputPath: outputPath,
		Rate:       &rate,
		Format:     "wav",
	}

	result, err := provider.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate() with custom rate failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(result); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}

func TestGenerateEmptyText(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test")
	}

	provider, err := NewProvider()
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_empty.wav")

	ctx := context.Background()
	req := tts.GenerateRequest{
		Text:       "   ",
		Voice:      "en-us",
		OutputPath: outputPath,
		Format:     "wav",
	}

	_, err = provider.Generate(ctx, req)
	if err == nil {
		t.Error("Expected error for empty text")
	}
}

func TestGenerateWithMarkdown(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test")
	}

	provider, err := NewProvider()
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_markdown.wav")

	ctx := context.Background()
	req := tts.GenerateRequest{
		Text:       "# Heading\n\nThis is **bold** text.",
		Voice:      "en-us",
		OutputPath: outputPath,
		Format:     "wav",
	}

	result, err := provider.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate() with markdown failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(result); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}
