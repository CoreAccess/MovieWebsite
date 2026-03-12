package monetization

import (
	"movieweb/internal/models"
)

// FetchAdvertisements returns mock ad campaigns for a given context
func FetchAdvertisements(context string) []models.Advertisement {
	// In production, this would query the DB for active overlapping campaigns
	return []models.Advertisement{
		{
			Title:       "NordVPN - 70% Off",
			Description: "Stay secure and unblock content worldwide.",
			Url:         "https://nordvpn.com",
			Image:       "https://placehold.co/400x250/0f172a/38bdf8?text=NordVPN+Promo",
		},
	}
}
