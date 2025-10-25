package espeak

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/indaco/md2audio/internal/text"
	"github.com/indaco/md2audio/internal/tts"
	"github.com/indaco/md2audio/internal/utils"
)

// Provider implements the TTS Provider interface for espeak-ng command.
type Provider struct {
	// No configuration needed - 'espeak-ng' is a system command
}

// NewProvider creates a new espeak-ng provider.
func NewProvider() (*Provider, error) {
	// Verify we're on Linux
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("espeak provider is only available on Linux")
	}

	// Try espeak-ng first, fall back to espeak
	cmd := "espeak-ng"
	if _, err := exec.LookPath(cmd); err != nil {
		cmd = "espeak"
		if _, err := exec.LookPath(cmd); err != nil {
			return nil, fmt.Errorf("neither espeak-ng nor espeak command found. Install with: sudo apt install espeak-ng")
		}
	}

	return &Provider{}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "espeak"
}

// Generate creates audio from text using the espeak-ng command.
func (p *Provider) Generate(ctx context.Context, req tts.GenerateRequest) (string, error) {
	// Clean markdown from text
	cleanText := text.CleanMarkdown(req.Text)
	if strings.TrimSpace(cleanText) == "" {
		return "", fmt.Errorf("no text to generate audio from")
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(req.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine speaking rate (espeak uses -s for speed)
	rate := 180 // default
	if req.Rate != nil {
		rate = *req.Rate
	}

	// Map macOS voice names to espeak voices
	voice := mapVoiceToEspeak(req.Voice)

	// Build espeak command
	// Format: espeak-ng -v voice -s rate -w output.wav "text"
	outputPath := req.OutputPath
	wavPath := outputPath

	// Ensure .wav extension for espeak command
	if filepath.Ext(outputPath) != ".wav" {
		wavPath = outputPath[:len(outputPath)-len(filepath.Ext(outputPath))] + ".wav"
	}

	// Try espeak-ng first, fall back to espeak
	cmdName := "espeak-ng"
	if _, err := exec.LookPath(cmdName); err != nil {
		cmdName = "espeak"
	}

	cmd := exec.CommandContext(ctx, cmdName, "-v", voice, "-s", strconv.Itoa(rate), "-w", wavPath, cleanText)

	// Execute espeak command
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("espeak command failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Fprintf(os.Stderr, "Generating: %s\n", wavPath)

	// Measure actual duration
	duration, err := utils.GetAudioDuration(wavPath)
	if err == nil {
		fmt.Fprintf(os.Stderr, "✓ Created: %s\n", wavPath)
		fmt.Fprintf(os.Stderr, "  Actual duration: %.2fs\n", duration)
	} else {
		fmt.Fprintf(os.Stderr, "✓ Created: %s\n", wavPath)
		fmt.Fprintf(os.Stderr, "  Warning: Could not measure duration: %v\n", err)
	}

	// Convert to other formats if requested
	if req.Format != "wav" && req.Format != "" {
		convertedPath := strings.Replace(wavPath, ".wav", "."+req.Format, 1)
		if err := convertAudio(ctx, wavPath, convertedPath, req.Format); err != nil {
			return wavPath, fmt.Errorf("audio created but conversion to %s failed: %w", req.Format, err)
		}

		// Remove the original wav file
		if err := os.Remove(wavPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not remove temporary wav file: %v\n", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Converted to: %s\n", convertedPath)
		return convertedPath, nil
	}

	return wavPath, nil
}

// ListVoices returns available voices from the espeak-ng command.
func (p *Provider) ListVoices(ctx context.Context) ([]tts.Voice, error) {
	// Try espeak-ng first, fall back to espeak
	cmdName := "espeak-ng"
	if _, err := exec.LookPath(cmdName); err != nil {
		cmdName = "espeak"
	}

	cmd := exec.CommandContext(ctx, cmdName, "--voices")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list voices: %w", err)
	}

	// Parse output (format: "Pty Language Age/Gender VoiceName File Other Languages")
	lines := strings.Split(string(output), "\n")
	voices := make([]tts.Voice, 0)

	// Skip header line
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// Extract voice info
		// Format: Pty Language Age/Gender VoiceName [File] [Other Languages]
		language := fields[1]
		voiceName := fields[3]

		// Parse gender from Age/Gender field
		ageGender := fields[2]
		gender := ""
		if strings.Contains(ageGender, "M") {
			gender = "male"
		} else if strings.Contains(ageGender, "F") {
			gender = "female"
		}

		// Create description
		description := fmt.Sprintf("%s voice", language)
		if gender != "" {
			// Capitalize first letter of gender
			capitalizedGender := strings.ToUpper(string(gender[0])) + gender[1:]
			description = fmt.Sprintf("%s %s voice", capitalizedGender, language)
		}

		voices = append(voices, tts.Voice{
			ID:          voiceName,
			Name:        voiceName,
			Language:    language,
			Description: description,
			Gender:      gender,
		})
	}

	return voices, nil
}

// mapVoiceToEspeak maps macOS voice names to espeak voice identifiers.
// This allows users to use the same voice names across platforms.
func mapVoiceToEspeak(voice string) string {
	// Common voice mappings
	voiceMap := map[string]string{
		// British voices
		"Kate":   "en-gb",
		"Daniel": "en-gb",
		"Oliver": "en-gb",
		"Serena": "en-gb",

		// US voices
		"Samantha": "en-us",
		"Alex":     "en-us",
		"Tom":      "en-us",
		"Fiona":    "en-us",

		// Australian voices
		"Karen": "en-au",

		// Indian voices
		"Veena": "en-in",

		// Other common languages
		"Thomas": "fr",    // French
		"Anna":   "de",    // German
		"Monica": "es",    // Spanish
		"Alice":  "it",    // Italian
		"Joana":  "pt-pt", // Portuguese
	}

	// Check if we have a mapping
	if espeakVoice, ok := voiceMap[voice]; ok {
		return espeakVoice
	}

	// If voice looks like an espeak voice code (e.g., "en-us"), use it directly
	if matched, _ := regexp.MatchString(`^[a-z]{2}(-[a-z]{2})?$`, voice); matched {
		return voice
	}

	// Default to en-us
	return "en-us"
}

// convertAudio converts a WAV file to another format using ffmpeg.
func convertAudio(ctx context.Context, inputPath, outputPath, format string) error {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg is required for audio conversion but not found. Install with: sudo apt install ffmpeg")
	}

	// Build ffmpeg command
	// ffmpeg -i input.wav -codec:a libmp3lame output.mp3 (for mp3)
	// ffmpeg -i input.wav -codec:a aac output.m4a (for m4a)
	var codec string
	switch format {
	case "mp3":
		codec = "libmp3lame"
	case "m4a", "mp4":
		codec = "aac"
	case "aiff":
		codec = "pcm_s16be"
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", "-i", inputPath, "-codec:a", codec, "-y", outputPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
