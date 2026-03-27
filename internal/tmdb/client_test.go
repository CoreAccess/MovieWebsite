package tmdb

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal string",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "Uppercase",
			input:    "ALL CAPS",
			expected: "all-caps",
		},
		{
			name:     "Punctuation",
			input:    "The Matrix: Reloaded!",
			expected: "the-matrix-reloaded",
		},
		{
			name:     "Multiple spaces and hyphens",
			input:    "A    B--C",
			expected: "a-b-c",
		},
		{
			name:     "Edge spaces and special",
			input:    "  @#% hello...  ",
			expected: "hello",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Special chars only",
			input:    "@@@!!!",
			expected: "",
		},
		{
			name:     "Already slugified",
			input:    "already-slugified-string",
			expected: "already-slugified-string",
		},
		{
			name:     "Numbers",
			input:    "Top 10 Movies of 2024",
			expected: "top-10-movies-of-2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			if result != tt.expected {
				t.Errorf("Slugify(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

