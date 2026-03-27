package dbrepo

import (
	"context"
	"database/sql"
	"filmgap/internal/models"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// PostgresDBRepo implements the DatabaseRepo interface for PostgreSQL
type PostgresDBRepo struct {
	DB *sql.DB
}

func (m *PostgresDBRepo) Connection() *sql.DB {
	return m.DB
}

func (m *PostgresDBRepo) InitDB(dataSourceName string, tmdbAPIKey string) (*sql.DB, error) {
	var err error
	m.DB, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = m.DB.Ping(); err != nil {
		return nil, err
	}

	log.Println("Connected to PostgreSQL Database")

	// Construct the complete PostgreSQL schema (which includes all Supertype tables natively)
	m.createTables()
	m.seedDataIfEmpty(tmdbAPIKey)

	return m.DB, nil
}

// User Operations
func (m *PostgresDBRepo) CreateUser(username, email, hash string) error {
	query := `INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)`
	_, err := m.DB.Exec(query, username, email, hash)
	return err
}

func (m *PostgresDBRepo) GetUserByEmail(email string) (models.User, error) {
	var u models.User
	query := `SELECT id, username, email, password_hash, COALESCE(avatar, '/static/img/default_avatar.webp'), COALESCE(role, 'user'), created_at FROM users WHERE email = $1`
	err := m.DB.QueryRow(query, email).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Avatar, &u.Role, &u.CreatedAt)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (m *PostgresDBRepo) GetUserByID(id int) (models.User, error) {
	var u models.User
	query := `SELECT id, username, email, COALESCE(avatar, '/static/img/default_avatar.webp'), COALESCE(role, 'user'), created_at FROM users WHERE id = $1`
	err := m.DB.QueryRow(query, id).Scan(&u.ID, &u.Username, &u.Email, &u.Avatar, &u.Role, &u.CreatedAt)
	return u, err
}

func (m *PostgresDBRepo) GetAllUsers(limit int, offset int) ([]models.User, error) {
	query := `SELECT id, username, email, COALESCE(avatar, '/static/img/default_avatar.webp'), COALESCE(role, 'user'), COALESCE(reputation_score, 0), created_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := m.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Avatar, &u.Role, &u.ReputationScore, &u.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (m *PostgresDBRepo) UpdateUserProfile(userID int, email string, avatar string) error {
	query := `UPDATE users SET email = $1, avatar = $2 WHERE id = $3`
	_, err := m.DB.Exec(query, email, avatar, userID)
	return err
}

// Subtype Operations
func (m *PostgresDBRepo) GetMovieByID(id int) (*models.Movie, error) {
	var mov models.Movie
	var budget, boxOffice, langCode, countryCode, tagline, subtitle sql.NullString
	query := `
		SELECT m.id, m.name, m.slug, COALESCE(m.date_published, ''), COALESCE(m.content_rating, ''), COALESCE(m.aggregate_rating, 0.0), COALESCE(m.description, ''), COALESCE(m.image, ''), 
		       COALESCE(s.budget, ''), COALESCE(s.box_office, ''), COALESCE(s.duration, 0), COALESCE(m.language_code, ''), COALESCE(m.country_code, ''), COALESCE(m.tagline, ''), 
		       COALESCE(m.rating_count, 0), COALESCE(m.review_count, 0), COALESCE(m.best_rating, 10.0), COALESCE(m.worst_rating, 1.0), COALESCE(m.is_family_friendly, TRUE), COALESCE(m.subtitle, ''),
		       COALESCE(m.tmdb_id, 0)
		FROM media m
		JOIN movies s ON m.id = s.media_id
		WHERE m.id = $1 AND m.media_type = 'Movie'
	`
	err := m.DB.QueryRow(query, id).Scan(
		&mov.ID, &mov.Name, &mov.Slug, &mov.DatePublished, &mov.ContentRating, &mov.AggregateRating, &mov.Description, &mov.Image,
		&budget, &boxOffice, &mov.Duration, &langCode, &countryCode, &tagline,
		&mov.RatingCount, &mov.ReviewCount, &mov.BestRating, &mov.WorstRating, &mov.IsFamilyFriendly, &subtitle, &mov.TmdbID,
	)
	if err != nil {
		return nil, err
	}
	mov.Budget = budget.String
	mov.BoxOffice = boxOffice.String
	mov.LanguageCode = langCode.String
	mov.CountryCode = countryCode.String
	mov.Tagline = tagline.String
	return &mov, nil
}

func (m *PostgresDBRepo) SearchMedia(query string, limit, offset int) ([]models.Media, error) {
	sqlQuery := `
		SELECT id, media_type, name, slug, COALESCE(image, ''), COALESCE(date_published, ''), COALESCE(aggregate_rating, 0.0)
		FROM media
		WHERE name ILIKE $1
		ORDER BY aggregate_rating DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := m.DB.Query(sqlQuery, "%"+query+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.Media
	for rows.Next() {
		var med models.Media
		if err := rows.Scan(&med.ID, &med.MediaType, &med.Name, &med.Slug, &med.Image, &med.DatePublished, &med.AggregateRating); err != nil {
			return nil, err
		}
		results = append(results, med)
	}
	return results, nil
}

func (m *PostgresDBRepo) GetMediaByID(id int) (*models.Media, error) {
	var med models.Media
	query := `SELECT id, media_type, name, slug, COALESCE(image, ''), COALESCE(date_published, ''), COALESCE(aggregate_rating, 0.0) FROM media WHERE id = $1`
	err := m.DB.QueryRow(query, id).Scan(&med.ID, &med.MediaType, &med.Name, &med.Slug, &med.Image, &med.DatePublished, &med.AggregateRating)
	if err != nil {
		return nil, err
	}
	return &med, nil
}
func (m *PostgresDBRepo) GetAllMovies(limit int, offset int, sort string) ([]models.Movie, error) {
	var query string
	switch sort {
	case "pop":
		query = `
			SELECT m.id, m.name, m.slug, COALESCE(m.image, ''), COALESCE(m.date_published, ''), COALESCE(m.aggregate_rating, 0.0)
			FROM media m
			JOIN movies s ON m.id = s.media_id
			WHERE m.media_type = 'Movie'
			ORDER BY m.aggregate_rating DESC, m.date_published DESC
			LIMIT $1 OFFSET $2`
	case "rating":
		query = `
			SELECT m.id, m.name, m.slug, COALESCE(m.image, ''), COALESCE(m.date_published, ''), COALESCE(m.aggregate_rating, 0.0)
			FROM media m
			JOIN movies s ON m.id = s.media_id
			WHERE m.media_type = 'Movie'
			ORDER BY m.aggregate_rating DESC
			LIMIT $1 OFFSET $2`
	default:
		query = `
			SELECT m.id, m.name, m.slug, COALESCE(m.image, ''), COALESCE(m.date_published, ''), COALESCE(m.aggregate_rating, 0.0)
			FROM media m
			JOIN movies s ON m.id = s.media_id
			WHERE m.media_type = 'Movie'
			ORDER BY m.date_published DESC
			LIMIT $1 OFFSET $2`
	}

	rows, err := m.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var mov models.Movie
		if err := rows.Scan(&mov.ID, &mov.Name, &mov.Slug, &mov.Image, &mov.DatePublished, &mov.AggregateRating); err != nil {
			return nil, err
		}
		movies = append(movies, mov)
	}
	return movies, nil
}

func (m *PostgresDBRepo) GetPopularMovies(limit int) ([]models.Movie, error) {
	return m.GetAllMovies(limit, 0, "pop")
}

func (m *PostgresDBRepo) GetUpcomingMovies(limit int) ([]models.Movie, error) {
	return m.GetAllMovies(limit, 0, "date")
}

func (m *PostgresDBRepo) GetAllShows(limit int, offset int, sort string) ([]models.TVSeries, error) {
	var query string
	switch sort {
	case "pop":
		query = `
			SELECT m.id, m.name, m.slug, COALESCE(m.image, ''), COALESCE(m.date_published, ''), COALESCE(m.aggregate_rating, 0.0)
			FROM media m
			JOIN tv_series s ON m.id = s.media_id
			WHERE m.media_type = 'TVSeries'
			ORDER BY m.aggregate_rating DESC, m.date_published DESC
			LIMIT $1 OFFSET $2`
	case "rating":
		query = `
			SELECT m.id, m.name, m.slug, COALESCE(m.image, ''), COALESCE(m.date_published, ''), COALESCE(m.aggregate_rating, 0.0)
			FROM media m
			JOIN tv_series s ON m.id = s.media_id
			WHERE m.media_type = 'TVSeries'
			ORDER BY m.aggregate_rating DESC
			LIMIT $1 OFFSET $2`
	default:
		query = `
			SELECT m.id, m.name, m.slug, COALESCE(m.image, ''), COALESCE(m.date_published, ''), COALESCE(m.aggregate_rating, 0.0)
			FROM media m
			JOIN tv_series s ON m.id = s.media_id
			WHERE m.media_type = 'TVSeries'
			ORDER BY m.date_published DESC
			LIMIT $1 OFFSET $2`
	}

	rows, err := m.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shows []models.TVSeries
	for rows.Next() {
		var s models.TVSeries
		if err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.Image, &s.StartDate, &s.AggregateRating); err != nil {
			return nil, err
		}
		shows = append(shows, s)
	}
	return shows, nil
}

func (m *PostgresDBRepo) GetPopularShows(limit int) ([]models.TVSeries, error) {
	return m.GetAllShows(limit, 0, "pop")
}

func (m *PostgresDBRepo) GetNewShows(limit int) ([]models.TVSeries, error) {
	return m.GetAllShows(limit, 0, "date")
}

func (m *PostgresDBRepo) GetAllPeople(limit int, offset int, sort string) ([]models.Person, error) {
	query := `
		SELECT id, name, slug, COALESCE(gender, ''), COALESCE(description, ''), COALESCE(image, ''), COALESCE(popularity_score, 0.0) 
		FROM people 
		ORDER BY popularity_score DESC 
		LIMIT $1 OFFSET $2
	`
	rows, err := m.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []models.Person
	for rows.Next() {
		var p models.Person
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Gender, &p.Biography, &p.Image, &p.PopularityScore); err != nil {
			return nil, err
		}
		people = append(people, p)
	}
	return people, nil
}

func (m *PostgresDBRepo) SearchMovies(searchQuery string, limit int, offset int) ([]models.Movie, error) {
	query := `
		SELECT m.id, m.name, m.slug, COALESCE(m.image, ''), COALESCE(m.date_published, ''), COALESCE(m.aggregate_rating, 0.0)
		FROM media m
		JOIN movies s ON m.id = s.media_id
		WHERE m.name ILIKE $1 AND m.media_type = 'Movie'
		ORDER BY m.aggregate_rating DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := m.DB.Query(query, "%"+searchQuery+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var mov models.Movie
		if err := rows.Scan(&mov.ID, &mov.Name, &mov.Slug, &mov.Image, &mov.DatePublished, &mov.AggregateRating); err != nil {
			return nil, err
		}
		movies = append(movies, mov)
	}
	return movies, nil
}

func (m *PostgresDBRepo) SearchShows(searchQuery string, limit int, offset int) ([]models.TVSeries, error) {
	query := `
		SELECT m.id, m.name, m.slug, COALESCE(m.image, ''), COALESCE(m.date_published, ''), COALESCE(m.aggregate_rating, 0.0)
		FROM media m
		JOIN tv_series s ON m.id = s.media_id
		WHERE m.name ILIKE $1 AND m.media_type = 'TVSeries'
		ORDER BY m.aggregate_rating DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := m.DB.Query(query, "%"+searchQuery+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shows []models.TVSeries
	for rows.Next() {
		var s models.TVSeries
		if err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.Image, &s.StartDate, &s.AggregateRating); err != nil {
			return nil, err
		}
		shows = append(shows, s)
	}
	return shows, nil
}

func (m *PostgresDBRepo) SearchPeople(searchQuery string, limit int, offset int) ([]models.Person, error) {
	query := `
		SELECT id, name, slug, COALESCE(gender, ''), COALESCE(description, ''), COALESCE(image, ''), COALESCE(popularity_score, 0.0)
		FROM people
		WHERE name ILIKE $1
		ORDER BY popularity_score DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := m.DB.Query(query, "%"+searchQuery+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []models.Person
	for rows.Next() {
		var p models.Person
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Gender, &p.Biography, &p.Image, &p.PopularityScore); err != nil {
			return nil, err
		}
		people = append(people, p)
	}
	return people, nil
}
func (m *PostgresDBRepo) GetCastForMedia(mediaID int) ([]models.CastMember, error) {
	query := `
		SELECT 
			p.id, p.name, p.slug, COALESCE(p.image, ''), 
			COALESCE(mc.character_name, ''), 
			COALESCE(mc.list_order, 0)
		FROM media_cast mc
		JOIN people p ON mc.person_id = p.id
		WHERE mc.media_id = $1
		ORDER BY mc.list_order ASC
	`
	rows, err := m.DB.Query(query, mediaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cast []models.CastMember
	for rows.Next() {
		var cm models.CastMember
		var pImg sql.NullString
		err := rows.Scan(
			&cm.Person.ID, &cm.Person.Name, &cm.Person.Slug, &pImg,
			&cm.Character.Name,
			&cm.BillingOrder,
		)
		if err != nil {
			return nil, err
		}
		cm.Person.Image = pImg.String
		// Use a local slugify or similar, for now assuming name is enough or handle in service
		cast = append(cast, cm)
	}
	return cast, nil
}

func (m *PostgresDBRepo) GetCrewForMedia(mediaID int) ([]models.CrewMember, error) {
	query := `
		SELECT p.id, p.name, p.slug, COALESCE(p.image, ''), mc.job, mc.department
		FROM media_crew mc
		JOIN people p ON mc.person_id = p.id
		WHERE mc.media_id = $1
	`
	rows, err := m.DB.Query(query, mediaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var crew []models.CrewMember
	for rows.Next() {
		var cm models.CrewMember
		var pImg sql.NullString
		if err := rows.Scan(&cm.Person.ID, &cm.Person.Name, &cm.Person.Slug, &pImg, &cm.Job, &cm.Department); err != nil {
			return nil, err
		}
		cm.Person.Image = pImg.String
		crew = append(crew, cm)
	}
	return crew, nil
}

func (m *PostgresDBRepo) GetMediaGenres(mediaID int) ([]models.Genre, error) {
	query := `
		SELECT g.id, g.name, g.slug
		FROM genres g
		JOIN media_genres mg ON g.id = mg.genre_id
		WHERE mg.media_id = $1
	`
	rows, err := m.DB.Query(query, mediaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []models.Genre
	for rows.Next() {
		var g models.Genre
		if err := rows.Scan(&g.ID, &g.Name, &g.Slug); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	return genres, nil
}

func (m *PostgresDBRepo) GetWatchProviders(mediaID int, countryCode string) ([]models.WatchProviderOption, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, media_id, country_code, provider_type, provider_id, provider_name,
		       COALESCE(logo_url, ''), COALESCE(display_priority, 0), COALESCE(deep_link_url, ''),
		       COALESCE(source, 'tmdb'), updated_at
		FROM watch_providers
		WHERE media_id = $1 AND country_code = $2
		ORDER BY provider_type ASC, display_priority ASC, provider_name ASC
	`
	rows, err := m.DB.QueryContext(ctx, query, mediaID, countryCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []models.WatchProviderOption
	for rows.Next() {
		var provider models.WatchProviderOption
		if err := rows.Scan(
			&provider.ID,
			&provider.MediaID,
			&provider.CountryCode,
			&provider.ProviderType,
			&provider.ProviderID,
			&provider.ProviderName,
			&provider.LogoURL,
			&provider.DisplayPriority,
			&provider.DeepLinkURL,
			&provider.Source,
			&provider.UpdatedAt,
		); err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

func (m *PostgresDBRepo) ReplaceWatchProviders(mediaID int, countryCode string, providers []models.WatchProviderOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx, `DELETE FROM watch_providers WHERE media_id = $1 AND country_code = $2`, mediaID, countryCode); err != nil {
		return err
	}

	for _, provider := range providers {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO watch_providers (
				media_id, country_code, provider_type, provider_id, provider_name,
				logo_url, display_priority, deep_link_url, source, updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		`,
			mediaID,
			countryCode,
			provider.ProviderType,
			provider.ProviderID,
			provider.ProviderName,
			provider.LogoURL,
			provider.DisplayPriority,
			provider.DeepLinkURL,
			provider.Source,
			provider.UpdatedAt,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (m *PostgresDBRepo) GetShowByID(id int) (*models.TVSeries, error) {
	var s models.TVSeries
	var langCode, countryCode, tagline, subtitle sql.NullString
	query := `
		SELECT m.id, m.name, m.slug, COALESCE(m.date_published, ''), COALESCE(m.content_rating, ''), COALESCE(s.end_date, ''), COALESCE(m.aggregate_rating, 0.0), 
		       COALESCE(m.description, ''), COALESCE(m.image, ''), COALESCE(s.number_of_seasons, 0), 
		       COALESCE(m.language_code, 'en'), COALESCE(m.country_code, 'US'), COALESCE(m.tagline, ''), 
		       COALESCE(m.rating_count, 0), COALESCE(m.review_count, 0), COALESCE(m.best_rating, 10.0), COALESCE(m.worst_rating, 1.0), COALESCE(m.subtitle, ''),
		       COALESCE(s.number_of_episodes, 0), COALESCE(m.tmdb_id, 0)
		FROM media m
		JOIN tv_series s ON m.id = s.media_id
		WHERE m.id = $1 AND m.media_type = 'TVSeries'
	`
	err := m.DB.QueryRow(query, id).Scan(
		&s.ID, &s.Name, &s.Slug, &s.StartDate, &s.ContentRating, &s.EndDate, &s.AggregateRating,
		&s.Description, &s.Image, &s.NumberOfSeasons, &langCode, &countryCode, &tagline,
		&s.RatingCount, &s.ReviewCount, &s.BestRating, &s.WorstRating, &subtitle,
		&s.NumberOfEpisodes, &s.TmdbID,
	)
	if err != nil {
		return nil, err
	}
	s.LanguageCode = langCode.String
	s.CountryCode = countryCode.String
	s.Tagline = tagline.String
	s.Subtitle = subtitle.String
	return &s, nil
}

func (m *PostgresDBRepo) GetPersonByID(id int) (*models.Person, error) {
	var p models.Person
	query := `SELECT id, name, slug, COALESCE(gender, ''), COALESCE(description, ''), COALESCE(image, ''), COALESCE(popularity_score, 0.0) FROM people WHERE id = $1`
	err := m.DB.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Slug, &p.Gender, &p.Biography, &p.Image, &p.PopularityScore)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (m *PostgresDBRepo) GetPersonMovies(personID int) ([]models.Movie, error) {
	query := `
		SELECT m.id, m.name, m.slug, COALESCE(m.image, ''), COALESCE(m.date_published, ''), COALESCE(m.aggregate_rating, 0.0)
		FROM media m
		JOIN movies s ON m.id = s.media_id
		JOIN media_cast mc ON m.id = mc.media_id
		WHERE mc.person_id = $1 AND m.media_type = 'Movie'
	`
	rows, err := m.DB.Query(query, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var mov models.Movie
		if err := rows.Scan(&mov.ID, &mov.Name, &mov.Slug, &mov.Image, &mov.DatePublished, &mov.AggregateRating); err != nil {
			return nil, err
		}
		movies = append(movies, mov)
	}
	return movies, nil
}

func (m *PostgresDBRepo) GetPersonShows(personID int) ([]models.TVSeries, error) {
	query := `
		SELECT m.id, m.name, m.slug, COALESCE(m.image, ''), COALESCE(m.date_published, ''), COALESCE(m.aggregate_rating, 0.0)
		FROM media m
		JOIN tv_series s ON m.id = s.media_id
		JOIN media_cast mc ON m.id = mc.media_id
		WHERE mc.person_id = $1 AND m.media_type = 'TVSeries'
	`
	rows, err := m.DB.Query(query, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shows []models.TVSeries
	for rows.Next() {
		var s models.TVSeries
		if err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.Image, &s.StartDate, &s.AggregateRating); err != nil {
			return nil, err
		}
		shows = append(shows, s)
	}
	return shows, nil
}
func (m *PostgresDBRepo) InsertMovie(mov models.Movie) (int, error) {
	// First insert into media
	var mediaID int
	mediaQuery := `
		INSERT INTO media (media_type, name, slug, description, image, date_published, content_rating, aggregate_rating, tmdb_id)
		VALUES ('Movie', $1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (slug, media_type) DO UPDATE SET tmdb_id = EXCLUDED.tmdb_id
		RETURNING id
	`
	err := m.DB.QueryRow(mediaQuery, mov.Name, mov.Slug, mov.Description, mov.Image, mov.DatePublished, mov.ContentRating, mov.AggregateRating, mov.TmdbID).Scan(&mediaID)
	if err != nil {
		return 0, err
	}

	// Then insert into movies
	movieQuery := `
		INSERT INTO movies (media_id, budget, box_office, duration)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (media_id) DO NOTHING
	`
	_, err = m.DB.Exec(movieQuery, mediaID, mov.Budget, mov.BoxOffice, mov.Duration)
	return mediaID, err
}

func (m *PostgresDBRepo) InsertShow(s models.TVSeries) (int, error) {
	var mediaID int
	mediaQuery := `
		INSERT INTO media (media_type, name, slug, description, image, date_published, content_rating, aggregate_rating, tmdb_id)
		VALUES ('TVSeries', $1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (slug, media_type) DO UPDATE SET tmdb_id = EXCLUDED.tmdb_id
		RETURNING id
	`
	err := m.DB.QueryRow(mediaQuery, s.Name, s.Slug, s.Description, s.Image, s.StartDate, s.ContentRating, s.AggregateRating, s.TmdbID).Scan(&mediaID)
	if err != nil {
		return 0, err
	}

	showQuery := `
		INSERT INTO tv_series (media_id, end_date, number_of_seasons, number_of_episodes)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (media_id) DO NOTHING
	`
	_, err = m.DB.Exec(showQuery, mediaID, s.EndDate, s.NumberOfSeasons, s.NumberOfEpisodes)
	return mediaID, err
}

func (m *PostgresDBRepo) InsertPerson(p models.Person) (int, error) {
	var personID int
	query := `
		INSERT INTO people (name, slug, gender, birth_date, birth_place, death_date, description, image, known_for_department, popularity_score)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (slug) DO UPDATE SET slug = EXCLUDED.slug
		RETURNING id
	`
	err := m.DB.QueryRow(query, p.Name, p.Slug, p.Gender, p.BirthDate, p.BirthPlace, p.Deathday, p.Biography, p.Image, p.KnownForDepartment, p.PopularityScore).Scan(&personID)
	return personID, err
}

func (m *PostgresDBRepo) InsertMediaCast(mediaID int, personID int, character string, order int) error {
	query := `
		INSERT INTO media_cast (media_id, person_id, character_name, list_order)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (media_id, person_id, character_name) DO NOTHING
	`
	_, err := m.DB.Exec(query, mediaID, personID, character, order)
	return err
}

func (m *PostgresDBRepo) GetUserWatchlist(userID int) ([]models.Movie, []models.TVSeries, error) {
	var watchlistID int
	err := m.DB.QueryRow("SELECT id FROM watchlists WHERE user_id = $1 LIMIT 1", userID).Scan(&watchlistID)
	if err != nil {
		return []models.Movie{}, []models.TVSeries{}, nil
	}
	_, movies, shows, err := m.GetWatchlistByID(watchlistID)
	return movies, shows, err
}

func (m *PostgresDBRepo) GetWatchlistByID(id int) (models.Watchlist, []models.Movie, []models.TVSeries, error) {
	var w models.Watchlist
	query := `
		SELECT w.id, w.user_id, w.name, w.description, w.is_public, w.created_at, COUNT(wi.id) as item_count 
		FROM watchlists w
		LEFT JOIN watchlist_items wi ON w.id = wi.watchlist_id
		WHERE w.id = $1 
		GROUP BY w.id`

	err := m.DB.QueryRow(query, id).Scan(&w.ID, &w.UserID, &w.Name, &w.Description, &w.IsPublic, &w.CreatedAt, &w.ItemCount)
	if err != nil {
		return w, nil, nil, err
	}

	movieQuery := `
		SELECT m.id, m.name, m.slug, COALESCE(m.date_published, ''), COALESCE(m.aggregate_rating, 0.0), COALESCE(m.description, ''), COALESCE(m.image, '')
		FROM media m
		JOIN movies s ON m.id = s.media_id
		JOIN watchlist_items wi ON m.id = wi.media_id
		WHERE wi.watchlist_id = $1 AND m.media_type = 'Movie'
		ORDER BY wi.added_at DESC
	`
	movieRows, err := m.DB.Query(movieQuery, id)
	if err != nil {
		return w, nil, nil, err
	}
	defer movieRows.Close()

	var movies []models.Movie
	for movieRows.Next() {
		var mov models.Movie
		if err := movieRows.Scan(&mov.ID, &mov.Name, &mov.Slug, &mov.DatePublished, &mov.AggregateRating, &mov.Description, &mov.Image); err != nil {
			return w, nil, nil, err
		}
		movies = append(movies, mov)
	}

	showQuery := `
		SELECT m.id, m.name, m.slug, COALESCE(m.date_published, ''), COALESCE(s.end_date, ''), COALESCE(m.aggregate_rating, 0.0), COALESCE(m.description, ''), COALESCE(m.image, ''), COALESCE(s.number_of_seasons, 0)
		FROM media m
		JOIN tv_series s ON m.id = s.media_id
		JOIN watchlist_items wi ON m.id = wi.media_id
		WHERE wi.watchlist_id = $1 AND m.media_type = 'TVSeries'
		ORDER BY wi.added_at DESC
	`
	showRows, err := m.DB.Query(showQuery, id)
	if err != nil {
		return w, nil, nil, err
	}
	defer showRows.Close()

	var shows []models.TVSeries
	for showRows.Next() {
		var s models.TVSeries
		if err := showRows.Scan(&s.ID, &s.Name, &s.Slug, &s.StartDate, &s.EndDate, &s.AggregateRating, &s.Description, &s.Image, &s.NumberOfSeasons); err != nil {
			return w, nil, nil, err
		}
		shows = append(shows, s)
	}

	return w, movies, shows, nil
}

func (m *PostgresDBRepo) GetUserWatchlists(userID int) ([]models.Watchlist, error) {
	query := `
		SELECT w.id, w.user_id, w.name, w.description, w.is_public, w.created_at, COUNT(wi.id) as item_count 
		FROM watchlists w
		LEFT JOIN watchlist_items wi ON w.id = wi.watchlist_id
		WHERE w.user_id = $1 
		GROUP BY w.id
		ORDER BY w.created_at DESC`
	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var watchlists []models.Watchlist
	for rows.Next() {
		var w models.Watchlist
		err := rows.Scan(&w.ID, &w.UserID, &w.Name, &w.Description, &w.IsPublic, &w.CreatedAt, &w.ItemCount)
		if err != nil {
			return nil, err
		}
		watchlists = append(watchlists, w)
	}
	return watchlists, nil
}

func (m *PostgresDBRepo) CreateWatchlist(userID int, name string, description string) (int, error) {
	query := `INSERT INTO watchlists (user_id, name, description, is_public) VALUES ($1, $2, $3, $4) RETURNING id`
	var newID int
	err := m.DB.QueryRow(query, userID, name, description, false).Scan(&newID)
	return newID, err
}

func (m *PostgresDBRepo) AddToWatchlist(watchlistID int, mediaType string, mediaID int) error {
	query := `INSERT INTO watchlist_items (watchlist_id, media_type, media_id) VALUES ($1, $2, $3) ON CONFLICT (watchlist_id, media_id) DO NOTHING`
	_, err := m.DB.Exec(query, watchlistID, mediaType, mediaID)
	return err
}

func (m *PostgresDBRepo) RemoveFromWatchlist(watchlistID int, mediaID int) error {
	query := `DELETE FROM watchlist_items WHERE watchlist_id = $1 AND media_id = $2`
	_, err := m.DB.Exec(query, watchlistID, mediaID)
	return err
}

// List Operations

func (m *PostgresDBRepo) CreateList(list models.List) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newID int
	query := `
		INSERT INTO lists (user_id, name, slug, description, is_ranked, is_collaborative, visibility)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	err := m.DB.QueryRowContext(ctx, query,
		list.UserID, list.Name, list.Slug, list.Description,
		list.IsRanked, list.IsCollaborative, list.Visibility,
	).Scan(&newID)
	if err != nil {
		return 0, err
	}
	return newID, nil
}

func (m *PostgresDBRepo) UpdateList(list models.List) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE lists
		SET name = $1, slug = $2, description = $3, is_ranked = $4, is_collaborative = $5, visibility = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
	`
	_, err := m.DB.ExecContext(ctx, query,
		list.Name, list.Slug, list.Description,
		list.IsRanked, list.IsCollaborative, list.Visibility,
		list.ID,
	)
	return err
}

func (m *PostgresDBRepo) DeleteList(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, "DELETE FROM lists WHERE id = $1", id)
	return err
}

func (m *PostgresDBRepo) GetListByID(id int) (models.List, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var list models.List
	query := `
		SELECT l.id, l.user_id, l.name, l.slug, COALESCE(l.description, ''), l.is_ranked, l.is_collaborative, l.visibility,
		       l.like_count, l.follower_count, l.item_count, l.is_featured, l.created_at, l.updated_at, u.username
		FROM lists l
		JOIN users u ON l.user_id = u.id
		WHERE l.id = $1
	`
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&list.ID, &list.UserID, &list.Name, &list.Slug, &list.Description, &list.IsRanked, &list.IsCollaborative, &list.Visibility,
		&list.LikeCount, &list.FollowerCount, &list.ItemCount, &list.IsFeatured, &list.CreatedAt, &list.UpdatedAt, &list.Username,
	)
	return list, err
}

func (m *PostgresDBRepo) GetListBySlug(userID int, slug string) (models.List, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var list models.List
	query := `
		SELECT l.id, l.user_id, l.name, l.slug, COALESCE(l.description, ''), l.is_ranked, l.is_collaborative, l.visibility,
		       l.like_count, l.follower_count, l.item_count, l.is_featured, l.created_at, l.updated_at, u.username
		FROM lists l
		JOIN users u ON l.user_id = u.id
		WHERE l.user_id = $1 AND l.slug = $2
	`
	err := m.DB.QueryRowContext(ctx, query, userID, slug).Scan(
		&list.ID, &list.UserID, &list.Name, &list.Slug, &list.Description, &list.IsRanked, &list.IsCollaborative, &list.Visibility,
		&list.LikeCount, &list.FollowerCount, &list.ItemCount, &list.IsFeatured, &list.CreatedAt, &list.UpdatedAt, &list.Username,
	)
	return list, err
}

func (m *PostgresDBRepo) GetListsByUserID(userID int) ([]models.List, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, name, slug, COALESCE(description, ''), is_ranked, is_collaborative, visibility,
		       like_count, follower_count, item_count, is_featured, created_at, updated_at
		FROM lists
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []models.List
	for rows.Next() {
		var l models.List
		err := rows.Scan(
			&l.ID, &l.UserID, &l.Name, &l.Slug, &l.Description, &l.IsRanked, &l.IsCollaborative, &l.Visibility,
			&l.LikeCount, &l.FollowerCount, &l.ItemCount, &l.IsFeatured, &l.CreatedAt, &l.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		lists = append(lists, l)
	}
	return lists, nil
}

func (m *PostgresDBRepo) GetPublicLists(limit, offset int) ([]models.List, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT l.id, l.user_id, l.name, l.slug, COALESCE(l.description, ''), l.is_ranked, l.is_collaborative, l.visibility,
		       l.like_count, l.follower_count, l.item_count, l.is_featured, l.created_at, l.updated_at, u.username
		FROM lists l
		JOIN users u ON l.user_id = u.id
		WHERE l.visibility = 'public'
		ORDER BY l.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := m.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []models.List
	for rows.Next() {
		var l models.List
		err := rows.Scan(
			&l.ID, &l.UserID, &l.Name, &l.Slug, &l.Description, &l.IsRanked, &l.IsCollaborative, &l.Visibility,
			&l.LikeCount, &l.FollowerCount, &l.ItemCount, &l.IsFeatured, &l.CreatedAt, &l.UpdatedAt, &l.Username,
		)
		if err != nil {
			return nil, err
		}
		lists = append(lists, l)
	}
	return lists, nil
}

// List Item Operations

func (m *PostgresDBRepo) AddListItem(item models.ListItem) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO list_items (list_id, media_id, rank, note, added_by)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (list_id, media_id) DO NOTHING
	`
	_, err = tx.ExecContext(ctx, query, item.ListID, item.MediaID, item.Rank, item.Note, item.AddedBy)
	if err != nil {
		return err
	}

	// Update item count on the list
	_, err = tx.ExecContext(ctx, "UPDATE lists SET item_count = (SELECT COUNT(*) FROM list_items WHERE list_id = $1) WHERE id = $1", item.ListID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *PostgresDBRepo) RemoveListItem(listID, mediaID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM list_items WHERE list_id = $1 AND media_id = $2", listID, mediaID)
	if err != nil {
		return err
	}

	// Update item count on the list
	_, err = tx.ExecContext(ctx, "UPDATE lists SET item_count = (SELECT COUNT(*) FROM list_items WHERE list_id = $1) WHERE id = $1", listID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *PostgresDBRepo) GetListItems(listID int) ([]models.ListItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT li.id, li.list_id, li.media_id, COALESCE(li.rank, 0), COALESCE(li.note, ''), li.added_by, li.added_at, 
		       m.name, m.slug, m.media_type, COALESCE(m.image, '')
		FROM list_items li
		JOIN media m ON li.media_id = m.id
		WHERE li.list_id = $1
		ORDER BY li.rank ASC, li.added_at ASC
	`
	rows, err := m.DB.QueryContext(ctx, query, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ListItem
	for rows.Next() {
		var i models.ListItem
		err := rows.Scan(
			&i.ID, &i.ListID, &i.MediaID, &i.Rank, &i.Note, &i.AddedBy, &i.AddedAt,
			&i.MediaName, &i.MediaSlug, &i.MediaType, &i.MediaImg,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (m *PostgresDBRepo) UpdateListItemRank(listID, mediaID, rank int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE list_items SET rank = $1 WHERE list_id = $2 AND media_id = $3`
	_, err := m.DB.ExecContext(ctx, query, rank, listID, mediaID)
	return err
}

// Session Operations
func (m *PostgresDBRepo) CreateSession(s models.Session) error {
	_, err := m.DB.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)", s.ID, s.UserID, s.ExpiresAt)
	return err
}

func (m *PostgresDBRepo) GetSession(id string) (models.Session, error) {
	var s models.Session
	err := m.DB.QueryRow("SELECT id, user_id, expires_at FROM sessions WHERE id = $1", id).
		Scan(&s.ID, &s.UserID, &s.ExpiresAt)
	return s, err
}

func (m *PostgresDBRepo) DeleteSession(id string) error {
	_, err := m.DB.Exec("DELETE FROM sessions WHERE id = $1", id)
	return err
}

// Admin & Moderation Operations
func (m *PostgresDBRepo) GetAdminMetrics() (userCount, mediaCount int, err error) {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM users),
			(SELECT COUNT(*) FROM movies) + (SELECT COUNT(*) FROM tv_series)
	`
	err = m.DB.QueryRow(query).Scan(&userCount, &mediaCount)
	return
}

// RecalculateMediaRating recomputes aggregate_rating and rating_count for a media record
// based on the current state of the reviews table. Called after every review upsert.
func (m *PostgresDBRepo) RecalculateMediaRating(mediaID int) error {
	query := `
		UPDATE media
		SET
			aggregate_rating = (
				SELECT COALESCE(ROUND(AVG(rating)::numeric, 1), 0)
				FROM reviews
				WHERE media_id = $1
			),
			rating_count = (
				SELECT COUNT(*)
				FROM reviews
				WHERE media_id = $1
			)
		WHERE id = $1
	`
	_, err := m.DB.Exec(query, mediaID)
	return err
}

func (m *PostgresDBRepo) GetTVEpisodes(seriesID int) ([]models.TVEpisode, error) {
	query := `
		SELECT id, series_id, season_number, episode_number, name, slug, COALESCE(date_published, ''), COALESCE(description, ''), COALESCE(image, ''), COALESCE(duration, 0)
		FROM tv_episodes
		WHERE series_id = $1
		ORDER BY season_number ASC, episode_number ASC
	`
	rows, err := m.DB.Query(query, seriesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var episodes []models.TVEpisode
	for rows.Next() {
		var e models.TVEpisode
		err := rows.Scan(&e.ID, &e.SeriesID, &e.SeasonNumber, &e.EpisodeNumber, &e.Name, &e.Slug, &e.DatePublished, &e.Description, &e.Image, &e.Duration)
		if err != nil {
			return nil, err
		}
		e.AirDate = e.DatePublished
		episodes = append(episodes, e)
	}

	return episodes, nil
}

// Homepage Hero Operations
func (m *PostgresDBRepo) GetCurrentHeroFeature() (int, string, error) {
	var mediaID int
	var selectedAt string
	query := `SELECT media_id, selected_at FROM hero_features ORDER BY selected_at DESC LIMIT 1`
	err := m.DB.QueryRow(query).Scan(&mediaID, &selectedAt)
	return mediaID, selectedAt, err
}

func (m *PostgresDBRepo) SetCurrentHeroFeature(mediaID int) error {
	query := `INSERT INTO hero_features (media_id) VALUES ($1)`
	_, err := m.DB.Exec(query, mediaID)
	return err
}

func (m *PostgresDBRepo) GetUpcomingPopularMedia() (int, error) {
	var mediaID int
	// Attempt to find upcoming movie first
	query := `
		SELECT m.id 
		FROM media m 
		WHERE m.media_type = 'Movie' AND m.date_published > CURRENT_DATE::TEXT
		ORDER BY m.aggregate_rating DESC, m.date_published DESC
		LIMIT 1
	`
	err := m.DB.QueryRow(query).Scan(&mediaID)
	if err == sql.ErrNoRows {
		// Fallback to highest rated movie if no future release exists
		queryFallback := `SELECT m.id FROM media m WHERE m.media_type = 'Movie' ORDER BY m.aggregate_rating DESC LIMIT 1`
		err = m.DB.QueryRow(queryFallback).Scan(&mediaID)
	}
	return mediaID, err
}

// ---------------------------------------------------------------------------
// Ingestor Operations
// ---------------------------------------------------------------------------

// EnqueueIngestion inserts a new job into pending_ingestion.
// Uses ON CONFLICT DO NOTHING so duplicates are silently skipped.
// Returns the inserted row ID, or 0 if the ID already existed.
func (m *PostgresDBRepo) EnqueueIngestion(tmdbID int, mediaType string) (int, error) {
	var id int
	err := m.DB.QueryRow(`
		INSERT INTO pending_ingestion (tmdb_id, media_type)
		VALUES ($1, $2)
		ON CONFLICT (tmdb_id, media_type) DO NOTHING
		RETURNING id
	`, tmdbID, mediaType).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil // already existed, not an error
	}
	return id, err
}

// ClaimNextIngestionJob atomically claims the next QUEUED job, marking it PROCESSING.
// Returns sql.ErrNoRows when the queue is empty.
func (m *PostgresDBRepo) ClaimNextIngestionJob(maxAttempts int) (models.IngestionJob, error) {
	var job models.IngestionJob
	err := m.DB.QueryRow(`
		UPDATE pending_ingestion
		SET status = 'PROCESSING', attempts = attempts + 1, last_attempt = NOW()
		WHERE id = (
			SELECT id FROM pending_ingestion
			WHERE status = 'QUEUED' AND attempts < $1
			ORDER BY created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, tmdb_id, media_type, status, attempts
	`, maxAttempts).Scan(&job.ID, &job.TmdbID, &job.MediaType, &job.Status, &job.Attempts)
	return job, err
}

// CompleteIngestionJob marks a job as COMPLETED.
func (m *PostgresDBRepo) CompleteIngestionJob(id int) error {
	_, err := m.DB.Exec(`UPDATE pending_ingestion SET status = 'COMPLETED' WHERE id = $1`, id)
	return err
}

// FailIngestionJob marks a job as FAILED (max attempts exceeded).
func (m *PostgresDBRepo) FailIngestionJob(id int) error {
	_, err := m.DB.Exec(`UPDATE pending_ingestion SET status = 'FAILED' WHERE id = $1`, id)
	return err
}

// ResetStuckIngestionJobs resets any PROCESSING jobs back to QUEUED.
// Called on ingestor startup to recover from ungraceful shutdowns.
func (m *PostgresDBRepo) ResetStuckIngestionJobs() error {
	_, err := m.DB.Exec(`UPDATE pending_ingestion SET status = 'QUEUED' WHERE status = 'PROCESSING'`)
	return err
}

// UpsertMovie inserts or updates a movie via its tmdb_id.
// Returns the local media.id.
func (m *PostgresDBRepo) UpsertMovie(mov models.Movie) (int, error) {
	var mediaID int

	// Upsert the media supertype row.
	err := m.DB.QueryRow(`
		INSERT INTO media (media_type, name, slug, description, image, date_published, aggregate_rating, tmdb_id, content_rating)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (tmdb_id) DO UPDATE SET
			name            = EXCLUDED.name,
			description     = EXCLUDED.description,
			image           = EXCLUDED.image,
			date_published  = EXCLUDED.date_published,
			aggregate_rating= EXCLUDED.aggregate_rating,
			content_rating  = EXCLUDED.content_rating
		RETURNING id
	`, mov.MediaType, mov.Name, mov.Slug, mov.Description, mov.Image,
		mov.DatePublished, mov.AggregateRating, mov.TmdbID, mov.ContentRating,
	).Scan(&mediaID)
	if err != nil {
		return 0, err
	}

	// Upsert the movies subtype row.
	_, err = m.DB.Exec(`
		INSERT INTO movies (media_id, duration, tagline)
		VALUES ($1, $2, $3)
		ON CONFLICT (media_id) DO UPDATE SET
			duration = EXCLUDED.duration,
			tagline  = EXCLUDED.tagline
	`, mediaID, mov.Duration, mov.Tagline)
	return mediaID, err
}

// UpsertShow inserts or updates a TV series via its tmdb_id.
// Returns the local media.id.
func (m *PostgresDBRepo) UpsertShow(show models.TVSeries) (int, error) {
	var mediaID int
	err := m.DB.QueryRow(`
		INSERT INTO media (media_type, name, slug, description, image, date_published, aggregate_rating, tmdb_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (tmdb_id) DO UPDATE SET
			name            = EXCLUDED.name,
			description     = EXCLUDED.description,
			image           = EXCLUDED.image,
			date_published  = EXCLUDED.date_published,
			aggregate_rating= EXCLUDED.aggregate_rating
		RETURNING id
	`, show.MediaType, show.Name, show.Slug, show.Description, show.Image,
		show.DatePublished, show.AggregateRating, show.TmdbID,
	).Scan(&mediaID)
	if err != nil {
		return 0, err
	}

	_, err = m.DB.Exec(`
		INSERT INTO tv_series (media_id, number_of_seasons)
		VALUES ($1, $2)
		ON CONFLICT (media_id) DO UPDATE SET number_of_seasons = EXCLUDED.number_of_seasons
	`, mediaID, show.NumberOfSeasons)
	return mediaID, err
}

// UpsertPerson inserts or updates a person row by slug.
// Returns the person's local ID.
func (m *PostgresDBRepo) UpsertPerson(name, slug, image string) (int, error) {
	var id int
	err := m.DB.QueryRow(`
		INSERT INTO people (name, slug, image)
		VALUES ($1, $2, $3)
		ON CONFLICT (slug) DO UPDATE SET
			name  = EXCLUDED.name,
			image = CASE WHEN EXCLUDED.image != '' THEN EXCLUDED.image ELSE people.image END
		RETURNING id
	`, name, slug, image).Scan(&id)
	return id, err
}

// UpsertGenreForMedia inserts a genre (if new) and links it to a media item.
func (m *PostgresDBRepo) UpsertGenreForMedia(mediaID, genreID int, name string) error {
	slug := name
	_, err := m.DB.Exec(`
		INSERT INTO genres (id, name, slug) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`, genreID, name, slug)
	if err != nil {
		return err
	}
	_, err = m.DB.Exec(`
		INSERT INTO media_genres (media_id, genre_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, mediaID, genreID)
	return err
}

// Review & Rating Operations

func (m *PostgresDBRepo) UpsertReview(r models.Review) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO reviews (user_id, media_id, rating, title, body, review_type, contains_spoilers, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, media_id) DO UPDATE SET
			rating = EXCLUDED.rating,
			title = EXCLUDED.title,
			body = EXCLUDED.body,
			review_type = EXCLUDED.review_type,
			contains_spoilers = EXCLUDED.contains_spoilers,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := m.DB.ExecContext(ctx, query, r.UserID, r.MediaID, r.Rating, r.Title, r.Body, r.ReviewType, r.ContainsSpoilers)
	return err
}

func (m *PostgresDBRepo) GetReviewsForMedia(mediaID int, limit, offset int) ([]models.Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT r.id, r.user_id, r.media_id, r.rating, COALESCE(r.title, ''), COALESCE(r.body, ''), r.review_type, r.contains_spoilers, r.like_count, r.created_at, r.updated_at, u.username, COALESCE(u.avatar, '/static/img/default_avatar.webp')
		FROM reviews r
		JOIN users u ON r.user_id = u.id
		WHERE r.media_id = $1 AND r.status = 'published'
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := m.DB.QueryContext(ctx, query, mediaID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var r models.Review
		err := rows.Scan(&r.ID, &r.UserID, &r.MediaID, &r.Rating, &r.Title, &r.Body, &r.ReviewType, &r.ContainsSpoilers, &r.LikeCount, &r.CreatedAt, &r.UpdatedAt, &r.Username, &r.Avatar)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}

	return reviews, nil
}

func (m *PostgresDBRepo) GetUserReviewForMedia(userID, mediaID int) (*models.Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT r.id, r.user_id, r.media_id, r.rating, COALESCE(r.title, ''), COALESCE(r.body, ''), r.review_type, r.contains_spoilers, r.like_count, r.created_at, r.updated_at, u.username, COALESCE(u.avatar, '/static/img/default_avatar.webp')
		FROM reviews r
		JOIN users u ON r.user_id = u.id
		WHERE r.user_id = $1 AND r.media_id = $2
	`
	var r models.Review
	err := m.DB.QueryRowContext(ctx, query, userID, mediaID).Scan(
		&r.ID, &r.UserID, &r.MediaID, &r.Rating, &r.Title, &r.Body, &r.ReviewType, &r.ContainsSpoilers, &r.LikeCount, &r.CreatedAt, &r.UpdatedAt, &r.Username, &r.Avatar,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (m *PostgresDBRepo) DeleteReview(userID, mediaID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE FROM reviews WHERE user_id = $1 AND media_id = $2`
	_, err := m.DB.ExecContext(ctx, query, userID, mediaID)
	return err
}

// --- Homepage Expansion Methods ---

// GetHomepageStats returns aggregate counts for the homepage matrix.
func (m *PostgresDBRepo) GetHomepageStats() (models.HomepageStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var stats models.HomepageStats
	query := `
		SELECT
			(SELECT COUNT(*) FROM users) as user_count,
			(SELECT COUNT(*) FROM media WHERE media_type = 'Movie') as movie_count,
			(SELECT COUNT(*) FROM media WHERE media_type = 'TVSeries') as show_count,
			(SELECT COUNT(*) FROM reviews) as review_count
	`
	err := m.DB.QueryRowContext(ctx, query).Scan(
		&stats.UserCount,
		&stats.MovieCount,
		&stats.ShowCount,
		&stats.ReviewCount,
	)
	return stats, err
}

// GetTrendingMedia returns popular media based on rating and interaction count.
func (m *PostgresDBRepo) GetTrendingMedia(mediaType string, limit, offset int) ([]models.MediaSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	where := ""
	args := []interface{}{limit, offset}
	var query string
	if mediaType != "" {
		where = "WHERE media_type = $3"
		args = append(args, mediaType)

		query = fmt.Sprintf(`
			SELECT id, media_type, name, slug, image, aggregate_rating, date_published
			FROM media
			%s
			ORDER BY (aggregate_rating * rating_count) DESC, created_at DESC
			LIMIT $1 OFFSET $2
		`, where)
	} else {
		// "All" case: Mix Movies, TV Shows, and People
		query = `
			SELECT id, media_type, name, slug, image, aggregate_rating, year, score
			FROM (
				SELECT id, media_type, name, slug, image, aggregate_rating, date_published as year, (aggregate_rating * rating_count) as score
				FROM media
				UNION ALL
				SELECT id, 'People' as media_type, name, slug, COALESCE(image, ''), 0.0 as aggregate_rating, 
				       COALESCE(EXTRACT(YEAR FROM AGE(NULLIF(birth_date, '')::DATE))::TEXT, '') as year, 
				       (popularity_score * 5 + 38) as score
				FROM people
			) AS mixed
			ORDER BY score DESC, id ASC
			LIMIT $1 OFFSET $2
		`
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.MediaSummary
	for rows.Next() {
		var s models.MediaSummary
		var err error
		if mediaType != "" {
			err = rows.Scan(&s.ID, &s.MediaType, &s.Name, &s.Slug, &s.Image, &s.AggregateRating, &s.Year)
		} else {
			var score float64
			err = rows.Scan(&s.ID, &s.MediaType, &s.Name, &s.Slug, &s.Image, &s.AggregateRating, &s.Year, &score)
		}
		if err != nil {
			return nil, err
		}
		if len(s.Year) >= 4 {
			s.Year = s.Year[:4]
		}
		list = append(list, s)
	}
	return list, nil
}

// GetTrendingPeople returns popular people based on their popularity score,
// mapped to MediaSummary for UI consistency.
func (m *PostgresDBRepo) GetTrendingPeople(limit, offset int) ([]models.Person, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, name, slug, COALESCE(image, ''), COALESCE(known_for_department, ''), COALESCE(birth_date, '')
		FROM people
		ORDER BY popularity_score DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := m.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []models.Person
	for rows.Next() {
		var p models.Person
		err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Image, &p.KnownForDepartment, &p.BirthDate)
		if err != nil {
			return nil, err
		}
		people = append(people, p)
	}
	return people, nil
}

// GetRecentActivity fetches the latest site-wide activities.
func (m *PostgresDBRepo) GetRecentActivity(limit int) ([]models.Activity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT a.id, a.user_id, u.username, COALESCE(u.avatar, '/static/img/default_avatar.webp'), a.activity_type, 
		       COALESCE(a.target_id, 0), COALESCE(a.target_type, ''), a.created_at,
		       COALESCE(m.name, ''), COALESCE(m.slug, ''), COALESCE(m.media_type, '')
		FROM activities a
		JOIN users u ON a.user_id = u.id
		LEFT JOIN media m ON a.target_id = m.id AND a.target_type IN ('Movie', 'TVSeries', 'media')
		ORDER BY a.created_at DESC
		LIMIT $1
	`
	rows, err := m.DB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []models.Activity
	for rows.Next() {
		var a models.Activity
		err := rows.Scan(
			&a.ID, &a.UserID, &a.Username, &a.UserAvatar, &a.ActivityType,
			&a.TargetID, &a.TargetType, &a.CreatedAt,
			&a.TargetName, &a.TargetSlug, &a.TargetMediaType,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}
	return activities, nil
}

// GetRecentBlogPosts fetches the latest blog entries.
func (m *PostgresDBRepo) GetRecentBlogPosts(limit int) ([]models.BlogPost, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT b.id, b.title, b.slug, b.content, COALESCE(b.image, ''), b.is_featured, b.created_at, COALESCE(u.username, 'Admin')
		FROM blog_posts b
		LEFT JOIN users u ON b.author_id = u.id
		ORDER BY b.created_at DESC
		LIMIT $1
	`
	rows, err := m.DB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.BlogPost
	for rows.Next() {
		var p models.BlogPost
		err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Content, &p.Image, &p.IsFeatured, &p.CreatedAt, &p.Author)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

// GetBlogPostBySlug fetches a single blog entry by its identifier.
func (m *PostgresDBRepo) GetBlogPostBySlug(slug string) (models.BlogPost, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT b.id, b.title, b.slug, b.content, COALESCE(b.image, ''), b.is_featured, b.created_at, COALESCE(u.username, 'Admin')
		FROM blog_posts b
		LEFT JOIN users u ON b.author_id = u.id
		WHERE b.slug = $1
	`
	var p models.BlogPost
	err := m.DB.QueryRowContext(ctx, query, slug).Scan(&p.ID, &p.Title, &p.Slug, &p.Content, &p.Image, &p.IsFeatured, &p.CreatedAt, &p.Author)
	return p, err
}

// GetAllBlogPosts returns a paginated list of blog entries.
func (m *PostgresDBRepo) GetAllBlogPosts(limit, offset int) ([]models.BlogPost, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT b.id, b.title, b.slug, b.content, COALESCE(b.image, ''), b.is_featured, b.created_at, COALESCE(u.username, 'Admin')
		FROM blog_posts b
		LEFT JOIN users u ON b.author_id = u.id
		ORDER BY b.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := m.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.BlogPost
	for rows.Next() {
		var p models.BlogPost
		err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Content, &p.Image, &p.IsFeatured, &p.CreatedAt, &p.Author)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

// GetActorBirthdays finds people born on the current day (MM-DD).
func (m *PostgresDBRepo) GetActorBirthdays(date string) ([]models.Person, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, name, slug, image, known_for_department
		FROM people
		WHERE birth_date LIKE '%-' || $1
		ORDER BY popularity_score DESC
		LIMIT 12
	`
	rows, err := m.DB.QueryContext(ctx, query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []models.Person
	for rows.Next() {
		var p models.Person
		err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Image, &p.KnownForDepartment)
		if err != nil {
			return nil, err
		}
		people = append(people, p)
	}
	return people, nil
}

// GetTopBoxOffice returns highest grossing movies.
func (m *PostgresDBRepo) GetTopBoxOffice(limit int) ([]models.MediaSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT m.id, m.media_type, m.name, m.slug, m.image, m.aggregate_rating, m.date_published
		FROM media m
		JOIN movies mov ON m.id = mov.media_id
		WHERE m.media_type = 'Movie' AND mov.box_office IS NOT NULL AND mov.box_office != 'Unknown' AND mov.box_office != ''
		ORDER BY mov.box_office DESC
		LIMIT $1
	`
	rows, err := m.DB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.MediaSummary
	for rows.Next() {
		var s models.MediaSummary
		err := rows.Scan(&s.ID, &s.MediaType, &s.Name, &s.Slug, &s.Image, &s.AggregateRating, &s.Year)
		if err != nil {
			return nil, err
		}
		if len(s.Year) >= 4 {
			s.Year = s.Year[:4]
		}
		list = append(list, s)
	}
	return list, nil
}

// GetFanFavorites returns recently popular movies.
func (m *PostgresDBRepo) GetFanFavorites(limit int) ([]models.MediaSummary, error) {
	return m.GetTrendingMedia("Movie", limit, 0)
}

// GetPopularLists fetches user lists with high interaction.
func (m *PostgresDBRepo) GetPopularLists(limit int) ([]models.List, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT l.id, l.user_id, u.username, l.name, l.slug, l.description, l.item_count, l.like_count
		FROM lists l
		JOIN users u ON l.user_id = u.id
		WHERE l.visibility = 'public'
		ORDER BY (l.like_count + l.follower_count) DESC, l.created_at DESC
		LIMIT $1
	`
	rows, err := m.DB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []models.List
	for rows.Next() {
		var l models.List
		err := rows.Scan(&l.ID, &l.UserID, &l.Username, &l.Name, &l.Slug, &l.Description, &l.ItemCount, &l.LikeCount)
		if err != nil {
			return nil, err
		}
		lists = append(lists, l)
	}
	return lists, nil
}

// GetFranchiseSpotlight picks a random franchise to feature.
func (m *PostgresDBRepo) GetFranchiseSpotlight() (models.Franchise, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, name, slug, COALESCE(description, ''), COALESCE(image, '') FROM franchises ORDER BY RANDOM() LIMIT 1`
	var f models.Franchise
	err := m.DB.QueryRowContext(ctx, query).Scan(&f.ID, &f.Name, &f.Slug, &f.Description, &f.Image)
	return f, err
}

// GetRecentPhotos fetches latest media/person imagery.
func (m *PostgresDBRepo) GetRecentPhotos(limit int) ([]models.Photo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, COALESCE(media_id, 0), COALESCE(person_id, 0), image_url, COALESCE(caption, ''), view_count, created_at FROM photos ORDER BY created_at DESC LIMIT $1`
	rows, err := m.DB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []models.Photo
	for rows.Next() {
		var p models.Photo
		err := rows.Scan(&p.ID, &p.MediaID, &p.PersonID, &p.ImageURL, &p.Caption, &p.ViewCount, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		photos = append(photos, p)
	}
	return photos, nil
}

// CreateBlogPost inserts a new blog entry.
func (m *PostgresDBRepo) CreateBlogPost(p models.BlogPost) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `INSERT INTO blog_posts (title, slug, content, image, author_id, is_featured) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := m.DB.ExecContext(ctx, query, p.Title, p.Slug, p.Content, p.Image, p.AuthorID, p.IsFeatured)
	return err
}

// LogActivity records a user action for the activity feed.
func (m *PostgresDBRepo) LogActivity(a models.Activity) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `INSERT INTO activities (user_id, activity_type, target_id, target_type) VALUES ($1, $2, $3, $4)`
	_, err := m.DB.ExecContext(ctx, query, a.UserID, a.ActivityType, a.TargetID, a.TargetType)
	return err
}

func (m *PostgresDBRepo) CreateFranchise(f models.Franchise) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `INSERT INTO franchises (name, slug, description, image) VALUES ($1, $2, $3, $4)`
	_, err := m.DB.ExecContext(ctx, query, f.Name, f.Slug, f.Description, f.Image)
	return err
}

func (m *PostgresDBRepo) GetFranchiseBySlug(slug string) (models.Franchise, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, name, slug, COALESCE(description, ''), COALESCE(image, '') FROM franchises WHERE slug = $1`
	var f models.Franchise
	err := m.DB.QueryRowContext(ctx, query, slug).Scan(&f.ID, &f.Name, &f.Slug, &f.Description, &f.Image)
	return f, err
}

func (m *PostgresDBRepo) GetFranchiseMedia(franchiseID int) ([]models.MediaSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT m.id, m.media_type, m.name, m.slug, m.image, m.aggregate_rating, m.date_published
		FROM media m
		JOIN media_franchises mf ON m.id = mf.media_id
		WHERE mf.franchise_id = $1
		ORDER BY mf.list_order ASC
	`
	rows, err := m.DB.QueryContext(ctx, query, franchiseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.MediaSummary
	for rows.Next() {
		var s models.MediaSummary
		err := rows.Scan(&s.ID, &s.MediaType, &s.Name, &s.Slug, &s.Image, &s.AggregateRating, &s.Year)
		if err != nil {
			return nil, err
		}
		if len(s.Year) >= 4 {
			s.Year = s.Year[:4]
		}
		list = append(list, s)
	}
	return list, nil
}

func (m *PostgresDBRepo) AddMediaToFranchise(mediaID, franchiseID, order int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `INSERT INTO media_franchises (media_id, franchise_id, list_order) VALUES ($1, $2, $3) ON CONFLICT (media_id, franchise_id) DO UPDATE SET list_order = EXCLUDED.list_order`
	_, err := m.DB.ExecContext(ctx, query, mediaID, franchiseID, order)
	return err
}

func (m *PostgresDBRepo) CreatePhoto(p models.Photo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `INSERT INTO photos (media_id, person_id, image_url, caption) VALUES ($1, $2, $3, $4)`
	mediaID := sql.NullInt64{Int64: int64(p.MediaID), Valid: p.MediaID > 0}
	personID := sql.NullInt64{Int64: int64(p.PersonID), Valid: p.PersonID > 0}
	_, err := m.DB.ExecContext(ctx, query, mediaID, personID, p.ImageURL, p.Caption)
	return err
}

// Social Graph (Follow System)

func (m *PostgresDBRepo) FollowUser(followerID, followedUserID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `INSERT INTO user_follows (follower_id, followed_user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := m.DB.ExecContext(ctx, query, followerID, followedUserID)
	return err
}

func (m *PostgresDBRepo) UnfollowUser(followerID, followedUserID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE FROM user_follows WHERE follower_id = $1 AND followed_user_id = $2`
	_, err := m.DB.ExecContext(ctx, query, followerID, followedUserID)
	return err
}

func (m *PostgresDBRepo) FollowPerson(followerID, personID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `INSERT INTO user_follows (follower_id, followed_person_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := m.DB.ExecContext(ctx, query, followerID, personID)
	return err
}

func (m *PostgresDBRepo) UnfollowPerson(followerID, personID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE FROM user_follows WHERE follower_id = $1 AND followed_person_id = $2`
	_, err := m.DB.ExecContext(ctx, query, followerID, personID)
	return err
}

func (m *PostgresDBRepo) FollowList(followerID, listID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `INSERT INTO user_follows (follower_id, followed_list_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := m.DB.ExecContext(ctx, query, followerID, listID)
	return err
}

func (m *PostgresDBRepo) UnfollowList(followerID, listID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE FROM user_follows WHERE follower_id = $1 AND followed_list_id = $2`
	_, err := m.DB.ExecContext(ctx, query, followerID, listID)
	return err
}

func (m *PostgresDBRepo) IsFollowingUser(followerID, followedUserID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $1 AND followed_user_id = $2)`
	err := m.DB.QueryRowContext(ctx, query, followerID, followedUserID).Scan(&exists)
	return exists, err
}

func (m *PostgresDBRepo) IsFollowingPerson(followerID, personID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $1 AND followed_person_id = $2)`
	err := m.DB.QueryRowContext(ctx, query, followerID, personID).Scan(&exists)
	return exists, err
}

func (m *PostgresDBRepo) IsFollowingList(followerID, listID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $1 AND followed_list_id = $2)`
	err := m.DB.QueryRowContext(ctx, query, followerID, listID).Scan(&exists)
	return exists, err
}

func (m *PostgresDBRepo) GetFollowers(userID int) ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT u.id, u.username, u.email, COALESCE(u.avatar, '/static/img/default_avatar.webp'), u.created_at
		FROM users u
		JOIN user_follows f ON u.id = f.follower_id
		WHERE f.followed_user_id = $1
		ORDER BY f.created_at DESC
	`
	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Avatar, &u.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (m *PostgresDBRepo) GetFollowingUsers(userID int) ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT u.id, u.username, u.email, COALESCE(u.avatar, '/static/img/default_avatar.webp'), u.created_at
		FROM users u
		JOIN user_follows f ON u.id = f.followed_user_id
		WHERE f.follower_id = $1
		ORDER BY f.created_at DESC
	`
	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Avatar, &u.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (m *PostgresDBRepo) GetFollowingPeople(userID int) ([]models.Person, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT p.id, p.name, p.slug, COALESCE(p.image, ''), p.known_for_department
		FROM people p
		JOIN user_follows f ON p.id = f.followed_person_id
		WHERE f.follower_id = $1
		ORDER BY f.created_at DESC
	`
	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []models.Person
	for rows.Next() {
		var p models.Person
		err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Image, &p.KnownForDepartment)
		if err != nil {
			return nil, err
		}
		people = append(people, p)
	}
	return people, nil
}

func (m *PostgresDBRepo) GetFollowCounts(userID int) (followers, following int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT
			(SELECT COUNT(*) FROM user_follows WHERE followed_user_id = $1) as followers,
			(SELECT COUNT(*) FROM user_follows WHERE follower_id = $1) as following
	`
	err = m.DB.QueryRowContext(ctx, query, userID).Scan(&followers, &following)
	return followers, following, err
}

func (m *PostgresDBRepo) GetPersonFollowCounts(personID int) (followers, following int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT COUNT(*) FROM user_follows WHERE followed_person_id = $1`
	err = m.DB.QueryRowContext(ctx, query, personID).Scan(&followers)
	return followers, 0, err
}

func (m *PostgresDBRepo) GetListFollowCounts(listID int) (followers, following int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT COUNT(*) FROM user_follows WHERE followed_list_id = $1`
	err = m.DB.QueryRowContext(ctx, query, listID).Scan(&followers)
	return followers, 0, err
}

// GetActivitiesByFollowed fetches activities from users that the given userID follows.
func (m *PostgresDBRepo) GetActivitiesByFollowed(userID int, limit int) ([]models.Activity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT a.id, a.user_id, u.username, COALESCE(u.avatar, '/static/img/default_avatar.webp'), 
		       a.activity_type, COALESCE(a.target_id, 0), COALESCE(a.target_type, ''), a.created_at,
		       COALESCE(m.name, ''), COALESCE(m.slug, ''), m.media_type
		FROM activities a
		JOIN users u ON a.user_id = u.id
		JOIN user_follows f ON a.user_id = f.followed_user_id
		LEFT JOIN media m ON a.target_id = m.id AND a.target_type IN ('Movie', 'TVSeries', 'media')
		WHERE f.follower_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2
	`
	rows, err := m.DB.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []models.Activity
	for rows.Next() {
		var a models.Activity
		err := rows.Scan(
			&a.ID, &a.UserID, &a.Username, &a.UserAvatar,
			&a.ActivityType, &a.TargetID, &a.TargetType, &a.CreatedAt,
			&a.TargetName, &a.TargetSlug, &a.TargetMediaType,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}
	return activities, nil
}
