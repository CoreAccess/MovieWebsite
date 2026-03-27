package service

import (
	"bytes"
	"database/sql"
	"filmgap/internal/metadata"
	"filmgap/internal/models"
	"filmgap/internal/repository"
	"filmgap/internal/tmdb"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
)

// AppService orchestrates complex business logic, abstracting
// repository calls (and later Vector DB calls) away from HTTP handlers.
type AppService struct {
	Repo repository.DatabaseRepo
	// Future: VectorRepo repository.VectorRepo
	cache      *expirable.LRU[string, any]
	tmdbClient *tmdb.Client
}

// NewAppService creates a new configured service layer.
func NewAppService(repo repository.DatabaseRepo, tmdbToken string) *AppService {
	// Initialize LRU cache with a maximum of 5000 items and a default TTL of 5 minutes.
	// This prevents memory leaks by evicting older/expired entries.
	cache := expirable.NewLRU[string, any](5000, nil, 5*time.Minute)

	var client *tmdb.Client
	if tmdbToken != "" {
		client = tmdb.NewClient(tmdbToken)
	}

	return &AppService{
		Repo:       repo,
		cache:      cache,
		tmdbClient: client,
	}
}

// GetMovieDetail fetches a movie, its cast, and its generated JSON-LD.
// This ensures handlers do not manually stitch these concepts together.
func (s *AppService) GetMovieDetail(id int, baseDomain string) (*models.Movie, []models.CastMember, []models.CrewMember, []models.Genre, string, error) {
	movie, err := s.Repo.GetMovieByID(id)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	cast, err := s.Repo.GetCastForMedia(movie.ID)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	crew, err := s.Repo.GetCrewForMedia(movie.ID)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	genres, _ := s.Repo.GetMediaGenres(movie.ID)

	// Generate the Schema.org payload for AI Agents and SEO
	jsonld, err := metadata.GenerateMovieJSONLD(*movie, baseDomain)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	return movie, cast, crew, genres, jsonld, nil
}

func (s *AppService) GetWatchProviderGroups(mediaID int, mediaType string, tmdbID int, countryCode string) ([]models.WatchProviderGroup, error) {
	countryCode = normalizeProviderCountry(countryCode)

	providers, err := s.Repo.GetWatchProviders(mediaID, countryCode)
	if err != nil {
		return nil, err
	}

	needsRefresh := len(providers) == 0
	if !needsRefresh {
		needsRefresh = time.Since(providers[0].UpdatedAt) > 24*time.Hour
	}

	if needsRefresh && s.tmdbClient != nil && tmdbID > 0 {
		if fresh, fetchErr := s.fetchProvidersFromTMDB(mediaType, tmdbID, mediaID, countryCode); fetchErr == nil {
			if replaceErr := s.Repo.ReplaceWatchProviders(mediaID, countryCode, fresh); replaceErr == nil {
				providers = fresh
			}
		}
	}

	return groupWatchProviders(providers), nil
}

// GetShowDetail fetches a TV show, its cast, and its generated JSON-LD.
func (s *AppService) GetShowDetail(id int, baseDomain string) (*models.TVSeries, []models.CastMember, []models.CrewMember, []models.TVEpisode, []models.Genre, string, error) {
	show, err := s.Repo.GetShowByID(id)
	if err != nil {
		return nil, nil, nil, nil, nil, "", err
	}

	cast, err := s.Repo.GetCastForMedia(show.ID)
	if err != nil {
		return nil, nil, nil, nil, nil, "", err
	}

	crew, err := s.Repo.GetCrewForMedia(show.ID)
	if err != nil {
		return nil, nil, nil, nil, nil, "", err
	}

	genres, _ := s.Repo.GetMediaGenres(show.ID)

	jsonld, err := metadata.GenerateTVSeriesJSONLD(*show, baseDomain)
	if err != nil {
		return nil, nil, nil, nil, nil, "", err
	}

	episodes, err := s.Repo.GetTVEpisodes(id)
	if err != nil {
		return nil, nil, nil, nil, nil, "", err
	}

	return show, cast, crew, episodes, genres, jsonld, nil
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
	cacheKey := fmt.Sprintf("popular_movies:%d", limit)
	if val, ok := s.getFromCache(cacheKey); ok {
		return val.([]models.Movie), nil
	}

	movies, err := s.Repo.GetPopularMovies(limit)
	if err == nil {
		s.setToCache(cacheKey, movies, 5*time.Minute)
	}
	return movies, err
}

func (s *AppService) GetUpcomingMovies(limit int) ([]models.Movie, error) {
	cacheKey := fmt.Sprintf("upcoming_movies:%d", limit)
	if val, ok := s.getFromCache(cacheKey); ok {
		return val.([]models.Movie), nil
	}

	movies, err := s.Repo.GetUpcomingMovies(limit)
	if err == nil {
		s.setToCache(cacheKey, movies, 5*time.Minute)
	}
	return movies, err
}

func (s *AppService) GetPopularShows(limit int) ([]models.TVSeries, error) {
	cacheKey := fmt.Sprintf("popular_shows:%d", limit)
	if val, ok := s.getFromCache(cacheKey); ok {
		return val.([]models.TVSeries), nil
	}

	shows, err := s.Repo.GetPopularShows(limit)
	if err == nil {
		s.setToCache(cacheKey, shows, 5*time.Minute)
	}
	return shows, err
}

func (s *AppService) GetNewShows(limit int) ([]models.TVSeries, error) {
	cacheKey := fmt.Sprintf("new_shows:%d", limit)
	if val, ok := s.getFromCache(cacheKey); ok {
		return val.([]models.TVSeries), nil
	}

	shows, err := s.Repo.GetNewShows(limit)
	if err == nil {
		s.setToCache(cacheKey, shows, 5*time.Minute)
	}
	return shows, err
}

func (s *AppService) getFromCache(key string) (any, bool) {
	return s.cache.Get(key)
}

func (s *AppService) setToCache(key string, value any, ttl time.Duration) {
	s.cache.Add(key, value)
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

func (s *AppService) GetWatchlistByID(id int) (models.Watchlist, []models.Movie, []models.TVSeries, error) {
	return s.Repo.GetWatchlistByID(id)
}

func (s *AppService) GetUserWatchlists(userID int) ([]models.Watchlist, error) {
	return s.Repo.GetUserWatchlists(userID)
}

func (s *AppService) CreateWatchlist(userID int, name, description string) (int, error) {
	return s.Repo.CreateWatchlist(userID, name, description)
}

func (s *AppService) ToggleWatchlist(userID int, mediaType string, mediaID int, add bool) error {
	watchlists, err := s.Repo.GetUserWatchlists(userID)
	var watchlistID int
	if err != nil || len(watchlists) == 0 {
		watchlistID, err = s.Repo.CreateWatchlist(userID, "My Watchlist", "My personal collection")
		if err != nil {
			return err
		}
	} else {
		watchlistID = watchlists[0].ID
	}

	if add {
		return s.Repo.AddToWatchlist(watchlistID, mediaType, mediaID)
	}
	return s.Repo.RemoveFromWatchlist(watchlistID, mediaID)
}

// List Operations

func (s *AppService) CreateList(list models.List) (int, error) {
	return s.Repo.CreateList(list)
}

func (s *AppService) UpdateList(list models.List) error {
	return s.Repo.UpdateList(list)
}

func (s *AppService) DeleteList(id int) error {
	return s.Repo.DeleteList(id)
}

func (s *AppService) GetListByID(id int) (models.List, error) {
	return s.Repo.GetListByID(id)
}

func (s *AppService) GetListBySlug(userID int, slug string) (models.List, error) {
	return s.Repo.GetListBySlug(userID, slug)
}

func (s *AppService) GetListsByUserID(userID int) ([]models.List, error) {
	return s.Repo.GetListsByUserID(userID)
}

func (s *AppService) GetPublicLists(limit, offset int) ([]models.List, error) {
	return s.Repo.GetPublicLists(limit, offset)
}

func (s *AppService) GetListDetail(userID int, slug string) (models.List, []models.ListItem, error) {
	list, err := s.Repo.GetListBySlug(userID, slug)
	if err != nil {
		return models.List{}, nil, err
	}

	items, err := s.Repo.GetListItems(list.ID)
	if err != nil {
		return list, nil, err
	}

	return list, items, nil
}

func (s *AppService) GetFranchiseDetail(slug string) (models.Franchise, []models.MediaSummary, error) {
	franchise, err := s.Repo.GetFranchiseBySlug(slug)
	if err != nil {
		return models.Franchise{}, nil, err
	}

	media, err := s.Repo.GetFranchiseMedia(franchise.ID)
	if err != nil {
		return franchise, nil, err
	}

	return franchise, media, nil
}

func (s *AppService) AddListItem(item models.ListItem) error {
	return s.Repo.AddListItem(item)
}

func (s *AppService) RemoveListItem(listID, mediaID int) error {
	return s.Repo.RemoveListItem(listID, mediaID)
}

func (s *AppService) UpdateListItemRank(listID, mediaID, rank int) error {
	return s.Repo.UpdateListItemRank(listID, mediaID, rank)
}

// Review & Rating Operations

func (s *AppService) SubmitReview(r models.Review) error {
	if err := s.Repo.UpsertReview(r); err != nil {
		return err
	}

	// Log the activity for the feed
	s.LogActivity(r.UserID, "review_post", r.MediaID, "media")

	// Side-effect: recompute aggregate_rating and rating_count on the media record.
	// This keeps the denormalised summary columns in sync after every review upsert.
	return s.Repo.RecalculateMediaRating(r.MediaID)
}

func (s *AppService) GetReviewsForMedia(mediaID int, limit, offset int) ([]models.Review, error) {
	return s.Repo.GetReviewsForMedia(mediaID, limit, offset)
}

func (s *AppService) GetUserReviewForMedia(userID, mediaID int) (*models.Review, error) {
	return s.Repo.GetUserReviewForMedia(userID, mediaID)
}

func (s *AppService) DeleteReview(userID, mediaID int) error {
	return s.Repo.DeleteReview(userID, mediaID)
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

func (s *AppService) GetHomepageHero(host string) (any, error) {
	mediaID, selectedAtStr, err := s.Repo.GetCurrentHeroFeature()

	needsNew := false
	if err != nil {
		if err == sql.ErrNoRows {
			needsNew = true
		} else {
			return nil, err
		}
	} else {
		selectedAt, parseErr := time.Parse(time.RFC3339, selectedAtStr)
		if parseErr != nil {
			needsNew = true
		} else if time.Since(selectedAt).Hours() > 72 {
			needsNew = true
		}
	}

	if needsNew {
		newMediaID, err := s.Repo.GetUpcomingPopularMedia()
		if err != nil {
			return nil, err
		}

		err = s.Repo.SetCurrentHeroFeature(newMediaID)
		if err != nil {
			return nil, err
		}
		mediaID = newMediaID
	}

	media, err := s.Repo.GetMediaByID(mediaID)
	if err != nil {
		return nil, err
	}

	// Hero Data Integrity Check: skip records that are "empty shells".
	// A valid hero candidate must have both a poster image and a description.
	// If the current feature fails this check, we force a rotation to the next
	// valid candidate and try again (one retry only to prevent an infinite loop).
	if media.Image == "" || media.Description == "" {
		newMediaID, err := s.Repo.GetUpcomingPopularMedia()
		if err != nil {
			return nil, fmt.Errorf("hero candidate %d failed integrity check and no fallback found: %w", mediaID, err)
		}
		_ = s.Repo.SetCurrentHeroFeature(newMediaID)
		mediaID = newMediaID
		media, err = s.Repo.GetMediaByID(mediaID)
		if err != nil {
			return nil, err
		}
	}
	if media.MediaType == "Movie" {
		movie, cast, crew, _, jsonld, err := s.GetMovieDetail(mediaID, host)
		genres, _ := s.Repo.GetMediaGenres(mediaID)
		return map[string]interface{}{
			"Type":   "Movie",
			"Movie":  movie,
			"Cast":   cast,
			"Crew":   crew,
			"Genres": genres,
			"JSONLD": jsonld,
		}, err
	} else {
		show, cast, crew, episodes, _, jsonld, err := s.GetShowDetail(mediaID, host)
		genres, _ := s.Repo.GetMediaGenres(mediaID)
		return map[string]interface{}{
			"Type":     "TVSeries",
			"Show":     show,
			"Cast":     cast,
			"Crew":     crew,
			"Episodes": episodes,
			"Genres":   genres,
			"JSONLD":   jsonld,
		}, err
	}
}

// GetHomepageData aggregates all data needed for the comprehensive homepage.
func (s *AppService) GetHomepageData(host string) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// 1. Hero Content
	hero, err := s.GetHomepageHero(host)
	if err == nil {
		data["Hero"] = hero
	}

	// 2. Community Stats (2x2 Matrix)
	stats, err := s.Repo.GetHomepageStats()
	if err == nil {
		data["Stats"] = stats
	}

	// 3. Trending Matrix (Initial load: Popular All)
	trending, err := s.Repo.GetTrendingMedia("", 12, 0)
	if err == nil {
		data["Trending"] = trending
	}

	// 4. Activity Feed
	activities, err := s.Repo.GetRecentActivity(10)
	if err == nil {
		data["Activity"] = activities
	}

	// 5. Popular Lists
	lists, err := s.Repo.GetPopularLists(3)
	if err == nil {
		data["PopularLists"] = lists
	}

	// 6. Franchise Spotlight
	franchise, err := s.Repo.GetFranchiseSpotlight()
	if err == nil {
		data["Franchise"] = franchise
	}

	// 7. Blog Posts
	blogPosts, err := s.Repo.GetRecentBlogPosts(4)
	if err == nil {
		data["BlogPosts"] = blogPosts
	}

	// 8. Photo Grid
	photos, err := s.Repo.GetRecentPhotos(8)
	if err == nil {
		data["Photos"] = photos
	}

	// 9. Extra Columns
	data["Birthdays"], _ = s.Repo.GetActorBirthdays(time.Now().Format("01-02"))
	data["FanFavorites"], _ = s.Repo.GetFanFavorites(5)
	data["BoxOffice"], _ = s.Repo.GetTopBoxOffice(5)
	data["PopularCelebs"], _ = s.Repo.GetTrendingPeople(5, 0)

	return data, nil
}

// LogActivity is a helper to record site-wide events.
func (s *AppService) LogActivity(userID int, activityType string, targetID int, targetType string) {
	activity := models.Activity{
		UserID:       userID,
		ActivityType: activityType,
		TargetID:     targetID,
		TargetType:   targetType,
	}
	_ = s.Repo.LogActivity(activity)
}

// SeedHomepageData bootstraps some initial content if the DB is empty.
func (s *AppService) SeedHomepageData() {
	// Seed Blog Posts if none exist
	posts, _ := s.Repo.GetRecentBlogPosts(1)
	if len(posts) == 0 {
		_ = s.Repo.CreateBlogPost(models.BlogPost{
			Title:      "Welcome to the New FilmGap",
			Slug:       "welcome-to-filmgap",
			Content:    "We are excited to launch our new community-driven platform for film lovers.",
			IsFeatured: true,
		})
		_ = s.Repo.CreateBlogPost(models.BlogPost{
			Title:   "Top 10 Sci-Fi Movies of 2026",
			Slug:    "top-scifi-2026",
			Content: "The year 2026 has been incredible for science fiction...",
		})
	}

	// Seed Franchises if none exist
	f, _ := s.Repo.GetFranchiseSpotlight()
	if f.ID == 0 {
		_ = s.Repo.CreateFranchise(models.Franchise{
			Name:        "The Matrix Franchise",
			Slug:        "the-matrix",
			Description: "The world of Neo and Trinity.",
			Image:       "https://image.tmdb.org/t/p/original/dXNAPwY7Vrq7oZsnH9o9h5I9I7n.jpg",
		})
	}

	// Seed Photos if none exist
	photos, _ := s.Repo.GetRecentPhotos(1)
	if len(photos) == 0 {
		samplePhotos := []string{
			"https://image.tmdb.org/t/p/w500/kY8p266U1f3v9x8wMUMGkM1h6S1.jpg",
			"https://image.tmdb.org/t/p/w500/7WsyChvgynoqvTslas6mS69J7vS.jpg",
			"https://image.tmdb.org/t/p/w500/mS9FvA9l4qO1WnZp9U0Mvq0vK9O.jpg",
			"https://image.tmdb.org/t/p/w500/uU9RbtS74AdvSclYv3H9uB3v2Uq.jpg",
			"https://image.tmdb.org/t/p/w500/8tS87vXE87yO1Kx0t4uFv1S4L5L.jpg",
			"https://image.tmdb.org/t/p/w500/4Y7S7tS7Y7S7tS7Y7S7tS7Y7S7t.jpg",
		}
		for _, url := range samplePhotos {
			_ = s.Repo.CreatePhoto(models.Photo{
				ImageURL: url,
				Caption:  "Community Flash",
			})
		}
	}
}

func (s *AppService) GetTrendingMedia(mediaType string, limit, offset int) ([]models.MediaSummary, error) {
	return s.Repo.GetTrendingMedia(mediaType, limit, offset)
}

func (s *AppService) GetTrendingPeople(limit, offset int) ([]models.Person, error) {
	return s.Repo.GetTrendingPeople(limit, offset)
}

func normalizeProviderCountry(countryCode string) string {
	countryCode = strings.TrimSpace(countryCode)
	if countryCode == "" {
		return "US"
	}

	if strings.Contains(countryCode, ",") {
		countryCode = strings.Split(countryCode, ",")[0]
	}

	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	if len(countryCode) != 2 {
		return "US"
	}
	return countryCode
}

func (s *AppService) fetchProvidersFromTMDB(mediaType string, tmdbID, mediaID int, countryCode string) ([]models.WatchProviderOption, error) {
	var (
		result tmdb.TMDBWatchProviderCountryResult
		err    error
	)

	switch mediaType {
	case "Movie":
		result, err = s.tmdbClient.FetchMovieWatchProviders(tmdbID, countryCode)
	case "TVSeries", "TV":
		result, err = s.tmdbClient.FetchTVWatchProviders(tmdbID, countryCode)
	default:
		return nil, fmt.Errorf("unsupported media type for watch providers: %s", mediaType)
	}
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var providers []models.WatchProviderOption
	appendProviders := func(providerType string, items []tmdb.TMDBWatchProvider) {
		for _, item := range items {
			logoURL := ""
			if item.LogoPath != "" {
				logoURL = tmdb.ImageBaseURL + item.LogoPath
			}
			providers = append(providers, models.WatchProviderOption{
				MediaID:         mediaID,
				CountryCode:     countryCode,
				ProviderType:    providerType,
				ProviderID:      item.ProviderID,
				ProviderName:    item.ProviderName,
				LogoURL:         logoURL,
				DisplayPriority: item.DisplayPriority,
				DeepLinkURL:     result.Link,
				Source:          "tmdb",
				UpdatedAt:       now,
			})
		}
	}

	appendProviders("subscription", result.Flatrate)
	appendProviders("free", result.Free)
	appendProviders("free", result.Ads)
	appendProviders("rent", result.Rent)
	appendProviders("buy", result.Buy)

	return providers, nil
}

func groupWatchProviders(providers []models.WatchProviderOption) []models.WatchProviderGroup {
	groupOrder := []struct {
		Key   string
		Label string
	}{
		{Key: "subscription", Label: "Subscription"},
		{Key: "free", Label: "Free"},
		{Key: "rent", Label: "Rent"},
		{Key: "buy", Label: "Buy"},
	}

	grouped := make(map[string][]models.WatchProviderOption)
	for _, provider := range providers {
		grouped[provider.ProviderType] = append(grouped[provider.ProviderType], provider)
	}

	var result []models.WatchProviderGroup
	for _, group := range groupOrder {
		items := grouped[group.Key]
		if len(items) == 0 {
			continue
		}

		sort.SliceStable(items, func(i, j int) bool {
			if items[i].DisplayPriority == items[j].DisplayPriority {
				return items[i].ProviderName < items[j].ProviderName
			}
			return items[i].DisplayPriority < items[j].DisplayPriority
		})

		seen := make(map[int]bool)
		deduped := make([]models.WatchProviderOption, 0, len(items))
		for _, item := range items {
			if seen[item.ProviderID] {
				continue
			}
			seen[item.ProviderID] = true
			deduped = append(deduped, item)
		}

		result = append(result, models.WatchProviderGroup{
			Key:       group.Key,
			Label:     group.Label,
			Providers: deduped,
		})
	}

	return result
}

// Social Graph (Follow System)

func (s *AppService) FollowUser(followerID, followedUserID int) error {
	if followerID == followedUserID {
		return fmt.Errorf("cannot follow yourself")
	}
	err := s.Repo.FollowUser(followerID, followedUserID)
	if err == nil {
		s.LogActivity(followerID, "followed_user", followedUserID, "user")
	}
	return err
}

func (s *AppService) UnfollowUser(followerID, followedUserID int) error {
	return s.Repo.UnfollowUser(followerID, followedUserID)
}

func (s *AppService) FollowPerson(followerID, personID int) error {
	err := s.Repo.FollowPerson(followerID, personID)
	if err == nil {
		s.LogActivity(followerID, "followed_person", personID, "person")
	}
	return err
}

func (s *AppService) UnfollowPerson(followerID, personID int) error {
	return s.Repo.UnfollowPerson(followerID, personID)
}

func (s *AppService) FollowList(followerID, listID int) error {
	err := s.Repo.FollowList(followerID, listID)
	if err == nil {
		s.LogActivity(followerID, "followed_list", listID, "list")
	}
	return err
}

func (s *AppService) UnfollowList(followerID, listID int) error {
	return s.Repo.UnfollowList(followerID, listID)
}

func (s *AppService) IsFollowingUser(followerID, followedUserID int) (bool, error) {
	return s.Repo.IsFollowingUser(followerID, followedUserID)
}

func (s *AppService) IsFollowingPerson(followerID, personID int) (bool, error) {
	return s.Repo.IsFollowingPerson(followerID, personID)
}

func (s *AppService) IsFollowingList(followerID, listID int) (bool, error) {
	return s.Repo.IsFollowingList(followerID, listID)
}

func (s *AppService) GetFollowers(userID int) ([]models.User, error) {
	return s.Repo.GetFollowers(userID)
}

func (s *AppService) GetFollowingUsers(userID int) ([]models.User, error) {
	return s.Repo.GetFollowingUsers(userID)
}

func (s *AppService) GetFollowingPeople(userID int) ([]models.Person, error) {
	return s.Repo.GetFollowingPeople(userID)
}

func (s *AppService) GetFollowCounts(userID int) (followers, following int, err error) {
	return s.Repo.GetFollowCounts(userID)
}

func (s *AppService) GetPersonFollowCounts(personID int) (followers, following int, err error) {
	return s.Repo.GetPersonFollowCounts(personID)
}

func (s *AppService) GetListFollowCounts(listID int) (followers, following int, err error) {
	return s.Repo.GetListFollowCounts(listID)
}

func (s *AppService) GetFollowerFeed(userID int, limit int) ([]models.Activity, error) {
	return s.Repo.GetActivitiesByFollowed(userID, limit)
}

func (s *AppService) GetAllBlogPosts(limit, offset int) ([]models.BlogPost, error) {
	return s.Repo.GetAllBlogPosts(limit, offset)
}

func (s *AppService) GetBlogPostBySlug(slug string, baseDomain string) (models.BlogPost, string, error) {
	post, err := s.Repo.GetBlogPostBySlug(slug)
	if err != nil {
		return post, "", err
	}

	// Render Markdown to HTML
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(post.Content), &buf); err != nil {
		return post, "", err
	}

	// Sanitize output to prevent Stored XSS
	p := bluemonday.UGCPolicy()
	post.Content = p.Sanitize(buf.String())

	// Generate JSON-LD
	jsonLD, err := metadata.GenerateBlogPostJSONLD(post, baseDomain)
	if err != nil {
		return post, "", err
	}

	return post, jsonLD, nil
}
