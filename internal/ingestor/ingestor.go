// Package ingestor implements the background TMDB data ingestion service.
// It operates completely independently of the web server and communicates
// solely through the shared PostgreSQL database.
package ingestor

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"filmgap/internal/models"
	"filmgap/internal/repository"
	"filmgap/internal/tmdb"

	"golang.org/x/time/rate"
)

// Config controls how aggressively the ingestor queries TMDB.
type Config struct {
	// RequestsPerSecond is the maximum rate of outbound TMDB API calls.
	// TMDB advertises ~40 req/sec; we default to 15 to stay well clear.
	RequestsPerSecond float64
	// BurstSize is the maximum burst above the steady-state rate.
	BurstSize int
	// PollInterval is how long the processing loop sleeps when the queue is empty.
	PollInterval time.Duration
	// MaxAttempts is the number of times a job is retried before being marked FAILED.
	MaxAttempts int
}

// DefaultConfig returns a conservative configuration appropriate for development.
func DefaultConfig() Config {
	return Config{
		RequestsPerSecond: 10,
		BurstSize:         5,
		PollInterval:      3 * time.Second,
		MaxAttempts:       3,
	}
}

// Ingestor orchestrates the two main loops:
//  1. Discovery  — finds new TMDB IDs and adds them to pending_ingestion.
//  2. Processing — pulls QUEUED jobs, fetches full metadata, and writes to Postgres.
type Ingestor struct {
	repo    repository.DatabaseRepo
	client  *tmdb.Client
	limiter *rate.Limiter
	cfg     Config
	log     *slog.Logger
}

// New creates an Ingestor ready to run.
func New(repo repository.DatabaseRepo, apiKey string, cfg Config, log *slog.Logger) *Ingestor {
	lim := rate.NewLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.BurstSize)
	return &Ingestor{
		repo:    repo,
		client:  tmdb.NewClient(apiKey),
		limiter: lim,
		cfg:     cfg,
		log:     log,
	}
}

// Run starts both loops and blocks until ctx is cancelled.
// On startup it resets any lingering PROCESSING jobs from a previous ungraceful
// shutdown back to QUEUED so they are retried rather than left stuck.
func (ing *Ingestor) Run(ctx context.Context) {
	// Crash recovery: reset any job that was mid-flight when we last died.
	if err := ing.repo.ResetStuckIngestionJobs(); err != nil {
		ing.log.Warn("failed to reset stuck jobs", "err", err)
	} else {
		ing.log.Info("stuck job recovery complete")
	}

	ing.log.Info("ingestor starting",
		"rate_per_sec", ing.cfg.RequestsPerSecond,
		"poll_interval", ing.cfg.PollInterval,
	)

	// Discovery loop: find new IDs on TMDB and queue them for processing.
	go ing.discoveryLoop(ctx)

	// Processing loop: consume QUEUED jobs and persist full detail to Postgres.
	ing.processingLoop(ctx)

	ing.log.Info("ingestor stopped")
}

// ---------------------------------------------------------------------------
// Discovery loop — newest first
// ---------------------------------------------------------------------------

func (ing *Ingestor) discoveryLoop(ctx context.Context) {
	moviePage := 1
	tvPage := 1

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Discover movies
		newMovies, done := ing.discoverMoviePage(ctx, moviePage)
		if done {
			ing.log.Info("movie discovery pass complete", "pages_scanned", moviePage)
			moviePage = 1
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Minute):
			}
		} else {
			if newMovies > 0 {
				ing.log.Info("queued new movies", "count", newMovies, "page", moviePage)
			}
			moviePage++
			// TMDB API caps /discover at 500 pages. Exceeding this returns 400 Bad Request.
			if moviePage > 500 {
				ing.log.Info("movie discovery hit TMDB page limit (500), resetting")
				moviePage = 1
				select {
				case <-ctx.Done():
					return
				case <-time.After(10 * time.Minute):
				}
			}
		}

		// Discover TV shows
		newTV, doneTv := ing.discoverTVPage(ctx, tvPage)
		if doneTv {
			ing.log.Info("tv discovery pass complete", "pages_scanned", tvPage)
			tvPage = 1
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Minute):
			}
		} else {
			if newTV > 0 {
				ing.log.Info("queued new tv shows", "count", newTV, "page", tvPage)
			}
			tvPage++
			// TMDB API caps /discover at 500 pages. Exceeding this returns 400 Bad Request.
			if tvPage > 500 {
				ing.log.Info("tv discovery hit TMDB page limit (500), resetting")
				tvPage = 1
				select {
				case <-ctx.Done():
					return
				case <-time.After(10 * time.Minute):
				}
			}
		}
	}
}

func (ing *Ingestor) discoverMoviePage(ctx context.Context, page int) (int, bool) {
	if err := ing.limiter.Wait(ctx); err != nil {
		return 0, true
	}
	movies, err := ing.client.DiscoverMovies(page)
	if err != nil {
		ing.log.Error("discover movies failed", "page", page, "err", err)
		return 0, false
	}
	if len(movies) == 0 {
		return 0, true
	}
	count := 0
	for _, m := range movies {
		if m.PosterPath == "" || m.Overview == "" {
			continue
		}
		if queued := ing.enqueue(m.ID, "Movie"); queued {
			count++
		}
	}
	return count, false
}

func (ing *Ingestor) discoverTVPage(ctx context.Context, page int) (int, bool) {
	if err := ing.limiter.Wait(ctx); err != nil {
		return 0, true
	}
	shows, err := ing.client.DiscoverTV(page)
	if err != nil {
		ing.log.Error("discover tv failed", "page", page, "err", err)
		return 0, false
	}
	if len(shows) == 0 {
		return 0, true
	}
	count := 0
	for _, s := range shows {
		if s.PosterPath == "" || s.Overview == "" {
			continue
		}
		if queued := ing.enqueue(s.ID, "TV"); queued {
			count++
		}
	}
	return count, false
}

func (ing *Ingestor) enqueue(tmdbID int, mediaType string) bool {
	_, err := ing.repo.EnqueueIngestion(tmdbID, mediaType)
	return err == nil
}

// ---------------------------------------------------------------------------
// Processing loop — transactional full metadata fetch
// ---------------------------------------------------------------------------

func (ing *Ingestor) processingLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		job, err := ing.repo.ClaimNextIngestionJob(ing.cfg.MaxAttempts)
		if err != nil {
			if err == sql.ErrNoRows {
				select {
				case <-ctx.Done():
					return
				case <-time.After(ing.cfg.PollInterval):
				}
				continue
			}
			ing.log.Error("failed to claim ingestion job", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}

		ing.processJob(ctx, job)
	}
}

func (ing *Ingestor) processJob(ctx context.Context, job models.IngestionJob) {
	ing.log.Info("processing job", "tmdb_id", job.TmdbID, "type", job.MediaType)

	var processErr error
	if job.MediaType == "Movie" {
		processErr = ing.processMovie(ctx, job.TmdbID)
	} else {
		processErr = ing.processTV(ctx, job.TmdbID)
	}

	if processErr != nil {
		ing.log.Error("job failed", "tmdb_id", job.TmdbID, "type", job.MediaType, "err", processErr)
		_ = ing.repo.FailIngestionJob(job.ID)
	} else {
		ing.log.Info("job completed", "tmdb_id", job.TmdbID, "type", job.MediaType)
		_ = ing.repo.CompleteIngestionJob(job.ID)
	}

	// Jitter / Cooldown: Add a small randomized break between jobs so it feels
	// more like a "background" process and less like a rapid-fire attack.
	select {
	case <-ctx.Done():
	case <-time.After(100 * time.Millisecond):
	}
}

// processMovie fetches full TMDB metadata and persists ALL writes inside a
// single database transaction. If the process is killed at any point before
// the transaction commits, Postgres automatically rolls back every write —
// leaving the database in a clean state with no partial records.
func (ing *Ingestor) processMovie(ctx context.Context, tmdbID int) error {
	// Rate-limit: detail fetch.
	if err := ing.limiter.Wait(ctx); err != nil {
		return err
	}
	detail, err := ing.client.FetchMovieDetail(tmdbID)
	if err != nil {
		return err
	}

	// Skip empty-shell records that have no poster or overview.
	if detail.PosterPath == "" || detail.Overview == "" {
		return fmt.Errorf("tmdb_id %d has no poster or overview — skipping", tmdbID)
	}

	// Rate-limit: credits fetch.
	if err := ing.limiter.Wait(ctx); err != nil {
		return err
	}
	credits, err := ing.client.FetchMovieCredits(tmdbID)
	if err != nil {
		return err
	}

	// Rate-limit: certification fetch.
	if err := ing.limiter.Wait(ctx); err != nil {
		return err
	}
	rating, _ := ing.client.FetchMovieCertification(tmdbID, "US")

	// --- All TMDB fetches are done. Now open a single DB transaction. ---
	db := ing.repo.Connection()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// If anything below fails, the deferred Rollback is a no-op after Commit.
	defer tx.Rollback() //nolint:errcheck

	// Generate a unique slug by appending the release year if available.
	// This prevents remakes/reboots from colliding on the (slug, media_type) constraint.
	year := "0000"
	if len(detail.ReleaseDate) >= 4 {
		year = detail.ReleaseDate[:4]
	}
	slug := tmdb.Slugify(fmt.Sprintf("%s %s", detail.Title, year))

	// Upsert media supertype row.
	var mediaID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO media (media_type, name, slug, description, image, date_published, aggregate_rating, tmdb_id, content_rating, subtitle, language_code, country_code)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (tmdb_id) DO UPDATE SET
			name             = EXCLUDED.name,
			slug             = EXCLUDED.slug,
			description      = EXCLUDED.description,
			image            = EXCLUDED.image,
			date_published   = EXCLUDED.date_published,
			aggregate_rating = EXCLUDED.aggregate_rating,
			content_rating   = EXCLUDED.content_rating,
			subtitle         = EXCLUDED.subtitle,
			language_code    = EXCLUDED.language_code,
			country_code     = EXCLUDED.country_code
		RETURNING id
	`, "Movie", detail.Title, slug, detail.Overview,
		tmdb.ImageBaseURL+detail.PosterPath, detail.ReleaseDate, detail.VoteAverage, tmdbID, rating,
		"", detail.OriginalLanguage, strings.Join(detail.OriginCountry, ","),
	).Scan(&mediaID)
	if err != nil {
		return err
	}

	// Upsert movies subtype row.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO movies (media_id, duration, tagline, budget, box_office)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (media_id) DO UPDATE SET 
			duration = EXCLUDED.duration, 
			tagline = EXCLUDED.tagline,
			budget = EXCLUDED.budget,
			box_office = EXCLUDED.box_office
	`, mediaID, detail.Runtime, detail.Tagline, fmt.Sprintf("%d", detail.Budget), fmt.Sprintf("%d", detail.Revenue))
	if err != nil {
		return err
	}

	// Upsert genres.
	for _, g := range detail.Genres {
		_, err = tx.ExecContext(ctx, `INSERT INTO genres (id, name, slug) VALUES ($1,$2,$3) ON CONFLICT (id) DO NOTHING`, g.ID, g.Name, g.Name)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO media_genres (media_id, genre_id) VALUES ($1,$2) ON CONFLICT (media_id, genre_id) DO NOTHING`, mediaID, g.ID)
		if err != nil {
			return err
		}
	}

	// Upsert top cast (up to 10).
	for i, c := range credits.Cast {
		if i >= 10 {
			break
		}

		if err := ing.limiter.Wait(ctx); err != nil {
			return err
		}
		personDetail, err := ing.client.FetchPersonDetail(c.ID)
		if err != nil {
			ing.log.Warn("failed to fetch person detail", "tmdb_id", c.ID, "err", err)
			continue // Skip if fetch fails to avoid breaking ingestion
		}

		profileImg := ""
		if personDetail.ProfilePath != "" {
			profileImg = tmdb.ImageBaseURL + personDetail.ProfilePath
		} else if c.ProfilePath != "" {
			profileImg = tmdb.ImageBaseURL + c.ProfilePath
		}

		gender := ""
		if c.Gender == 1 {
			gender = "Female"
		} else if c.Gender == 2 {
			gender = "Male"
		} else if c.Gender == 3 {
			gender = "Non-binary"
		}

		var personID int
		err = tx.QueryRowContext(ctx, `
			INSERT INTO people (name, slug, gender, birth_date, death_date, birth_place, description, image, known_for_department, popularity_score) 
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			ON CONFLICT (slug) DO UPDATE SET 
				name=EXCLUDED.name, 
				gender=EXCLUDED.gender,
				birth_date=EXCLUDED.birth_date,
				death_date=EXCLUDED.death_date,
				birth_place=EXCLUDED.birth_place,
				description=EXCLUDED.description,
				image=CASE WHEN EXCLUDED.image!='' THEN EXCLUDED.image ELSE people.image END,
				known_for_department=EXCLUDED.known_for_department,
				popularity_score=EXCLUDED.popularity_score
			RETURNING id
		`, c.Name, tmdb.Slugify(c.Name), gender, personDetail.Birthday, personDetail.Deathday, personDetail.PlaceOfBirth, personDetail.Biography, profileImg, personDetail.KnownForDepartment, personDetail.Popularity).Scan(&personID)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO media_cast (media_id, person_id, character_name, list_order)
			VALUES ($1,$2,$3,$4)
			ON CONFLICT (media_id, person_id, character_name) DO NOTHING
		`, mediaID, personID, c.Character, c.Order)
		if err != nil {
			return err
		}
	}

	// All writes succeeded — commit atomically.
	return tx.Commit()
}

// processTV is identical to processMovie but for TV series.
func (ing *Ingestor) processTV(ctx context.Context, tmdbID int) error {
	if err := ing.limiter.Wait(ctx); err != nil {
		return err
	}
	detail, err := ing.client.FetchTVDetail(tmdbID)
	if err != nil {
		return err
	}

	if detail.PosterPath == "" || detail.Overview == "" {
		return fmt.Errorf("tmdb_id %d (TV) has no poster or overview — skipping", tmdbID)
	}

	if err := ing.limiter.Wait(ctx); err != nil {
		return err
	}
	credits, err := ing.client.FetchTVCredits(tmdbID)
	if err != nil {
		return err
	}

	db := ing.repo.Connection()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	// Generate a unique slug by appending the release year.
	year := "0000"
	if len(detail.FirstAirDate) >= 4 {
		year = detail.FirstAirDate[:4]
	}
	slug := tmdb.Slugify(fmt.Sprintf("%s %s", detail.Name, year))

	var mediaID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO media (media_type, name, slug, description, image, date_published, aggregate_rating, tmdb_id, subtitle, language_code, country_code)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (tmdb_id) DO UPDATE SET
			name             = EXCLUDED.name,
			slug             = EXCLUDED.slug,
			description      = EXCLUDED.description,
			image            = EXCLUDED.image,
			date_published   = EXCLUDED.date_published,
			aggregate_rating = EXCLUDED.aggregate_rating,
			subtitle         = EXCLUDED.subtitle,
			language_code    = EXCLUDED.language_code,
			country_code     = EXCLUDED.country_code
		RETURNING id
	`, "TVSeries", detail.Name, slug, detail.Overview,
		tmdb.ImageBaseURL+detail.PosterPath, detail.FirstAirDate, detail.VoteAverage, tmdbID,
		"", detail.OriginalLanguage, strings.Join(detail.OriginCountry, ","),
	).Scan(&mediaID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO tv_series (media_id, number_of_seasons, end_date, number_of_episodes) 
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (media_id) DO UPDATE SET 
			number_of_seasons = EXCLUDED.number_of_seasons,
			end_date = EXCLUDED.end_date,
			number_of_episodes = EXCLUDED.number_of_episodes
	`, mediaID, detail.NumberOfSeasons, detail.LastAirDate, detail.NumberOfEpisodes)
	if err != nil {
		return err
	}

	for _, g := range detail.Genres {
		_, err = tx.ExecContext(ctx, `INSERT INTO genres (id, name, slug) VALUES ($1,$2,$3) ON CONFLICT (id) DO NOTHING`, g.ID, g.Name, g.Name)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO media_genres (media_id, genre_id) VALUES ($1,$2) ON CONFLICT (media_id, genre_id) DO NOTHING`, mediaID, g.ID)
		if err != nil {
			return err
		}
	}

	for i, c := range credits.Cast {
		if i >= 10 {
			break
		}
		profileImg := ""
		if c.ProfilePath != "" {
			profileImg = tmdb.ImageBaseURL + c.ProfilePath
		}
		var personID int
		err = tx.QueryRowContext(ctx, `
			INSERT INTO people (name, slug, image) VALUES ($1,$2,$3)
			ON CONFLICT (slug) DO UPDATE SET name=EXCLUDED.name, image=CASE WHEN EXCLUDED.image!='' THEN EXCLUDED.image ELSE people.image END
			RETURNING id
		`, c.Name, tmdb.Slugify(c.Name), profileImg).Scan(&personID)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO media_cast (media_id, person_id, character_name, list_order)
			VALUES ($1,$2,$3,$4)
			ON CONFLICT (media_id, person_id, character_name) DO NOTHING
		`, mediaID, personID, c.Character, c.Order)
		if err != nil {
			return err
		}
	}

	// Fetch and insert TV Seasons
	for seasonNum := 1; seasonNum <= detail.NumberOfSeasons; seasonNum++ {
		if err := ing.limiter.Wait(ctx); err != nil {
			return err
		}
		seasonDetail, err := ing.client.FetchTVSeason(tmdbID, seasonNum)
		if err != nil {
			ing.log.Warn("failed to fetch tv season detail", "tmdb_id", tmdbID, "season_number", seasonNum, "err", err)
			continue
		}

		seasonPoster := ""
		if seasonDetail.PosterPath != "" {
			seasonPoster = tmdb.ImageBaseURL + seasonDetail.PosterPath
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO tv_seasons (series_id, season_number, name, description, image, date_published, episode_count, aggregate_rating)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			ON CONFLICT (series_id, season_number) DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				image = EXCLUDED.image,
				date_published = EXCLUDED.date_published,
				episode_count = EXCLUDED.episode_count,
				aggregate_rating = EXCLUDED.aggregate_rating
		`, mediaID, seasonNum, seasonDetail.Name, seasonDetail.Overview, seasonPoster, seasonDetail.AirDate, len(seasonDetail.Episodes), seasonDetail.VoteAverage)
		if err != nil {
			return err
		}

		for _, ep := range seasonDetail.Episodes {
			epStill := ""
			if ep.StillPath != "" {
				epStill = tmdb.ImageBaseURL + ep.StillPath
			}

			epSlug := tmdb.Slugify(fmt.Sprintf("%s season %d episode %d", detail.Name, seasonNum, ep.EpisodeNumber))

			_, err = tx.ExecContext(ctx, `
				INSERT INTO tv_episodes (series_id, season_number, episode_number, name, slug, date_published, description, image, duration)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
				ON CONFLICT (series_id, season_number, episode_number) DO UPDATE SET
					name = EXCLUDED.name,
					slug = EXCLUDED.slug,
					date_published = EXCLUDED.date_published,
					description = EXCLUDED.description,
					image = EXCLUDED.image,
					duration = EXCLUDED.duration
			`, mediaID, seasonNum, ep.EpisodeNumber, ep.Name, epSlug, ep.AirDate, ep.Overview, epStill, ep.Runtime)
			if err != nil {
				return err
			}
		}
	}

	// All writes succeeded — commit atomically.
	return tx.Commit()
}
