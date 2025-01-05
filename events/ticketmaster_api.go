package events

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lthummus/seattle-sports-today/secrets"
)

type TicketmasterAPIResponse struct {
	Embedded struct {
		Events []struct {
			Name   string `json:"name"`
			Type   string `json:"type"`
			Id     string `json:"id"`
			Test   bool   `json:"test"`
			Url    string `json:"url"`
			Locale string `json:"locale"`
			Dates  struct {
				Start struct {
					LocalDate      string    `json:"localDate"`
					LocalTime      string    `json:"localTime"`
					DateTime       time.Time `json:"dateTime"`
					DateTBD        bool      `json:"dateTBD"`
					DateTBA        bool      `json:"dateTBA"`
					TimeTBA        bool      `json:"timeTBA"`
					NoSpecificTime bool      `json:"noSpecificTime"`
				} `json:"start"`
				Timezone string `json:"timezone"`
				Status   struct {
					Code string `json:"code"`
				} `json:"status"`
				SpanMultipleDays bool `json:"spanMultipleDays"`
				InitialStartDate struct {
					LocalDate string    `json:"localDate"`
					LocalTime string    `json:"localTime"`
					DateTime  time.Time `json:"dateTime"`
				} `json:"initialStartDate,omitempty"`
			} `json:"dates"`
			Classifications []struct {
				Primary bool `json:"primary"`
				Segment struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"segment"`
				Genre struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"genre"`
				SubGenre struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"subGenre"`
				Type struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"type"`
				SubType struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"subType"`
				Family bool `json:"family"`
			} `json:"classifications"`
			Promoter struct {
				Id          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"promoter,omitempty"`
			Promoters []struct {
				Id          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"promoters,omitempty"`
			Info       string `json:"info,omitempty"`
			DoorsTimes struct {
				LocalDate string    `json:"localDate"`
				LocalTime string    `json:"localTime"`
				DateTime  time.Time `json:"dateTime"`
				Id        string    `json:"id"`
			} `json:"doorsTimes,omitempty"`
		} `json:"events"`
	} `json:"_embedded"`
}

const (
	VenueIDClimatePledgeArena = "KovZ917Ahkk"
	VenueIDLumenField         = "KovZpZAEknnA"
	VenueIDTMobilePark        = "KovZpZAEevAA"
	VenueIDWAMUTheater        = "KovZpZAFFE7A"

	SegmentIDMusic  = "KZFzniwnSyZfZ7v7nJ"
	SegmentIDSports = "KZFzniwnSyZfZ7v7nE"

	TicketmasterEventSearchAPI   = "https://app.ticketmaster.com/discovery/v2/events"
	TicketmasterApiKeySecretName = "TICKETMASTER_API_KEY_SECRET_NAME"
)

var seattleVenueMap = map[string]string{
	"Climate Pledge Arena": VenueIDClimatePledgeArena,
	"Lumen Field":          VenueIDLumenField,
	"T-Mobile Park":        VenueIDTMobilePark,
	"WAMU Theater":         VenueIDWAMUTheater,
}

func beginningOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func beginningOfTomorrow(t time.Time) time.Time {
	tomorrow := t.AddDate(0, 0, 1)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, t.Location())
}

func getEventForVenueID(ctx context.Context, apiKey string, venueName string, venueID string, searchStart string, searchEnd string) (*Event, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, TicketmasterEventSearchAPI, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("venueId", venueID)
	q.Add("apikey", apiKey)
	q.Add("startDateTime", searchStart)
	q.Add("endDateTime", searchEnd)
	q.Add("segmentId", SegmentIDMusic)
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Str("status", resp.Status).Msg("could not read error response body")
			return nil, fmt.Errorf("events: getEventForVenueID: could not read error body: %w", err)
		}
		log.Error().Str("status", resp.Status).Msg("error retrieving data from ticketmaster")
		return nil, fmt.Errorf("events: getEventForVenueID: could not retireve data from ticketmaster: %s", string(body))
	}

	var payload TicketmasterAPIResponse
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}

	if len(payload.Embedded.Events) == 0 {
		return nil, nil
	}

	e := payload.Embedded.Events[0]

	startTime := e.Dates.Start.DateTime.In(SeattleTimeZone)

	return &Event{
		RawDescription: fmt.Sprintf("%s is performing at %s. The show starts at %s", e.Name, venueName, startTime.Format(localTimeDateFormat)),
		RawTime:        e.Dates.Start.DateTime.Unix(),
	}, nil
}

// GetMusicalEvents only finds musical events. This should probably be expanded to non-music events that happen at these
// places, but that's TODO later.
func GetMusicalEvents(ctx context.Context) ([]*Event, error) {
	ticketmasterApiKeySecretName := os.Getenv(TicketmasterApiKeySecretName)
	if ticketmasterApiKeySecretName == "" {
		log.Warn().Str("env_var_name", TicketmasterApiKeySecretName).Msg("environment variable not set. Not querying ticketmaster")
		return nil, nil
	}

	apiKey, err := secrets.GetSecretString(ctx, ticketmasterApiKeySecretName)
	if err != nil {
		return nil, fmt.Errorf("events: GetMusicalEvents: could not get ticketmaster secret: %w", err)
	}

	today := time.Now().In(SeattleTimeZone)

	start := beginningOfDay(today).Format(time.RFC3339)
	end := beginningOfTomorrow(today).Format(time.RFC3339)

	var res []*Event

	for venueName, venueID := range seattleVenueMap {
		var e *Event
		e, err = getEventForVenueID(ctx, apiKey, venueName, venueID, start, end)
		if err != nil {
			return nil, fmt.Errorf("events: GetMusicalEvents: could not query for ticketmaster data: %w", err)
		}
		if e != nil {
			res = append(res, e)
		}

		// ticketmaster limits us to 5 calls per second -- add a delay so we don't blow through that
		time.Sleep(300 * time.Millisecond)
	}

	return res, nil
}
