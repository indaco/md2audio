// Package config provides configuration management for the md2audio CLI.
// It handles command-line argument parsing, environment variable loading,
// voice presets, and configuration validation.
//
// Key features:
//   - CLI flag parsing with sensible defaults
//   - Voice preset management (british-female, us-male, etc.)
//   - Environment variable integration (.env file support)
//   - Provider-specific configuration (say, elevenlabs)
//   - Configuration validation
//   - Secure API key masking in output
package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/indaco/md2audio/internal/env"
	"github.com/indaco/md2audio/internal/logger"
)

// VoicePresets maps common voice configurations to voice names
var VoicePresets = map[string]string{
	"british-female":    "Kate",
	"british-male":      "Daniel",
	"us-female":         "Samantha",
	"us-male":           "Alex",
	"australian-female": "Karen",
	"indian-female":     "Veena",
}

// DefaultElevenLabsVoiceID is the default voice for ElevenLabs (Rachel)
const DefaultElevenLabsVoiceID = "21m00Tcm4TlvDq8ikWAM"

// Config holds the application configuration
type Config struct {
	// Input/Output Options
	MarkdownFile string // Path to input markdown file (mutually exclusive with InputDir)
	InputDir     string // Path to input directory for recursive processing (mutually exclusive with MarkdownFile)
	OutputDir    string // Path to output directory for generated audio files (default: "./audio_sections")

	// Say Provider Options (macOS only)
	Voice string // Voice name for say provider (default: "Kate")
	Rate  int    // Speaking rate in words per minute for say provider (default: 180)

	// Common Audio Options
	Format string // Output audio format: "aiff", "m4a", or "mp3" (default: "aiff")
	Prefix string // Prefix for output filenames (default: "section")

	// Command Options
	ListVoices   bool   // List all available voices for the selected provider
	RefreshCache bool   // Force refresh voice cache when listing voices
	ExportVoices string // Export cached voices to JSON file (e.g., "voices.json")
	Version      bool   // Print version and exit
	Debug        bool   // Enable debug logging
	DryRun       bool   // Dry-run mode: show what would be generated without creating files

	// TTS Provider Configuration
	Provider          string // TTS provider: "say" (macOS) or "elevenlabs" (default: "say")
	ElevenLabsVoiceID string // ElevenLabs voice ID (required when using elevenlabs provider)
	ElevenLabsModel   string // ElevenLabs model ID (default: "eleven_monolingual_v1")
	ElevenLabsAPIKey  string // ElevenLabs API key (prefer ELEVENLABS_API_KEY env var)

	// ElevenLabs Voice Settings (optional, loaded from environment variables with defaults)
	ElevenLabsStability       float64 // Voice consistency (0.0-1.0, default: 0.5, higher = more consistent but less expressive)
	ElevenLabsSimilarityBoost float64 // Voice similarity to original (0.0-1.0, default: 0.5, higher = closer to voice characteristics)
	ElevenLabsStyle           float64 // Voice style/emotional range (0.0-1.0, default: 0.0 = disabled, higher = more expressive)
	ElevenLabsUseSpeakerBoost bool    // Boost similarity of synthesized speech (default: true)
	ElevenLabsSpeed           float64 // Speaking speed multiplier (0.7-1.2, default: 1.0, only for non-timed sections)
}

// Parse parses command-line flags and returns the configuration
func Parse() Config {
	// Load .env file if it exists (won't override existing env vars)
	if _, err := env.Load(".env"); err != nil {
		// Only warn if there's an actual error (not just file not found)
		fmt.Fprintf(os.Stderr, "Warning: Failed to load .env file: %v\n", err)
	}

	// Create logger for help message
	log := logger.NewDefaultLogger()

	config := Config{}

	flag.StringVar(&config.MarkdownFile, "f", "", "Input markdown file (use -f or -d, not both)")
	flag.StringVar(&config.InputDir, "d", "", "Input directory to process recursively (use -f or -d, not both)")
	flag.StringVar(&config.OutputDir, "o", "./audio_sections", "Output directory for audio files")

	// TTS Provider
	flag.StringVar(&config.Provider, "provider", "say", "TTS provider: 'say' (macOS) or 'elevenlabs'")

	// Say provider options
	var preset string
	flag.StringVar(&preset, "p", "", "Voice preset for say provider (british-female, british-male, us-female, us-male, australian-female, indian-female)")
	flag.StringVar(&config.Voice, "v", "", "Specific voice name for say provider (overrides preset)")
	flag.IntVar(&config.Rate, "r", 180, "Speaking rate for say provider (lower = slower)")

	// ElevenLabs provider options
	flag.StringVar(&config.ElevenLabsVoiceID, "elevenlabs-voice-id", "", "ElevenLabs voice ID (e.g., '21m00Tcm4TlvDq8ikWAM')")
	flag.StringVar(&config.ElevenLabsModel, "elevenlabs-model", "eleven_monolingual_v1", "ElevenLabs model ID")
	flag.StringVar(&config.ElevenLabsAPIKey, "elevenlabs-api-key", "", "ElevenLabs API key (prefer ELEVENLABS_API_KEY env var)")

	// Common options
	flag.StringVar(&config.Format, "format", "aiff", "Output audio format (aiff, m4a, mp3)")
	flag.StringVar(&config.Prefix, "prefix", "section", "Prefix for output filenames")
	flag.BoolVar(&config.ListVoices, "list-voices", false, "List all available voices (uses cache if available)")
	flag.BoolVar(&config.RefreshCache, "refresh-cache", false, "Force refresh of voice cache when listing voices")
	flag.StringVar(&config.ExportVoices, "export-voices", "", "Export cached voices to JSON file (e.g., voices.json)")
	flag.BoolVar(&config.Version, "version", false, "Print version and exit")
	flag.BoolVar(&config.Debug, "debug", false, "Enable debug logging")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Show what would be generated without creating files")

	flag.Usage = func() {
		log.Default("Markdown to Audio Generator")
		log.Faint("Convert markdown H2 sections to audio files using TTS providers.")
		log.Blank()
		log.Default("Usage:")
		log.Faint(fmt.Sprintf("  %s [options]", os.Args[0]))
		log.Blank()
		log.Default("Options:")
		flag.PrintDefaults()
		log.Blank()
		log.Default("Examples (macOS say provider):")
		log.Faint("  # Process a single file with say (default)")
		log.Faint(fmt.Sprintf("  %s -f script.md -p british-female", os.Args[0]))
		log.Blank()
		log.Faint("  # Process directory with custom voice and rate")
		log.Faint(fmt.Sprintf("  %s -d ./docs -v Kate -r 170", os.Args[0]))
		log.Blank()
		log.Faint("  # Generate m4a files")
		log.Faint(fmt.Sprintf("  %s -d ./docs -p british-female -format m4a", os.Args[0]))
		log.Blank()
		log.Faint("  # List available say voices")
		log.Faint(fmt.Sprintf("  %s -list-voices", os.Args[0]))
		log.Blank()
		log.Default("Examples (ElevenLabs provider):")
		log.Faint("  # Use ElevenLabs with environment variable")
		log.Faint("  export ELEVENLABS_API_KEY='your-key'")
		log.Faint(fmt.Sprintf("  %s -provider elevenlabs -elevenlabs-voice-id 21m00Tcm4TlvDq8ikWAM -f script.md", os.Args[0]))
		log.Blank()
		log.Faint("  # Use ElevenLabs with .env file")
		log.Faint("  echo 'ELEVENLABS_API_KEY=your-key' > .env")
		log.Faint(fmt.Sprintf("  %s -provider elevenlabs -elevenlabs-voice-id 21m00Tcm4TlvDq8ikWAM -d ./docs", os.Args[0]))
		log.Blank()
		log.Faint("  # List ElevenLabs voices")
		log.Faint(fmt.Sprintf("  %s -provider elevenlabs -list-voices", os.Args[0]))
		log.Blank()
		log.Default("Say Voice Presets:")
		log.Faint("  british-female, british-male, us-female, us-male,")
		log.Faint("  australian-female, indian-female")
	}

	flag.Parse()

	// Return early if version flag is set (skip all initialization)
	if config.Version {
		return config
	}

	// Determine voice to use (only for say provider)
	if config.Provider == "say" || config.Provider == "" {
		if config.Voice != "" {
			// Explicit voice specified, use it
		} else if preset != "" {
			if voice, ok := VoicePresets[preset]; ok {
				config.Voice = voice
			} else {
				fmt.Printf("Unknown preset: %s, using default voice 'Kate'\n", preset)
				config.Voice = "Kate"
			}
		} else if !config.ListVoices {
			config.Voice = "Kate"
			fmt.Println("No voice specified, using default: Kate")
		}
	}

	// Normalize provider name
	if config.Provider == "" {
		config.Provider = "say"
	}

	// Set default ElevenLabs voice if not specified and not listing voices
	if config.Provider == "elevenlabs" && config.ElevenLabsVoiceID == "" && !config.ListVoices {
		config.ElevenLabsVoiceID = DefaultElevenLabsVoiceID
		fmt.Println("No ElevenLabs voice specified, using default: Rachel (21m00Tcm4TlvDq8ikWAM)")
	}

	// Load ElevenLabs voice settings from environment variables (with defaults)
	if config.Provider == "elevenlabs" {
		config.ElevenLabsStability = getEnvFloat("ELEVENLABS_STABILITY", 0.5)
		config.ElevenLabsSimilarityBoost = getEnvFloat("ELEVENLABS_SIMILARITY_BOOST", 0.5)
		config.ElevenLabsStyle = getEnvFloat("ELEVENLABS_STYLE", 0.0)
		config.ElevenLabsUseSpeakerBoost = getEnvBool("ELEVENLABS_USE_SPEAKER_BOOST", true)
		config.ElevenLabsSpeed = getEnvFloat("ELEVENLABS_SPEED", 1.0)
	}

	return config
}

// getEnvFloat retrieves a float64 value from environment variable with a default fallback
func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvBool retrieves a bool value from environment variable with a default fallback
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// Validate checks if the configuration is valid
func (c Config) Validate() error {
	// Check mutual exclusivity of -f and -d
	if c.MarkdownFile != "" && c.InputDir != "" {
		return fmt.Errorf("cannot use both -f and -d flags; use one or the other")
	}

	// Check that at least one input is provided (unless listing voices)
	if !c.ListVoices && c.MarkdownFile == "" && c.InputDir == "" {
		return fmt.Errorf("either -f (file) or -d (directory) is required")
	}

	// Validate provider
	if c.Provider != "say" && c.Provider != "elevenlabs" {
		return fmt.Errorf("invalid provider %q: must be 'say' or 'elevenlabs'", c.Provider)
	}

	// Validate provider-specific requirements
	if c.Provider == "elevenlabs" && !c.ListVoices {
		if c.ElevenLabsVoiceID == "" {
			return fmt.Errorf("ElevenLabs voice ID is required: use -elevenlabs-voice-id flag")
		}
	}

	return nil
}

// IsDirectoryMode returns true if processing a directory
func (c Config) IsDirectoryMode() bool {
	return c.InputDir != ""
}

// maskSecret masks sensitive string data for safe display in logs
// Shows first 4 and last 4 characters, masks the middle with asterisks
func maskSecret(secret string) string {
	if secret == "" {
		return "[not set]"
	}
	if len(secret) <= 8 {
		return "****" // Too short to partially reveal
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

// Print displays the configuration
// NOTE: This method is safe for logging - sensitive data (API keys) are never printed
func (c Config) Print() {
	fmt.Println("\nConfiguration:")
	if c.IsDirectoryMode() {
		fmt.Printf("  Input directory: %s\n", c.InputDir)
	} else {
		fmt.Printf("  Markdown file: %s\n", c.MarkdownFile)
	}
	fmt.Printf("  TTS Provider: %s\n", c.Provider)

	// Provider-specific configuration
	switch c.Provider {
	case "say":
		fmt.Printf("  Voice: %s\n", c.Voice)
		fmt.Printf("  Rate: %d\n", c.Rate)
	case "elevenlabs":
		fmt.Printf("  Voice ID: %s\n", c.ElevenLabsVoiceID)
		fmt.Printf("  Model: %s\n", c.ElevenLabsModel)
		// API key is intentionally not printed for security
		// If debugging is needed, check environment variable ELEVENLABS_API_KEY
		if c.ElevenLabsAPIKey != "" {
			fmt.Printf("  API Key: %s\n", maskSecret(c.ElevenLabsAPIKey))
		}
	}

	fmt.Printf("  Format: %s\n", c.Format)
	fmt.Printf("  Output directory: %s\n\n", c.OutputDir)
}
