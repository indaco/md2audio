// Package audio provides audio generation orchestration.
// It coordinates TTS providers to generate audio files from markdown sections,
// with support for timing control and format conversion.
//
// Key features:
//   - Audio generation orchestration
//   - Timing annotation support
//   - Speaking rate calculation
//   - Multiple output formats (AIFF, M4A, MP3)
//   - Duration measurement and validation
package audio

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/indaco/md2audio/internal/logger"
	"github.com/indaco/md2audio/internal/parser"
	"github.com/indaco/md2audio/internal/text"
	"github.com/indaco/md2audio/internal/tts"
	"github.com/indaco/md2audio/internal/utils"
)

// GeneratorConfig holds configuration for audio generation
type GeneratorConfig struct {
	Voice     string
	Rate      int
	Format    string
	Prefix    string
	OutputDir string
	Provider  tts.Provider // TTS provider to use
}

// Generator handles audio file generation
type Generator struct {
	config GeneratorConfig
	log    logger.LoggerInterface
}

// NewGenerator creates a new audio generator
func NewGenerator(config GeneratorConfig, log logger.LoggerInterface) *Generator {
	return &Generator{
		config: config,
		log:    log,
	}
}

// ListAvailableVoices lists all available macOS voices
func ListAvailableVoices() error {
	fmt.Println("Available voices:")
	cmd := exec.Command("say", "-v", "?")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error listing voices: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

// Generate generates an audio file for a section
func (g *Generator) Generate(section parser.Section, index int) error {
	if g.config.Provider == nil {
		return fmt.Errorf("no TTS provider configured")
	}

	safeTitle := text.SanitizeFilename(section.Title)

	// Build output path based on format
	var outputPath string
	fileExt := g.config.Format

	// For say provider with m4a, we need to use .aiff initially
	// For elevenlabs, use the format directly (it outputs mp3)
	if g.config.Provider.Name() == "say" {
		if g.config.Format == "m4a" {
			fileExt = "aiff" // say provider will convert after generation
		}
	} else if g.config.Provider.Name() == "elevenlabs" {
		fileExt = "mp3" // ElevenLabs outputs MP3
	}

	outputPath = filepath.Join(g.config.OutputDir, fmt.Sprintf("%s_%02d_%s.%s", g.config.Prefix, index, safeTitle, fileExt))

	// Determine speaking rate (only used by say provider)
	speakingRate := g.config.Rate
	var targetDuration *float64
	if section.HasTiming {
		// Calculate required rate to fit the duration (for say provider)
		estimatedRate := estimateSpeakingRate(section.Content, section.Duration, g.log)
		speakingRate = estimatedRate
		g.log.Hint(fmt.Sprintf("Target duration: %.1fs, Calculated rate: %d wpm", section.Duration, speakingRate))

		// Also pass target duration for providers that support it (e.g., ElevenLabs)
		targetDuration = &section.Duration
	}

	// Build TTS request
	request := tts.GenerateRequest{
		Text:           section.Content,
		Voice:          g.config.Voice,
		OutputPath:     outputPath,
		Rate:           &speakingRate,
		Format:         g.config.Format,
		TargetDuration: targetDuration,
	}

	// Generate audio using TTS provider
	ctx := context.Background()
	finalPath, err := g.config.Provider.Generate(ctx, request)
	if err != nil {
		return fmt.Errorf("error generating audio: %w", err)
	}

	// Show timing info if applicable
	if section.HasTiming {
		// Try to get actual duration (provider-dependent)
		if g.config.Provider.Name() == "say" {
			if actualDuration, err := utils.GetAudioDuration(finalPath); err == nil {
				diff := actualDuration - section.Duration
				g.log.WithIndent(true)
				g.log.Hint(fmt.Sprintf("target: %.1fs, diff: %+.2fs", section.Duration, diff))
				g.log.WithIndent(false)
			}
		}
	}

	return nil
}

// estimateSpeakingRate calculates the words per minute needed to fit target duration
func estimateSpeakingRate(textContent string, targetDuration float64, log logger.LoggerInterface) int {
	const (
		minWPM           = 90
		maxWPM           = 360
		defaultWPM       = 180
		adjustmentFactor = 0.95 // Empirical adjustment - say command seems slightly faster
	)

	wordCount := utils.CountWords(textContent)
	requiredWPM := utils.CalculateWPM(wordCount, targetDuration)

	if requiredWPM <= 0 {
		return defaultWPM // default fallback
	}

	// Add a small adjustment factor (say command seems to be slightly faster in practice)
	adjustedWPM := requiredWPM * adjustmentFactor

	// Clamp to reasonable values (say supports roughly 90-360 wpm)
	if adjustedWPM > maxWPM {
		log.Warning(fmt.Sprintf("Required rate (%.0f wpm) exceeds maximum, capping at %d wpm", requiredWPM, maxWPM))
	}

	return utils.ClampInt(int(adjustedWPM), minWPM, maxWPM)
}
