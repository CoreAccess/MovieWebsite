package tmdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	APIBaseURL   = "https://api.themoviedb.org/3"
	ImageBaseURL = "https://image.tmdb.org/t/p/w500"
)

type Client struct {
	token  string
	isV4   bool
	client *http.Client
}

func NewClient(token string) *Client {
	// TMDB v3 keys are 32 chars, v4 Access Tokens are much longer (JWT-like)
	isV4 := len(token) > 100
	return &Client{
		token:  token,
		isV4:   isV4,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) authenticateRequest(req *http.Request) {
	req.Header.Add("accept", "application/json")
	if c.isV4 {
		req.Header.Add("Authorization", "Bearer "+c.token)
	} else {
		q := req.URL.Query()
		q.Add("api_key", c.token)
		req.URL.RawQuery = q.Encode()
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

// TMDBMovieDetail is the full detail response for a single movie.
type TMDBMovieDetail struct {
	ID               int         `json:"id"`
	Title            string      `json:"title"`
	Overview         string      `json:"overview"`
	ReleaseDate      string      `json:"release_date"`
	PosterPath       string      `json:"poster_path"`
	VoteAverage      float64     `json:"vote_average"`
	Runtime          int         `json:"runtime"`
	Tagline          string      `json:"tagline"`
	Genres           []TMDBGenre `json:"genres"`
	Budget           int         `json:"budget"`
	Revenue          int         `json:"revenue"`
	OriginalLanguage string      `json:"original_language"`
	OriginCountry    []string    `json:"origin_country"`
}

// TMDBReleaseDatesResponse wraps the release_dates append response.
type TMDBReleaseDatesResponse struct {
	Results []struct {
		Iso31661     string `json:"iso_3166_1"`
		ReleaseDates []struct {
			Certification string `json:"certification"`
		} `json:"release_dates"`
	} `json:"results"`
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

// TMDBTVDetail is the full detail response for a single TV series.
type TMDBTVDetail struct {
	ID               int         `json:"id"`
	Name             string      `json:"name"`
	Overview         string      `json:"overview"`
	FirstAirDate     string      `json:"first_air_date"`
	LastAirDate      string      `json:"last_air_date"`
	PosterPath       string      `json:"poster_path"`
	VoteAverage      float64     `json:"vote_average"`
	NumberOfSeasons  int         `json:"number_of_seasons"`
	NumberOfEpisodes int         `json:"number_of_episodes"`
	Genres           []TMDBGenre `json:"genres"`
	OriginalLanguage string      `json:"original_language"`
	OriginCountry    []string    `json:"origin_country"`
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

type TMDBPersonDetail struct {
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	Biography          string  `json:"biography"`
	Birthday           string  `json:"birthday"`
	Deathday           string  `json:"deathday"`
	PlaceOfBirth       string  `json:"place_of_birth"`
	ProfilePath        string  `json:"profile_path"`
	KnownForDepartment string  `json:"known_for_department"`
	Popularity         float64 `json:"popularity"`
}

type TMDBSeasonDetail struct {
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	Overview     string        `json:"overview"`
	AirDate      string        `json:"air_date"`
	PosterPath   string        `json:"poster_path"`
	SeasonNumber int           `json:"season_number"`
	VoteAverage  float64       `json:"vote_average"`
	Episodes     []TMDBEpisode `json:"episodes"`
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

type TMDBWatchProvider struct {
	DisplayPriority int    `json:"display_priority"`
	LogoPath        string `json:"logo_path"`
	ProviderID      int    `json:"provider_id"`
	ProviderName    string `json:"provider_name"`
}

type TMDBWatchProviderCountryResult struct {
	Link     string              `json:"link"`
	Flatrate []TMDBWatchProvider `json:"flatrate"`
	Rent     []TMDBWatchProvider `json:"rent"`
	Buy      []TMDBWatchProvider `json:"buy"`
	Free     []TMDBWatchProvider `json:"free"`
	Ads      []TMDBWatchProvider `json:"ads"`
}

type TMDBWatchProvidersResponse struct {
	Results map[string]TMDBWatchProviderCountryResult `json:"results"`
}

func (c *Client) FetchMovieGenres() ([]TMDBGenre, error) {
	req, _ := http.NewRequest("GET", APIBaseURL+"/genre/movie/list?language=en", nil)
	c.authenticateRequest(req)

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
	req, _ := http.NewRequest("GET", APIBaseURL+"/genre/tv/list?language=en", nil)
	c.authenticateRequest(req)

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
	req, _ := http.NewRequest("GET", APIBaseURL+"/trending/movie/week?language=en-US", nil)
	c.authenticateRequest(req)

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
	req, _ := http.NewRequest("GET", APIBaseURL+"/trending/tv/week?language=en-US", nil)
	c.authenticateRequest(req)

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
	req, _ := http.NewRequest("GET", fmt.Sprintf(APIBaseURL+"/movie/%d/credits?language=en-US", movieID), nil)
	c.authenticateRequest(req)

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
	req, _ := http.NewRequest("GET", fmt.Sprintf(APIBaseURL+"/tv/%d/credits?language=en-US", seriesID), nil)
	c.authenticateRequest(req)

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
	req, _ := http.NewRequest("GET", fmt.Sprintf(APIBaseURL+"/tv/%d/season/%d?language=en-US", seriesID, seasonNumber), nil)
	c.authenticateRequest(req)

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

// DiscoverMovies fetches a page of movies sorted by release date descending (newest first).
func (c *Client) DiscoverMovies(page int) ([]TMDBMovie, error) {
	url := fmt.Sprintf("%s/discover/movie?sort_by=primary_release_date.desc&page=%d&language=en-US", APIBaseURL, page)
	req, _ := http.NewRequest("GET", url, nil)
	c.authenticateRequest(req)

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
	return data.Results, nil
}

// DiscoverTV fetches a page of TV shows sorted by first air date descending (newest first).
func (c *Client) DiscoverTV(page int) ([]TMDBTV, error) {
	url := fmt.Sprintf("%s/discover/tv?sort_by=first_air_date.desc&page=%d&language=en-US", APIBaseURL, page)
	req, _ := http.NewRequest("GET", url, nil)
	c.authenticateRequest(req)

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
	return data.Results, nil
}

// FetchMovieDetail retrieves full metadata for a single movie.
func (c *Client) FetchMovieDetail(movieID int) (TMDBMovieDetail, error) {
	url := fmt.Sprintf("%s/movie/%d?language=en-US", APIBaseURL, movieID)
	req, _ := http.NewRequest("GET", url, nil)
	c.authenticateRequest(req)

	res, err := c.client.Do(req)
	if err != nil {
		return TMDBMovieDetail{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return TMDBMovieDetail{}, fmt.Errorf("TMDB API error: %s", res.Status)
	}

	var detail TMDBMovieDetail
	if err := json.NewDecoder(res.Body).Decode(&detail); err != nil {
		return TMDBMovieDetail{}, err
	}
	return detail, nil
}

// FetchTVDetail retrieves full metadata for a single TV series.
func (c *Client) FetchTVDetail(seriesID int) (TMDBTVDetail, error) {
	url := fmt.Sprintf("%s/tv/%d?language=en-US", APIBaseURL, seriesID)
	req, _ := http.NewRequest("GET", url, nil)
	c.authenticateRequest(req)

	res, err := c.client.Do(req)
	if err != nil {
		return TMDBTVDetail{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return TMDBTVDetail{}, fmt.Errorf("TMDB API error: %s", res.Status)
	}

	var detail TMDBTVDetail
	if err := json.NewDecoder(res.Body).Decode(&detail); err != nil {
		return TMDBTVDetail{}, err
	}
	return detail, nil
}

// FetchMovieCertification returns the content rating for a movie in a specific country.
func (c *Client) FetchMovieCertification(movieID int, countryCode string) (string, error) {
	url := fmt.Sprintf("%s/movie/%d/release_dates", APIBaseURL, movieID)
	req, _ := http.NewRequest("GET", url, nil)
	c.authenticateRequest(req)

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var data TMDBReleaseDatesResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", err
	}

	for _, r := range data.Results {
		if r.Iso31661 == countryCode {
			for _, rd := range r.ReleaseDates {
				if rd.Certification != "" {
					return rd.Certification, nil
				}
			}
		}
	}
	return "", nil
}

// FetchPersonDetail retrieves full metadata for a single person.
func (c *Client) FetchPersonDetail(personID int) (TMDBPersonDetail, error) {
	url := fmt.Sprintf("%s/person/%d?language=en-US", APIBaseURL, personID)
	req, _ := http.NewRequest("GET", url, nil)
	c.authenticateRequest(req)

	res, err := c.client.Do(req)
	if err != nil {
		return TMDBPersonDetail{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return TMDBPersonDetail{}, fmt.Errorf("TMDB API error: %s", res.Status)
	}

	var detail TMDBPersonDetail
	if err := json.NewDecoder(res.Body).Decode(&detail); err != nil {
		return TMDBPersonDetail{}, err
	}
	return detail, nil
}

// FetchTVSeason retrieves full metadata for a single TV season.
func (c *Client) FetchTVSeason(seriesID, seasonNumber int) (TMDBSeasonDetail, error) {
	url := fmt.Sprintf("%s/tv/%d/season/%d?language=en-US", APIBaseURL, seriesID, seasonNumber)
	req, _ := http.NewRequest("GET", url, nil)
	c.authenticateRequest(req)

	res, err := c.client.Do(req)
	if err != nil {
		return TMDBSeasonDetail{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return TMDBSeasonDetail{}, fmt.Errorf("TMDB API error: %s", res.Status)
	}

	var detail TMDBSeasonDetail
	if err := json.NewDecoder(res.Body).Decode(&detail); err != nil {
		return TMDBSeasonDetail{}, err
	}
	return detail, nil
}

func (c *Client) FetchMovieWatchProviders(movieID int, countryCode string) (TMDBWatchProviderCountryResult, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf(APIBaseURL+"/movie/%d/watch/providers", movieID), nil)
	c.authenticateRequest(req)

	res, err := c.client.Do(req)
	if err != nil {
		return TMDBWatchProviderCountryResult{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return TMDBWatchProviderCountryResult{}, fmt.Errorf("TMDB API error: %s", res.Status)
	}

	var data TMDBWatchProvidersResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return TMDBWatchProviderCountryResult{}, err
	}

	result, ok := data.Results[countryCode]
	if !ok {
		return TMDBWatchProviderCountryResult{}, nil
	}
	return result, nil
}

func (c *Client) FetchTVWatchProviders(seriesID int, countryCode string) (TMDBWatchProviderCountryResult, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf(APIBaseURL+"/tv/%d/watch/providers", seriesID), nil)
	c.authenticateRequest(req)

	res, err := c.client.Do(req)
	if err != nil {
		return TMDBWatchProviderCountryResult{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return TMDBWatchProviderCountryResult{}, fmt.Errorf("TMDB API error: %s", res.Status)
	}

	var data TMDBWatchProvidersResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return TMDBWatchProviderCountryResult{}, err
	}

	result, ok := data.Results[countryCode]
	if !ok {
		return TMDBWatchProviderCountryResult{}, nil
	}
	return result, nil
}

type TMDBCountry struct {
	Iso31661    string `json:"iso_3166_1"`
	EnglishName string `json:"english_name"`
	NativeName  string `json:"native_name"`
}

// FetchCountries retrieves the list of configuration countries from TMDB.
func (c *Client) FetchCountries() ([]TMDBCountry, error) {
	url := fmt.Sprintf("%s/configuration/countries?language=en-US", APIBaseURL)
	req, _ := http.NewRequest("GET", url, nil)
	c.authenticateRequest(req)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API error: %s", res.Status)
	}

	var countries []TMDBCountry
	if err := json.NewDecoder(res.Body).Decode(&countries); err != nil {
		return nil, err
	}
	return countries, nil
}
