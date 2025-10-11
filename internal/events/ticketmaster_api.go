package events

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"github.com/rs/zerolog/log"
)

const (
	TicketmasterDefaultBaseURL   = "https://app.ticketmaster.com"
	TicketmasterEventSearchAPI   = "%s/discovery/v2/events"
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

var seattleTeamAttractionIDs = map[string]string{
	"K8vZ917_vgV": "Seattle Kraken",
	"K8vZ9171oU7": "Seattle Seahawks",
	"K8vZ9171o6f": "Seattle Mariners",
	"K8vZ917G8RV": "Seattle Sounders",
	"K8vZ9178Dm7": "Seattle Reign",
	"K8vZ9171xo0": "Seattle Storm",
}

type ticketmasterFetcher struct {
	venues        map[string]string
	attractionIDs map[string]string
	limiter       *rate.Limiter
	apiKey        string
	baseURL       string
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
	eventTimeFormatted := eventTime.Format(localTimeDateFormat)
	if e.Dates.Start.TimeTBA && !e.Dates.Start.DateTBD {
		eventTimeFormatted = "TBA"
		var err error
		eventTime, err = time.ParseInLocation("2006-01-02", e.Dates.Start.LocalDate, SeattleTimeZone)
		if err != nil {
			log.Error().Err(err).Str("venue_name", venueName).Str("event_name", e.Name).Msg("could not parse event date for TBA")
			eventTime = e.Dates.Start.DateTime.In(SeattleTimeZone)
		}

		// we don't know the time, so just set it to noon for sorting purposes
		eventTime = eventTime.Add(12 * time.Hour)
	}

	if seattleTeam == "" {
		// not a seattle sports team, just take event name and build that event
		return &Event{
			RawDescription: fmt.Sprintf("%s is at %s. It starts at %s", e.Name, venueName, eventTimeFormatted),
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
		LocalTime: eventTimeFormatted,
		Opponent:  opponentTeam,
		RawTime:   eventTime.Unix(),
	}, nil
}

func (tm *ticketmasterFetcher) getEventsForVenueID(ctx context.Context, venueName string, venueID string, startDate time.Time, endDate time.Time) ([]*Event, []*Event, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(TicketmasterEventSearchAPI, tm.baseURL), nil)
	if err != nil {
		return nil, nil, err
	}

	q := req.URL.Query()
	q.Add("venueId", venueID)
	q.Add("apikey", tm.apiKey)
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

func (tm *ticketmasterFetcher) GetEvents(ctx context.Context) ([]*Event, []*Event, error) {
	var err error
	today := time.Now().In(SeattleTimeZone)

	start := beginningOfDay(today)
	end := start.AddDate(0, 0, 2)

	var todayEvents []*Event
	var tomorrowEvents []*Event

	for venueName, venueID := range seattleVenueMap {
		var foundToday []*Event
		var foundTomorrow []*Event
		foundToday, foundTomorrow, err = tm.getEventsForVenueID(ctx, venueName, venueID, start, end)
		if err != nil {
			return nil, nil, fmt.Errorf("events: getTicketmasterEvents: could not query for ticketmaster data: %w", err)
		}
		if len(foundToday) > 0 {
			todayEvents = append(todayEvents, foundToday...)
		}
		if len(foundTomorrow) > 0 {
			tomorrowEvents = append(tomorrowEvents, foundTomorrow...)
		}

		err = tm.limiter.Wait(ctx)
		if err != nil {
			log.Error().Err(err).Msg("could not wait for ticketmaster rate limiter")
		}
	}

	return todayEvents, tomorrowEvents, nil
}
