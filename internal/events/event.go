package events

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"

	"github.com/rs/zerolog/log"

	"github.com/lthummus/seattle-sports-today/internal/secrets"
)

type eventFetcher func(ctx context.Context, today time.Time, tomorrow time.Time) ([]*Event, []*Event, error)

var (
	SeattleTimeZone *time.Location

	httpClient = xray.Client(http.DefaultClient)
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
	//SeattleToday = time.Now().In(SeattleTimeZone)
	//SeattleTomorrow = SeattleToday.AddDate(0, 0, 1)

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
	TeamName  string `json:"team_name"`
	Venue     string `json:"venue"`
	LocalTime string `json:"local_time"`
	Opponent  string `json:"opponent"`

	RawDescription string `json:"raw_description,omitempty"`

	RawTime int64 `json:"raw_time"`
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
	eg, ctx := errgroup.WithContext(ctx)

	res := &EventResults{}
	var eventLock sync.Mutex

	eg.Go(func() error {
		ticketmasterApiKeySecretName := os.Getenv(TicketmasterApiKeySecretName)
		if ticketmasterApiKeySecretName == "" {
			log.Warn().Str("env_var_name", TicketmasterApiKeySecretName).Msg("environment variable not set. Not querying ticketmaster")
			return nil
		}

		apiKey, err := secrets.GetSecretString(ctx, ticketmasterApiKeySecretName)
		if err != nil {
			return fmt.Errorf("events: getTicketmasterEvents: could not get ticketmaster secret: %w", err)
		}

		tm := &ticketmasterFetcher{
			venues:        seattleVenueMap,
			attractionIDs: seattleTeamAttractionIDs,
			limiter:       rate.NewLimiter(4, 1),
			apiKey:        apiKey,
			baseURL:       TicketmasterDefaultBaseURL,
		}

		return fetchAndAppendEvents(ctx, tm.GetEvents, res, &eventLock, seattleToday, seattleTomorrow)
	})

	eg.Go(func() error {
		return fetchAndAppendEvents(ctx, getSpecialEvents, res, &eventLock, seattleToday, seattleTomorrow)
	})

	eg.Go(func() error {
		return fetchAndAppendEvents(ctx, GetUWGames, res, &eventLock, seattleToday, seattleTomorrow)
	})

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	return res, nil
}
