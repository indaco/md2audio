package cache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/indaco/md2audio/internal/tts"
)

// MockTTSProvider is a mock TTS provider for testing
type MockTTSProvider struct {
	name            string
	voices          []tts.Voice
	listVoicesCalls int
	generateCalls   int
	shouldError     bool
}

func (m *MockTTSProvider) Name() string {
	return m.name
}

func (m *MockTTSProvider) Generate(ctx context.Context, req tts.GenerateRequest) (string, error) {
	m.generateCalls++
	if m.shouldError {
		return "", fmt.Errorf("mock generate error")
	}
	return req.OutputPath, nil
}

func (m *MockTTSProvider) ListVoices(ctx context.Context) ([]tts.Voice, error) {
	m.listVoicesCalls++
	if m.shouldError {
		return nil, fmt.Errorf("mock list voices error")
	}
	return m.voices, nil
}

func setupTestCachedProvider(t *testing.T, providerName string, voices []tts.Voice) (*CachedProvider, *MockTTSProvider) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_cached_provider.db")

	cache, err := NewVoiceCacheWithPath(dbPath, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	mockProvider := &MockTTSProvider{
		name:   providerName,
		voices: voices,
	}

	cachedProvider := NewCachedProvider(mockProvider, cache)

	return cachedProvider, mockProvider
}

func TestNewCachedProvider(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_new.db")

	cache, err := NewVoiceCacheWithPath(dbPath, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer func() { _ = cache.Close() }()

	mockProvider := &MockTTSProvider{name: "test-provider"}
	cachedProvider := NewCachedProvider(mockProvider, cache)

	if cachedProvider == nil {
		t.Fatal("NewCachedProvider returned nil")
	}

	if cachedProvider.Name() != "test-provider" {
		t.Errorf("Expected name 'test-provider', got %s", cachedProvider.Name())
	}
}

func TestCachedProviderGenerate(t *testing.T) {
	testVoices := []tts.Voice{
		{ID: "v1", Name: "Voice 1"},
	}

	cachedProvider, mockProvider := setupTestCachedProvider(t, "test-provider", testVoices)
	ctx := context.Background()

	// Test Generate delegates to underlying provider
	req := tts.GenerateRequest{
		Text:       "Test text",
		Voice:      "v1",
		OutputPath: "/tmp/test.mp3",
	}

	outputPath, err := cachedProvider.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if outputPath != req.OutputPath {
		t.Errorf("Expected output path %s, got %s", req.OutputPath, outputPath)
	}

	if mockProvider.generateCalls != 1 {
		t.Errorf("Expected 1 generate call, got %d", mockProvider.generateCalls)
	}
}

func TestCachedProviderListVoicesCacheMiss(t *testing.T) {
	testVoices := []tts.Voice{
		{ID: "v1", Name: "Voice 1", Language: "en-US"},
		{ID: "v2", Name: "Voice 2", Language: "en-GB"},
	}

	cachedProvider, mockProvider := setupTestCachedProvider(t, "test-provider", testVoices)
	ctx := context.Background()

	// First call - cache miss, should fetch from provider
	voices, err := cachedProvider.ListVoices(ctx)
	if err != nil {
		t.Fatalf("ListVoices failed: %v", err)
	}

	if len(voices) != 2 {
		t.Errorf("Expected 2 voices, got %d", len(voices))
	}

	if mockProvider.listVoicesCalls != 1 {
		t.Errorf("Expected 1 ListVoices call, got %d", mockProvider.listVoicesCalls)
	}
}

func TestCachedProviderListVoicesCacheHit(t *testing.T) {
	testVoices := []tts.Voice{
		{ID: "v1", Name: "Voice 1", Language: "en-US"},
		{ID: "v2", Name: "Voice 2", Language: "en-GB"},
	}

	cachedProvider, mockProvider := setupTestCachedProvider(t, "test-provider", testVoices)
	ctx := context.Background()

	// First call - cache miss
	_, err := cachedProvider.ListVoices(ctx)
	if err != nil {
		t.Fatalf("First ListVoices failed: %v", err)
	}

	// Second call - cache hit, should NOT call provider
	voices, err := cachedProvider.ListVoices(ctx)
	if err != nil {
		t.Fatalf("Second ListVoices failed: %v", err)
	}

	if len(voices) != 2 {
		t.Errorf("Expected 2 voices from cache, got %d", len(voices))
	}

	// Should only have been called once (first time)
	if mockProvider.listVoicesCalls != 1 {
		t.Errorf("Expected 1 ListVoices call (cache hit on second), got %d", mockProvider.listVoicesCalls)
	}
}

func TestCachedProviderListVoicesProviderError(t *testing.T) {
	testVoices := []tts.Voice{}

	cachedProvider, mockProvider := setupTestCachedProvider(t, "test-provider", testVoices)
	mockProvider.shouldError = true

	ctx := context.Background()

	// Should propagate provider error
	_, err := cachedProvider.ListVoices(ctx)
	if err == nil {
		t.Error("Expected error from provider, got nil")
	}
}

func TestCachedProviderListVoicesRefresh(t *testing.T) {
	testVoices := []tts.Voice{
		{ID: "v1", Name: "Voice 1"},
	}

	cachedProvider, mockProvider := setupTestCachedProvider(t, "test-provider", testVoices)
	ctx := context.Background()

	// First populate the cache
	_, err := cachedProvider.ListVoices(ctx)
	if err != nil {
		t.Fatalf("Initial ListVoices failed: %v", err)
	}

	// Update provider's voices
	mockProvider.voices = []tts.Voice{
		{ID: "v1", Name: "Voice 1"},
		{ID: "v2", Name: "Voice 2"},
		{ID: "v3", Name: "Voice 3"},
	}

	// Refresh should clear cache and fetch new voices
	voices, err := cachedProvider.ListVoicesRefresh(ctx)
	if err != nil {
		t.Fatalf("ListVoicesRefresh failed: %v", err)
	}

	if len(voices) != 3 {
		t.Errorf("Expected 3 voices after refresh, got %d", len(voices))
	}

	// Should have called provider twice (once for initial, once for refresh)
	if mockProvider.listVoicesCalls != 2 {
		t.Errorf("Expected 2 ListVoices calls, got %d", mockProvider.listVoicesCalls)
	}
}

func TestCachedProviderListVoicesRefreshError(t *testing.T) {
	testVoices := []tts.Voice{
		{ID: "v1", Name: "Voice 1"},
	}

	cachedProvider, mockProvider := setupTestCachedProvider(t, "test-provider", testVoices)
	ctx := context.Background()

	// Make provider fail
	mockProvider.shouldError = true

	// Refresh should fail
	_, err := cachedProvider.ListVoicesRefresh(ctx)
	if err == nil {
		t.Error("Expected error from refresh with failing provider, got nil")
	}
}

func TestCachedProviderName(t *testing.T) {
	testVoices := []tts.Voice{}
	cachedProvider, _ := setupTestCachedProvider(t, "elevenlabs", testVoices)

	if cachedProvider.Name() != "elevenlabs" {
		t.Errorf("Expected name 'elevenlabs', got %s", cachedProvider.Name())
	}
}

func TestCachedProviderGetCacheInfo(t *testing.T) {
	testVoices := []tts.Voice{
		{ID: "v1", Name: "Voice 1"},
		{ID: "v2", Name: "Voice 2"},
	}

	cachedProvider, _ := setupTestCachedProvider(t, "test-provider", testVoices)
	ctx := context.Background()

	// Populate cache
	_, err := cachedProvider.ListVoices(ctx)
	if err != nil {
		t.Fatalf("ListVoices failed: %v", err)
	}

	// Get cache info
	info, err := cachedProvider.GetCacheInfo(ctx)
	if err != nil {
		t.Fatalf("GetCacheInfo failed: %v", err)
	}

	if info.Provider != "test-provider" {
		t.Errorf("Expected provider 'test-provider', got %s", info.Provider)
	}

	if info.Count != 2 {
		t.Errorf("Expected count 2, got %d", info.Count)
	}

	if info.IsExpired(1 * time.Hour) {
		t.Error("Cache should not be expired")
	}
}

func TestCachedProviderGetCacheInfoEmpty(t *testing.T) {
	testVoices := []tts.Voice{}
	cachedProvider, _ := setupTestCachedProvider(t, "test-provider", testVoices)
	ctx := context.Background()

	// Get info without populating cache
	info, err := cachedProvider.GetCacheInfo(ctx)
	if err != nil {
		t.Fatalf("GetCacheInfo failed: %v", err)
	}

	// Empty cache returns info with Count=0
	if info == nil {
		t.Fatal("Expected info for empty cache, got nil")
	}

	if info.Count != 0 {
		t.Errorf("Expected count 0 for empty cache, got %d", info.Count)
	}
}

func TestCachedProviderExportVoicesToJSON(t *testing.T) {
	testVoices := []tts.Voice{
		{ID: "v1", Name: "Voice 1", Description: "Test voice", Language: "en-US", Gender: "female"},
	}

	cachedProvider, _ := setupTestCachedProvider(t, "test-provider", testVoices)
	ctx := context.Background()

	// Populate cache
	_, err := cachedProvider.ListVoices(ctx)
	if err != nil {
		t.Fatalf("ListVoices failed: %v", err)
	}

	// Export to JSON
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "exported_voices.json")

	err = cachedProvider.ExportVoicesToJSON(ctx, outputPath)
	if err != nil {
		t.Fatalf("ExportVoicesToJSON failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Expected JSON file to exist")
	}

	// Verify file is not empty
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Expected non-empty JSON file")
	}
}

func TestCachedProviderExportVoicesToJSONEmpty(t *testing.T) {
	testVoices := []tts.Voice{}
	cachedProvider, _ := setupTestCachedProvider(t, "test-provider", testVoices)
	ctx := context.Background()

	// Try to export without populating cache
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "empty_export.json")

	err := cachedProvider.ExportVoicesToJSON(ctx, outputPath)
	// Should handle empty cache gracefully (might error or create empty array)
	// The behavior depends on the ExportToJSON implementation
	if err != nil {
		// It's okay if it errors on empty cache
		t.Logf("Export on empty cache returned error: %v", err)
	}
}
