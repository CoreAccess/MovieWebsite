package dbrepo

import (
	"database/sql"
	"log"


	"movieweb/internal/models"

	_ "modernc.org/sqlite"
)

// SqliteDBRepo implements the DatabaseRepo interface for SQLite
type SqliteDBRepo struct {
	DB *sql.DB
}

// Ensure interface compliance at compile time
//var _ repository.DatabaseRepo = (*SqliteDBRepo)(nil) // Uncomment once fully implemented

func (m *SqliteDBRepo) Connection() *sql.DB {
	return m.DB
}

func (m *SqliteDBRepo) InitDB(dataSourceName string, tmdbAPIKey string) (*sql.DB, error) {
	var err error
	m.DB, err = sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = m.DB.Ping(); err != nil {
		return nil, err
	}

	log.Println("Connected to SQLite Database")
	return m.DB, nil
}

// Methods to be filled with proper queries mapping to Phase 1's supertype structure:
// SearchMedia, GetMediaByID, GetMovieBySlug, GetShowBySlug, GetPersonBySlug, GetCastForMedia, GetCrewForMedia
// InsertMovie, InsertShow, InsertPerson

// The actual SQLite logic will live here.
// By migrating functions out of internal/database/*.go into receiver methods
// on m *SqliteDBRepo, we isolate the HTTP handlers from global SQL state.
// We will do this carefully in the service layer migration phase.

func (m *SqliteDBRepo) CreateUser(username, email, hash string) error { return nil }
func (m *SqliteDBRepo) GetUserByEmail(email string) (models.User, error) { return models.User{}, nil }
func (m *SqliteDBRepo) GetUserByID(id int) (models.User, error) { return models.User{}, nil }
func (m *SqliteDBRepo) UpdateUserProfile(userID int, email string, avatar string) error { return nil }
func (m *SqliteDBRepo) SearchMedia(query string, limit, offset int) ([]models.Media, error) { return nil, nil }
func (m *SqliteDBRepo) GetMediaByID(id int) (*models.Media, error) { return nil, nil }
func (m *SqliteDBRepo) GetAllMovies(limit int, offset int, sort string) ([]models.Movie, error) { return nil, nil }
func (m *SqliteDBRepo) GetPopularMovies(limit int) ([]models.Movie, error) { return nil, nil }
func (m *SqliteDBRepo) GetUpcomingMovies(limit int) ([]models.Movie, error) { return nil, nil }
func (m *SqliteDBRepo) GetMovieByID(id int) (*models.Movie, error) { return nil, nil }
func (m *SqliteDBRepo) GetAllShows(limit int, offset int, sort string) ([]models.TVSeries, error) { return nil, nil }
func (m *SqliteDBRepo) GetPopularShows(limit int) ([]models.TVSeries, error) { return nil, nil }
func (m *SqliteDBRepo) GetNewShows(limit int) ([]models.TVSeries, error) { return nil, nil }
func (m *SqliteDBRepo) GetShowByID(id int) (*models.TVSeries, error) { return nil, nil }
func (m *SqliteDBRepo) GetAllPeople(limit int, offset int, sort string) ([]models.Person, error) { return nil, nil }
func (m *SqliteDBRepo) GetPersonByID(id int) (*models.Person, error) { return nil, nil }
func (m *SqliteDBRepo) SearchMovies(searchQuery string, limit int, offset int) ([]models.Movie, error) { return nil, nil }
func (m *SqliteDBRepo) SearchShows(searchQuery string, limit int, offset int) ([]models.TVSeries, error) { return nil, nil }
func (m *SqliteDBRepo) SearchPeople(searchQuery string, limit int, offset int) ([]models.Person, error) { return nil, nil }
func (m *SqliteDBRepo) GetCastForMedia(mediaID int) ([]models.CastMember, error) { return nil, nil }
func (m *SqliteDBRepo) GetCrewForMedia(mediaID int) ([]models.CrewMember, error) { return nil, nil }
func (m *SqliteDBRepo) GetPersonMovies(personID int) ([]models.Movie, error) { return nil, nil }
func (m *SqliteDBRepo) GetPersonShows(personID int) ([]models.TVSeries, error) { return nil, nil }
func (m *SqliteDBRepo) InsertMovie(mov models.Movie) (int, error) { return 0, nil }
func (m *SqliteDBRepo) InsertShow(s models.TVSeries) (int, error) { return 0, nil }
func (m *SqliteDBRepo) InsertPerson(p models.Person) (int, error) { return 0, nil }
func (m *SqliteDBRepo) InsertMediaCast(mediaID int, personID int, character string, order int) error { return nil }
