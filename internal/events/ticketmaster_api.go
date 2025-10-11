package events

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/time/rate"

	"github.com/rs/zerolog/log"

	"github.com/lthummus/seattle-sports-today/internal/secrets"
)

const (
	TicketmasterEventSearchAPI   = "https://app.ticketmaster.com/discovery/v2/events"
	TicketmasterApiKeySecretName = "TICKETMASTER_API_KEY_SECRET_NAME"

	SubTypeIDTouringFacility = "KZFzBErXgnZfZ7vAvv"
)

var ticketmasterRateLimiter = rate.NewLimiter(4, 4)

// seattleVenueMap is a map of venues to ticketmaster's internal venue ID for venues we should look at
var seattleVenueMap = map[string]string{
	"Climate Pledge Arena": "KovZ917Ahkk",
	"Lumen Field":          "KovZpZAEknnA",
	"T-Mobile Park":        "KovZpZAEevAA",
	"WAMU Theater":         "KovZpZAFFE7A",
}

var seattleTeamAttractionIDs = map[string]string{
	"K8vZ917_vgV": "Seattle Kraken",
	"K8vZ9171oU7": "Seattle Seahawks",
	"K8vZ9171o6f": "Seattle Mariners",
	"K8vZ917G8RV": "Seattle Sounders",
	"K8vZ9178Dm7": "Seattle Reign",
	"K8vZ9171xo0": "Seattle Storm",
}

func beginningOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
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

	if e.Dates.Start.DateTBD || e.Dates.Start.DateTBA {
		log.Info().Str("name", e.Name).Str("venue_name", venueName).Msg("time is TBA")
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

	return false
}

func buildInternalEvent(e TicketmasterEvent, venueName string) (*Event, error) {
	var seattleTeam string
	for _, curr := range e.Embedded.Attractions {
		if team := seattleTeamAttractionIDs[curr.Id]; team != "" {
			seattleTeam = team
			break
		}
	}

	eventTime := e.Dates.Start.DateTime.In(SeattleTimeZone)
	if e.Dates.Start.TimeTBA && !e.Dates.Start.DateTBD {
		var err error
		eventTime, err = time.ParseInLocation("2006-01-02", e.Dates.Start.LocalDate, SeattleTimeZone)
		if err != nil {
			log.Error().Err(err).Str("venue_name", venueName).Str("event_name", e.Name).Msg("could not parse start time")
			return nil, fmt.Errorf("events: buildInternalEvent: could not parse start time: %w", err)
		}

		// since the event is Time TBA, just put it at noon
		eventTime = eventTime.Add(12 * time.Hour)
	}

	if seattleTeam == "" {
		// not a seattle sports team, just take event name and build that event
		return &Event{
			RawDescription: fmt.Sprintf("%s is at %s. It starts at %s", e.Name, venueName, eventTime.Format(localTimeDateFormat)),
			RawTime:        eventTime.Unix(),
		}, nil
	}

	// this code assumes there are only two "attractions" ... that should be good for any sports match?
	var opponentTeam string
	for _, curr := range e.Embedded.Attractions {
		if team := seattleTeamAttractionIDs[curr.Id]; team == "" {
			opponentTeam = curr.Name
		}
	}

	if opponentTeam == "" {
		log.Warn().Str("venue_name", venueName).Msg("could not find opponent attraction ID")
		opponentTeam = "some unknown opponent"
	}

	return &Event{
		TeamName:  seattleTeam,
		Venue:     venueName,
		LocalTime: eventTime.Format(localTimeDateFormat),
		Opponent:  opponentTeam,
		RawTime:   eventTime.Unix(),
	}, nil
}

func getEventsForVenueID(ctx context.Context, apiKey string, venueName string, venueID string, startDate time.Time, endDate time.Time) ([]*Event, []*Event, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, TicketmasterEventSearchAPI, nil)
	if err != nil {
		return nil, nil, err
	}

	q := req.URL.Query()
	q.Add("venueId", venueID)
	q.Add("apikey", apiKey)
	q.Add("startDateTime", startDate.Format(time.RFC3339))
	q.Add("endDateTime", endDate.Format(time.RFC3339))
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Str("status", resp.Status).Msg("could not read error response body")
			return nil, nil, fmt.Errorf("events: getEventForVenueID: could not read error body: %w", err)
		}
		log.Error().Str("status", resp.Status).Msg("error retrieving data from ticketmaster")
		return nil, nil, fmt.Errorf("events: getEventForVenueID: could not retireve data from ticketmaster: %s", string(body))
	}

	remainingRequestCount := resp.Header.Get("Rate-Limit-Available")
	rateLimitResetTime := resp.Header.Get("Rate-Limit-Reset")

	log.Info().Str("venue_name", venueName).Str("remaining_requests", remainingRequestCount).Str("rate_limit_reset_time", rateLimitResetTime).Msg("completed ticketmaster API request")

	var payload TicketmasterEventSearchResponse
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return nil, nil, err
	}

	var today []*Event
	var tomorrow []*Event

	for _, e := range payload.Embedded.Events {

		if eventShouldBeIgnored(&e) {
			log.Info().Str("venue", venueName).Str("event_name", e.Name).Msg("ignoring event")
			continue
		}

		log.Info().Str("venue_name", venueName).Str("event_name", e.Name).Msg("found event from ticketmaster")

		event, err := buildInternalEvent(e, venueName)
		if err != nil {
			continue
		}
		eventTime := e.Dates.Start.DateTime.In(SeattleTimeZone)
		if isDay(SeattleToday, eventTime) {
			today = append(today, event)
		} else if isDay(SeattleTomorrow, eventTime) {
			tomorrow = append(tomorrow, event)
		}
	}

	return today, tomorrow, nil

}

func getTicketmasterEvents(ctx context.Context) ([]*Event, []*Event, error) {
	ticketmasterApiKeySecretName := os.Getenv(TicketmasterApiKeySecretName)
	if ticketmasterApiKeySecretName == "" {
		log.Warn().Str("env_var_name", TicketmasterApiKeySecretName).Msg("environment variable not set. Not querying ticketmaster")
		return nil, nil, nil
	}

	apiKey, err := secrets.GetSecretString(ctx, ticketmasterApiKeySecretName)
	if err != nil {
		return nil, nil, fmt.Errorf("events: getTicketmasterEvents: could not get ticketmaster secret: %w", err)
	}

	today := time.Now().In(SeattleTimeZone)

	start := beginningOfDay(today)
	end := start.AddDate(0, 0, 2)

	var todayEvents []*Event
	var tomorrowEvents []*Event

	for venueName, venueID := range seattleVenueMap {
		var foundToday []*Event
		var foundTomorrow []*Event
		foundToday, foundTomorrow, err = getEventsForVenueID(ctx, apiKey, venueName, venueID, start, end)
		if err != nil {
			return nil, nil, fmt.Errorf("events: getTicketmasterEvents: could not query for ticketmaster data: %w", err)
		}
		if len(foundToday) > 0 {
			todayEvents = append(todayEvents, foundToday...)
		}
		if len(foundTomorrow) > 0 {
			tomorrowEvents = append(tomorrowEvents, foundTomorrow...)
		}

		err = ticketmasterRateLimiter.Wait(ctx)
		if err != nil {
			log.Error().Err(err).Msg("could not wait for ticketmaster rate limiter")
		}
	}

	return todayEvents, tomorrowEvents, nil
}
