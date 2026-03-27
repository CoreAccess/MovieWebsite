package repository

import (
	"database/sql"
	"filmgap/internal/models"
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
	GetMediaGenres(id int) ([]models.Genre, error)
	GetWatchProviders(mediaID int, countryCode string) ([]models.WatchProviderOption, error)
	ReplaceWatchProviders(mediaID int, countryCode string, providers []models.WatchProviderOption) error

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
	GetWatchlistByID(id int) (models.Watchlist, []models.Movie, []models.TVSeries, error)
	CreateWatchlist(userID int, name, description string) (int, error)
	AddToWatchlist(watchlistID int, mediaType string, mediaID int) error
	RemoveFromWatchlist(watchlistID int, mediaID int) error

	// List Operations
	CreateList(list models.List) (int, error)
	UpdateList(list models.List) error
	DeleteList(id int) error
	GetListByID(id int) (models.List, error)
	GetListBySlug(userID int, slug string) (models.List, error)
	GetListsByUserID(userID int) ([]models.List, error)
	GetPublicLists(limit, offset int) ([]models.List, error)
	AddListItem(item models.ListItem) error
	RemoveListItem(listID, mediaID int) error
	GetListItems(listID int) ([]models.ListItem, error)
	UpdateListItemRank(listID, mediaID, rank int) error

	// Review & Rating Operations
	UpsertReview(r models.Review) error
	GetReviewsForMedia(mediaID int, limit, offset int) ([]models.Review, error)
	GetUserReviewForMedia(userID, mediaID int) (*models.Review, error)
	DeleteReview(userID, mediaID int) error

	// Session Operations
	CreateSession(s models.Session) error
	GetSession(id string) (models.Session, error)
	DeleteSession(id string) error

	// Homepage Hero Operations
	GetCurrentHeroFeature() (int, string, error)
	SetCurrentHeroFeature(mediaID int) error
	GetUpcomingPopularMedia() (int, error)

	// Ingestor Operations
	EnqueueIngestion(tmdbID int, mediaType string) (int, error)
	ClaimNextIngestionJob(maxAttempts int) (models.IngestionJob, error)
	CompleteIngestionJob(id int) error
	FailIngestionJob(id int) error
	// ResetStuckIngestionJobs resets any PROCESSING jobs left over from a previous
	// crash back to QUEUED so they will be retried on the next startup.
	ResetStuckIngestionJobs() error
	UpsertMovie(m models.Movie) (int, error)
	UpsertShow(s models.TVSeries) (int, error)
	UpsertPerson(name, slug, image string) (int, error)
	UpsertGenreForMedia(mediaID, genreID int, name string) error

	GetAdminMetrics() (userCount, mediaCount int, err error)

	// Homepage Expansion Methods
	GetHomepageStats() (models.HomepageStats, error)
	GetTrendingMedia(mediaType string, limit, offset int) ([]models.MediaSummary, error)
	GetTrendingPeople(limit, offset int) ([]models.Person, error)
	GetRecentActivity(limit int) ([]models.Activity, error)
	GetRecentBlogPosts(limit int) ([]models.BlogPost, error)
	GetBlogPostBySlug(slug string) (models.BlogPost, error)
	GetAllBlogPosts(limit, offset int) ([]models.BlogPost, error)
	GetActorBirthdays(date string) ([]models.Person, error)
	GetTopBoxOffice(limit int) ([]models.MediaSummary, error)
	GetFanFavorites(limit int) ([]models.MediaSummary, error)
	GetPopularLists(limit int) ([]models.List, error)
	GetFranchiseSpotlight() (models.Franchise, error)
	GetFranchiseBySlug(slug string) (models.Franchise, error)
	GetFranchiseMedia(franchiseID int) ([]models.MediaSummary, error)
	GetRecentPhotos(limit int) ([]models.Photo, error)

	// Data Seeding/Mutation for Homepage
	CreateBlogPost(post models.BlogPost) error
	LogActivity(activity models.Activity) error
	CreateFranchise(f models.Franchise) error
	CreatePhoto(p models.Photo) error

	// RecalculateMediaRating recomputes and persists aggregate_rating and rating_count
	// for a given media item based on all stored reviews.
	RecalculateMediaRating(mediaID int) error

	// Social Graph (Follow System)
	FollowUser(followerID, followedUserID int) error
	UnfollowUser(followerID, followedUserID int) error
	FollowPerson(followerID, personID int) error
	UnfollowPerson(followerID, personID int) error
	FollowList(followerID, listID int) error
	UnfollowList(followerID, listID int) error
	IsFollowingUser(followerID, followedUserID int) (bool, error)
	IsFollowingPerson(followerID, personID int) (bool, error)
	IsFollowingList(followerID, listID int) (bool, error)
	GetFollowers(userID int) ([]models.User, error)
	GetFollowingUsers(userID int) ([]models.User, error)
	GetFollowingPeople(userID int) ([]models.Person, error)
	GetFollowCounts(userID int) (followers, following int, err error)
	GetPersonFollowCounts(personID int) (followers, following int, err error)
	GetListFollowCounts(listID int) (followers, following int, err error)
	GetActivitiesByFollowed(userID int, limit int) ([]models.Activity, error)
}
