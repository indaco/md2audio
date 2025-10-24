package say

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
)

// Provider implements the TTS Provider interface for macOS 'say' command.
type Provider struct {
	// No configuration needed - 'say' is a system command
}

// NewProvider creates a new macOS say provider.
func NewProvider() (*Provider, error) {
	// Verify we're on macOS
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("say provider is only available on macOS")
	}

	// Verify say command exists
	if _, err := exec.LookPath("say"); err != nil {
		return nil, fmt.Errorf("say command not found: %w", err)
	}

	return &Provider{}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "say"
}

// Generate creates audio from text using the macOS say command.
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

	// Determine speaking rate
	rate := 180 // default
	if req.Rate != nil {
		rate = *req.Rate
	}

	// Build say command
	// Format: say -v Voice -r Rate -o output.aiff "text"
	outputPath := req.OutputPath
	// Ensure .aiff extension for say command
	if filepath.Ext(outputPath) != ".aiff" {
		outputPath = outputPath[:len(outputPath)-len(filepath.Ext(outputPath))] + ".aiff"
	}

	cmd := exec.CommandContext(ctx, "say", "-v", req.Voice, "-r", strconv.Itoa(rate), "-o", outputPath, cleanText)

	// Execute say command
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("say command failed: %w\nOutput: %s", err, string(output))
	}

	// Note: Using stderr for progress messages to avoid polluting stdout
	// TODO: Consider passing logger via context or provider interface for better integration
	fmt.Fprintf(os.Stderr, "Generating: %s\n", outputPath)

	// Measure actual duration using afinfo
	duration, err := getAudioDuration(outputPath)
	if err == nil {
		fmt.Fprintf(os.Stderr, "✓ Created: %s\n", outputPath)
		fmt.Fprintf(os.Stderr, "  Actual duration: %.2fs\n", duration)
	} else {
		fmt.Fprintf(os.Stderr, "✓ Created: %s\n", outputPath)
		fmt.Fprintf(os.Stderr, "  Warning: Could not measure duration: %v\n", err)
	}

	// Convert to M4A if requested
	if req.Format == "m4a" || req.Format == "mp4" {
		m4aPath := strings.Replace(outputPath, ".aiff", ".m4a", 1)
		if err := convertToM4A(ctx, outputPath, m4aPath); err != nil {
			return outputPath, fmt.Errorf("audio created but conversion to m4a failed: %w", err)
		}

		// Remove the original aiff file
		if err := os.Remove(outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not remove temporary aiff file: %v\n", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Converted to: %s\n", m4aPath)
		return m4aPath, nil
	}

	return outputPath, nil
}

// ListVoices returns available voices from the macOS say command.
func (p *Provider) ListVoices(ctx context.Context) ([]tts.Voice, error) {
	cmd := exec.CommandContext(ctx, "say", "-v", "?")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list voices: %w", err)
	}

	// Parse output (format: "VoiceName    locale  # Description")
	lines := strings.Split(string(output), "\n")
	voices := make([]tts.Voice, 0, len(lines))

	// Regex to parse voice lines
	voicePattern := regexp.MustCompile(`^([^\s]+(?:\s+\([^)]+\))?)\s+([a-z]{2}_[A-Z]{2})\s+#\s+(.+)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := voicePattern.FindStringSubmatch(line)
		if len(matches) == 4 {
			voices = append(voices, tts.Voice{
				ID:          matches[1], // Voice name is the ID
				Name:        matches[1],
				Language:    matches[2],
				Description: matches[3],
			})
		}
	}

	return voices, nil
}

// getAudioDuration measures the duration of an audio file using afinfo.
func getAudioDuration(audioPath string) (float64, error) {
	cmd := exec.Command("afinfo", audioPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("afinfo command failed: %w", err)
	}

	// Parse duration from afinfo output
	// Looking for line like: "estimated duration: 5.123456 sec"
	re := regexp.MustCompile(`estimated duration:\s+([\d.]+)\s+sec`)
	matches := re.FindStringSubmatch(string(output))

	if len(matches) < 2 {
		return 0, fmt.Errorf("could not parse duration from afinfo output")
	}

	duration, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse duration value: %w", err)
	}

	return duration, nil
}

// convertToM4A converts an AIFF file to M4A format using afconvert.
func convertToM4A(ctx context.Context, aiffPath, m4aPath string) error {
	cmd := exec.CommandContext(ctx, "afconvert", "-f", "mp4f", "-d", "aac", aiffPath, m4aPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("afconvert failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}
