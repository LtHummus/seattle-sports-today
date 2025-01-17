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
			Status     struct {
				Type struct {
					Name        string `json:"name"`
					Description string `json:"description"`
				} `json:"type"`
			} `json:"status"`
			Venue struct {
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

// TODO: the next two functions can be DRY'd out a bit

func queryESPNTeam(ctx context.Context, url string, seattleTeam string, expectedVenue string) ([]*Event, error) {
	log.Info().Str("seattle_team", seattleTeam).Msg("querying for espn team info")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Error().Err(err).Str("seattle_team", seattleTeam).Msg("could not build request")
		return nil, fmt.Errorf("events: queryESPNTeam: could not build request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Str("seattle_team", seattleTeam).Msg("could not contact ESPN API")
		return nil, fmt.Errorf("events: queryESPNTeam: could not contact ESPN API: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Str("seattle_team", seattleTeam).Err(err).Str("status_code", resp.Status).Msg("could not read error response body")
			return nil, fmt.Errorf("events: queryESPNTeam: could not read error body: %w", err)
		}
		log.Error().Str("seattle_team", seattleTeam).Str("status_code", resp.Status).Msg("error retrieving data from ESPN")
		return nil, fmt.Errorf("events: queryESPNTeam: could not retireve data from ESPN: %s", string(body))
	}

	var payload espnTeamResponse
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		log.Error().Err(err).Str("seattle_team", seattleTeam).Msg("could not decode ESPN response")
		return nil, fmt.Errorf("events: queryESPNTeam: could not decode ESPN response: %w", err)
	}

	if payload.Team.ID == "" || payload.Team.UID == "" {
		log.Error().Str("seattle_team", seattleTeam).Str("url", url).Msg("empty ESPN payload")
		return nil, fmt.Errorf("events: queryESPNTeam: empty response payload")
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
			log.Error().Err(err).Str("seattle_team", seattleTeam).Msg("could not parse start time")
			return nil, fmt.Errorf("events: queryESPNTeam: could not parse start time: %w", err)
		}

		seattleStart := gameTime.In(SeattleTimeZone)
		if SeattleCurrentTime.Year() != seattleStart.Year() || SeattleCurrentTime.YearDay() != seattleStart.YearDay() {
			continue
		}

		if competition.Venue.FullName == expectedVenue {
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

			if competition.Status.Type.Name == "STATUS_CANCELED" {
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
