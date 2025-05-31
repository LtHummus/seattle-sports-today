package events

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lthummus/seattle-sports-today/secrets"
)

type WNBACompetition struct {
	Broadcasts []struct {
		Market string `json:"market"`
		Name   string `json:"name"`
		Type   string `json:"type"`
	} `json:"broadcasts"`
	Competitors []struct {
		Abbrev      string `json:"abbrev"`
		AltColor    string `json:"altColor"`
		DisplayName string `json:"displayName"`
		Id          string `json:"id"`
		IsHome      bool   `json:"isHome"`
		Leader      struct {
			DisplayValue string `json:"displayValue"`
			Href         string `json:"href"`
			Name         string `json:"name"`
			ShortName    string `json:"shortName"`
			Uid          string `json:"uid"`
		} `json:"leader"`
		Links            string `json:"links"`
		Location         string `json:"location"`
		Logo             string `json:"logo"`
		Name             string `json:"name"`
		RecordSummary    string `json:"recordSummary"`
		Score            int    `json:"score"`
		ShortDisplayName string `json:"shortDisplayName"`
		ShortName        string `json:"shortName"`
		StandingSummary  string `json:"standingSummary"`
		TeamColor        string `json:"teamColor"`
		Uid              string `json:"uid"`
		Winner           bool   `json:"winner,omitempty"`
	} `json:"competitors"`
	Completed bool   `json:"completed"`
	Date      string `json:"date"`
	Format    struct {
		Regulation struct {
			Periods int `json:"periods"`
		} `json:"regulation"`
	} `json:"format"`
	HeaderPostfix string `json:"headerPostfix"`
	Id            string `json:"id"`
	IsTie         bool   `json:"isTie"`
	Link          string `json:"link"`
	Season        struct {
		Slug string `json:"slug"`
		Type int    `json:"type"`
		Year int    `json:"year"`
	} `json:"season"`
	Status struct {
		Detail string `json:"detail"`
		Id     string `json:"id"`
		State  string `json:"state"`
	} `json:"status"`
	Tbd   bool `json:"tbd"`
	Teams []struct {
		Abbrev      string `json:"abbrev"`
		AltColor    string `json:"altColor"`
		DisplayName string `json:"displayName"`
		Id          string `json:"id"`
		IsHome      bool   `json:"isHome"`
		Leader      struct {
			DisplayValue string `json:"displayValue"`
			Href         string `json:"href"`
			Name         string `json:"name"`
			ShortName    string `json:"shortName"`
			Uid          string `json:"uid"`
		} `json:"leader"`
		Links            string `json:"links"`
		Location         string `json:"location"`
		Logo             string `json:"logo"`
		Name             string `json:"name"`
		RecordSummary    string `json:"recordSummary"`
		Score            int    `json:"score"`
		ShortDisplayName string `json:"shortDisplayName"`
		ShortName        string `json:"shortName"`
		StandingSummary  string `json:"standingSummary"`
		TeamColor        string `json:"teamColor"`
		Uid              string `json:"uid"`
		Winner           bool   `json:"winner,omitempty"`
	} `json:"teams"`
	Tickets struct {
	} `json:"tickets"`
	TimeValid bool `json:"timeValid"`
	Venue     struct {
		Address struct {
			City  string `json:"city"`
			State string `json:"state"`
		} `json:"address"`
		FullName string `json:"fullName"`
		Id       string `json:"id"`
		Indoor   bool   `json:"indoor"`
	} `json:"venue"`
}

func GetStormGame(ctx context.Context) ([]*Event, error) {
	apiKeySecretName := os.Getenv("WBNA_API_KEY_SECRET_NAME")
	if apiKeySecretName == "" {
		log.Warn().Msg("WBNA_API_KEY_SECRET_NAME not set")
		return nil, nil
	}

	apiKey, err := secrets.GetSecretString(ctx, apiKeySecretName)
	if err != nil {
		log.Error().Err(err).Str("api_key_secret_name", apiKeySecretName).Msg("could not retrieve wbna api key")
		return nil, nil
	}

	year := SeattleCurrentTime.Year()
	month := SeattleCurrentTime.Month()
	day := SeattleCurrentTime.Day()

	v := url.Values{}
	v.Add("year", fmt.Sprintf("%d", year))
	v.Add("month", fmt.Sprintf("%02d", month))
	v.Add("day", fmt.Sprintf("%02d", day))

	log.Info().Str("wbna_api_key_secret_name", apiKeySecretName).Msg("querying for WNBA games")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://wnba-api.p.rapidapi.com/wnbaschedule?%s", v.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("events: GetStormGame: could not build request: %w", err)
	}

	req.Header.Set("x-rapidapi-key", apiKey)
	req.Header.Set("x-rapidapi-host", "wnba-api.p.rapidapi.com")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("events: GetStormGame: could not get data from WNBA API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Str("status_code", resp.Status).Msg("could not read error response body")
			return nil, fmt.Errorf("events: GetStormGame: could not read error body: %w", err)
		}
		return nil, fmt.Errorf("events: GetStormGame: could not retireve WNBA schedule: %s", string(body))
	}

	log.Info().Str("status", resp.Status).Msg("got response from WNBA api")

	var payload map[string][]WNBACompetition
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return nil, fmt.Errorf("event: GetStormGame: invalid API response")
	}

	formattedDate := SeattleCurrentTime.Format("20060102")
	todaysGames := payload[formattedDate]
	if todaysGames == nil {
		return nil, nil
	}

	for _, curr := range todaysGames {
		var homeTeamAbbrev string
		var awayTeamName string

		for _, team := range curr.Teams {
			if team.IsHome {
				homeTeamAbbrev = team.Abbrev
			} else {
				awayTeamName = team.Name
			}
		}

		if homeTeamAbbrev != "SEA" {
			continue
		}

		// not RFC3339 because there are no seconds
		startTime, err := time.ParseInLocation("2006-01-02T15:04Z", curr.Date, time.UTC)
		if err != nil {
			log.Error().Err(err).Str("given_date", curr.Date).Msg("could not parse start time")
			return nil, fmt.Errorf("event: GetStormGame: could not parse start time")
		}
		seattleStart := startTime.In(SeattleTimeZone)

		log.Info().
			Str("team_name", "Seattle Storm").
			Str("opponent", awayTeamName).
			Str("start_time", seattleStart.Format(localTimeDateFormat)).
			Msg("found game")
		return []*Event{
			{
				TeamName:  "Seattle Storm",
				Venue:     curr.Venue.FullName,
				LocalTime: seattleStart.Format(localTimeDateFormat),
				Opponent:  awayTeamName,
				RawTime:   seattleStart.Unix(),
			},
		}, nil
	}

	log.Info().Msg("no WNBA games today")

	return nil, nil
}
