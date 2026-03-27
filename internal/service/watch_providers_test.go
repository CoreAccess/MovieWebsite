package service

import (
	"filmgap/internal/models"
	"testing"
	"time"
)

func TestNormalizeProviderCountry(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty defaults to US", input: "", expected: "US"},
		{name: "single country preserved", input: "ca", expected: "CA"},
		{name: "comma list uses first entry", input: "us,ca", expected: "US"},
		{name: "invalid falls back to US", input: "USA", expected: "US"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeProviderCountry(tt.input); got != tt.expected {
				t.Fatalf("normalizeProviderCountry(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGroupWatchProviders(t *testing.T) {
	now := time.Now()
	providers := []models.WatchProviderOption{
		{ProviderType: "buy", ProviderID: 4, ProviderName: "Apple TV", DisplayPriority: 2, UpdatedAt: now},
		{ProviderType: "subscription", ProviderID: 2, ProviderName: "Netflix", DisplayPriority: 3, UpdatedAt: now},
		{ProviderType: "subscription", ProviderID: 1, ProviderName: "Max", DisplayPriority: 1, UpdatedAt: now},
		{ProviderType: "subscription", ProviderID: 1, ProviderName: "Max", DisplayPriority: 1, UpdatedAt: now},
		{ProviderType: "rent", ProviderID: 3, ProviderName: "Prime Video", DisplayPriority: 1, UpdatedAt: now},
	}

	grouped := groupWatchProviders(providers)
	if len(grouped) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(grouped))
	}

	if grouped[0].Key != "subscription" {
		t.Fatalf("expected first group to be subscription, got %s", grouped[0].Key)
	}

	if len(grouped[0].Providers) != 2 {
		t.Fatalf("expected duplicate subscription providers to be deduped, got %d providers", len(grouped[0].Providers))
	}

	if grouped[0].Providers[0].ProviderName != "Max" {
		t.Fatalf("expected providers to be sorted by display priority, got %s first", grouped[0].Providers[0].ProviderName)
	}
}
