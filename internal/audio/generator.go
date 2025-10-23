package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/indaco/md2audio/internal/parser"
	"github.com/indaco/md2audio/internal/text"
)

// GeneratorConfig holds configuration for audio generation
type GeneratorConfig struct {
	Voice     string
	Rate      int
	Format    string
	Prefix    string
	OutputDir string
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
	safeTitle := text.SanitizeFilename(section.Title)

	var outputPath string
	if g.config.Format == "m4a" {
		// Start with aiff, will convert later
		outputPath = filepath.Join(g.config.OutputDir, fmt.Sprintf("%s_%02d_%s.aiff", g.config.Prefix, index, safeTitle))
	} else {
		outputPath = filepath.Join(g.config.OutputDir, fmt.Sprintf("%s_%02d_%s.%s", g.config.Prefix, index, safeTitle, g.config.Format))
	}

	// Determine speaking rate
	speakingRate := g.config.Rate

	if section.HasTiming {
		// Calculate required rate to fit the duration
		estimatedRate := estimateSpeakingRate(section.Content, section.Duration)
		speakingRate = estimatedRate
		fmt.Printf("Target duration: %.1fs, Calculated rate: %d wpm\n", section.Duration, speakingRate)
	}

	fmt.Printf("Generating: %s\n", outputPath)

	// Generate audio using say command
	cmd := exec.Command("say", "-v", g.config.Voice, "-r", fmt.Sprintf("%d", speakingRate), "-o", outputPath, section.Content)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error generating audio: %w\n%s", err, string(output))
	}

	fmt.Printf("✓ Created: %s\n", outputPath)

	// Measure actual duration
	actualDuration := getAudioDuration(outputPath)
	if actualDuration > 0 {
		fmt.Printf("  Actual duration: %.2fs", actualDuration)
		if section.HasTiming {
			diff := actualDuration - section.Duration
			fmt.Printf(" (target: %.1fs, diff: %+.2fs)\n", section.Duration, diff)
		} else {
			fmt.Printf("\n")
		}
	}

	// Convert to m4a if requested
	if g.config.Format == "m4a" {
		m4aPath := strings.Replace(outputPath, ".aiff", ".m4a", 1)
		convertCmd := exec.Command("afconvert", "-f", "mp4f", "-d", "aac", outputPath, m4aPath)

		if output, err := convertCmd.CombinedOutput(); err != nil {
			fmt.Printf("Warning: Could not convert to m4a: %v\n%s\n", err, string(output))
		} else {
			// Remove the original aiff file
			if err := os.Remove(outputPath); err != nil {
				fmt.Printf("Warning: Could not remove temporary aiff file: %v\n", err)
			}
			fmt.Printf("✓ Converted to: %s\n", m4aPath)
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
