package main

import (
	"sort"

	"filmgap/internal/models"
)

func buildTVSeasonGroups(episodes []models.TVEpisode) []models.TVSeasonGroup {
	if len(episodes) == 0 {
		return nil
	}

	groupedEpisodes := make(map[int][]models.TVEpisode)
	seasonNumbers := make([]int, 0)
	seenSeasons := make(map[int]bool)

	for _, episode := range episodes {
		groupedEpisodes[episode.SeasonNumber] = append(groupedEpisodes[episode.SeasonNumber], episode)
		if !seenSeasons[episode.SeasonNumber] {
			seasonNumbers = append(seasonNumbers, episode.SeasonNumber)
			seenSeasons[episode.SeasonNumber] = true
		}
	}

	sort.Ints(seasonNumbers)

	seasonGroups := make([]models.TVSeasonGroup, 0, len(seasonNumbers))
	for _, seasonNumber := range seasonNumbers {
		seasonEpisodes := append([]models.TVEpisode(nil), groupedEpisodes[seasonNumber]...)
		sort.SliceStable(seasonEpisodes, func(i, j int) bool {
			if seasonEpisodes[i].EpisodeNumber == seasonEpisodes[j].EpisodeNumber {
				return seasonEpisodes[i].Name < seasonEpisodes[j].Name
			}
			return seasonEpisodes[i].EpisodeNumber < seasonEpisodes[j].EpisodeNumber
		})

		seasonGroups = append(seasonGroups, models.TVSeasonGroup{
			SeasonNumber: seasonNumber,
			EpisodeCount: len(seasonEpisodes),
			Episodes:     seasonEpisodes,
		})
	}

	return seasonGroups
}
