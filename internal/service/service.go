package service

import (
	"movieweb/internal/models"
	"movieweb/internal/repository"
	"movieweb/internal/metadata"
)

// AppService orchestrates complex business logic, abstracting
// repository calls (and later Vector DB calls) away from HTTP handlers.
type AppService struct {
	Repo repository.DatabaseRepo
	// Future: VectorRepo repository.VectorRepo
}

// NewAppService creates a new configured service layer.
func NewAppService(repo repository.DatabaseRepo) *AppService {
	return &AppService{
		Repo: repo,
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
func (s *AppService) GetShowDetail(id int, baseDomain string) (*models.TVSeries, []models.CastMember, []models.CrewMember, string, error) {
	show, err := s.Repo.GetShowByID(id)
	if err != nil {
		return nil, nil, nil, "", err
	}

	cast, err := s.Repo.GetCastForMedia(show.ID)
	if err != nil {
		return nil, nil, nil, "", err
	}

	crew, err := s.Repo.GetCrewForMedia(show.ID)
	if err != nil {
		return nil, nil, nil, "", err
	}

	jsonld, err := metadata.GenerateTVSeriesJSONLD(*show, baseDomain)
	if err != nil {
		return nil, nil, nil, "", err
	}

	return show, cast, crew, jsonld, nil
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
