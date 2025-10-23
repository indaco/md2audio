package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_BasicKeyValuePairs(t *testing.T) {
	envContent := `API_KEY=test123
DATABASE_URL=postgres://localhost`

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	loaded, err := Load(envFile)
	defer cleanupEnvVars(t, "API_KEY", "DATABASE_URL")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if loaded != 2 {
		t.Errorf("Expected 2 variables loaded, got %d", loaded)
	}

	assertEnvVar(t, "API_KEY", "test123")
	assertEnvVar(t, "DATABASE_URL", "postgres://localhost")
}

func TestLoad_WithCommentsAndEmptyLines(t *testing.T) {
	envContent := `# This is a comment
API_KEY=test123

# Another comment
DATABASE_URL=postgres://localhost
`

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	loaded, err := Load(envFile)
	defer cleanupEnvVars(t, "API_KEY", "DATABASE_URL")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if loaded != 2 {
		t.Errorf("Expected 2 variables loaded, got %d", loaded)
	}
}

func TestLoad_WithQuotedValues(t *testing.T) {
	envContent := `API_KEY="test 123"
DATABASE_URL='postgres://localhost'`

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	loaded, err := Load(envFile)
	defer cleanupEnvVars(t, "API_KEY", "DATABASE_URL")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if loaded != 2 {
		t.Errorf("Expected 2 variables loaded, got %d", loaded)
	}

	// Verify quoted values are unquoted
	assertEnvVar(t, "API_KEY", "test 123")
}

func TestLoad_DoesNotOverrideExistingEnvVars(t *testing.T) {
	envContent := `API_KEY=from_file
NEW_KEY=new_value`

	// Set up existing environment variable
	_ = os.Setenv("API_KEY", "existing_value")
	defer func() { _ = os.Unsetenv("API_KEY") }()

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	loaded, err := Load(envFile)
	defer cleanupEnvVars(t, "NEW_KEY")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Only NEW_KEY should be loaded
	if loaded != 1 {
		t.Errorf("Expected 1 variable loaded, got %d", loaded)
	}

	// Verify existing env vars are not overridden
	assertEnvVar(t, "API_KEY", "existing_value")
	assertEnvVar(t, "NEW_KEY", "new_value")
}

func TestLoad_HandlesSpacesAroundEquals(t *testing.T) {
	envContent := `API_KEY = test123
DATABASE_URL= postgres://localhost`

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	loaded, err := Load(envFile)
	defer cleanupEnvVars(t, "API_KEY", "DATABASE_URL")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if loaded != 2 {
		t.Errorf("Expected 2 variables loaded, got %d", loaded)
	}
}

func TestLoad_InvalidFormat(t *testing.T) {
	envContent := `INVALID LINE WITHOUT EQUALS`

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	_, err := Load(envFile)

	if err == nil {
		t.Error("Expected error but got nil")
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "nonexistent.env")

	loaded, err := Load(nonExistent)

	if err != nil {
		t.Errorf("Load should not error on non-existent file, got: %v", err)
	}

	if loaded != 0 {
		t.Errorf("Expected 0 variables loaded, got %d", loaded)
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "returns env value when set",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "from_env",
			expected:     "from_env",
		},
		{
			name:         "returns default when env not set",
			key:          "MISSING_KEY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				_ = os.Setenv(tt.key, tt.envValue)
				defer func() { _ = os.Unsetenv(tt.key) }()
			}

			got := Get(tt.key, tt.defaultValue)
			if got != tt.expected {
				t.Errorf("Get(%q, %q) = %q, want %q", tt.key, tt.defaultValue, got, tt.expected)
			}
		})
	}
}

func TestMustGet(t *testing.T) {
	t.Run("returns value when set", func(t *testing.T) {
		_ = os.Setenv("TEST_KEY", "test_value")
		defer func() { _ = os.Unsetenv("TEST_KEY") }()

		got := MustGet("TEST_KEY")
		if got != "test_value" {
			t.Errorf("MustGet() = %q, want %q", got, "test_value")
		}
	})

	t.Run("panics when not set", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustGet should panic when env var not set")
			}
		}()

		MustGet("MISSING_KEY")
	})
}

// Helper functions

// cleanupEnvVars unsets the given environment variables
func cleanupEnvVars(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		_ = os.Unsetenv(key)
	}
}

// assertEnvVar checks that an environment variable has the expected value
func assertEnvVar(t *testing.T, key, expected string) {
	t.Helper()
	if got := os.Getenv(key); got != expected {
		t.Errorf("%s = %q, want %q", key, got, expected)
	}
}
