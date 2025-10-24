// Package tts provides the text-to-speech provider abstraction and implementations.
// It defines the Provider interface that all TTS services must implement,
// along with common types and request/response structures.
//
// Key features:
//   - Provider interface for TTS abstraction
//   - Support for multiple providers (say, elevenlabs)
//   - Timing control and speed adjustment
//   - Voice listing and selection
//
// Providers:
//   - say: macOS built-in TTS (AIFF, M4A output)
//   - elevenlabs: ElevenLabs API (MP3 output)
package tts

import "context"

// Provider defines the interface for text-to-speech providers.
// Implementations include macOS 'say' command and ElevenLabs API.
type Provider interface {
	// Generate creates audio from text and returns the output file path.
	// The provider is responsible for creating the audio file at the specified path.
	Generate(ctx context.Context, req GenerateRequest) (string, error)

	// ListVoices returns available voices for this provider.
	ListVoices(ctx context.Context) ([]Voice, error)

	// Name returns the provider name (e.g., "say", "elevenlabs").
	Name() string
}

// GenerateRequest contains all parameters needed to generate audio.
type GenerateRequest struct {
	// Text is the content to convert to speech
	Text string

	// Voice is the voice identifier (provider-specific)
	Voice string

	// OutputPath is where the audio file should be created
	OutputPath string

	// Rate is the speaking rate in words per minute (optional, used by 'say' provider)
	Rate *int

	// ModelID is the TTS model identifier (optional, used by ElevenLabs)
	ModelID *string

	// Format is the desired audio format (e.g., "aiff", "mp3")
	Format string

	// TargetDuration is the desired duration in seconds (optional, for timing control)
	TargetDuration *float64
}

// Voice represents a TTS voice.
type Voice struct {
	// ID is the unique voice identifier
	ID string

	// Name is the human-readable voice name
	Name string

	// Description provides additional information about the voice
	Description string

	// Language is the voice language code (e.g., "en-US")
	Language string

	// Gender is the voice gender (if applicable)
	Gender string
}
