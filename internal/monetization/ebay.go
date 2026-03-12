package monetization

import (
	"movieweb/internal/models"
)

// FetchEbayListings returns mock eBay affiliate listings for a given keyword
func FetchEbayListings(keyword string) []models.EbayListing {
	// In production, this would make an API call to eBay Partner Network
	return []models.EbayListing{
		{
			Title:    keyword + " Funko POP!",
			Price:    "$14.99",
			Url:      "https://partnernetwork.ebay.com/",
			ImageUrl: "https://via.placeholder.com/250x250/ffffff/5A5F6D?text=Funko+Pop",
			IsHot:    false,
		},
		{
			Title:    keyword + " Heavyweight T-Shirt",
			Price:    "$29.99",
			Url:      "https://partnernetwork.ebay.com/",
			ImageUrl: "https://via.placeholder.com/250x250/ffffff/5A5F6D?text=T-Shirt",
			IsHot:    false,
		},
		{
			Title:    "The Art of " + keyword,
			Price:    "$55.00",
			Url:      "https://partnernetwork.ebay.com/",
			ImageUrl: "https://via.placeholder.com/250x250/ffffff/5A5F6D?text=Book",
			IsHot:    false,
		},
		{
			Title:    keyword + " Prop Replica",
			Price:    "$149.99",
			Url:      "https://partnernetwork.ebay.com/",
			ImageUrl: "https://via.placeholder.com/250x250/ffffff/5A5F6D?text=Prop+Replica",
			IsHot:    true,
		},
	}
}
