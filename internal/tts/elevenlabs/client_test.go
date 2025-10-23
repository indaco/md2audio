package elevenlabs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/md2audio/internal/tts"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		envVar      string
		expectError bool
		errorMsg    string
	}{
		{
			name: "with API key in config",
			config: Config{
				APIKey: "test-api-key",
			},
			expectError: false,
		},
		{
			name:        "with API key in env var",
			config:      Config{},
			envVar:      "test-env-key",
			expectError: false,
		},
		{
			name:        "without API key",
			config:      Config{},
			expectError: true,
			errorMsg:    "API key not found",
		},
		{
			name: "with custom base URL",
			config: Config{
				APIKey:  "test-api-key",
				BaseURL: "https://custom.api.com",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envVar != "" {
				_ = os.Setenv(EnvVarAPIKey, tt.envVar)
				defer func() { _ = os.Unsetenv(EnvVarAPIKey) }()
			} else {
				_ = os.Unsetenv(EnvVarAPIKey)
			}

			client, err := NewClient(tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("Expected client but got nil")
				return
			}

			// Verify API key was set
			expectedKey := tt.config.APIKey
			if expectedKey == "" {
				expectedKey = tt.envVar
			}
			if client.apiKey != expectedKey {
				t.Errorf("API key = %q, want %q", client.apiKey, expectedKey)
			}

			// Verify base URL
			expectedURL := tt.config.BaseURL
			if expectedURL == "" {
				expectedURL = DefaultBaseURL
			}
			if client.baseURL != expectedURL {
				t.Errorf("Base URL = %q, want %q", client.baseURL, expectedURL)
			}
		})
	}
}

func TestClient_Name(t *testing.T) {
	client := &Client{apiKey: "test"}
	if got := client.Name(); got != "elevenlabs" {
		t.Errorf("Name() = %q, want %q", got, "elevenlabs")
	}
}

func TestClient_Generate(t *testing.T) {
	tests := []struct {
		name         string
		request      tts.GenerateRequest
		serverStatus int
		serverBody   string
		expectError  bool
		errorMsg     string
	}{
		{
			name: "successful generation",
			request: tts.GenerateRequest{
				Text:       "Hello world",
				Voice:      "21m00Tcm4TlvDq8ikWAM",
				OutputPath: "/tmp/test.mp3",
			},
			serverStatus: http.StatusOK,
			serverBody:   "fake-audio-data",
			expectError:  false,
		},
		{
			name: "with custom model",
			request: tts.GenerateRequest{
				Text:       "Hello world",
				Voice:      "21m00Tcm4TlvDq8ikWAM",
				OutputPath: "/tmp/test.mp3",
				ModelID:    stringPtr("eleven_multilingual_v2"),
			},
			serverStatus: http.StatusOK,
			serverBody:   "fake-audio-data",
			expectError:  false,
		},
		{
			name: "API error - unauthorized",
			request: tts.GenerateRequest{
				Text:       "Hello world",
				Voice:      "invalid-voice",
				OutputPath: "/tmp/test.mp3",
			},
			serverStatus: http.StatusUnauthorized,
			serverBody:   `{"detail":"Invalid API key"}`,
			expectError:  true,
			errorMsg:     "401",
		},
		{
			name: "API error - quota exceeded",
			request: tts.GenerateRequest{
				Text:       "Hello world",
				Voice:      "21m00Tcm4TlvDq8ikWAM",
				OutputPath: "/tmp/test.mp3",
			},
			serverStatus: http.StatusTooManyRequests,
			serverBody:   `{"detail":"Quota exceeded"}`,
			expectError:  true,
			errorMsg:     "429",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				if r.Header.Get("xi-api-key") != "test-api-key" {
					t.Errorf("Expected xi-api-key header, got %q", r.Header.Get("xi-api-key"))
				}

				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected application/json content-type, got %q", r.Header.Get("Content-Type"))
				}

				// Verify request body contains text
				body, _ := io.ReadAll(r.Body)
				if !strings.Contains(string(body), tt.request.Text) {
					t.Errorf("Request body should contain %q, got %q", tt.request.Text, string(body))
				}

				// Send response
				w.WriteHeader(tt.serverStatus)
				_, _ = fmt.Fprint(w, tt.serverBody)
			}))
			defer server.Close()

			// Create client with mock server
			client := &Client{
				apiKey:     "test-api-key",
				baseURL:    server.URL,
				httpClient: server.Client(),
			}

			// Create temp directory for output
			tmpDir := t.TempDir()
			tt.request.OutputPath = filepath.Join(tmpDir, "test.mp3")

			// Execute Generate
			outputPath, err := client.Generate(context.Background(), tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify output file was created
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Errorf("Output file not created at %s", outputPath)
			}

			// Verify file contains expected data
			data, err := os.ReadFile(outputPath)
			if err != nil {
				t.Errorf("Failed to read output file: %v", err)
			}
			if string(data) != tt.serverBody {
				t.Errorf("Output file content = %q, want %q", string(data), tt.serverBody)
			}

			// Verify output path has .mp3 extension
			if filepath.Ext(outputPath) != ".mp3" {
				t.Errorf("Output path should have .mp3 extension, got %s", outputPath)
			}
		})
	}
}

func TestClient_ListVoices(t *testing.T) {
	tests := []struct {
		name         string
		serverStatus int
		serverBody   string
		expectError  bool
		errorMsg     string
		expectedLen  int
	}{
		{
			name:         "successful list",
			serverStatus: http.StatusOK,
			serverBody: `{
				"voices": [
					{
						"voice_id": "21m00Tcm4TlvDq8ikWAM",
						"name": "Rachel",
						"description": "American Female",
						"labels": {
							"language": "en-US",
							"gender": "female"
						}
					},
					{
						"voice_id": "AZnzlk1XvdvUeBnXmlld",
						"name": "Domi",
						"description": "American Female",
						"labels": {
							"language": "en-US",
							"gender": "female"
						}
					}
				]
			}`,
			expectError: false,
			expectedLen: 2,
		},
		{
			name:         "API error - unauthorized",
			serverStatus: http.StatusUnauthorized,
			serverBody:   `{"detail":"Invalid API key"}`,
			expectError:  true,
			errorMsg:     "401",
		},
		{
			name:         "invalid JSON response",
			serverStatus: http.StatusOK,
			serverBody:   `invalid json`,
			expectError:  true,
			errorMsg:     "failed to decode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}

				if r.Header.Get("xi-api-key") != "test-api-key" {
					t.Errorf("Expected xi-api-key header, got %q", r.Header.Get("xi-api-key"))
				}

				// Send response
				w.WriteHeader(tt.serverStatus)
				_, _ = fmt.Fprint(w, tt.serverBody)
			}))
			defer server.Close()

			// Create client with mock server
			client := &Client{
				apiKey:     "test-api-key",
				baseURL:    server.URL,
				httpClient: server.Client(),
			}

			// Execute ListVoices
			voices, err := client.ListVoices(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(voices) != tt.expectedLen {
				t.Errorf("Expected %d voices, got %d", tt.expectedLen, len(voices))
			}

			// Verify first voice details if available
			if len(voices) > 0 {
				v := voices[0]
				if v.ID != "21m00Tcm4TlvDq8ikWAM" {
					t.Errorf("Voice ID = %q, want %q", v.ID, "21m00Tcm4TlvDq8ikWAM")
				}
				if v.Name != "Rachel" {
					t.Errorf("Voice Name = %q, want %q", v.Name, "Rachel")
				}
				if v.Language != "en-US" {
					t.Errorf("Voice Language = %q, want %q", v.Language, "en-US")
				}
			}
		})
	}
}

func TestClient_GenerateOutputPathExtension(t *testing.T) {
	// Create mock server that returns successful response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "audio-data")
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-api-key",
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		inputPath      string
		expectedExt    string
		expectedExists bool
	}{
		{
			name:           "aiff extension becomes mp3",
			inputPath:      filepath.Join(tmpDir, "test.aiff"),
			expectedExt:    ".mp3",
			expectedExists: true,
		},
		{
			name:           "wav extension becomes mp3",
			inputPath:      filepath.Join(tmpDir, "test.wav"),
			expectedExt:    ".mp3",
			expectedExists: true,
		},
		{
			name:           "mp3 extension stays mp3",
			inputPath:      filepath.Join(tmpDir, "test.mp3"),
			expectedExt:    ".mp3",
			expectedExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tts.GenerateRequest{
				Text:       "Hello",
				Voice:      "test-voice",
				OutputPath: tt.inputPath,
			}

			outputPath, err := client.Generate(context.Background(), req)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if filepath.Ext(outputPath) != tt.expectedExt {
				t.Errorf("Output extension = %q, want %q", filepath.Ext(outputPath), tt.expectedExt)
			}

			if tt.expectedExists {
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Errorf("Output file not created at %s", outputPath)
				}
			}
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
