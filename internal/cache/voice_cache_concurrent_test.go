package cache

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/indaco/md2audio/internal/tts"
)

// setupTestCache creates a test cache with a temporary database
func setupTestCache(t *testing.T) *VoiceCache {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_cache.db")

	cache, err := NewVoiceCacheWithPath(dbPath, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create test cache: %v", err)
	}

	t.Cleanup(func() {
		_ = cache.Close()
	})

	return cache
}

func TestVoiceCacheConcurrentReads(t *testing.T) {
	cache := setupTestCache(t)
	ctx := context.Background()

	// Save some test data
	voices := []tts.Voice{
		{ID: "test-voice-1", Name: "Test Voice 1"},
		{ID: "test-voice-2", Name: "Test Voice 2"},
	}
	err := cache.Set(ctx, "test-provider", voices)
	if err != nil {
		t.Fatalf("Failed to save voices: %v", err)
	}

	var wg sync.WaitGroup
	numReaders := 20
	errors := make(chan error, numReaders)

	// Go 1.25.3: Use WaitGroup.Go() for safer goroutine management
	for range numReaders {
		wg.Go(func() {
			cachedVoices, err := cache.Get(ctx, "test-provider")
			if err != nil {
				errors <- fmt.Errorf("concurrent read failed: %w", err)
				return
			}
			if len(cachedVoices) != 2 {
				errors <- fmt.Errorf("expected 2 voices, got %d", len(cachedVoices))
			}
		})
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Error(err)
	}
}

func TestVoiceCacheConcurrentWrites(t *testing.T) {
	cache := setupTestCache(t)
	ctx := context.Background()

	var wg sync.WaitGroup
	numWriters := 10
	errors := make(chan error, numWriters)

	// Go 1.25.3: Use WaitGroup.Go()
	for i := range numWriters {
		// capture loop variable
		wg.Go(func() {
			voices := []tts.Voice{{
				ID:   fmt.Sprintf("voice-%d", i),
				Name: fmt.Sprintf("Voice %d", i),
			}}
			err := cache.Set(ctx, fmt.Sprintf("provider-%d", i), voices)
			if err != nil {
				errors <- fmt.Errorf("concurrent write failed for provider-%d: %w", i, err)
			}
		})
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Error(err)
	}

	// Verify all writes succeeded
	for i := range numWriters {
		voices, err := cache.Get(ctx, fmt.Sprintf("provider-%d", i))
		if err != nil {
			t.Errorf("Failed to get voices for provider-%d: %v", i, err)
		}
		if len(voices) != 1 {
			t.Errorf("Expected 1 voice for provider-%d, got %d", i, len(voices))
		}
	}
}

func TestVoiceCacheMixedOperations(t *testing.T) {
	cache := setupTestCache(t)
	ctx := context.Background()

	// Pre-populate with some data
	initialVoices := []tts.Voice{
		{ID: "initial-voice", Name: "Initial Voice"},
	}
	err := cache.Set(ctx, "shared-provider", initialVoices)
	if err != nil {
		t.Fatalf("Failed to save initial voices: %v", err)
	}

	var wg sync.WaitGroup
	numOperations := 20
	errors := make(chan error, numOperations)

	// Mix of concurrent reads and writes
	for i := range numOperations {
		if i%2 == 0 {
			// Even: Read operation
			wg.Go(func() {
				_, err := cache.Get(ctx, "shared-provider")
				if err != nil {
					errors <- fmt.Errorf("read operation %d failed: %w", i, err)
				}
			})
		} else {
			// Odd: Write operation
			wg.Go(func() {
				voices := []tts.Voice{{
					ID:   fmt.Sprintf("voice-%d", i),
					Name: fmt.Sprintf("Voice %d", i),
				}}
				err := cache.Set(ctx, fmt.Sprintf("provider-%d", i), voices)
				if err != nil {
					errors <- fmt.Errorf("write operation %d failed: %w", i, err)
				}
			})
		}
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	errorCount := 0
	for err := range errors {
		t.Error(err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("Mixed operations had %d errors", errorCount)
	}
}

func TestWALModeEnabled(t *testing.T) {
	cache := setupTestCache(t)

	// Query current journal mode
	var journalMode string
	err := cache.db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("Failed to query journal mode: %v", err)
	}

	if journalMode != "wal" {
		t.Errorf("Expected WAL mode, got: %s", journalMode)
	}
}

func TestCachePragmaSettings(t *testing.T) {
	cache := setupTestCache(t)

	tests := []struct {
		name     string
		pragma   string
		expected string
	}{
		{
			name:     "journal_mode",
			pragma:   "PRAGMA journal_mode",
			expected: "wal",
		},
		{
			name:     "synchronous",
			pragma:   "PRAGMA synchronous",
			expected: "1", // NORMAL = 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var value string
			err := cache.db.QueryRow(tt.pragma).Scan(&value)
			if err != nil {
				t.Fatalf("Failed to query %s: %v", tt.pragma, err)
			}
			if value != tt.expected {
				t.Errorf("%s: expected %s, got %s", tt.name, tt.expected, value)
			}
		})
	}
}

// TestConcurrentCacheStress is a stress test with many concurrent operations
func TestConcurrentCacheStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	cache := setupTestCache(t)
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 50
	operationsPerGoroutine := 10
	errors := make(chan error, numGoroutines*operationsPerGoroutine)

	for g := range numGoroutines {
		wg.Go(func() {
			for op := range operationsPerGoroutine {
				providerName := fmt.Sprintf("provider-%d", g)
				voices := []tts.Voice{{
					ID:   fmt.Sprintf("voice-%d-%d", g, op),
					Name: fmt.Sprintf("Voice %d-%d", g, op),
				}}

				// Write
				if err := cache.Set(ctx, providerName, voices); err != nil {
					errors <- fmt.Errorf("stress test write failed (g=%d, op=%d): %w", g, op, err)
					continue
				}

				// Read
				if _, err := cache.Get(ctx, providerName); err != nil {
					errors <- fmt.Errorf("stress test read failed (g=%d, op=%d): %w", g, op, err)
				}
			}
		})
	}

	wg.Wait()
	close(errors)

	errorCount := 0
	for err := range errors {
		if errorCount < 10 { // Only log first 10 errors
			t.Error(err)
		}
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("Stress test completed with %d errors", errorCount)
	}
}
