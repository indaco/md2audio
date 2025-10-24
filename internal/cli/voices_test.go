package cli

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/indaco/md2audio/internal/cache"
	"github.com/indaco/md2audio/internal/config"
	"github.com/indaco/md2audio/internal/logger"
	"github.com/indaco/md2audio/internal/testhelpers"
	"github.com/indaco/md2audio/internal/tts"
)

func TestCreateProvider(t *testing.T) {
	tests := []struct {
		name            string
		cfg             config.Config
		expectError     bool
		expectedName    string
		skipOnNonDarwin bool
	}{
		{
			name: "say provider",
			cfg: config.Config{
				Provider: "say",
			},
			expectError:     false,
			expectedName:    "say",
			skipOnNonDarwin: true,
		},
		{
			name: "default provider (say)",
			cfg: config.Config{
				Provider: "",
			},
			expectError:     false,
			expectedName:    "say",
			skipOnNonDarwin: true,
		},
		{
			name: "elevenlabs provider with API key",
			cfg: config.Config{
				Provider:         "elevenlabs",
				ElevenLabsAPIKey: "test-key-123",
			},
			expectError:  false,
			expectedName: "elevenlabs",
		},
		{
			name: "elevenlabs provider without API key",
			cfg: config.Config{
				Provider: "elevenlabs",
			},
			expectError: true,
		},
		{
			name: "unsupported provider",
			cfg: config.Config{
				Provider: "unknown",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOnNonDarwin && runtime.GOOS != "darwin" {
				t.Skip("Skipping macOS-specific test")
			}

			provider, err := CreateProvider(tt.cfg)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if provider == nil {
				t.Error("Provider is nil")
				return
			}

			if provider.Name() != tt.expectedName {
				t.Errorf("Expected provider name %q, got %q", tt.expectedName, provider.Name())
			}
		})
	}
}

func TestDisplayVoicesOutput(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	voices := []tts.Voice{
		{
			ID:          "voice1",
			Name:        "Kate",
			Language:    "en_US",
			Description: "American female voice",
		},
		{
			ID:          "voice2",
			Name:        "Daniel",
			Language:    "en_GB",
			Description: "British male voice",
		},
	}

	tests := []struct {
		name         string
		providerName string
		voices       []tts.Voice
		checkOutput  func(string) bool
	}{
		{
			name:         "say provider simple format",
			providerName: "say",
			voices:       voices,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "Kate") &&
					strings.Contains(output, "en_US") &&
					strings.Contains(output, "Daniel")
			},
		},
		{
			name:         "elevenlabs provider with IDs",
			providerName: "elevenlabs",
			voices:       voices,
			checkOutput: func(output string) bool {
				return strings.Contains(output, "voice1") &&
					strings.Contains(output, "voice2") &&
					strings.Contains(output, "Kate") &&
					strings.Contains(output, "ID")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.NewDefaultLogger()

			output, err := testhelpers.CaptureStdout(func() {
				displayVoices(tt.providerName, tt.voices, log)
			})
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			if !tt.checkOutput(output) {
				t.Errorf("Output validation failed.\nGot:\n%s", output)
			}
		})
	}
}

func TestDisplayElevenLabsVoicesLongDescription(t *testing.T) {
	voices := []tts.Voice{
		{
			ID:          "voice1",
			Name:        "TestVoice",
			Language:    "en",
			Description: "This is a very long description that should be truncated because it exceeds the maximum length allowed",
		},
	}

	log := logger.NewDefaultLogger()

	output, err := testhelpers.CaptureStdout(func() {
		displayElevenLabsVoices(voices, log)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	if !strings.Contains(output, "...") {
		t.Error("Long description should be truncated with '...'")
	}

	if len(strings.Split(output, "\n")[1]) > 150 {
		t.Error("Truncated line is still too long")
	}
}

func TestDisplaySimpleVoicesWithAndWithoutDescription(t *testing.T) {
	voices := []tts.Voice{
		{
			Name:        "Voice1",
			Language:    "en_US",
			Description: "Has description",
		},
		{
			Name:     "Voice2",
			Language: "en_GB",
			// No description
		},
	}

	log := logger.NewDefaultLogger()

	output, err := testhelpers.CaptureStdout(func() {
		displaySimpleVoices(voices, log)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	if !strings.Contains(output, "Voice1") || !strings.Contains(output, "Voice2") {
		t.Error("Output should contain both voices")
	}

	if !strings.Contains(output, "Has description") {
		t.Error("Output should contain voice description when available")
	}
}

func TestExportVoices(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "voices.json")

	// Create voice cache
	voiceCache, err := cache.NewVoiceCache()
	if err != nil {
		t.Fatalf("Failed to create voice cache: %v", err)
	}
	defer func() {
		if err := voiceCache.Close(); err != nil {
			t.Logf("Warning: failed to close voice cache: %v", err)
		}
	}()

	// Create provider
	cfg := config.Config{
		Provider: "say",
	}
	provider, err := CreateProvider(cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	cachedProvider := cache.NewCachedProvider(provider, voiceCache)
	ctx := context.Background()

	// Get voices to populate cache
	_, err = cachedProvider.ListVoices(ctx)
	if err != nil {
		t.Fatalf("Failed to list voices: %v", err)
	}

	log := logger.NewDefaultLogger()

	// Test export
	err = ExportVoices(ctx, cachedProvider, provider.Name(), outputPath, log)
	if err != nil {
		t.Errorf("ExportVoices() failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Export file was not created")
	}

	// Verify file is valid JSON
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Export file is empty")
	}

	if !strings.HasPrefix(string(data), "[") {
		t.Error("Export file does not appear to be valid JSON array")
	}
}

func TestExportVoicesNoVoices(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "voices.json")

	// Create empty voice cache
	voiceCache, err := cache.NewVoiceCache()
	if err != nil {
		t.Fatalf("Failed to create voice cache: %v", err)
	}
	defer func() {
		if err := voiceCache.Close(); err != nil {
			t.Logf("Warning: failed to close voice cache: %v", err)
		}
	}()

	cfg := config.Config{
		Provider: "say",
	}
	provider, err := CreateProvider(cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	cachedProvider := cache.NewCachedProvider(provider, voiceCache)
	ctx := context.Background()

	// Clear cache to ensure no voices
	if err := voiceCache.Clear(ctx, provider.Name()); err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	log := logger.NewDefaultLogger()

	// Test export with no voices - should still work as it fetches from provider
	err = ExportVoices(ctx, cachedProvider, provider.Name(), outputPath, log)
	if err != nil {
		t.Logf("ExportVoices() error (expected if no voices available): %v", err)
	}
}

func TestListVoices(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	voiceCache, err := cache.NewVoiceCache()
	if err != nil {
		t.Fatalf("Failed to create voice cache: %v", err)
	}
	defer func() {
		if err := voiceCache.Close(); err != nil {
			t.Logf("Warning: failed to close voice cache: %v", err)
		}
	}()

	cfg := config.Config{
		Provider: "say",
	}
	provider, err := CreateProvider(cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	cachedProvider := cache.NewCachedProvider(provider, voiceCache)
	ctx := context.Background()
	log := logger.NewDefaultLogger()

	tests := []struct {
		name         string
		refreshCache bool
	}{
		{
			name:         "list voices from cache",
			refreshCache: false,
		},
		{
			name:         "list voices with refresh",
			refreshCache: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := testhelpers.CaptureStdout(func() {
				err := ListVoices(ctx, cachedProvider, provider.Name(), tt.refreshCache, log)
				if err != nil {
					t.Errorf("ListVoices() error = %v", err)
				}
			})
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			if !strings.Contains(output, "Available voices") {
				t.Error("Output should contain 'Available voices'")
			}

			if tt.refreshCache && !strings.Contains(output, "Refreshing") {
				t.Error("Output should mention refreshing cache")
			}
		})
	}
}

func TestHandleVoiceCommands(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	voiceCache, err := cache.NewVoiceCache()
	if err != nil {
		t.Fatalf("Failed to create voice cache: %v", err)
	}
	defer func() {
		if err := voiceCache.Close(); err != nil {
			t.Logf("Warning: failed to close voice cache: %v", err)
		}
	}()

	log := logger.NewDefaultLogger()

	t.Run("list voices", func(t *testing.T) {
		cfg := config.Config{
			Provider:   "say",
			ListVoices: true,
		}

		err := HandleVoiceCommands(cfg, voiceCache, log)
		if err != nil {
			t.Errorf("HandleVoiceCommands() error = %v", err)
		}
	})

	t.Run("export voices", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "voices.json")

		cfg := config.Config{
			Provider:     "say",
			ExportVoices: outputPath,
		}

		err := HandleVoiceCommands(cfg, voiceCache, log)
		if err != nil {
			t.Errorf("HandleVoiceCommands() error = %v", err)
		}

		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Export file was not created")
		}
	})

	t.Run("invalid provider", func(t *testing.T) {
		cfg := config.Config{
			Provider:   "invalid",
			ListVoices: true,
		}

		err := HandleVoiceCommands(cfg, voiceCache, log)
		if err == nil {
			t.Error("Expected error for invalid provider")
		}
	})
}

func TestGetVoicesWithCacheInfo(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	voiceCache, err := cache.NewVoiceCache()
	if err != nil {
		t.Fatalf("Failed to create voice cache: %v", err)
	}
	defer func() {
		if err := voiceCache.Close(); err != nil {
			t.Logf("Warning: failed to close voice cache: %v", err)
		}
	}()

	cfg := config.Config{
		Provider: "say",
	}
	provider, err := CreateProvider(cfg)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	cachedProvider := cache.NewCachedProvider(provider, voiceCache)
	ctx := context.Background()
	log := logger.NewDefaultLogger()

	// Populate cache first
	_, err = cachedProvider.ListVoices(ctx)
	if err != nil {
		t.Fatalf("Failed to populate cache: %v", err)
	}

	cacheInfo, err := cachedProvider.GetCacheInfo(ctx)
	if err != nil {
		t.Fatalf("Failed to get cache info: %v", err)
	}

	t.Run("get voices from cache with info", func(t *testing.T) {
		voices, err := getVoices(ctx, cachedProvider, provider.Name(), false, cacheInfo, log)
		if err != nil {
			t.Errorf("getVoices() error = %v", err)
		}

		if len(voices) == 0 {
			t.Error("Expected voices but got none")
		}
	})

	t.Run("get voices with refresh", func(t *testing.T) {
		voices, err := getVoices(ctx, cachedProvider, provider.Name(), true, cacheInfo, log)
		if err != nil {
			t.Errorf("getVoices() error = %v", err)
		}

		if len(voices) == 0 {
			t.Error("Expected voices but got none")
		}
	})
}
