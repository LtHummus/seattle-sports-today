package events

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-xray-sdk-go/v2/xray"
	"golang.org/x/time/rate"

	"github.com/rs/zerolog/log"

	"github.com/lthummus/seattle-sports-today/internal/secrets"
)

type eventFetcher func(ctx context.Context, today time.Time, tomorrow time.Time) ([]*Event, []*Event, error)

var (
	SeattleTimeZone *time.Location

	httpClient = xray.Client(&http.Client{
		Timeout: 5 * time.Second,
	})
)

const (
	localTimeDateFormat = "3:04 PM"
)

func init() {
	stz, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatal().Err(err).Msg("could not load seattle timezeone")
	}

	SeattleTimeZone = stz

	log.Info().Str("seattle_time_zone", SeattleTimeZone.String()).Msg("initialized time")
}

func isDay(target time.Time, specimen time.Time) bool {
	return target.Year() == specimen.Year() && target.YearDay() == specimen.YearDay()
}

type EventResults struct {
	TodayEvent     []*Event
	TomorrowEvents []*Event
}

type Event struct {
	ID               string `json:"id"`
	TeamName         string `json:"team_name"`
	Venue            string `json:"venue"`
	LocalTime        string `json:"local_time"`
	Opponent         string `json:"opponent"`
	ShortDescription string `json:"short_description"`

	RawDescription string `json:"raw_description,omitempty"`

	RawTime int64 `json:"raw_time"`
}

func (e *Event) CalendarSummary() string {
	if e.ShortDescription != "" {
		return e.ShortDescription
	}

	if e.RawDescription != "" {
		return e.RawDescription
	}

	return fmt.Sprintf("%s are playing against the %s at %s", e.TeamName, e.Opponent, e.Venue)
}

func (e *Event) String() string {
	if e.RawDescription != "" {
		return e.RawDescription
	}

	return fmt.Sprintf("%s are playing against the %s at %s. The game starts at %s.", e.TeamName, e.Opponent, e.Venue, e.LocalTime)
}

func fetchAndAppendEvents(ctx context.Context, fetcher eventFetcher, res *EventResults, eventLock *sync.Mutex, todayDate time.Time, tomorrowDate time.Time) error {
	today, tomorrow, e := fetcher(ctx, todayDate, tomorrowDate)
	if e != nil {
		return e
	}
	if len(today) == 0 && len(tomorrow) == 0 {
		return nil
	}

	eventLock.Lock()
	defer eventLock.Unlock()
	res.TodayEvent = append(res.TodayEvent, today...)
	res.TomorrowEvents = append(res.TomorrowEvents, tomorrow...)
	return nil
}

func GetTodayAndTomorrowGames(ctx context.Context, seattleToday time.Time, seattleTomorrow time.Time) (*EventResults, error) {
	var wg sync.WaitGroup

	res := &EventResults{}
	var eventLock sync.Mutex

	var errLock sync.Mutex
	var errs []error
	recordErr := func(source string, err error) {
		errLock.Lock()
		defer errLock.Unlock()
		errs = append(errs, fmt.Errorf("%s: %w", source, err))
	}

	wg.Go(func() {
		ticketmasterApiKeySecretName := os.Getenv(TicketmasterApiKeySecretName)
		if ticketmasterApiKeySecretName == "" {
			log.Warn().Str("env_var_name", TicketmasterApiKeySecretName).Msg("environment variable not set. Not querying ticketmaster")
			return
		}

		apiKey, err := secrets.GetSecretString(ctx, ticketmasterApiKeySecretName)
		if err != nil {
			recordErr("ticketmater", fmt.Errorf("events: getTicketmasterEvents: could not get ticketmaster secret: %w", err))
			return
		}

		tm := &ticketmasterFetcher{
			venues:        seattleVenueMap,
			attractionIDs: seattleTeamAttractionIDs,
			limiter:       rate.NewLimiter(3, 1),
			apiKey:        apiKey,
			baseURL:       TicketmasterDefaultBaseURL,
		}

		err = fetchAndAppendEvents(ctx, tm.GetEvents, res, &eventLock, seattleToday, seattleTomorrow)
		if err != nil {
			recordErr("ticketmaster", err)
		}
	})

	wg.Go(func() {
		err := fetchAndAppendEvents(ctx, getSpecialEvents, res, &eventLock, seattleToday, seattleTomorrow)
		if err != nil {
			recordErr("special_events", err)
		}
	})

	wg.Go(func() {
		err := fetchAndAppendEvents(ctx, GetUWGames, res, &eventLock, seattleToday, seattleTomorrow)
		if err != nil {
			recordErr("uw", err)
		}
	})

	wg.Wait()

	return res, errors.Join(errs...)
}
