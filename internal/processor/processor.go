package processor

import (
	"fmt"
	"os"
	"strings"

	"github.com/indaco/md2audio/internal/audio"
	"github.com/indaco/md2audio/internal/config"
	"github.com/indaco/md2audio/internal/parser"
	"github.com/indaco/md2audio/internal/tts"
	"github.com/indaco/md2audio/internal/tts/elevenlabs"
	"github.com/indaco/md2audio/internal/tts/say"
)

// createTTSProvider creates the appropriate TTS provider based on configuration
func createTTSProvider(cfg config.Config) (tts.Provider, error) {
	switch cfg.Provider {
	case "say":
		return say.NewProvider()
	case "elevenlabs":
		return elevenlabs.NewClient(elevenlabs.Config{
			APIKey:          cfg.ElevenLabsAPIKey,
			BaseURL:         "", // Use default
			Stability:       cfg.ElevenLabsStability,
			SimilarityBoost: cfg.ElevenLabsSimilarityBoost,
			Style:           cfg.ElevenLabsStyle,
			UseSpeakerBoost: cfg.ElevenLabsUseSpeakerBoost,
			Speed:           cfg.ElevenLabsSpeed,
		})
	default:
		return nil, fmt.Errorf("unsupported TTS provider: %s", cfg.Provider)
	}
}

// ProcessDirectory processes all markdown files in a directory recursively
func ProcessDirectory(cfg config.Config) error {
	fmt.Printf("Scanning directory: %s\n", cfg.InputDir)

	// Find all markdown files
	mdFiles, err := parser.FindMarkdownFiles(cfg.InputDir)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(mdFiles) == 0 {
		return fmt.Errorf("no markdown files found in directory: %s", cfg.InputDir)
	}

	fmt.Printf("Found %d markdown file(s)\n\n", len(mdFiles))

	totalSuccess := 0
	totalSections := 0

	// Process each markdown file
	for i, mdFile := range mdFiles {
		fmt.Printf("\n%s\n", strings.Repeat("=", 60))
		fmt.Printf("Processing file %d/%d: %s\n", i+1, len(mdFiles), mdFile.RelPath)
		fmt.Printf("%s\n", strings.Repeat("=", 60))

		// Get output directory for this file
		outputDir := mdFile.GetOutputDir(cfg.OutputDir)

		// Process the file
		successCount, sectionCount, err := processSingleFile(mdFile.AbsPath, outputDir, cfg)
		if err != nil {
			fmt.Printf("Warning: Failed to process %s: %v\n", mdFile.RelPath, err)
			continue
		}

		totalSuccess += successCount
		totalSections += sectionCount
	}

	// Final summary
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("Directory processing complete!\n")
	fmt.Printf("Generated %d/%d audio files from %d markdown file(s)\n", totalSuccess, totalSections, len(mdFiles))
	fmt.Printf("Output directory: %s\n", cfg.OutputDir)
	fmt.Printf("%s\n", strings.Repeat("=", 60))

	return nil
}

// ProcessFile processes a single markdown file
func ProcessFile(markdownFile, outputDir string, cfg config.Config) error {
	_, _, err := processSingleFile(markdownFile, outputDir, cfg)
	return err
}

// processSingleFile processes one markdown file and returns success count and section count
func processSingleFile(markdownFile, outputDir string, cfg config.Config) (int, int, error) {
	// Parse markdown file
	fmt.Println("Parsing markdown file...")
	sections, err := parser.ParseMarkdownFile(markdownFile)
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing markdown: %w", err)
	}

	if len(sections) == 0 {
		fmt.Println("Warning: No H2 sections found in the markdown file.")
		return 0, 0, nil
	}

	fmt.Printf("Found %d section(s)\n\n", len(sections))

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return 0, 0, fmt.Errorf("error creating output directory: %w", err)
	}

	// Create TTS provider
	provider, err := createTTSProvider(cfg)
	if err != nil {
		return 0, 0, fmt.Errorf("error creating TTS provider: %w", err)
	}

	fmt.Printf("Using TTS provider: %s\n\n", provider.Name())

	// Determine voice to use based on provider
	voice := cfg.Voice
	if cfg.Provider == "elevenlabs" {
		voice = cfg.ElevenLabsVoiceID
	}

	// Create audio generator
	generator := audio.NewGenerator(audio.GeneratorConfig{
		Voice:     voice,
		Rate:      cfg.Rate,
		Format:    cfg.Format,
		Prefix:    cfg.Prefix,
		OutputDir: outputDir,
		Provider:  provider,
	})

	// Generate audio for each section
	successCount := 0
	for i, section := range sections {
		fmt.Printf("\n--- Section %d/%d: %s ---\n", i+1, len(sections), section.Title)
		if section.HasTiming {
			fmt.Printf("Target Duration: %.1f seconds\n", section.Duration)
		}
		preview := section.Content
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		fmt.Printf("Text: %s\n", preview)

		if err := generator.Generate(section, i+1); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			successCount++
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 50))
	fmt.Printf("Complete! Generated %d/%d audio files\n", successCount, len(sections))
	fmt.Printf("Files saved to: %s\n", outputDir)
	fmt.Printf("%s\n", strings.Repeat("=", 50))

	return successCount, len(sections), nil
}
