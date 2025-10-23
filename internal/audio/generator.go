package audio

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/indaco/md2audio/internal/parser"
	"github.com/indaco/md2audio/internal/text"
	"github.com/indaco/md2audio/internal/tts"
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
}

// NewGenerator creates a new audio generator
func NewGenerator(config GeneratorConfig) *Generator {
	return &Generator{config: config}
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
		estimatedRate := estimateSpeakingRate(section.Content, section.Duration)
		speakingRate = estimatedRate
		fmt.Printf("Target duration: %.1fs, Calculated rate: %d wpm\n", section.Duration, speakingRate)

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
			actualDuration := getAudioDuration(finalPath)
			if actualDuration > 0 {
				diff := actualDuration - section.Duration
				fmt.Printf("  (target: %.1fs, diff: %+.2fs)\n", section.Duration, diff)
			}
		}
	}

	return nil
}

// estimateSpeakingRate calculates the words per minute needed to fit target duration
func estimateSpeakingRate(textContent string, targetDuration float64) int {
	// Count words
	words := strings.Fields(textContent)
	wordCount := len(words)

	// Calculate required words per minute
	// targetDuration is in seconds, convert to minutes
	targetMinutes := targetDuration / 60.0

	if targetMinutes <= 0 {
		return 180 // default fallback
	}

	requiredWPM := float64(wordCount) / targetMinutes

	// Add a small adjustment factor (say command seems to be slightly faster in practice)
	// This is empirical and may need tuning
	adjustedWPM := requiredWPM * 0.95

	// Clamp to reasonable values (say supports roughly 90-360 wpm)
	if adjustedWPM < 90 {
		adjustedWPM = 90
	} else if adjustedWPM > 360 {
		adjustedWPM = 360
		fmt.Printf("Warning: Required rate (%.0f wpm) exceeds maximum, capping at 360 wpm\n", requiredWPM)
	}

	return int(adjustedWPM)
}

// getAudioDuration returns the duration of an audio file in seconds
func getAudioDuration(filepath string) float64 {
	cmd := exec.Command("afinfo", filepath)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	// Parse output for duration
	// Looking for line like: "estimated duration: 8.413764 sec"
	durationPattern := regexp.MustCompile(`estimated duration:\s+([\d.]+)\s+sec`)
	if match := durationPattern.FindStringSubmatch(string(output)); match != nil {
		var duration float64
		if _, err := fmt.Sscanf(match[1], "%f", &duration); err == nil {
			return duration
		}
	}

	return 0
}
