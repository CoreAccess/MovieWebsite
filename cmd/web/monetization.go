package main

import (
	"fmt"
	"log"
	"net/http"

	"movieweb/internal/database"
)

// adsPortalView renders the management interface for advertisers
func (app *application) adsPortalView(w http.ResponseWriter, r *http.Request) {
	user := app.getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := app.getTemplateData("Advertiser Portal", r)

	// Fetch campaigns for this user (company_id)
	campaigns, _ := database.GetAdCampaigns(user.ID) // Assumes user.ID is also their company ID
	data.AdCampaigns = campaigns

	app.render(w, http.StatusOK, "ads.html", data)
}

// createAdCampaignPost handles submission of a new advertising campaign
func (app *application) createAdCampaignPost(w http.ResponseWriter, r *http.Request) {
	user := app.getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	budgetStr := r.PostForm.Get("budget")
	targetPages := r.PostForm.Get("target_pages")

	var budget float64
	fmt.Sscanf(budgetStr, "%f", &budget)

	title := r.PostForm.Get("title")
	description := r.PostForm.Get("description")
	url := r.PostForm.Get("url")
	image := r.PostForm.Get("image")

	// 1. Create Campaign
	campaignID, err := database.CreateAdCampaign(user.ID, budget, `["`+targetPages+`"]`)
	if err == nil {
		// 2. Create Advertisement Creative mapping to the Campaign
		database.CreateAdvertisement(campaignID, title, description, url, image)
	} else {
		log.Println("Error creating campaign:", err)
	}

	http.Redirect(w, r, "/ads?success=campaign_created", http.StatusSeeOther)
}
