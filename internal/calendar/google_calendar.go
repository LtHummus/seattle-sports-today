package calendar

import (
	"context"
	"encoding/base32"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2/google"
	gcalendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/lthummus/seattle-sports-today/internal/events"
)

type Google struct {
	client       *gcalendar.Service
	eventService *gcalendar.EventsService

	calendarID string
}

func NewGoogleCalendar(ctx context.Context, apiKey string, calendarID string) (*Google, error) {
	creds, err := google.CredentialsFromJSONWithType(ctx, []byte(apiKey), google.ServiceAccount, gcalendar.CalendarScope)
	if err != nil {
		return nil, err
	}

	client, err := gcalendar.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, err
	}
	return &Google{
		client:       client,
		eventService: gcalendar.NewEventsService(client),
		calendarID:   calendarID,
	}, nil
}

func (g *Google) CreateEvent(ctx context.Context, event *events.Event) error {
	googleValidID := base32.HexEncoding.EncodeToString([]byte(event.ID))
	googleValidID = strings.ToLower(strings.TrimRight(googleValidID, "="))

	// for now, assume every event is 3 hours...but we could probably do better
	start := time.Unix(event.RawTime, 0)
	end := start.Add(3 * time.Hour)

	googleEvent := gcalendar.Event{
		Id:          googleValidID,
		Description: event.String(),
		Summary:     event.String(),
		Location:    event.Venue,
		Start:       &gcalendar.EventDateTime{DateTime: start.Format(time.RFC3339)},
		End:         &gcalendar.EventDateTime{DateTime: end.Format(time.RFC3339)},
	}
	createdEvent, err := g.eventService.Insert(g.calendarID, &googleEvent).Do()
	if gerr, ok := errors.AsType[*googleapi.Error](err); ok && gerr.Code == http.StatusConflict {
		// event already exists, so try to update
		log.Warn().Ctx(ctx).Str("event_id", googleValidID).Msg("event already exists, attempting to update")
		createdEvent, err = g.eventService.Update(g.calendarID, googleValidID, &googleEvent).Do()
	}
	if err != nil && !googleapi.IsNotModified(err) {
		log.Error().Ctx(ctx).Err(err).Str("event_description", event.String()).Msg("could not insert event")
		return fmt.Errorf("calendar: CreateEvent: could not create event: %w", err)
	}

	if googleapi.IsNotModified(err) {
		log.Warn().Ctx(ctx).Str("event_id", createdEvent.Id).Str("event_description", event.String()).Msg("event already existed in google calendar")
		return nil
	}

	log.Info().Ctx(ctx).Str("event_id", createdEvent.Id).Str("event_description", event.String()).Msg("event created in google calendar")
	return nil
}
