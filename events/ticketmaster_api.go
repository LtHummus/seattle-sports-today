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

const (
	AttractionIDKraken   = "K8vZ917_vgV"
	AttractionIDSeahawks = "K8vZ9171oU7"
	AttractionIDMariners = "K8vZ9171o6f"
	AttractionIDSounders = "K8vZ917G8RV"
	AttractionIDReign    = "K8vZ9178Dm7"
	AttractionIDStorm    = "K8vZ9171xo0"

	TicketmasterEventSearchAPI   = "https://app.ticketmaster.com/discovery/v2/events"
	TicketmasterApiKeySecretName = "TICKETMASTER_API_KEY_SECRET_NAME"

	SubTypeIDTouringFacility = "KZFzBErXgnZfZ7vAvv"
)

// seattleVenueMap is a map of venues to ticketmaster's internal venue ID for venues we should look at
var seattleVenueMap = map[string]string{
	"Climate Pledge Arena": "KovZ917Ahkk",
	"Lumen Field":          "KovZpZAEknnA",
	"T-Mobile Park":        "KovZpZAEevAA",
	"WAMU Theater":         "KovZpZAFFE7A",
}

// seattleTeamsMap is a set of attraction IDs (read: sports teams) that we want to ignore for this so we don't
// count these events twice (since we check for them in other places). Note for later ... should we just do _everything_
// via the tikcetmaster API? probably
var seattleTeamsMap = map[string]struct{}{
	AttractionIDKraken:   {},
	AttractionIDSeahawks: {},
	AttractionIDMariners: {},
	AttractionIDSounders: {},
	AttractionIDReign:    {},
	AttractionIDStorm:    {},
}

func beginningOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func beginningOfTomorrow(t time.Time) time.Time {
	tomorrow := t.AddDate(0, 0, 1)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, t.Location())
}

func eventShouldBeIgnored(e *TicketmasterEvent) bool {
	venueName := ""
	if len(e.Embedded.Venues) != 0 {
		venueName = e.Embedded.Venues[0].Name
	}

	if e.Dates.Status.Code == "cancelled" {
		log.Info().Str("name", e.Name).Str("venue_name", venueName).Msg("event is cancelled")
		return true
	}

	if e.Classifications == nil || len(e.Classifications) == 0 {
		log.Info().Str("name", e.Name).Str("venue_name", venueName).Msg("no classifications")
		return true
	}

	if len(e.Embedded.Attractions) == 0 {
		log.Info().Str("name", e.Name).Str("venue_name", venueName).Msg("no attractions")
		return true
	}

	if e.Classifications[0].SubType.Id == SubTypeIDTouringFacility {
		// we don't want to list arena tours (as cool as they are)
		log.Info().Str("name", e.Name).Str("venue_name", venueName).Msg("is facility tour")
		return true
	}

	for _, curr := range e.Embedded.Attractions {
		if _, contained := seattleTeamsMap[curr.Id]; contained {
			log.Info().Str("name", e.Name).Str("venue_name", venueName).Str("attraction", curr.Id).Msg("attraction is a seattle team we check separately")
			return true
		}
	}

	return false
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

	remainingRequestCount := resp.Header.Get("Rate-Limit-Available")
	rateLimitResetTime := resp.Header.Get("Rate-Limit-Reset")

	log.Info().Str("venue_name", venueName).Str("remaining_requests", remainingRequestCount).Str("rate_limit_reset_time", rateLimitResetTime).Msg("completed ticketmaster API request")

	var payload TicketmasterEventSearchResponse
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}

	// TODO: make this return multiple events if needed
	for _, e := range payload.Embedded.Events {

		if eventShouldBeIgnored(&e) {
			log.Info().Str("venue", venueName).Str("event_name", e.Name).Msg("ignoring event")
			continue
		}

		startTime := e.Dates.Start.DateTime.In(SeattleTimeZone)

		log.Info().Str("venue_name", venueName).Str("event_name", e.Name).Msg("found event from ticketmaster")

		return &Event{
			RawDescription: fmt.Sprintf("%s is at %s. It starts at %s", e.Name, venueName, startTime.Format(localTimeDateFormat)),
			RawTime:        e.Dates.Start.DateTime.Unix(),
		}, nil
	}

	return nil, nil

}

func TicketmasterEvents(ctx context.Context) ([]*Event, error) {
	ticketmasterApiKeySecretName := os.Getenv(TicketmasterApiKeySecretName)
	if ticketmasterApiKeySecretName == "" {
		log.Warn().Str("env_var_name", TicketmasterApiKeySecretName).Msg("environment variable not set. Not querying ticketmaster")
		return nil, nil
	}

	apiKey, err := secrets.GetSecretString(ctx, ticketmasterApiKeySecretName)
	if err != nil {
		return nil, fmt.Errorf("events: TicketmasterEvents: could not get ticketmaster secret: %w", err)
	}

	today := time.Now().In(SeattleTimeZone)

	start := beginningOfDay(today).Format(time.RFC3339)
	end := beginningOfTomorrow(today).Format(time.RFC3339)

	var res []*Event

	for venueName, venueID := range seattleVenueMap {
		var e *Event
		e, err = getEventForVenueID(ctx, apiKey, venueName, venueID, start, end)
		if err != nil {
			return nil, fmt.Errorf("events: TicketmasterEvents: could not query for ticketmaster data: %w", err)
		}
		if e != nil {
			res = append(res, e)
		}

		// ticketmaster limits us to 5 calls per second -- add a delay so we don't blow through that
		time.Sleep(300 * time.Millisecond)
	}

	return res, nil
}
