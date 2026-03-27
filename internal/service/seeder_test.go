package service

import (
	"fmt"
	"filmgap/internal/config"
	"filmgap/internal/models"
	"filmgap/internal/repository/dbrepo"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSeedMockData(t *testing.T) {
	config.LoadEnv("../../.env")

	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	repo := &dbrepo.PostgresDBRepo{}
	_, err := repo.InitDB(dsn, "")
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	// 1. Seed Users
	t.Log("Seeding users...")
	users := []struct {
		username string
		email    string
	}{
		{"alex_cinema", "alex@example.com"},
		{"sarah_reviews", "sarah@example.com"},
		{"mj_watson", "mj@example.com"},
		{"peter_p", "peter@example.com"},
		{"bruce_w", "bruce@example.com"},
		{"clark_k", "clark@example.com"},
		{"diana_p", "diana@example.com"},
		{"tony_s", "tony@example.com"},
		{"steve_r", "steve@example.com"},
		{"nat_r", "nat@example.com"},
	}

	dummyHash := "$2a$10$WpP9vI.a1H7Qk.D8V1z5e.7Q0O0O0O0O0O0O0O0O0O0O0O0O0O0O" 

	for _, u := range users {
		_ = repo.CreateUser(u.username, u.email, dummyHash)
	}

	allUsers, _ := repo.GetAllUsers(20, 0)

	// 2. Fetch Media and Assign Genres/BoxOffice/Popularity
	t.Log("Enhancing media...")
	movies, _ := repo.GetAllMovies(50, 0, "")
	allMediaIDs := []int{}
	for i, m := range movies {
		allMediaIDs = append(allMediaIDs, m.ID)
		if i % 2 == 0 {
			_ = repo.UpsertGenreForMedia(m.ID, 1, "Action")
			_ = repo.UpsertGenreForMedia(m.ID, 2, "Sci-Fi")
		}
		if i < 10 {
			m.BoxOffice = fmt.Sprintf("$%d,000,000", 800 - (i*40))
			m.AggregateRating = 7.5 + (float64(i) * 0.2)
			if m.AggregateRating > 10 { m.AggregateRating = 9.8 }
			_, _ = repo.UpsertMovie(m)
		}
	}
	shows, _ := repo.GetAllShows(50, 0, "")
	for _, s := range shows {
		allMediaIDs = append(allMediaIDs, s.ID)
	}

	// 3. Seed Reviews & Ratings
	t.Log("Seeding reviews...")
	reviewBodies := []string{
		"Absolutely stunning visuals and a gripping story.",
		"A bit slow in the middle, but the ending made up for it.",
		"Masterpiece. I've watched this three times already.",
        "Overrated, but still worth a watch for the cinematography.",
        "The acting was top-notch, especially the lead role.",
	}

	for i := 0; i < 100; i++ {
		user := allUsers[rand.Intn(len(allUsers))]
		mediaID := allMediaIDs[rand.Intn(len(allMediaIDs))]
		rating := float64(rand.Intn(6) + 5) 

		review := models.Review{
			UserID:     user.ID,
			MediaID:    mediaID,
			Rating:     rating,
			Body:       reviewBodies[rand.Intn(len(reviewBodies))],
			ReviewType: "user",
		}
		_ = repo.UpsertReview(review)
		_ = repo.RecalculateMediaRating(mediaID)
        _ = repo.LogActivity(models.Activity{
            UserID:       user.ID,
            ActivityType: "rating",
            TargetID:     mediaID,
            TargetType:   "Media",
        })
	}

	// 4. Seed Lists
	t.Log("Seeding lists...")
	listNames := []string{"Best of 2025", "Sci-Fi Masterpieces", "Weekend Binge", "Hidden Gems"}
	for i := 0; i < 12; i++ {
		user := allUsers[rand.Intn(len(allUsers))]
		name := listNames[rand.Intn(len(listNames))]
        if i > 3 { name = fmt.Sprintf("%s Vol. %d", name, i) }
        
		list := models.List{
			UserID:      user.ID,
			Name:        name,
			Description: "A curated collection for testing.",
			Visibility:  "public",
            Slug:        fmt.Sprintf("test-list-%d-%d", user.ID, i),
		}
		listID, err := repo.CreateList(list)
		if err == nil {
			for j := 0; j < 8; j++ {
				mediaID := allMediaIDs[rand.Intn(len(allMediaIDs))]
				_ = repo.AddListItem(models.ListItem{
					ListID:  listID,
					MediaID: mediaID,
					AddedBy: user.ID,
					Rank:    j + 1,
				})
			}
            _ = repo.LogActivity(models.Activity{
                UserID:       user.ID,
                ActivityType: "list_created",
                TargetID:     listID,
                TargetType:   "List",
            })
		}
	}

    // 5. Seed Birthdays for TODAY
    t.Log("Seeding birthdays for today...")
    today := time.Now().Format("01-02")
    celebs := []string{"Keanu Reeves", "Zendaya", "Tom Hardy", "Emily Blunt", "Pedro Pascal"}
    for i, name := range celebs {
        p := models.Person{
            Name: name,
            Slug: strings.ToLower(strings.ReplaceAll(name, " ", "-")),
            BirthDate: fmt.Sprintf("19%d-%s", 70+i, today),
            KnownForDepartment: "Acting",
            Image: fmt.Sprintf("https://image.tmdb.org/t/p/w200/avatar%d.png", (i%5)+1),
        }
        _, _ = repo.InsertPerson(p)
    }

    // 6. Seed Blog Posts
    t.Log("Seeding blog posts...")
    for i := 1; i <= 6; i++ {
        _ = repo.CreateBlogPost(models.BlogPost{
            Title:      fmt.Sprintf("Why Cinematic Universe %d is the Future", i),
            Slug:       fmt.Sprintf("future-cinema-%d", i),
            Content:    "In this post, we explore the deep philosophical implications of modern franchises...",
            IsFeatured: i == 1,
        })
    }
    
    // 7. Seed Franchises
    t.Log("Seeding franchises and links...")
    franchisesData := []struct {
        name string
        slug string
    }{
        {"The Matrix", "the-matrix"},
        {"Marvel", "marvel"},
        {"Star Wars", "star-wars"},
        {"Dune", "dune"},
        {"John Wick", "john-wick"},
    }
    for _, fData := range franchisesData {
        _ = repo.CreateFranchise(models.Franchise{
            Name:        fData.name,
            Slug:        fData.slug,
            Description: fmt.Sprintf("All about the %s universe.", fData.name),
            Image:       "https://image.tmdb.org/t/p/original/dXNAPwY7Vrq7oZsnH9o9h5I9I7n.jpg",
        })
        
        // Fetch the created franchise to get the ID (or just use a subquery/returning but easier for test)
        f, _ := repo.GetFranchiseBySlug(fData.slug)
        if f.ID > 0 {
            // Link 3 random movies to each franchise
            for i := 0; i < 3; i++ {
                mID := allMediaIDs[rand.Intn(len(allMediaIDs))]
                _ = repo.AddMediaToFranchise(mID, f.ID, i+1)
            }
        }
    }

	t.Log("Mock data enrichment completed successfully.")
}
