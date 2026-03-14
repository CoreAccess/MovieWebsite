package service

import (
	"database/sql"
	"movieweb/internal/models"
	"testing"
)

type mockRepo struct {
	movieCalls int
	showCalls  int
	userCalls  int
}

func (m *mockRepo) Connection() *sql.DB { return nil }
func (m *mockRepo) InitDB(dsn, key string) (*sql.DB, error) { return nil, nil }
func (m *mockRepo) CreateUser(u, e, h string) error { return nil }
func (m *mockRepo) GetUserByEmail(e string) (models.User, error) { return models.User{}, nil }
func (m *mockRepo) GetUserByID(id int) (models.User, error) { return models.User{}, nil }
func (m *mockRepo) GetAllUsers(limit, offset int) ([]models.User, error) {
	m.userCalls++
	return []models.User{{Username: "test"}}, nil
}
func (m *mockRepo) UpdateUserProfile(id int, e, a string) error { return nil }
func (m *mockRepo) SearchMedia(q string, l, o int) ([]models.Media, error) { return nil, nil }
func (m *mockRepo) GetMediaByID(id int) (*models.Media, error) { return nil, nil }
func (m *mockRepo) GetAllMovies(limit, offset int, sort string) ([]models.Movie, error) {
	m.movieCalls++
	return []models.Movie{{Media: models.Media{Name: "test"}}}, nil
}
func (m *mockRepo) GetPopularMovies(l int) ([]models.Movie, error) { return nil, nil }
func (m *mockRepo) GetUpcomingMovies(l int) ([]models.Movie, error) { return nil, nil }
func (m *mockRepo) GetMovieByID(id int) (*models.Movie, error) { return nil, nil }
func (m *mockRepo) GetAllShows(limit, offset int, sort string) ([]models.TVSeries, error) {
	m.showCalls++
	return []models.TVSeries{{Media: models.Media{Name: "test"}}}, nil
}
func (m *mockRepo) GetPopularShows(l int) ([]models.TVSeries, error) { return nil, nil }
func (m *mockRepo) GetNewShows(l int) ([]models.TVSeries, error) { return nil, nil }
func (m *mockRepo) GetShowByID(id int) (*models.TVSeries, error) { return nil, nil }
func (m *mockRepo) GetTVEpisodes(id int) ([]models.TVEpisode, error) { return nil, nil }
func (m *mockRepo) GetAllPeople(limit, offset int, sort string) ([]models.Person, error) { return nil, nil }
func (m *mockRepo) GetPersonByID(id int) (*models.Person, error) { return nil, nil }
func (m *mockRepo) SearchMovies(q string, l, o int) ([]models.Movie, error) { return nil, nil }
func (m *mockRepo) SearchShows(q string, l, o int) ([]models.TVSeries, error) { return nil, nil }
func (m *mockRepo) SearchPeople(q string, l, o int) ([]models.Person, error) { return nil, nil }
func (m *mockRepo) GetCastForMedia(id int) ([]models.CastMember, error) { return nil, nil }
func (m *mockRepo) GetCrewForMedia(id int) ([]models.CrewMember, error) { return nil, nil }
func (m *mockRepo) GetPersonMovies(id int) ([]models.Movie, error) { return nil, nil }
func (m *mockRepo) GetPersonShows(id int) ([]models.TVSeries, error) { return nil, nil }
func (m *mockRepo) InsertMovie(m1 models.Movie) (int, error) { return 0, nil }
func (m *mockRepo) InsertShow(s models.TVSeries) (int, error) { return 0, nil }
func (m *mockRepo) InsertPerson(p models.Person) (int, error) { return 0, nil }
func (m *mockRepo) InsertMediaCast(m1, p int, c string, o int) error { return nil }
func (m *mockRepo) GetUserWatchlist(id int) ([]models.Movie, []models.TVSeries, error) { return nil, nil, nil }
func (m *mockRepo) GetUserWatchlists(id int) ([]models.Watchlist, error) { return nil, nil }
func (m *mockRepo) CreateWatchlist(id int, n, d string) error { return nil }
func (m *mockRepo) AddToWatchlist(id int, mt string, mi int) error { return nil }
func (m *mockRepo) CreateSession(s models.Session) error { return nil }
func (m *mockRepo) GetSession(id string) (models.Session, error) { return models.Session{}, nil }
func (m *mockRepo) DeleteSession(id string) error { return nil }
func (m *mockRepo) GetAdminMetrics() (int, int, error) { return 0, 0, nil }

func TestAppService_Cache(t *testing.T) {
	repo := &mockRepo{}
	s := NewAppService(repo)

	// Test GetAllMovies caching
	for i := 0; i < 5; i++ {
		_, err := s.GetAllMovies(10, 0, "")
		if err != nil {
			t.Fatalf("GetAllMovies failed: %v", err)
		}
	}
	if repo.movieCalls != 1 {
		t.Errorf("Expected 1 movie call, got %d", repo.movieCalls)
	}

	// Test GetAllShows caching
	for i := 0; i < 5; i++ {
		_, err := s.GetAllShows(10, 0, "")
		if err != nil {
			t.Fatalf("GetAllShows failed: %v", err)
		}
	}
	if repo.showCalls != 1 {
		t.Errorf("Expected 1 show call, got %d", repo.showCalls)
	}

	// Test GetAllUsers caching
	for i := 0; i < 5; i++ {
		_, err := s.GetAllUsers(10, 0)
		if err != nil {
			t.Fatalf("GetAllUsers failed: %v", err)
		}
	}
	if repo.userCalls != 1 {
		t.Errorf("Expected 1 user call, got %d", repo.userCalls)
	}
}
