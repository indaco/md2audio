package processor

import (
	"fmt"
	"os"

	"github.com/indaco/md2audio/internal/audio"
	"github.com/indaco/md2audio/internal/cli"
	"github.com/indaco/md2audio/internal/config"
	"github.com/indaco/md2audio/internal/logger"
	"github.com/indaco/md2audio/internal/parser"
)

// ProcessDirectory processes all markdown files in a directory recursively
func ProcessDirectory(cfg config.Config, log logger.LoggerInterface) error {
	log.Info("Scanning directory:", cfg.InputDir)

	// Find all markdown files
	mdFiles, err := parser.FindMarkdownFiles(cfg.InputDir)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(mdFiles) == 0 {
		return fmt.Errorf("no markdown files found in directory: %s", cfg.InputDir)
	}

	log.Success(fmt.Sprintf("Found %d markdown file(s)", len(mdFiles)))
	log.Blank()

	totalSuccess := 0
	totalSections := 0

	// Process each markdown file
	for i, mdFile := range mdFiles {
		log.Blank()
		log.Info(fmt.Sprintf("Processing file %d/%d:", i+1, len(mdFiles))).WithAttrs("file", mdFile.RelPath)

		// Get output directory for this file
		outputDir := mdFile.GetOutputDir(cfg.OutputDir)

		// Process the file
		successCount, sectionCount, err := processSingleFile(mdFile.AbsPath, outputDir, cfg, log)
		if err != nil {
			log.Warning(fmt.Sprintf("Failed to process %s: %v", mdFile.RelPath, err))
			continue
		}

		totalSuccess += successCount
		totalSections += sectionCount
	}

	// Final summary
	log.Blank()
	log.Success("Directory processing complete!")
	log.Info(fmt.Sprintf("Generated %d/%d audio files from %d markdown file(s)", totalSuccess, totalSections, len(mdFiles)))
	log.Info("Output directory:", cfg.OutputDir)

	return nil
}

// ProcessFile processes a single markdown file
func ProcessFile(markdownFile, outputDir string, cfg config.Config, log logger.LoggerInterface) error {
	_, _, err := processSingleFile(markdownFile, outputDir, cfg, log)
	return err
}

// processSingleFile processes one markdown file and returns success count and section count
func processSingleFile(markdownFile, outputDir string, cfg config.Config, log logger.LoggerInterface) (int, int, error) {
	// Parse markdown file
	log.Info("Parsing markdown file...")
	sections, err := parser.ParseMarkdownFile(markdownFile)
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing markdown: %w", err)
	}

	if len(sections) == 0 {
		log.Warning("No H2 sections found in the markdown file.")
		return 0, 0, nil
	}

	log.Success(fmt.Sprintf("Found %d section(s)", len(sections)))
	log.Blank()

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return 0, 0, fmt.Errorf("error creating output directory: %w", err)
	}

	// Create TTS provider
	provider, err := cli.CreateProvider(cfg)
	if err != nil {
		return 0, 0, fmt.Errorf("error creating TTS provider: %w", err)
	}

	log.Info("Using TTS provider:", provider.Name())
	log.Blank()

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
	}, log)

	// Generate audio for each section
	successCount := 0
	for i, section := range sections {
		log.Blank()
		log.Info(fmt.Sprintf("Section %d/%d:", i+1, len(sections))).WithAttrs("title", section.Title)

		if section.HasTiming {
			log.WithIndent(true)
			log.Hint(fmt.Sprintf("Target duration: %.1f seconds", section.Duration))
			log.WithIndent(false)
		}

		preview := section.Content
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		log.WithIndent(true)
		log.Hint(fmt.Sprintf("Text: %s", preview))
		log.WithIndent(false)

		if err := generator.Generate(section, i+1); err != nil {
			log.Error("Failed:", err)
		} else {
			successCount++
		}
	}

	log.Blank()
	log.Success(fmt.Sprintf("Complete! Generated %d/%d audio files", successCount, len(sections)))
	log.Info("Files saved to:", outputDir)

	return successCount, len(sections), nil
}
