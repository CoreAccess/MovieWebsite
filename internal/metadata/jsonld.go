package metadata

import (
	"encoding/json"
	"fmt"
	"filmgap/internal/models"
)

// GenerateMovieJSONLD converts a models.Movie into a valid Schema.org/Movie JSON-LD string.
func GenerateMovieJSONLD(m models.Movie, baseDomain string) (string, error) {
	// Construct the foundational Map.
	payload := map[string]interface{}{
		"@context":      "https://schema.org",
		"@type":         "Movie",
		"name":          m.Name,
		"url":           fmt.Sprintf("%s/movie/%s", baseDomain, m.Slug),
		"image":         m.Image,
		"description":   m.Description,
		"datePublished": m.DatePublished,
	}

	if m.Duration > 0 {
		// ISO 8601 Duration format for Runtime in minutes
		payload["duration"] = fmt.Sprintf("PT%dM", m.Duration)
	}
	if m.AggregateRating > 0 {
		payload["aggregateRating"] = map[string]interface{}{
			"@type":       "AggregateRating",
			"ratingValue": m.AggregateRating,
			"bestRating":  "10", // filmgap scale
		}
	}
	if len(m.Genres) > 0 {
		payload["genre"] = m.Genres
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GenerateTVSeriesJSONLD converts a models.TVSeries into a valid Schema.org/TVSeries JSON-LD string.
func GenerateTVSeriesJSONLD(s models.TVSeries, baseDomain string) (string, error) {
	payload := map[string]interface{}{
		"@context":      "https://schema.org",
		"@type":         "TVSeries",
		"name":          s.Name,
		"url":           fmt.Sprintf("%s/show/%s", baseDomain, s.Slug),
		"image":         s.Image,
		"description":   s.Description,
		"startDate":     s.DatePublished,
		"endDate":       s.EndDate,
	}

	if s.NumberOfSeasons > 0 {
		payload["numberOfSeasons"] = s.NumberOfSeasons
	}
	if s.NumberOfEpisodes > 0 {
		payload["numberOfEpisodes"] = s.NumberOfEpisodes
	}
	if s.AggregateRating > 0 {
		payload["aggregateRating"] = map[string]interface{}{
			"@type":       "AggregateRating",
			"ratingValue": s.AggregateRating,
			"bestRating":  "10",
		}
	}
	if len(s.Genres) > 0 {
		payload["genre"] = s.Genres
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GeneratePersonJSONLD converts a models.Person into a valid Schema.org/Person JSON-LD string.
func GeneratePersonJSONLD(p models.Person, baseDomain string) (string, error) {
	payload := map[string]interface{}{
		"@context":    "https://schema.org",
		"@type":       "Person",
		"name":        p.Name,
		"url":         fmt.Sprintf("%s/person/%s", baseDomain, p.Slug),
		"image":       p.Image,
		"description": p.Biography,
		"birthDate":   p.BirthDate,
		"deathDate":   p.Deathday,
		"birthPlace":  p.BirthPlace,
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
// GenerateBlogPostJSONLD converts a models.BlogPost into a valid Schema.org/BlogPosting JSON-LD string.
func GenerateBlogPostJSONLD(p models.BlogPost, baseDomain string) (string, error) {
	payload := map[string]interface{}{
		"@context":      "https://schema.org",
		"@type":         "BlogPosting",
		"headline":      p.Title,
		"url":           fmt.Sprintf("%s/blog/%s", baseDomain, p.Slug),
		"image":         p.Image,
		"datePublished": p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		"author": map[string]interface{}{
			"@type": "Person",
			"name":  p.Author,
		},
		"articleBody": p.Content,
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
