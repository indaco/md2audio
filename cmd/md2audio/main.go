package main

import (
	"fmt"
	"os"

	"github.com/indaco/md2audio/internal/cache"
	"github.com/indaco/md2audio/internal/cli"
	"github.com/indaco/md2audio/internal/config"
	"github.com/indaco/md2audio/internal/logger"
	"github.com/indaco/md2audio/internal/processor"
	"github.com/indaco/md2audio/internal/version"
)

// run executes the main application logic and returns an error if something fails.
func run(cfg config.Config, log logger.LoggerInterface) error {
	// Initialize voice cache
	voiceCache, err := cache.NewVoiceCache()
	if err != nil {
		return fmt.Errorf("failed to initialize voice cache: %w", err)
	}
	voiceCache.SetLogger(log) // Enable debug logging for cache operations
	defer func() {
		if closeErr := voiceCache.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close voice cache: %v\n", closeErr)
		}
	}()

	// Handle voice-related commands
	if cfg.ListVoices || cfg.ExportVoices != "" {
		return cli.HandleVoiceCommands(cfg, voiceCache, log)
	}

	// Validate configuration for audio processing
	if err := cfg.Validate(); err != nil {
		return err
	}

	cfg.Print()

	// Process based on mode
	if cfg.IsDirectoryMode() {
		return processor.ProcessDirectory(cfg, log)
	}
	return processor.ProcessFile(cfg.MarkdownFile, cfg.OutputDir, cfg, log)
}

func main() {
	// Create logger instance
	log := logger.NewDefaultLogger()

	cfg := config.Parse()

	// Enable debug logging if requested
	log.SetDebug(cfg.Debug)

	// Handle version flag
	if cfg.Version {
		fmt.Printf("md2audio version %s\n", version.GetVersion())
		return
	}

	if err := run(cfg, log); err != nil {
		log.Error("Fatal error:", err)
		os.Exit(1)
	}
}
