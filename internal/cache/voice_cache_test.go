package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/indaco/md2audio/internal/logger"
	"github.com/indaco/md2audio/internal/tts"
)

func TestVoiceCache(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_cache.db")

	cache, err := NewVoiceCacheWithPath(dbPath, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer func() { _ = cache.Close() }()

	ctx := context.Background()

	// Test data
	testVoices := []tts.Voice{
		{
			ID:          "voice1",
			Name:        "Test Voice 1",
			Description: "A test voice",
			Language:    "en-US",
			Gender:      "female",
		},
		{
			ID:          "voice2",
			Name:        "Test Voice 2",
			Description: "Another test voice",
			Language:    "en-GB",
			Gender:      "male",
		},
	}

	// Test Set
	t.Run("Set", func(t *testing.T) {
		err := cache.Set(ctx, "test-provider", testVoices)
		if err != nil {
			t.Errorf("Failed to set cache: %v", err)
		}
	})

	// Test Get
	t.Run("Get", func(t *testing.T) {
		voices, err := cache.Get(ctx, "test-provider")
		if err != nil {
			t.Errorf("Failed to get cache: %v", err)
		}

		if len(voices) != len(testVoices) {
			t.Errorf("Expected %d voices, got %d", len(testVoices), len(voices))
		}

		if voices[0].ID != testVoices[0].ID {
			t.Errorf("Expected voice ID %s, got %s", testVoices[0].ID, voices[0].ID)
		}
	})

	// Test GetCacheInfo
	t.Run("GetCacheInfo", func(t *testing.T) {
		info, err := cache.GetCacheInfo(ctx, "test-provider")
		if err != nil {
			t.Errorf("Failed to get cache info: %v", err)
		}

		if info.Count != len(testVoices) {
			t.Errorf("Expected count %d, got %d", len(testVoices), info.Count)
		}

		if info.Provider != "test-provider" {
			t.Errorf("Expected provider %s, got %s", "test-provider", info.Provider)
		}
	})

	// Test Clear
	t.Run("Clear", func(t *testing.T) {
		err := cache.Clear(ctx, "test-provider")
		if err != nil {
			t.Errorf("Failed to clear cache: %v", err)
		}

		voices, err := cache.Get(ctx, "test-provider")
		if err != nil {
			t.Errorf("Failed to get cache after clear: %v", err)
		}

		if voices != nil {
			t.Errorf("Expected nil voices after clear, got %d voices", len(voices))
		}
	})
}

func TestCacheInfo(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_info.db")

	cache, err := NewVoiceCacheWithPath(dbPath, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer func() { _ = cache.Close() }()

	ctx := context.Background()

	testVoices := []tts.Voice{
		{ID: "v1", Name: "Voice 1", Language: "en-US"},
		{ID: "v2", Name: "Voice 2", Language: "en-GB"},
	}

	// Set cache
	err = cache.Set(ctx, "test-provider", testVoices)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Get cache info
	info, err := cache.GetCacheInfo(ctx, "test-provider")
	if err != nil {
		t.Fatalf("Failed to get cache info: %v", err)
	}

	if info.Count != 2 {
		t.Errorf("Expected count 2, got %d", info.Count)
	}

	if info.Provider != "test-provider" {
		t.Errorf("Expected provider 'test-provider', got %s", info.Provider)
	}

	if info.IsExpired(1 * time.Hour) {
		t.Error("Cache should not be expired")
	}
}

func TestExportToJSON(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_export.db")

	cache, err := NewVoiceCacheWithPath(dbPath, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer func() { _ = cache.Close() }()

	ctx := context.Background()

	testVoices := []tts.Voice{
		{ID: "v1", Name: "Voice 1", Description: "Test", Language: "en-US", Gender: "female"},
	}

	err = cache.Set(ctx, "test-provider", testVoices)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "voices.json")
	err = cache.ExportToJSON(ctx, "test-provider", outputPath)
	if err != nil {
		t.Fatalf("Failed to export JSON: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Expected JSON file to exist")
	}
}

// TestNewVoiceCache tests the default constructor
func TestNewVoiceCache(t *testing.T) {
	// NewVoiceCache uses the default path, so we need to handle that
	// We'll just verify it can be created
	cache, err := NewVoiceCache()
	if err != nil {
		t.Fatalf("Failed to create cache with default path: %v", err)
	}
	defer func() { _ = cache.Close() }()

	if cache == nil {
		t.Fatal("Expected cache to be created, got nil")
	}
}

// TestSetLogger tests the SetLogger method
func TestSetLogger(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_logger.db")

	cache, err := NewVoiceCacheWithPath(dbPath, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer func() { _ = cache.Close() }()

	// Use the real logger for testing
	log := logger.NewDefaultLogger()

	// This is just a test that SetLogger doesn't panic
	cache.SetLogger(log)

	// Verify that operations still work after setting logger
	ctx := context.Background()
	testVoices := []tts.Voice{
		{ID: "v1", Name: "Voice 1"},
	}

	err = cache.Set(ctx, "test-provider", testVoices)
	if err != nil {
		t.Errorf("Set failed after SetLogger: %v", err)
	}

	voices, err := cache.Get(ctx, "test-provider")
	if err != nil {
		t.Errorf("Get failed after SetLogger: %v", err)
	}

	if len(voices) != 1 {
		t.Errorf("Expected 1 voice, got %d", len(voices))
	}
}

// TestClearAll tests the ClearAll method
func TestClearAll(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_clear_all.db")

	cache, err := NewVoiceCacheWithPath(dbPath, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer func() { _ = cache.Close() }()

	ctx := context.Background()

	// Add voices for multiple providers
	provider1Voices := []tts.Voice{{ID: "v1", Name: "Provider 1 Voice"}}
	provider2Voices := []tts.Voice{{ID: "v2", Name: "Provider 2 Voice"}}

	err = cache.Set(ctx, "provider1", provider1Voices)
	if err != nil {
		t.Fatalf("Failed to set provider1 cache: %v", err)
	}

	err = cache.Set(ctx, "provider2", provider2Voices)
	if err != nil {
		t.Fatalf("Failed to set provider2 cache: %v", err)
	}

	// Verify both are cached
	voices1, _ := cache.Get(ctx, "provider1")
	voices2, _ := cache.Get(ctx, "provider2")

	if len(voices1) != 1 || len(voices2) != 1 {
		t.Fatal("Expected both providers to have voices before clear")
	}

	// Clear all
	err = cache.ClearAll(ctx)
	if err != nil {
		t.Fatalf("ClearAll failed: %v", err)
	}

	// Verify both are cleared
	voices1, _ = cache.Get(ctx, "provider1")
	voices2, _ = cache.Get(ctx, "provider2")

	if voices1 != nil {
		t.Errorf("Expected provider1 to be cleared, got %d voices", len(voices1))
	}

	if voices2 != nil {
		t.Errorf("Expected provider2 to be cleared, got %d voices", len(voices2))
	}
}

// TestIsExpired tests the IsExpired method on CacheInfo
func TestIsExpired(t *testing.T) {
	tests := []struct {
		name         string
		newestEntry  time.Time
		ttl          time.Duration
		shouldExpire bool
	}{
		{
			name:         "not expired - cached 10 minutes ago with 1 hour TTL",
			newestEntry:  time.Now().Add(-10 * time.Minute),
			ttl:          1 * time.Hour,
			shouldExpire: false,
		},
		{
			name:         "expired - cached 2 hours ago with 1 hour TTL",
			newestEntry:  time.Now().Add(-2 * time.Hour),
			ttl:          1 * time.Hour,
			shouldExpire: true,
		},
		{
			name:         "just expired - cached exactly 1 hour ago with 1 hour TTL",
			newestEntry:  time.Now().Add(-1 * time.Hour),
			ttl:          1 * time.Hour,
			shouldExpire: true,
		},
		{
			name:         "not expired - cached 1 second ago with 1 minute TTL",
			newestEntry:  time.Now().Add(-1 * time.Second),
			ttl:          1 * time.Minute,
			shouldExpire: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &CacheInfo{
				Provider:    "test",
				Count:       5,
				NewestEntry: tt.newestEntry,
				OldestEntry: tt.newestEntry,
			}

			if info.IsExpired(tt.ttl) != tt.shouldExpire {
				t.Errorf("IsExpired() = %v, want %v", info.IsExpired(tt.ttl), tt.shouldExpire)
			}
		})
	}
}
