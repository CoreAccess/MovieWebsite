package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"movieweb/internal/database"
)

// adsPortalView renders the management interface for advertisers
func adsPortalView(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/partials/sidebar.tmpl",
		"./ui/html/pages/ads.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := getTemplateData("Advertiser Portal", r)
	
	// Fetch campaigns for this user (company_id)
	campaigns, _ := database.GetAdCampaigns(user.ID) // Assumes user.ID is also their company ID
	data.AdCampaigns = campaigns

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err.Error())
	}
}

// createAdCampaignPost handles submission of a new advertising campaign
func createAdCampaignPost(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
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
	campaignID, err := database.CreateAdCampaign(user.ID, budget, `["` + targetPages + `"]`)
	if err == nil {
		// 2. Create Advertisement Creative mapping to the Campaign
		database.CreateAdvertisement(campaignID, title, description, url, image)
	} else {
		log.Println("Error creating campaign:", err)
	}

	http.Redirect(w, r, "/ads?success=campaign_created", http.StatusSeeOther)
}
