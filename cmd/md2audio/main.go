package main

import (
	"context"
	"fmt"
	"os"

	"github.com/indaco/md2audio/internal/config"
	"github.com/indaco/md2audio/internal/processor"
	"github.com/indaco/md2audio/internal/tts"
	"github.com/indaco/md2audio/internal/tts/elevenlabs"
	"github.com/indaco/md2audio/internal/tts/say"
)

// run executes the main application logic and returns an error if something fails.
func run(cfg config.Config) error {
	if cfg.ListVoices {
		// Create appropriate TTS provider
		var provider tts.Provider
		var err error

		switch cfg.Provider {
		case "say", "":
			provider, err = say.NewProvider()
		case "elevenlabs":
			provider, err = elevenlabs.NewClient(elevenlabs.Config{
				APIKey: cfg.ElevenLabsAPIKey,
			})
		default:
			return fmt.Errorf("unsupported provider: %s", cfg.Provider)
		}

		if err != nil {
			return fmt.Errorf("failed to create TTS provider: %w", err)
		}

		fmt.Printf("Available voices for %s provider:\n\n", provider.Name())

		voices, err := provider.ListVoices(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list voices: %w", err)
		}

		for _, voice := range voices {
			fmt.Printf("%-20s %-10s", voice.Name, voice.Language)
			if voice.Description != "" {
				fmt.Printf(" - %s", voice.Description)
			}
			fmt.Println()
		}

		return nil
	}

	// Validate configuration
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

func main() {
	cfg := config.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
