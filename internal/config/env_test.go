package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnv(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected map[string]string
	}{
		{
			name: "Basic key-value pairs",
			content: `
KEY1=VALUE1
KEY2=VALUE2
`,
			expected: map[string]string{
				"KEY1": "VALUE1",
				"KEY2": "VALUE2",
			},
		},
		{
			name: "Comments and empty lines",
			content: `
# This is a comment
KEY1=VALUE1

# Another comment
KEY2=VALUE2
`,
			expected: map[string]string{
				"KEY1": "VALUE1",
				"KEY2": "VALUE2",
			},
		},
		{
			name: "Whitespace handling",
			content: `
  KEY1  =  VALUE1
KEY2=VALUE2
   KEY3=VALUE3
`,
			expected: map[string]string{
				"KEY1": "VALUE1",
				"KEY2": "VALUE2",
				"KEY3": "VALUE3",
			},
		},
		{
			name: "Quoted values",
			content: `
KEY1="VALUE1"
KEY2='VALUE2'
KEY3="VALUE WITH SPACES"
KEY4='VALUE WITH "QUOTES"'
`,
			expected: map[string]string{
				"KEY1": "VALUE1",
				"KEY2": "VALUE2",
				"KEY3": "VALUE WITH SPACES",
				"KEY4": `VALUE WITH "QUOTES"`,
			},
		},
		{
			name: "Values with equals signs",
			content: `
KEY1=VALUE=WITH=EQUALS
KEY2=VALUE=
`,
			expected: map[string]string{
				"KEY1": "VALUE=WITH=EQUALS",
				"KEY2": "VALUE=",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, ".env")

			if err := os.WriteFile(envFile, []byte(tc.content), 0644); err != nil {
				t.Fatalf("Failed to create temp env file: %v", err)
			}

			// Clear environment variables before each test
			for k := range tc.expected {
				os.Unsetenv(k)
			}

			if err := LoadEnv(envFile); err != nil {
				t.Fatalf("LoadEnv failed: %v", err)
			}

			for k, expectedVal := range tc.expected {
				actualVal := os.Getenv(k)
				if actualVal != expectedVal {
					t.Errorf("Expected %s=%s, got %s", k, expectedVal, actualVal)
				}
				os.Unsetenv(k)
			}
		})
	}
}

func TestLoadEnv_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "non_existent_file.env")
	err := LoadEnv(nonExistentFile)
	if err != nil {
		t.Errorf("LoadEnv should not return error for non-existent file, got: %v", err)
	}
}

func TestLoadEnv_MalformedLine(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	const key = "MALFORMED_LINE_WITHOUT_EQUALS"
	content := key

	os.Unsetenv(key)
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp env file: %v", err)
	}

	if err := LoadEnv(envFile); err != nil {
		t.Fatalf("LoadEnv should not return error for malformed line: %v", err)
	}

	if val := os.Getenv(key); val != "" {
		t.Errorf("Expected environment variable %s to be empty, got %q", key, val)
	}
}
