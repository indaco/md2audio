package env

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Load reads environment variables from a .env file and sets them in the process environment.
// It does NOT override existing environment variables.
// Returns the number of variables loaded.
func Load(filepath string) (int, error) {
	file, err := os.Open(filepath)
	if err != nil {
		// .env file is optional - not an error if it doesn't exist
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to open .env file: %w", err)
	}
	defer func() { _ = file.Close() }()

	loaded := 0
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return loaded, fmt.Errorf("invalid format at line %d: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		value = strings.Trim(value, `"'`)

		// Only set if not already in environment
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				return loaded, fmt.Errorf("failed to set env var %s: %w", key, err)
			}
			loaded++
		}
	}

	if err := scanner.Err(); err != nil {
		return loaded, fmt.Errorf("error reading .env file: %w", err)
	}

	return loaded, nil
}

// Get retrieves an environment variable value.
// It first checks the process environment, then falls back to the default value.
func Get(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// MustGet retrieves an environment variable value.
// It panics if the variable is not set.
func MustGet(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return value
}
