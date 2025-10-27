// Package cli provides command-line interface utilities.
// It handles voice-related commands (list, export) and provider creation.
//
// Key features:
//   - Voice listing with caching support
//   - Voice export to JSON
//   - Provider factory pattern
//   - Cache management
//   - Formatted voice output
package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/indaco/md2audio/internal/cache"
	"github.com/indaco/md2audio/internal/config"
	"github.com/indaco/md2audio/internal/logger"
	"github.com/indaco/md2audio/internal/tts"
	"github.com/indaco/md2audio/internal/tts/elevenlabs"
	"github.com/indaco/md2audio/internal/tts/espeak"
	"github.com/indaco/md2audio/internal/tts/google"
	"github.com/indaco/md2audio/internal/tts/say"
	"github.com/indaco/md2audio/internal/utils"
)

// HandleVoiceCommands handles all voice-related commands (list, export).
func HandleVoiceCommands(cfg config.Config, voiceCache *cache.VoiceCache, log logger.LoggerInterface) error {
	provider, err := CreateProvider(cfg)
	if err != nil {
		return err
	}

	// Set logger on provider if it supports it
	if elevenlabsClient, ok := provider.(*elevenlabs.Client); ok {
		elevenlabsClient.SetLogger(log)
	}
	if googleClient, ok := provider.(*google.Client); ok {
		googleClient.SetLogger(log)
	}

	cachedProvider := cache.NewCachedProvider(provider, voiceCache)
	ctx := context.Background()

	// Ensure Google client is closed when done
	if googleClient, ok := provider.(*google.Client); ok {
		defer func() { _ = googleClient.Close() }()
	}

	if cfg.Commands.ExportVoices != "" {
		return ExportVoices(ctx, cachedProvider, provider.Name(), cfg.Commands.ExportVoices, log)
	}

	if cfg.Commands.ListVoices {
		return ListVoices(ctx, cachedProvider, provider.Name(), cfg.Commands.RefreshCache, log)
	}

	return nil
}

// CreateProvider creates a TTS provider based on configuration.
func CreateProvider(cfg config.Config) (tts.Provider, error) {
	ctx := context.Background()

	// Handle empty provider (use platform default)
	provider := cfg.Provider
	if provider == "" {
		provider = config.GetDefaultProvider()
	}

	switch provider {
	case "say":
		return say.NewProvider()
	case "espeak":
		return espeak.NewProvider()
	case "elevenlabs":
		return elevenlabs.NewClient(elevenlabs.Config{
			APIKey:          cfg.ElevenLabs.APIKey,
			Stability:       cfg.ElevenLabs.VoiceSettings.Stability,
			SimilarityBoost: cfg.ElevenLabs.VoiceSettings.SimilarityBoost,
			Style:           cfg.ElevenLabs.VoiceSettings.Style,
			UseSpeakerBoost: cfg.ElevenLabs.VoiceSettings.UseSpeakerBoost,
			Speed:           cfg.ElevenLabs.VoiceSettings.Speed,
		})
	case "google":
		return google.NewClient(ctx, google.Config{
			CredentialsFile: cfg.Google.CredentialsFile,
			LanguageCode:    cfg.Google.LanguageCode,
			SpeakingRate:    cfg.Google.SpeakingRate,
			Pitch:           cfg.Google.Pitch,
			VolumeGainDb:    cfg.Google.VolumeGainDb,
		})
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// ExportVoices exports cached voices to a JSON file.
func ExportVoices(ctx context.Context, cachedProvider *cache.CachedProvider, providerName, outputPath string, log logger.LoggerInterface) error {
	log.Info(fmt.Sprintf("Exporting cached voices for %s provider to %s...", providerName, outputPath))

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

	log.Success(fmt.Sprintf("Exported %d voices to %s", len(voices), outputPath))
	return nil
}

// ListVoices lists available voices, using cache or refreshing as needed.
func ListVoices(ctx context.Context, cachedProvider *cache.CachedProvider, providerName string, refreshCache bool, log logger.LoggerInterface) error {
	// Show cache info
	cacheInfo, err := cachedProvider.GetCacheInfo(ctx)
	if err == nil && cacheInfo.Count > 0 {
		log.Hint(fmt.Sprintf("Voice cache for %s provider: %d voices (cached %v ago)",
			providerName, cacheInfo.Count, utils.FormatDuration(cacheInfo.NewestEntry)))
		log.Blank()
	}

	// Get voices (refresh or use cache)
	voices, err := getVoices(ctx, cachedProvider, providerName, refreshCache, cacheInfo, log)
	if err != nil {
		return err
	}

	// Display voices
	displayVoices(providerName, voices, log)
	return nil
}

// getVoices retrieves voices either from cache or by refreshing.
func getVoices(ctx context.Context, cachedProvider *cache.CachedProvider, providerName string, refreshCache bool, cacheInfo *cache.CacheInfo, log logger.LoggerInterface) ([]tts.Voice, error) {
	if refreshCache {
		log.Info(fmt.Sprintf("Refreshing voice cache for %s provider...", providerName))
		voices, err := cachedProvider.ListVoicesRefresh(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh voices: %w", err)
		}
		log.Success(fmt.Sprintf("Cache refreshed with %d voices", len(voices)))
		log.Blank()
		return voices, nil
	}

	voices, err := cachedProvider.ListVoices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list voices: %w", err)
	}

	// Indicate cache usage
	if cacheInfo != nil && cacheInfo.Count > 0 {
		log.Hint("(using cached voices - use -refresh-cache to update)")
		log.Blank()
	}

	return voices, nil
}

// displayVoices displays the voice list in an appropriate format.
func displayVoices(providerName string, voices []tts.Voice, log logger.LoggerInterface) {
	log.Info(fmt.Sprintf("Available voices for %s provider:", providerName))
	log.Blank()

	if providerName == "elevenlabs" {
		displayElevenLabsVoices(voices, log)
	} else {
		displaySimpleVoices(voices, log)
	}
}

// displayElevenLabsVoices displays voices in ElevenLabs format with IDs.
func displayElevenLabsVoices(voices []tts.Voice, log logger.LoggerInterface) {
	log.Default(fmt.Sprintf("%-40s %-20s %-10s %s", "ID", "Name", "Language", "Description"))
	log.Default(strings.Repeat("-", 100))
	for _, voice := range voices {
		desc := voice.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		log.Default(fmt.Sprintf("%-40s %-20s %-10s %s", voice.ID, voice.Name, voice.Language, desc))
	}
}

// displaySimpleVoices displays voices in simple format (for say provider).
func displaySimpleVoices(voices []tts.Voice, log logger.LoggerInterface) {
	for _, voice := range voices {
		line := fmt.Sprintf("%-20s %-10s", voice.Name, voice.Language)
		if voice.Description != "" {
			line += fmt.Sprintf(" - %s", voice.Description)
		}
		log.Default(line)
	}
}
