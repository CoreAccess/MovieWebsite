package dbrepo

import (
	"database/sql"
	"fmt"
	"log"
	"movieweb/internal/models"

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
	query := `SELECT id, username, email, password_hash, COALESCE(avatar, ''), COALESCE(role, 'user') FROM users WHERE email = $1`
	err := m.DB.QueryRow(query, email).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Avatar, &u.Role)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (m *PostgresDBRepo) GetUserByID(id int) (models.User, error) {
	var u models.User
	query := `SELECT id, username, email, COALESCE(avatar, ''), COALESCE(role, 'user') FROM users WHERE id = $1`
	err := m.DB.QueryRow(query, id).Scan(&u.ID, &u.Username, &u.Email, &u.Avatar, &u.Role)
	return u, err
}

func (m *PostgresDBRepo) GetAllUsers(limit int, offset int) ([]models.User, error) {
	query := `SELECT id, username, email, COALESCE(avatar, ''), COALESCE(role, 'user'), COALESCE(reputation_score, 0), created_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
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
		       COALESCE(m.rating_count, 0), COALESCE(m.review_count, 0), COALESCE(m.best_rating, 10.0), COALESCE(m.worst_rating, 1.0), COALESCE(m.is_family_friendly, TRUE), COALESCE(m.subtitle, '') 
		FROM media m
		JOIN movies s ON m.id = s.media_id
		WHERE m.id = $1 AND m.media_type = 'Movie'
	`
	err := m.DB.QueryRow(query, id).Scan(
		&mov.ID, &mov.Name, &mov.Slug, &mov.DatePublished, &mov.ContentRating, &mov.AggregateRating, &mov.Description, &mov.Image,
		&budget, &boxOffice, &mov.Duration, &langCode, &countryCode, &tagline,
		&mov.RatingCount, &mov.ReviewCount, &mov.BestRating, &mov.WorstRating, &mov.IsFamilyFriendly, &subtitle,
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
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Gender, &p.Description, &p.Image, &p.PopularityScore); err != nil {
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
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Gender, &p.Description, &p.Image, &p.PopularityScore); err != nil {
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

func (m *PostgresDBRepo) GetShowByID(id int) (*models.TVSeries, error) {
	var s models.TVSeries
	var langCode, countryCode, tagline, subtitle sql.NullString
	query := `
		SELECT m.id, m.name, m.slug, COALESCE(m.date_published, ''), COALESCE(m.content_rating, ''), COALESCE(s.end_date, ''), COALESCE(m.aggregate_rating, 0.0), 
		       COALESCE(m.description, ''), COALESCE(m.image, ''), COALESCE(s.number_of_seasons, 0), 
		       COALESCE(m.language_code, 'en'), COALESCE(m.country_code, 'US'), COALESCE(m.tagline, ''), 
		       COALESCE(m.rating_count, 0), COALESCE(m.review_count, 0), COALESCE(m.best_rating, 10.0), COALESCE(m.worst_rating, 1.0), COALESCE(m.subtitle, '') 
		FROM media m
		JOIN tv_series s ON m.id = s.media_id
		WHERE m.id = $1 AND m.media_type = 'TVSeries'
	`
	err := m.DB.QueryRow(query, id).Scan(
		&s.ID, &s.Name, &s.Slug, &s.StartDate, &s.ContentRating, &s.EndDate, &s.AggregateRating,
		&s.Description, &s.Image, &s.NumberOfSeasons, &langCode, &countryCode, &tagline,
		&s.RatingCount, &s.ReviewCount, &s.BestRating, &s.WorstRating, &subtitle,
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
	err := m.DB.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Slug, &p.Gender, &p.Description, &p.Image, &p.PopularityScore)
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
	err := m.DB.QueryRow(query, p.Name, p.Slug, p.Gender, p.Birthday, p.BirthPlace, p.Deathday, p.Description, p.Image, p.KnownForDepartment, p.PopularityScore).Scan(&personID)
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

	movieQuery := `
		SELECT m.id, m.name, m.slug, COALESCE(m.date_published, ''), COALESCE(m.aggregate_rating, 0.0), COALESCE(m.description, ''), COALESCE(m.image, '')
		FROM media m
		JOIN movies s ON m.id = s.media_id
		JOIN watchlist_items wi ON m.id = wi.media_id
		WHERE wi.watchlist_id = $1 AND m.media_type = 'Movie'
		ORDER BY wi.added_at DESC
	`
	movieRows, err := m.DB.Query(movieQuery, watchlistID)
	if err != nil {
		return nil, nil, err
	}
	defer movieRows.Close()

	var movies []models.Movie
	for movieRows.Next() {
		var mov models.Movie
		if err := movieRows.Scan(&mov.ID, &mov.Name, &mov.Slug, &mov.DatePublished, &mov.AggregateRating, &mov.Description, &mov.Image); err != nil {
			return nil, nil, err
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
	showRows, err := m.DB.Query(showQuery, watchlistID)
	if err != nil {
		return nil, nil, err
	}
	defer showRows.Close()

	var shows []models.TVSeries
	for showRows.Next() {
		var s models.TVSeries
		if err := showRows.Scan(&s.ID, &s.Name, &s.Slug, &s.StartDate, &s.EndDate, &s.AggregateRating, &s.Description, &s.Image, &s.NumberOfSeasons); err != nil {
			return nil, nil, err
		}
		shows = append(shows, s)
	}

	return movies, shows, nil
}

func (m *PostgresDBRepo) GetUserWatchlists(userID int) ([]models.Watchlist, error) {
	query := `SELECT id, user_id, name, description, is_public, created_at FROM watchlists WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var watchlists []models.Watchlist
	for rows.Next() {
		var w models.Watchlist
		err := rows.Scan(&w.ID, &w.UserID, &w.Name, &w.Description, &w.IsPublic, &w.CreatedAt)
		if err != nil {
			return nil, err
		}
		watchlists = append(watchlists, w)
	}
	return watchlists, nil
}

func (m *PostgresDBRepo) CreateWatchlist(userID int, name string, description string) error {
	query := `INSERT INTO watchlists (user_id, name, description, is_public) VALUES ($1, $2, $3, $4)`
	_, err := m.DB.Exec(query, userID, name, description, false)
	return err
}

func (m *PostgresDBRepo) AddToWatchlist(watchlistID int, mediaType string, mediaID int) error {
	query := `INSERT INTO watchlist_items (watchlist_id, media_type, media_id) VALUES ($1, $2, $3)`
	_, err := m.DB.Exec(query, watchlistID, mediaType, mediaID)
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
