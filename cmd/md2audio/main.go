package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/indaco/md2audio/internal/cache"
	"github.com/indaco/md2audio/internal/config"
	"github.com/indaco/md2audio/internal/processor"
	"github.com/indaco/md2audio/internal/tts"
	"github.com/indaco/md2audio/internal/tts/elevenlabs"
	"github.com/indaco/md2audio/internal/tts/say"
)

// run executes the main application logic and returns an error if something fails.
func run(cfg config.Config) error {
	// Initialize voice cache
	voiceCache, err := cache.NewVoiceCache()
	if err != nil {
		return fmt.Errorf("failed to initialize voice cache: %w", err)
	}
	defer func() {
		if closeErr := voiceCache.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close voice cache: %v\n", closeErr)
		}
	}()

	// Handle voice-related commands
	if cfg.ListVoices || cfg.ExportVoices != "" {
		return handleVoiceCommands(cfg, voiceCache)
	}

	// Validate configuration for audio processing
	if err := cfg.Validate(); err != nil {
		return err
	}

	cfg.Print()

	// Process based on mode
	if cfg.IsDirectoryMode() {
		return processor.ProcessDirectory(cfg)
	}
	return processor.ProcessFile(cfg.MarkdownFile, cfg.OutputDir, cfg)
}

// handleVoiceCommands handles all voice-related commands (list, export).
func handleVoiceCommands(cfg config.Config, voiceCache *cache.VoiceCache) error {
	provider, err := createProvider(cfg)
	if err != nil {
		return err
	}

	cachedProvider := cache.NewCachedProvider(provider, voiceCache)
	ctx := context.Background()

	if cfg.ExportVoices != "" {
		return exportVoices(ctx, cachedProvider, provider.Name(), cfg.ExportVoices)
	}

	if cfg.ListVoices {
		return listVoices(ctx, cachedProvider, provider.Name(), cfg.RefreshCache)
	}

	return nil
}

// createProvider creates a TTS provider based on configuration.
func createProvider(cfg config.Config) (tts.Provider, error) {
	switch cfg.Provider {
	case "say", "":
		return say.NewProvider()
	case "elevenlabs":
		return elevenlabs.NewClient(elevenlabs.Config{
			APIKey:          cfg.ElevenLabsAPIKey,
			Stability:       cfg.ElevenLabsStability,
			SimilarityBoost: cfg.ElevenLabsSimilarityBoost,
			Style:           cfg.ElevenLabsStyle,
			UseSpeakerBoost: cfg.ElevenLabsUseSpeakerBoost,
			Speed:           cfg.ElevenLabsSpeed,
		})
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}

// exportVoices exports cached voices to a JSON file.
func exportVoices(ctx context.Context, cachedProvider *cache.CachedProvider, providerName, outputPath string) error {
	fmt.Printf("Exporting cached voices for %s provider to %s...\n", providerName, outputPath)

	voices, err := cachedProvider.ListVoices(ctx)
	if err != nil {
		return fmt.Errorf("failed to get voices: %w", err)
	}

	if len(voices) == 0 {
		return fmt.Errorf("no voices available to export")
	}

	if err := cachedProvider.ExportVoicesToJSON(ctx, outputPath); err != nil {
		return fmt.Errorf("failed to export voices: %w", err)
	}

	fmt.Printf("✓ Exported %d voices to %s\n", len(voices), outputPath)
	return nil
}

// listVoices lists available voices, using cache or refreshing as needed.
func listVoices(ctx context.Context, cachedProvider *cache.CachedProvider, providerName string, refreshCache bool) error {
	// Show cache info
	cacheInfo, err := cachedProvider.GetCacheInfo(ctx)
	if err == nil && cacheInfo.Count > 0 {
		fmt.Printf("Voice cache for %s provider: %d voices (cached %v ago)\n\n",
			providerName, cacheInfo.Count, formatDuration(cacheInfo.NewestEntry))
	}

	// Get voices (refresh or use cache)
	voices, err := getVoices(ctx, cachedProvider, providerName, refreshCache, cacheInfo)
	if err != nil {
		return err
	}

	// Display voices
	displayVoices(providerName, voices)
	return nil
}

// getVoices retrieves voices either from cache or by refreshing.
func getVoices(ctx context.Context, cachedProvider *cache.CachedProvider, providerName string, refreshCache bool, cacheInfo *cache.CacheInfo) ([]tts.Voice, error) {
	if refreshCache {
		fmt.Printf("Refreshing voice cache for %s provider...\n", providerName)
		voices, err := cachedProvider.ListVoicesRefresh(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh voices: %w", err)
		}
		fmt.Printf("✓ Cache refreshed with %d voices\n\n", len(voices))
		return voices, nil
	}

	voices, err := cachedProvider.ListVoices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list voices: %w", err)
	}

	// Indicate cache usage
	if cacheInfo != nil && cacheInfo.Count > 0 {
		fmt.Printf("(using cached voices - use -refresh-cache to update)\n\n")
	}

	return voices, nil
}

// displayVoices displays the voice list in an appropriate format.
func displayVoices(providerName string, voices []tts.Voice) {
	fmt.Printf("Available voices for %s provider:\n\n", providerName)

	if providerName == "elevenlabs" {
		displayElevenLabsVoices(voices)
	} else {
		displaySimpleVoices(voices)
	}
}

// displayElevenLabsVoices displays voices in ElevenLabs format with IDs.
func displayElevenLabsVoices(voices []tts.Voice) {
	fmt.Printf("%-40s %-20s %-10s %s\n", "ID", "Name", "Language", "Description")
	fmt.Println("--------------------------------------------------------------------------------------------")
	for _, voice := range voices {
		desc := voice.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		fmt.Printf("%-40s %-20s %-10s %s\n", voice.ID, voice.Name, voice.Language, desc)
	}
}

// displaySimpleVoices displays voices in simple format (for say provider).
func displaySimpleVoices(voices []tts.Voice) {
	for _, voice := range voices {
		fmt.Printf("%-20s %-10s", voice.Name, voice.Language)
		if voice.Description != "" {
			fmt.Printf(" - %s", voice.Description)
		}
		fmt.Println()
	}
}

func main() {
	cfg := config.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// formatDuration formats a time.Time as a human-readable duration from now.
func formatDuration(t time.Time) string {
	d := time.Since(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}
