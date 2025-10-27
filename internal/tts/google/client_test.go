package google

import (
	"testing"
)

func TestDetermineVoiceType(t *testing.T) {
	tests := []struct {
		name      string
		voiceName string
		want      string
	}{
		{
			name:      "Neural2 voice",
			voiceName: "en-US-Neural2-F",
			want:      "Neural2",
		},
		{
			name:      "WaveNet voice (lowercase)",
			voiceName: "en-GB-Wavenet-A",
			want:      "WaveNet",
		},
		{
			name:      "WaveNet voice (uppercase)",
			voiceName: "en-US-WaveNet-B",
			want:      "WaveNet",
		},
		{
			name:      "Studio voice",
			voiceName: "en-US-Studio-M",
			want:      "Studio",
		},
		{
			name:      "Polyglot voice",
			voiceName: "en-US-Polyglot-1",
			want:      "Polyglot",
		},
		{
			name:      "Standard voice",
			voiceName: "en-US-Standard-A",
			want:      "Standard",
		},
		{
			name:      "Unknown voice defaults to Standard",
			voiceName: "en-US-Unknown-Voice",
			want:      "Standard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineVoiceType(tt.voiceName)
			if got != tt.want {
				t.Errorf("determineVoiceType(%q) = %q, want %q", tt.voiceName, got, tt.want)
			}
		})
	}
}

func TestCalculateSpeed(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		targetDuration float64
		wantMin        float64
		wantMax        float64
	}{
		{
			name:           "Normal speed for balanced text",
			text:           "This is a test sentence with about ten words in it.",
			targetDuration: 5.0,
			wantMin:        0.5,
			wantMax:        2.0,
		},
		{
			name:           "Fast speed for short target",
			text:           "This is a long sentence with many words that should require faster speech to fit in a very short duration.",
			targetDuration: 2.0,
			wantMin:        1.0,
			wantMax:        4.0,
		},
		{
			name:           "Slow speed for long target",
			text:           "Short text.",
			targetDuration: 10.0,
			wantMin:        0.25,
			wantMax:        1.0,
		},
		{
			name:           "Clamping at maximum speed",
			text:           "This is a very long piece of text with many words that would require extremely fast speech to fit into an impossibly short duration of just one second which is not realistic.",
			targetDuration: 0.5,
			wantMin:        3.0,
			wantMax:        4.0, // Should be clamped at 4.0
		},
		{
			name:           "Clamping at minimum speed",
			text:           "Word.",
			targetDuration: 30.0,
			wantMin:        0.25, // Should be clamped at 0.25
			wantMax:        0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateSpeed(tt.text, tt.targetDuration)

			// Check if within valid Google Cloud TTS range
			if got < 0.25 || got > 4.0 {
				t.Errorf("calculateSpeed() = %v, which is outside valid range [0.25, 4.0]", got)
			}

			// Check if within expected range for this test
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateSpeed(%q, %.1f) = %v, want between %.2f and %.2f",
					tt.text, tt.targetDuration, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestEnsureCorrectExtension(t *testing.T) {
	tests := []struct {
		name       string
		outputPath string
		format     string
		want       string
	}{
		{
			name:       "Correct MP3 extension",
			outputPath: "/path/to/file.mp3",
			format:     "mp3",
			want:       "/path/to/file.mp3",
		},
		{
			name:       "Incorrect extension, should change to MP3",
			outputPath: "/path/to/file.wav",
			format:     "mp3",
			want:       "/path/to/file.mp3",
		},
		{
			name:       "Correct WAV extension",
			outputPath: "/path/to/file.wav",
			format:     "wav",
			want:       "/path/to/file.wav",
		},
		{
			name:       "Incorrect extension, should change to WAV",
			outputPath: "/path/to/file.ogg",
			format:     "wav",
			want:       "/path/to/file.wav",
		},
		{
			name:       "Correct OGG extension",
			outputPath: "/path/to/file.ogg",
			format:     "ogg",
			want:       "/path/to/file.ogg",
		},
		{
			name:       "No extension, should add format",
			outputPath: "/path/to/file",
			format:     "mp3",
			want:       "/path/to/file.mp3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ensureCorrectExtension(tt.outputPath, tt.format)
			if got != tt.want {
				t.Errorf("ensureCorrectExtension(%q, %q) = %q, want %q",
					tt.outputPath, tt.format, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{
			name:   "Substring present",
			s:      "en-US-Neural2-F",
			substr: "Neural2",
			want:   true,
		},
		{
			name:   "Substring not present",
			s:      "en-US-WaveNet-A",
			substr: "Neural2",
			want:   false,
		},
		{
			name:   "Empty substring",
			s:      "test",
			substr: "",
			want:   true,
		},
		{
			name:   "Substring at start",
			s:      "Neural2Voice",
			substr: "Neural2",
			want:   true,
		},
		{
			name:   "Substring at end",
			s:      "VoiceNeural2",
			substr: "Neural2",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}
