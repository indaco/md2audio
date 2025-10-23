package say

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/indaco/md2audio/internal/tts"
)

func TestNewProvider(t *testing.T) {
	provider, err := NewProvider()

	if runtime.GOOS != "darwin" {
		if err == nil {
			t.Error("Expected error on non-macOS systems")
		}
		return
	}

	if err != nil {
		t.Errorf("Unexpected error on macOS: %v", err)
		return
	}

	if provider == nil {
		t.Error("Expected provider but got nil")
	}
}

func TestProvider_Name(t *testing.T) {
	provider := &Provider{}
	if got := provider.Name(); got != "say" {
		t.Errorf("Name() = %q, want %q", got, "say")
	}
}

func TestProvider_Generate(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	provider := &Provider{}
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		request     tts.GenerateRequest
		expectError bool
	}{
		{
			name: "basic generation with default rate",
			request: tts.GenerateRequest{
				Text:       "Hello world",
				Voice:      "Kate",
				OutputPath: filepath.Join(tmpDir, "test1.aiff"),
				Format:     "aiff",
			},
			expectError: false,
		},
		{
			name: "generation with custom rate",
			request: tts.GenerateRequest{
				Text:       "Hello world",
				Voice:      "Kate",
				OutputPath: filepath.Join(tmpDir, "test2.aiff"),
				Rate:       intPtr(200),
				Format:     "aiff",
			},
			expectError: false,
		},
		{
			name: "generation with markdown formatting",
			request: tts.GenerateRequest{
				Text:       "**Bold text** and _italic text_ with [links](http://example.com)",
				Voice:      "Kate",
				OutputPath: filepath.Join(tmpDir, "test3.aiff"),
				Format:     "aiff",
			},
			expectError: false,
		},
		{
			name: "empty text after markdown cleaning",
			request: tts.GenerateRequest{
				Text:       "   ",
				Voice:      "Kate",
				OutputPath: filepath.Join(tmpDir, "test4.aiff"),
				Format:     "aiff",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath, err := provider.Generate(context.Background(), tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify output file exists
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Errorf("Output file not created at %s", outputPath)
			}

			// Verify file has content
			info, err := os.Stat(outputPath)
			if err != nil {
				t.Errorf("Failed to stat output file: %v", err)
			} else if info.Size() == 0 {
				t.Error("Output file is empty")
			}
		})
	}
}

func TestProvider_GenerateM4A(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	provider := &Provider{}
	tmpDir := t.TempDir()

	req := tts.GenerateRequest{
		Text:       "Hello world",
		Voice:      "Kate",
		OutputPath: filepath.Join(tmpDir, "test.aiff"),
		Format:     "m4a",
	}

	outputPath, err := provider.Generate(context.Background(), req)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	// Verify output file is M4A
	if filepath.Ext(outputPath) != ".m4a" {
		t.Errorf("Expected .m4a extension, got %s", filepath.Ext(outputPath))
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("M4A file not created at %s", outputPath)
	}

	// Verify AIFF file was removed
	aiffPath := filepath.Join(tmpDir, "test.aiff")
	if _, err := os.Stat(aiffPath); !os.IsNotExist(err) {
		t.Error("AIFF file should have been removed after M4A conversion")
	}
}

func TestProvider_ListVoices(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	provider := &Provider{}
	voices, err := provider.ListVoices(context.Background())

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if len(voices) == 0 {
		t.Error("Expected at least one voice, got 0")
	}

	// Verify voice structure
	for _, voice := range voices {
		if voice.ID == "" {
			t.Error("Voice ID should not be empty")
		}
		if voice.Name == "" {
			t.Error("Voice Name should not be empty")
		}
		if voice.Language == "" {
			t.Error("Voice Language should not be empty")
		}
	}

	// Verify at least one known voice exists (Kate is standard on macOS)
	foundKate := false
	for _, voice := range voices {
		if voice.Name == "Kate" {
			foundKate = true
			if voice.Language != "en_GB" {
				t.Errorf("Kate voice should be en_GB, got %s", voice.Language)
			}
			break
		}
	}

	if !foundKate {
		t.Log("Warning: Kate voice not found in available voices")
	}
}

func TestGetAudioDuration(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	// Create a temporary audio file first
	tmpDir := t.TempDir()
	audioPath := filepath.Join(tmpDir, "test.aiff")

	// Generate a simple audio file
	provider := &Provider{}
	req := tts.GenerateRequest{
		Text:       "Test",
		Voice:      "Kate",
		OutputPath: audioPath,
		Format:     "aiff",
	}

	_, err := provider.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to generate test audio: %v", err)
	}

	// Test duration measurement
	duration, err := getAudioDuration(audioPath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if duration <= 0 {
		t.Errorf("Duration should be positive, got %f", duration)
	}

	// Reasonable duration for single word
	if duration > 10 {
		t.Errorf("Duration seems too long for single word: %f seconds", duration)
	}
}

func TestGetAudioDurationNonExistentFile(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	_, err := getAudioDuration("/nonexistent/path/file.aiff")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestConvertToM4A(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	// Create a temporary audio file first
	tmpDir := t.TempDir()
	aiffPath := filepath.Join(tmpDir, "test.aiff")

	// Generate a simple audio file
	provider := &Provider{}
	req := tts.GenerateRequest{
		Text:       "Test",
		Voice:      "Kate",
		OutputPath: aiffPath,
		Format:     "aiff",
	}

	_, err := provider.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to generate test audio: %v", err)
	}

	// Test conversion
	m4aPath := filepath.Join(tmpDir, "test.m4a")
	err = convertToM4A(context.Background(), aiffPath, m4aPath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	// Verify M4A file was created
	if _, err := os.Stat(m4aPath); os.IsNotExist(err) {
		t.Error("M4A file was not created")
	}

	// Verify file has content
	info, err := os.Stat(m4aPath)
	if err != nil {
		t.Errorf("Failed to stat M4A file: %v", err)
	} else if info.Size() == 0 {
		t.Error("M4A file is empty")
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}
