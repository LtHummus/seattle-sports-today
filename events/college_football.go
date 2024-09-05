package events

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type huskiesFootballPayload struct {
	Team struct {
		NextEvent []struct {
			Id        string `json:"id"`
			Date      string `json:"date"`
			Name      string `json:"name"`
			ShortName string `json:"shortName"`
			Week      struct {
				Number int    `json:"number"`
				Text   string `json:"text"`
			} `json:"week"`
			TimeValid    bool `json:"timeValid"`
			Competitions []struct {
				Id         string `json:"id"`
				Date       string `json:"date"`
				Attendance int    `json:"attendance"`
				Type       struct {
					Id           string `json:"id"`
					Text         string `json:"text"`
					Abbreviation string `json:"abbreviation"`
					Slug         string `json:"slug"`
					Type         string `json:"type"`
				} `json:"type"`
				TimeValid         bool `json:"timeValid"`
				NeutralSite       bool `json:"neutralSite"`
				BoxscoreAvailable bool `json:"boxscoreAvailable"`
				TicketsAvailable  bool `json:"ticketsAvailable"`
				Venue             struct {
					FullName string `json:"fullName"`
					Address  struct {
						City    string `json:"city"`
						State   string `json:"state"`
						ZipCode string `json:"zipCode"`
					} `json:"address"`
				} `json:"venue"`
				Competitors []struct {
					Id       string `json:"id"`
					Type     string `json:"type"`
					Order    int    `json:"order"`
					HomeAway string `json:"homeAway"`
					Team     struct {
						Id               string `json:"id"`
						Location         string `json:"location"`
						Nickname         string `json:"nickname"`
						Abbreviation     string `json:"abbreviation"`
						DisplayName      string `json:"displayName"`
						ShortDisplayName string `json:"shortDisplayName"`
						Logos            []struct {
							Href        string   `json:"href"`
							Width       int      `json:"width"`
							Height      int      `json:"height"`
							Alt         string   `json:"alt"`
							Rel         []string `json:"rel"`
							LastUpdated string   `json:"lastUpdated"`
						} `json:"logos"`
						Links []struct {
							Rel  []string `json:"rel"`
							Href string   `json:"href"`
							Text string   `json:"text"`
						} `json:"links"`
					} `json:"team"`
					CuratedRank struct {
						Current int `json:"current"`
					} `json:"curatedRank"`
				} `json:"competitors"`
				Notes      []interface{} `json:"notes"`
				Broadcasts []struct {
					Type struct {
						Id        string `json:"id"`
						ShortName string `json:"shortName"`
					} `json:"type"`
					Market struct {
						Id   string `json:"id"`
						Type string `json:"type"`
					} `json:"market"`
					Media struct {
						ShortName string `json:"shortName"`
					} `json:"media"`
					Lang   string `json:"lang"`
					Region string `json:"region"`
				} `json:"broadcasts"`
				Tickets []struct {
					Id              string  `json:"id"`
					Summary         string  `json:"summary"`
					Description     string  `json:"description"`
					MaxPrice        float64 `json:"maxPrice"`
					StartingPrice   float64 `json:"startingPrice"`
					NumberAvailable int     `json:"numberAvailable"`
					TotalPostings   int     `json:"totalPostings"`
					Links           []struct {
						Rel  []string `json:"rel"`
						Href string   `json:"href"`
					} `json:"links"`
				} `json:"tickets"`
				Status struct {
					Clock        float64 `json:"clock"`
					DisplayClock string  `json:"displayClock"`
					Period       int     `json:"period"`
					Type         struct {
						Id          string `json:"id"`
						Name        string `json:"name"`
						State       string `json:"state"`
						Completed   bool   `json:"completed"`
						Description string `json:"description"`
						Detail      string `json:"detail"`
						ShortDetail string `json:"shortDetail"`
					} `json:"type"`
					IsTBDFlex bool `json:"isTBDFlex"`
				} `json:"status"`
			} `json:"competitions"`
			Links []struct {
				Language   string   `json:"language"`
				Rel        []string `json:"rel"`
				Href       string   `json:"href"`
				Text       string   `json:"text"`
				ShortText  string   `json:"shortText"`
				IsExternal bool     `json:"isExternal"`
				IsPremium  bool     `json:"isPremium"`
			} `json:"links"`
		} `json:"nextEvent"`
		StandingSummary string `json:"standingSummary"`
	} `json:"team"`
}

const (
	huskiesURL  = "https://site.api.espn.com/apis/site/v2/sports/football/college-football/teams/WASH"
	huskiesName = "Washington Huskies"
)

func GetHuskiesFootballGame(ctx context.Context) (*Event, error) {
	log.Info().Msg("querying ESPN API for Huskies")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, huskiesURL, nil)
	if err != nil {
		log.Error().Err(err).Str("seattle_team", huskiesName).Msg("could not build request for huskies")
		return nil, fmt.Errorf("events: GetHuskiesFootballGame: could not build request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Str("seattle_team", huskiesName).Msg("could not make ESPN API request")
		return nil, fmt.Errorf("events: GetHuskiesFootballGame: could not make API request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Str("seattle_team", huskiesName).Err(err).Str("status_code", resp.Status).Msg("could not read error response body")
			return nil, fmt.Errorf("events: GetHuskiesFootballGame: could not read error body: %w", err)
		}
		log.Error().Str("seattle_team", huskiesName).Str("status_code", resp.Status).Msg("error retrieving data from ESPN")
		return nil, fmt.Errorf("events: GetHuskiesFootballGame: could not retireve data from ESPN: %s", string(body))
	}

	var payload huskiesFootballPayload
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		log.Error().Err(err).Str("seattle_team", huskiesName).Msg("could not decode JSON")
		return nil, fmt.Errorf("events: GetHuskiesFootballGame: could not decode JSON: %w", err)
	}

	for _, curr := range payload.Team.NextEvent {
		competition := curr.Competitions[0]
		homeTeam := competition.Competitors[0]
		awayTeam := competition.Competitors[1]
		if homeTeam.HomeAway != "home" {
			homeTeam = competition.Competitors[1]
			awayTeam = competition.Competitors[0]
		}

		gameTime, err := time.Parse("2006-01-02T15:04Z", competition.Date)
		if err != nil {
			log.Error().Err(err).Str("seattle_team", huskiesName).Msg("could not parse start time")
			return nil, fmt.Errorf("events: queryESPN: could not parse start time: %w", err)
		}

		seattleStart := gameTime.In(SeattleTimeZone)
		if SeattleCurrentTime.Year() != seattleStart.Year() || SeattleCurrentTime.YearDay() != seattleStart.YearDay() {
			continue
		}

		if competition.Venue.FullName == "Husky Stadium" {
			return &Event{
				TeamName:  huskiesName,
				Venue:     competition.Venue.FullName,
				LocalTime: gameTime.In(SeattleTimeZone).Format(localTimeDateFormat),
				Opponent:  awayTeam.Team.DisplayName,
			}, nil
		}
	}

	return nil, nil
}
