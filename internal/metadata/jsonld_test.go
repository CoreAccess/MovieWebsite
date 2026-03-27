package metadata

import (
	"encoding/json"
	"filmgap/internal/models"
	"strings"
	"testing"
)

func TestGenerateMovieJSONLD(t *testing.T) {
	baseDomain := "https://example.com"
	tests := []struct {
		name    string
		movie   models.Movie
		wantErr bool
		checks  []string
	}{
		{
			name: "Full Movie Data",
			movie: models.Movie{
				Media: models.Media{
					Name:            "Inception",
					Slug:            "inception",
					Image:           "https://example.com/image.jpg",
					Description:     "A dream within a dream",
					DatePublished:   "2010-07-16",
					AggregateRating: 8.8,
				},
				Duration: 148,
				Genres: []models.Genre{
					{Name: "Action"},
					{Name: "Sci-Fi"},
				},
			},
			wantErr: false,
			checks: []string{
				`"@type":"Movie"`,
				`"name":"Inception"`,
				`"url":"https://example.com/movie/inception"`,
				`"duration":"PT148M"`,
				`"ratingValue":8.8`,
				`"genre"`,
			},
		},
		{
			name: "Minimal Movie Data",
			movie: models.Movie{
				Media: models.Media{
					Name: "Minimal",
					Slug: "minimal",
				},
			},
			wantErr: false,
			checks: []string{
				`"@type":"Movie"`,
				`"name":"Minimal"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateMovieJSONLD(tt.movie, baseDomain)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateMovieJSONLD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for _, check := range tt.checks {
					if !strings.Contains(got, check) {
						t.Errorf("GenerateMovieJSONLD() output missing expected string: %s\nGot: %s", check, got)
					}
				}
				// Verify it's valid JSON
				var js map[string]interface{}
				if err := json.Unmarshal([]byte(got), &js); err != nil {
					t.Errorf("GenerateMovieJSONLD() output is not valid JSON: %v", err)
				}
			}
		})
	}
}

func TestGenerateTVSeriesJSONLD(t *testing.T) {
	baseDomain := "https://example.com"
	tests := []struct {
		name    string
		series  models.TVSeries
		wantErr bool
		checks  []string
	}{
		{
			name: "Full TV Series Data",
			series: models.TVSeries{
				Media: models.Media{
					Name:            "Breaking Bad",
					Slug:            "breaking-bad",
					Image:           "https://example.com/bb.jpg",
					Description:     "A high school chemistry teacher...",
					DatePublished:   "2008-01-20",
					AggregateRating: 9.5,
				},
				EndDate:          "2013-09-29",
				NumberOfSeasons:  5,
				NumberOfEpisodes: 62,
				Genres: []models.Genre{
					{Name: "Crime"},
					{Name: "Drama"},
				},
			},
			wantErr: false,
			checks: []string{
				`"@type":"TVSeries"`,
				`"name":"Breaking Bad"`,
				`"url":"https://example.com/show/breaking-bad"`,
				`"startDate":"2008-01-20"`,
				`"endDate":"2013-09-29"`,
				`"numberOfSeasons":5`,
				`"numberOfEpisodes":62`,
				`"ratingValue":9.5`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateTVSeriesJSONLD(tt.series, baseDomain)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateTVSeriesJSONLD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for _, check := range tt.checks {
					if !strings.Contains(got, check) {
						t.Errorf("GenerateTVSeriesJSONLD() output missing expected string: %s\nGot: %s", check, got)
					}
				}
				var js map[string]interface{}
				if err := json.Unmarshal([]byte(got), &js); err != nil {
					t.Errorf("GenerateTVSeriesJSONLD() output is not valid JSON: %v", err)
				}
			}
		})
	}
}

func TestGeneratePersonJSONLD(t *testing.T) {
	baseDomain := "https://example.com"
	tests := []struct {
		name    string
		person  models.Person
		wantErr bool
		checks  []string
	}{
		{
			name: "Full Person Data",
			person: models.Person{
				Name:       "Christopher Nolan",
				Slug:       "christopher-nolan",
				Image:      "https://example.com/nolan.jpg",
				Biography:  "Acclaimed director...",
				BirthDate:  "1970-07-30",
				BirthPlace: "London, England",
			},
			wantErr: false,
			checks: []string{
				`"@type":"Person"`,
				`"name":"Christopher Nolan"`,
				`"url":"https://example.com/person/christopher-nolan"`,
				`"birthDate":"1970-07-30"`,
				`"birthPlace":"London, England"`,
				`"description":"Acclaimed director..."`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GeneratePersonJSONLD(tt.person, baseDomain)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePersonJSONLD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for _, check := range tt.checks {
					if !strings.Contains(got, check) {
						t.Errorf("GeneratePersonJSONLD() output missing expected string: %s\nGot: %s", check, got)
					}
				}
				var js map[string]interface{}
				if err := json.Unmarshal([]byte(got), &js); err != nil {
					t.Errorf("GeneratePersonJSONLD() output is not valid JSON: %v", err)
				}
			}
		})
	}
}

