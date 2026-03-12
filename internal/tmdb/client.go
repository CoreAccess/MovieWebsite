package tmdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	apiKey string
	client *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Basic Slugifier
func Slugify(s string) string {
	s = strings.ToLower(s)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// TMDB API Structs
type TrendingMoviesResponse struct {
	Results []TMDBMovie `json:"results"`
}

type TMDBMovie struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	Overview         string  `json:"overview"`
	ReleaseDate      string  `json:"release_date"`
	PosterPath       string  `json:"poster_path"`
	VoteAverage      float64 `json:"vote_average"`
	OriginalLanguage string  `json:"original_language"`
	GenreIDs         []int   `json:"genre_ids"`
}

type TrendingTVResponse struct {
	Results []TMDBTV `json:"results"`
}

type TMDBTV struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	Overview         string  `json:"overview"`
	FirstAirDate     string  `json:"first_air_date"`
	PosterPath       string  `json:"poster_path"`
	VoteAverage      float64 `json:"vote_average"`
	OriginalLanguage string  `json:"original_language"`
	GenreIDs         []int   `json:"genre_ids"`
}

type CreditsResponse struct {
	Cast []TMDBCast `json:"cast"`
	Crew []TMDBCrew `json:"crew"`
}

type TMDBCast struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Character   string `json:"character"`
	ProfilePath string `json:"profile_path"`
	Order       int    `json:"order"`
	Gender      int    `json:"gender"` // 1: Female, 2: Male
}

type TMDBCrew struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Job         string `json:"job"`
	Department  string `json:"department"`
	ProfilePath string `json:"profile_path"`
}

type TMDBSeasonDetail struct {
	Episodes []TMDBEpisode `json:"episodes"`
}

type TMDBEpisode struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Overview      string `json:"overview"`
	AirDate       string `json:"air_date"`
	StillPath     string `json:"still_path"`
	EpisodeNumber int    `json:"episode_number"`
	SeasonNumber  int    `json:"season_number"`
	Runtime       int    `json:"runtime"`
}

type TMDBGenreResponse struct {
	Genres []TMDBGenre `json:"genres"`
}

type TMDBGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (c *Client) FetchMovieGenres() ([]TMDBGenre, error) {
	req, _ := http.NewRequest("GET", "https://api.themoviedb.org/3/genre/movie/list?language=en", nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.apiKey)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API error: %s", res.Status)
	}

	var data TMDBGenreResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Genres, nil
}

func (c *Client) FetchTVGenres() ([]TMDBGenre, error) {
	req, _ := http.NewRequest("GET", "https://api.themoviedb.org/3/genre/tv/list?language=en", nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.apiKey)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API error: %s", res.Status)
	}

	var data TMDBGenreResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Genres, nil
}

func (c *Client) FetchTrendingMovies() ([]TMDBMovie, error) {
	req, _ := http.NewRequest("GET", "https://api.themoviedb.org/3/trending/movie/week?language=en-US", nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.apiKey)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API error: %s", res.Status)
	}

	var data TrendingMoviesResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	// Limit to top 10
	limit := 10
	if len(data.Results) < limit {
		limit = len(data.Results)
	}
	return data.Results[:limit], nil
}

func (c *Client) FetchTrendingShows() ([]TMDBTV, error) {
	req, _ := http.NewRequest("GET", "https://api.themoviedb.org/3/trending/tv/week?language=en-US", nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.apiKey)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API error: %s", res.Status)
	}

	var data TrendingTVResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	// Limit to top 10
	limit := 10
	if len(data.Results) < limit {
		limit = len(data.Results)
	}
	return data.Results[:limit], nil
}

func (c *Client) FetchMovieCredits(movieID int) (CreditsResponse, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.themoviedb.org/3/movie/%d/credits?language=en-US", movieID), nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.apiKey)

	res, err := c.client.Do(req)
	if err != nil {
		return CreditsResponse{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return CreditsResponse{}, fmt.Errorf("TMDB API error: %s", res.Status)
	}

	var data CreditsResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return CreditsResponse{}, err
	}

	// Limit to top 10 cast members
	limit := 10
	if len(data.Cast) > limit {
		data.Cast = data.Cast[:limit]
	}
	return data, nil
}

func (c *Client) FetchTVCredits(seriesID int) (CreditsResponse, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.themoviedb.org/3/tv/%d/credits?language=en-US", seriesID), nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.apiKey)

	res, err := c.client.Do(req)
	if err != nil {
		return CreditsResponse{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return CreditsResponse{}, fmt.Errorf("TMDB API error: %s", res.Status)
	}

	var data CreditsResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return CreditsResponse{}, err
	}

	// Limit to top 10 cast members
	limit := 10
	if len(data.Cast) > limit {
		data.Cast = data.Cast[:limit]
	}
	return data, nil
}

func (c *Client) FetchTVSeasonEpisodes(seriesID int, seasonNumber int) ([]TMDBEpisode, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.themoviedb.org/3/tv/%d/season/%d?language=en-US", seriesID, seasonNumber), nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.apiKey)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API error: %s", res.Status)
	}

	var data TMDBSeasonDetail
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	// Limit to top 10 episodes max
	limit := 10
	if len(data.Episodes) < limit {
		limit = len(data.Episodes)
	}
	return data.Episodes[:limit], nil
}
