package audio

import (
	"os/exec"
	"runtime"
	"testing"

	"github.com/indaco/md2audio/internal/logger"
	"github.com/indaco/md2audio/internal/parser"
)

func TestEstimateSpeakingRate(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		targetDuration float64
		expectedRate   int
		description    string
	}{
		{
			name:           "8 seconds for short text",
			text:           "This is a test with exactly eight words in total content.",
			targetDuration: 8.0,
			expectedRate:   90, // Clamped to minimum (11 words / 8 sec = low rate)
			description:    "Short text in 8 seconds should hit minimum rate",
		},
		{
			name:           "short duration high rate",
			text:           "Short text here",
			targetDuration: 1.0,
			expectedRate:   171, // Will be adjusted
			description:    "Very short durations should increase rate",
		},
		{
			name:           "long duration low rate",
			text:           "This is a much longer piece of text that should result in a slower speaking rate when given a long target duration.",
			targetDuration: 30.0,
			expectedRate:   95, // Lower rate for longer duration
			description:    "Long durations should decrease rate",
		},
		{
			name:           "minimum rate clamping",
			text:           "Few words",
			targetDuration: 60.0,
			expectedRate:   90, // Should clamp to minimum
			description:    "Should clamp to minimum 90 WPM",
		},
		{
			name:           "maximum rate clamping",
			text:           "Many words here to test the maximum rate clamping functionality which should cap at three hundred sixty words per minute regardless of the calculated requirement based on content length",
			targetDuration: 1.0,
			expectedRate:   360, // Should clamp to maximum
			description:    "Should clamp to maximum 360 WPM",
		},
		{
			name:           "zero duration fallback",
			text:           "Some text",
			targetDuration: 0.0,
			expectedRate:   180, // Default fallback
			description:    "Zero duration should return default rate",
		},
		{
			name:           "negative duration fallback",
			text:           "Some text",
			targetDuration: -5.0,
			expectedRate:   180, // Default fallback
			description:    "Negative duration should return default rate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.NewDefaultLogger()
			result := estimateSpeakingRate(tt.text, tt.targetDuration, log)

			// Allow small variance due to rounding and 0.95 adjustment factor
			tolerance := 5
			if result < tt.expectedRate-tolerance || result > tt.expectedRate+tolerance {
				t.Errorf("%s: estimateSpeakingRate() = %d, want ~%d (±%d)",
					tt.description, result, tt.expectedRate, tolerance)
			}

			// Verify rate is within valid bounds
			if result < 90 || result > 360 {
				t.Errorf("Rate %d is outside valid bounds [90, 360]", result)
			}
		})
	}
}

func TestEstimateSpeakingRateBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		duration float64
		minRate  int
		maxRate  int
	}{
		{
			name:     "minimum boundary",
			text:     "a",
			duration: 100.0,
			minRate:  90,
			maxRate:  90,
		},
		{
			name:     "maximum boundary",
			text:     "word " + repeat("word ", 100),
			duration: 0.5,
			minRate:  360,
			maxRate:  360,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.NewDefaultLogger()
			result := estimateSpeakingRate(tt.text, tt.duration, log)
			if result < tt.minRate || result > tt.maxRate {
				t.Errorf("Expected rate between %d and %d, got %d", tt.minRate, tt.maxRate, result)
			}
		})
	}
}

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name   string
		config GeneratorConfig
	}{
		{
			name: "standard config",
			config: GeneratorConfig{
				Voice:     "Kate",
				Rate:      180,
				Format:    "aiff",
				Prefix:    "section",
				OutputDir: "./audio",
			},
		},
		{
			name: "m4a format config",
			config: GeneratorConfig{
				Voice:     "Samantha",
				Rate:      170,
				Format:    "m4a",
				Prefix:    "demo",
				OutputDir: "./output",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.NewDefaultLogger()
			gen := NewGenerator(tt.config, log)

			if gen == nil {
				t.Fatal("NewGenerator() returned nil")
				return
			}

			// Verify config is set correctly
			if gen.config.Voice != tt.config.Voice {
				t.Errorf("Voice = %q, want %q", gen.config.Voice, tt.config.Voice)
			}
			if gen.config.Rate != tt.config.Rate {
				t.Errorf("Rate = %d, want %d", gen.config.Rate, tt.config.Rate)
			}
			if gen.config.Format != tt.config.Format {
				t.Errorf("Format = %q, want %q", gen.config.Format, tt.config.Format)
			}
			if gen.config.Prefix != tt.config.Prefix {
				t.Errorf("Prefix = %q, want %q", gen.config.Prefix, tt.config.Prefix)
			}
			if gen.config.OutputDir != tt.config.OutputDir {
				t.Errorf("OutputDir = %q, want %q", gen.config.OutputDir, tt.config.OutputDir)
			}
		})
	}
}

func TestGeneratorConfigDefaults(t *testing.T) {
	config := GeneratorConfig{}
	log := logger.NewDefaultLogger()
	gen := NewGenerator(config, log)

	if gen == nil {
		t.Fatal("NewGenerator() with empty config returned nil")
	}
}

// Integration test: Only runs on macOS and checks if required commands exist
func TestMacOSCommandsExist(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-macOS platform")
	}

	commands := []string{"say", "afinfo"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			_, err := exec.LookPath(cmd)
			if err != nil {
				t.Errorf("Required command %q not found in PATH: %v", cmd, err)
			}
		})
	}
}

// Integration test: Check if say command accepts basic flags
func TestSayCommandBasicUsage(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	// Test if say command can list voices
	cmd := exec.Command("say", "-v", "?")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("say -v ? failed: %v", err)
	}

	if len(output) == 0 {
		t.Error("say -v ? returned empty output")
	}
}

// Test that GeneratorConfig can be created with various combinations
func TestGeneratorConfigVariations(t *testing.T) {
	configs := []GeneratorConfig{
		{Voice: "Kate", Rate: 180, Format: "aiff", Prefix: "s", OutputDir: "."},
		{Voice: "Alex", Rate: 200, Format: "m4a", Prefix: "audio", OutputDir: "/tmp"},
		{Voice: "Samantha", Rate: 150, Format: "aiff", Prefix: "test_", OutputDir: "./test"},
	}

	for i, cfg := range configs {
		log := logger.NewDefaultLogger()
		gen := NewGenerator(cfg, log)
		if gen == nil {
			t.Errorf("Config %d: NewGenerator() returned nil", i)
		}
	}
}

// Test estimateSpeakingRate with realistic content
func TestEstimateSpeakingRateRealistic(t *testing.T) {
	// Real-world example from demo_script_example.md
	text := `ServiceSage is an AI-powered support services recommendation system that helps
	organizations match project requirements with the right expertise—instantly. Our catalog
	contains 34 support services across 3 categories, delivering personalized recommendations in seconds.`

	targetDuration := 8.0 // From "SCENE 1: Hero Section (8s)"

	log := logger.NewDefaultLogger()
	rate := estimateSpeakingRate(text, targetDuration, log)

	// Should be somewhere in the reasonable range
	if rate < 150 || rate > 300 {
		t.Errorf("Rate %d seems unrealistic for normal speech", rate)
	}
}

// Helper function to repeat a string
func repeat(s string, count int) string {
	result := ""
	for range count {
		result += s
	}
	return result
}

// Test that Generate method exists and has correct signature
func TestGenerateMethodExists(t *testing.T) {
	log := logger.NewDefaultLogger()
	gen := NewGenerator(GeneratorConfig{
		Voice:     "Kate",
		Rate:      180,
		Format:    "aiff",
		Prefix:    "test",
		OutputDir: t.TempDir(),
	}, log)

	section := parser.Section{
		Title:     "Test",
		Content:   "Test content",
		Duration:  0,
		HasTiming: false,
	}

	// We're not actually testing Generate here (requires macOS commands)
	// Just verifying the method exists and can be called
	// This will fail on the actual say command, but that's expected
	_ = gen.Generate(section, 1)
	// We don't check the error because it's expected to fail without proper setup
}
