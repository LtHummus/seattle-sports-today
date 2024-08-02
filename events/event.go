package events

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/rs/zerolog/log"
)

type eventFetcher func(ctx context.Context) (*Event, error)

var (
	SeattleTimeZone    *time.Location
	SeattleCurrentTime time.Time
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
	SeattleCurrentTime = time.Now().In(SeattleTimeZone)

	log.Info().Str("current_seattle_time", SeattleCurrentTime.Format(time.RFC850)).Str("seattle_time_zone", SeattleTimeZone.String()).Msg("initialized time")
}

type Event struct {
	TeamName  string `json:"string"`
	Venue     string `json:"venue"`
	LocalTime string `json:"local_time"`
	Opponent  string `json:"opponent"`
}

func fetchAndAdd(ctx context.Context, f eventFetcher, eventList *[]*Event, lock *sync.Mutex) func() error {
	return func() error {
		event, err := f(ctx)
		if err != nil {
			return err
		}
		lock.Lock()
		defer lock.Unlock()

		if event == nil {
			return nil
		}
		*eventList = append(*eventList, event)
		return nil
	}
}

func GetTodaysGames(ctx context.Context) ([]*Event, error) {

	eg, ctx2 := errgroup.WithContext(ctx)

	events := make([]*Event, 0)
	eventLock := &sync.Mutex{}

	eg.Go(fetchAndAdd(ctx2, GetSoundersGame, &events, eventLock))
	eg.Go(fetchAndAdd(ctx2, GetKrakenGame, &events, eventLock))
	eg.Go(fetchAndAdd(ctx2, GetMarinersGame, &events, eventLock))
	eg.Go(fetchAndAdd(ctx2, GetSeahawksGame, &events, eventLock))
	eg.Go(fetchAndAdd(ctx2, GetStormGame, &events, eventLock))
	eg.Go(fetchAndAdd(ctx2, GetReignGame, &events, eventLock))
	eg.Go(fetchAndAdd(ctx2, GetHuskiesFootballGame, &events, eventLock))
	eg.Go(fetchAndAdd(ctx2, GetSoundersLeagueCupGame, &events, eventLock))

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	return events, nil
}
