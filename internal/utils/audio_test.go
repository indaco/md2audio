package utils

import (
	"runtime"
	"strings"
	"testing"
)

func TestCountWords(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "empty string",
			text:     "",
			expected: 0,
		},
		{
			name:     "single word",
			text:     "hello",
			expected: 1,
		},
		{
			name:     "multiple words",
			text:     "hello world foo bar",
			expected: 4,
		},
		{
			name:     "words with extra whitespace",
			text:     "hello   world\n\tfoo",
			expected: 3,
		},
		{
			name:     "sentence with punctuation",
			text:     "Hello, world! How are you?",
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountWords(tt.text)
			if result != tt.expected {
				t.Errorf("CountWords(%q) = %d, want %d", tt.text, result, tt.expected)
			}
		})
	}
}

func TestCalculateWPM(t *testing.T) {
	tests := []struct {
		name            string
		wordCount       int
		durationSeconds float64
		expected        float64
	}{
		{
			name:            "120 words in 60 seconds",
			wordCount:       120,
			durationSeconds: 60.0,
			expected:        120.0,
		},
		{
			name:            "150 words in 60 seconds",
			wordCount:       150,
			durationSeconds: 60.0,
			expected:        150.0,
		},
		{
			name:            "75 words in 30 seconds",
			wordCount:       75,
			durationSeconds: 30.0,
			expected:        150.0,
		},
		{
			name:            "zero duration",
			wordCount:       100,
			durationSeconds: 0,
			expected:        0,
		},
		{
			name:            "negative duration",
			wordCount:       100,
			durationSeconds: -10,
			expected:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateWPM(tt.wordCount, tt.durationSeconds)
			if result != tt.expected {
				t.Errorf("CalculateWPM(%d, %.1f) = %.1f, want %.1f",
					tt.wordCount, tt.durationSeconds, result, tt.expected)
			}
		})
	}
}

func TestEstimateDuration(t *testing.T) {
	tests := []struct {
		name            string
		text            string
		wordsPerMinute  float64
		expectedSeconds float64
	}{
		{
			name:            "150 WPM with 150 words",
			text:            strings.Repeat("word ", 150),
			wordsPerMinute:  150.0,
			expectedSeconds: 60.0,
		},
		{
			name:            "120 WPM with 120 words",
			text:            strings.Repeat("word ", 120),
			wordsPerMinute:  120.0,
			expectedSeconds: 60.0,
		},
		{
			name:            "zero WPM",
			text:            "some text here",
			wordsPerMinute:  0,
			expectedSeconds: 0,
		},
		{
			name:            "negative WPM",
			text:            "some text here",
			wordsPerMinute:  -10,
			expectedSeconds: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateDuration(tt.text, tt.wordsPerMinute)
			if result != tt.expectedSeconds {
				t.Errorf("EstimateDuration(%q, %.1f) = %.1f, want %.1f",
					tt.text, tt.wordsPerMinute, result, tt.expectedSeconds)
			}
		})
	}
}

func TestClampInt(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		min      int
		max      int
		expected int
	}{
		{
			name:     "value within range",
			value:    50,
			min:      0,
			max:      100,
			expected: 50,
		},
		{
			name:     "value below min",
			value:    -10,
			min:      0,
			max:      100,
			expected: 0,
		},
		{
			name:     "value above max",
			value:    150,
			min:      0,
			max:      100,
			expected: 100,
		},
		{
			name:     "value equals min",
			value:    0,
			min:      0,
			max:      100,
			expected: 0,
		},
		{
			name:     "value equals max",
			value:    100,
			min:      0,
			max:      100,
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClampInt(tt.value, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("ClampInt(%d, %d, %d) = %d, want %d",
					tt.value, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestClampFloat64(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		min      float64
		max      float64
		expected float64
	}{
		{
			name:     "value within range",
			value:    0.5,
			min:      0.0,
			max:      1.0,
			expected: 0.5,
		},
		{
			name:     "value below min",
			value:    -0.5,
			min:      0.0,
			max:      1.0,
			expected: 0.0,
		},
		{
			name:     "value above max",
			value:    1.5,
			min:      0.0,
			max:      1.0,
			expected: 1.0,
		},
		{
			name:     "value equals min",
			value:    0.0,
			min:      0.0,
			max:      1.0,
			expected: 0.0,
		},
		{
			name:     "value equals max",
			value:    1.0,
			min:      0.0,
			max:      1.0,
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClampFloat64(tt.value, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("ClampFloat64(%.1f, %.1f, %.1f) = %.1f, want %.1f",
					tt.value, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestGetAudioDuration(t *testing.T) {
	// This test can only run on macOS
	if runtime.GOOS != "darwin" {
		t.Skip("GetAudioDuration only works on macOS")
	}

	// Test with non-existent file
	_, err := GetAudioDuration("/nonexistent/file.aiff")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}

	// Test with invalid platform check is covered by the function itself
	// We can't easily test the actual duration parsing without creating a real audio file
	// which would be too heavy for a unit test
}

func TestGetAudioDurationNonMacOS(t *testing.T) {
	// Save original GOOS
	originalGOOS := runtime.GOOS

	// This test verifies the error message when not on macOS
	// In reality, we can't change runtime.GOOS, but we document the expected behavior
	if originalGOOS != "darwin" {
		_, err := GetAudioDuration("/some/file.aiff")
		if err == nil {
			t.Error("Expected error on non-macOS platform, got nil")
		}
	}
}
