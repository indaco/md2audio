package google

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/api/option"

	"github.com/indaco/md2audio/internal/logger"
	"github.com/indaco/md2audio/internal/tts"
	"github.com/indaco/md2audio/internal/utils"
)

const (
	// DefaultVoiceName is the default Google Cloud TTS voice
	DefaultVoiceName = "en-US-Neural2-F"

	// DefaultLanguageCode is the default language
	DefaultLanguageCode = "en-US"

	// EnvVarCredentials is the environment variable for service account credentials
	EnvVarCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
)

// VoiceType represents the voice quality tier
type VoiceType string

const (
	VoiceTypeStandard VoiceType = "Standard"
	VoiceTypeWaveNet  VoiceType = "WaveNet"
	VoiceTypeNeural2  VoiceType = "Neural2"
	VoiceTypeStudio   VoiceType = "Studio"
	VoiceTypePolyglot VoiceType = "Polyglot"
)

// Client implements the TTS Provider interface for Google Cloud Text-to-Speech API.
type Client struct {
	client       *texttospeech.Client
	log          logger.LoggerInterface
	languageCode string
	speakingRate float64 // 0.25 to 4.0
	pitch        float64 // -20.0 to 20.0
	volumeGainDb float64 // -96.0 to 16.0
}

// Config holds configuration for the Google Cloud TTS client.
type Config struct {
	// CredentialsFile is the path to the service account JSON file.
	// If empty, uses GOOGLE_APPLICATION_CREDENTIALS environment variable.
	CredentialsFile string

	// LanguageCode is the voice language (e.g., "en-US", "en-GB").
	// Default: "en-US"
	LanguageCode string

	// SpeakingRate is the speed multiplier (0.25 to 4.0).
	// Default: 1.0 (normal speed)
	SpeakingRate float64

	// Pitch adjustment in semitones (-20.0 to 20.0).
	// Default: 0.0 (no change)
	Pitch float64

	// VolumeGainDb is the volume gain in decibels (-96.0 to 16.0).
	// Default: 0.0 (no change)
	VolumeGainDb float64
}

// NewClient creates a new Google Cloud TTS client.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	// Prepare client options
	var opts []option.ClientOption

	// Use credentials file if provided
	if cfg.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.CredentialsFile))
	} else if credsPath := os.Getenv(EnvVarCredentials); credsPath != "" {
		// Environment variable is set, client will use it automatically
		opts = append(opts, option.WithCredentialsFile(credsPath))
	} else {
		return nil, fmt.Errorf("credentials not found for Google Cloud: set %s environment variable or provide CredentialsFile", EnvVarCredentials)
	}

	// Create the client
	client, err := texttospeech.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Cloud TTS client: %w", err)
	}

	// Set defaults
	languageCode := cfg.LanguageCode
	if languageCode == "" {
		languageCode = DefaultLanguageCode
	}

	speakingRate := cfg.SpeakingRate
	if speakingRate == 0 {
		speakingRate = 1.0 // Default normal speed
	}

	pitch := cfg.Pitch
	// pitch defaults to 0.0 if not set

	volumeGainDb := cfg.VolumeGainDb
	// volumeGainDb defaults to 0.0 if not set

	return &Client{
		client:       client,
		languageCode: languageCode,
		speakingRate: speakingRate,
		pitch:        pitch,
		volumeGainDb: volumeGainDb,
	}, nil
}

// Name returns the provider name.
func (c *Client) Name() string {
	return "google"
}

// SetLogger sets the logger for debug output.
func (c *Client) SetLogger(log logger.LoggerInterface) {
	c.log = log
}

// Close closes the client connection.
func (c *Client) Close() error {
	return c.client.Close()
}

// Generate creates audio from text using Google Cloud TTS.
func (c *Client) Generate(ctx context.Context, req tts.GenerateRequest) (string, error) {
	// Determine voice name
	voiceName := req.Voice
	if voiceName == "" {
		voiceName = DefaultVoiceName
	}

	// Parse voice name to extract language code (e.g., "en-US-Neural2-F" -> "en-US")
	languageCode := c.languageCode
	if len(voiceName) >= 5 && voiceName[2] == '-' {
		// Extract language code from voice name (first 5 chars: "en-US")
		languageCode = voiceName[:5]
	}

	// Determine speaking rate
	speakingRate := c.speakingRate
	if req.TargetDuration != nil && *req.TargetDuration > 0 {
		// Calculate speed to match target duration
		speakingRate = calculateSpeed(req.Text, *req.TargetDuration)
		if c.log != nil {
			c.log.Debug(fmt.Sprintf("Target duration: %.1fs, Calculated speed: %.2fx", *req.TargetDuration, speakingRate))
		}
	}

	// Prepare the synthesis request
	ttsReq := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{
				Text: req.Text,
			},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: languageCode,
			Name:         voiceName,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding:   getAudioEncoding(req.Format),
			SpeakingRate:    speakingRate,
			Pitch:           c.pitch,
			VolumeGainDb:    c.volumeGainDb,
			SampleRateHertz: getSampleRate(req.Format),
		},
	}

	// Log API request
	if c.log != nil {
		c.log.Debug(fmt.Sprintf("Google Cloud TTS API: Synthesize (voice: %s, lang: %s, rate: %.2f)", voiceName, languageCode, speakingRate))
	}

	// Execute request
	resp, err := c.client.SynthesizeSpeech(ctx, ttsReq)
	if err != nil {
		return "", fmt.Errorf("failed to synthesize speech: %w", err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(req.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine output path with correct extension
	outputPath := ensureCorrectExtension(req.OutputPath, req.Format)

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	// Write audio data to file
	if _, err := outFile.Write(resp.AudioContent); err != nil {
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}

	return outputPath, nil
}

// ListVoices retrieves available voices from Google Cloud TTS.
func (c *Client) ListVoices(ctx context.Context) ([]tts.Voice, error) {
	// Request voice list
	req := &texttospeechpb.ListVoicesRequest{
		// LanguageCode can be empty to list all voices
	}

	if c.log != nil {
		c.log.Debug("Google Cloud TTS API: ListVoices")
	}

	resp, err := c.client.ListVoices(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list voices: %w", err)
	}

	// Convert to tts.Voice
	voices := make([]tts.Voice, 0, len(resp.Voices))
	for _, v := range resp.Voices {
		// Get first language code (voices can support multiple languages)
		languageCode := ""
		if len(v.LanguageCodes) > 0 {
			languageCode = v.LanguageCodes[0]
		}

		// Determine voice type from name
		voiceType := determineVoiceType(v.Name)

		// Build description
		description := fmt.Sprintf("%s voice", voiceType)
		if v.NaturalSampleRateHertz > 0 {
			description += fmt.Sprintf(" (%d Hz)", v.NaturalSampleRateHertz)
		}

		// Determine gender
		gender := ""
		switch v.SsmlGender {
		case texttospeechpb.SsmlVoiceGender_MALE:
			gender = "male"
		case texttospeechpb.SsmlVoiceGender_FEMALE:
			gender = "female"
		case texttospeechpb.SsmlVoiceGender_NEUTRAL:
			gender = "neutral"
		}

		voices = append(voices, tts.Voice{
			ID:          v.Name,
			Name:        v.Name,
			Description: description,
			Language:    languageCode,
			Gender:      gender,
		})
	}

	return voices, nil
}

// getAudioEncoding returns the audio encoding for the specified format.
func getAudioEncoding(format string) texttospeechpb.AudioEncoding {
	switch format {
	case "mp3":
		return texttospeechpb.AudioEncoding_MP3
	case "wav":
		return texttospeechpb.AudioEncoding_LINEAR16
	case "ogg":
		return texttospeechpb.AudioEncoding_OGG_OPUS
	default:
		// Default to MP3
		return texttospeechpb.AudioEncoding_MP3
	}
}

// getSampleRate returns the appropriate sample rate for the format.
func getSampleRate(format string) int32 {
	switch format {
	case "wav":
		return 24000 // High quality for WAV
	case "mp3", "ogg":
		return 24000 // Standard quality for compressed formats
	default:
		return 24000
	}
}

// ensureCorrectExtension ensures the output path has the correct file extension.
func ensureCorrectExtension(outputPath, format string) string {
	expectedExt := "." + format
	currentExt := filepath.Ext(outputPath)

	if currentExt != expectedExt {
		return outputPath[:len(outputPath)-len(currentExt)] + expectedExt
	}
	return outputPath
}

// determineVoiceType extracts the voice type from the voice name.
// Examples: "en-US-Neural2-F" -> "Neural2", "en-US-Wavenet-A" -> "WaveNet"
func determineVoiceType(voiceName string) string {
	// Voice name format: {languageCode}-{voiceType}-{variant}
	// Example: en-US-Neural2-F, en-GB-Wavenet-A

	// Look for voice type keywords
	if contains(voiceName, "Neural2") {
		return "Neural2"
	}
	if contains(voiceName, "Wavenet") || contains(voiceName, "WaveNet") {
		return "WaveNet"
	}
	if contains(voiceName, "Studio") {
		return "Studio"
	}
	if contains(voiceName, "Polyglot") {
		return "Polyglot"
	}
	if contains(voiceName, "Standard") {
		return "Standard"
	}

	return "Standard"
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// calculateSpeed determines the speed multiplier needed to match target duration.
// Google Cloud TTS speed ranges from 0.25 (slower) to 4.0 (faster), with 1.0 being normal.
func calculateSpeed(text string, targetDuration float64) float64 {
	const (
		naturalWPM   = 150.0 // Assume natural speaking rate at speed 1.0 is ~150 words per minute
		minSpeed     = 0.25  // Google Cloud TTS minimum speed
		maxSpeed     = 4.0   // Google Cloud TTS maximum speed
		defaultSpeed = 1.0
	)

	wordCount := utils.CountWords(text)
	if wordCount == 0 {
		return defaultSpeed
	}

	// Calculate natural duration at speed 1.0
	naturalDuration := utils.EstimateDuration(text, naturalWPM)

	// Calculate required speed: naturalDuration / targetDuration
	// If target is shorter, we need faster speed (>1.0)
	// If target is longer, we need slower speed (<1.0)
	speed := naturalDuration / targetDuration
	originalSpeed := speed

	// Clamp to Google Cloud TTS valid range
	speed = utils.ClampFloat64(speed, minSpeed, maxSpeed)

	// Warn if we had to clamp
	if speed != originalSpeed {
		if originalSpeed < minSpeed {
			fmt.Fprintf(os.Stderr, "Warning: Required speed (%.2f) is below minimum, clamping to %.2f (audio will be longer than target)\n", originalSpeed, minSpeed)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Required speed (%.2f) exceeds maximum, clamping to %.2f (audio will be shorter than target)\n", originalSpeed, maxSpeed)
		}
	}

	return speed
}
