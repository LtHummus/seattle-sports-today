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

const (
	huskiesFootballURL         = "https://site.api.espn.com/apis/site/v2/sports/football/college-football/teams/WASH"
	huskiesMensBasketballURL   = "https://site.api.espn.com/apis/site/v2/sports/basketball/mens-college-basketball/teams/264"
	huskiesWomensBasketballURL = "https://site.api.espn.com/apis/site/v2/sports/basketball/womens-college-basketball/teams/264"

	huskiesFootballName         = "Washington Huskies (Football)"
	huskiesMensBasketballName   = "Washington Huskies (Men's Basketball)"
	huskiesWomensBasketballName = "Washington Huskies (Women's Basketball)"

	huskyStadium        = "Husky Stadium"
	alaskaAirlinesArena = "Alaska Airlines Arena"

	seattleTeamKey = "seattle_team"
)

type espnTeamResponse struct {
	Team struct {
		ID        string `json:"id"`
		UID       string `json:"uid"`
		NextEvent []struct {
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
					} `json:"team"`
				} `json:"competitors"`
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
		} `json:"nextEvent"`
	} `json:"team"`
}

func queryESPNAndAdd(ctx context.Context, url string, teamName string, venue string, today *[]*Event, tomorrow *[]*Event, seattleToday time.Time, seattleTomorrow time.Time) error {
	log.Info().Str(seattleTeamKey, teamName).Msg("querying espn for team info")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Error().Err(err).Str(seattleTeamKey, teamName).Msg("could not build request")
		return fmt.Errorf("events: queryESPN: could not build request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Str(seattleTeamKey, teamName).Msg("could not contact ESPN API")
		return fmt.Errorf("events: queryESPN: could not contact API: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Str(seattleTeamKey, teamName).Err(err).Str("status_code", resp.Status).Msg("could not read error response body")
			return fmt.Errorf("events: queryESPN: could not read error body: %w", err)
		}
		log.Error().Str(seattleTeamKey, teamName).Str("status_code", resp.Status).Msg("error retrieving data from ESPN")
		return fmt.Errorf("events: queryESPN: could not retireve data from ESPN: %s", string(body))
	}

	var payload espnTeamResponse
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		log.Error().Err(err).Str(seattleTeamKey, teamName).Msg("could not decode ESPN response")
		return fmt.Errorf("events: queryESPNTeam: could not decode ESPN response: %w", err)
	}

	if payload.Team.ID == "" || payload.Team.UID == "" {
		log.Error().Str(seattleTeamKey, teamName).Str("url", url).Msg("empty ESPN payload")
		return fmt.Errorf("events: queryESPNTeam: empty response payload")
	}

	for _, curr := range payload.Team.NextEvent {
		if len(curr.Competitions) <= 0 {
			log.Warn().Str(seattleTeamKey, teamName).Msg("no games found")
			continue
		}

		competition := curr.Competitions[0]

		competitorLength := len(competition.Competitors)
		if competitorLength < 2 {
			log.Warn().Str(seattleTeamKey, teamName).Int("count", competitorLength).Msg("insufficient competitors")
			continue
		} else if competitorLength > 2 {
			log.Warn().Str(seattleTeamKey, teamName).Int("count", competitorLength).Msg("unexpected number of competitors, only using first 2")
			// Keep going, assume the first two are the home and away teams
		}

		homeTeam := competition.Competitors[0]
		awayTeam := competition.Competitors[1]
		if homeTeam.HomeAway != "home" {
			homeTeam = competition.Competitors[1]
			awayTeam = competition.Competitors[0]
		}

		gameTime, err := time.Parse("2006-01-02T15:04Z", competition.Date)
		if err != nil {
			log.Error().Err(err).Str(seattleTeamKey, teamName).Msg("could not parse start time")
			return fmt.Errorf("events: queryESPNTeam: could not parse start time: %w", err)
		}

		seattleStart := gameTime.In(SeattleTimeZone)

		if competition.Venue.FullName == venue {
			if isDay(seattleToday, seattleStart) {
				log.Info().Str(seattleTeamKey, teamName).Str("opponent", awayTeam.Team.DisplayName).Msg("found game for today")
				*today = append(*today, &Event{
					TeamName:  teamName,
					Venue:     competition.Venue.FullName,
					LocalTime: gameTime.In(SeattleTimeZone).Format(localTimeDateFormat),
					Opponent:  awayTeam.Team.DisplayName,
					RawTime:   gameTime.Unix(),
				})
			} else if isDay(seattleTomorrow, seattleStart) {
				log.Info().Str(seattleTeamKey, teamName).Str("opponent", awayTeam.Team.DisplayName).Msg("found game for tomorrow")
				*tomorrow = append(*tomorrow, &Event{
					TeamName:  teamName,
					Venue:     competition.Venue.FullName,
					LocalTime: gameTime.In(SeattleTimeZone).Format(localTimeDateFormat),
					Opponent:  awayTeam.Team.DisplayName,
					RawTime:   gameTime.Unix(),
				})
			}
		}
	}

	return nil
}

func GetUWGames(ctx context.Context, seattleToday time.Time, seattleTomorrow time.Time) ([]*Event, []*Event, error) {
	var today []*Event
	var tomorrow []*Event

	err := queryESPNAndAdd(ctx, huskiesFootballURL, huskiesFootballName, huskyStadium, &today, &tomorrow, seattleToday, seattleTomorrow)
	if err != nil {
		return nil, nil, err
	}

	err = queryESPNAndAdd(ctx, huskiesMensBasketballURL, huskiesMensBasketballName, alaskaAirlinesArena, &today, &tomorrow, seattleToday, seattleTomorrow)
	if err != nil {
		return nil, nil, err
	}

	err = queryESPNAndAdd(ctx, huskiesWomensBasketballURL, huskiesWomensBasketballName, alaskaAirlinesArena, &today, &tomorrow, seattleToday, seattleTomorrow)
	if err != nil {
		return nil, nil, err
	}

	return today, tomorrow, nil
}
