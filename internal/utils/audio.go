// Package utils provides utility functions for audio processing and common operations.
// This file contains audio-specific utilities including duration measurement,
// word counting, WPM calculations, and clamping functions.
//
// Key features:
//   - Audio duration measurement (macOS afinfo)
//   - Word counting and WPM calculations
//   - Duration estimation utilities
//   - Value clamping functions
package utils

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// GetAudioDuration measures the duration of an audio file using macOS afinfo.
// Returns duration in seconds, or an error if the file cannot be read or parsed.
// This function is macOS-specific and requires the afinfo command.
func GetAudioDuration(audioPath string) (float64, error) {
	// Verify we're on macOS
	if runtime.GOOS != "darwin" {
		return 0, fmt.Errorf("audio duration measurement is only available on macOS")
	}

	cmd := exec.Command("afinfo", audioPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("afinfo command failed: %w", err)
	}

	// Parse duration from afinfo output
	// Looking for line like: "estimated duration: 5.123456 sec"
	re := regexp.MustCompile(`estimated duration:\s+([\d.]+)\s+sec`)
	matches := re.FindStringSubmatch(string(output))

	if len(matches) < 2 {
		return 0, fmt.Errorf("could not parse duration from afinfo output")
	}

	duration, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse duration value: %w", err)
	}

	return duration, nil
}

// CountWords counts the number of words in a text string.
// Words are defined as whitespace-separated tokens.
func CountWords(text string) int {
	return len(strings.Fields(text))
}

// CalculateWPM calculates words per minute given word count and duration.
// Duration should be in seconds. Returns 0 if duration is invalid.
func CalculateWPM(wordCount int, durationSeconds float64) float64 {
	if durationSeconds <= 0 {
		return 0
	}
	durationMinutes := durationSeconds / 60.0
	return float64(wordCount) / durationMinutes
}

// EstimateDuration estimates the duration in seconds for a given text at a specified WPM.
func EstimateDuration(text string, wordsPerMinute float64) float64 {
	if wordsPerMinute <= 0 {
		return 0
	}
	wordCount := CountWords(text)
	return (float64(wordCount) / wordsPerMinute) * 60.0
}

// ClampInt clamps an integer value between min and max.
func ClampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ClampFloat64 clamps a float64 value between min and max.
func ClampFloat64(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
