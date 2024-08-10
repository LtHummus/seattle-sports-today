package events

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"golang.org/x/sync/errgroup"

	"github.com/rs/zerolog/log"
)

type eventFetcher func(ctx context.Context) (*Event, error)

var (
	SeattleTimeZone    *time.Location
	SeattleCurrentTime time.Time

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
	SeattleCurrentTime = time.Now().In(SeattleTimeZone)

	log.Info().Str("current_seattle_time", SeattleCurrentTime.Format(time.RFC850)).Str("seattle_time_zone", SeattleTimeZone.String()).Msg("initialized time")
}

type Event struct {
	TeamName  string `json:"string"`
	Venue     string `json:"venue"`
	LocalTime string `json:"local_time"`
	Opponent  string `json:"opponent"`
}

func fetchAndAdd(ctx context.Context, teamName string, f eventFetcher, eventList *[]*Event, lock *sync.Mutex) func() error {
	return func() error {
		return xray.Capture(ctx, fmt.Sprintf("Fetch.%s", teamName), func(ctx2 context.Context) error {
			event, err := f(ctx2)
			_ = xray.AddAnnotation(ctx2, "teamname", teamName)
			if err != nil {
				_ = xray.AddError(ctx2, err)
				return err
			}
			lock.Lock()
			defer lock.Unlock()

			if event == nil {
				_ = xray.AddAnnotation(ctx2, "gamefound", false)
				return nil
			}
			_ = xray.AddAnnotation(ctx2, "gamefound", true)
			*eventList = append(*eventList, event)
			return nil
		})
	}
}

func GetTodaysGames(ctx context.Context) ([]*Event, error) {

	eg, ctx2 := errgroup.WithContext(ctx)

	events := make([]*Event, 0)
	eventLock := &sync.Mutex{}

	eg.Go(fetchAndAdd(ctx2, "Sounders", GetSoundersGame, &events, eventLock))
	eg.Go(fetchAndAdd(ctx2, "Kraken", GetKrakenGame, &events, eventLock))
	eg.Go(fetchAndAdd(ctx2, "Mariners", GetMarinersGame, &events, eventLock))
	eg.Go(fetchAndAdd(ctx2, "Seahawks", GetSeahawksGame, &events, eventLock))
	eg.Go(fetchAndAdd(ctx2, "Storm", GetStormGame, &events, eventLock))
	eg.Go(fetchAndAdd(ctx2, "Reign", GetReignGame, &events, eventLock))
	eg.Go(fetchAndAdd(ctx2, "HuskiesFootball", GetHuskiesFootballGame, &events, eventLock))
	eg.Go(fetchAndAdd(ctx2, "SoundersLeaguesCup", GetSoundersLeagueCupGame, &events, eventLock))

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	return events, nil
}
