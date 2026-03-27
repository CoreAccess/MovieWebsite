package models

import "time"

// Data structs derived from Schema.org for maximum AI Agent/SEO compatibility.

// User represents a system user.
type User struct {
	ID              int
	Username        string
	Email           string
	PasswordHash    string
	GoogleID        string
	FacebookID      string
	Avatar          string // URL or path to the user's profile picture
	ReputationScore int    // Score based on community engagement and contributions
	Role            string // 'user', 'contributor', 'moderator', 'admin'
	FollowerCount   int
	FollowingCount  int
	CreatedAt       time.Time
}

// Media is the core "Supertype" mapping to Schema.org's CreativeWork, Movie, TVSeries.
// In the database, all subtypes (movies, tv_series) will reference `media.id`.
type Media struct {
	ID              int       `json:"-"`
	MediaType       string    `json:"@type"` // "Movie" or "TVSeries"
	Name            string    `json:"name"`
	Slug            string    `json:"url"`
	Description     string    `json:"description,omitempty"`
	Image           string    `json:"image,omitempty"`
	DatePublished   string    `json:"datePublished,omitempty"`
	AggregateRating float64   `json:"aggregateRating,omitempty"`
	CreatedAt       time.Time `json:"-"`
}

// Movie is a subtype of Media.
type Movie struct {
	Media
	ReleaseDate      string `json:"-"`
	Duration         int    `json:"duration,omitempty"`
	ContentRating    string
	TmdbID           int     `json:"-"`
	Genres           []Genre `json:"genre,omitempty"`
	RatingCount      int
	ReviewCount      int
	BestRating       float64
	WorstRating      float64
	IsFamilyFriendly bool
	Budget           string
	BoxOffice        string
	LanguageCode     string
	CountryCode      string
	Tagline          string
	Subtitle         string
}

// TVSeries is a subtype of Media.
type TVSeries struct {
	Media
	StartDate        string `json:"-"`
	EndDate          string `json:"endDate,omitempty"`
	NumberOfSeasons  int    `json:"numberOfSeasons,omitempty"`
	NumberOfEpisodes int    `json:"numberOfEpisodes,omitempty"`
	ContentRating    string
	TmdbID           int     `json:"-"`
	Genres           []Genre `json:"genre,omitempty"`
	RatingCount      int
	ReviewCount      int
	BestRating       float64
	WorstRating      float64
	IsFamilyFriendly bool
	LanguageCode     string
	CountryCode      string
	Tagline          string
	Subtitle         string
}

// Person represents a cast or crew member (Schema.org Person).
type Person struct {
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	Slug               string  `json:"slug"`
	Biography          string  `json:"biography,omitempty"`
	Image              string  `json:"image,omitempty"`
	Gender             string  `json:"gender,omitempty"`
	BirthPlace         string  `json:"birthPlace,omitempty"`
	BirthDate          string  `json:"birthDate,omitempty"`
	Deathday           string  `json:"deathDate,omitempty"`
	PopularityScore    float64 `json:"popularity_score"`
	KnowsLanguage      string  `json:"knowsLanguage,omitempty"`
	NationalityCode    string  `json:"nationalityCode,omitempty"`
	KnownForDepartment string  `json:"known_for_department,omitempty"`
	FollowerCount      int     `json:"followerCount"`
	TmdbID             int     `json:"-"`
	KnownFor           string  `json:"jobTitle,omitempty"`
}

// CastMember represents a role a Person played in Media.
type CastMember struct {
	Person
	Character    Person `json:"characterName"`
	BillingOrder int
	Order        int `json:"-"`
}

// CrewMember represents a behind-the-scenes role a Person played in Media.
type CrewMember struct {
	Person
	Job        string `json:"jobTitle"`
	Department string `json:"-"`
}

// Organization represents a production company or similar entity (Schema.org Organization).
type Organization struct {
	ID            int    `json:"-"`
	Name          string `json:"name"`
	Slug          string `json:"url"`
	Description   string `json:"description,omitempty"`
	LogoURL       string `json:"logo,omitempty"`
	OriginCountry string `json:"location,omitempty"`
	TmdbID        int    `json:"-"`
}

// Watchlist and WatchlistItem
type Watchlist struct {
	ID          int
	UserID      int
	Name        string
	Description string
	IsPublic    bool
	ItemCount   int
	CreatedAt   time.Time
}

type WatchlistItem struct {
	ID          int
	WatchlistID int
	MediaID     int
	MediaType   string
	AddedAt     time.Time
}

// Session stores authenticated user sessions.
type Session struct {
	ID        string
	UserID    int
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Follow represents a user following another user, person, or list.
type Follow struct {
	ID               int       `json:"id"`
	FollowerID       int       `json:"followerId"`
	FollowedUserID   *int      `json:"followedUserId,omitempty"`
	FollowedPersonID *int      `json:"followedPersonId,omitempty"`
	FollowedListID   *int      `json:"followedListId,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
}

// List represents a user-created collection of media.
type List struct {
	ID              int       `json:"id"`
	UserID          int       `json:"userId"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Description     string    `json:"description"`
	IsRanked        bool      `json:"isRanked"`
	IsCollaborative bool      `json:"isCollaborative"`
	Visibility      string    `json:"visibility"` // 'public', 'private', 'friends'
	LikeCount       int       `json:"likeCount"`
	FollowerCount   int       `json:"followerCount"`
	ItemCount       int       `json:"itemCount"`
	IsFeatured      bool      `json:"isFeatured"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`

	// Joins
	Username string `json:"username,omitempty"`
}

// ListItem represents an entry in a user List.
type ListItem struct {
	ID      int       `json:"id"`
	ListID  int       `json:"listId"`
	MediaID int       `json:"mediaId"`
	Rank    int       `json:"rank,omitempty"`
	Note    string    `json:"note,omitempty"`
	AddedBy int       `json:"addedBy"`
	AddedAt time.Time `json:"addedAt"`

	// Joins
	MediaName string `json:"mediaName,omitempty"`
	MediaSlug string `json:"mediaSlug,omitempty"`
	MediaType string `json:"mediaType,omitempty"`
	MediaImg  string `json:"mediaImg,omitempty"`
}

// ListCollaborator represents a user who can edit a collaborative list.
type ListCollaborator struct {
	ID       int       `json:"id"`
	ListID   int       `json:"listId"`
	UserID   int       `json:"userId"`
	Role     string    `json:"role"` // 'contributor', 'editor'
	JoinedAt time.Time `json:"joinedAt"`
}

// Genre
type Genre struct {
	ID   int
	Name string
	Slug string
}

// TVEpisode (Schema.org TVEpisode)
type TVEpisode struct {
	ID            int    `json:"-"`
	SeriesID      int    `json:"-"`
	SeasonNumber  int    `json:"seasonNumber"`
	EpisodeNumber int    `json:"episodeNumber"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	AirDate       string `json:"datePublished,omitempty"`
	DatePublished string
	Slug          string
	Image         string
	Duration      int
}

type TVSeasonGroup struct {
	SeasonNumber int
	EpisodeCount int
	Episodes     []TVEpisode
}

// Composite detail types for handler/template use
type MovieDetail struct {
	Movie
	Cast      []CastMember
	Directors []Person
	Writers   []Person
	Genres    []Genre
}

// WatchProviderOption represents one provider/method pairing for a media title.
type WatchProviderOption struct {
	ID              int
	MediaID         int
	CountryCode     string
	ProviderType    string
	ProviderID      int
	ProviderName    string
	LogoURL         string
	DisplayPriority int
	DeepLinkURL     string
	Source          string
	UpdatedAt       time.Time
}

// WatchProviderGroup is a template-friendly grouping of provider options.
type WatchProviderGroup struct {
	Key       string
	Label     string
	Providers []WatchProviderOption
}

type TVSeriesDetail struct {
	TVSeries
	Series       TVSeries
	Cast         []CastMember
	Creators     []Person
	Genres       []Genre
	Episodes     []TVEpisode
	SeasonGroups []TVSeasonGroup
	Directors    []Person
	Writers      []Person
}

type PersonDetail struct {
	Person
	Movies []Movie
	Shows  []TVSeries
}

// IngestionJob represents one row in the pending_ingestion table.
type IngestionJob struct {
	ID        int
	TmdbID    int
	MediaType string // "Movie" or "TV"
	Status    string // "QUEUED", "PROCESSING", "COMPLETED", "FAILED"
	Attempts  int
}

// Review represents a user rating and written review for a piece of media.
type Review struct {
	ID               int       `json:"id"`
	UserID           int       `json:"userId"`
	MediaID          int       `json:"mediaId"`
	Rating           float64   `json:"rating"` // 1-10 scale
	Title            string    `json:"title,omitempty"`
	Body             string    `json:"body,omitempty"`
	ReviewType       string    `json:"reviewType"` // "user", "critic", "quick"
	ContainsSpoilers bool      `json:"containsSpoilers"`
	LikeCount        int       `json:"likeCount"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`

	// Joins
	Username string `json:"username,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

// BlogPost represents a blog entry.
type BlogPost struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Slug       string    `json:"slug"`
	Content    string    `json:"content"` // Markdown or pre-rendered HTML
	Image      string    `json:"image,omitempty"`
	AuthorID   int       `json:"authorId,omitempty"`
	Author     string    `json:"author,omitempty"`
	IsFeatured bool      `json:"isFeatured"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// Activity represents a user action in the system.
type Activity struct {
	ID              int       `json:"id"`
	UserID          int       `json:"userId"`
	Username        string    `json:"username,omitempty"`
	UserAvatar      string    `json:"userAvatar,omitempty"`
	ActivityType    string    `json:"activityType"` // 'follow', 'list_create', 'review_post', etc.
	TargetID        int       `json:"targetId,omitempty"`
	TargetType      string    `json:"targetType,omitempty"` // 'User', 'List', 'Review', 'Media'
	TargetName      string    `json:"targetName,omitempty"`
	TargetSlug      string    `json:"targetSlug,omitempty"`
	TargetMediaType string    `json:"targetMediaType,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
}

// Franchise represents a collection of related media (e.g., "The Matrix").
type Franchise struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	Image       string    `json:"image,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Photo represents a media or person asset for the homepage gallery.
type Photo struct {
	ID        int       `json:"id"`
	MediaID   int       `json:"mediaId,omitempty"`
	PersonID  int       `json:"personId,omitempty"`
	ImageURL  string    `json:"imageUrl"`
	Caption   string    `json:"caption,omitempty"`
	ViewCount int       `json:"viewCount"`
	CreatedAt time.Time `json:"createdAt"`
}

// HomepageStats holds the calculation result for the homepage counters.
type HomepageStats struct {
	UserCount   int
	MovieCount  int
	ShowCount   int
	ReviewCount int
}

// MediaSummary is a lightweight version of media for grids and trending lists.
type MediaSummary struct {
	ID              int     `json:"id"`
	MediaType       string  `json:"media_type"`
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	Image           string  `json:"image"`
	AggregateRating float64 `json:"aggregate_rating"`
	Year            string  `json:"year"`
}
