package game

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
)

// SuperheroClient fetches character data from the akabab/superhero-api,
// a free, open-source, CDN-hosted REST API with 700+ characters from
// Marvel, DC, and other comic publishers.
//
// Data is fetched once on first use and cached in-memory (~2MB for ~700 entries).
// No API key or authentication is required.
type SuperheroClient struct {
	cdnURL       string
	client       *http.Client
	geminiAPIKey string

	// Lazy-loaded full dataset
	onceFetch sync.Once
	allHeroes []superheroEntry
	fetchErr  error
}

// superheroEntry represents a single character from the Superhero API.
type superheroEntry struct {
	ID         int              `json:"id"`
	Name       string           `json:"name"`
	Slug       string           `json:"slug"`
	Powerstats superheroPower   `json:"powerstats"`
	Biography  superheroBio     `json:"biography"`
	Images     superheroImages  `json:"images"`
	Connections superheroConns  `json:"connections"`
}

type superheroPower struct {
	Intelligence int `json:"intelligence"`
	Strength     int `json:"strength"`
	Speed        int `json:"speed"`
	Durability   int `json:"durability"`
	Power        int `json:"power"`
	Combat       int `json:"combat"`
}

type superheroBio struct {
	FullName        string   `json:"fullName"`
	AlterEgos       string   `json:"alterEgos"`
	Aliases         []string `json:"aliases"`
	Publisher       string   `json:"publisher"`
	Alignment       string   `json:"alignment"`
	FirstAppearance string   `json:"firstAppearance"`
}

type superheroImages struct {
	XS string `json:"xs"`
	SM string `json:"sm"`
	MD string `json:"md"`
	LG string `json:"lg"`
}

type superheroConns struct {
	GroupAffiliation string `json:"groupAffiliation"`
}

// NewSuperheroClient creates a client for the Superhero API.
func NewSuperheroClient(geminiKey string) *SuperheroClient {
	return &SuperheroClient{
		cdnURL:       "https://cdn.jsdelivr.net/gh/akabab/superhero-api@0.3.0/api",
		client:       &http.Client{Timeout: 15 * time.Second},
		geminiAPIKey: geminiKey,
	}
}

// publisherAliases maps common theme strings to their Superhero API publisher names.
var publisherAliases = map[string]string{
	"marvel":           "Marvel Comics",
	"marvel comics":    "Marvel Comics",
	"avengers":         "Marvel Comics",
	"x-men":            "Marvel Comics",
	"x men":            "Marvel Comics",
	"xmen":             "Marvel Comics",
	"spider-man":       "Marvel Comics",
	"spiderman":        "Marvel Comics",
	"fantastic four":   "Marvel Comics",
	"guardians":        "Marvel Comics",
	"dc":               "DC Comics",
	"dc comics":        "DC Comics",
	"justice league":   "DC Comics",
	"batman":           "DC Comics",
	"superman":         "DC Comics",
	"teen titans":      "DC Comics",
}

// teamAliases maps theme strings to group affiliation keywords for filtering.
var teamAliases = map[string][]string{
	"avengers":       {"Avengers"},
	"x-men":          {"X-Men"},
	"x men":          {"X-Men"},
	"xmen":           {"X-Men"},
	"justice league": {"Justice League"},
	"teen titans":    {"Teen Titans"},
	"fantastic four": {"Fantastic Four"},
	"guardians":      {"Guardians of the Galaxy"},
	"suicide squad":  {"Suicide Squad"},
	"batman":         {"Batman", "Gotham"},
	"spider-man":     {"Spider"},
	"spiderman":      {"Spider"},
}

// FetchDeck returns a deck of characters matching the given theme from the
// Superhero API dataset. Returns an error if not enough characters are found.
func (sc *SuperheroClient) FetchDeck(theme string) (models.Deck, error) {
	if err := sc.ensureLoaded(); err != nil {
		return nil, fmt.Errorf("superhero dataset unavailable: %w", err)
	}

	themeKey := strings.ToLower(strings.TrimSpace(theme))

	// Step 1: Filter by publisher and/or team affiliation
	var matched []superheroEntry

	publisher := publisherAliases[themeKey]
	teamKeywords := teamAliases[themeKey]

	if publisher != "" || len(teamKeywords) > 0 {
		// Score-based matching: team affiliation matches rank higher
		type scored struct {
			entry superheroEntry
			score int
		}
		var candidates []scored

		for _, hero := range sc.allHeroes {
			s := 0

			// Publisher match
			if publisher != "" && strings.EqualFold(hero.Biography.Publisher, publisher) {
				s += 1
			}

			// Team affiliation match (higher priority)
			if len(teamKeywords) > 0 {
				affiliation := strings.ToLower(hero.Connections.GroupAffiliation)
				for _, kw := range teamKeywords {
					if strings.Contains(affiliation, strings.ToLower(kw)) {
						s += 10
						break
					}
				}
			}

			// Name contains theme
			if strings.Contains(strings.ToLower(hero.Name), themeKey) {
				s += 5
			}

			if s > 0 {
				candidates = append(candidates, scored{entry: hero, score: s})
			}
		}

		// Sort by score descending (simple insertion sort — small dataset)
		for i := 1; i < len(candidates); i++ {
			for j := i; j > 0 && candidates[j].score > candidates[j-1].score; j-- {
				candidates[j], candidates[j-1] = candidates[j-1], candidates[j]
			}
		}

		for _, c := range candidates {
			matched = append(matched, c.entry)
			if len(matched) >= models.TotalCards {
				break
			}
		}
	}

	// Step 2: Fallback — fuzzy name search across all heroes
	if len(matched) < models.TotalCards {
		seen := make(map[int]bool)
		for _, m := range matched {
			seen[m.ID] = true
		}

		for _, hero := range sc.allHeroes {
			if seen[hero.ID] {
				continue
			}

			nameMatch := strings.Contains(strings.ToLower(hero.Name), themeKey)
			aliasMatch := false
			for _, alias := range hero.Biography.Aliases {
				if strings.Contains(strings.ToLower(alias), themeKey) {
					aliasMatch = true
					break
				}
			}
			fullNameMatch := strings.Contains(strings.ToLower(hero.Biography.FullName), themeKey)

			if nameMatch || aliasMatch || fullNameMatch {
				matched = append(matched, hero)
				if len(matched) >= models.TotalCards {
					break
				}
			}
		}
	}

	if len(matched) < models.TotalCards {
		return nil, fmt.Errorf("not enough superhero characters found for %q (%d/%d)", theme, len(matched), models.TotalCards)
	}

	// Step 3: Convert to Meta Clash cards
	deck := make(models.Deck, 0, models.TotalCards)
	for _, hero := range matched[:models.TotalCards] {
		stats := mapSuperheroStats(hero.Powerstats)
		image := hero.Images.MD
		if image == "" {
			image = hero.Images.LG
		}
		if image == "" {
			image = fmt.Sprintf("https://picsum.photos/seed/%s/320/420", url.QueryEscape(hero.Name))
		}

		deck = append(deck, models.Card{
			ID:    generateCardID(hero.Name),
			Name:  hero.Name,
			Image: image,
			Stats: stats,
		})
	}

	return deck, nil
}

// mapSuperheroStats converts Superhero API powerstats (1-100 scale, 6 attrs)
// to Meta Clash stats (1-99 scale, 4 attrs).
func mapSuperheroStats(ps superheroPower) models.Stats {
	// Rank = overall average of all 6 powerstats
	avg := (ps.Intelligence + ps.Strength + ps.Speed + ps.Durability + ps.Power + ps.Combat) / 6
	return models.Stats{
		Rank:     clampStat(avg),
		Strength: clampStat(ps.Strength),
		Speed:    clampStat(ps.Speed),
		IQ:       clampStat(ps.Intelligence),
	}
}

// clampStat ensures a stat is in the valid [1, 99] range.
func clampStat(v int) int {
	if v < 1 {
		return 1
	}
	if v > 99 {
		return 99
	}
	return v
}

// ensureLoaded fetches the full Superhero API dataset exactly once.
func (sc *SuperheroClient) ensureLoaded() error {
	sc.onceFetch.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), sc.client.Timeout)
		defer cancel()

		allURL := sc.cdnURL + "/all.json"
		slog.Info("fetching superhero dataset", "url", allURL)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, allURL, nil)
		if err != nil {
			sc.fetchErr = fmt.Errorf("request creation failed: %w", err)
			return
		}

		resp, err := sc.client.Do(req)
		if err != nil {
			sc.fetchErr = fmt.Errorf("superhero API fetch failed: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			sc.fetchErr = fmt.Errorf("superhero API returned status %d", resp.StatusCode)
			return
		}

		if err := json.NewDecoder(resp.Body).Decode(&sc.allHeroes); err != nil {
			sc.fetchErr = fmt.Errorf("failed to decode superhero data: %w", err)
			return
		}

		slog.Info("superhero dataset loaded", "count", len(sc.allHeroes))
	})
	return sc.fetchErr
}
