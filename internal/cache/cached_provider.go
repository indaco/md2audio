package cache

import (
	"context"
	"fmt"

	"github.com/indaco/md2audio/internal/tts"
)

// CachedProvider wraps a TTS provider with voice caching capabilities.
type CachedProvider struct {
	provider tts.Provider
	cache    *VoiceCache
}

// NewCachedProvider creates a new cached provider wrapper.
func NewCachedProvider(provider tts.Provider, cache *VoiceCache) *CachedProvider {
	return &CachedProvider{
		provider: provider,
		cache:    cache,
	}
}

// Generate delegates to the underlying provider.
func (p *CachedProvider) Generate(ctx context.Context, req tts.GenerateRequest) (string, error) {
	return p.provider.Generate(ctx, req)
}

// ListVoices returns cached voices if available, otherwise fetches from provider.
func (p *CachedProvider) ListVoices(ctx context.Context) ([]tts.Voice, error) {
	// Try to get from cache first
	cachedVoices, err := p.cache.Get(ctx, p.provider.Name())
	if err != nil {
		return nil, fmt.Errorf("cache error: %w", err)
	}

	// Return cached voices if available
	if len(cachedVoices) > 0 {
		return cachedVoices, nil
	}

	// Cache miss - fetch from provider
	voices, err := p.provider.ListVoices(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache for next time
	if err := p.cache.Set(ctx, p.provider.Name(), voices); err != nil {
		// Log warning but don't fail - we have the voices
		fmt.Printf("Warning: Failed to cache voices: %v\n", err)
	}

	return voices, nil
}

// ListVoicesRefresh forces a refresh of the voice cache.
func (p *CachedProvider) ListVoicesRefresh(ctx context.Context) ([]tts.Voice, error) {
	// Clear existing cache
	if err := p.cache.Clear(ctx, p.provider.Name()); err != nil {
		return nil, fmt.Errorf("failed to clear cache: %w", err)
	}

	// Fetch fresh voices
	voices, err := p.provider.ListVoices(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := p.cache.Set(ctx, p.provider.Name(), voices); err != nil {
		return nil, fmt.Errorf("failed to cache voices: %w", err)
	}

	return voices, nil
}

// Name returns the underlying provider's name.
func (p *CachedProvider) Name() string {
	return p.provider.Name()
}

// GetCacheInfo returns cache information for the provider.
func (p *CachedProvider) GetCacheInfo(ctx context.Context) (*CacheInfo, error) {
	return p.cache.GetCacheInfo(ctx, p.provider.Name())
}

// ExportVoicesToJSON exports cached voices to a JSON file.
func (p *CachedProvider) ExportVoicesToJSON(ctx context.Context, outputPath string) error {
	return p.cache.ExportToJSON(ctx, p.provider.Name(), outputPath)
}
