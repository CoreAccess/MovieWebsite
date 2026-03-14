package repository

import (
	"database/sql"
	"movieweb/internal/models"
)

// DatabaseRepo defines the core data access contract.
type DatabaseRepo interface {
	Connection() *sql.DB

	// InitDB is responsible for setup, schema creation, and any necessary initial seed data.
	InitDB(dataSourceName string, tmdbAPIKey string) (*sql.DB, error)

	// User Operations
	CreateUser(username, email, hash string) error
	GetUserByEmail(email string) (models.User, error)
	GetUserByID(id int) (models.User, error)
	GetAllUsers(limit int, offset int) ([]models.User, error)
	UpdateUserProfile(userID int, email string, avatar string) error

	// Core Media Operations (Supertype queries)
	SearchMedia(query string, limit, offset int) ([]models.Media, error)
	GetMediaByID(id int) (*models.Media, error)

	// Subtype Operations
	GetAllMovies(limit int, offset int, sort string) ([]models.Movie, error)
	GetPopularMovies(limit int) ([]models.Movie, error)
	GetUpcomingMovies(limit int) ([]models.Movie, error)
	GetMovieByID(id int) (*models.Movie, error)

	GetAllShows(limit int, offset int, sort string) ([]models.TVSeries, error)
	GetPopularShows(limit int) ([]models.TVSeries, error)
	GetNewShows(limit int) ([]models.TVSeries, error)
	GetShowByID(id int) (*models.TVSeries, error)
	GetTVEpisodes(seriesID int) ([]models.TVEpisode, error)

	GetAllPeople(limit int, offset int, sort string) ([]models.Person, error)
	GetPersonByID(id int) (*models.Person, error)

	// Search Operations (Vector Database ready mappings)
	SearchMovies(searchQuery string, limit int, offset int) ([]models.Movie, error)
	SearchShows(searchQuery string, limit int, offset int) ([]models.TVSeries, error)
	SearchPeople(searchQuery string, limit int, offset int) ([]models.Person, error)

	// Relational Operations
	GetCastForMedia(mediaID int) ([]models.CastMember, error)
	GetCrewForMedia(mediaID int) ([]models.CrewMember, error)
	GetPersonMovies(personID int) ([]models.Movie, error)
	GetPersonShows(personID int) ([]models.TVSeries, error)

	// Write Operations (For Seeding & Wiki Edits)
	InsertMovie(m models.Movie) (int, error)
	InsertShow(s models.TVSeries) (int, error)
	InsertPerson(p models.Person) (int, error)
	InsertMediaCast(mediaID int, personID int, character string, order int) error

	// Watchlist Operations
	GetUserWatchlist(userID int) ([]models.Movie, []models.TVSeries, error)
	GetUserWatchlists(userID int) ([]models.Watchlist, error)
	CreateWatchlist(userID int, name, description string) error
	AddToWatchlist(watchlistID int, mediaType string, mediaID int) error

	// Session Operations
	CreateSession(s models.Session) error
	GetSession(id string) (models.Session, error)
	DeleteSession(id string) error

	GetAdminMetrics() (userCount, mediaCount int, err error)
}
