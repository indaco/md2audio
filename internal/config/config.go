package config

import (
	"flag"
	"fmt"
	"os"
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

// Config holds the application configuration
type Config struct {
	MarkdownFile string
	InputDir     string
	OutputDir    string
	Voice        string
	Rate         int
	Format       string
	Prefix       string
	ListVoices   bool
}

// Parse parses command-line flags and returns the configuration
func Parse() Config {
	config := Config{}

	flag.StringVar(&config.MarkdownFile, "f", "", "Input markdown file (use -f or -d, not both)")
	flag.StringVar(&config.InputDir, "d", "", "Input directory to process recursively (use -f or -d, not both)")
	flag.StringVar(&config.OutputDir, "o", "./audio_sections", "Output directory for audio files")

	var preset string
	flag.StringVar(&preset, "p", "", "Voice preset (british-female, british-male, us-female, us-male, australian-female, indian-female)")
	flag.StringVar(&config.Voice, "v", "", "Specific voice name (overrides preset)")
	flag.IntVar(&config.Rate, "r", 180, "Speaking rate (lower = slower)")
	flag.StringVar(&config.Format, "format", "aiff", "Output audio format (aiff or m4a)")
	flag.StringVar(&config.Prefix, "prefix", "section", "Prefix for output filenames")
	flag.BoolVar(&config.ListVoices, "list-voices", false, "List all available voices and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Markdown to Audio Generator\n")
		fmt.Fprintf(os.Stderr, "Convert markdown H2 sections to audio files using macOS say command.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Process a single file\n")
		fmt.Fprintf(os.Stderr, "  %s -f script.md -p british-female\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Process entire directory recursively\n")
		fmt.Fprintf(os.Stderr, "  %s -d ./docs -p british-female\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Use specific voice with custom rate\n")
		fmt.Fprintf(os.Stderr, "  %s -f script.md -v Kate -r 170\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Generate m4a files instead of aiff\n")
		fmt.Fprintf(os.Stderr, "  %s -d ./docs -p british-female -format m4a\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # List all available voices\n")
		fmt.Fprintf(os.Stderr, "  %s -list-voices\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Voice Presets:\n")
		fmt.Fprintf(os.Stderr, "  british-female, british-male, us-female, us-male,\n")
		fmt.Fprintf(os.Stderr, "  australian-female, indian-female\n")
	}

	flag.Parse()

	// Determine voice to use
	if config.Voice != "" {
		// Explicit voice specified, use it
	} else if preset != "" {
		if voice, ok := VoicePresets[preset]; ok {
			config.Voice = voice
		} else {
			fmt.Printf("Unknown preset: %s, using default voice 'Kate'\n", preset)
			config.Voice = "Kate"
		}
	} else {
		config.Voice = "Kate"
		fmt.Println("No voice specified, using default: Kate")
	}

	return config
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

	return nil
}

// IsDirectoryMode returns true if processing a directory
func (c Config) IsDirectoryMode() bool {
	return c.InputDir != ""
}

// Print displays the configuration
func (c Config) Print() {
	fmt.Println("\nConfiguration:")
	if c.IsDirectoryMode() {
		fmt.Printf("  Input directory: %s\n", c.InputDir)
	} else {
		fmt.Printf("  Markdown file: %s\n", c.MarkdownFile)
	}
	fmt.Printf("  Voice: %s\n", c.Voice)
	fmt.Printf("  Rate: %d\n", c.Rate)
	fmt.Printf("  Format: %s\n", c.Format)
	fmt.Printf("  Output directory: %s\n\n", c.OutputDir)
}
