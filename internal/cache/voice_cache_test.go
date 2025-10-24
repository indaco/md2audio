package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

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
