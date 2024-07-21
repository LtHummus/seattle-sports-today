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

type espnNFLResponse struct {
	Leagues []struct {
		Id           string `json:"id"`
		Uid          string `json:"uid"`
		Name         string `json:"name"`
		Abbreviation string `json:"abbreviation"`
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
		CalendarType        string `json:"calendarType"`
		CalendarIsWhitelist bool   `json:"calendarIsWhitelist"`
		CalendarStartDate   string `json:"calendarStartDate"`
		CalendarEndDate     string `json:"calendarEndDate"`
		Calendar            []struct {
			Label     string `json:"label"`
			Value     string `json:"value"`
			StartDate string `json:"startDate"`
			EndDate   string `json:"endDate"`
			Entries   []struct {
				Label          string `json:"label"`
				AlternateLabel string `json:"alternateLabel"`
				Detail         string `json:"detail"`
				Value          string `json:"value"`
				StartDate      string `json:"startDate"`
				EndDate        string `json:"endDate"`
			} `json:"entries"`
		} `json:"calendar"`
	} `json:"leagues"`
	Season struct {
		Type int `json:"type"`
		Year int `json:"year"`
	} `json:"season"`
	Week struct {
		Number int `json:"number"`
	} `json:"week"`
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
		Week struct {
			Number int `json:"number"`
		} `json:"week"`
		Competitions []struct {
			Id         string `json:"id"`
			Uid        string `json:"uid"`
			Date       string `json:"date"`
			Attendance int    `json:"attendance"`
			Type       struct {
				Id           string `json:"id"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type"`
			TimeValid             bool `json:"timeValid"`
			NeutralSite           bool `json:"neutralSite"`
			ConferenceCompetition bool `json:"conferenceCompetition"`
			PlayByPlayAvailable   bool `json:"playByPlayAvailable"`
			Recent                bool `json:"recent"`
			Venue                 struct {
				Id       string `json:"id"`
				FullName string `json:"fullName"`
				Address  struct {
					City  string `json:"city"`
					State string `json:"state,omitempty"`
				} `json:"address"`
				Indoor bool `json:"indoor"`
			} `json:"venue"`
			Competitors []struct {
				Id       string `json:"id"`
				Uid      string `json:"uid"`
				Type     string `json:"type"`
				Order    int    `json:"order"`
				HomeAway string `json:"homeAway"`
				Team     struct {
					Id               string `json:"id"`
					Uid              string `json:"uid"`
					Location         string `json:"location"`
					Name             string `json:"name"`
					Abbreviation     string `json:"abbreviation"`
					DisplayName      string `json:"displayName"`
					ShortDisplayName string `json:"shortDisplayName"`
					Color            string `json:"color"`
					AlternateColor   string `json:"alternateColor"`
					IsActive         bool   `json:"isActive"`
					Venue            struct {
						Id string `json:"id"`
					} `json:"venue"`
					Links []struct {
						Rel        []string `json:"rel"`
						Href       string   `json:"href"`
						Text       string   `json:"text"`
						IsExternal bool     `json:"isExternal"`
						IsPremium  bool     `json:"isPremium"`
					} `json:"links"`
					Logo string `json:"logo"`
				} `json:"team"`
				Score      string        `json:"score"`
				Statistics []interface{} `json:"statistics"`
			} `json:"competitors"`
			Notes []struct {
				Type     string `json:"type"`
				Headline string `json:"headline"`
			} `json:"notes"`
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
			Broadcasts []struct {
				Market string   `json:"market"`
				Names  []string `json:"names"`
			} `json:"broadcasts"`
			Format struct {
				Regulation struct {
					Periods int `json:"periods"`
				} `json:"regulation"`
			} `json:"format"`
			Tickets []struct {
				Summary         string `json:"summary"`
				NumberAvailable int    `json:"numberAvailable"`
				Links           []struct {
					Href string `json:"href"`
				} `json:"links"`
			} `json:"tickets"`
			StartDate     string `json:"startDate"`
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
			Odds []struct {
				Provider struct {
					Id       string `json:"id"`
					Name     string `json:"name"`
					Priority int    `json:"priority"`
				} `json:"provider"`
				Details      string  `json:"details"`
				OverUnder    float64 `json:"overUnder"`
				Spread       float64 `json:"spread"`
				AwayTeamOdds struct {
					Favorite bool `json:"favorite"`
					Underdog bool `json:"underdog"`
					Team     struct {
						Id           string `json:"id"`
						Uid          string `json:"uid"`
						Abbreviation string `json:"abbreviation"`
						Name         string `json:"name"`
						DisplayName  string `json:"displayName"`
						Logo         string `json:"logo"`
					} `json:"team"`
				} `json:"awayTeamOdds"`
				HomeTeamOdds struct {
					Favorite bool `json:"favorite"`
					Underdog bool `json:"underdog"`
					Team     struct {
						Id           string `json:"id"`
						Uid          string `json:"uid"`
						Abbreviation string `json:"abbreviation"`
						Name         string `json:"name"`
						DisplayName  string `json:"displayName"`
						Logo         string `json:"logo"`
					} `json:"team"`
				} `json:"homeTeamOdds"`
				Open struct {
					Over struct {
						Value                 float64 `json:"value"`
						DisplayValue          string  `json:"displayValue"`
						AlternateDisplayValue string  `json:"alternateDisplayValue"`
						Decimal               float64 `json:"decimal"`
						Fraction              string  `json:"fraction"`
						American              string  `json:"american"`
					} `json:"over"`
					Under struct {
						Value                 float64 `json:"value"`
						DisplayValue          string  `json:"displayValue"`
						AlternateDisplayValue string  `json:"alternateDisplayValue"`
						Decimal               float64 `json:"decimal"`
						Fraction              string  `json:"fraction"`
						American              string  `json:"american"`
					} `json:"under"`
					Total struct {
						AlternateDisplayValue string `json:"alternateDisplayValue"`
						American              string `json:"american"`
					} `json:"total"`
				} `json:"open"`
				Current struct {
					Over struct {
						Value                 float64 `json:"value"`
						DisplayValue          string  `json:"displayValue"`
						AlternateDisplayValue string  `json:"alternateDisplayValue"`
						Decimal               float64 `json:"decimal"`
						Fraction              string  `json:"fraction"`
						American              string  `json:"american"`
					} `json:"over"`
					Under struct {
						Value                 float64 `json:"value"`
						DisplayValue          string  `json:"displayValue"`
						AlternateDisplayValue string  `json:"alternateDisplayValue"`
						Decimal               float64 `json:"decimal"`
						Fraction              string  `json:"fraction"`
						American              string  `json:"american"`
					} `json:"under"`
					Total struct {
						AlternateDisplayValue string `json:"alternateDisplayValue"`
						American              string `json:"american"`
					} `json:"total"`
				} `json:"current"`
			} `json:"odds"`
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
	} `json:"events"`
}

func GetSeahawksGame(ctx context.Context) (*Event, error) {
	today := time.Now().In(seattleTimeZone)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard", nil)
	if err != nil {
		return nil, fmt.Errorf("events: GetSeahawksGame: could not build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("events: GetSeahawksGame: could not make NFL request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Str("status_code", resp.Status).Msg("could not read error response body")
			return nil, fmt.Errorf("events: GetSeahawksGame: could not read error body: %w", err)
		}
		return nil, fmt.Errorf("events: GetSeahawksGame: could not retireve MLB schedule: %s", string(body))
	}

	var payload espnNFLResponse
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return nil, fmt.Errorf("events: GetSeahawksGame: could not decode JSON: %w", err)
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
				TeamName:  "Seattle Seahawks",
				Venue:     competition.Venue.FullName,
				LocalTime: seattleStart.Format(localTimeDateFormat),
				Opponent:  awayTeam.Team.Name,
			}, nil
		}
	}

	return nil, nil
}
