package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff/v5"

	"github.com/indaco/md2audio/internal/env"
	"github.com/indaco/md2audio/internal/logger"
	"github.com/indaco/md2audio/internal/tts"
	"github.com/indaco/md2audio/internal/utils"
)

const (
	// TextToSpeechBaseURL is the default Text to Speech ElevenLabs API endpoint
	TextToSpeechBaseURL = "https://api.elevenlabs.io/v1"

	// VoicesBaseURL is the default Voices ElevenLabs API endpoint
	VoicesBaseURL = "https://api.elevenlabs.io/v2"

	// DefaultModel is the default TTS model
	DefaultModel = "eleven_multilingual_v2"

	// EnvVarAPIKey is the environment variable name for the API key
	EnvVarAPIKey = "ELEVENLABS_API_KEY"
)

// Client implements the TTS Provider interface for ElevenLabs API.
type Client struct {
	apiKey              string
	textToSpeechBaseURL string // Base URL for text-to-speech operations (v1)
	voicesBaseURL       string // Base URL for voices operations (v2)
	httpClient          *http.Client
	log                 logger.LoggerInterface // Optional logger for debug output

	// Default voice settings
	stability       float64
	similarityBoost float64
	style           float64
	useSpeakerBoost bool
	speed           float64
}

// Config holds configuration for the ElevenLabs client.
type Config struct {
	APIKey              string
	TextToSpeechBaseURL string // Base URL for text-to-speech operations (defaults to v1)
	VoicesBaseURL       string // Base URL for voices operations (defaults to v2)
	HTTPClient          *http.Client

	// Voice Settings (optional, with defaults)
	Stability       float64 // Voice consistency (0.0-1.0, default: 0.5)
	SimilarityBoost float64 // Voice similarity (0.0-1.0, default: 0.5)
	Style           float64 // Voice style/emotion (0.0-1.0, default: 0.0 = disabled)
	UseSpeakerBoost bool    // Boost similarity of synthesized speech (default: true)
	Speed           float64 // Speaking speed (0.7-1.2, default: 1.0, only for non-timed sections)
}

// NewClient creates a new ElevenLabs client.
// It loads the API key from environment variable or .env file.
func NewClient(cfg Config) (*Client, error) {
	// Load .env file if it exists (won't override existing env vars)
	if _, err := env.Load(".env"); err != nil {
		// Log warning but don't fail - env vars may already be set
		fmt.Fprintf(os.Stderr, "Warning: Failed to load .env file: %v\n", err)
	}

	// Get API key from config, env var, or error
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv(EnvVarAPIKey)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("ElevenLabs API key not found: set %s environment variable or provide in Config", EnvVarAPIKey)
	}

	// Set text-to-speech base URL
	textToSpeechBaseURL := cfg.TextToSpeechBaseURL
	if textToSpeechBaseURL == "" {
		textToSpeechBaseURL = TextToSpeechBaseURL
	}

	// Set voices base URL
	voicesBaseURL := cfg.VoicesBaseURL
	if voicesBaseURL == "" {
		voicesBaseURL = VoicesBaseURL
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 60 * time.Second,
		}
	}

	// Set voice settings with defaults if not provided
	stability := cfg.Stability
	if stability == 0 {
		stability = 0.5 // Default
	}

	similarityBoost := cfg.SimilarityBoost
	if similarityBoost == 0 {
		similarityBoost = 0.5 // Default
	}

	style := cfg.Style
	// Style defaults to 0.0 (disabled), so we don't need to check

	useSpeakerBoost := cfg.UseSpeakerBoost
	// Note: false is default for bool, but we want true as default
	// This is handled by config parsing setting true as default

	speed := cfg.Speed
	if speed == 0 {
		speed = 1.0 // Default natural speed
	}

	return &Client{
		apiKey:              apiKey,
		textToSpeechBaseURL: textToSpeechBaseURL,
		voicesBaseURL:       voicesBaseURL,
		httpClient:          httpClient,
		stability:           stability,
		similarityBoost:     similarityBoost,
		style:               style,
		useSpeakerBoost:     useSpeakerBoost,
		speed:               speed,
	}, nil
}

// Name returns the provider name.
func (c *Client) Name() string {
	return "elevenlabs"
}

// SetLogger sets the logger for debug output.
func (c *Client) SetLogger(log logger.LoggerInterface) {
	c.log = log
}

// retryableHTTPRequest executes an HTTP request with exponential backoff retry logic.
// Retries on network errors and 429/500/502/503 status codes.
func (c *Client) retryableHTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var lastErr error

	operation := func() error {
		// Clone the request for retry (in case body needs to be re-read)
		reqClone := req.Clone(ctx)

		var err error
		resp, err = c.httpClient.Do(reqClone)
		if err != nil {
			// Network error - retry
			lastErr = err
			return err
		}

		// Check if we should retry based on status code
		if shouldRetry(resp.StatusCode) {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
			return lastErr
		}

		// Success or non-retryable error
		return nil
	}

	// Configure exponential backoff: 3 retries
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 1 * time.Second
	expBackoff.MaxInterval = 10 * time.Second
	expBackoff.Reset()

	// Wrap with max retries and context awareness
	retryCount := 0
	maxRetries := 3

	for retryCount < maxRetries {
		err := operation()
		if err == nil {
			return resp, nil
		}

		retryCount++
		if retryCount >= maxRetries {
			return nil, lastErr
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Wait with exponential backoff
		wait := expBackoff.NextBackOff()
		if wait == backoff.Stop {
			return nil, lastErr
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
			// Continue to next retry
		}
	}

	return nil, lastErr
}

// shouldRetry returns true if the HTTP status code indicates a retryable error.
func shouldRetry(statusCode int) bool {
	return statusCode == 429 || // Too Many Requests
		statusCode == 500 || // Internal Server Error
		statusCode == 502 || // Bad Gateway
		statusCode == 503 // Service Unavailable
}

// Generate creates audio from text using the ElevenLabs API.
func (c *Client) Generate(ctx context.Context, req tts.GenerateRequest) (string, error) {
	// Determine model
	modelID := DefaultModel
	if req.ModelID != nil && *req.ModelID != "" {
		modelID = *req.ModelID
	}

	// Prepare voice settings from client defaults
	voiceSettings := c.prepareVoiceSettings(req)

	// Prepare request body
	reqBody := TTSRequest{
		Text:          req.Text,
		ModelID:       modelID,
		VoiceSettings: voiceSettings,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/text-to-speech/%s", c.textToSpeechBaseURL, req.Voice)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("xi-api-key", c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "audio/mpeg")

	// Log API request
	if c.log != nil {
		c.log.Debug(fmt.Sprintf("ElevenLabs API: POST /text-to-speech/%s (model: %s)", req.Voice, modelID))
	}

	// Execute request with retry logic
	resp, err := c.retryableHTTPRequest(ctx, httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response status (non-retryable errors)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(req.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine output path with correct extension
	outputPath := req.OutputPath
	// ElevenLabs returns MP3, ensure correct extension
	if filepath.Ext(outputPath) != ".mp3" {
		outputPath = outputPath[:len(outputPath)-len(filepath.Ext(outputPath))] + ".mp3"
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	// Copy audio data to file
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}

	return outputPath, nil
}

// ListVoices retrieves available voices from ElevenLabs API.
func (c *Client) ListVoices(ctx context.Context) ([]tts.Voice, error) {
	url := fmt.Sprintf("%s/voices", c.voicesBaseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("xi-api-key", c.apiKey)

	// Log API request
	if c.log != nil {
		c.log.Debug("ElevenLabs API: GET /voices")
	}

	// Execute request with retry logic
	resp, err := c.retryableHTTPRequest(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response status (non-retryable errors)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var voicesResp VoicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&voicesResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to tts.Voice
	voices := make([]tts.Voice, len(voicesResp.Voices))
	for i, v := range voicesResp.Voices {
		voices[i] = tts.Voice{
			ID:          v.VoiceID,
			Name:        v.Name,
			Description: v.Description,
			Language:    v.Labels.Language,
			Gender:      v.Labels.Gender,
		}
	}

	return voices, nil
}

// TTSRequest represents the request body for text-to-speech API.
type TTSRequest struct {
	Text          string         `json:"text"`
	ModelID       string         `json:"model_id"`
	VoiceSettings *VoiceSettings `json:"voice_settings,omitempty"`
}

// VoiceSettings contains voice configuration parameters.
type VoiceSettings struct {
	Stability       float64  `json:"stability,omitempty"`
	SimilarityBoost float64  `json:"similarity_boost,omitempty"`
	Style           *float64 `json:"style,omitempty"`             // Range: 0.0-1.0
	UseSpeakerBoost *bool    `json:"use_speaker_boost,omitempty"` // Boost the similarity of the synthesized speech
	Speed           *float64 `json:"speed,omitempty"`             // Range: 0.7-1.2 (default: 1.0)
}

// VoicesResponse represents the response from the voices API.
type VoicesResponse struct {
	Voices []VoiceInfo `json:"voices"`
}

// VoiceInfo contains information about a voice.
type VoiceInfo struct {
	VoiceID     string      `json:"voice_id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Labels      VoiceLabels `json:"labels"`
}

// VoiceLabels contains metadata about a voice.
type VoiceLabels struct {
	Language string `json:"language"`
	Gender   string `json:"gender"`
}

// prepareVoiceSettings creates voice settings for the TTS request.
// It uses client defaults and handles speed settings based on timing annotations.
func (c *Client) prepareVoiceSettings(req tts.GenerateRequest) *VoiceSettings {
	settings := &VoiceSettings{
		Stability:       c.stability,
		SimilarityBoost: c.similarityBoost,
	}

	// Add optional settings if non-default
	if c.style > 0 {
		settings.Style = &c.style
	}
	if c.useSpeakerBoost {
		useSpeakerBoost := true
		settings.UseSpeakerBoost = &useSpeakerBoost
	}

	// Speed handling: timing annotation overrides default speed
	if req.TargetDuration != nil && *req.TargetDuration > 0 {
		// Calculate speed to match target duration
		speed := calculateSpeed(req.Text, *req.TargetDuration)
		settings.Speed = &speed
		// Note: Using stderr for progress messages to avoid polluting stdout
		// TODO: Consider passing logger via context or provider interface for better integration
		fmt.Fprintf(os.Stderr, "Target duration: %.1fs, Calculated speed: %.2fx\n", *req.TargetDuration, speed)
	} else if c.speed != 1.0 && c.speed > 0 {
		// Use configured default speed for non-timed sections (only if explicitly set)
		settings.Speed = &c.speed
		fmt.Fprintf(os.Stderr, "Using configured speed: %.2fx\n", c.speed)
	}

	return settings
}

// calculateSpeed determines the speed multiplier needed to match target duration.
// ElevenLabs speed ranges from 0.7 (slower) to 1.2 (faster), with 1.0 being normal.
func calculateSpeed(text string, targetDuration float64) float64 {
	const (
		naturalWPM   = 150.0 // Assume natural speaking rate at speed 1.0 is ~150 words per minute
		minSpeed     = 0.7   // ElevenLabs minimum speed
		maxSpeed     = 1.2   // ElevenLabs maximum speed
		defaultSpeed = 1.0
	)

	wordCount := utils.CountWords(text)
	if wordCount == 0 {
		return defaultSpeed
	}

	// Calculate natural duration at speed 1.0 using utility
	naturalDuration := utils.EstimateDuration(text, naturalWPM)

	// Calculate required speed: naturalDuration / targetDuration
	// If target is shorter, we need faster speed (>1.0)
	// If target is longer, we need slower speed (<1.0)
	speed := naturalDuration / targetDuration
	originalSpeed := speed

	// Clamp to ElevenLabs valid range
	speed = utils.ClampFloat64(speed, minSpeed, maxSpeed)

	// Warn if we had to clamp
	if speed != originalSpeed {
		if originalSpeed < minSpeed {
			fmt.Fprintf(os.Stderr, "Warning: Required speed (%.2f) is below minimum, clamping to %.1f (audio will be longer than target)\n", originalSpeed, minSpeed)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Required speed (%.2f) exceeds maximum, clamping to %.1f (audio will be shorter than target)\n", originalSpeed, maxSpeed)
		}
	}

	return speed
}
