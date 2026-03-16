package dbrepo

import (
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"movieweb/internal/models"
	"movieweb/internal/tmdb"
)

// seedDataIfEmpty checks if the media table is empty and populates it with TMDB data
func (m *PostgresDBRepo) seedDataIfEmpty(tmdbAPIKey string) {
	var count int
	err := m.DB.QueryRow("SELECT COUNT(*) FROM media").Scan(&count)
	if err != nil {
		log.Printf("Note: media table might not be seeded yet: %v\n", err)
	}

	if count == 0 {
		log.Println("Seeding TMDB data into new Schema.org database (PostgreSQL)...")

		// 1. Seed Users
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
		users := []models.User{
			{Username: "adamd", Email: "adam@example.com", Avatar: "/static/img/avatar1.png", ReputationScore: 50, Role: "admin"},
			{Username: "sarah_k", Email: "sarah@example.com", Avatar: "/static/img/avatar2.png", ReputationScore: 10, Role: "user"},
			{Username: "moviebuff99", Email: "buff99@example.com", Avatar: "/static/img/avatar3.png", ReputationScore: 100, Role: "moderator"},
		}

		query := "INSERT INTO users (username, email, password_hash, avatar, reputation_score, role) VALUES "
		var args []interface{}
		var placeholders []string

		for i, u := range users {
			offset := i * 6
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", offset+1, offset+2, offset+3, offset+4, offset+5, offset+6))
			args = append(args, u.Username, u.Email, string(hashedPassword), u.Avatar, u.ReputationScore, u.Role)
		}

		query += strings.Join(placeholders, ", ") + " ON CONFLICT (email) DO NOTHING"
		_, err = m.DB.Exec(query, args...)
		if err != nil {
			log.Println("Error inserting users in bulk:", err)
		}

		client := tmdb.NewClient(tmdbAPIKey)

		// Pre-seed genres using ON CONFLICT DO NOTHING
		mGenres, err := client.FetchMovieGenres()
		if err == nil {
			for _, g := range mGenres {
				_, _ = m.DB.Exec("INSERT INTO genres (id, name, slug) VALUES ($1, $2, $3) ON CONFLICT (name) DO NOTHING", g.ID, g.Name, tmdb.Slugify(g.Name))
			}
		}
		tGenres, err := client.FetchTVGenres()
		if err == nil {
			for _, g := range tGenres {
				_, _ = m.DB.Exec("INSERT INTO genres (id, name, slug) VALUES ($1, $2, $3) ON CONFLICT (name) DO NOTHING", g.ID, g.Name, tmdb.Slugify(g.Name))
			}
		}

		// Pre-seed languages
		_, _ = m.DB.Exec("INSERT INTO languages (code, name) VALUES ('en', 'English'), ('ja', 'Japanese'), ('ko', 'Korean'), ('es', 'Spanish'), ('fr', 'French') ON CONFLICT (code) DO NOTHING")

		// 2. Fetch and Seed Movies
		movies, err := client.FetchTrendingMovies()
		if err != nil {
			log.Println("Error fetching movies from TMDB:", err)
		} else {
			for _, mov := range movies {
				slug := tmdb.Slugify(mov.Title)
				if slug == "" {
					slug = "movie"
				}
				langCode := mov.OriginalLanguage
				if langCode == "" {
					langCode = "en"
				}
				
				movieModel := models.Movie{}
				movieModel.MediaType = "Movie"
				movieModel.Name = mov.Title
				movieModel.Slug = slug
				movieModel.DatePublished = mov.ReleaseDate
				movieModel.AggregateRating = mov.VoteAverage
				movieModel.Description = mov.Overview
				movieModel.Image = tmdb.ImageBaseURL + mov.PosterPath
				movieModel.TmdbID = mov.ID
				movieModel.Duration = 0
				movieModel.LanguageCode = langCode

				mediaID, err := m.InsertMovie(movieModel)
				if err != nil {
					log.Printf("Error inserting media (movie) %s: %v", mov.Title, err)
					continue
				}

				for _, gID := range mov.GenreIDs {
					_, _ = m.DB.Exec("INSERT INTO media_genres (media_id, genre_id) VALUES ($1, $2) ON CONFLICT (media_id, genre_id) DO NOTHING", mediaID, gID)
				}

				// Fetch Credits
				credits, err := client.FetchMovieCredits(mov.ID)
				if err == nil {
					for _, cast := range credits.Cast {
						personSlug := tmdb.Slugify(cast.Name)
						if personSlug == "" {
							personSlug = "person"
						}
						var image string
						if cast.ProfilePath != "" {
							image = tmdb.ImageBaseURL + cast.ProfilePath
						}
						
						p := models.Person{
							Name: cast.Name,
							Slug: personSlug,
							Gender: fmt.Sprintf("%d", cast.Gender),
							Image: image,
						}
						personID, _ := m.InsertPerson(p)

						characterSlug := tmdb.Slugify(cast.Character)
						if characterSlug == "" {
							characterSlug = "character"
						}
						var charID int
						err = m.DB.QueryRow("SELECT id FROM characters WHERE name = $1", cast.Character).Scan(&charID)
						if err != nil {
							err = m.DB.QueryRow("INSERT INTO characters (name, slug, gender) VALUES ($1, $2, $3) RETURNING id", cast.Character, characterSlug, cast.Gender).Scan(&charID)
						}

						_ = m.InsertMediaCast(mediaID, personID, cast.Character, cast.Order)
					}

					for _, crew := range credits.Crew {
						if crew.Job == "Director" || crew.Job == "Writer" || crew.Job == "Screenplay" || crew.Job == "Author" {
							crewSlug := tmdb.Slugify(crew.Name)
							if crewSlug == "" {
								crewSlug = "crew"
							}
							var image string
							if crew.ProfilePath != "" {
								image = tmdb.ImageBaseURL + crew.ProfilePath
							}
							
							p := models.Person{
								Name: crew.Name,
								Slug: crewSlug,
								Gender: "0",
								Image: image,
							}
							personID, _ := m.InsertPerson(p)

							job := "writer"
							if crew.Job == "Director" {
								job = "director"
							}
							_, _ = m.DB.Exec("INSERT INTO media_crew (media_id, person_id, job, department) VALUES ($1, $2, $3, $4)", mediaID, personID, job, "production")
						}
					}
				}
			}
		}

		// 3. Fetch and Seed TV Shows
		shows, err := client.FetchTrendingShows()
		if err != nil {
			log.Println("Error fetching shows from TMDB:", err)
		} else {
			for _, s := range shows {
				slug := tmdb.Slugify(s.Name)
				if slug == "" {
					slug = "show"
				}

				showModel := models.TVSeries{}
				showModel.MediaType = "TVSeries"
				showModel.Name = s.Name
				showModel.Slug = slug
				showModel.StartDate = s.FirstAirDate
				showModel.AggregateRating = s.VoteAverage
				showModel.Description = s.Overview
				showModel.Image = tmdb.ImageBaseURL + s.PosterPath
				showModel.TmdbID = s.ID
				showModel.NumberOfSeasons = 1
				showModel.LanguageCode = s.OriginalLanguage

				mediaID, err := m.InsertShow(showModel)
				if err != nil {
					log.Printf("Error inserting media (show) %s: %v", s.Name, err)
					continue
				}

				for _, gID := range s.GenreIDs {
					_, _ = m.DB.Exec("INSERT INTO media_genres (media_id, genre_id) VALUES ($1, $2) ON CONFLICT (media_id, genre_id) DO NOTHING", mediaID, gID)
				}

				// Fetch Credits
				credits, err := client.FetchTVCredits(s.ID)
				if err == nil {
					for _, cast := range credits.Cast {
						personSlug := tmdb.Slugify(cast.Name)
						if personSlug == "" {
							personSlug = "person"
						}
						var image string
						if cast.ProfilePath != "" {
							image = tmdb.ImageBaseURL + cast.ProfilePath
						}
						
						p := models.Person{
							Name: cast.Name,
							Slug: personSlug,
							Gender: fmt.Sprintf("%d", cast.Gender),
							Image: image,
						}
						personID, _ := m.InsertPerson(p)

						characterSlug := tmdb.Slugify(cast.Character)
						if characterSlug == "" {
							characterSlug = "character"
						}
						var charID int
						err = m.DB.QueryRow("SELECT id FROM characters WHERE name = $1", cast.Character).Scan(&charID)
						if err != nil {
							err = m.DB.QueryRow("INSERT INTO characters (name, slug, gender) VALUES ($1, $2, $3) RETURNING id", cast.Character, characterSlug, cast.Gender).Scan(&charID)
						}

						_ = m.InsertMediaCast(mediaID, personID, cast.Character, cast.Order)
					}

					for _, crew := range credits.Crew {
						if crew.Job == "Executive Producer" || crew.Job == "Creator" || crew.Job == "Writer" {
							crewSlug := tmdb.Slugify(crew.Name)
							if crewSlug == "" {
								crewSlug = "crew"
							}
							var image string
							if crew.ProfilePath != "" {
								image = tmdb.ImageBaseURL + crew.ProfilePath
							}
							
							p := models.Person{
								Name: crew.Name,
								Slug: crewSlug,
								Gender: "0",
								Image: image,
							}
							personID, _ := m.InsertPerson(p)

							job := "writer"
							if strings.Contains(crew.Job, "Producer") {
								job = "director" 
							}

							_, _ = m.DB.Exec("INSERT INTO media_crew (media_id, person_id, job, department) VALUES ($1, $2, $3, $4)", mediaID, personID, job, "production")
						}
					}
				}

				// Fetch Episodes for Season 1
				episodes, err := client.FetchTVSeasonEpisodes(s.ID, 1)
				if err == nil && len(episodes) > 0 {
					for _, ep := range episodes {
						epSlug := tmdb.Slugify(ep.Name)
						if epSlug == "" {
							epSlug = fmt.Sprintf("episode-%d", ep.EpisodeNumber)
						}
						epSlug = fmt.Sprintf("%s-%s", slug, epSlug)

						var image string
						if ep.StillPath != "" {
							image = tmdb.ImageBaseURL + ep.StillPath
						}

						queryEp := `INSERT INTO tv_episodes (series_id, season_number, episode_number, name, slug, date_published, description, image, duration)
							VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT (series_id, season_number, episode_number) DO NOTHING`
						_, err := m.DB.Exec(queryEp, mediaID, ep.SeasonNumber, ep.EpisodeNumber, ep.Name, epSlug, ep.AirDate, ep.Overview, image, ep.Runtime)

						if err != nil {
							log.Printf("Error inserting episode for series %s: %v", s.Name, err)
						}
					}
				}
			}
		}
	}
}
