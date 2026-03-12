package database

import (
	"log"
	"time"
)

// AdCampaign represents a marketing target with a set budget
type AdCampaign struct {
	ID          int
	CompanyID   int
	Budget      float64
	TargetPages string
	Impressions int
	Clicks      int
	StartDate   time.Time
	EndDate     time.Time
}

// Advertisement represents the creative associated with a campaign
type Advertisement struct {
	ID          int
	CampaignID  int
	Image       string
	URL         string
	Title       string
	Description string
}

// GetAdCampaigns retrieves all campaigns for a specific company or user
func GetAdCampaigns(companyID int) ([]AdCampaign, error) {
	var campaigns []AdCampaign
	query := `SELECT id, company_id, budget, target_pages, impressions, clicks FROM ad_campaigns WHERE company_id = ?`

	rows, err := DB.Query(query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c AdCampaign
		err := rows.Scan(&c.ID, &c.CompanyID, &c.Budget, &c.TargetPages, &c.Impressions, &c.Clicks)
		if err != nil {
			log.Println("Error scanning ad campaign:", err)
			continue
		}
		campaigns = append(campaigns, c)
	}

	return campaigns, nil
}

// CreateAdCampaign initializes a new budget ad campaign
func CreateAdCampaign(companyID int, budget float64, targetPages string) (int, error) {
	query := `INSERT INTO ad_campaigns (company_id, budget, target_pages) VALUES (?, ?, ?)`
	res, err := DB.Exec(query, companyID, budget, targetPages)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	return int(id), err
}

// CreateAdvertisement attaches creative URL content to a campaign
func CreateAdvertisement(campaignID int, title, description, url, image string) error {
	query := `INSERT INTO advertisements (campaign_id, title, description, url, image) VALUES (?, ?, ?, ?, ?)`
	_, err := DB.Exec(query, campaignID, title, description, url, image)
	return err
}
