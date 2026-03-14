package service

import (
	"fmt"
	"movieweb/internal/metadata"
	"movieweb/internal/models"
	"movieweb/internal/repository"
	"sync"
	"time"
)

type cacheItem struct {
	value      any
	expiration time.Time
}

// AppService orchestrates complex business logic, abstracting
// repository calls (and later Vector DB calls) away from HTTP handlers.
type AppService struct {
	Repo repository.DatabaseRepo
	// Future: VectorRepo repository.VectorRepo
	cache map[string]cacheItem
	mu    sync.RWMutex
}

// NewAppService creates a new configured service layer.
func NewAppService(repo repository.DatabaseRepo) *AppService {
	return &AppService{
		Repo:  repo,
		cache: make(map[string]cacheItem),
	}
}

// GetMovieDetail fetches a movie, its cast, and its generated JSON-LD.
// This ensures handlers do not manually stitch these concepts together.
func (s *AppService) GetMovieDetail(id int, baseDomain string) (*models.Movie, []models.CastMember, []models.CrewMember, string, error) {
	movie, err := s.Repo.GetMovieByID(id)
	if err != nil {
		return nil, nil, nil, "", err
	}

	cast, err := s.Repo.GetCastForMedia(movie.ID)
	if err != nil {
		return nil, nil, nil, "", err
	}

	crew, err := s.Repo.GetCrewForMedia(movie.ID)
	if err != nil {
		return nil, nil, nil, "", err
	}

	// Generate the Schema.org payload for AI Agents and SEO
	jsonld, err := metadata.GenerateMovieJSONLD(*movie, baseDomain)
	if err != nil {
		return nil, nil, nil, "", err
	}

	return movie, cast, crew, jsonld, nil
}

// GetShowDetail fetches a TV show, its cast, and its generated JSON-LD.
func (s *AppService) GetShowDetail(id int, baseDomain string) (*models.TVSeries, []models.CastMember, []models.CrewMember, []models.TVEpisode, string, error) {
	show, err := s.Repo.GetShowByID(id)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	cast, err := s.Repo.GetCastForMedia(show.ID)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	crew, err := s.Repo.GetCrewForMedia(show.ID)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	jsonld, err := metadata.GenerateTVSeriesJSONLD(*show, baseDomain)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	episodes, err := s.Repo.GetTVEpisodes(id)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	return show, cast, crew, episodes, jsonld, nil
}

// GetPersonDetail fetches a person, their credited works, and generated JSON-LD.
func (s *AppService) GetPersonDetail(id int, baseDomain string) (*models.Person, []models.Movie, []models.TVSeries, string, error) {
	person, err := s.Repo.GetPersonByID(id)
	if err != nil {
		return nil, nil, nil, "", err
	}

	movies, err := s.Repo.GetPersonMovies(id)
	if err != nil {
		return nil, nil, nil, "", err
	}

	shows, err := s.Repo.GetPersonShows(id)
	if err != nil {
		return nil, nil, nil, "", err
	}

	jsonld, err := metadata.GeneratePersonJSONLD(*person, baseDomain)
	if err != nil {
		return nil, nil, nil, "", err
	}

	return person, movies, shows, jsonld, nil
}

// Listing Operations
func (s *AppService) GetPopularMovies(limit int) ([]models.Movie, error) {
	return s.Repo.GetPopularMovies(limit)
}

func (s *AppService) GetUpcomingMovies(limit int) ([]models.Movie, error) {
	return s.Repo.GetUpcomingMovies(limit)
}

func (s *AppService) GetPopularShows(limit int) ([]models.TVSeries, error) {
	return s.Repo.GetPopularShows(limit)
}

func (s *AppService) GetNewShows(limit int) ([]models.TVSeries, error) {
	return s.Repo.GetNewShows(limit)
}

func (s *AppService) getFromCache(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.cache[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(item.expiration) {
		return nil, false
	}

	return item.value, true
}

func (s *AppService) setToCache(key string, value any, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}

func (s *AppService) GetAllMovies(limit int, offset int, sort string) ([]models.Movie, error) {
	cacheKey := fmt.Sprintf("movies:%d:%d:%s", limit, offset, sort)
	if val, ok := s.getFromCache(cacheKey); ok {
		return val.([]models.Movie), nil
	}

	movies, err := s.Repo.GetAllMovies(limit, offset, sort)
	if err == nil {
		s.setToCache(cacheKey, movies, 5*time.Minute)
	}
	return movies, err
}

func (s *AppService) GetAllShows(limit int, offset int, sort string) ([]models.TVSeries, error) {
	cacheKey := fmt.Sprintf("shows:%d:%d:%s", limit, offset, sort)
	if val, ok := s.getFromCache(cacheKey); ok {
		return val.([]models.TVSeries), nil
	}

	shows, err := s.Repo.GetAllShows(limit, offset, sort)
	if err == nil {
		s.setToCache(cacheKey, shows, 5*time.Minute)
	}
	return shows, err
}

func (s *AppService) GetAllPeople(limit int, offset int, sort string) ([]models.Person, error) {
	return s.Repo.GetAllPeople(limit, offset, sort)
}

func (s *AppService) GetUserWatchlist(userID int) ([]models.Movie, []models.TVSeries, error) {
	return s.Repo.GetUserWatchlist(userID)
}

func (s *AppService) GetUserWatchlists(userID int) ([]models.Watchlist, error) {
	return s.Repo.GetUserWatchlists(userID)
}

func (s *AppService) CreateWatchlist(userID int, name, description string) error {
	return s.Repo.CreateWatchlist(userID, name, description)
}

func (s *AppService) AddToWatchlist(watchlistID int, mediaType string, mediaID int) error {
	return s.Repo.AddToWatchlist(watchlistID, mediaType, mediaID)
}

// User & Session Operations
func (s *AppService) CreateUser(username, email, hash string) error {
	return s.Repo.CreateUser(username, email, hash)
}

func (s *AppService) GetUserByEmail(email string) (models.User, error) {
	return s.Repo.GetUserByEmail(email)
}

func (s *AppService) GetUserByID(id int) (models.User, error) {
	return s.Repo.GetUserByID(id)
}

func (s *AppService) GetAllUsers(limit int, offset int) ([]models.User, error) {
	cacheKey := fmt.Sprintf("users:%d:%d", limit, offset)
	if val, ok := s.getFromCache(cacheKey); ok {
		return val.([]models.User), nil
	}

	users, err := s.Repo.GetAllUsers(limit, offset)
	if err == nil {
		s.setToCache(cacheKey, users, 5*time.Minute)
	}
	return users, err
}

func (s *AppService) UpdateUserProfile(userID int, email string, avatar string) error {
	return s.Repo.UpdateUserProfile(userID, email, avatar)
}

func (s *AppService) CreateSession(session models.Session) error {
	return s.Repo.CreateSession(session)
}

func (s *AppService) GetSession(id string) (models.Session, error) {
	return s.Repo.GetSession(id)
}

func (s *AppService) DeleteSession(id string) error {
	return s.Repo.DeleteSession(id)
}

// Admin & Moderation Operations
func (s *AppService) GetAdminMetrics() (userCount, mediaCount int, err error) {
	return s.Repo.GetAdminMetrics()
}

// Search Operations
func (s *AppService) SearchMovies(q string, limit int, offset int) ([]models.Movie, error) {
	return s.Repo.SearchMovies(q, limit, offset)
}

func (s *AppService) SearchShows(q string, limit int, offset int) ([]models.TVSeries, error) {
	return s.Repo.SearchShows(q, limit, offset)
}

func (s *AppService) SearchPeople(q string, limit int, offset int) ([]models.Person, error) {
	return s.Repo.SearchPeople(q, limit, offset)
}
