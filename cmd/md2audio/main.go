package main

import (
	"fmt"
	"os"

	"github.com/indaco/md2audio/internal/audio"
	"github.com/indaco/md2audio/internal/config"
	"github.com/indaco/md2audio/internal/processor"
)

// run executes the main application logic and returns an error if something fails.
func run(cfg config.Config) error {
	if cfg.ListVoices {
		if err := audio.ListAvailableVoices(); err != nil {
			return fmt.Errorf("failed to list voices: %w", err)
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
