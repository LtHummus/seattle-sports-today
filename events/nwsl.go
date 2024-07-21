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

type espnNWSLResponse struct {
	Leagues []struct {
		Id           string `json:"id"`
		Uid          string `json:"uid"`
		Name         string `json:"name"`
		Abbreviation string `json:"abbreviation"`
		MidsizeName  string `json:"midsizeName"`
		Slug         string `json:"slug"`
		Season       struct {
			Year        int    `json:"year"`
			StartDate   string `json:"startDate"`
			EndDate     string `json:"endDate"`
			DisplayName string `json:"displayName"`
			Type        struct {
				Id           string `json:"id"`
				Type         int    `json:"type"`
				Name         string `json:"name"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type"`
		} `json:"season"`
		Logos []struct {
			Href        string   `json:"href"`
			Width       int      `json:"width"`
			Height      int      `json:"height"`
			Alt         string   `json:"alt"`
			Rel         []string `json:"rel"`
			LastUpdated string   `json:"lastUpdated"`
		} `json:"logos"`
		CalendarType        string   `json:"calendarType"`
		CalendarIsWhitelist bool     `json:"calendarIsWhitelist"`
		CalendarStartDate   string   `json:"calendarStartDate"`
		CalendarEndDate     string   `json:"calendarEndDate"`
		Calendar            []string `json:"calendar"`
	} `json:"leagues"`
	Season struct {
		Type int `json:"type"`
		Year int `json:"year"`
	} `json:"season"`
	Day struct {
		Date string `json:"date"`
	} `json:"day"`
	Events []struct {
		Id        string `json:"id"`
		Uid       string `json:"uid"`
		Date      string `json:"date"`
		Name      string `json:"name"`
		ShortName string `json:"shortName"`
		Season    struct {
			Year int    `json:"year"`
			Type int    `json:"type"`
			Slug string `json:"slug"`
		} `json:"season"`
		Competitions []struct {
			Id         string `json:"id"`
			Uid        string `json:"uid"`
			Date       string `json:"date"`
			StartDate  string `json:"startDate"`
			Attendance int    `json:"attendance"`
			TimeValid  bool   `json:"timeValid"`
			Recent     bool   `json:"recent"`
			Status     struct {
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
			} `json:"status"`
			Venue struct {
				Id       string `json:"id"`
				FullName string `json:"fullName"`
				Address  struct {
					City    string `json:"city"`
					Country string `json:"country"`
				} `json:"address"`
			} `json:"venue"`
			Format struct {
				Regulation struct {
					Periods int `json:"periods"`
				} `json:"regulation"`
			} `json:"format"`
			Notes         []interface{} `json:"notes"`
			GeoBroadcasts []struct {
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
			} `json:"geoBroadcasts"`
			Broadcasts []struct {
				Market string   `json:"market"`
				Names  []string `json:"names"`
			} `json:"broadcasts"`
			Competitors []struct {
				Id       string `json:"id"`
				Uid      string `json:"uid"`
				Type     string `json:"type"`
				Order    int    `json:"order"`
				HomeAway string `json:"homeAway"`
				Winner   bool   `json:"winner"`
				Form     string `json:"form"`
				Score    string `json:"score"`
				Records  []struct {
					Name         string `json:"name"`
					Type         string `json:"type"`
					Summary      string `json:"summary"`
					Abbreviation string `json:"abbreviation"`
				} `json:"records"`
				Team struct {
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
				Statistics []struct {
					Name         string `json:"name"`
					Abbreviation string `json:"abbreviation"`
					DisplayValue string `json:"displayValue"`
				} `json:"statistics"`
			} `json:"competitors"`
			Details []struct {
				Type struct {
					Id   string `json:"id"`
					Text string `json:"text"`
				} `json:"type"`
				Clock struct {
					Value        float64 `json:"value"`
					DisplayValue string  `json:"displayValue"`
				} `json:"clock"`
				Team struct {
					Id string `json:"id"`
				} `json:"team"`
				ScoreValue       int  `json:"scoreValue"`
				ScoringPlay      bool `json:"scoringPlay"`
				RedCard          bool `json:"redCard"`
				YellowCard       bool `json:"yellowCard"`
				PenaltyKick      bool `json:"penaltyKick"`
				OwnGoal          bool `json:"ownGoal"`
				Shootout         bool `json:"shootout"`
				AthletesInvolved []struct {
					Id          string `json:"id"`
					DisplayName string `json:"displayName"`
					ShortName   string `json:"shortName"`
					FullName    string `json:"fullName"`
					Jersey      string `json:"jersey"`
					Team        struct {
						Id string `json:"id"`
					} `json:"team"`
					Links []struct {
						Rel      []string `json:"rel"`
						Href     string   `json:"href"`
						IsHidden bool     `json:"isHidden"`
					} `json:"links"`
					Position string `json:"position"`
				} `json:"athletesInvolved"`
			} `json:"details"`
			WasSuspended        bool `json:"wasSuspended"`
			PlayByPlayAvailable bool `json:"playByPlayAvailable"`
			PlayByPlayAthletes  bool `json:"playByPlayAthletes"`
		} `json:"competitions"`
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
		} `json:"status"`
		Venue struct {
			DisplayName string `json:"displayName"`
		} `json:"venue"`
		Links []struct {
			Language   string   `json:"language"`
			Rel        []string `json:"rel"`
			Href       string   `json:"href"`
			Text       string   `json:"text"`
			ShortText  string   `json:"shortText"`
			IsExternal bool     `json:"isExternal"`
			IsPremium  bool     `json:"isPremium"`
			IsHidden   bool     `json:"isHidden,omitempty"`
		} `json:"links"`
	} `json:"events"`
}

func GetReignGame(ctx context.Context) (*Event, error) {
	today := time.Now().In(seattleTimeZone)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://site.api.espn.com/apis/site/v2/sports/soccer/usa.nwsl/scoreboard", nil)
	if err != nil {
		return nil, fmt.Errorf("events: GetStormGame: could not build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("events: GetStormGame: could not make WNBA request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Str("status_code", resp.Status).Msg("could not read error response body")
			return nil, fmt.Errorf("events: GetStormGame: could not read error body: %w", err)
		}
		return nil, fmt.Errorf("events: GetStormGame: could not retireve MLB schedule: %s", string(body))
	}

	var payload espnNWSLResponse
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return nil, fmt.Errorf("events: GetStormGame: could not decode JSON: %w", err)
	}

	for _, curr := range payload.Events {
		competition := curr.Competitions[0]
		homeTeam := competition.Competitors[0]
		awayTeam := competition.Competitors[1]
		if homeTeam.HomeAway != "home" {
			homeTeam = competition.Competitors[1]
			awayTeam = competition.Competitors[0]
		}

		if homeTeam.Team.Abbreviation == "SEA" {
			gameTime, err := time.Parse("2006-01-02T15:04Z", competition.StartDate)
			if err != nil {
				log.Error().Err(err).Str("start_date", competition.StartDate).Msg("could not parse start date")
				return nil, fmt.Errorf("events: GetSeahawksGame: could not parse start time: %w", err)
			}

			seattleStart := gameTime.In(seattleTimeZone)
			if today.Year() != seattleStart.Year() || today.YearDay() != seattleStart.YearDay() {
				continue
			}

			return &Event{
				TeamName:  "Seattle Reign",
				Venue:     competition.Venue.FullName,
				LocalTime: seattleStart.Format(localTimeDateFormat),
				Opponent:  awayTeam.Team.Name,
			}, nil
		}
	}

	return nil, nil

}
