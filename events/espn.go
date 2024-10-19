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

// a lot of our sports sources use the ESPN api, so all that common code is here

type espnResponse struct {
	Events []struct {
		Id           string `json:"id"`
		Uid          string `json:"uid"`
		Date         string `json:"date"`
		Name         string `json:"name"`
		ShortName    string `json:"shortName"`
		Competitions []struct {
			Id         string `json:"id"`
			Uid        string `json:"uid"`
			Date       string `json:"date"`
			StartDate  string `json:"startDate"`
			Attendance int    `json:"attendance"`
			TimeValid  bool   `json:"timeValid"`
			Recent     bool   `json:"recent"`
			Venue      struct {
				Id       string `json:"id"`
				FullName string `json:"fullName"`
				Address  struct {
					City    string `json:"city"`
					Country string `json:"country"`
				} `json:"address"`
			} `json:"venue"`
			Competitors []struct {
				Id       string `json:"id"`
				Uid      string `json:"uid"`
				Type     string `json:"type"`
				Order    int    `json:"order"`
				HomeAway string `json:"homeAway"`
				Winner   bool   `json:"winner"`
				Form     string `json:"form"`
				Score    string `json:"score"`
				Team     struct {
					Id               string `json:"id"`
					Uid              string `json:"uid"`
					Abbreviation     string `json:"abbreviation"`
					DisplayName      string `json:"displayName"`
					ShortDisplayName string `json:"shortDisplayName"`
					Name             string `json:"name"`
					Location         string `json:"location"`
					Color            string `json:"color"`
					AlternateColor   string `json:"alternateColor"`
					IsActive         bool   `json:"isActive"`
					Logo             string `json:"logo"`
					Links            []struct {
						Rel        []string `json:"rel"`
						Href       string   `json:"href"`
						Text       string   `json:"text"`
						IsExternal bool     `json:"isExternal"`
						IsPremium  bool     `json:"isPremium"`
						IsHidden   bool     `json:"isHidden"`
					} `json:"links"`
					Venue struct {
						Id string `json:"id"`
					} `json:"venue"`
				} `json:"team"`
			} `json:"competitors"`
		} `json:"competitions"`
		Venue struct {
			DisplayName string `json:"displayName"`
		} `json:"venue"`
	} `json:"events"`
}

func queryESPN(ctx context.Context, url string, seattleTeam string, abbreviation string) ([]*Event, error) {
	log.Info().Str("url", url).Str("team", seattleTeam).Msg("querying ESPN API")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Error().Err(err).Str("seattle_team", seattleTeam).Msg("could not build request")
		return nil, fmt.Errorf("events: queryESPN: could not build request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Str("seattle_team", seattleTeam).Msg("could not make ESPN API request")
		return nil, fmt.Errorf("events: queryESPN: could not make ESPN API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Str("seattle_team", seattleTeam).Err(err).Str("status_code", resp.Status).Msg("could not read error response body")
			return nil, fmt.Errorf("events: queryESPN: could not read error body: %w", err)
		}
		log.Error().Str("seattle_team", seattleTeam).Str("status_code", resp.Status).Msg("error retrieving data from ESPN")
		return nil, fmt.Errorf("events: queryESPN: could not retireve data from ESPN: %s", string(body))
	}

	var payload espnResponse
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		log.Error().Str("seattle_team", seattleTeam).Err(err).Msg("could not decode JSON")
		return nil, fmt.Errorf("events: queryESPN: could not decode JSON: %w", err)
	}

	for _, curr := range payload.Events {
		competition := curr.Competitions[0]
		homeTeam := competition.Competitors[0]
		awayTeam := competition.Competitors[1]
		if homeTeam.HomeAway != "home" {
			homeTeam = competition.Competitors[1]
			awayTeam = competition.Competitors[0]
		}

		if homeTeam.Team.Abbreviation == abbreviation {
			gameTime, err := time.Parse("2006-01-02T15:04Z", competition.StartDate)
			if err != nil {
				log.Error().Err(err).Str("seattle_team", seattleTeam).Msg("could not parse start time")
				return nil, fmt.Errorf("events: queryESPN: could not parse start time: %w", err)
			}

			seattleStart := gameTime.In(SeattleTimeZone)
			if SeattleCurrentTime.Year() != seattleStart.Year() || SeattleCurrentTime.YearDay() != seattleStart.YearDay() {
				continue
			}

			return []*Event{
				{
					TeamName:  seattleTeam,
					Venue:     competition.Venue.FullName,
					LocalTime: gameTime.In(SeattleTimeZone).Format(localTimeDateFormat),
					Opponent:  awayTeam.Team.DisplayName,
					RawTime:   gameTime.Unix(),
				},
			}, nil
		}
	}

	return nil, nil
}
